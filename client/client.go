// Package client holds one accepted connection: the state a login moves
// through, the framing the connection is on, and the read loop that drives it.
// Keep alives are sent into it from outside, on the server's sweep over every
// connection it holds. Everything server-wide reaches it through the Config it
// is built with.
package client

import (
	"bytes"
	"crypto/cipher"
	"crypto/subtle"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/Shonz1/go-void-limbo/auth"
	"github.com/Shonz1/go-void-limbo/gamedata"
	clientboundConfiguration "github.com/Shonz1/go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	"github.com/Shonz1/go-void-limbo/protocol"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// serverId is the name a login goes by at the session server. It has been the
// empty string since the protocol stopped having anything to put in it, and both
// ends hash it all the same.
const serverId = ""

// maxPrePlayPacketSize is the largest frame a connection that has not reached
// the play phase may claim. The biggest thing a client legitimately sends
// before then is a login plugin response carrying a signed forwarding payload,
// a few kilobytes, so 32KB leaves room many times over -- without letting a
// connection that has not even logged in reserve the protocol's full two
// megabytes a frame.
const maxPrePlayPacketSize = 32767

// SessionServer is the side of a login this server does not hold: the service
// that knows which accounts really logged in and what they look like.
type SessionServer interface {
	HasJoined(username, serverHash string) (types.GameProfile, error)
}

// StatusProvider is the server-wide state a ping is answered from and the
// player count a connection joins and leaves. It is an interface here because
// the state belongs to the server, which is the package that constructs
// clients: the connection only ever asks about it.
type StatusProvider interface {
	// Status assembles what a ping arriving on a connection speaking version is
	// answered with.
	Status(version types.ProtocolVersion) types.ServerStatus

	// PlayerJoined counts a client that has just reached the play phase, and
	// PlayerLeft stops counting one whose connection has ended.
	PlayerJoined()
	PlayerLeft()
}

// WorldProvider is the world every joined connection is shown, prebuilt per
// protocol version. It is an interface here for the same reason StatusProvider
// is: the world belongs to the server, and the connection only ever asks.
type WorldProvider interface {
	// PacketsFor returns the packets that put the world on the wire of a
	// client speaking version, in sending order. The slice is shared across
	// connections and must not be modified.
	PacketsFor(version types.ProtocolVersion) []types.ClientboundPacket

	// Spawn is where the world puts a joining player.
	Spawn() (x, y, z float64)
}

// Config is everything a connection shares with the rest of its server.
type Config struct {
	PacketRegistry *protocol.Registry
	GameData       *gamedata.Provider

	// World is the world a joined connection is shown, and nil on a server
	// that has none to show, which is the void this server is named for.
	World WorldProvider

	// EntityId is the id the play phase gives this connection's player, unique
	// among every connection the server accepts for as long as it runs.
	EntityId int32

	// GameMode is the mode this connection's player is put in, which the
	// operator chose for the server. The zero value is survival, the way the
	// protocol numbers the modes; the creative default this server ships with
	// comes from package config rather than from here.
	GameMode types.GameMode

	// PlayerSync is the roster this player joins on reaching the world, shared
	// by every connection so the players on it can be shown each other. Nil
	// leaves each player alone in the world, which is what a test that built a
	// client by hand gets.
	PlayerSync *PlayerSync

	// KeyPair and SessionServer are shared by every connection: one key is
	// generated for the process, and one client talks to Mojang for all of them.
	KeyPair       *auth.KeyPair
	SessionServer SessionServer

	// Status is what a ping on this connection is answered from, and is the one
	// the server keeps rather than a copy: a ping asks about the server, and
	// nothing in it belongs to the connection it arrived on.
	Status StatusProvider

	// EncryptionEnabled is the setting the server was started with, and decides
	// what a login is worth: checked with Mojang behind a cipher of this
	// server's own, or taken on the word of whoever is on the connection.
	EncryptionEnabled bool

	// ForwardingSecret is what a modern proxy signs the logins it forwards
	// with, shared by every connection and empty on a server that no proxy was
	// configured in front of.
	ForwardingSecret []byte
}

// Client is one accepted connection and the state a packet handler may observe
// and mutate on it. It implements types.Client.
type Client struct {
	conn           net.Conn
	stream         *streams.MinecraftStream
	packetRegistry *protocol.Registry
	gameRegistries *gamedata.Provider

	// world is what a joined connection is shown, and nil on a server with
	// none. The pointer is handed over when the connection is accepted and
	// never changes.
	world WorldProvider

	// entityId, gameMode and playerSync are this player's standing among the
	// others: the id everything names its entity by, the mode it plays in,
	// and the roster it joins on reaching the world. All are handed over when
	// the connection is accepted and never change.
	entityId   int32
	gameMode   types.GameMode
	playerSync *PlayerSync

	// status is what a ping on this connection is answered from. The pointer is
	// handed over when the connection is accepted and never changes.
	status StatusProvider

	// keyPair and sessionServer are shared by every connection: one key is
	// generated for the process, and one client talks to Mojang for all of them.
	keyPair       *auth.KeyPair
	sessionServer SessionServer

	// encryptionEnabled is the setting the server was started with, and never
	// changes for the life of a connection, so it is read without the lock.
	encryptionEnabled bool

	// forwardingSecret is what a modern proxy signs the logins it forwards with,
	// shared by every connection and empty on a server that no proxy was
	// configured in front of. Like the setting above it, it is fixed for the
	// life of the process and read without the lock.
	forwardingSecret []byte

	// mu guards everything below it, and every write to the connection. Keep
	// alives are sent from a goroutine of their own while the read loop is
	// handling packets, so the state a write reads its packet id from is state
	// two goroutines reach for.
	//
	// Reading from the connection is not guarded and does not need to be: the
	// read and write halves of the stream buffer nothing in common, and a
	// net.Conn takes a read and a write at the same time. Turning the cipher on
	// rebuilds both halves, but it happens on the read loop's own goroutine and
	// under the lock, so it is not a write either half can be caught mid-read
	// by.
	mu              sync.Mutex
	protocolVersion types.ProtocolVersion
	phase           types.Phase
	profile         types.GameProfile

	// Where this connection's player is and how it stands, per the move and
	// input packets it sent, so a player joining later can be shown everyone
	// mid-pose rather than piled at the spawn.
	x, y, z    float64
	yaw, pitch float32
	onGround   bool
	sneaking   bool
	sprinting  bool

	// shownPlayers is the other players this connection has been shown, by
	// entity id. It is this connection's record rather than anything global,
	// and every spawn, removal and relay checks it under the same lock the
	// packets go out under, so what the client was told always matches it.
	shownPlayers map[int32]struct{}

	// forwardedLogin is the account a proxy vouched for, and forwarded says
	// whether one arrived at all. A BungeeCord proxy writes it into the
	// handshake, where nothing is configured about it: a connection carries one
	// or it does not, and only an unencrypted server looks. A modern proxy signs
	// it into a payload of its own, which only a server holding the secret ever
	// asks for.
	forwarded      bool
	forwardedLogin types.ForwardedLogin

	// pendingForwardingMessageId is the message id the forwarding request went
	// out under and has not been answered yet, or zero when this connection is
	// not waiting on one. It is what makes a payload an answer to something this
	// end asked rather than something a client volunteered.
	pendingForwardingMessageId int32

	// verifyToken is the token an encryption request went out with and has not
	// been answered yet, and sharedSecret is what the connection ended up
	// encrypted with. The secret outlives the packet that carried it because the
	// session server matches a login by a hash taken over it.
	verifyToken  []byte
	sharedSecret []byte

	// pendingKeepAlive is the id of the keep alive the client has not answered
	// yet, or zero when the server is not waiting on one.
	pendingKeepAlive int64

	// configurationFinished says finish configuration has gone out. The client
	// leaves the configuration phase the moment it reads that packet, before
	// its acknowledgement has crossed back and moved this end on, so the two
	// ends disagree about the phase for a round trip. A configuration packet
	// sent into that gap reaches a client already in play, which reads it under
	// a play id and fails on it; from this point until the acknowledgement,
	// nothing of the configuration phase may be sent.
	configurationFinished bool

	// readScratch backs the frame body the read loop is consuming, reused from
	// one packet to the next so a client streaming position updates is not a
	// stream of allocations. Only the read loop touches it, and nothing decoded
	// from a body keeps a byte of it: every field a decoder returns is a copy.
	readScratch []byte

	// decodeReader and decodeStream are the one stream every packet of this
	// connection is decoded through, pointed at each new body in turn rather
	// than built around every one. Only the read loop touches them, and they
	// are made on first use so a client built field by field works too.
	decodeReader *bytes.Reader
	decodeStream *streams.MinecraftStream

	// compressionEnabled says whether the client has been told a compression
	// threshold, which is what puts a data length in front of every packet body
	// in both directions. compressionThreshold is the size at or above which a
	// body is deflated; a threshold of zero deflates every one of them, so the
	// two cannot be the same field.
	//
	// A client is only ever told once, and the framing has to match on both
	// ends, so the zero value is the connection every client starts on: no
	// threshold sent, nothing framed for one.
	compressionEnabled   bool
	compressionThreshold int32
}

// New builds the client for a connection the server just accepted, at the
// start of everything: protocol zero, the handshake phase, and the plain
// framing.
func New(conn net.Conn, cfg Config) *Client {
	return &Client{
		protocolVersion:   types.ProtocolVersions.ZERO,
		phase:             types.PhaseHandshake,
		conn:              conn,
		stream:            streams.NewMinecraftStreamFromNetConn(conn),
		packetRegistry:    cfg.PacketRegistry,
		gameRegistries:    cfg.GameData,
		world:             cfg.World,
		entityId:          cfg.EntityId,
		gameMode:          cfg.GameMode,
		playerSync:        cfg.PlayerSync,
		keyPair:           cfg.KeyPair,
		sessionServer:     cfg.SessionServer,
		status:            cfg.Status,
		encryptionEnabled: cfg.EncryptionEnabled,
		forwardingSecret:  cfg.ForwardingSecret,
	}
}

// Close ends the connection, which is what ends the read loop driving it and
// everything the loop's return tears down behind it.
func (c *Client) Close() error {
	return c.conn.Close()
}

// RemoteAddr names the other end of the connection, for whoever has something
// to log about it.
func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// compression reports the threshold packets are deflated at, and whether the
// client has been told one at all.
func (c *Client) compression() (int32, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.compressionThreshold, c.compressionEnabled
}

// EnableCompression tells the client the body size at or above which packets
// are deflated, and frames everything written afterwards that way.
//
// The lock is held across both because the packet announcing the threshold is
// the last one framed uncompressed: a keep alive slipping in between would be
// framed one way and read the other, and the connection would never recover.
func (c *Client) EnableCompression(threshold int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if threshold < 0 {
		return fmt.Errorf("invalid compression threshold: %d", threshold)
	}

	if c.compressionEnabled {
		return errors.New("compression is already enabled")
	}

	if err := c.writePacket(&clientboundLogin.SetCompressionClientboundPacket{Threshold: threshold}); err != nil {
		return err
	}

	c.compressionEnabled = true
	c.compressionThreshold = threshold

	return nil
}

// EncryptionEnabled reports whether this connection is to be encrypted, and
// with it whether its login is checked with Mojang.
func (c *Client) EncryptionEnabled() bool {
	return c.encryptionEnabled
}

// SetForwardedLogin records the account a proxy vouched for in the handshake.
func (c *Client) SetForwardedLogin(forwarded types.ForwardedLogin) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.forwarded = true
	c.forwardedLogin = forwarded
}

// ForwardedLogin returns what the proxy vouched for, and whether anything did.
func (c *Client) ForwardedLogin() (types.ForwardedLogin, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.forwardedLogin, c.forwarded
}

// ModernForwardingEnabled reports whether this server holds a forwarding secret,
// and with it whether a login is expected to arrive signed by a proxy.
func (c *Client) ModernForwardingEnabled() bool {
	return len(c.forwardingSecret) > 0
}

// modernForwardingMessageId is the id the one forwarding request a connection
// ever sends goes out under.
//
// One fixed id is enough because there is only ever one request outstanding and
// only one that is ever sent. Nothing rests on it being hard to guess either: an
// answer is worth something because of the signature over it, and a client that
// echoes the right id without one is a client that gets no further than a client
// that echoes the wrong one.
const modernForwardingMessageId = 1

// BeginModernForwarding records that this connection is waiting on a forwarding
// payload, and returns the message id the request goes out under.
func (c *Client) BeginModernForwarding() (int32, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.ModernForwardingEnabled() {
		return 0, errors.New("no forwarding secret is configured")
	}

	if c.pendingForwardingMessageId != 0 {
		return 0, errors.New("the forwarding request has already been sent")
	}

	c.pendingForwardingMessageId = modernForwardingMessageId

	return c.pendingForwardingMessageId, nil
}

// CompleteModernForwarding checks a payload against the request this connection
// is waiting on and the secret the proxy shares, and returns the login it
// vouches for.
//
// A payload that gets this far settles the login on its own: the account, the
// name and the signed textures all come out of it, under a signature over the
// lot. What the client said about itself in login start is not consulted, since
// the point of the secret is that nothing on the connection has to be taken at
// its word.
func (c *Client) CompleteModernForwarding(messageId int32, payload []byte) (types.ForwardedLogin, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.checkPendingForwardingRequest(messageId); err != nil {
		return types.ForwardedLogin{}, err
	}

	forwarded, err := auth.ParseModernForwarding(c.forwardingSecret, payload)
	if err != nil {
		return types.ForwardedLogin{}, err
	}

	// The request has been answered, and a connection with nothing outstanding
	// is one a second payload cannot rewrite the login of.
	c.pendingForwardingMessageId = 0

	c.forwarded = true
	c.forwardedLogin = forwarded

	return forwarded, nil
}

// DeclineModernForwarding records that the connection answered the forwarding
// request without a login in it, which is what a client that has never heard of
// the channel answers.
//
// The request is answered either way, and a connection with nothing outstanding
// is one a payload arriving afterwards cannot rewrite the login of. That matters
// here as much as it does on the answer that worked: the login goes on to be
// settled without a proxy, and a signed payload turning up behind it would be
// settling it twice.
func (c *Client) DeclineModernForwarding(messageId int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.checkPendingForwardingRequest(messageId); err != nil {
		return err
	}

	c.pendingForwardingMessageId = 0

	return nil
}

// checkPendingForwardingRequest reports whether messageId answers the forwarding
// request this connection is actually waiting on, which is what makes an answer
// a reply to something this end asked rather than something a client
// volunteered.
//
// It is the one question both ways of answering the request have to ask, so it
// is asked in one place: an answer that carried a login and an answer that
// carried none are told apart by what they settle, not by what makes them an
// answer. Clearing the request is left to the caller, because the two clear it
// at different moments -- a payload only once it has been read, so that a
// signature this end could not verify leaves the question still open.
//
// The lock is held by the caller, which took it to decide what the answer was
// worth in the first place.
func (c *Client) checkPendingForwardingRequest(messageId int32) error {
	if c.pendingForwardingMessageId == 0 {
		return errors.New("no forwarding request is waiting on an answer")
	}

	if messageId != c.pendingForwardingMessageId {
		return fmt.Errorf("forwarding payload %d answers the wrong request, expected %d", messageId, c.pendingForwardingMessageId)
	}

	return nil
}

// BeginEncryption generates the verify token this connection's encryption
// request goes out with, and hands back what that packet has to carry.
func (c *Client) BeginEncryption() ([]byte, []byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.verifyToken != nil || c.sharedSecret != nil {
		return nil, nil, errors.New("encryption has already been requested")
	}

	verifyToken, err := auth.NewVerifyToken()
	if err != nil {
		return nil, nil, err
	}

	c.verifyToken = verifyToken

	return c.keyPair.PublicKey(), verifyToken, nil
}

// CompleteEncryption decrypts an encryption response and puts the connection
// under the secret it carried.
//
// The token is checked first and against the one this connection sent, which is
// what keeps a response from being worth anything anywhere but here. Nothing is
// written back before the cipher is on, because the client turned its own on the
// moment it sent this.
//
// A response with no token at all is what a client on a version that may sign
// the challenge sent instead of encrypting it: a signature under the profile
// key it announced in its hello, which the 1.19.3 step's upgrade takes off
// because nothing above 1.19.1 has a field for it, and which this server has
// no key left to check against. Such a response is let through on the
// session server's word alone, which is what settles every login here anyway:
// the client joined Mojang's session under the same secret before it answered,
// and a secret nobody joined under is one Authenticate refuses. A version that
// always encrypts the challenge sends no such thing, and an empty token from
// it is refused as any wrong token is.
func (c *Client) CompleteEncryption(sharedSecret, verifyToken []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.verifyToken == nil {
		return errors.New("encryption was never requested")
	}

	if len(verifyToken) == 0 && !c.protocolVersion.MaySignEncryptionChallenge() {
		return fmt.Errorf("the verify token is empty, which protocol %d has no way to answer with", c.protocolVersion.ID)
	}

	if len(verifyToken) != 0 {
		token, err := c.keyPair.Decrypt(verifyToken)
		if err != nil {
			return fmt.Errorf("failed to decrypt the verify token: %w", err)
		}

		if subtle.ConstantTimeCompare(token, c.verifyToken) != 1 {
			return errors.New("the verify token is not the one that was sent")
		}
	}

	secret, err := c.keyPair.Decrypt(sharedSecret)
	if err != nil {
		return fmt.Errorf("failed to decrypt the shared secret: %w", err)
	}

	if err := c.stream.EnableEncryption(secret); err != nil {
		return err
	}

	// The token has done its work, and a connection with nothing outstanding is
	// one a second encryption response cannot restart.
	c.verifyToken = nil
	c.sharedSecret = secret

	return nil
}

// Authenticate asks the session server about the login this connection is in the
// middle of, and returns the profile Mojang holds for it.
//
// The hash is the whole of the question: the client sent Mojang one taken over
// the same secret and the same key before it answered the encryption request, so
// a session server that recognizes this one is a session server that saw the
// same client on the other side of this connection.
func (c *Client) Authenticate() (types.GameProfile, error) {
	// The session server is on the far side of the internet and answers when it
	// answers. Holding the lock across that would stop a keep alive going out
	// for as long as it takes.
	c.mu.Lock()
	username := c.profile.Username
	secret := c.sharedSecret
	c.mu.Unlock()

	if secret == nil {
		return types.GameProfile{}, errors.New("the connection is not encrypted")
	}

	return c.sessionServer.HasJoined(username, auth.ServerHash(serverId, secret, c.keyPair.PublicKey()))
}

func (c *Client) RegistryPackets() []types.ClientboundPacket {
	return c.gameRegistries.PacketsFor(c.ProtocolVersion())
}

func (c *Client) WorldPackets() []types.ClientboundPacket {
	if c.world == nil {
		return nil
	}

	return c.world.PacketsFor(c.ProtocolVersion())
}

func (c *Client) WorldSpawn() (x, y, z float64, ok bool) {
	if c.world == nil {
		return 0, 0, 0, false
	}

	x, y, z = c.world.Spawn()

	return x, y, z, true
}

func (c *Client) ProtocolVersion() types.ProtocolVersion {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.protocolVersion
}

func (c *Client) SetProtocolVersion(protocolVersion types.ProtocolVersion) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.protocolVersion = protocolVersion
}

func (c *Client) Phase() types.Phase {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.phase
}

// SetPhase moves the connection on, and counts a client that has arrived in the
// play phase among the players a ping reports.
//
// Nothing moves a connection back out of play, so the count is joined once, and
// the connection ending is what leaves it. That is also why the phase alone says
// whether a connection ever counted, and no second field records it.
func (c *Client) SetPhase(phase types.Phase) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if phase == types.PhasePlay && c.phase != types.PhasePlay {
		c.status.PlayerJoined()
	}

	c.phase = phase
}

// LeavePlay stops counting a client whose connection has ended, if it ever got
// as far as being counted. It is called by whoever saw the connection end,
// which is the read loop for a connection that never left it and the reactor
// for one that did.
//
// The phase is the whole of that question: nothing moves a connection back out
// of play, so one that ends there is a player leaving, and one that ends
// anywhere else never arrived.
func (c *Client) LeavePlay() {
	c.mu.Lock()
	left := c.phase == types.PhasePlay
	if left {
		c.status.PlayerLeft()
	}
	c.mu.Unlock()

	// The roster is left off this connection's lock: taking a player off it
	// writes to every other connection, and each of those writes takes the
	// other connection's lock of its own.
	if left {
		c.leavePlayerSync()
	}
}

// TakeoverRead hands the read half of the connection to whoever will read it
// from here on: everything already read and not consumed, and the cipher the
// bytes still on the wire are under, nil when the connection is plain. The
// client's own read loop must be done reading before this is called, and
// nothing may read through the client again after it.
func (c *Client) TakeoverRead() ([]byte, cipher.Stream, error) {
	return c.stream.TakeoverRead()
}

// ServerStatus assembles what a ping on this connection is answered with, which
// is the server's own status read at the version this connection speaks.
func (c *Client) ServerStatus() types.ServerStatus {
	return c.status.Status(c.ProtocolVersion())
}

func (c *Client) Profile() types.GameProfile {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.profile
}

func (c *Client) SetProfile(profile types.GameProfile) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.profile = profile
}

// ReadPacket decodes the next packet and returns the handler registered for it,
// which may be nil when the packet needs no reaction. The packet body is consumed
// from the connection in full before decoding, so an unknown packet id or a failed
// decode cannot desynchronize subsequent reads. Those two are reported as a
// *packetError; anything else is the connection itself failing.
//
// The id says which packet arrived on the version the client speaks, and the
// body is then carried up to the latest version before it is decoded, since
// that is the only version a decoder exists for.
func (c *Client) ReadPacket() (types.ServerboundPacket, types.PacketHandler, error) {
	// The frame length and the size a compressed body claims to inflate to are
	// both the client's word, allocated for before a byte behind them is read.
	// Nothing legitimate comes near the protocol maximum before play, so a
	// connection that has not got that far is not taken at two megabytes of it.
	phase := c.Phase()

	maxSize := int32(streams.MaxPacketSize)
	if phase != types.PhasePlay {
		maxSize = maxPrePlayPacketSize
	}

	body, err := c.stream.ReadFrameInto(&c.readScratch, maxSize)
	if err != nil {
		return nil, nil, err
	}

	return c.decodeBody(phase, maxSize, body)
}

// decodeBody turns one framed body, already off the wire in full, into the
// packet it carries and the handler registered for it. It is the half of
// ReadPacket both read paths share: the read loop hands it what ReadFrame
// returned, and the reactor hands it the frames it reassembled itself. The
// body may alias a buffer the caller reuses; nothing decoded keeps a byte of
// it.
func (c *Client) decodeBody(phase types.Phase, maxSize int32, body []byte) (types.ServerboundPacket, types.PacketHandler, error) {
	var err error

	// A body that cannot be inflated is still a body that was read in full, so
	// the frames after it start where they should.
	if threshold, enabled := c.compression(); enabled {
		body, err = streams.DecompressBody(body, threshold, maxSize)
		if err != nil {
			return nil, nil, &packetError{err: err}
		}
	}

	packetId, idSize, err := streams.ReadVarIntFrom(body)
	if err != nil {
		return nil, nil, &packetError{err: err}
	}

	protocolVersion := c.ProtocolVersion()

	packetType, ok := c.packetRegistry.GetServerboundType(phase, protocolVersion, packetId)
	if !ok {
		return nil, nil, &packetError{err: fmt.Errorf("unknown packet id: %d", packetId)}
	}

	entry, ok := c.packetRegistry.GetServerbound(phase, packetType)
	if !ok || entry.Decoder == nil {
		return nil, nil, &packetError{err: fmt.Errorf("no decoder for packet id %d", packetId)}
	}

	// The body is whole and in memory, so the payload is the body past the id
	// rather than anything read back out of it.
	payload, err := c.packetRegistry.UpgradeBody(phase, packetType, protocolVersion, body[idSize:])
	if err != nil {
		return nil, nil, &packetError{err: err}
	}

	if c.decodeReader == nil {
		c.decodeReader = bytes.NewReader(nil)
		c.decodeStream = streams.NewMinecraftStreamFromBytesReader(c.decodeReader)
	}

	c.decodeReader.Reset(payload)

	packet, err := entry.Decoder(c.decodeStream)
	if err != nil {
		return nil, nil, &packetError{err: fmt.Errorf("failed to decode packet: %w", err)}
	}

	logPacket("packet received", packet)

	return packet, entry.Handler, nil
}

func (c *Client) WritePacket(packet types.ClientboundPacket) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.writePacket(packet)
}

// writePacket is WritePacket with the lock already held, for the callers that
// took it to decide what to write in the first place.
func (c *Client) writePacket(packet types.ClientboundPacket) error {
	// The client left configuration on reading finish configuration, and reads
	// whatever arrives next as play. Refused here rather than sent, since a
	// packet the client cannot make sense of is a connection it drops.
	if c.phase == types.PhaseConfiguration && c.configurationFinished {
		return fmt.Errorf("%s follows finish configuration, which the client has already left configuration on", packet)
	}

	if prepared, ok := packet.(*types.PreparedPacket); ok {
		return c.writePrepared(prepared)
	}

	// The packet encodes itself at the latest version, which is the only one it
	// knows how to be, and is then carried back down to the version the client
	// speaks, with the id it goes out under at that version in front.
	body, err := c.packetRegistry.EncodeClientbound(c.phase, c.protocolVersion, packet)
	if err != nil {
		return err
	}

	if c.compressionEnabled {
		body, err = streams.CompressBody(body, c.compressionThreshold)
		if err != nil {
			return err
		}
	}

	if err := c.stream.WriteFrame(body); err != nil {
		return err
	}

	// Sent, so from here on the client is on its way out of configuration and
	// nothing else of that phase may follow it.
	if _, ok := packet.(*clientboundConfiguration.FinishConfigurationClientboundPacket); ok {
		c.configurationFinished = true
	}

	logPacket("packet sent", packet)

	return nil
}

// writePrepared writes a packet that was put in wire form ahead of time, for
// this client's phase and version -- a packet prepared for any other is a
// mistake in whoever handed it over, and is refused rather than sent as
// something the client would read as noise.
//
// The deflated bytes go out as they are when this connection would have
// deflated the body itself: a threshold was announced and the body reaches it.
// Otherwise the body is inflated and framed the way the connection frames any
// other, so what was prepared is only ever a saving, never a difference.
func (c *Client) writePrepared(prepared *types.PreparedPacket) error {
	if prepared == nil {
		return errors.New("packet is nil")
	}

	if prepared.Phase != c.phase || prepared.Version != c.protocolVersion.ID {
		return fmt.Errorf("%s is not prepared for phase %d on protocol %d", prepared, c.phase, c.protocolVersion.ID)
	}

	if c.compressionEnabled && prepared.Size >= c.compressionThreshold {
		if err := c.stream.WriteCompressedFrame(prepared.Size, prepared.Deflated); err != nil {
			return err
		}
	} else {
		body, err := prepared.Body()
		if err != nil {
			return err
		}

		if c.compressionEnabled {
			err = c.stream.WriteCompressedFrame(0, body)
		} else {
			err = c.stream.WriteFrame(body)
		}

		if err != nil {
			return err
		}
	}

	logPacket("packet sent", prepared)

	return nil
}
