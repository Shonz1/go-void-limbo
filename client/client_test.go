package client

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/Shonz1/go-void-limbo/auth"
	"github.com/Shonz1/go-void-limbo/internal/testutil"
	clientboundCommon "github.com/Shonz1/go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "github.com/Shonz1/go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	clientboundPlay "github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	serverboundCommon "github.com/Shonz1/go-void-limbo/packets/serverbound/common"
	"github.com/Shonz1/go-void-limbo/protocol"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// fakeStatus stands in for the server-wide status a real connection is handed,
// counting what the connection tells it.
type fakeStatus struct {
	mu     sync.Mutex
	joins  int
	leaves int
}

func (s *fakeStatus) Status(types.ProtocolVersion) types.ServerStatus {
	return types.ServerStatus{}
}

func (s *fakeStatus) PlayerJoined() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.joins++
}

func (s *fakeStatus) PlayerLeft() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.leaves++
}

func (s *fakeStatus) counts() (int, int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.joins, s.leaves
}

// newTestClient builds a client whose writes land in the returned buffer. The
// packet ids come from the real registration, so a keep alive written here is
// framed exactly as one written to a connection.
func newTestClient(phase types.Phase) (*Client, *bytes.Buffer) {
	return newTestClientOn(phase, types.ProtocolVersions.MINECRAFT_26_2)
}

// newTestClientOn is newTestClient for the tests that care which version the
// client is on, which are the ones about what the transformers do to a packet
// on its way past.
func newTestClientOn(phase types.Phase, protocolVersion types.ProtocolVersion) (*Client, *bytes.Buffer) {
	buf := new(bytes.Buffer)

	return &Client{
		protocolVersion: protocolVersion,
		phase:           phase,
		stream:          streams.NewMinecraftStreamFromBuffer(buf),
		packetRegistry:  protocol.NewDefaultRegistry(),
		status:          new(fakeStatus),
	}, buf
}

// newLoginClient builds a client on one end of a connection, part way through a
// login, with the other end of the connection returned to play the client.
func newLoginClient(t *testing.T, sessionServer SessionServer) (*Client, net.Conn) {
	t.Helper()

	server, client := net.Pipe()

	t.Cleanup(func() {
		server.Close()
		client.Close()
	})

	if err := client.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return &Client{
		protocolVersion:   types.ProtocolVersions.MINECRAFT_26_2,
		phase:             types.PhaseLogin,
		conn:              server,
		stream:            streams.NewMinecraftStreamFromNetConn(server),
		packetRegistry:    protocol.NewDefaultRegistry(),
		keyPair:           testutil.KeyPair(),
		sessionServer:     sessionServer,
		status:            new(fakeStatus),
		encryptionEnabled: true,
	}, client
}

// enableCompressionOn puts a client on a compressed connection without sending
// the packet that announces it, for the tests that are about the framing rather
// than about announcing it.
func enableCompressionOn(client *Client, threshold int32) {
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

	mc := &Client{
		protocolVersion: types.ProtocolVersions.MINECRAFT_26_2,
		phase:           types.PhasePlay,
		conn:            server,
		stream:          streams.NewMinecraftStreamFromNetConn(server),
		packetRegistry:  protocol.NewDefaultRegistry(),
		status:          new(fakeStatus),
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

func TestKeepAliveLoopStopsWithTheConnection(t *testing.T) {
	server, client := net.Pipe()
	defer client.Close()
	defer server.Close()

	mc := &Client{
		protocolVersion: types.ProtocolVersions.MINECRAFT_26_2,
		phase:           types.PhasePlay,
		conn:            server,
		stream:          streams.NewMinecraftStreamFromNetConn(server),
		packetRegistry:  protocol.NewDefaultRegistry(),
		status:          new(fakeStatus),
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

	if dataLength, _ := testutil.FramedBody(t, buf.Bytes()); dataLength != threshold {
		t.Errorf("data length = %d, want a body of exactly %d deflated", dataLength, threshold)
	}

	buf.Reset()

	// One byte under it is not, however close it came.
	if err := client.WritePacket(clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:worldgen/biome", bytes.Repeat([]byte("a"), threshold-2))); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dataLength, _ := testutil.FramedBody(t, buf.Bytes()); dataLength != 0 {
		t.Errorf("data length = %d, want a body one byte under the threshold left in full", dataLength)
	}

	buf.Reset()

	// A registry is the one thing a limbo sends that is worth deflating, and
	// repetitive enough for it to pay off.
	registry := bytes.Repeat([]byte("minecraft:dimension_type"), 32)

	if err := client.WritePacket(clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:dimension_type", registry)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dataLength, deflated := testutil.FramedBody(t, buf.Bytes())

	// The configuration phase's registry data id, and then the body as given.
	wantBody := append([]byte{0x07}, registry...)

	if int(dataLength) != len(wantBody) {
		t.Errorf("data length = %d, want the %d bytes the body inflates to", dataLength, len(wantBody))
	}

	if len(deflated) >= len(wantBody) {
		t.Errorf("deflated %d bytes into %d, want fewer", len(wantBody), len(deflated))
	}

	if got := testutil.Inflate(t, deflated); !bytes.Equal(got, wantBody) {
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
		{name: "no threshold", frame: testutil.Frame(t, keepAliveBody)},
		{name: "under the threshold", threshold: 256, enabled: true, frame: testutil.CompressedFrame(t, keepAliveBody, 0)},
		{name: "deflated", threshold: 4, enabled: true, frame: testutil.CompressedFrame(t, keepAliveBody, int32(len(keepAliveBody)))},
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
	buf.Write(testutil.CompressedFrame(t, keepAliveBody, int32(len(keepAliveBody))))
	buf.Write(testutil.Frame(t, append([]byte{0x80, 0x02}, testutil.Deflate(t, keepAliveBody)...)))

	// Behind them, a frame there is nothing wrong with.
	buf.Write(testutil.CompressedFrame(t, keepAliveBody, 0))

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

// A connection that has not reached play may not claim the protocol's full
// frame size: nothing legitimate before then comes anywhere near it, so the
// length is refused before a body that big is allocated for. The same frame is
// taken at its word in play.
func TestReadPacketCapsWhatAPrePlayFrameMayClaim(t *testing.T) {
	body := make([]byte, maxPrePlayPacketSize+1)
	body[0] = 0x7F

	tests := []struct {
		name   string
		phase  types.Phase
		capped bool
	}{
		{name: "capped before play", phase: types.PhaseLogin, capped: true},
		{name: "at the protocol maximum in play", phase: types.PhasePlay, capped: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, buf := newTestClient(test.phase)
			buf.Write(testutil.Frame(t, body))

			_, _, err := client.ReadPacket()

			var packetErr *packetError
			if test.capped {
				if err == nil || errors.As(err, &packetErr) {
					t.Fatalf("error = %v, want the connection refused on the length alone", err)
				}

				return
			}

			// The frame was read in full: what fails is the unknown id inside
			// it, which is the read loop carrying on rather than the
			// connection ending.
			if !errors.As(err, &packetErr) {
				t.Fatalf("error = %v, want a packet error about the id inside the frame", err)
			}
		})
	}
}

// A connection on a server that holds a secret asks its question once, and takes
// one answer to it. A second payload arrives too late to name anybody else.
func TestModernForwardingIsAnsweredOnce(t *testing.T) {
	client, _ := newLoginClient(t, &testutil.FakeSessionServer{})
	client.forwardingSecret = testutil.ForwardingSecret

	messageId, err := client.BeginModernForwarding()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := client.BeginModernForwarding(); err == nil {
		t.Error("error = nil, want a second request refused")
	}

	payload := testutil.SignedForwardingPayload(t, testutil.ForwardingSecret, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

	forwarded, err := client.CompleteModernForwarding(messageId, payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if forwarded.Uuid != "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Errorf("uuid = %q, want the account the payload was signed for", forwarded.Uuid)
	}

	kept, ok := client.ForwardedLogin()
	if !ok || kept.Username != "Notch" {
		t.Errorf("kept %s (%t), want the login the payload carried", kept, ok)
	}

	second := testutil.SignedForwardingPayload(t, testutil.ForwardingSecret, "203.0.113.7", "00000000-0000-0000-0000-000000000002", "Somebody", nil)

	if _, err := client.CompleteModernForwarding(messageId, second); err == nil {
		t.Error("error = nil, want a second payload refused")
	}

	if kept, _ := client.ForwardedLogin(); kept.Username != "Notch" {
		t.Errorf("kept %s, want the login the first payload carried", kept)
	}
}

// The message id ties an answer to the question this connection asked. A payload
// answering some other number is not this connection's to read, however well it
// is signed.
func TestModernForwardingRefusesAnAnswerToAnotherRequest(t *testing.T) {
	client, _ := newLoginClient(t, &testutil.FakeSessionServer{})
	client.forwardingSecret = testutil.ForwardingSecret

	messageId, err := client.BeginModernForwarding()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := testutil.SignedForwardingPayload(t, testutil.ForwardingSecret, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

	if _, err := client.CompleteModernForwarding(messageId+1, payload); err == nil {
		t.Error("error = nil, want a payload answering another request refused")
	}

	if _, ok := client.ForwardedLogin(); ok {
		t.Error("kept a forwarded login, want the one that answered nothing left out")
	}
}

// The request is answered by an answer with no login in it as much as by one that
// works, so the connection gives up on it and has nothing outstanding afterwards.
// The login goes on to be settled without a proxy, and a payload turning up
// behind that would be settling it a second time.
func TestModernForwardingGivesUpOnARequestTheConnectionCannotAnswer(t *testing.T) {
	client, _ := newLoginClient(t, &testutil.FakeSessionServer{})
	client.forwardingSecret = testutil.ForwardingSecret

	messageId, err := client.BeginModernForwarding()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The id ties the answer to the question, whichever kind of answer it is.
	if err := client.DeclineModernForwarding(messageId + 1); err == nil {
		t.Error("error = nil, want an answer to another request refused")
	}

	if err := client.DeclineModernForwarding(messageId); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := testutil.SignedForwardingPayload(t, testutil.ForwardingSecret, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

	if _, err := client.CompleteModernForwarding(messageId, payload); err == nil {
		t.Error("error = nil, want a payload behind a request already given up on refused")
	}

	if _, ok := client.ForwardedLogin(); ok {
		t.Error("kept a forwarded login, want the one that answered nothing left out")
	}

	// And the connection cannot give up twice, since the second time there is
	// nothing left to give up on.
	if err := client.DeclineModernForwarding(messageId); err == nil {
		t.Error("error = nil, want a second answer to the same request refused")
	}
}

// A payload nobody asked for. A connection that has not sent the request has
// nothing outstanding, and a login start is what sends it, so this is a client
// trying to name itself before the question was put to it.
func TestModernForwardingRefusesAPayloadNothingAskedFor(t *testing.T) {
	client, _ := newLoginClient(t, &testutil.FakeSessionServer{})
	client.forwardingSecret = testutil.ForwardingSecret

	payload := testutil.SignedForwardingPayload(t, testutil.ForwardingSecret, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

	if _, err := client.CompleteModernForwarding(1, payload); err == nil {
		t.Error("error = nil, want a payload nothing asked for refused")
	}
}

func TestBeginEncryptionOffersTheServerKeyAndANewToken(t *testing.T) {
	client, _ := newLoginClient(t, &testutil.FakeSessionServer{})

	publicKey, verifyToken, err := client.BeginEncryption()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(publicKey, testutil.KeyPair().PublicKey()) {
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
	client, peer := newLoginClient(t, &testutil.FakeSessionServer{})

	_, verifyToken, err := client.BeginEncryption()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secret := []byte("0123456789abcdef")

	if err := client.CompleteEncryption(testutil.EncryptForServer(t, secret), testutil.EncryptForServer(t, verifyToken)); err != nil {
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

		read <- testutil.DecryptFromServer(t, secret, frame)
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

	if err := client.CompleteEncryption(testutil.EncryptForServer(t, secret), testutil.EncryptForServer(t, verifyToken)); err == nil {
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
			sharedSecret: func(t *testing.T, _ []byte) []byte { return testutil.EncryptForServer(t, secret) },
			verifyToken: func(t *testing.T, _ []byte) []byte {
				return testutil.EncryptForServer(t, []byte{0x01, 0x02, 0x03, 0x04})
			},
		},
		{
			name:         "with another connection's token",
			begin:        true,
			sharedSecret: func(t *testing.T, _ []byte) []byte { return testutil.EncryptForServer(t, secret) },
			verifyToken: func(t *testing.T, _ []byte) []byte {
				return testutil.EncryptForServer(t, []byte{0x01, 0x02, 0x03, 0x04})
			},
		},
		{
			name:         "with a token encrypted to nobody",
			begin:        true,
			sharedSecret: func(t *testing.T, _ []byte) []byte { return testutil.EncryptForServer(t, secret) },
			verifyToken:  func(t *testing.T, _ []byte) []byte { return []byte("not a ciphertext") },
		},
		{
			name:         "with a secret encrypted to nobody",
			begin:        true,
			sharedSecret: func(t *testing.T, _ []byte) []byte { return []byte("not a ciphertext") },
			verifyToken:  func(t *testing.T, verifyToken []byte) []byte { return testutil.EncryptForServer(t, verifyToken) },
		},
		{
			name:         "with a secret that is not a key",
			begin:        true,
			sharedSecret: func(t *testing.T, _ []byte) []byte { return testutil.EncryptForServer(t, []byte("too short")) },
			verifyToken:  func(t *testing.T, verifyToken []byte) []byte { return testutil.EncryptForServer(t, verifyToken) },
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := newLoginClient(t, &testutil.FakeSessionServer{})

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
	sessionServer := &testutil.FakeSessionServer{Profile: authenticated}

	client, _ := newLoginClient(t, sessionServer)
	client.SetProfile(types.GameProfile{Username: "Notch"})

	_, verifyToken, err := client.BeginEncryption()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	secret := []byte("0123456789abcdef")

	if err := client.CompleteEncryption(testutil.EncryptForServer(t, secret), testutil.EncryptForServer(t, verifyToken)); err != nil {
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
	if !slices.Equal(sessionServer.Usernames, []string{"Notch"}) {
		t.Errorf("asked about %v, want the name the client logged in under", sessionServer.Usernames)
	}

	want := auth.ServerHash(serverId, secret, testutil.KeyPair().PublicKey())
	if !slices.Equal(sessionServer.Hashes, []string{want}) {
		t.Errorf("asked about %v, want the hash over this login %q", sessionServer.Hashes, want)
	}
}

// The phases before play are ones a player is still waiting behind, so only the
// move into play joins the count, only the first such move joins it, and only a
// connection that ended in play leaves it.
func TestSetPhaseCountsTheMoveIntoPlayExactlyOnce(t *testing.T) {
	status := new(fakeStatus)
	joining := &Client{protocolVersion: types.ProtocolVersions.MINECRAFT_26_2, status: status}

	joining.SetPhase(types.PhaseLogin)
	joining.SetPhase(types.PhaseConfiguration)

	if joins, _ := status.counts(); joins != 0 {
		t.Errorf("joins = %d, want none while the client is still arriving", joins)
	}

	joining.SetPhase(types.PhasePlay)

	if joins, _ := status.counts(); joins != 1 {
		t.Errorf("joins = %d, want the arrival in play counted", joins)
	}

	// Being put into the phase it is already in is not a second player.
	joining.SetPhase(types.PhasePlay)

	if joins, _ := status.counts(); joins != 1 {
		t.Errorf("joins = %d, want a phase that did not change to count nothing", joins)
	}

	joining.leavePlay()

	if _, leaves := status.counts(); leaves != 1 {
		t.Errorf("leaves = %d, want the ended connection counted out", leaves)
	}

	// A connection that ended without ever joining leaves nothing behind.
	// Counting it out would make every ping after it report fewer players than
	// are there.
	pinging := &Client{protocolVersion: types.ProtocolVersions.MINECRAFT_26_2, status: status}
	pinging.leavePlay()

	if _, leaves := status.counts(); leaves != 1 {
		t.Errorf("leaves = %d, want a connection that never joined to leave nothing", leaves)
	}
}

// A join is the packet the two versions disagree about, so what a 26.1 client is
// sent has to be what a 26.2 client is sent with the online mode flag taken back
// out of it.
func TestWritePacketCarriesTheBodyDownToTheClientVersion(t *testing.T) {
	join := func() *clientboundPlay.LoginClientboundPacket {
		return &clientboundPlay.LoginClientboundPacket{
			EntityId:           1,
			Dimensions:         []string{"minecraft:overworld"},
			ViewDistance:       2,
			SimulationDistance: 2,
			ShowDeathScreen:    true,
			OnlineMode:         true,
			SpawnInfo: clientboundPlay.SpawnInfo{
				Dimension:        "minecraft:overworld",
				GameMode:         clientboundPlay.GameModeSpectator,
				PreviousGameMode: clientboundPlay.GameModeNone,
			},
		}
	}

	latest, latestBuf := newTestClient(types.PhasePlay)
	if err := latest.WritePacket(join()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	older, olderBuf := newTestClientOn(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_1)
	if err := older.WritePacket(join()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latestId, latestBody := testutil.IdAndBody(t, latestBuf.Bytes())
	olderId, olderBody := testutil.IdAndBody(t, olderBuf.Bytes())

	// Both versions number the join the same, so the id is not what changes.
	if latestId != 0x31 || olderId != 0x31 {
		t.Errorf("join went out as %#02x to 26.2 and %#02x to 26.1, want %#02x to both", latestId, olderId, 0x31)
	}

	if len(olderBody) != len(latestBody)-1 {
		t.Fatalf("26.1 body is %d bytes and 26.2's is %d, want exactly one fewer", len(olderBody), len(latestBody))
	}

	// The flag is the second to last byte, and enforces secure chat behind it is
	// what a 26.1 client reads in its place.
	want := append(append([]byte{}, latestBody[:len(latestBody)-2]...), latestBody[len(latestBody)-1])
	if !bytes.Equal(olderBody, want) {
		t.Errorf("26.1 body is\n%v\nwant\n%v", olderBody, want)
	}
}

// A packet neither version changed goes out to both untouched, which is all of
// them but the join.
func TestWritePacketLeavesAPacketNeitherVersionChanged(t *testing.T) {
	latest, latestBuf := newTestClient(types.PhaseConfiguration)
	if err := latest.WritePacket(&clientboundCommon.KeepAliveClientboundPacket{Id: 7}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	older, olderBuf := newTestClientOn(types.PhaseConfiguration, types.ProtocolVersions.MINECRAFT_26_1)
	if err := older.WritePacket(&clientboundCommon.KeepAliveClientboundPacket{Id: 7}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(latestBuf.Bytes(), olderBuf.Bytes()) {
		t.Errorf("26.1 was sent\n% x\nand 26.2\n% x", olderBuf.Bytes(), latestBuf.Bytes())
	}
}

// What a 26.1 client sends is resolved through 26.1's id table and decoded by
// the one decoder the packet has, which belongs to 26.2.
func TestReadPacketResolvesIdsAtTheClientVersion(t *testing.T) {
	client, buf := newTestClientOn(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_1)

	body := new(bytes.Buffer)
	bodyStream := streams.NewMinecraftStreamFromBuffer(body)

	// The id play gives the keep alive, which is the same on both versions.
	if err := bodyStream.WriteVarInt(0x1C); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := bodyStream.WriteLong(99); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := bodyStream.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	frame := streams.NewMinecraftStreamFromBuffer(buf)
	if err := frame.WriteVarInt(int32(body.Len())); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := frame.WriteBytes(body.Bytes()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := frame.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	packet, handler, err := client.ReadPacket()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	keepAlive, ok := packet.(*serverboundCommon.KeepAliveServerboundPacket)
	if !ok {
		t.Fatalf("expected a keep alive, got %T", packet)
	}

	if keepAlive.Id != 99 {
		t.Errorf("expected id 99, got %d", keepAlive.Id)
	}

	if handler == nil {
		t.Error("expected the keep alive handler, got nil")
	}
}

// A 26.1 client has no field for the session id 26.2 appended to the login
// success packet, and reads a packet longer than it expects as a connection to
// drop. It is the one packet of the login phase the two versions disagree about.
func TestWritePacketDropsTheSessionIdForAnOlderClient(t *testing.T) {
	signature := "signed"

	loginSuccess := func() *clientboundLogin.LoginSuccessClientboundPacket {
		return &clientboundLogin.LoginSuccessClientboundPacket{
			Profile: types.GameProfile{
				Uuid:       "01020304-0506-0708-090a-0b0c0d0e0f10",
				Username:   "Steve",
				Properties: []types.ProfileProperty{{Name: "textures", Value: "skin", Signature: &signature}},
			},
			SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20",
		}
	}

	latest, latestBuf := newTestClient(types.PhaseLogin)
	if err := latest.WritePacket(loginSuccess()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	older, olderBuf := newTestClientOn(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_1)
	if err := older.WritePacket(loginSuccess()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latestId, latestBody := testutil.IdAndBody(t, latestBuf.Bytes())
	olderId, olderBody := testutil.IdAndBody(t, olderBuf.Bytes())

	if latestId != 0x02 || olderId != 0x02 {
		t.Errorf("login success went out as %#02x to 26.2 and %#02x to 26.1, want %#02x to both", latestId, olderId, 0x02)
	}

	// The session id is a uuid on the end and nothing else moved, so what 26.1
	// is sent is what 26.2 is sent without its last sixteen bytes.
	want := latestBody[:len(latestBody)-16]
	if !bytes.Equal(olderBody, want) {
		t.Errorf("26.1 body is\n%v\nwant\n%v", olderBody, want)
	}
}
