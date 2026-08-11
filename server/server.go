// Package server accepts connections and hands each one to a client of its
// own, holding the one copy of everything those clients share: the packet
// registry, the game data, the server's key, and what a ping is answered from.
package server

import (
	"log/slog"
	"net"

	"github.com/Shonz1/go-void-limbo/auth"
	"github.com/Shonz1/go-void-limbo/client"
	"github.com/Shonz1/go-void-limbo/gamedata"
	"github.com/Shonz1/go-void-limbo/protocol"
)

// Config is what an operator decides about a server before it starts.
type Config struct {
	PacketRegistry *protocol.Registry
	GameData       *gamedata.Provider
	KeyPair        *auth.KeyPair
	SessionServer  client.SessionServer

	// Description is what a ping describes this server as.
	Description string

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
	keyPair        *auth.KeyPair
	sessionServer  client.SessionServer

	// status is the whole of what the status phase answers with, and the only
	// state this server keeps about more than one connection at a time.
	status status

	encryptionEnabled bool
	forwardingSecret  []byte
}

// New builds a server from what the operator decided about it.
func New(cfg Config) *Server {
	return &Server{
		packetRegistry:    cfg.PacketRegistry,
		gameRegistries:    cfg.GameData,
		keyPair:           cfg.KeyPair,
		sessionServer:     cfg.SessionServer,
		status:            status{description: cfg.Description},
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

	defer listener.Close()

	slog.Info("TCP server is running", "address", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "err", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	slog.Info("new client connected", "addr", conn.RemoteAddr().String())

	client.New(conn, client.Config{
		PacketRegistry:    s.packetRegistry,
		GameData:          s.gameRegistries,
		KeyPair:           s.keyPair,
		SessionServer:     s.sessionServer,
		Status:            &s.status,
		EncryptionEnabled: s.encryptionEnabled,
		ForwardingSecret:  s.forwardingSecret,
	}).Run()
}
