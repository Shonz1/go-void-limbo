// Package server accepts connections and hands each one to a client of its
// own, holding the one copy of everything those clients share: the packet
// registry, the game data, the server's key, and what a ping is answered from.
package server

import (
	"errors"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Shonz1/go-void-limbo/auth"
	"github.com/Shonz1/go-void-limbo/client"
	"github.com/Shonz1/go-void-limbo/gamedata"
	"github.com/Shonz1/go-void-limbo/protocol"
	"github.com/Shonz1/go-void-limbo/types"
)

// Config is what an operator decides about a server before it starts.
// WorldProvider is the world every joined connection is shown. The connection
// is what asks for it, so the connection's package is where it is defined.
type WorldProvider = client.WorldProvider

type Config struct {
	PacketRegistry *protocol.Registry
	GameData       *gamedata.Provider
	KeyPair        *auth.KeyPair
	SessionServer  client.SessionServer

	// World is the world a joined connection is shown, and nil on a server
	// that has none, which shows the void instead.
	World WorldProvider

	// Description is what a ping describes this server as.
	Description string

	// GameMode is the mode every joining player is put in. The zero value is
	// survival, the way the protocol numbers the modes; the creative default
	// this server ships with is package config's, applied where the
	// environment is read rather than here.
	GameMode types.GameMode

	// EncryptionEnabled decides what a login is worth: checked with Mojang
	// behind a cipher of this server's own, or taken on the word of whoever is
	// on the connection -- the proxy that forwarded it, or the client itself.
	EncryptionEnabled bool

	// ForwardingSecret is what a modern proxy signs the logins it forwards
	// with, and is empty on a server that no proxy was configured in front of.
	// Holding one puts a question to the proxy in front of every login: the
	// ones it answers are settled by the signature, and the ones it does not
	// are left to the setting above, since a secret does not stop anything else
	// reaching the port.
	ForwardingSecret []byte
}

// Server is what every connection shares: the registries a packet is resolved
// through, the game data the configuration phase hands out, and the key and
// session server a login is checked with.
type Server struct {
	packetRegistry *protocol.Registry
	gameRegistries *gamedata.Provider
	world          client.WorldProvider
	keyPair        *auth.KeyPair
	sessionServer  client.SessionServer

	// status is the whole of what the status phase answers with.
	status status

	// playerSync is the roster the joined players see each other through, one
	// for every connection like the status above it.
	playerSync *client.PlayerSync

	// entityIds hands each connection the id the play phase gives its player.
	// It only ever counts up, so a player that left never has its entity
	// confused with one that joined after it.
	entityIds atomic.Int32

	// clients, under clientsMu, is every connection currently being served,
	// held so the keep alive sweep can reach them all from one goroutine. The
	// map is made on first use, so a server built field by field works too.
	clientsMu sync.Mutex
	clients   map[*client.Client]struct{}

	// reactor serves the joined connections from one goroutine, on the
	// platforms that give it a poller. Nil where they do not, and everything
	// asked of it declines gracefully: joined connections then stay on
	// goroutines of their own.
	reactor *reactor

	// gameMode is the mode every joining player is put in.
	gameMode types.GameMode

	encryptionEnabled bool
	forwardingSecret  []byte
}

// New builds a server from what the operator decided about it.
func New(cfg Config) *Server {
	return &Server{
		packetRegistry:    cfg.PacketRegistry,
		gameRegistries:    cfg.GameData,
		world:             cfg.World,
		keyPair:           cfg.KeyPair,
		sessionServer:     cfg.SessionServer,
		status:            status{description: cfg.Description},
		playerSync:        client.NewPlayerSync(),
		gameMode:          cfg.GameMode,
		encryptionEnabled: cfg.EncryptionEnabled,
		forwardingSecret:  cfg.ForwardingSecret,
	}
}

// ListenAndServe listens on address and serves every connection that arrives,
// each on a goroutine of its own, until the listener fails.
func (s *Server) ListenAndServe(address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return s.Serve(listener)
}

// Serve serves every connection that arrives on listener, each on a goroutine
// of its own, until the listener fails or is closed. Taking the listener rather
// than an address lets the caller be the one that chose the port -- a test
// binds port zero and reads back what it got before the server is asked to run.
func (s *Server) Serve(listener net.Listener) error {
	defer listener.Close()

	slog.Info("TCP server is running", "address", listener.Addr().String())

	// One clock serves every connection its keep alives. A ticker and a
	// goroutine per connection would each wake the scheduler once per interval,
	// and across thousands of mostly idle connections that waking is most of
	// what serving them costs.
	stopSweep := make(chan struct{})
	defer close(stopSweep)

	go s.sweepKeepAlives(stopSweep, client.KeepAliveInterval)

	// Joined connections move onto the reactor where the platform has a
	// poller: one goroutine reading all of them beats a goroutine sleeping and
	// waking per packet on each. Where it has none, they stay on their own
	// goroutines, which serves fewer connections for the same money but serves
	// them all the same.
	if reactor, err := newReactor(s); err == nil {
		s.reactor = reactor
		go reactor.run()
	} else {
		slog.Info("joined connections stay on goroutines of their own", "err", err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			// A closed listener is the caller ending the serving, not a
			// connection that failed to arrive, and every Accept after it
			// would fail the same way forever.
			if errors.Is(err, net.ErrClosed) {
				return err
			}

			slog.Error("failed to accept connection", "err", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	slog.Info("new client connected", "addr", conn.RemoteAddr().String())

	c := client.New(conn, client.Config{
		PacketRegistry:    s.packetRegistry,
		GameData:          s.gameRegistries,
		World:             s.world,
		EntityId:          s.entityIds.Add(1),
		GameMode:          s.gameMode,
		PlayerSync:        s.playerSync,
		KeyPair:           s.keyPair,
		SessionServer:     s.sessionServer,
		Status:            &s.status,
		EncryptionEnabled: s.encryptionEnabled,
		ForwardingSecret:  s.forwardingSecret,
	})

	// Registered for as long as the connection is served, which is what the
	// keep alive sweep reaches it by. A connection the reactor took over is
	// its to close and unregister; one it did not is still this goroutine's.
	s.addClient(c)

	handedOff := false

	defer func() {
		if handedOff {
			return
		}

		s.removeClient(c)
		conn.Close()
	}()

	c.Run(func(joined *client.Client) bool {
		handedOff = s.reactor.take(joined, conn)
		return handedOff
	})
}

// sweepKeepAlives sends every connection its keep alives on the one shared
// clock until stop is closed. A connection that cannot take one, or let the
// last go unanswered, is closed, which is what ends its read loop.
func (s *Server) sweepKeepAlives(stop <-chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			for _, c := range s.snapshotClients() {
				if err := c.SendKeepAlive(); err != nil {
					slog.Error("dropping connection", "addr", c.RemoteAddr(), "err", err)
					s.disconnect(c)
				}
			}
		}
	}
}

// disconnect closes c's connection by whichever hand may close it: the
// reactor's loop when there is one, because a descriptor it polls has to be
// unmapped before it is closed, and directly when there is not.
func (s *Server) disconnect(c *client.Client) {
	if s.reactor != nil {
		s.reactor.dropClient(c)
		return
	}

	c.Close()
}

// snapshotClients copies the registry out from under its lock, so the sweep
// writes to connections without holding up the connections coming and going.
func (s *Server) snapshotClients() []*client.Client {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	clients := make([]*client.Client, 0, len(s.clients))
	for c := range s.clients {
		clients = append(clients, c)
	}

	return clients
}

func (s *Server) addClient(c *client.Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	if s.clients == nil {
		s.clients = make(map[*client.Client]struct{})
	}

	s.clients[c] = struct{}{}
}

func (s *Server) removeClient(c *client.Client) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	delete(s.clients, c)
}
