package main

import (
	"bytes"
	"crypto/subtle"
	"errors"
	"fmt"
	"go-void-limbo/auth"
	"go-void-limbo/gamedata"
	"go-void-limbo/handlers"
	clientboundCommon "go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	clientboundPlay "go-void-limbo/packets/clientbound/play"
	serverboundCommon "go-void-limbo/packets/serverbound/common"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/packets/serverbound/login"
	serverboundPlay "go-void-limbo/packets/serverbound/play"
	"go-void-limbo/registries"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"io"
	"log/slog"
	"net"
	"os"
	"reflect"
	"sync"
	"time"
)

const address = ":25565"

// maxPacketSize is the largest packet body the protocol allows (2^21 - 1 bytes).
const maxPacketSize = 2097151

// packetLogBlacklist holds the packet types that are never logged, in either
// direction. A joined client sends some of these on every tick, and at that rate
// they bury everything else the log has to say.
var packetLogBlacklist = map[reflect.Type]bool{
	reflect.TypeOf(serverboundPlay.ClientTickEndServerboundPacket{}): true,
}

// logPacket records a packet crossing the connection, unless its type is
// blacklisted. Every connection carries the same traffic, so this is detail one
// asks for rather than detail one is told.
func logPacket(message string, packet any) {
	packetType := reflect.TypeOf(packet)
	if packetType != nil && packetType.Kind() == reflect.Pointer {
		packetType = packetType.Elem()
	}

	if packetType != nil && packetLogBlacklist[packetType] {
		return
	}

	slog.Debug(message, "packet", packet)
}

// keepAliveInterval is how often a keep alive goes out, and equally how long an
// unanswered one has before the connection is given up on.
//
// Both ends drop a connection they have read nothing from for thirty seconds,
// so the interval has to leave room for an answer to come back inside that
// window; fifteen seconds is what a vanilla server uses and is half of it.
const keepAliveInterval = 15 * time.Second

// errKeepAliveTimeout is what a client that let a keep alive go unanswered for
// a whole interval leaves behind. Nothing on that connection is worth waiting
// for any longer, since the client either stopped reading or stopped existing.
var errKeepAliveTimeout = errors.New("keep alive went unanswered")

// packetError is a failure to make sense of a packet whose body was already
// read from the connection in full: an unknown id, or a decode that did not
// work out. The connection is still in sync and still usable, so the read loop
// reports these and carries on, unlike the read failures that end it.
type packetError struct {
	err error
}

func (e *packetError) Error() string { return e.err.Error() }

func (e *packetError) Unwrap() error { return e.err }

// compressBody frames a packet body for a connection that has been told a
// compression threshold, deflating it when it is big enough to be worth it. The
// var int in front carries the size the body inflates to, or zero for a body
// small enough to be left in full.
func compressBody(body []byte, threshold int32) ([]byte, error) {
	size := int32(0)
	payload := body

	if int32(len(body)) >= threshold {
		compressed, err := streams.Compress(body)
		if err != nil {
			return nil, fmt.Errorf("failed to compress packet: %w", err)
		}

		size = int32(len(body))
		payload = compressed
	}

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := stream.WriteVarInt(size); err != nil {
		return nil, err
	}

	if err := stream.Flush(); err != nil {
		return nil, err
	}

	return append(buf.Bytes(), payload...), nil
}

// decompressBody undoes compressBody on a body that arrived from a client that
// was told the same threshold. A size the client had no business compressing at
// is refused: a body it should have sent in full is one this end cannot tell
// from a frame that lost its place.
func decompressBody(body []byte, threshold int32) ([]byte, error) {
	size, read, err := streams.ReadVarIntFrom(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}

	payload := body[read:]

	if size == 0 {
		return payload, nil
	}

	if size < threshold || size > maxPacketSize {
		return nil, fmt.Errorf("invalid data length: %d", size)
	}

	return streams.Decompress(payload, size)
}

// serverId is the name a login goes by at the session server. It has been the
// empty string since the protocol stopped having anything to put in it, and both
// ends hash it all the same.
const serverId = ""

// sessionServer is the side of a login this server does not hold: the service
// that knows which accounts really logged in and what they look like.
type sessionServer interface {
	HasJoined(username, serverHash string) (types.GameProfile, error)
}

type MinecraftClient struct {
	conn           net.Conn
	stream         *streams.MinecraftStream
	packetRegistry *registries.PacketRegistry
	gameRegistries *gamedata.Provider

	// keyPair and sessionServer are shared by every connection: one key is
	// generated for the process, and one client talks to Mojang for all of them.
	keyPair       *auth.KeyPair
	sessionServer sessionServer

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

	// verifyToken is the token an encryption request went out with and has not
	// been answered yet, and sharedSecret is what the connection ended up
	// encrypted with. The secret outlives the packet that carried it because the
	// session server matches a login by a hash taken over it.
	verifyToken  []byte
	sharedSecret []byte

	// pendingKeepAlive is the id of the keep alive the client has not answered
	// yet, or zero when the server is not waiting on one.
	pendingKeepAlive int64

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

// compression reports the threshold packets are deflated at, and whether the
// client has been told one at all.
func (c *MinecraftClient) compression() (int32, bool) {
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
func (c *MinecraftClient) EnableCompression(threshold int32) error {
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

// BeginEncryption generates the verify token this connection's encryption
// request goes out with, and hands back what that packet has to carry.
func (c *MinecraftClient) BeginEncryption() ([]byte, []byte, error) {
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
func (c *MinecraftClient) CompleteEncryption(sharedSecret, verifyToken []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.verifyToken == nil {
		return errors.New("encryption was never requested")
	}

	token, err := c.keyPair.Decrypt(verifyToken)
	if err != nil {
		return fmt.Errorf("failed to decrypt the verify token: %w", err)
	}

	if subtle.ConstantTimeCompare(token, c.verifyToken) != 1 {
		return errors.New("the verify token is not the one that was sent")
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
func (c *MinecraftClient) Authenticate() (types.GameProfile, error) {
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

func (c *MinecraftClient) RegistryPackets() []types.ClientboundPacket {
	return c.gameRegistries.PacketsFor(c.ProtocolVersion())
}

func (c *MinecraftClient) ProtocolVersion() types.ProtocolVersion {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.protocolVersion
}

func (c *MinecraftClient) SetProtocolVersion(protocolVersion types.ProtocolVersion) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.protocolVersion = protocolVersion
}

func (c *MinecraftClient) Phase() types.Phase {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.phase
}

func (c *MinecraftClient) SetPhase(phase types.Phase) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.phase = phase
}

func (c *MinecraftClient) Profile() types.GameProfile {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.profile
}

func (c *MinecraftClient) SetProfile(profile types.GameProfile) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.profile = profile
}

// ReadPacket decodes the next packet and returns the handler registered for it,
// which may be nil when the packet needs no reaction. The packet body is consumed
// from the connection in full before decoding, so an unknown packet id or a failed
// decode cannot desynchronize subsequent reads. Those two are reported as a
// *packetError; anything else is the connection itself failing.
func (c *MinecraftClient) ReadPacket() (types.ServerboundPacket, types.PacketHandler, error) {
	length, err := c.stream.ReadVarInt()
	if err != nil {
		return nil, nil, err
	}

	if length < 1 || length > maxPacketSize {
		return nil, nil, fmt.Errorf("invalid packet length: %d", length)
	}

	body, err := c.stream.ReadBytes(length)
	if err != nil {
		return nil, nil, err
	}

	// A body that cannot be inflated is still a body that was read in full, so
	// the frames after it start where they should.
	if threshold, enabled := c.compression(); enabled {
		body, err = decompressBody(body, threshold)
		if err != nil {
			return nil, nil, &packetError{err: err}
		}
	}

	bodyStream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	packetId, err := bodyStream.ReadVarInt()
	if err != nil {
		return nil, nil, &packetError{err: err}
	}

	entry, ok := c.packetRegistry.GetServerbound(c.Phase(), c.ProtocolVersion(), packetId)
	if !ok || entry.Decoder == nil {
		return nil, nil, &packetError{err: fmt.Errorf("unknown packet id: %d", packetId)}
	}

	packet, err := entry.Decoder(bodyStream)
	if err != nil {
		return nil, nil, &packetError{err: fmt.Errorf("failed to decode packet: %w", err)}
	}

	logPacket("packet received", packet)

	return packet, entry.Handler, nil
}

func (c *MinecraftClient) WritePacket(packet types.ClientboundPacket) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.writePacket(packet)
}

// writePacket is WritePacket with the lock already held, for the callers that
// took it to decide what to write in the first place.
func (c *MinecraftClient) writePacket(packet types.ClientboundPacket) error {
	if packet == nil {
		return errors.New("packet is nil")
	}

	packetId := c.packetRegistry.GetClientboundId(c.phase, reflect.TypeOf(packet).Elem(), c.protocolVersion)
	if packetId == -1 {
		return errors.New("unknown packet id")
	}

	buf := new(bytes.Buffer)
	tempStream := streams.NewMinecraftStreamFromBuffer(buf)

	err := tempStream.WriteVarInt(packetId)
	if err != nil {
		return err
	}

	err = packet.Encode(tempStream)
	if err != nil {
		return err
	}

	err = tempStream.Flush()
	if err != nil {
		return err
	}

	body := buf.Bytes()

	if c.compressionEnabled {
		body, err = compressBody(body, c.compressionThreshold)
		if err != nil {
			return err
		}
	}

	err = c.stream.WriteVarInt(int32(len(body)))
	if err != nil {
		return err
	}

	err = c.stream.WriteBytes(body)
	if err != nil {
		return err
	}

	err = c.stream.Flush()
	if err != nil {
		return err
	}

	logPacket("packet sent", packet)

	return nil
}

// configureLogging sets the level the default logger keeps, read from LOG_LEVEL
// as one of DEBUG, INFO, WARN or ERROR. Packet traffic is logged at DEBUG, so it
// is silent until asked for.
func configureLogging() {
	level := slog.LevelInfo
	unrecognized := ""

	if raw, ok := os.LookupEnv("LOG_LEVEL"); ok {
		if err := level.UnmarshalText([]byte(raw)); err != nil {
			level = slog.LevelInfo
			unrecognized = raw
		}
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	if unrecognized != "" {
		slog.Warn("unrecognized LOG_LEVEL, falling back to INFO", "value", unrecognized)
	}
}

// ConfirmKeepAlive records the client's answer to the keep alive the server is
// waiting on.
func (c *MinecraftClient) ConfirmKeepAlive(id int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pendingKeepAlive == 0 {
		return fmt.Errorf("keep alive %d answers nothing that was sent", id)
	}

	if id != c.pendingKeepAlive {
		return fmt.Errorf("keep alive %d answers the wrong packet, expected %d", id, c.pendingKeepAlive)
	}

	c.pendingKeepAlive = 0

	return nil
}

// sendKeepAlive asks the client to prove it is still there, unless the last ask
// is still unanswered, which is errKeepAliveTimeout.
//
// Only configuration and play have a keep alive packet, so only they get one.
// The phases before them need none: a handshake and a login are exchanges the
// client drives from one packet to the next, and a client that stops driving
// one has stopped connecting rather than gone quiet.
func (c *MinecraftClient) sendKeepAlive() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.phase != types.PhaseConfiguration && c.phase != types.PhasePlay {
		return nil
	}

	if c.pendingKeepAlive != 0 {
		return errKeepAliveTimeout
	}

	// Any id the answer can be matched against works. The clock is what vanilla
	// uses, and it never repeats a value inside a connection.
	id := time.Now().UnixMilli()
	c.pendingKeepAlive = id

	return c.writePacket(&clientboundCommon.KeepAliveClientboundPacket{Id: id})
}

// keepAliveLoop sends a keep alive every interval until done is closed, and
// closes the connection when one is not answered or cannot be sent. Closing is
// what ends the read loop, which is what closes done.
func (c *MinecraftClient) keepAliveLoop(done <-chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := c.sendKeepAlive(); err != nil {
				slog.Error("dropping connection", "addr", c.conn.RemoteAddr(), "err", err)
				c.conn.Close()

				return
			}
		}
	}
}

// server is what every connection shares: the registries a packet is resolved
// through, the game data the configuration phase hands out, and the key and
// session server a login is checked with.
type server struct {
	packetRegistry *registries.PacketRegistry
	gameRegistries *gamedata.Provider
	keyPair        *auth.KeyPair
	sessionServer  sessionServer
}

func main() {
	configureLogging()

	packetRegistry := registries.NewPacketRegistry()

	registerPackets(packetRegistry)

	gameRegistries, err := gamedata.NewDefaultProvider()
	if err != nil {
		slog.Error("failed to encode game registries", "err", err)
		return
	}

	// One key for the process, generated before the first client can ask for it.
	// It is only ever used to get a login's secret across, so nothing is lost by
	// it going away with the process.
	keyPair, err := auth.NewKeyPair()
	if err != nil {
		slog.Error("failed to generate the server key", "err", err)
		return
	}

	srv := &server{
		packetRegistry: packetRegistry,
		gameRegistries: gameRegistries,
		keyPair:        keyPair,
		sessionServer:  auth.NewSessionServer(),
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("failed to start server", "err", err)
		return
	}

	defer listener.Close()

	slog.Info("TCP server is running", "address", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "err", err)
			continue
		}

		go srv.handleConnection(conn)
	}
}

func registerPackets(packetRegistry *registries.PacketRegistry) {
	packetRegistry.RegisterServerbound(types.PhaseHandshake, types.ProtocolVersions.ZERO, 0x00, handshake.DecodeHandshakeServerboundPacket, handlers.HandleHandshakeServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x00, login.DecodeLoginStartServerboundPacket, handlers.HandleLoginStartServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x01, login.DecodeEncryptionResponseServerboundPacket, handlers.HandleEncryptionResponseServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x03, login.DecodeLoginAcknowledgedServerboundPacket, handlers.HandleLoginAcknowledgedServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhaseConfiguration, types.ProtocolVersions.MINECRAFT_26_2, 0x03, configuration.DecodeAcknowledgeFinishConfigurationServerboundPacket, handlers.HandleAcknowledgeFinishConfigurationServerboundPacket)

	// The same keep alive in both phases that have one, under the id each phase
	// gives it.
	packetRegistry.RegisterServerbound(types.PhaseConfiguration, types.ProtocolVersions.MINECRAFT_26_2, 0x04, serverboundCommon.DecodeKeepAliveServerboundPacket, handlers.HandleKeepAliveServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x1C, serverboundCommon.DecodeKeepAliveServerboundPacket, handlers.HandleKeepAliveServerboundPacket)

	// What a joined client sends on its own. None of it needs a reaction from a
	// limbo, but a packet with no decoder is one the read loop can only report
	// as an unknown id, and the client sends these every tick.
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x00, serverboundPlay.DecodeAcceptTeleportationServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x0D, serverboundPlay.DecodeClientTickEndServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x1E, serverboundPlay.DecodeMovePlayerPositionServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x1F, serverboundPlay.DecodeMovePlayerPositionRotationServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x20, serverboundPlay.DecodeMovePlayerRotationServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x21, serverboundPlay.DecodeMovePlayerStatusServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x2C, serverboundPlay.DecodePlayerLoadedServerboundPacket, nil)

	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.DisconnectClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x00)
	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.EncryptionRequestClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x01)
	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x02)
	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.SetCompressionClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x03)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.RegistryDataClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x07)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.UpdateTagsClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x0D)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.FinishConfigurationClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x03)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x04)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x2C)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundPlay.GameEventClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x26)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x31)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x46)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x48)
}

func (s *server) handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	slog.Info("new client connected", "addr", remoteAddr)

	mc := &MinecraftClient{
		protocolVersion: types.ProtocolVersions.ZERO,
		phase:           types.PhaseHandshake,
		conn:            conn,
		stream:          streams.NewMinecraftStreamFromNetConn(conn),
		packetRegistry:  s.packetRegistry,
		gameRegistries:  s.gameRegistries,
		keyPair:         s.keyPair,
		sessionServer:   s.sessionServer,
	}

	// A limbo has nothing to say to a client that has arrived, and thirty
	// seconds of having nothing to say is what both ends treat as a dead
	// connection. Keep alives are the something, and they go out on a clock of
	// their own rather than in reaction to what the client sends.
	done := make(chan struct{})
	defer close(done)

	go mc.keepAliveLoop(done, keepAliveInterval)

	for {
		packet, handler, err := mc.ReadPacket()
		if err != nil {
			// A packet the server could not make sense of is one packet lost,
			// since its body was read in full and the next one starts where it
			// should. Anything else is the connection, including the close a
			// keep alive that went unanswered performs.
			var packetErr *packetError
			if errors.As(err, &packetErr) {
				slog.Error("failed to read packet", "err", err)
				continue
			}

			// A client that left, and a connection this server closed on a keep
			// alive that went unanswered, are both connections already
			// accounted for.
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) {
				slog.Error("connection lost", "addr", remoteAddr, "err", err)
			}

			return
		}

		if handler == nil {
			continue
		}

		if err := handler(mc, packet); err != nil {
			slog.Error("failed to handle packet", "packet", packet, "err", err)
		}
	}
}
