package server

import (
	"net"
	"testing"
	"time"

	"github.com/Shonz1/go-void-limbo/gamedata"
	"github.com/Shonz1/go-void-limbo/internal/testutil"
	"github.com/Shonz1/go-void-limbo/protocol"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// The play phase packet ids at the version the peers below speak, 26.2.
const (
	addEntityId        = 0x01
	gameEventId        = 0x26
	playLoginId        = 0x31
	playerInfoRemoveId = 0x45
	playerInfoUpdateId = 0x46
	playerPositionId   = 0x48
	removeEntitiesId   = 0x4D
)

// joinPlay drives one connection all the way into the world: an offline login,
// the configuration exchange, and the join itself, read until the game event
// that ends it. What comes over the connection after this is about the other
// players, which is what the test below is about.
func joinPlay(t *testing.T, srv *Server, username string) *testutil.LoginPeer {
	t.Helper()

	conn, clientConn := net.Pipe()

	t.Cleanup(func() {
		conn.Close()
		clientConn.Close()
	})

	if err := clientConn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	go srv.handleConnection(conn)

	peer := &testutil.LoginPeer{T: t, Conn: clientConn}

	sendHandshake(t, peer, "localhost", types.ProtocolVersions.MINECRAFT_26_2.ID, int32(types.PhaseLogin))
	sendLoginStart(t, peer, username)
	readSetCompression(t, peer)
	readLoginSuccess(t, peer)

	// Login acknowledged, which moves the connection into configuration and
	// brings the registries and finish configuration back.
	peer.WritePacket([]byte{0x03})

	for {
		packet := peer.ReadPacket()

		packetId, err := packet.ReadVarInt()
		if err != nil {
			t.Fatalf("reading a configuration packet id: %v", err)
		}

		// Finish configuration, which the peer acknowledges to join.
		if packetId == 0x03 {
			break
		}
	}

	peer.WritePacket([]byte{0x03})

	// The join: the play login, the spawn teleport, the peer's own player list
	// entry, and the game event that says chunks are next -- of which a server
	// with no world has none, so the game event is the join's last packet.
	for _, want := range []types.PacketId{playLoginId, playerPositionId, playerInfoUpdateId, gameEventId} {
		packet := peer.ReadPacket()

		packetId, err := packet.ReadVarInt()
		if err != nil {
			t.Fatalf("reading a join packet id: %v", err)
		}

		if packetId != want {
			t.Fatalf("join packet id = %#02x, want %#02x", packetId, want)
		}
	}

	return peer
}

// readPacketId reads one packet and returns its id and the rest of its body,
// which is all the assertions below need.
func readPacketId(t *testing.T, peer *testutil.LoginPeer) (types.PacketId, *streams.MinecraftStream) {
	t.Helper()

	packet := peer.ReadPacket()

	packetId, err := packet.ReadVarInt()
	if err != nil {
		t.Fatalf("reading a packet id: %v", err)
	}

	return packetId, packet
}

// The whole of the feature through the real server: two connections log in,
// join, and are shown each other -- and the leaving of one reaches the other.
func TestTwoPlayersAreShownEachOther(t *testing.T) {
	gameData, err := gamedata.NewDefaultProvider()
	if err != nil {
		t.Fatalf("building the game data: %v", err)
	}

	srv := New(Config{
		PacketRegistry: protocol.NewDefaultRegistry(),
		GameData:       gameData,
		KeyPair:        testutil.KeyPair(),
		SessionServer:  &testutil.FakeSessionServer{},
	})

	alice := joinPlay(t, srv, "alice")
	bob := joinPlay(t, srv, "bob")

	// Bob was shown the player already there: the list entry, then the
	// entity. The reads interleave with alice's below because a pipe buffers
	// nothing, and bob's goroutine writes his own connection first.
	if packetId, _ := readPacketId(t, bob); packetId != playerInfoUpdateId {
		t.Fatalf("bob's first packet after joining = %#02x, want alice's list entry %#02x", packetId, playerInfoUpdateId)
	}

	packetId, addEntity := readPacketId(t, bob)
	if packetId != addEntityId {
		t.Fatalf("bob's second packet after joining = %#02x, want alice's entity %#02x", packetId, addEntityId)
	}

	bobSawEntity, err := addEntity.ReadVarInt()
	if err != nil {
		t.Fatalf("reading the entity id: %v", err)
	}

	// And alice was shown bob the same way.
	if packetId, _ := readPacketId(t, alice); packetId != playerInfoUpdateId {
		t.Fatalf("alice's packet after bob joined = %#02x, want bob's list entry %#02x", packetId, playerInfoUpdateId)
	}

	packetId, addEntity = readPacketId(t, alice)
	if packetId != addEntityId {
		t.Fatalf("alice's next packet = %#02x, want bob's entity %#02x", packetId, addEntityId)
	}

	aliceSawEntity, err := addEntity.ReadVarInt()
	if err != nil {
		t.Fatalf("reading the entity id: %v", err)
	}

	// Each saw the other, and the ids are the server's own counter: one per
	// connection, never shared.
	if bobSawEntity == aliceSawEntity {
		t.Errorf("both players were shown entity %d, want each connection an id of its own", bobSawEntity)
	}

	// Bob leaves; alice is told twice over, the body first.
	bob.Conn.Close()

	if packetId, _ := readPacketId(t, alice); packetId != removeEntitiesId {
		t.Fatalf("alice's packet after bob left = %#02x, want remove entities %#02x", packetId, removeEntitiesId)
	}

	if packetId, _ := readPacketId(t, alice); packetId != playerInfoRemoveId {
		t.Fatalf("alice's next packet = %#02x, want player info remove %#02x", packetId, playerInfoRemoveId)
	}
}
