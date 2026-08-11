package main

import (
	"bytes"
	"encoding/binary"
	"errors"
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
