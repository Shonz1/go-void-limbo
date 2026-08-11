package main

import (
	"bytes"
	"compress/zlib"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/binary"
	"errors"
	"go-void-limbo/auth"
	clientboundCommon "go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	serverboundCommon "go-void-limbo/packets/serverbound/common"
	"go-void-limbo/registries"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"io"
	"net"
	"slices"
	"sync"
	"testing"
	"time"
)

// newTestClient builds a client whose writes land in the returned buffer. The
// packet ids come from the real registration, so a keep alive written here is
// framed exactly as one written to a connection.
func newTestClient(phase types.Phase) (*MinecraftClient, *bytes.Buffer) {
	packetRegistry := registries.NewPacketRegistry()
	registerPackets(packetRegistry)

	buf := new(bytes.Buffer)

	return &MinecraftClient{
		protocolVersion: types.ProtocolVersions.MINECRAFT_26_2,
		phase:           phase,
		stream:          streams.NewMinecraftStreamFromBuffer(buf),
		packetRegistry:  packetRegistry,
	}, buf
}

// fakeSessionServer stands in for Mojang's, recording what it was asked.
type fakeSessionServer struct {
	usernames []string
	hashes    []string

	profile types.GameProfile
	err     error
}

func (s *fakeSessionServer) HasJoined(username, serverHash string) (types.GameProfile, error) {
	s.usernames = append(s.usernames, username)
	s.hashes = append(s.hashes, serverHash)

	return s.profile, s.err
}

// testKeyPair is generated once for the package. Generating one is the slowest
// thing in these tests and none of them care which key they get.
var testKeyPair = sync.OnceValue(func() *auth.KeyPair {
	keyPair, err := auth.NewKeyPair()
	if err != nil {
		panic(err)
	}

	return keyPair
})

// newLoginClient builds a client on one end of a connection, part way through a
// login, with the other end of the connection returned to play the client.
func newLoginClient(t *testing.T, sessionServer sessionServer) (*MinecraftClient, net.Conn) {
	t.Helper()

	server, client := net.Pipe()

	t.Cleanup(func() {
		server.Close()
		client.Close()
	})

	if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	packetRegistry := registries.NewPacketRegistry()
	registerPackets(packetRegistry)

	return &MinecraftClient{
		protocolVersion: types.ProtocolVersions.MINECRAFT_26_2,
		phase:           types.PhaseLogin,
		conn:            server,
		stream:          streams.NewMinecraftStreamFromNetConn(server),
		packetRegistry:  packetRegistry,
		keyPair:         testKeyPair(),
		sessionServer:   sessionServer,
	}, client
}

// encryptForServer is what a client does to the two fields of its encryption
// response: each one under the server's public key, with the padding the
// client's cipher uses.
func encryptForServer(t *testing.T, plaintext []byte) []byte {
	t.Helper()

	parsed, err := x509.ParsePKIXPublicKey(testKeyPair().PublicKey())
	if err != nil {
		t.Fatalf("reading the server key: %v", err)
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, parsed.(*rsa.PublicKey), plaintext)
	if err != nil {
		t.Fatalf("encrypting for the server: %v", err)
	}

	return ciphertext
}

// testCipher is the client's side of the connection cipher: AES in 8-bit cipher
// feedback mode, keyed by the shared secret and started from it. It stands apart
// from the server's own so that what is checked is that the two agree.
type testCipher struct {
	block     cipher.Block
	register  []byte
	keyStream []byte
	decrypt   bool
}

func newTestCipher(t *testing.T, secret []byte, decrypt bool) *testCipher {
	t.Helper()

	block, err := aes.NewCipher(secret)
	if err != nil {
		t.Fatalf("keying the cipher: %v", err)
	}

	register := make([]byte, len(secret))
	copy(register, secret)

	return &testCipher{block: block, register: register, keyStream: make([]byte, block.BlockSize()), decrypt: decrypt}
}

func (c *testCipher) apply(data []byte) []byte {
	out := make([]byte, len(data))

	for i, in := range data {
		c.block.Encrypt(c.keyStream, c.register)

		out[i] = in ^ c.keyStream[0]

		feedback := out[i]
		if c.decrypt {
			feedback = in
		}

		copy(c.register, c.register[1:])
		c.register[len(c.register)-1] = feedback
	}

	return out
}

func decryptFromServer(t *testing.T, secret, ciphertext []byte) []byte {
	t.Helper()

	return newTestCipher(t, secret, true).apply(ciphertext)
}

// deflate is what a client that was told a threshold does to a body big enough
// for it. It stands apart from the server's own compression so that a frame
// built here is a frame the other end would have built.
func deflate(t *testing.T, body []byte) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	writer := zlib.NewWriter(buf)

	if _, err := writer.Write(body); err != nil {
		t.Fatalf("compressing the body: %v", err)
	}

	if err := writer.Close(); err != nil {
		t.Fatalf("compressing the body: %v", err)
	}

	return buf.Bytes()
}

// frame puts the length in front of a packet body, which is the whole of the
// framing on a connection that has not been told a threshold.
func frame(t *testing.T, body []byte) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := stream.WriteVarInt(int32(len(body))); err != nil {
		t.Fatalf("framing the body: %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("framing the body: %v", err)
	}

	return append(buf.Bytes(), body...)
}

// compressedFrame is frame for a connection that was told a threshold: the body
// gains a var int saying what it inflates to, or zero when it was small enough
// to travel in full. dataLength is written as given so that a client behaving
// badly can be framed too.
func compressedFrame(t *testing.T, body []byte, dataLength int32) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := stream.WriteVarInt(dataLength); err != nil {
		t.Fatalf("framing the body: %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("framing the body: %v", err)
	}

	if dataLength != 0 {
		body = deflate(t, body)
	}

	return frame(t, append(buf.Bytes(), body...))
}

// inflate undoes deflate, for reading back what the server framed.
func inflate(t *testing.T, data []byte) []byte {
	t.Helper()

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("inflating the body: %v", err)
	}

	defer reader.Close()

	body, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("inflating the body: %v", err)
	}

	return body
}

// framedBody splits a frame written to a compressed connection into the size
// the body inflates to, which is zero for one left in full, and the body
// itself. It fails the test on a frame whose length does not match what
// follows it, since a client reads the next frame from where this one ends.
func framedBody(t *testing.T, written []byte) (int32, []byte) {
	t.Helper()

	length, read, err := streams.ReadVarIntFrom(written)
	if err != nil {
		t.Fatalf("reading the packet length: %v", err)
	}

	body := written[read:]
	if int(length) != len(body) {
		t.Fatalf("length says %d bytes, frame carries %d", length, len(body))
	}

	dataLength, read, err := streams.ReadVarIntFrom(body)
	if err != nil {
		t.Fatalf("reading the data length: %v", err)
	}

	return dataLength, body[read:]
}

// enableCompressionOn puts a client on a compressed connection without sending
// the packet that announces it, for the tests that are about the framing rather
// than about announcing it.
func enableCompressionOn(client *MinecraftClient, threshold int32) {
	client.compressionEnabled = true
	client.compressionThreshold = threshold
}

// The keep alive a client answers with, as it arrives in the play phase: the
// packet id and the eight byte id it was sent.
var keepAliveBody = []byte{0x1C, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x04, 0xD2}

// The client answers a keep alive with the id it carried, so what was sent has
// to be readable back out of the frame.
func keepAliveIdIn(t *testing.T, frame []byte, wantPacketId byte) int64 {
	t.Helper()

	// A one byte length, a one byte packet id and the eight byte id itself.
	want := len(frame) == 10 && frame[0] == 0x09 && frame[1] == wantPacketId
	if !want {
		t.Fatalf("frame = % x, want a 9 byte body starting with packet id %#02x", frame, wantPacketId)
	}

	return int64(binary.BigEndian.Uint64(frame[2:]))
}

func TestSendKeepAliveUsesThePacketIdOfThePhaseItIsSentIn(t *testing.T) {
	// The same packet, and a different id in each of the two phases that have
	// one. Sending a configuration keep alive to a client in play is a packet
	// that client reads as something else entirely.
	tests := []struct {
		name     string
		phase    types.Phase
		packetId byte
	}{
		{name: "configuration", phase: types.PhaseConfiguration, packetId: 0x04},
		{name: "play", phase: types.PhasePlay, packetId: 0x2C},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, buf := newTestClient(test.phase)

			if err := client.sendKeepAlive(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			id := keepAliveIdIn(t, buf.Bytes(), test.packetId)

			// An answer is matched against what is being waited on, so the two
			// have to be the same number.
			if id != client.pendingKeepAlive {
				t.Errorf("sent id %d, waiting on %d", id, client.pendingKeepAlive)
			}

			if id == 0 {
				t.Error("sent id 0, which is the value that means nothing is being waited on")
			}
		})
	}
}

func TestSendKeepAliveSendsNothingInPhasesWithoutOne(t *testing.T) {
	// Neither phase has a keep alive packet to send. Both are exchanges the
	// client drives packet by packet, so silence in them is the server's turn
	// rather than a connection going quiet.
	for _, phase := range []types.Phase{types.PhaseHandshake, types.PhaseLogin} {
		client, buf := newTestClient(phase)

		if err := client.sendKeepAlive(); err != nil {
			t.Fatalf("phase %d: unexpected error: %v", phase, err)
		}

		if buf.Len() != 0 {
			t.Errorf("phase %d: wrote % x, want nothing", phase, buf.Bytes())
		}

		// Waiting on an answer that was never asked for would drop the next
		// connection to reach configuration.
		if client.pendingKeepAlive != 0 {
			t.Errorf("phase %d: waiting on keep alive %d, want none", phase, client.pendingKeepAlive)
		}
	}
}

func TestSendKeepAliveReportsOneThatWentUnanswered(t *testing.T) {
	client, _ := newTestClient(types.PhasePlay)

	if err := client.sendKeepAlive(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A whole interval later with no answer. Asking again would leave a client
	// that stopped reading alive on a connection nothing can reach it through.
	if err := client.sendKeepAlive(); !errors.Is(err, errKeepAliveTimeout) {
		t.Errorf("error = %v, want %v", err, errKeepAliveTimeout)
	}
}

func TestConfirmKeepAlive(t *testing.T) {
	client, buf := newTestClient(types.PhasePlay)

	if err := client.sendKeepAlive(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	id := keepAliveIdIn(t, buf.Bytes(), 0x2C)

	// An id that answers a keep alive the server never sent, or a different one
	// from the one it is waiting on, says nothing about the connection being
	// live.
	if err := client.ConfirmKeepAlive(id + 1); err == nil {
		t.Error("error = nil, want an error for an answer to a different keep alive")
	}

	if client.pendingKeepAlive != id {
		t.Errorf("waiting on %d, want the wrong answer to have changed nothing", client.pendingKeepAlive)
	}

	if err := client.ConfirmKeepAlive(id); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.pendingKeepAlive != 0 {
		t.Errorf("waiting on %d, want an answered keep alive to be waited on no longer", client.pendingKeepAlive)
	}

	if err := client.ConfirmKeepAlive(id); err == nil {
		t.Error("error = nil, want an error for an answer to nothing")
	}

	// An answered keep alive is what lets the next one go out.
	buf.Reset()

	if err := client.sendKeepAlive(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	keepAliveIdIn(t, buf.Bytes(), 0x2C)
}

// The loop is what turns an unanswered keep alive into a closed connection, and
// closing is what ends the read loop on the other side of it.
func TestKeepAliveLoopDropsAClientThatNeverAnswers(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()

	packetRegistry := registries.NewPacketRegistry()
	registerPackets(packetRegistry)

	mc := &MinecraftClient{
		protocolVersion: types.ProtocolVersions.MINECRAFT_26_2,
		phase:           types.PhasePlay,
		conn:            server,
		stream:          streams.NewMinecraftStreamFromNetConn(server),
		packetRegistry:  packetRegistry,
	}

	done := make(chan struct{})
	defer close(done)

	// A tick short enough to keep the test quick. What it stands in for is the
	// fifteen seconds a real connection gets.
	go mc.keepAliveLoop(done, 10*time.Millisecond)

	if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frame := make([]byte, 10)
	if _, err := io.ReadFull(client, frame); err != nil {
		t.Fatalf("reading the keep alive: %v", err)
	}

	keepAliveIdIn(t, frame, 0x2C)

	// Nothing is sent back, so the next tick finds the keep alive unanswered.
	if _, err := client.Read(make([]byte, 1)); !errors.Is(err, io.EOF) {
		t.Errorf("error = %v, want the connection closed", err)
	}
}

func TestEnableCompressionAnnouncesTheThresholdInThePlainFraming(t *testing.T) {
	client, buf := newTestClient(types.PhaseLogin)

	if err := client.EnableCompression(256); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A length, the login phase's set compression id, and 256 as a var int. A
	// data length in front of this body is one the client is not reading for
	// yet, since this is the packet that tells it to start.
	want := []byte{0x03, 0x03, 0x80, 0x02}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("wrote % x, want % x", buf.Bytes(), want)
	}

	threshold, enabled := client.compression()
	if !enabled || threshold != 256 {
		t.Errorf("threshold = %d enabled = %t, want 256 and enabled", threshold, enabled)
	}

	// Everything after it is framed for the threshold, whether or not it is big
	// enough to be deflated.
	buf.Reset()

	if err := client.WritePacket(&clientboundLogin.DisconnectClientboundPacket{Reason: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A length, a data length of zero for a body left in full, the login
	// disconnect id, and a one character reason.
	want = []byte{0x04, 0x00, 0x00, 0x01, 'x'}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("wrote % x, want % x", buf.Bytes(), want)
	}

	// A client can only be told once, and telling it again would leave it
	// reading for a threshold while another one is being written for.
	if err := client.EnableCompression(256); err == nil {
		t.Error("error = nil, want compression to be refused a second time")
	}
}

func TestEnableCompressionRefusesAThresholdThatIsNotOne(t *testing.T) {
	client, buf := newTestClient(types.PhaseLogin)

	// A negative threshold is how the protocol says compression is off. Sending
	// it and then framing everything for it is the one thing that cannot be
	// meant, so it is refused rather than half done.
	if err := client.EnableCompression(-1); err == nil {
		t.Error("error = nil, want a negative threshold refused")
	}

	if buf.Len() != 0 {
		t.Errorf("wrote % x, want nothing", buf.Bytes())
	}

	if _, enabled := client.compression(); enabled {
		t.Error("compression is on after a threshold that was refused")
	}
}

func TestWritePacketDeflatesOnlyTheBodiesAtTheThreshold(t *testing.T) {
	const threshold = 256

	client, buf := newTestClient(types.PhaseConfiguration)
	enableCompressionOn(client, threshold)

	if err := client.WritePacket(&clientboundCommon.KeepAliveClientboundPacket{Id: 1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Nine bytes of keep alive behind a data length of zero: deflating a body
	// this size costs more bytes than it saves.
	want := []byte{0x0A, 0x00, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Fatalf("wrote % x, want % x", buf.Bytes(), want)
	}

	buf.Reset()

	// A body of exactly the threshold is deflated: the threshold is the size
	// deflating starts at, not the size it starts after, and a client reading
	// the boundary the other way reads the frame as something else entirely.
	// One byte of the body is the packet id.
	if err := client.WritePacket(clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:worldgen/biome", bytes.Repeat([]byte("a"), threshold-1))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dataLength, _ := framedBody(t, buf.Bytes()); dataLength != threshold {
		t.Errorf("data length = %d, want a body of exactly %d deflated", dataLength, threshold)
	}

	buf.Reset()

	// One byte under it is not, however close it came.
	if err := client.WritePacket(clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:worldgen/biome", bytes.Repeat([]byte("a"), threshold-2))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dataLength, _ := framedBody(t, buf.Bytes()); dataLength != 0 {
		t.Errorf("data length = %d, want a body one byte under the threshold left in full", dataLength)
	}

	buf.Reset()

	// A registry is the one thing a limbo sends that is worth deflating, and
	// repetitive enough for it to pay off.
	registry := bytes.Repeat([]byte("minecraft:dimension_type"), 32)

	if err := client.WritePacket(clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:dimension_type", registry)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataLength, deflated := framedBody(t, buf.Bytes())

	// The configuration phase's registry data id, and then the body as given.
	wantBody := append([]byte{0x07}, registry...)

	if int(dataLength) != len(wantBody) {
		t.Errorf("data length = %d, want the %d bytes the body inflates to", dataLength, len(wantBody))
	}

	if len(deflated) >= len(wantBody) {
		t.Errorf("deflated %d bytes into %d, want fewer", len(wantBody), len(deflated))
	}

	if got := inflate(t, deflated); !bytes.Equal(got, wantBody) {
		t.Errorf("body inflates to % x, want % x", got, wantBody)
	}
}

func TestReadPacketReadsWhicheverFramingTheConnectionIsOn(t *testing.T) {
	tests := []struct {
		name      string
		threshold int32
		enabled   bool
		frame     []byte
	}{
		{name: "no threshold", frame: frame(t, keepAliveBody)},
		{name: "under the threshold", threshold: 256, enabled: true, frame: compressedFrame(t, keepAliveBody, 0)},
		{name: "deflated", threshold: 4, enabled: true, frame: compressedFrame(t, keepAliveBody, int32(len(keepAliveBody)))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, buf := newTestClient(types.PhasePlay)
			if test.enabled {
				enableCompressionOn(client, test.threshold)
			}

			buf.Write(test.frame)

			packet, handler, err := client.ReadPacket()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			keepAlive, ok := packet.(*serverboundCommon.KeepAliveServerboundPacket)
			if !ok {
				t.Fatalf("read %T, want a keep alive", packet)
			}

			if keepAlive.Id != 1234 {
				t.Errorf("id = %d, want 1234", keepAlive.Id)
			}

			// An answer the server does not act on is a connection it drops on
			// the next keep alive.
			if handler == nil {
				t.Error("handler = nil, want the keep alive handled")
			}
		})
	}
}

func TestReadPacketReportsABodyItCannotInflate(t *testing.T) {
	client, buf := newTestClient(types.PhasePlay)
	enableCompressionOn(client, 256)

	// A data length under the threshold is a body the client had no business
	// deflating, and one this end cannot tell apart from a frame that lost its
	// place. A body claiming to inflate to more than it does is the other way
	// the same frame goes wrong.
	buf.Write(compressedFrame(t, keepAliveBody, int32(len(keepAliveBody))))
	buf.Write(frame(t, append([]byte{0x80, 0x02}, deflate(t, keepAliveBody)...)))

	// Behind them, a frame there is nothing wrong with.
	buf.Write(compressedFrame(t, keepAliveBody, 0))

	for _, name := range []string{"below the threshold", "shorter than it claims"} {
		_, _, err := client.ReadPacket()

		var packetErr *packetError
		if !errors.As(err, &packetErr) {
			t.Fatalf("%s: error = %v, want a packet error the read loop carries on from", name, err)
		}
	}

	// The bodies were read in full, so the frame behind them starts where it
	// should and the connection is still worth reading.
	packet, _, err := client.ReadPacket()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := packet.(*serverboundCommon.KeepAliveServerboundPacket); !ok {
		t.Errorf("read %T, want the keep alive behind the frames that failed", packet)
	}
}

// loginPeer is the far side of a login: it sends what a client sends and reads
// what a client reads, taking on the cipher and then the compressed framing at
// the points a real client takes them on.
type loginPeer struct {
	t    *testing.T
	conn net.Conn

	encrypter  *testCipher
	decrypter  *testCipher
	compressed bool
}

// encrypt turns the connection cipher on, as a client does the instant it has
// sent its encryption response.
func (p *loginPeer) encrypt(secret []byte) {
	p.encrypter = newTestCipher(p.t, secret, false)
	p.decrypter = newTestCipher(p.t, secret, true)
}

func (p *loginPeer) writePacket(body []byte) {
	p.t.Helper()

	// Everything this side sends is small enough to travel in full, so a zero
	// data length is the whole of what compression adds to it.
	if p.compressed {
		body = append([]byte{0x00}, body...)
	}

	written := frame(p.t, body)
	if p.encrypter != nil {
		written = p.encrypter.apply(written)
	}

	if _, err := p.conn.Write(written); err != nil {
		p.t.Fatalf("writing to the connection: %v", err)
	}
}

func (p *loginPeer) readByte() byte {
	p.t.Helper()

	buf := make([]byte, 1)
	if _, err := io.ReadFull(p.conn, buf); err != nil {
		p.t.Fatalf("reading from the connection: %v", err)
	}

	if p.decrypter != nil {
		buf = p.decrypter.apply(buf)
	}

	return buf[0]
}

func (p *loginPeer) readVarInt() int32 {
	p.t.Helper()

	value := int32(0)

	for position := 0; position < 32; position += 7 {
		current := p.readByte()
		value |= int32(current&0x7F) << position

		if current&0x80 == 0 {
			return value
		}
	}

	p.t.Fatal("reading from the connection: var int too big")

	return 0
}

// readPacket reads one frame and returns the body inside it, through whichever
// of the cipher and the compressed framing the connection has reached.
func (p *loginPeer) readPacket() *streams.MinecraftStream {
	p.t.Helper()

	length := p.readVarInt()

	body := make([]byte, length)
	for i := range body {
		body[i] = p.readByte()
	}

	if p.compressed {
		size, read, err := streams.ReadVarIntFrom(body)
		if err != nil {
			p.t.Fatalf("reading the data length: %v", err)
		}

		body = body[read:]
		if size != 0 {
			body = inflate(p.t, body)
		}
	}

	return streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))
}

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

	sessionServer := &fakeSessionServer{profile: authenticated}

	packetRegistry := registries.NewPacketRegistry()
	registerPackets(packetRegistry)

	srv := &server{packetRegistry: packetRegistry, keyPair: testKeyPair(), sessionServer: sessionServer}

	conn, client := net.Pipe()
	defer client.Close()

	if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go srv.handleConnection(conn)

	peer := &loginPeer{t: t, conn: client}

	// A handshake announcing the protocol version and an intent to log in, and
	// then the name and uuid the client claims for itself.
	handshake := new(bytes.Buffer)
	handshakeStream := streams.NewMinecraftStreamFromBuffer(handshake)

	for _, write := range []func() error{
		func() error { return handshakeStream.WriteVarInt(0x00) },
		func() error { return handshakeStream.WriteVarInt(int32(types.ProtocolVersions.MINECRAFT_26_2.ID)) },
		func() error { return handshakeStream.WriteString("localhost") },
		func() error { return handshakeStream.WriteShort(25565) },
		func() error { return handshakeStream.WriteVarInt(int32(types.PhaseLogin)) },
		handshakeStream.Flush,
	} {
		if err := write(); err != nil {
			t.Fatalf("building the handshake: %v", err)
		}
	}

	peer.writePacket(handshake.Bytes())

	loginStart := new(bytes.Buffer)
	loginStartStream := streams.NewMinecraftStreamFromBuffer(loginStart)

	for _, write := range []func() error{
		func() error { return loginStartStream.WriteVarInt(0x00) },
		func() error { return loginStartStream.WriteString("notch") },
		func() error { return loginStartStream.WriteUuid("00000000-0000-0000-0000-000000000001") },
		loginStartStream.Flush,
	} {
		if err := write(); err != nil {
			t.Fatalf("building the login start: %v", err)
		}
	}

	peer.writePacket(loginStart.Bytes())

	request := peer.readPacket()

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

	response := new(bytes.Buffer)
	responseStream := streams.NewMinecraftStreamFromBuffer(response)

	for _, write := range []func() error{
		func() error { return responseStream.WriteVarInt(0x01) },
		func() error {
			return responseStream.WriteByteArray(encryptForServer(t, secret))
		},
		func() error {
			return responseStream.WriteByteArray(encryptForServer(t, verifyToken))
		},
		responseStream.Flush,
	} {
		if err := write(); err != nil {
			t.Fatalf("building the encryption response: %v", err)
		}
	}

	peer.writePacket(response.Bytes())

	// From here the client reads and writes through the cipher, whatever the
	// server has managed.
	peer.encrypt(secret)

	setCompression := peer.readPacket()

	if packetId, err := setCompression.ReadVarInt(); err != nil || packetId != 0x03 {
		t.Fatalf("packet id = %#02x err = %v, want the login phase's set compression %#02x", packetId, err, 0x03)
	}

	threshold, err := setCompression.ReadVarInt()
	if err != nil {
		t.Fatalf("reading the threshold: %v", err)
	}

	// The client frames everything after this packet for the threshold, and so
	// does the server.
	peer.compressed = true

	loginSuccess := peer.readPacket()

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
	if !slices.Equal(sessionServer.usernames, []string{"notch"}) {
		t.Errorf("asked about %v, want the name the client logged in under", sessionServer.usernames)
	}

	// The vanilla threshold, which is what the login announces.
	if threshold != 256 {
		t.Errorf("threshold = %d, want %d", threshold, 256)
	}
}

func TestBeginEncryptionOffersTheServerKeyAndANewToken(t *testing.T) {
	client, _ := newLoginClient(t, &fakeSessionServer{})

	publicKey, verifyToken, err := client.BeginEncryption()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(publicKey, testKeyPair().PublicKey()) {
		t.Errorf("public key = % x, want the server's", publicKey)
	}

	if !bytes.Equal(verifyToken, client.verifyToken) {
		t.Errorf("sent token % x, waiting on % x", verifyToken, client.verifyToken)
	}

	if len(verifyToken) == 0 {
		t.Error("sent an empty verify token, which is one anybody can answer with")
	}

	// A second request would leave the connection waiting on a token it never
	// sent, and the client answering the one it was.
	if _, _, err := client.BeginEncryption(); err == nil {
		t.Error("error = nil, want encryption refused a second time")
	}
}

func TestCompleteEncryptionPutsTheConnectionUnderTheSecret(t *testing.T) {
	client, peer := newLoginClient(t, &fakeSessionServer{})

	_, verifyToken, err := client.BeginEncryption()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secret := []byte("0123456789abcdef")

	if err := client.CompleteEncryption(encryptForServer(t, secret), encryptForServer(t, verifyToken)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The client is reading through its own cipher from the moment it answered,
	// so a packet it cannot decrypt is a connection it drops.
	read := make(chan []byte, 1)
	go func() {
		// A length, the login disconnect id, and a one character reason.
		frame := make([]byte, 4)
		if _, err := io.ReadFull(peer, frame); err != nil {
			t.Errorf("reading the packet: %v", err)
			close(read)

			return
		}

		read <- decryptFromServer(t, secret, frame)
	}()

	if err := client.WritePacket(&clientboundLogin.DisconnectClientboundPacket{Reason: "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := <-read; !bytes.Equal(got, []byte{0x03, 0x00, 0x01, 'x'}) {
		t.Errorf("the connection carried % x, want the packet under the client's cipher", got)
	}

	// The secret outlives the packet that carried it, because the session server
	// matches a login by a hash taken over it.
	if !bytes.Equal(client.sharedSecret, secret) {
		t.Errorf("kept secret % x, want the one the client sent", client.sharedSecret)
	}

	// A connection with nothing outstanding is one a second response cannot
	// rekey.
	if client.verifyToken != nil {
		t.Errorf("waiting on token % x, want an answered request waited on no longer", client.verifyToken)
	}

	if err := client.CompleteEncryption(encryptForServer(t, secret), encryptForServer(t, verifyToken)); err == nil {
		t.Error("error = nil, want a second encryption response refused")
	}
}

func TestCompleteEncryptionRefusesAResponseItCannotTie(t *testing.T) {
	secret := []byte("0123456789abcdef")

	tests := []struct {
		name         string
		begin        bool
		sharedSecret func(t *testing.T, verifyToken []byte) []byte
		verifyToken  func(t *testing.T, verifyToken []byte) []byte
	}{
		{
			name:         "before a request was sent",
			sharedSecret: func(t *testing.T, _ []byte) []byte { return encryptForServer(t, secret) },
			verifyToken:  func(t *testing.T, _ []byte) []byte { return encryptForServer(t, []byte{0x01, 0x02, 0x03, 0x04}) },
		},
		{
			name:         "with another connection's token",
			begin:        true,
			sharedSecret: func(t *testing.T, _ []byte) []byte { return encryptForServer(t, secret) },
			verifyToken:  func(t *testing.T, _ []byte) []byte { return encryptForServer(t, []byte{0x01, 0x02, 0x03, 0x04}) },
		},
		{
			name:         "with a token encrypted to nobody",
			begin:        true,
			sharedSecret: func(t *testing.T, _ []byte) []byte { return encryptForServer(t, secret) },
			verifyToken:  func(t *testing.T, _ []byte) []byte { return []byte("not a ciphertext") },
		},
		{
			name:         "with a secret encrypted to nobody",
			begin:        true,
			sharedSecret: func(t *testing.T, _ []byte) []byte { return []byte("not a ciphertext") },
			verifyToken:  func(t *testing.T, verifyToken []byte) []byte { return encryptForServer(t, verifyToken) },
		},
		{
			name:         "with a secret that is not a key",
			begin:        true,
			sharedSecret: func(t *testing.T, _ []byte) []byte { return encryptForServer(t, []byte("too short")) },
			verifyToken:  func(t *testing.T, verifyToken []byte) []byte { return encryptForServer(t, verifyToken) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := newLoginClient(t, &fakeSessionServer{})

			var verifyToken []byte
			if test.begin {
				var err error
				if _, verifyToken, err = client.BeginEncryption(); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			if err := client.CompleteEncryption(test.sharedSecret(t, verifyToken), test.verifyToken(t, verifyToken)); err == nil {
				t.Fatal("error = nil, want the response refused")
			}

			// A login that got this far and failed is one nothing else should be
			// able to act on.
			if client.sharedSecret != nil {
				t.Errorf("kept secret % x, want a refused response to leave the connection alone", client.sharedSecret)
			}

			if _, err := client.Authenticate(); err == nil {
				t.Error("error = nil, want a connection that was never encrypted refused")
			}
		})
	}
}

func TestAuthenticateAsksAboutTheLoginTheConnectionIsIn(t *testing.T) {
	authenticated := types.GameProfile{Uuid: "069a79f4-44e9-4726-a5be-fca90e38aaf5", Username: "Notch"}
	sessionServer := &fakeSessionServer{profile: authenticated}

	client, _ := newLoginClient(t, sessionServer)
	client.SetProfile(types.GameProfile{Username: "Notch"})

	_, verifyToken, err := client.BeginEncryption()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secret := []byte("0123456789abcdef")

	if err := client.CompleteEncryption(encryptForServer(t, secret), encryptForServer(t, verifyToken)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	profile, err := client.Authenticate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if profile.String() != authenticated.String() {
		t.Errorf("profile = %s, want the one the session server answered with %s", profile, authenticated)
	}

	// The client is asked about by the name it logged in under, and by the hash
	// it derived over the same secret and the same key before it answered.
	if !slices.Equal(sessionServer.usernames, []string{"Notch"}) {
		t.Errorf("asked about %v, want the name the client logged in under", sessionServer.usernames)
	}

	want := auth.ServerHash(serverId, secret, testKeyPair().PublicKey())
	if !slices.Equal(sessionServer.hashes, []string{want}) {
		t.Errorf("asked about %v, want the hash over this login %q", sessionServer.hashes, want)
	}
}

func TestKeepAliveLoopStopsWithTheConnection(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	defer server.Close()

	packetRegistry := registries.NewPacketRegistry()
	registerPackets(packetRegistry)

	mc := &MinecraftClient{
		protocolVersion: types.ProtocolVersions.MINECRAFT_26_2,
		phase:           types.PhasePlay,
		conn:            server,
		stream:          streams.NewMinecraftStreamFromNetConn(server),
		packetRegistry:  packetRegistry,
	}

	done := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		mc.keepAliveLoop(done, time.Hour)
		close(stopped)
	}()

	// The read loop closes done when the connection ends, and a keep alive
	// goroutine left running would outlive every connection the server takes.
	close(done)

	select {
	case <-stopped:
	case <-time.After(5 * time.Second):
		t.Error("the keep alive loop is still running after the connection ended")
	}
}
