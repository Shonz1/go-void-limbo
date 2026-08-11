package main

import (
	"bytes"
	"crypto/subtle"
	"errors"
	"flag"
	"fmt"
	"go-void-limbo/auth"
	"go-void-limbo/gamedata"
	clientboundCommon "go-void-limbo/packets/clientbound/common"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	serverboundPlay "go-void-limbo/packets/serverbound/play"
	"go-void-limbo/registries"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"io"
	"log/slog"
	"net"
	"os"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
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

// playerCount is how many clients have reached the play phase, which is the
// number a server list ping is answered with.
//
// One is shared by every connection a server accepts: a ping asks about the
// server rather than about the connection it arrived on. It is read while a ping
// is being answered and written as connections join and leave, each on a
// goroutine of its own, so it is atomic rather than under any one connection's
// lock.
type playerCount struct {
	count atomic.Int64
}

// join counts a client that has just reached the play phase, leave stops
// counting one whose connection has ended, and online is what a ping reports.
//
// A nil count is one nothing is counted in, which is what a connection with no
// server behind it has: the number belongs to the server, and there is none here
// to hold one.
func (p *playerCount) join() {
	if p != nil {
		p.count.Add(1)
	}
}

func (p *playerCount) leave() {
	if p != nil {
		p.count.Add(-1)
	}
}

func (p *playerCount) online() int32 {
	if p == nil {
		return 0
	}

	return int32(p.count.Load())
}

type MinecraftClient struct {
	conn           net.Conn
	stream         *streams.MinecraftStream
	packetRegistry *registries.PacketRegistry
	gameRegistries *gamedata.Provider

	// description is what a ping describes this server as, and players is the
	// count every connection on this server shares. A ping is answered before
	// anything about the connection has been settled, so both are handed over
	// when it is accepted and neither changes afterwards.
	description string
	players     *playerCount

	// keyPair and sessionServer are shared by every connection: one key is
	// generated for the process, and one client talks to Mojang for all of them.
	keyPair       *auth.KeyPair
	sessionServer sessionServer

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

// EncryptionEnabled reports whether this connection is to be encrypted, and
// with it whether its login is checked with Mojang.
func (c *MinecraftClient) EncryptionEnabled() bool {
	return c.encryptionEnabled
}

// SetForwardedLogin records the account a proxy vouched for in the handshake.
func (c *MinecraftClient) SetForwardedLogin(forwarded types.ForwardedLogin) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.forwarded = true
	c.forwardedLogin = forwarded
}

// ForwardedLogin returns what the proxy vouched for, and whether anything did.
func (c *MinecraftClient) ForwardedLogin() (types.ForwardedLogin, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.forwardedLogin, c.forwarded
}

// ModernForwardingEnabled reports whether this server holds a forwarding secret,
// and with it whether a login is expected to arrive signed by a proxy.
func (c *MinecraftClient) ModernForwardingEnabled() bool {
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
func (c *MinecraftClient) BeginModernForwarding() (int32, error) {
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
func (c *MinecraftClient) CompleteModernForwarding(messageId int32, payload []byte) (types.ForwardedLogin, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pendingForwardingMessageId == 0 {
		return types.ForwardedLogin{}, errors.New("no forwarding request is waiting on an answer")
	}

	if messageId != c.pendingForwardingMessageId {
		return types.ForwardedLogin{}, fmt.Errorf("forwarding payload %d answers the wrong request, expected %d", messageId, c.pendingForwardingMessageId)
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
func (c *MinecraftClient) DeclineModernForwarding(messageId int32) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pendingForwardingMessageId == 0 {
		return errors.New("no forwarding request is waiting on an answer")
	}

	if messageId != c.pendingForwardingMessageId {
		return fmt.Errorf("forwarding payload %d answers the wrong request, expected %d", messageId, c.pendingForwardingMessageId)
	}

	c.pendingForwardingMessageId = 0

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

// SetPhase moves the connection on, and counts a client that has arrived in the
// play phase among the players a ping reports.
//
// Nothing moves a connection back out of play, so the count is joined once, and
// the connection ending is what leaves it. That is also why the phase alone says
// whether a connection ever counted, and no second field records it.
func (c *MinecraftClient) SetPhase(phase types.Phase) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if phase == types.PhasePlay && c.phase != types.PhasePlay {
		c.players.join()
	}

	c.phase = phase
}

// leavePlay stops counting a client whose connection has ended, if it ever got
// as far as being counted.
//
// The phase is the whole of that question: nothing moves a connection back out
// of play, so one that ends there is a player leaving, and one that ends
// anywhere else never arrived.
func (c *MinecraftClient) leavePlay() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.phase == types.PhasePlay {
		c.players.leave()
	}
}

// statusVersion is the version a ping is answered with: the client's own
// whenever this server speaks it, so that a client on any supported version sees
// a server it can join rather than one number it has to call incompatible.
//
// A client on a version this server does not speak left the connection on
// protocol zero, and is told the latest instead. That is the answer it came for:
// it has no use for a version it cannot join at, only for something to draw
// beside the fact that it cannot.
func statusVersion(version types.ProtocolVersion) types.ServerVersion {
	if !types.IsSupportedProtocolVersion(version) {
		version = types.LatestProtocolVersion
	}

	return types.ServerVersion{Name: version.Names[0], Protocol: version.ID}
}

// ServerStatus assembles what a ping on this connection is answered with.
func (c *MinecraftClient) ServerStatus() types.ServerStatus {
	online := c.players.online()

	return types.ServerStatus{
		Version: statusVersion(c.ProtocolVersion()),
		Players: types.ServerPlayers{
			Online: online,

			// A limbo turns nobody away, and the protocol has no way of saying
			// so: the field is a number, and a client draws a server as full when
			// the two are equal. One more than however many are on is the closest
			// this can come to the truth, and it is a truth about this server
			// rather than a number an operator was asked to invent.
			Max: online + 1,
		},
		Description: types.TextComponent{Text: c.description},
	}
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
//
// The id says which packet arrived on the version the client speaks, and the
// body is then carried up to the latest version before it is decoded, since
// that is the only version a decoder exists for.
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

	phase := c.Phase()
	protocolVersion := c.ProtocolVersion()

	packetType, ok := c.packetRegistry.GetServerboundType(phase, protocolVersion, packetId)
	if !ok {
		return nil, nil, &packetError{err: fmt.Errorf("unknown packet id: %d", packetId)}
	}

	entry, ok := c.packetRegistry.GetServerbound(phase, packetType)
	if !ok || entry.Decoder == nil {
		return nil, nil, &packetError{err: fmt.Errorf("no decoder for packet id %d", packetId)}
	}

	payload, err := bodyStream.ReadRest()
	if err != nil {
		return nil, nil, &packetError{err: err}
	}

	payload, err = c.packetRegistry.UpgradeBody(phase, packetType, protocolVersion, payload)
	if err != nil {
		return nil, nil, &packetError{err: err}
	}

	packet, err := entry.Decoder(streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(payload)))
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

	packetType := reflect.TypeOf(packet).Elem()

	packetId := c.packetRegistry.GetClientboundId(c.phase, packetType, c.protocolVersion)
	if packetId == -1 {
		return errors.New("unknown packet id")
	}

	// The packet encodes itself at the latest version, which is the only one it
	// knows how to be, and is then carried back down to the version the client
	// speaks.
	payloadBuf := new(bytes.Buffer)
	payloadStream := streams.NewMinecraftStreamFromBuffer(payloadBuf)

	err := packet.Encode(payloadStream)
	if err != nil {
		return err
	}

	err = payloadStream.Flush()
	if err != nil {
		return err
	}

	payload, err := c.packetRegistry.DowngradeBody(c.phase, packetType, c.protocolVersion, payloadBuf.Bytes())
	if err != nil {
		return err
	}

	// The id goes in front of the body it was resolved for, at the version the
	// client reads both at.
	buf := new(bytes.Buffer)
	tempStream := streams.NewMinecraftStreamFromBuffer(buf)

	err = tempStream.WriteVarInt(packetId)
	if err != nil {
		return err
	}

	err = tempStream.Flush()
	if err != nil {
		return err
	}

	body := append(buf.Bytes(), payload...)

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

// encryptionSetting reports whether connections are to be encrypted, read from
// ENCRYPTION as anything strconv.ParseBool accepts.
//
// It defaults to on, and an unrecognized value is treated as on rather than
// refused, because every way of misreading this setting has to land on the safe
// side: an unencrypted server is one anyone can log in to under anyone's name,
// and nothing about a connection says which of the two it got.
func encryptionSetting() bool {
	raw, ok := os.LookupEnv("ENCRYPTION")
	if !ok {
		return true
	}

	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		slog.Warn("unrecognized ENCRYPTION, falling back to enabled", "value", raw)
		return true
	}

	return enabled
}

// defaultDescription is what the server list draws under this server's address
// when the operator has said nothing about it. It says what the server is, which
// is the one thing a player looking at a list of them needs to be told.
const defaultDescription = "A void limbo"

// descriptionSetting reports what a ping describes this server as, read from
// MOTD.
//
// An empty value is treated as nothing said rather than as an empty description,
// because a blank line in a server list is indistinguishable from a server that
// failed to answer.
func descriptionSetting() string {
	if raw, ok := os.LookupEnv("MOTD"); ok && raw != "" {
		return raw
	}

	return defaultDescription
}

// forwardingSecretFlag is the secret on the command line, which is the one place
// an operator can put it that does not outlive the process. It wins over the
// environment when both are set, because it is the more deliberate of the two:
// the environment is inherited and a flag is typed.
var forwardingSecretFlag = flag.String("forwarding-secret", "", "the secret a modern proxy signs the logins it forwards with, taken from FORWARDING_SECRET when this is empty")

// forwardingSecretSetting reports the secret a modern proxy shares with this
// server, taken from the flag when it has one and from FORWARDING_SECRET
// otherwise.
//
// There is no setting that turns forwarding on. The secret is the setting: a
// server that holds one is a server behind a proxy, and asks every login for a
// payload signed with it. A server that holds none never asks, and logins there
// are settled as they were before any of this existed.
//
// So an empty value is no secret rather than an empty one. A secret nobody set
// is a secret everybody has, and a server that asked for a signature under it
// would be checking a signature anyone can produce, which is worse than not
// asking at all.
func forwardingSecretSetting(argument string) []byte {
	if argument != "" {
		return []byte(argument)
	}

	if raw, ok := os.LookupEnv("FORWARDING_SECRET"); ok && raw != "" {
		return []byte(raw)
	}

	return nil
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
// The phases before them need none: a handshake, a status ping and a login are
// exchanges the client drives from one packet to the next, and a client that
// stops driving one has stopped connecting rather than gone quiet.
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

	// description is what a ping describes this server as, and players is the
	// count every connection joins and leaves. They are the whole of what the
	// status phase answers with, and the only state this server keeps about more
	// than one connection at a time.
	description string
	players     playerCount

	// encryptionEnabled is what every connection this server accepts is handed,
	// and it decides what a login is worth: checked with Mojang behind a cipher
	// of this server's own, or taken on the word of whoever is on the connection
	// -- the proxy that forwarded it, or the client itself.
	encryptionEnabled bool

	// forwardingSecret is what a modern proxy signs the logins it forwards with,
	// and is empty on a server that no proxy was configured in front of. Holding
	// one puts a question to the proxy in front of every login: the ones it
	// answers are settled by the signature, and the ones it does not are left to
	// the setting above, since a secret does not stop anything else reaching the
	// port.
	forwardingSecret []byte
}

func main() {
	configureLogging()

	flag.Parse()

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

	encryptionEnabled := encryptionSetting()

	// The two settings answer different questions, so neither overrules the
	// other. A secret says what a forwarded login is worth: the proxy holds the
	// connection with the player and asked Mojang there, so nothing on this side
	// of it is asked to encrypt anything, and the signature is the whole of the
	// check. Encryption says what a login nobody forwarded is worth, which is
	// still a login this server has to settle, since holding a secret does not
	// stop anything else reaching the port.
	forwardingSecret := forwardingSecretSetting(*forwardingSecretFlag)
	if len(forwardingSecret) > 0 {
		slog.Info("a forwarding secret is configured, a login signed with it is taken from the proxy that signed it and is not checked with Mojang here")
	}

	if !encryptionEnabled {
		// The one thing this costs is the only thing anyone would want back, so
		// it is said out loud rather than left to be discovered. A login here is
		// taken on the word of whoever is on the connection, which is the proxy's
		// when one forwarded it and the client's when none did, so the port
		// should be one only what the operator trusts can reach.
		slog.Warn("encryption is disabled, logins nobody forwarded are taken on the word of whoever connects and are not checked with Mojang")
	}

	srv := &server{
		packetRegistry:    packetRegistry,
		gameRegistries:    gameRegistries,
		keyPair:           keyPair,
		sessionServer:     auth.NewSessionServer(),
		description:       descriptionSetting(),
		encryptionEnabled: encryptionEnabled,
		forwardingSecret:  forwardingSecret,
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

func (s *server) handleConnection(conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	slog.Info("new client connected", "addr", remoteAddr)

	mc := &MinecraftClient{
		protocolVersion:   types.ProtocolVersions.ZERO,
		phase:             types.PhaseHandshake,
		conn:              conn,
		stream:            streams.NewMinecraftStreamFromNetConn(conn),
		packetRegistry:    s.packetRegistry,
		gameRegistries:    s.gameRegistries,
		keyPair:           s.keyPair,
		sessionServer:     s.sessionServer,
		description:       s.description,
		players:           &s.players,
		encryptionEnabled: s.encryptionEnabled,
		forwardingSecret:  s.forwardingSecret,
	}

	// A client that joined stops being counted when its connection ends, and one
	// that never joined was never counted.
	defer mc.leavePlay()

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
