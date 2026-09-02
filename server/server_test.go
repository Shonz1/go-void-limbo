package server

import (
	"bytes"
	"crypto/sha256"
	"crypto/x509"
	"encoding/json"
	"errors"
	"io"
	"net"
	"slices"
	"testing"
	"time"

	"github.com/Shonz1/go-void-limbo/auth"
	"github.com/Shonz1/go-void-limbo/client"
	"github.com/Shonz1/go-void-limbo/internal/testutil"
	"github.com/Shonz1/go-void-limbo/protocol"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// A login end to end: the handshake and login start a client sends in the clear,
// the encryption request it answers, and then the two packets that finish the
// login, which reach it only if the cipher, the packet ids and the compressed
// framing all line up with what it turned on and when.
func TestALoginIsEncryptedAndAuthenticated(t *testing.T) {
	signature := "a signature"
	authenticated := types.GameProfile{
		Uuid:       "069a79f4-44e9-4726-a5be-fca90e38aaf5",
		Username:   "Notch",
		Properties: []types.ProfileProperty{{Name: "textures", Value: "a base64 blob", Signature: &signature}},
	}

	sessionServer := &testutil.FakeSessionServer{Profile: authenticated}

	srv := &Server{packetRegistry: protocol.NewDefaultRegistry(), keyPair: testutil.KeyPair(), sessionServer: sessionServer, encryptionEnabled: true}

	conn, clientConn := net.Pipe()
	defer clientConn.Close()

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go srv.handleConnection(conn)

	peer := &testutil.LoginPeer{T: t, Conn: clientConn}

	sendHandshake(t, peer, "localhost", types.ProtocolVersions.MINECRAFT_26_2.ID, int32(types.PhaseLogin))
	sendLoginStart(t, peer, "notch")

	request := peer.ReadPacket()

	packetId, err := request.ReadVarInt()
	if err != nil {
		t.Fatalf("reading the encryption request: %v", err)
	}

	if packetId != 0x01 {
		t.Fatalf("packet id = %#02x, want the login phase's encryption request %#02x", packetId, 0x01)
	}

	if _, err := request.ReadString(); err != nil {
		t.Fatalf("reading the server id: %v", err)
	}

	publicKey, err := request.ReadByteArray(1024)
	if err != nil {
		t.Fatalf("reading the public key: %v", err)
	}

	verifyToken, err := request.ReadByteArray(1024)
	if err != nil {
		t.Fatalf("reading the verify token: %v", err)
	}

	if _, err := x509.ParsePKIXPublicKey(publicKey); err != nil {
		t.Fatalf("the public key is not readable as the client reads it: %v", err)
	}

	secret := []byte("0123456789abcdef")

	sendEncryptionResponse(t, peer, secret, verifyToken)

	// From here the client reads and writes through the cipher, whatever the
	// server has managed.
	peer.Encrypt(secret)

	threshold := readSetCompression(t, peer)

	loginSuccess := peer.ReadPacket()

	if packetId, err := loginSuccess.ReadVarInt(); err != nil || packetId != 0x02 {
		t.Fatalf("packet id = %#02x err = %v, want the login phase's login success %#02x", packetId, err, 0x02)
	}

	uuid, err := loginSuccess.ReadUuid()
	if err != nil {
		t.Fatalf("reading the uuid: %v", err)
	}

	// The account Mojang answered with, rather than the one the client claimed.
	if uuid != authenticated.Uuid {
		t.Errorf("uuid = %q, want the authenticated %q", uuid, authenticated.Uuid)
	}

	username, err := loginSuccess.ReadString()
	if err != nil {
		t.Fatalf("reading the username: %v", err)
	}

	if username != authenticated.Username {
		t.Errorf("username = %q, want the authenticated %q", username, authenticated.Username)
	}

	properties, err := loginSuccess.ReadVarInt()
	if err != nil {
		t.Fatalf("reading the property count: %v", err)
	}

	// The textures are the only reason the profile is worth authenticating for
	// beyond the name, since they are what everyone else sees.
	if properties != int32(len(authenticated.Properties)) {
		t.Fatalf("carried %d properties, want the %d Mojang answered with", properties, len(authenticated.Properties))
	}

	name, err := loginSuccess.ReadString()
	if err != nil {
		t.Fatalf("reading the property name: %v", err)
	}

	if name != "textures" {
		t.Errorf("property name = %q, want %q", name, "textures")
	}

	// The client was asked about under the name it logged in with, which is the
	// only name anything knew before Mojang answered.
	if !slices.Equal(sessionServer.Usernames, []string{"notch"}) {
		t.Errorf("asked about %v, want the name the client logged in under", sessionServer.Usernames)
	}

	// The vanilla threshold, which is what the login announces.
	if threshold != 256 {
		t.Errorf("threshold = %d, want %d", threshold, 256)
	}
}

// sendHandshake sends the handshake that opens every exchange: the version the
// client speaks, the address it reached the server at, and the phase it is
// asking to be put into. The intent is the raw number rather than a phase,
// since what a handshake may ask for is itself under test in places.
func sendHandshake(t *testing.T, peer *testutil.LoginPeer, serverAddress string, protocolId types.ProtocolId, intent int32) {
	t.Helper()

	handshake := new(bytes.Buffer)
	handshakeStream := streams.NewMinecraftStreamFromBuffer(handshake)

	for _, write := range []func() error{
		func() error { return handshakeStream.WriteVarInt(0x00) },
		func() error { return handshakeStream.WriteVarInt(int32(protocolId)) },
		func() error { return handshakeStream.WriteString(serverAddress) },
		func() error { return handshakeStream.WriteShort(25565) },
		func() error { return handshakeStream.WriteVarInt(intent) },
		handshakeStream.Flush,
	} {
		if err := write(); err != nil {
			t.Fatalf("building the handshake: %v", err)
		}
	}

	peer.WritePacket(handshake.Bytes())
}

// sendLoginStart sends the name a client claims for itself, with a uuid it
// picked on its own, which is never the one the login is finished with.
func sendLoginStart(t *testing.T, peer *testutil.LoginPeer, username string) {
	t.Helper()

	loginStart := new(bytes.Buffer)
	loginStartStream := streams.NewMinecraftStreamFromBuffer(loginStart)

	for _, write := range []func() error{
		func() error { return loginStartStream.WriteVarInt(0x00) },
		func() error { return loginStartStream.WriteString(username) },
		func() error { return loginStartStream.WriteUuid("00000000-0000-0000-0000-000000000001") },
		loginStartStream.Flush,
	} {
		if err := write(); err != nil {
			t.Fatalf("building the login start: %v", err)
		}
	}

	peer.WritePacket(loginStart.Bytes())
}

// sendEncryptionResponse answers an encryption request the way a client does:
// the secret it chose and the token it was sent, each under the server's key.
func sendEncryptionResponse(t *testing.T, peer *testutil.LoginPeer, secret, verifyToken []byte) {
	t.Helper()

	response := new(bytes.Buffer)
	responseStream := streams.NewMinecraftStreamFromBuffer(response)

	for _, write := range []func() error{
		func() error { return responseStream.WriteVarInt(0x01) },
		func() error { return responseStream.WriteByteArray(testutil.EncryptForServer(t, secret)) },
		func() error { return responseStream.WriteByteArray(testutil.EncryptForServer(t, verifyToken)) },
		responseStream.Flush,
	} {
		if err := write(); err != nil {
			t.Fatalf("building the encryption response: %v", err)
		}
	}

	peer.WritePacket(response.Bytes())
}

// readSetCompression reads the set compression packet that opens the end of
// every login here, returns the threshold it announced, and puts the peer on
// the compressed framing, as a client is from that packet on.
func readSetCompression(t *testing.T, peer *testutil.LoginPeer) int32 {
	t.Helper()

	setCompression := peer.ReadPacket()

	if packetId, err := setCompression.ReadVarInt(); err != nil || packetId != 0x03 {
		t.Fatalf("packet id = %#02x err = %v, want the login phase's set compression %#02x", packetId, err, 0x03)
	}

	threshold, err := setCompression.ReadVarInt()
	if err != nil {
		t.Fatalf("reading the threshold: %v", err)
	}

	peer.Compressed = true

	return threshold
}

// readLoginSuccess reads the login success packet and returns the uuid, the
// username and the property count it carried, which are what every test about
// how a login ends checks.
func readLoginSuccess(t *testing.T, peer *testutil.LoginPeer) (string, string, int32) {
	t.Helper()

	loginSuccess := peer.ReadPacket()

	if packetId, err := loginSuccess.ReadVarInt(); err != nil || packetId != 0x02 {
		t.Fatalf("packet id = %#02x err = %v, want the login phase's login success %#02x", packetId, err, 0x02)
	}

	uuid, err := loginSuccess.ReadUuid()
	if err != nil {
		t.Fatalf("reading the uuid: %v", err)
	}

	username, err := loginSuccess.ReadString()
	if err != nil {
		t.Fatalf("reading the username: %v", err)
	}

	properties, err := loginSuccess.ReadVarInt()
	if err != nil {
		t.Fatalf("reading the property count: %v", err)
	}

	return uuid, username, properties
}

// The same login on a server that was told not to encrypt: no encryption
// request goes out, Mojang is never asked, and the two packets that finish the
// login reach a client that never turned a cipher on.
func TestALoginWithoutEncryptionIsTakenAtTheClientsWord(t *testing.T) {
	sessionServer := &testutil.FakeSessionServer{}

	srv := &Server{packetRegistry: protocol.NewDefaultRegistry(), keyPair: testutil.KeyPair(), sessionServer: sessionServer, encryptionEnabled: false}

	conn, clientConn := net.Pipe()
	defer clientConn.Close()

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go srv.handleConnection(conn)

	peer := &testutil.LoginPeer{T: t, Conn: clientConn}

	sendHandshake(t, peer, "localhost", types.ProtocolVersions.MINECRAFT_26_2.ID, int32(types.PhaseLogin))
	sendLoginStart(t, peer, "notch")

	// Set compression rather than an encryption request: the login went
	// straight to its end, and everything is still in the clear.
	readSetCompression(t, peer)

	uuid, username, properties := readLoginSuccess(t, peer)

	if uuid != types.OfflineUuid("notch") {
		t.Errorf("uuid = %q, want the offline %q", uuid, types.OfflineUuid("notch"))
	}

	if username != "notch" {
		t.Errorf("username = %q, want the name the client logged in under %q", username, "notch")
	}

	// No session server answered, so there are no signed textures to carry, and
	// the client shows the default skin.
	if properties != 0 {
		t.Errorf("carried %d properties, want none for a profile nobody signed", properties)
	}

	// A login with no secret behind it is one the session server could not have
	// been asked about.
	if len(sessionServer.Usernames) != 0 {
		t.Errorf("asked the session server about %v, want a login nobody can vouch for left alone", sessionServer.Usernames)
	}
}

// forwardedHandshakeAddress is what a proxy puts in the handshake in place of
// the address it was reached at: the player's own address, the account it
// authenticated, and the textures the session server gave it for that account.
const forwardedHandshakeAddress = "limbo.example\x00203.0.113.7\x00069a79f444e94726a5befca90e38aaf5\x00" +
	`[{"name":"textures","value":"a base64 blob","signature":"a signature"}]`

// The same login as the one taken at the client's word, on the same server,
// with a proxy in front of it: nothing is configured differently, and the
// handshake is the whole of the difference. It carries the account, the login
// start carries the name, and the two packets that finish the login welcome the
// player as who the proxy said they were, without a word to Mojang.
func TestALoginBehindAProxyIsTheAccountTheProxyForwarded(t *testing.T) {
	sessionServer := &testutil.FakeSessionServer{}

	srv := &Server{packetRegistry: protocol.NewDefaultRegistry(), keyPair: testutil.KeyPair(), sessionServer: sessionServer, encryptionEnabled: false}

	conn, clientConn := net.Pipe()
	defer clientConn.Close()

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go srv.handleConnection(conn)

	peer := &testutil.LoginPeer{T: t, Conn: clientConn}

	sendHandshake(t, peer, forwardedHandshakeAddress, types.ProtocolVersions.MINECRAFT_26_2.ID, int32(types.PhaseLogin))
	sendLoginStart(t, peer, "Notch")

	// Set compression rather than an encryption request: the account was settled
	// by the handshake, so the login went straight to its end.
	readSetCompression(t, peer)

	uuid, username, properties := readLoginSuccess(t, peer)

	if uuid != "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Errorf("uuid = %q, want the account the proxy forwarded", uuid)
	}

	// The handshake carries no name, so this is the one place it comes from.
	if username != "Notch" {
		t.Errorf("username = %q, want the name the login start carried %q", username, "Notch")
	}

	// The textures the proxy forwarded are signed by Mojang, and are what every
	// client draws the skin from.
	if properties != 1 {
		t.Fatalf("carried %d properties, want the textures the proxy forwarded", properties)
	}

	// The proxy asked Mojang on the connection it holds with the player, and
	// this end has no secret to ask about even if it wanted to.
	if len(sessionServer.Usernames) != 0 {
		t.Errorf("asked the session server about %v, want a login the proxy already settled left alone", sessionServer.Usernames)
	}
}

// forwardingServer is a limbo with a proxy configured in front of it: a secret
// to check a forwarded login against, and no encryption, since the proxy holds
// the connection with the player and answered for them already. A login it does
// not sign for is left to the client's own word here, which is what an operator
// who turned encryption off asked for.
func forwardingServer(t *testing.T) (*Server, *testutil.LoginPeer) {
	t.Helper()

	return forwardingServerOn(t, false, &testutil.FakeSessionServer{})
}

// forwardingServerOn is forwardingServer for the tests about what a login the
// proxy did not sign for is worth on it, which is the one thing holding a secret
// does not decide: whether the connection is to be encrypted, and what the
// session server would then say about it.
func forwardingServerOn(t *testing.T, encryptionEnabled bool, sessionServer client.SessionServer) (*Server, *testutil.LoginPeer) {
	t.Helper()

	srv := &Server{
		packetRegistry:    protocol.NewDefaultRegistry(),
		keyPair:           testutil.KeyPair(),
		sessionServer:     sessionServer,
		encryptionEnabled: encryptionEnabled,
		forwardingSecret:  testutil.ForwardingSecret,
	}

	conn, clientConn := net.Pipe()
	t.Cleanup(func() { clientConn.Close() })

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go srv.handleConnection(conn)

	return srv, &testutil.LoginPeer{T: t, Conn: clientConn}
}

// openForwardedLogin sends the handshake and login start a proxy sends, and
// reads the question the server asks back: which channel it wants the login on,
// and under which message id.
func openForwardedLogin(t *testing.T, peer *testutil.LoginPeer, serverAddress, username string) (int32, string) {
	t.Helper()

	sendHandshake(t, peer, serverAddress, types.ProtocolVersions.MINECRAFT_26_2.ID, int32(types.PhaseLogin))
	sendLoginStart(t, peer, username)

	request := peer.ReadPacket()

	if packetId, err := request.ReadVarInt(); err != nil || packetId != 0x04 {
		t.Fatalf("packet id = %#02x err = %v, want the login phase's plugin request %#02x", packetId, err, 0x04)
	}

	messageId, err := request.ReadVarInt()
	if err != nil {
		t.Fatalf("reading the message id: %v", err)
	}

	channel, err := request.ReadString()
	if err != nil {
		t.Fatalf("reading the channel: %v", err)
	}

	return messageId, channel
}

// answerForwardingRequest sends what a proxy answers a forwarding request with,
// under the message id it came in on.
func answerForwardingRequest(t *testing.T, peer *testutil.LoginPeer, messageId int32, payload []byte) {
	t.Helper()

	response := new(bytes.Buffer)
	responseStream := streams.NewMinecraftStreamFromBuffer(response)

	writes := []func() error{
		func() error { return responseStream.WriteVarInt(0x02) },
		func() error { return responseStream.WriteVarInt(messageId) },
		func() error { return responseStream.WriteBoolean(payload != nil) },
	}

	if payload != nil {
		writes = append(writes, func() error { return responseStream.WriteBytes(payload) })
	}

	writes = append(writes, responseStream.Flush)

	for _, write := range writes {
		if err := write(); err != nil {
			t.Fatalf("building the plugin response: %v", err)
		}
	}

	peer.WritePacket(response.Bytes())
}

// A login through a modern proxy, end to end. Nothing about the connection says
// who the player is: the handshake carries the address it was reached at, the
// login start carries a name nobody checked, and the account arrives afterwards
// in a payload signed with the secret the proxy and this server share.
func TestALoginBehindAModernProxyIsTheAccountItSigned(t *testing.T) {
	signature := "a signature"
	properties := []types.ProfileProperty{{Name: "textures", Value: "a base64 blob", Signature: &signature}}

	_, peer := forwardingServer(t)

	messageId, channel := openForwardedLogin(t, peer, "limbo.example", "somebody else")

	if channel != auth.ModernForwardingChannel {
		t.Fatalf("asked on %q, want the channel a proxy forwards a login on %q", channel, auth.ModernForwardingChannel)
	}

	payload := testutil.SignedForwardingPayload(t, testutil.ForwardingSecret, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", properties)

	answerForwardingRequest(t, peer, messageId, payload)

	// Set compression rather than an encryption request: the payload settled
	// the account, so the login went straight to its end.
	readSetCompression(t, peer)

	uuid, username, count := readLoginSuccess(t, peer)

	if uuid != "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Errorf("uuid = %q, want the account the proxy signed for", uuid)
	}

	// The name comes out of the payload as well, and not out of the login start,
	// which said something else entirely.
	if username != "Notch" {
		t.Errorf("username = %q, want the name the payload carried %q", username, "Notch")
	}

	if count != 1 {
		t.Fatalf("carried %d properties, want the textures the proxy forwarded", count)
	}
}

// The same server, answered by something that produced a login without holding
// the secret. This is the answer a client is never in a position to give: it
// names an account under the proxy's authority, and the signature is the whole of
// that authority, so every way of getting it wrong ends the same way.
//
// A connection that answers with no login at all is a different thing and is not
// refused. It is a player who came to the port directly, and is settled by what
// the server would do with any login nobody forwarded.
func TestALoginTheForwardingSecretDoesNotVouchForIsRefused(t *testing.T) {
	tests := []struct {
		name string

		// payload is what the connection answers with.
		payload func(t *testing.T) []byte
	}{
		{
			name: "a login signed under a guess at the secret",
			payload: func(t *testing.T) []byte {
				return testutil.SignedForwardingPayload(t, []byte("a guess"), "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)
			},
		},
		{
			name: "a login with no signature at all",
			payload: func(t *testing.T) []byte {
				signed := testutil.SignedForwardingPayload(t, testutil.ForwardingSecret, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)
				return signed[sha256.Size:]
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, peer := forwardingServer(t)

			messageId, _ := openForwardedLogin(t, peer, "limbo.example", "Notch")

			answerForwardingRequest(t, peer, messageId, test.payload(t))

			disconnect := peer.ReadPacket()

			if packetId, err := disconnect.ReadVarInt(); err != nil || packetId != 0x00 {
				t.Fatalf("packet id = %#02x err = %v, want the login phase's disconnect %#02x", packetId, err, 0x00)
			}

			// A client sitting on a connection it has no reason to think went
			// wrong is told why it is being let go.
			reason, err := disconnect.ReadString()
			if err != nil {
				t.Fatalf("reading the reason: %v", err)
			}

			if reason == "" {
				t.Error("disconnected without a word, want a reason the player can read")
			}
		})
	}
}

// A player who came to the port itself, on a server a proxy also points at. The
// connection has never heard of the forwarding channel and says so, and that is
// not a refusal: the login carries on to be settled the way this server settles
// any login nobody forwarded, which here means asking Mojang behind a cipher of
// its own.
func TestALoginNoProxyForwardedIsCheckedWithMojangHere(t *testing.T) {
	signature := "a signature"
	authenticated := types.GameProfile{
		Uuid:       "069a79f4-44e9-4726-a5be-fca90e38aaf5",
		Username:   "Notch",
		Properties: []types.ProfileProperty{{Name: "textures", Value: "a base64 blob", Signature: &signature}},
	}

	sessionServer := &testutil.FakeSessionServer{Profile: authenticated}

	_, peer := forwardingServerOn(t, true, sessionServer)

	messageId, _ := openForwardedLogin(t, peer, "limbo.example", "notch")

	// The answer of a client that has never heard of the channel.
	answerForwardingRequest(t, peer, messageId, nil)

	// The question this server asks a login it has to settle itself, which it
	// only gets to once the proxy has had its chance.
	request := peer.ReadPacket()

	if packetId, err := request.ReadVarInt(); err != nil || packetId != 0x01 {
		t.Fatalf("packet id = %#02x err = %v, want the login phase's encryption request %#02x", packetId, err, 0x01)
	}

	if _, err := request.ReadString(); err != nil {
		t.Fatalf("reading the server id: %v", err)
	}

	publicKey, err := request.ReadByteArray(1024)
	if err != nil {
		t.Fatalf("reading the public key: %v", err)
	}

	verifyToken, err := request.ReadByteArray(1024)
	if err != nil {
		t.Fatalf("reading the verify token: %v", err)
	}

	if _, err := x509.ParsePKIXPublicKey(publicKey); err != nil {
		t.Fatalf("the public key is not readable as the client reads it: %v", err)
	}

	secret := []byte("0123456789abcdef")

	sendEncryptionResponse(t, peer, secret, verifyToken)

	peer.Encrypt(secret)

	readSetCompression(t, peer)

	uuid, _, _ := readLoginSuccess(t, peer)

	// The account Mojang answered with. Nothing about the login came from the
	// proxy, and nothing about it came from the client either.
	if uuid != authenticated.Uuid {
		t.Errorf("uuid = %q, want the authenticated %q", uuid, authenticated.Uuid)
	}

	if !slices.Equal(sessionServer.Usernames, []string{"notch"}) {
		t.Errorf("asked about %v, want the name the client logged in under", sessionServer.Usernames)
	}
}

// The same player on a server that encrypts nothing, which is how a limbo behind
// a proxy is usually run. There is nobody left to ask, so the login is finished
// on the name the client logged in under, with the uuid that name is worth.
func TestALoginNoProxyForwardedIsTakenAtTheClientsWordWithoutEncryption(t *testing.T) {
	srv, peer := forwardingServer(t)

	messageId, _ := openForwardedLogin(t, peer, "limbo.example", "notch")

	answerForwardingRequest(t, peer, messageId, nil)

	readSetCompression(t, peer)

	uuid, username, properties := readLoginSuccess(t, peer)

	if uuid != types.OfflineUuid("notch") {
		t.Errorf("uuid = %q, want the offline %q", uuid, types.OfflineUuid("notch"))
	}

	if username != "notch" {
		t.Errorf("username = %q, want the name the client logged in under %q", username, "notch")
	}

	// Nobody signed for anything, so there are no textures to carry and the
	// client shows the default skin.
	if properties != 0 {
		t.Errorf("carried %d properties, want none for a profile nobody signed", properties)
	}

	// A login nobody vouched for is not one Mojang was asked about: the
	// connection has no secret on it to ask under.
	if asked := srv.sessionServer.(*testutil.FakeSessionServer).Usernames; len(asked) != 0 {
		t.Errorf("asked the session server about %v, want a login nobody can vouch for left alone", asked)
	}
}

// A BungeeCord proxy's fields written into the handshake of a server that holds
// a secret. They are plain text anyone who can reach the port can write, and a
// server that asks for a signature reads none of them: the login is held to the
// question it asked, and the account in the address never becomes anybody.
func TestAForwardedAddressCannotStandInForASignedLogin(t *testing.T) {
	_, peer := forwardingServer(t)

	messageId, _ := openForwardedLogin(t, peer, forwardedHandshakeAddress, "Notch")

	// The question went out rather than a success packet, which is already the
	// address having bought nothing. Answering it properly is what settles the
	// login, and the account that arrives is the signed one.
	payload := testutil.SignedForwardingPayload(t, testutil.ForwardingSecret, "203.0.113.7", "3d5b1b4e-08d9-484c-9ba1-ae3f6c02a1e2", "Notch", nil)

	answerForwardingRequest(t, peer, messageId, payload)

	readSetCompression(t, peer)

	uuid, _, _ := readLoginSuccess(t, peer)

	if uuid != "3d5b1b4e-08d9-484c-9ba1-ae3f6c02a1e2" {
		t.Errorf("uuid = %q, want the account the payload was signed for rather than the one written into the address", uuid)
	}
}

// newStatusServer is a server with nothing behind it but what a ping is answered
// from, which is all the status phase ever reaches for.
func newStatusServer(t *testing.T, description string) *Server {
	t.Helper()

	return &Server{packetRegistry: protocol.NewDefaultRegistry(), status: status{description: description}}
}

// statusServer builds a server that answers pings, opens a connection to it and
// sends the handshake a client sends before one, returning the far side of that
// connection to play the client.
//
// A ping needs nothing else of a server: no key, no session server and no game
// registries, because nothing in the status phase is anybody's login.
// protocolId is what the handshake says the client speaks, which may be a
// version this server has never heard of.
func statusServer(t *testing.T, description string, protocolId types.ProtocolId) *testutil.LoginPeer {
	t.Helper()

	return connect(t, newStatusServer(t, description), protocolId, int32(types.PhaseStatus))
}

// connect opens one connection to srv and sends the handshake that opens it,
// naming the version the client speaks and the phase it is asking to be put
// into.
func connect(t *testing.T, srv *Server, protocolId types.ProtocolId, intent int32) *testutil.LoginPeer {
	t.Helper()

	conn, clientConn := net.Pipe()
	t.Cleanup(func() { clientConn.Close() })

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go srv.handleConnection(conn)

	peer := &testutil.LoginPeer{T: t, Conn: clientConn}

	sendHandshake(t, peer, "limbo.example", protocolId, intent)

	return peer
}

// askStatus sends the status request a client sends and reads the document the
// server answers with, parsed the way a client parses it rather than field by
// field: the packet is one JSON string, and a client that cannot read it is left
// with the defaults for everything it could not find.
func askStatus(t *testing.T, peer *testutil.LoginPeer) types.ServerStatus {
	t.Helper()

	request := new(bytes.Buffer)
	requestStream := streams.NewMinecraftStreamFromBuffer(request)

	// A packet id and nothing behind it.
	for _, write := range []func() error{
		func() error { return requestStream.WriteVarInt(0x00) },
		requestStream.Flush,
	} {
		if err := write(); err != nil {
			t.Fatalf("building the status request: %v", err)
		}
	}

	peer.WritePacket(request.Bytes())

	response := peer.ReadPacket()

	if packetId, err := response.ReadVarInt(); err != nil || packetId != 0x00 {
		t.Fatalf("packet id = %#02x err = %v, want the status phase's status response %#02x", packetId, err, 0x00)
	}

	document, err := response.ReadString()
	if err != nil {
		t.Fatalf("reading the document: %v", err)
	}

	var serverStatus types.ServerStatus
	if err := json.Unmarshal([]byte(document), &serverStatus); err != nil {
		t.Fatalf("parsing %s: %v", document, err)
	}

	return serverStatus
}

// ping sends a ping request and returns the number that came back, which is what
// the client times the round trip of.
func ping(t *testing.T, peer *testutil.LoginPeer, payload int64) int64 {
	t.Helper()

	request := new(bytes.Buffer)
	requestStream := streams.NewMinecraftStreamFromBuffer(request)

	for _, write := range []func() error{
		func() error { return requestStream.WriteVarInt(0x01) },
		func() error { return requestStream.WriteLong(payload) },
		requestStream.Flush,
	} {
		if err := write(); err != nil {
			t.Fatalf("building the ping request: %v", err)
		}
	}

	peer.WritePacket(request.Bytes())

	response := peer.ReadPacket()

	if packetId, err := response.ReadVarInt(); err != nil || packetId != 0x01 {
		t.Fatalf("packet id = %#02x err = %v, want the status phase's pong response %#02x", packetId, err, 0x01)
	}

	pong, err := response.ReadLong()
	if err != nil {
		t.Fatalf("reading the payload: %v", err)
	}

	return pong
}

// A server list ping end to end: the handshake and the two questions a client
// asks before it has decided to connect, and the two answers it draws the entry
// in its server list from.
func TestAPingIsAnsweredWithWhatTheServerSaysAboutItself(t *testing.T) {
	peer := statusServer(t, "A void limbo", types.ProtocolVersions.MINECRAFT_26_2.ID)

	serverStatus := askStatus(t, peer)

	if serverStatus.Description.Text != "A void limbo" {
		t.Errorf("description = %q, want the one the server was started with", serverStatus.Description.Text)
	}

	// The version the client speaks, since this server speaks it too. Anything
	// else here is a client that draws the server as one it cannot join.
	want := types.ServerVersion{Name: "26.2", Protocol: types.ProtocolVersions.MINECRAFT_26_2.ID}
	if serverStatus.Version != want {
		t.Errorf("version = %+v, want %+v", serverStatus.Version, want)
	}

	// Nobody has joined, and a limbo is never full.
	if wantPlayers := (types.ServerPlayers{Online: 0, Max: 1}); serverStatus.Players != wantPlayers {
		t.Errorf("players = %+v, want %+v", serverStatus.Players, wantPlayers)
	}

	// The client times the answer to this and matches it against what it sent,
	// so the only wrong answer is a different number.
	payload := int64(0x0000019870A5E1D3)
	if pong := ping(t, peer, payload); pong != payload {
		t.Errorf("pong = %d, want the %d that was asked", pong, payload)
	}
}

// A client on a version this server cannot be joined on still gets an answer.
// Its handshake left the connection on protocol zero, so nothing in the status
// phase could be resolved at the version it speaks, and the point of answering
// is that its own server list is what says the versions do not match.
func TestAPingFromAVersionThisServerDoesNotSpeakIsStillAnswered(t *testing.T) {
	peer := statusServer(t, "A void limbo", 47)

	serverStatus := askStatus(t, peer)

	want := types.ServerVersion{Name: types.LatestProtocolVersion.Names[0], Protocol: types.LatestProtocolVersion.ID}
	if serverStatus.Version != want {
		t.Errorf("version = %+v, want the latest this server speaks %+v", serverStatus.Version, want)
	}

	if serverStatus.Description.Text != "A void limbo" {
		t.Errorf("description = %q, want the one the server was started with", serverStatus.Description.Text)
	}

	if pong := ping(t, peer, 1); pong != 1 {
		t.Errorf("pong = %d, want the 1 that was asked", pong)
	}
}

func TestStatusVersionIsTheClientsWhenThisServerSpeaksIt(t *testing.T) {
	latest := types.ServerVersion{Name: types.LatestProtocolVersion.Names[0], Protocol: types.LatestProtocolVersion.ID}

	tests := []struct {
		name    string
		version types.ProtocolVersion
		want    types.ServerVersion
	}{
		// Each supported version is reported as itself, which is what makes a
		// client on either of them see a server it can join. The name is the
		// first the version goes by, since a release that shares a protocol with
		// another shares everything a client checks.
		{name: "1.20.3", version: types.ProtocolVersions.MINECRAFT_1_20_3, want: types.ServerVersion{Name: "1.20.3", Protocol: types.ProtocolVersions.MINECRAFT_1_20_3.ID}},
		{name: "1.20.5", version: types.ProtocolVersions.MINECRAFT_1_20_5, want: types.ServerVersion{Name: "1.20.5", Protocol: types.ProtocolVersions.MINECRAFT_1_20_5.ID}},
		{name: "1.21", version: types.ProtocolVersions.MINECRAFT_1_21, want: types.ServerVersion{Name: "1.21", Protocol: types.ProtocolVersions.MINECRAFT_1_21.ID}},
		{name: "1.21.2", version: types.ProtocolVersions.MINECRAFT_1_21_2, want: types.ServerVersion{Name: "1.21.2", Protocol: types.ProtocolVersions.MINECRAFT_1_21_2.ID}},
		{name: "1.21.4", version: types.ProtocolVersions.MINECRAFT_1_21_4, want: types.ServerVersion{Name: "1.21.4", Protocol: types.ProtocolVersions.MINECRAFT_1_21_4.ID}},
		{name: "1.21.5", version: types.ProtocolVersions.MINECRAFT_1_21_5, want: types.ServerVersion{Name: "1.21.5", Protocol: types.ProtocolVersions.MINECRAFT_1_21_5.ID}},
		{name: "1.21.6", version: types.ProtocolVersions.MINECRAFT_1_21_6, want: types.ServerVersion{Name: "1.21.6", Protocol: types.ProtocolVersions.MINECRAFT_1_21_6.ID}},
		{name: "1.21.7", version: types.ProtocolVersions.MINECRAFT_1_21_7, want: types.ServerVersion{Name: "1.21.7", Protocol: types.ProtocolVersions.MINECRAFT_1_21_7.ID}},
		{name: "1.21.9", version: types.ProtocolVersions.MINECRAFT_1_21_9, want: types.ServerVersion{Name: "1.21.9", Protocol: types.ProtocolVersions.MINECRAFT_1_21_9.ID}},
		{name: "1.21.11", version: types.ProtocolVersions.MINECRAFT_1_21_11, want: types.ServerVersion{Name: "1.21.11", Protocol: types.ProtocolVersions.MINECRAFT_1_21_11.ID}},
		{name: "26.1", version: types.ProtocolVersions.MINECRAFT_26_1, want: types.ServerVersion{Name: "26.1", Protocol: types.ProtocolVersions.MINECRAFT_26_1.ID}},
		{name: "26.2", version: types.ProtocolVersions.MINECRAFT_26_2, want: types.ServerVersion{Name: "26.2", Protocol: types.ProtocolVersions.MINECRAFT_26_2.ID}},

		// A version this server does not speak, which is what a handshake it
		// could not place leaves behind, is told the latest instead.
		{name: "protocol zero", version: types.ProtocolVersions.ZERO, want: latest},
		{name: "a version from before any of this", version: types.GetProtocolVersionById(47), want: latest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := statusVersion(test.version); got != test.want {
				t.Errorf("statusVersion(%d) = %+v, want %+v", test.version.ID, got, test.want)
			}
		})
	}
}

// The sweep is what turns an unanswered keep alive into a closed connection,
// and closing is what ends the read loop on the other side of it. The clients
// the sweep visits are the registered ones, so the test registers its client
// the way handleConnection registers every real one.
func TestTheKeepAliveSweepDropsAClientThatNeverAnswers(t *testing.T) {
	srv := &Server{packetRegistry: protocol.NewDefaultRegistry()}

	conn, clientConn := net.Pipe()
	defer clientConn.Close()

	c := client.New(conn, client.Config{PacketRegistry: srv.packetRegistry, Status: &srv.status})
	c.SetProtocolVersion(types.ProtocolVersions.MINECRAFT_26_2)
	c.SetPhase(types.PhasePlay)

	srv.addClient(c)
	defer srv.removeClient(c)

	stop := make(chan struct{})
	defer close(stop)

	// A tick short enough to keep the test quick. What it stands in for is the
	// fifteen seconds a real crowd gets.
	go srv.sweepKeepAlives(stop, 10*time.Millisecond)

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The whole keep alive as the play phase frames it: a length, the packet id
	// and the eight bytes of the id it asks the answer to carry back.
	frame := make([]byte, 10)
	if _, err := io.ReadFull(clientConn, frame); err != nil {
		t.Fatalf("reading the keep alive: %v", err)
	}

	if frame[1] != 0x2C {
		t.Fatalf("packet id = %#02x, want the play phase's keep alive %#02x", frame[1], 0x2C)
	}

	// Nothing is sent back, so the next sweep finds the keep alive unanswered.
	if _, err := clientConn.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Errorf("error = %v, want the connection closed", err)
	}
}

// The count a ping reports is the clients in the play phase, reported through
// the one status every connection shares. The connection's side of the count --
// that only the move into play joins it, and only once -- lives with the client
// package's own tests.
func TestStatusCountsPlayersJoiningAndLeaving(t *testing.T) {
	s := &status{description: "A void limbo"}
	version := types.ProtocolVersions.MINECRAFT_26_2

	if got, want := s.Status(version).Players, (types.ServerPlayers{Online: 0, Max: 1}); got != want {
		t.Errorf("players = %+v, want %+v on a server nobody has joined", got, want)
	}

	s.PlayerJoined()

	if got, want := s.Status(version).Players, (types.ServerPlayers{Online: 1, Max: 2}); got != want {
		t.Errorf("players = %+v, want %+v once a client is in play", got, want)
	}

	s.PlayerLeft()

	if got, want := s.Status(version).Players, (types.ServerPlayers{Online: 0, Max: 1}); got != want {
		t.Errorf("players = %+v, want %+v once the connection has ended", got, want)
	}
}

// A handshake names the phase it wants to be put into, and the phases are
// numbered, so play is one of the numbers a client could name. A connection that
// names it has logged in to nothing and is not a player, and the ping that
// follows on another connection is where that shows.
func TestAHandshakeCannotPutAConnectionStraightIntoPlay(t *testing.T) {
	srv := newStatusServer(t, "A void limbo")

	// The phases a connection is only supposed to reach by getting through the
	// one before, and the numbers that arrive as play once an intent has been
	// narrowed to a byte.
	for _, intent := range []int32{
		int32(types.PhaseHandshake),
		int32(types.PhaseConfiguration),
		int32(types.PhasePlay),
		int32(types.PhasePlay) + 256,
		int32(types.PhasePlay) + 512,
	} {
		connect(t, srv, types.ProtocolVersions.MINECRAFT_26_2.ID, intent)
	}

	pinging := connect(t, srv, types.ProtocolVersions.MINECRAFT_26_2.ID, int32(types.PhaseStatus))

	if got, want := askStatus(t, pinging).Players, (types.ServerPlayers{Online: 0, Max: 1}); got != want {
		t.Errorf("players = %+v, want %+v: a handshake is not a login", got, want)
	}
}
