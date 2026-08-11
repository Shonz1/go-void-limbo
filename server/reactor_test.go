package server

import (
	"encoding/binary"
	"io"
	"net"
	"testing"
	"time"

	"github.com/Shonz1/go-void-limbo/client"
	"github.com/Shonz1/go-void-limbo/protocol"
	"github.com/Shonz1/go-void-limbo/types"
)

// tcpPair opens a real TCP connection to itself, because a takeover needs a
// real descriptor: the in-memory pipes the other tests serve are exactly what
// the reactor declines.
func tcpPair(t *testing.T) (net.Conn, net.Conn) {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listening: %v", err)
	}
	defer listener.Close()

	peer, err := net.Dial("tcp", listener.Addr().String())
	if err != nil {
		t.Fatalf("dialling: %v", err)
	}

	serverConn, err := listener.Accept()
	if err != nil {
		peer.Close()
		t.Fatalf("accepting: %v", err)
	}

	t.Cleanup(func() {
		peer.Close()
		serverConn.Close()
	})

	return serverConn, peer
}

// A joined connection end to end on the reactor: taken over, sent a keep
// alive, answered across two writes the loop has to reassemble, and let go of
// -- with the player counted out and the registry emptied -- when the peer
// leaves.
func TestTheReactorServesAJoinedConnection(t *testing.T) {
	srv := &Server{packetRegistry: protocol.NewDefaultRegistry()}

	reactor, err := newReactor(srv)
	if err != nil {
		t.Skipf("no poller on this platform: %v", err)
	}

	srv.reactor = reactor

	// The loop runs for the life of the server, so here for the life of the
	// test process; one loop and one poller are all it leaves behind.
	go reactor.run()

	serverConn, peer := tcpPair(t)

	c := client.New(serverConn, client.Config{PacketRegistry: srv.packetRegistry, Status: &srv.status})
	c.SetProtocolVersion(types.ProtocolVersions.MINECRAFT_26_2)
	c.SetPhase(types.PhasePlay)

	srv.addClient(c)

	if !reactor.take(c, serverConn) {
		t.Fatal("the reactor declined a TCP connection")
	}

	// The server asks, the same way the sweep does.
	if err := c.SendKeepAlive(); err != nil {
		t.Fatalf("sending the keep alive: %v", err)
	}

	if err := peer.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The whole keep alive as the play phase frames it: a length, the packet
	// id and the eight byte id the answer has to carry back.
	frame := make([]byte, 10)
	if _, err := io.ReadFull(peer, frame); err != nil {
		t.Fatalf("reading the keep alive: %v", err)
	}

	if frame[1] != 0x2C {
		t.Fatalf("packet id = %#02x, want the play phase's keep alive %#02x", frame[1], 0x2C)
	}

	// The answer, deliberately split across two writes with a breath between
	// them, so the loop has to hold the head of the frame until the rest
	// arrives.
	answer := append([]byte{9, 0x1C}, frame[2:]...)

	if _, err := peer.Write(answer[:3]); err != nil {
		t.Fatalf("writing the answer: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	if _, err := peer.Write(answer[3:]); err != nil {
		t.Fatalf("writing the answer: %v", err)
	}

	// The answer landing is what lets the next keep alive out, and that keep
	// alive arriving at the peer is what proves the send was a send: a nil
	// from SendKeepAlive alone could also be the sweep's skip of a busy lock.
	answered := false

	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if err := c.SendKeepAlive(); err != nil {
			time.Sleep(10 * time.Millisecond)
			continue
		}

		if err := peer.SetReadDeadline(time.Now().Add(time.Second)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := io.ReadFull(peer, frame); err == nil {
			answered = true
			break
		}
	}

	if !answered {
		t.Fatal("the keep alive answer never made it through the reactor")
	}

	if id := int64(binary.BigEndian.Uint64(frame[2:])); id == 0 {
		t.Error("the second keep alive carries no id")
	}

	// A peer that leaves takes the connection with it: the reactor lets go,
	// the registry empties, and the player is counted out.
	peer.Close()

	left := false

	version := types.ProtocolVersions.MINECRAFT_26_2

	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline); {
		if len(srv.snapshotClients()) == 0 && srv.status.Status(version).Players.Online == 0 {
			left = true
			break
		}

		time.Sleep(10 * time.Millisecond)
	}

	if !left {
		t.Errorf("clients = %d online = %d, want the departed connection let go of",
			len(srv.snapshotClients()), srv.status.Status(version).Players.Online)
	}
}
