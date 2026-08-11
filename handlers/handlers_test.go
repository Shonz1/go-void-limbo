package handlers

import (
	"bytes"
	"errors"
	"go-void-limbo/auth"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	clientboundPlay "go-void-limbo/packets/clientbound/play"
	clientboundStatus "go-void-limbo/packets/clientbound/status"
	"go-void-limbo/packets/serverbound/common"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/packets/serverbound/login"
	serverboundStatus "go-void-limbo/packets/serverbound/status"
	"go-void-limbo/types"
	"slices"
	"testing"
)

type fakeClient struct {
	protocolVersion types.ProtocolVersion
	phase           types.Phase
	profile         types.GameProfile
	written         []types.ClientboundPacket
	// writePhases records the phase the client was in as each packet was
	// written, since that phase is what the real client resolves a clientbound
	// packet id from.
	writePhases     []types.Phase
	registryPackets []types.ClientboundPacket

	// confirmedKeepAlives records the ids answered, and keepAliveErr is what the
	// real client would report for an answer that matches nothing.
	confirmedKeepAlives []int64
	keepAliveErr        error

	// compressionThresholds records the thresholds compression was enabled at,
	// and compressionAfter how many packets had been written by then. The
	// threshold only applies to what follows it, so when it was announced
	// matters as much as that it was.
	compressionThresholds []int32
	compressionAfter      int
	compressionErr        error

	// encryptionEnabled is the setting the server was started with. The zero
	// value is a server that encrypts nothing, so the tests about an encrypted
	// login say so.
	encryptionEnabled bool

	// forwarded and forwardedLogin are what the handshake left behind. The zero
	// value is a connection no proxy forwarded, which is every connection to a
	// server nothing is pointed at.
	forwarded      bool
	forwardedLogin types.ForwardedLogin

	// forwardingSecret stands in for the one the server holds: the zero value is
	// a server no proxy was configured in front of, and anything else is one
	// that asks every login for a payload signed with it.
	forwardingSecret []byte

	// forwardingMessageIds records the ids the requests went out under, and
	// beginForwardingErr is a connection that could not ask.
	forwardingMessageIds []int32
	beginForwardingErr   error

	// completedForwarding records the message id and payload of every answer
	// that got as far as the connection. forwardedResult is the login the real
	// client would read out of a payload it verified, and completeForwardingErr
	// one it would refuse.
	completedForwardingIds      []int32
	completedForwardingPayloads [][]byte
	forwardedResult             types.ForwardedLogin
	completeForwardingErr       error

	// declinedForwardingIds records the requests given up on, which is what an
	// answer carrying no login leaves behind, and declineForwardingErr is a
	// connection that had nothing to give up on.
	declinedForwardingIds []int32
	declineForwardingErr  error

	// publicKey and verifyToken are what the real client hands back to put in an
	// encryption request, and beginErr is a connection that could not produce
	// them.
	publicKey   []byte
	verifyToken []byte
	beginErr    error

	// completedSecrets and completedTokens record the two fields of every
	// encryption response that got as far as the connection, and encryptedAfter
	// how many packets had been written by then. Everything the client sends
	// after its response is encrypted, so a reply written before the cipher is
	// on is a reply it cannot read.
	completedSecrets [][]byte
	completedTokens  [][]byte
	encryptedAfter   int
	completeErr      error

	// serverStatus is what the real client would answer a ping with. The handler
	// has no say in any of it, so a status the tests can recognize is enough to
	// tell a response built from the connection's own from one built from
	// anything else.
	serverStatus types.ServerStatus

	// authenticated is the profile the session server would answer with, and
	// authenticateErr a client it has no record of. authenticateAfter is how
	// many packets had been written when it was asked.
	authenticated     types.GameProfile
	authenticateCalls int
	authenticateAfter int
	authenticateErr   error
}

func (c *fakeClient) RegistryPackets() []types.ClientboundPacket { return c.registryPackets }

func (c *fakeClient) ProtocolVersion() types.ProtocolVersion { return c.protocolVersion }

func (c *fakeClient) SetProtocolVersion(protocolVersion types.ProtocolVersion) {
	c.protocolVersion = protocolVersion
}

func (c *fakeClient) Phase() types.Phase { return c.phase }

func (c *fakeClient) SetPhase(phase types.Phase) { c.phase = phase }

func (c *fakeClient) Profile() types.GameProfile { return c.profile }

func (c *fakeClient) SetProfile(profile types.GameProfile) { c.profile = profile }

func (c *fakeClient) WritePacket(packet types.ClientboundPacket) error {
	c.written = append(c.written, packet)
	c.writePhases = append(c.writePhases, c.phase)
	return nil
}

func (c *fakeClient) ServerStatus() types.ServerStatus { return c.serverStatus }

func (c *fakeClient) EnableCompression(threshold int32) error {
	if c.compressionErr != nil {
		return c.compressionErr
	}

	c.compressionThresholds = append(c.compressionThresholds, threshold)
	c.compressionAfter = len(c.written)

	return nil
}

func (c *fakeClient) ConfirmKeepAlive(id int64) error {
	c.confirmedKeepAlives = append(c.confirmedKeepAlives, id)
	return c.keepAliveErr
}

func (c *fakeClient) EncryptionEnabled() bool { return c.encryptionEnabled }

func (c *fakeClient) SetForwardedLogin(forwarded types.ForwardedLogin) {
	c.forwarded = true
	c.forwardedLogin = forwarded
}

func (c *fakeClient) ForwardedLogin() (types.ForwardedLogin, bool) {
	return c.forwardedLogin, c.forwarded
}

func (c *fakeClient) ModernForwardingEnabled() bool { return len(c.forwardingSecret) > 0 }

func (c *fakeClient) BeginModernForwarding() (int32, error) {
	if c.beginForwardingErr != nil {
		return 0, c.beginForwardingErr
	}

	// Any id the answer can be matched against works, and the real client sends
	// one request per connection under one of them.
	messageId := int32(len(c.forwardingMessageIds) + 1)
	c.forwardingMessageIds = append(c.forwardingMessageIds, messageId)

	return messageId, nil
}

func (c *fakeClient) CompleteModernForwarding(messageId int32, payload []byte) (types.ForwardedLogin, error) {
	c.completedForwardingIds = append(c.completedForwardingIds, messageId)
	c.completedForwardingPayloads = append(c.completedForwardingPayloads, payload)

	if c.completeForwardingErr != nil {
		return types.ForwardedLogin{}, c.completeForwardingErr
	}

	c.forwarded = true
	c.forwardedLogin = c.forwardedResult

	return c.forwardedResult, nil
}

func (c *fakeClient) DeclineModernForwarding(messageId int32) error {
	if c.declineForwardingErr != nil {
		return c.declineForwardingErr
	}

	c.declinedForwardingIds = append(c.declinedForwardingIds, messageId)

	return nil
}

func (c *fakeClient) BeginEncryption() ([]byte, []byte, error) {
	if c.beginErr != nil {
		return nil, nil, c.beginErr
	}

	return c.publicKey, c.verifyToken, nil
}

func (c *fakeClient) CompleteEncryption(sharedSecret, verifyToken []byte) error {
	if c.completeErr != nil {
		return c.completeErr
	}

	c.completedSecrets = append(c.completedSecrets, sharedSecret)
	c.completedTokens = append(c.completedTokens, verifyToken)
	c.encryptedAfter = len(c.written)

	return nil
}

func (c *fakeClient) Authenticate() (types.GameProfile, error) {
	c.authenticateCalls++
	c.authenticateAfter = len(c.written)

	if c.authenticateErr != nil {
		return types.GameProfile{}, c.authenticateErr
	}

	return c.authenticated, nil
}

func TestHandleHandshakeServerboundPacket(t *testing.T) {
	client := &fakeClient{}
	packet := &handshake.HandshakeServerboundPacket{
		ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
		ServerAddress:   "localhost",
		ServerPort:      25565,
		Intent:          int32(types.PhaseLogin),
	}

	if err := HandleHandshakeServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.ProtocolVersion().ID != types.ProtocolVersions.MINECRAFT_26_2.ID {
		t.Errorf("expected protocol version %d, got %d", types.ProtocolVersions.MINECRAFT_26_2.ID, client.ProtocolVersion().ID)
	}

	if client.Phase() != types.PhaseLogin {
		t.Errorf("expected phase %d, got %d", types.PhaseLogin, client.Phase())
	}
}

// Two phases are a handshake's to ask for. The rest are ones a connection
// reaches by getting through the phase before, so a handshake that names one is
// refused rather than obeyed, and the connection is left where it started.
func TestHandleHandshakeServerboundPacketRefusesAnIntentThatIsNotAPingOrALogin(t *testing.T) {
	// Play, the two phases either side of what a handshake may name, and the
	// numbers that arrive as play once an intent has been narrowed to a byte.
	for _, intent := range []int32{
		int32(types.PhaseHandshake),
		int32(types.PhaseConfiguration),
		int32(types.PhasePlay),
		int32(types.PhasePlay) + 256,
		int32(types.PhasePlay) + 512,
	} {
		client := &fakeClient{}
		packet := &handshake.HandshakeServerboundPacket{
			ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
			ServerAddress:   "localhost",
			ServerPort:      25565,
			Intent:          intent,
		}

		if err := HandleHandshakeServerboundPacket(client, packet); err == nil {
			t.Errorf("intent %d: error = nil, want an intent a handshake may not name refused", intent)
		}

		if client.Phase() != types.PhaseHandshake {
			t.Errorf("intent %d: phase = %d, want the connection left in the handshake phase %d", intent, client.Phase(), types.PhaseHandshake)
		}
	}
}

// forwardedAddress is the handshake address a proxy sends: the address it was
// reached at, and then the login it is forwarding, joined by null bytes.
const forwardedAddress = "limbo.example\x00203.0.113.7\x00069a79f444e94726a5befca90e38aaf5\x00" +
	`[{"name":"textures","value":"a base64 blob","signature":"a signature"}]`

func TestHandleHandshakeServerboundPacketKeepsTheLoginTheProxyForwarded(t *testing.T) {
	client := &fakeClient{}
	packet := &handshake.HandshakeServerboundPacket{
		ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
		ServerAddress:   forwardedAddress,
		ServerPort:      25565,
		Intent:          int32(types.PhaseLogin),
	}

	if err := HandleHandshakeServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only the handshake carries it, and the login it belongs to arrives in the
	// packet after, so a handshake it is not taken from is a login that has to
	// fall back to the name the client typed.
	forwarded, ok := client.ForwardedLogin()
	if !ok {
		t.Fatal("kept no forwarded login, want the one the handshake carried")
	}

	if forwarded.Uuid != "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Errorf("uuid = %q, want the account the proxy vouched for", forwarded.Uuid)
	}

	if len(forwarded.Properties) != 1 {
		t.Errorf("kept %d properties, want the textures the proxy forwarded", len(forwarded.Properties))
	}
}

func TestHandleHandshakeServerboundPacketKeepsNothingFromAPlainAddress(t *testing.T) {
	client := &fakeClient{}

	// What a client that connected here itself sends: the address it was told to
	// connect to, with nothing appended to it and no account in it.
	packet := &handshake.HandshakeServerboundPacket{
		ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
		ServerAddress:   "limbo.example",
		Intent:          int32(types.PhaseLogin),
	}

	if err := HandleHandshakeServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := client.ForwardedLogin(); ok {
		t.Error("kept a forwarded login, want an address nobody wrote a login into read as an address")
	}
}

// A handshake that carries an account is only worth reading on a connection
// whose login was going to be taken on somebody's word regardless. A server
// that asks Mojang does not look, since believing the fields would be believing
// whoever wrote them, and that is the one way left to get past the question.
func TestHandleHandshakeServerboundPacketIgnoresAForwardedAddressOnAnEncryptedServer(t *testing.T) {
	client := &fakeClient{encryptionEnabled: true}
	packet := &handshake.HandshakeServerboundPacket{
		ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
		ServerAddress:   forwardedAddress,
		Intent:          int32(types.PhaseLogin),
	}

	if err := HandleHandshakeServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := client.ForwardedLogin(); ok {
		t.Error("kept a forwarded login, want an account nobody here vouched for ignored")
	}
}

func TestHandleHandshakeServerboundPacketReportsAForwardedLoginItCannotRead(t *testing.T) {
	client := &fakeClient{}

	// An address with a proxy's separators in it and no account this end can
	// make out. Something wrote a login here, and reading past it silently is
	// how a player ends up as somebody else.
	packet := &handshake.HandshakeServerboundPacket{
		ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
		ServerAddress:   "limbo.example\x00203.0.113.7\x00not a uuid",
		Intent:          int32(types.PhaseLogin),
	}

	if err := HandleHandshakeServerboundPacket(client, packet); err == nil {
		t.Error("error = nil, want a forwarded login that could not be read reported")
	}

	// The login start behind it falls back to the client's own word for want of
	// this, so a half-read handshake must not leave anything for it to find.
	if _, ok := client.ForwardedLogin(); ok {
		t.Error("kept a forwarded login, want nothing from a handshake that could not be read")
	}

	// The protocol version and the phase are what the packet says regardless:
	// they are the client's own, and the fields behind them are the proxy's.
	if client.Phase() != types.PhaseLogin {
		t.Errorf("expected phase %d, got %d", types.PhaseLogin, client.Phase())
	}
}

func TestHandleLoginStartServerboundPacketAsksForEncryption(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin, encryptionEnabled: true, publicKey: []byte("a public key"), verifyToken: []byte{0x01, 0x02, 0x03, 0x04}}
	packet := &login.LoginStartServerboundPacket{Name: "Notch", Uuid: "00000000-0000-0000-0000-000000000001"}

	if err := HandleLoginStartServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The client encrypts everything it sends after answering this, so a second
	// packet sent alongside it is one that arrives in a framing the client has
	// already stopped reading for.
	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	request, ok := client.written[0].(*clientboundLogin.EncryptionRequestClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.EncryptionRequestClientboundPacket, got %T", client.written[0])
	}

	if !bytes.Equal(request.PublicKey, client.publicKey) {
		t.Errorf("public key = % x, want the server's % x", request.PublicKey, client.publicKey)
	}

	// A token the client cannot send back leaves nothing to check its answer
	// against.
	if !bytes.Equal(request.VerifyToken, client.verifyToken) {
		t.Errorf("verify token = % x, want the one the connection is waiting on % x", request.VerifyToken, client.verifyToken)
	}

	// A client that is not asked to authenticate never tells Mojang it joined,
	// and the session server then has no record of a login this end is about to
	// ask it about.
	if !request.ShouldAuthenticate {
		t.Error("should authenticate = false, want the client to tell Mojang it joined")
	}

	// Compression is announced on the far side of the cipher, so nothing is
	// framed for a threshold the client has not been told yet.
	if len(client.compressionThresholds) != 0 {
		t.Errorf("enabled compression at %v, want it left until the connection is encrypted", client.compressionThresholds)
	}

	// The client stays in the login phase until it acknowledges the success packet.
	if client.Phase() != types.PhaseLogin {
		t.Errorf("expected phase %d, got %d", types.PhaseLogin, client.Phase())
	}

	// Nothing else on the connection carries the name the client logged in
	// under, and the session server is asked about it by that name.
	if client.Profile().Username != packet.Name || client.Profile().Uuid != packet.Uuid {
		t.Errorf("kept profile %s, want the one the client logged in with", client.Profile())
	}
}

func TestHandleLoginStartServerboundPacketReportsAFailureToBeginEncryption(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin, encryptionEnabled: true, beginErr: errors.New("no verify token")}

	// A request sent without a token to check the answer against is a login that
	// cannot be finished, so it is not sent.
	if err := HandleLoginStartServerboundPacket(client, &login.LoginStartServerboundPacket{Name: "Notch"}); err == nil {
		t.Error("error = nil, want the failure to begin encryption passed back")
	}

	if len(client.written) != 0 {
		t.Errorf("wrote %d packets, want none after encryption could not be begun", len(client.written))
	}
}

func TestHandleLoginStartServerboundPacketFinishesALoginWithoutEncryption(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin, publicKey: []byte("a public key"), verifyToken: []byte{0x01, 0x02, 0x03, 0x04}}
	packet := &login.LoginStartServerboundPacket{Name: "Notch", Uuid: "00000000-0000-0000-0000-000000000001"}

	if err := HandleLoginStartServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// An encryption request is what puts the client's cipher on, so sending one
	// on a server that will never encrypt leaves the client reading through a
	// cipher this end is not writing through.
	for _, written := range client.written {
		if _, ok := written.(*clientboundLogin.EncryptionRequestClientboundPacket); ok {
			t.Fatal("asked for encryption, want a login left in the clear")
		}
	}

	// There is no secret to hash, so there is nothing the session server could
	// be asked about.
	if client.authenticateCalls != 0 {
		t.Errorf("asked the session server %d times, want a login nobody can vouch for left alone", client.authenticateCalls)
	}

	if !slices.Equal(client.compressionThresholds, []int32{compressionThreshold}) {
		t.Errorf("enabled compression at %v, want the threshold announced once at %d", client.compressionThresholds, compressionThreshold)
	}

	if client.compressionAfter != 0 {
		t.Errorf("enabled compression after %d packets, want it announced before any reply", client.compressionAfter)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	loginSuccess, ok := client.written[0].(*clientboundLogin.LoginSuccessClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginSuccessClientboundPacket, got %T", client.written[0])
	}

	// The uuid is derived from the name, so the same name is the same account
	// every time rather than whatever the client felt like sending.
	want := types.OfflineUuid(packet.Name)

	if loginSuccess.Profile.Uuid != want {
		t.Errorf("login success uuid = %q, want the offline %q", loginSuccess.Profile.Uuid, want)
	}

	if loginSuccess.Profile.Username != packet.Name {
		t.Errorf("login success username = %q, want the name the client logged in under %q", loginSuccess.Profile.Username, packet.Name)
	}

	// Nothing signed a skin for an account nobody vouched for, and an unsigned
	// property is one the client refuses.
	if len(loginSuccess.Profile.Properties) != 0 {
		t.Errorf("carried %d properties, want none for a profile nobody signed", len(loginSuccess.Profile.Properties))
	}

	if client.Profile().String() != loginSuccess.Profile.String() {
		t.Errorf("kept profile = %s, want the one the client was welcomed with %s", client.Profile(), loginSuccess.Profile)
	}

	if loginSuccess.SessionId == "" {
		t.Error("expected a generated session id, got an empty string")
	}

	// The client stays in the login phase until it acknowledges the success
	// packet.
	if client.Phase() != types.PhaseLogin {
		t.Errorf("expected phase %d, got %d", types.PhaseLogin, client.Phase())
	}
}

func TestHandleLoginStartServerboundPacketFinishesTheLoginTheProxyForwarded(t *testing.T) {
	signature := "a signature"
	forwarded := types.ForwardedLogin{
		Address:    "203.0.113.7",
		Uuid:       "069a79f4-44e9-4726-a5be-fca90e38aaf5",
		Properties: []types.ProfileProperty{{Name: "textures", Value: "a base64 blob", Signature: &signature}},
	}

	client := &fakeClient{phase: types.PhaseLogin, forwarded: true, forwardedLogin: forwarded}

	// The uuid the packet carries is the proxy's own choosing and is not the one
	// the login is finished with; the name is, because the proxy sends it here
	// and nowhere else.
	packet := &login.LoginStartServerboundPacket{Name: "Notch", Uuid: "00000000-0000-0000-0000-000000000001"}

	if err := HandleLoginStartServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The proxy holds the connection with the player and asked Mojang on it
	// already. There is nobody on this connection to ask, and no secret to ask
	// about.
	for _, written := range client.written {
		if _, ok := written.(*clientboundLogin.EncryptionRequestClientboundPacket); ok {
			t.Fatal("asked for encryption, want a login the proxy already settled")
		}
	}

	if client.authenticateCalls != 0 {
		t.Errorf("asked the session server %d times, want a login the proxy already asked about left alone", client.authenticateCalls)
	}

	if !slices.Equal(client.compressionThresholds, []int32{compressionThreshold}) {
		t.Errorf("enabled compression at %v, want the threshold announced once at %d", client.compressionThresholds, compressionThreshold)
	}

	if client.compressionAfter != 0 {
		t.Errorf("enabled compression after %d packets, want it announced before any reply", client.compressionAfter)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	loginSuccess, ok := client.written[0].(*clientboundLogin.LoginSuccessClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginSuccessClientboundPacket, got %T", client.written[0])
	}

	want := types.GameProfile{Uuid: forwarded.Uuid, Username: packet.Name, Properties: forwarded.Properties}

	// The account is the one the proxy authenticated, down to the textures,
	// which are the only way anyone is shown a skin.
	if loginSuccess.Profile.String() != want.String() {
		t.Errorf("login success profile = %s, want the forwarded %s", loginSuccess.Profile, want)
	}

	if client.Profile().String() != want.String() {
		t.Errorf("kept profile = %s, want the forwarded %s", client.Profile(), want)
	}

	if loginSuccess.SessionId == "" {
		t.Error("expected a generated session id, got an empty string")
	}

	// The client stays in the login phase until it acknowledges the success
	// packet.
	if client.Phase() != types.PhaseLogin {
		t.Errorf("expected phase %d, got %d", types.PhaseLogin, client.Phase())
	}
}

// A handshake and the login start behind it, on a server that asks Mojang.
// Writing a proxy's account into the address is what anyone wanting past that
// question would try, and it buys nothing: the account is never read, and the
// login is held to the same exchange as any other.
func TestAForwardedAddressCannotStandInForTheMojangCheck(t *testing.T) {
	client := &fakeClient{encryptionEnabled: true, publicKey: []byte("a public key"), verifyToken: []byte{0x01, 0x02, 0x03, 0x04}}

	handshakePacket := &handshake.HandshakeServerboundPacket{
		ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
		ServerAddress:   forwardedAddress,
		Intent:          int32(types.PhaseLogin),
	}

	if err := HandleHandshakeServerboundPacket(client, handshakePacket); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := HandleLoginStartServerboundPacket(client, &login.LoginStartServerboundPacket{Name: "Notch"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	// An encryption request, which is a login that has not been settled by
	// anything the client sent, rather than the success packet the forwarded
	// account was written there to earn.
	if _, ok := client.written[0].(*clientboundLogin.EncryptionRequestClientboundPacket); !ok {
		t.Fatalf("expected *login.EncryptionRequestClientboundPacket, got %T", client.written[0])
	}

	if client.Profile().Uuid == "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Error("took the account out of the address, want it left to Mojang to answer")
	}
}

// forwardingSecret stands in for the one an operator shares between the proxy
// and this server. Nothing in these tests takes a digest under it, since the
// connection is what checks that; what they are about is that the question is
// asked and that no login is finished without an answer to it.
var forwardingSecret = []byte("a shared secret")

// A login start on a server that holds a forwarding secret. Nothing is settled
// by it: the question goes to whoever is on the connection, and the login waits
// for the answer.
func TestHandleLoginStartServerboundPacketAsksTheProxyForTheForwardedLogin(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin, forwardingSecret: forwardingSecret}

	if err := HandleLoginStartServerboundPacket(client, &login.LoginStartServerboundPacket{Name: "Notch"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	request, ok := client.written[0].(*clientboundLogin.LoginPluginRequestClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginPluginRequestClientboundPacket, got %T", client.written[0])
	}

	if request.Channel != auth.ModernForwardingChannel {
		t.Errorf("asked on %q, want the channel a proxy forwards a login on %q", request.Channel, auth.ModernForwardingChannel)
	}

	// The version asked for is the version answered at, so it is the one field
	// the request carries.
	if !bytes.Equal(request.Data, []byte{auth.ModernForwardingVersion}) {
		t.Errorf("asked with % x, want the forwarding version %d", request.Data, auth.ModernForwardingVersion)
	}

	if !slices.Equal(client.forwardingMessageIds, []int32{request.MessageId}) {
		t.Errorf("sent message ids %v, want the one the request went out under %d", client.forwardingMessageIds, request.MessageId)
	}

	// Neither of the two ways a login is settled without a proxy: no encryption
	// request, and no success packet taking the client at its word.
	if client.compressionAfter != 0 || len(client.compressionThresholds) != 0 {
		t.Errorf("enabled compression at %v, want a login that has not been finished", client.compressionThresholds)
	}

	if client.authenticateCalls != 0 {
		t.Errorf("asked the session server %d times, want a login the proxy answers for", client.authenticateCalls)
	}
}

// The answer, from a proxy holding the secret. Everything the login is finished
// with comes out of the payload, including the name: the one in login start is
// what the connection claimed, and a server that asked for a signature has no
// reason to prefer it.
func TestHandleLoginPluginResponseServerboundPacketFinishesTheLoginTheProxySigned(t *testing.T) {
	signature := "a signature"
	forwarded := types.ForwardedLogin{
		Address:    "203.0.113.7",
		Uuid:       "069a79f4-44e9-4726-a5be-fca90e38aaf5",
		Username:   "Notch",
		Properties: []types.ProfileProperty{{Name: "textures", Value: "a base64 blob", Signature: &signature}},
	}

	client := &fakeClient{
		phase:            types.PhaseLogin,
		profile:          types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001", Username: "somebody else"},
		forwardingSecret: forwardingSecret,
		forwardedResult:  forwarded,
	}

	payload := []byte("a signed payload")

	if err := HandleLoginPluginResponseServerboundPacket(client, &login.LoginPluginResponseServerboundPacket{MessageId: 1, Successful: true, Data: payload}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !slices.Equal(client.completedForwardingIds, []int32{1}) {
		t.Errorf("checked message ids %v, want the one the answer carried", client.completedForwardingIds)
	}

	if len(client.completedForwardingPayloads) != 1 || !bytes.Equal(client.completedForwardingPayloads[0], payload) {
		t.Errorf("checked payloads %q, want the one the answer carried", client.completedForwardingPayloads)
	}

	if !slices.Equal(client.compressionThresholds, []int32{compressionThreshold}) {
		t.Errorf("enabled compression at %v, want the threshold announced once at %d", client.compressionThresholds, compressionThreshold)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	loginSuccess, ok := client.written[0].(*clientboundLogin.LoginSuccessClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginSuccessClientboundPacket, got %T", client.written[0])
	}

	want := types.GameProfile{Uuid: forwarded.Uuid, Username: forwarded.Username, Properties: forwarded.Properties}

	if loginSuccess.Profile.String() != want.String() {
		t.Errorf("login success profile = %s, want the signed %s", loginSuccess.Profile, want)
	}

	if client.Profile().String() != want.String() {
		t.Errorf("kept profile = %s, want the signed %s", client.Profile(), want)
	}

	if client.authenticateCalls != 0 {
		t.Errorf("asked the session server %d times, want a login the proxy already asked about left alone", client.authenticateCalls)
	}
}

// What a client that came straight here answers with, since it has never heard
// of the channel. The secret says what a forwarded login is worth and nothing
// about this one, so the login carries on to be settled the way this server
// settles a login nobody forwarded: Mojang is asked when the connection is to be
// encrypted.
func TestHandleLoginPluginResponseServerboundPacketAsksMojangWhenNoProxyForwarded(t *testing.T) {
	client := &fakeClient{
		phase:             types.PhaseLogin,
		profile:           types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001", Username: "Notch"},
		forwardingSecret:  forwardingSecret,
		encryptionEnabled: true,
		publicKey:         []byte("a public key"),
		verifyToken:       []byte("a verify token"),
	}

	if err := HandleLoginPluginResponseServerboundPacket(client, &login.LoginPluginResponseServerboundPacket{MessageId: 1, Successful: false}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	request, ok := client.written[0].(*clientboundLogin.EncryptionRequestClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.EncryptionRequestClientboundPacket, got %T", client.written[0])
	}

	// A login this end is checking is one this end has to be told the answer
	// about, which is what asking the client to authenticate is for.
	if !request.ShouldAuthenticate {
		t.Error("asked for a secret without asking Mojang, want a login checked with the session server")
	}

	// The request has been answered, even though it carried nothing, so a payload
	// arriving behind this cannot settle the login a second time.
	if !slices.Equal(client.declinedForwardingIds, []int32{1}) {
		t.Errorf("gave up on %v, want the request the answer came in on", client.declinedForwardingIds)
	}

	// Nothing is finished yet: the client answers the encryption request, and the
	// session server answers after that.
	if len(client.compressionThresholds) != 0 {
		t.Errorf("enabled compression at %v, want a login still waiting on the secret", client.compressionThresholds)
	}
}

// The same answer on a server that encrypts nothing. There is nobody left to
// ask, so the login is finished on the name the client logged in under, exactly
// as it would be on a server no proxy was ever pointed at.
func TestHandleLoginPluginResponseServerboundPacketTakesTheClientsWordWhenNoProxyForwarded(t *testing.T) {
	client := &fakeClient{
		phase:            types.PhaseLogin,
		profile:          types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001", Username: "Notch"},
		forwardingSecret: forwardingSecret,
	}

	if err := HandleLoginPluginResponseServerboundPacket(client, &login.LoginPluginResponseServerboundPacket{MessageId: 1, Successful: false}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !slices.Equal(client.declinedForwardingIds, []int32{1}) {
		t.Errorf("gave up on %v, want the request the answer came in on", client.declinedForwardingIds)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	loginSuccess, ok := client.written[0].(*clientboundLogin.LoginSuccessClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginSuccessClientboundPacket, got %T", client.written[0])
	}

	// The uuid the name is worth, rather than the one the client sent: a name
	// nobody checked is at least the same account every time it is used.
	if loginSuccess.Profile.Uuid != types.OfflineUuid("Notch") {
		t.Errorf("uuid = %q, want the offline %q", loginSuccess.Profile.Uuid, types.OfflineUuid("Notch"))
	}

	if loginSuccess.Profile.Username != "Notch" {
		t.Errorf("username = %q, want the name the client logged in under", loginSuccess.Profile.Username)
	}

	if len(loginSuccess.Profile.Properties) != 0 {
		t.Errorf("carried %d properties, want none for a profile nobody signed", len(loginSuccess.Profile.Properties))
	}

	if client.authenticateCalls != 0 {
		t.Errorf("asked the session server %d times, want a login with no secret behind it left alone", client.authenticateCalls)
	}
}

// A second answer to the one question. The first was given up on, so there is
// nothing outstanding for this to answer, and the connection says so rather than
// letting a login be settled twice.
//
// It is let go rather than left where it is. Nothing about this connection is
// going to be settled now, and the login phase has no keep alive to notice a
// connection nobody is going to say anything else to.
func TestHandleLoginPluginResponseServerboundPacketReportsASecondAnswerToAGivenUpRequest(t *testing.T) {
	client := &fakeClient{
		phase:                types.PhaseLogin,
		profile:              types.GameProfile{Username: "Notch"},
		forwardingSecret:     forwardingSecret,
		declineForwardingErr: errors.New("no forwarding request is waiting on an answer"),
	}

	err := HandleLoginPluginResponseServerboundPacket(client, &login.LoginPluginResponseServerboundPacket{MessageId: 1, Successful: false})
	if err == nil {
		t.Fatal("error = nil, want the connection's rejection passed back")
	}

	if len(client.written) != 1 {
		t.Fatalf("wrote %d packets, want the connection let go", len(client.written))
	}

	if _, ok := client.written[0].(*clientboundLogin.DisconnectClientboundPacket); !ok {
		t.Fatalf("expected *login.DisconnectClientboundPacket, got %T", client.written[0])
	}

	if len(client.compressionThresholds) != 0 {
		t.Error("finished the login, want an answer to nothing let go")
	}
}

// A payload the connection would not vouch for: the wrong secret, a digest that
// does not match, or an answer to a request that was never sent. They are one
// thing from here, and the login ends the same way.
func TestHandleLoginPluginResponseServerboundPacketRefusesAPayloadTheSecretDoesNotVouchFor(t *testing.T) {
	client := &fakeClient{
		phase:                 types.PhaseLogin,
		forwardingSecret:      forwardingSecret,
		completeForwardingErr: errors.New("the forwarded login is not signed with the forwarding secret"),
	}

	err := HandleLoginPluginResponseServerboundPacket(client, &login.LoginPluginResponseServerboundPacket{MessageId: 1, Successful: true, Data: []byte("a forged payload")})
	if err == nil {
		t.Fatal("error = nil, want a payload the secret does not vouch for refused")
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	if _, ok := client.written[0].(*clientboundLogin.DisconnectClientboundPacket); !ok {
		t.Fatalf("expected *login.DisconnectClientboundPacket, got %T", client.written[0])
	}

	if len(client.compressionThresholds) != 0 {
		t.Error("finished the login, want a payload that could not be vouched for let go")
	}
}

// A server with no secret asks nothing on the channel, so an answer on it
// answers nothing. It is reported rather than acted on: a limbo has no other use
// for the channel and no way to know what the client meant by it.
func TestHandleLoginPluginResponseServerboundPacketReportsAnAnswerToNothing(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin}

	err := HandleLoginPluginResponseServerboundPacket(client, &login.LoginPluginResponseServerboundPacket{MessageId: 1, Successful: true, Data: []byte("a payload")})
	if err == nil {
		t.Fatal("error = nil, want an answer to a question nobody asked reported")
	}

	if len(client.written) != 0 {
		t.Errorf("wrote %d packets, want nothing said back", len(client.written))
	}

	if len(client.completedForwardingPayloads) != 0 {
		t.Error("checked the payload, want it left alone on a server that asked nothing")
	}
}

// A handshake carrying a BungeeCord proxy's fields, on a server that holds a
// secret. Those fields are plain text anyone who can reach the port can write,
// and reading them would put back the account-for-the-asking the secret was
// configured to remove, one packet ahead of the one that has to be signed.
func TestAForwardedAddressCannotStandInForTheSignedLogin(t *testing.T) {
	client := &fakeClient{forwardingSecret: forwardingSecret}

	handshakePacket := &handshake.HandshakeServerboundPacket{
		ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
		ServerAddress:   forwardedAddress,
		Intent:          int32(types.PhaseLogin),
	}

	if err := HandleHandshakeServerboundPacket(client, handshakePacket); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := client.ForwardedLogin(); ok {
		t.Error("kept the account written into the address, want it left to the signature to name")
	}

	if err := HandleLoginStartServerboundPacket(client, &login.LoginStartServerboundPacket{Name: "Notch"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	// The question the proxy answers, rather than the success packet the
	// address was written that way to earn.
	request, ok := client.written[0].(*clientboundLogin.LoginPluginRequestClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginPluginRequestClientboundPacket, got %T", client.written[0])
	}

	// Whoever wrote the address cannot sign for it, so the answer is that the
	// channel means nothing here. The login falls back, and what it falls back to
	// is the client's own name rather than the account in the address: the fields
	// bought nothing on the way past, and the fallback does not go looking for
	// them.
	response := &login.LoginPluginResponseServerboundPacket{MessageId: request.MessageId, Successful: false}

	if err := HandleLoginPluginResponseServerboundPacket(client, response); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.written) != 2 {
		t.Fatalf("expected 2 written packets, got %d", len(client.written))
	}

	loginSuccess, ok := client.written[1].(*clientboundLogin.LoginSuccessClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginSuccessClientboundPacket, got %T", client.written[1])
	}

	if loginSuccess.Profile.Uuid != types.OfflineUuid("Notch") {
		t.Errorf("uuid = %q, want the offline %q rather than the account written into the address", loginSuccess.Profile.Uuid, types.OfflineUuid("Notch"))
	}

	if len(loginSuccess.Profile.Properties) != 0 {
		t.Errorf("carried %d properties, want none of the textures the address named", len(loginSuccess.Profile.Properties))
	}
}

func TestHandleEncryptionResponseServerboundPacketAuthenticatesAndWritesLoginSuccess(t *testing.T) {
	claimed := types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001", Username: "notch"}
	signature := "a signature"
	authenticated := types.GameProfile{
		Uuid:       "069a79f4-44e9-4726-a5be-fca90e38aaf5",
		Username:   "Notch",
		Properties: []types.ProfileProperty{{Name: "textures", Value: "a skin", Signature: &signature}},
	}

	client := &fakeClient{phase: types.PhaseLogin, profile: claimed, authenticated: authenticated}
	packet := &login.EncryptionResponseServerboundPacket{SharedSecret: []byte("an encrypted secret"), VerifyToken: []byte("an encrypted token")}

	if err := HandleEncryptionResponseServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.completedSecrets) != 1 || !bytes.Equal(client.completedSecrets[0], packet.SharedSecret) || !bytes.Equal(client.completedTokens[0], packet.VerifyToken) {
		t.Fatalf("completed encryption with %q and %q, want the two fields the response carried", client.completedSecrets, client.completedTokens)
	}

	// The client turned its own cipher on the moment it sent this, so anything
	// written before this end catches up is a packet it reads as noise.
	if client.encryptedAfter != 0 {
		t.Errorf("encrypted the connection after %d packets, want it encrypted before any reply", client.encryptedAfter)
	}

	if client.authenticateCalls != 1 {
		t.Fatalf("asked the session server %d times, want once", client.authenticateCalls)
	}

	// A login that is only checked after the client has been welcomed in is a
	// login that was not checked.
	if client.authenticateAfter != 0 {
		t.Errorf("authenticated after %d packets, want it settled before any reply", client.authenticateAfter)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	// The threshold has to reach the client before anything framed for it does.
	if !slices.Equal(client.compressionThresholds, []int32{compressionThreshold}) {
		t.Errorf("enabled compression at %v, want the threshold announced once at %d", client.compressionThresholds, compressionThreshold)
	}

	if client.compressionAfter != 0 {
		t.Errorf("enabled compression after %d packets, want it announced before any reply", client.compressionAfter)
	}

	loginSuccess, ok := client.written[0].(*clientboundLogin.LoginSuccessClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginSuccessClientboundPacket, got %T", client.written[0])
	}

	// What the client claimed about itself is worth nothing next to what Mojang
	// answered, down to the case of the name and the textures that give it a
	// skin.
	if loginSuccess.Profile.String() != authenticated.String() {
		t.Errorf("login success profile = %s, want the authenticated %s", loginSuccess.Profile, authenticated)
	}

	if client.Profile().String() != authenticated.String() {
		t.Errorf("kept profile = %s, want the authenticated %s", client.Profile(), authenticated)
	}

	if loginSuccess.SessionId == "" {
		t.Error("expected a generated session id, got an empty string")
	}

	if loginSuccess.SessionId == loginSuccess.Profile.Uuid {
		t.Error("expected the session id to differ from the profile uuid")
	}
}

func TestHandleEncryptionResponseServerboundPacketDisconnectsAClientMojangDoesNotVouchFor(t *testing.T) {
	client := &fakeClient{
		phase:           types.PhaseLogin,
		profile:         types.GameProfile{Username: "Notch"},
		authenticateErr: errors.New("no record of this login"),
	}

	err := HandleEncryptionResponseServerboundPacket(client, &login.EncryptionResponseServerboundPacket{})
	if err == nil {
		t.Fatal("error = nil, want the failed authentication reported")
	}

	// The client is sitting on a connection it has no reason to think went
	// wrong, so it is told why it is being let go.
	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	if _, ok := client.written[0].(*clientboundLogin.DisconnectClientboundPacket); !ok {
		t.Fatalf("expected *login.DisconnectClientboundPacket, got %T", client.written[0])
	}

	// Nothing about a login that was refused should look like one that worked.
	if len(client.compressionThresholds) != 0 {
		t.Errorf("enabled compression at %v, want a refused login left alone", client.compressionThresholds)
	}
}

func TestHandleEncryptionResponseServerboundPacketStopsWhereEncryptionFails(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin, completeErr: errors.New("the verify token is not the one that was sent")}

	if err := HandleEncryptionResponseServerboundPacket(client, &login.EncryptionResponseServerboundPacket{}); err == nil {
		t.Error("error = nil, want the failure to encrypt passed back")
	}

	// The client is reading through a cipher this end could not turn on, so
	// there is nothing worth writing to it, and no login worth asking about.
	if len(client.written) != 0 {
		t.Errorf("wrote %d packets, want none once the connection cannot be encrypted", len(client.written))
	}

	if client.authenticateCalls != 0 {
		t.Errorf("asked the session server %d times, want a login abandoned before it is asked about", client.authenticateCalls)
	}
}

func TestHandleLoginAcknowledgedServerboundPacketFinishesConfiguration(t *testing.T) {
	registries := []types.ClientboundPacket{
		clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:dimension_type", []byte{0x01}),
		clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:worldgen/biome", []byte{0x02}),
	}
	client := &fakeClient{phase: types.PhaseLogin, registryPackets: registries}

	if err := HandleLoginAcknowledgedServerboundPacket(client, &login.LoginAcknowledgedServerboundPacket{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Phase() != types.PhaseConfiguration {
		t.Errorf("expected phase %d, got %d", types.PhaseConfiguration, client.Phase())
	}

	if len(client.written) != len(registries)+1 {
		t.Fatalf("expected %d written packets, got %d", len(registries)+1, len(client.written))
	}

	// The client cannot resolve anything the play phase refers to by id until
	// the registries arrive, so they have to precede finish configuration.
	for i, want := range registries {
		if client.written[i] != want {
			t.Errorf("packet %d = %v, want %v", i, client.written[i], want)
		}
	}

	last := client.written[len(client.written)-1]
	if _, ok := last.(*clientboundConfiguration.FinishConfigurationClientboundPacket); !ok {
		t.Errorf("expected *configuration.FinishConfigurationClientboundPacket, got %T", last)
	}

	// Writing before the phase moves would resolve the packet ids in the login
	// phase, where neither packet is registered.
	for i, phase := range client.writePhases {
		if phase != types.PhaseConfiguration {
			t.Errorf("expected packet %d to be written in phase %d, got %d", i, types.PhaseConfiguration, phase)
		}
	}
}

func TestHandleAcknowledgeFinishConfigurationServerboundPacketEntersPlay(t *testing.T) {
	profile := types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001", Username: "Notch"}
	client := &fakeClient{phase: types.PhaseConfiguration, profile: profile, encryptionEnabled: true}

	if err := HandleAcknowledgeFinishConfigurationServerboundPacket(client, &configuration.AcknowledgeFinishConfigurationServerboundPacket{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Phase() != types.PhasePlay {
		t.Errorf("expected phase %d, got %d", types.PhasePlay, client.Phase())
	}

	if len(client.written) != 4 {
		t.Fatalf("expected 4 written packets, got %d", len(client.written))
	}

	// Writing before the phase moves would resolve the packet ids in the
	// configuration phase, where none of these are registered.
	for i, phase := range client.writePhases {
		if phase != types.PhasePlay {
			t.Errorf("expected packet %d to be written in phase %d, got %d", i, types.PhasePlay, phase)
		}
	}

	login, ok := client.written[0].(*clientboundPlay.LoginClientboundPacket)
	if !ok {
		t.Fatalf("expected *play.LoginClientboundPacket first, got %T", client.written[0])
	}

	// A dimension the login packet does not list is one the client refuses to
	// be spawned in.
	if !slices.Contains(login.Dimensions, login.SpawnInfo.Dimension) {
		t.Errorf("spawn dimension %q is not among the listed dimensions %v", login.SpawnInfo.Dimension, login.Dimensions)
	}

	// Any other mode leaves the client waiting for the chunk it stands in,
	// which a limbo that sends no chunks never provides.
	if login.SpawnInfo.GameMode != clientboundPlay.GameModeSpectator {
		t.Errorf("game mode = %s, want spectator", login.SpawnInfo.GameMode)
	}

	if login.SpawnInfo.PreviousGameMode != clientboundPlay.GameModeNone {
		t.Errorf("previous game mode = %s, want none", login.SpawnInfo.PreviousGameMode)
	}

	// This login is one Mojang vouched for, and a client told otherwise draws no
	// head beside any name in the player list.
	if !login.OnlineMode {
		t.Error("online mode = false, want the client told its login was authenticated")
	}

	position, ok := client.written[1].(*clientboundPlay.PlayerPositionClientboundPacket)
	if !ok {
		t.Fatalf("expected *play.PlayerPositionClientboundPacket second, got %T", client.written[1])
	}

	// The client answers with this id, so a teleport it can only report as zero
	// cannot be told from one that was never sent.
	if position.TeleportId == 0 {
		t.Error("expected a non-zero teleport id")
	}

	playerInfo, ok := client.written[2].(*clientboundPlay.PlayerInfoUpdateClientboundPacket)
	if !ok {
		t.Fatalf("expected *play.PlayerInfoUpdateClientboundPacket third, got %T", client.written[2])
	}

	if len(playerInfo.Entries) != 1 {
		t.Fatalf("expected 1 player list entry, got %d", len(playerInfo.Entries))
	}

	entry := playerInfo.Entries[0]

	// The entry is the client's own, so it has to carry the profile the client
	// logged in with rather than a fresh one.
	if entry.Profile.String() != client.Profile().String() {
		t.Errorf("player list entry profile = %s, want %s", entry.Profile, client.Profile())
	}

	// An entry the client is never told about is one it ignores every later
	// update for.
	if playerInfo.Actions&clientboundPlay.PlayerInfoAddPlayer == 0 {
		t.Errorf("actions = %s, want the entry to be added", playerInfo.Actions)
	}

	if playerInfo.Actions&clientboundPlay.PlayerInfoUpdateListed == 0 || !entry.Listed {
		t.Errorf("actions = %s listed = %t, want the player listed", playerInfo.Actions, entry.Listed)
	}

	// The client reads its own mode from both packets, so disagreeing would
	// leave it in one mode holding a list entry that says another.
	if playerInfo.Actions&clientboundPlay.PlayerInfoUpdateGameMode == 0 || entry.GameMode != login.SpawnInfo.GameMode {
		t.Errorf("player list game mode = %s, want the login packet's %s", entry.GameMode, login.SpawnInfo.GameMode)
	}

	// A new entry holds no hat, so an entry that does not say otherwise draws
	// the head as its base skin layer alone.
	if playerInfo.Actions&clientboundPlay.PlayerInfoUpdateHat == 0 || !entry.ShowHat {
		t.Errorf("actions = %s show hat = %t, want the hat shown", playerInfo.Actions, entry.ShowHat)
	}

	// The client sits on its loading screen until this arrives, so it has to be
	// the packet that ends the join rather than one in the middle of it.
	chunksNext, ok := client.written[3].(*clientboundPlay.GameEventClientboundPacket)
	if !ok {
		t.Fatalf("expected *play.GameEventClientboundPacket last, got %T", client.written[3])
	}

	if chunksNext.Event != clientboundPlay.GameEventStartWaitingForChunks {
		t.Errorf("game event = %s, want start_waiting_for_chunks", chunksNext.Event)
	}
}

// TestHandleAcknowledgeFinishConfigurationServerboundPacketTellsTheClientWhatItsLoginWasWorth
// pins the online mode flag to the setting the server runs on. It is the claim
// that a name here was checked with Mojang, and a server that claims it without
// asking is a server telling the player list to draw heads it cannot stand
// behind.
func TestHandleAcknowledgeFinishConfigurationServerboundPacketTellsTheClientWhatItsLoginWasWorth(t *testing.T) {
	tests := []struct {
		name       string
		encryption bool
		forwarded  bool
		want       bool
	}{
		{name: "encrypted", encryption: true, want: true},
		// A proxy asked Mojang on the connection it holds with the player, and
		// forwarded the signed textures it got back, so the heads the player
		// list draws are ones this server can stand behind.
		{name: "forwarded", forwarded: true, want: true},
		{name: "neither", want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &fakeClient{phase: types.PhaseConfiguration, encryptionEnabled: test.encryption, forwarded: test.forwarded}

			if err := HandleAcknowledgeFinishConfigurationServerboundPacket(client, &configuration.AcknowledgeFinishConfigurationServerboundPacket{}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			login, ok := client.written[0].(*clientboundPlay.LoginClientboundPacket)
			if !ok {
				t.Fatalf("expected *play.LoginClientboundPacket first, got %T", client.written[0])
			}

			if login.OnlineMode != test.want {
				t.Errorf("online mode = %t, want the client told what its login was worth %t", login.OnlineMode, test.want)
			}
		})
	}
}

func TestHandleKeepAliveServerboundPacketConfirmsTheIdItCarries(t *testing.T) {
	// The same packet arrives in both phases that have one, and means the same
	// thing in both.
	for _, phase := range []types.Phase{types.PhaseConfiguration, types.PhasePlay} {
		client := &fakeClient{phase: phase}

		if err := HandleKeepAliveServerboundPacket(client, &common.KeepAliveServerboundPacket{Id: 1234}); err != nil {
			t.Fatalf("phase %d: unexpected error: %v", phase, err)
		}

		if !slices.Equal(client.confirmedKeepAlives, []int64{1234}) {
			t.Errorf("phase %d: confirmed %v, want the id the packet carried", phase, client.confirmedKeepAlives)
		}

		// The server asks and the client answers, so an answer needs no answer.
		if len(client.written) != 0 {
			t.Errorf("phase %d: wrote %d packets, want none", phase, len(client.written))
		}
	}
}

func TestHandleKeepAliveServerboundPacketReportsAnAnswerToNothing(t *testing.T) {
	client := &fakeClient{phase: types.PhasePlay, keepAliveErr: errors.New("answers nothing that was sent")}

	if err := HandleKeepAliveServerboundPacket(client, &common.KeepAliveServerboundPacket{Id: 1}); err == nil {
		t.Error("error = nil, want the client's rejection passed back")
	}
}

func TestHandleStatusRequestServerboundPacketAnswersWithWhatTheServerSays(t *testing.T) {
	status := types.ServerStatus{
		Version:     types.ServerVersion{Name: "26.2", Protocol: types.ProtocolVersions.MINECRAFT_26_2.ID},
		Players:     types.ServerPlayers{Online: 3, Max: 4},
		Description: types.TextComponent{Text: "A void limbo"},
	}

	client := &fakeClient{phase: types.PhaseStatus, serverStatus: status}

	if err := HandleStatusRequestServerboundPacket(client, &serverboundStatus.StatusRequestServerboundPacket{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.written) != 1 {
		t.Fatalf("wrote %d packets, want the status response alone", len(client.written))
	}

	response, ok := client.written[0].(*clientboundStatus.StatusResponseClientboundPacket)
	if !ok {
		t.Fatalf("wrote %T, want a status response", client.written[0])
	}

	// Nothing about it is the handler's to decide, so it is the connection's own
	// status or it is wrong.
	if response.Status != status {
		t.Errorf("answered with %+v, want %+v", response.Status, status)
	}

	// A ping is answered where it was asked. The response is registered in the
	// status phase alone, so a client that had moved on would be writing a packet
	// with no id.
	if client.writePhases[0] != types.PhaseStatus {
		t.Errorf("answered in phase %d, want the status phase %d", client.writePhases[0], types.PhaseStatus)
	}
}

func TestHandlePingRequestServerboundPacketSendsThePayloadBack(t *testing.T) {
	// Whatever the client picked, including a number that says nothing about a
	// clock: the client is the only end that knows what it means.
	for _, payload := range []int64{0, 1, -1, 0x0000019870A5E1D3} {
		client := &fakeClient{phase: types.PhaseStatus}

		if err := HandlePingRequestServerboundPacket(client, &serverboundStatus.PingRequestServerboundPacket{Payload: payload}); err != nil {
			t.Fatalf("payload %d: unexpected error: %v", payload, err)
		}

		if len(client.written) != 1 {
			t.Fatalf("payload %d: wrote %d packets, want the pong alone", payload, len(client.written))
		}

		pong, ok := client.written[0].(*clientboundStatus.PongResponseClientboundPacket)
		if !ok {
			t.Fatalf("payload %d: wrote %T, want a pong response", payload, client.written[0])
		}

		if pong.Payload != payload {
			t.Errorf("answered %d, want the %d that was asked", pong.Payload, payload)
		}
	}
}

func TestHandlersRejectUnexpectedPacketType(t *testing.T) {
	client := &fakeClient{}

	if err := HandleHandshakeServerboundPacket(client, &login.LoginStartServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleLoginStartServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleLoginAcknowledgedServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleAcknowledgeFinishConfigurationServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleKeepAliveServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleEncryptionResponseServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleLoginPluginResponseServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleStatusRequestServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandlePingRequestServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}
}
