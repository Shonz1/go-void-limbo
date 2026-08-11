package main

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	clientboundCommon "go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	serverboundCommon "go-void-limbo/packets/serverbound/common"
	"go-void-limbo/registries"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"io"
	"net"
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
