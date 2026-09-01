package client

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// The packet ids the play phase gives the entity packets at the latest
// version, which is what every client below speaks unless its test says
// otherwise.
const (
	addEntityId26_2        = 0x01
	animateId26_2          = 0x02
	positionSyncId26_2     = 0x23
	playerInfoRemoveId26_2 = 0x45
	playerInfoUpdateId26_2 = 0x46
	removeEntitiesId26_2   = 0x4D
	rotateHeadId26_2       = 0x53
	setEntityDataId26_2    = 0x63
)

// newSyncClient builds a joined player on the roster's server: in play, with
// an identity of its own, writing into the returned buffer.
func newSyncClient(ps *PlayerSync, entityId int32, username, uuid string) (*Client, *bytes.Buffer) {
	c, buf := newTestClient(types.PhasePlay)
	c.entityId = entityId
	c.playerSync = ps
	c.profile = types.GameProfile{Uuid: uuid, Username: username}

	return c, buf
}

// frames splits everything a client was sent into packet bodies, one per
// frame, each starting with its packet id.
func frames(t *testing.T, buf *bytes.Buffer) [][]byte {
	t.Helper()

	var out [][]byte

	rest := buf.Bytes()
	for len(rest) > 0 {
		size, sizeLen, err := streams.ReadVarIntFrom(rest)
		if err != nil {
			t.Fatalf("unreadable frame length in % x: %v", rest, err)
		}

		rest = rest[sizeLen:]
		if int(size) > len(rest) {
			t.Fatalf("frame claims %d bytes of the %d left", size, len(rest))
		}

		out = append(out, rest[:size])
		rest = rest[size:]
	}

	return out
}

// packetIds reduces frames to their leading id bytes, for the assertions that
// are about which packets went out rather than what was in them.
func packetIds(t *testing.T, buf *bytes.Buffer) []byte {
	t.Helper()

	var ids []byte
	for _, frame := range frames(t, buf) {
		ids = append(ids, frame[0])
	}

	return ids
}

const (
	uuidA = "11111111-1111-1111-1111-111111111111"
	uuidB = "22222222-2222-2222-2222-222222222222"
)

func TestJoinPlayerSyncShowsThePlayersToEachOther(t *testing.T) {
	ps := NewPlayerSync()

	a, aOut := newSyncClient(ps, 1, "Alice", uuidA)
	b, bOut := newSyncClient(ps, 2, "Bob", uuidB)

	// The first player in has nobody to be shown and nobody to be shown to.
	a.JoinPlayerSync()

	if aOut.Len() != 0 {
		t.Fatalf("a was sent % x joining an empty world, want nothing", aOut.Bytes())
	}

	b.JoinPlayerSync()

	// Each side gets the same pair about the other: the list entry first,
	// because the client refuses a player entity whose uuid it has no entry
	// for, then the entity.
	wantIds := []byte{playerInfoUpdateId26_2, addEntityId26_2}

	if got := packetIds(t, aOut); !bytes.Equal(got, wantIds) {
		t.Errorf("a was sent packets % x, want % x", got, wantIds)
	}

	if got := packetIds(t, bOut); !bytes.Equal(got, wantIds) {
		t.Errorf("b was sent packets % x, want % x", got, wantIds)
	}

	// The entity a was shown is b's, and the other way around.
	aFrames := frames(t, aOut)
	if got := aFrames[1][1]; got != 2 {
		t.Errorf("a was shown entity %d, want b's 2", got)
	}

	bFrames := frames(t, bOut)
	if got := bFrames[1][1]; got != 1 {
		t.Errorf("b was shown entity %d, want a's 1", got)
	}
}

func TestJoinPlayerSyncShowsAPlayerOnlyOnce(t *testing.T) {
	ps := NewPlayerSync()

	a, _ := newSyncClient(ps, 1, "Alice", uuidA)
	b, bOut := newSyncClient(ps, 2, "Bob", uuidB)

	a.JoinPlayerSync()
	b.JoinPlayerSync()

	sent := bOut.Len()

	// A join replayed -- which two racing connections can amount to -- must
	// not spawn anyone twice: the recipient's own record is what decides.
	b.showPlayer(a.syncSnapshot())

	if bOut.Len() != sent {
		t.Errorf("b was sent % x a second time, want nothing beyond the first showing", bOut.Bytes()[sent:])
	}
}

func TestMovesAreRelayedToThoseShown(t *testing.T) {
	ps := NewPlayerSync()

	a, aOut := newSyncClient(ps, 1, "Alice", uuidA)
	b, bOut := newSyncClient(ps, 2, "Bob", uuidB)

	a.JoinPlayerSync()
	b.JoinPlayerSync()
	aOut.Reset()
	bOut.Reset()

	// A move that turned carries the head along: the position packet steers
	// the body, and only the head packet turns what everyone watches.
	a.SyncPositionRotation(1, 65, 2, 90, 10, true)

	if got, want := packetIds(t, bOut), []byte{positionSyncId26_2, rotateHeadId26_2}; !bytes.Equal(got, want) {
		t.Errorf("b was sent packets % x after a turn, want % x", got, want)
	}

	if aOut.Len() != 0 {
		t.Errorf("a was sent % x for its own move, want nothing", aOut.Bytes())
	}

	bOut.Reset()

	// A move that did not turn moves the entity and nothing else.
	a.SyncPosition(2, 65, 2, true)

	if got, want := packetIds(t, bOut), []byte{positionSyncId26_2}; !bytes.Equal(got, want) {
		t.Errorf("b was sent packets % x after a straight move, want % x", got, want)
	}
}

func TestMovesAreDroppedWhereThePlayerWasNeverShown(t *testing.T) {
	ps := NewPlayerSync()

	a, _ := newSyncClient(ps, 1, "Alice", uuidA)
	c, cOut := newSyncClient(ps, 3, "Carol", "33333333-3333-3333-3333-333333333333")

	// c is on the roster but was never shown a -- the window a join is in the
	// middle of -- so a's moves have no entity on c's side to land on.
	ps.add(a)
	ps.add(c)

	a.SyncPosition(1, 65, 2, true)
	a.SyncSwing(false)
	a.SyncInput(true, false)

	if cOut.Len() != 0 {
		t.Errorf("c was sent % x about a player it was never shown, want nothing", cOut.Bytes())
	}
}

func TestLeaveHidesThePlayerEverywhere(t *testing.T) {
	ps := NewPlayerSync()

	a, aOut := newSyncClient(ps, 1, "Alice", uuidA)
	b, bOut := newSyncClient(ps, 2, "Bob", uuidB)

	a.JoinPlayerSync()
	b.JoinPlayerSync()
	aOut.Reset()
	bOut.Reset()

	b.LeavePlay()

	// The body goes first and the list entry after it, the reverse of the
	// spawn.
	if got, want := packetIds(t, aOut), []byte{removeEntitiesId26_2, playerInfoRemoveId26_2}; !bytes.Equal(got, want) {
		t.Errorf("a was sent packets % x after b left, want % x", got, want)
	}

	if bOut.Len() != 0 {
		t.Errorf("b was sent % x for its own leaving, want nothing", bOut.Bytes())
	}

	aOut.Reset()

	// Both loops that can see a connection end report it, and only the first
	// report says anything.
	b.LeavePlay()

	if aOut.Len() != 0 {
		t.Errorf("a was sent % x for a second leave, want nothing", aOut.Bytes())
	}

	// A move from the departed player lands on nobody.
	b.SyncPosition(1, 65, 2, true)

	if aOut.Len() != 0 {
		t.Errorf("a was sent % x about a player already hidden, want nothing", aOut.Bytes())
	}
}

func TestSyncInputRelaysOnlyChanges(t *testing.T) {
	ps := NewPlayerSync()

	a, _ := newSyncClient(ps, 1, "Alice", uuidA)
	b, bOut := newSyncClient(ps, 2, "Bob", uuidB)

	a.JoinPlayerSync()
	b.JoinPlayerSync()
	bOut.Reset()

	a.SyncInput(true, false)

	if got, want := packetIds(t, bOut), []byte{setEntityDataId26_2}; !bytes.Equal(got, want) {
		t.Errorf("b was sent packets % x after a sneak, want % x", got, want)
	}

	bOut.Reset()

	// The client resends its input byte for every key change, and most of the
	// bits are movement keys nobody else can see.
	a.SyncInput(true, false)

	if bOut.Len() != 0 {
		t.Errorf("b was sent % x for an input that changed no stance, want nothing", bOut.Bytes())
	}

	a.SyncInput(false, false)

	if got, want := packetIds(t, bOut), []byte{setEntityDataId26_2}; !bytes.Equal(got, want) {
		t.Errorf("b was sent packets % x after standing back up, want % x", got, want)
	}
}

func TestSyncSwingRelaysTheArm(t *testing.T) {
	ps := NewPlayerSync()

	a, _ := newSyncClient(ps, 1, "Alice", uuidA)
	b, bOut := newSyncClient(ps, 2, "Bob", uuidB)

	a.JoinPlayerSync()
	b.JoinPlayerSync()
	bOut.Reset()

	a.SyncSwing(false)
	a.SyncSwing(true)

	got := frames(t, bOut)
	if len(got) != 2 || got[0][0] != animateId26_2 || got[1][0] != animateId26_2 {
		t.Fatalf("b was sent % x, want two animate packets", bOut.Bytes())
	}

	// Each animate is the entity, then the animation: main arm 0, offhand 3.
	if got[0][2] != 0x00 || got[1][2] != 0x03 {
		t.Errorf("animations = %#x and %#x, want the main arm then the offhand", got[0][2], got[1][2])
	}
}

func TestShowPlayerCarriesTheStance(t *testing.T) {
	ps := NewPlayerSync()

	a, _ := newSyncClient(ps, 1, "Alice", uuidA)
	b, bOut := newSyncClient(ps, 2, "Bob", uuidB)

	a.JoinPlayerSync()

	// a is already sneaking by the time b arrives, with nobody yet to see it.
	a.SyncInput(true, false)

	b.JoinPlayerSync()

	want := []byte{playerInfoUpdateId26_2, addEntityId26_2, setEntityDataId26_2}
	if got := packetIds(t, bOut); !bytes.Equal(got, want) {
		t.Errorf("b was sent packets % x joining beside a sneaking player, want % x", got, want)
	}
}

// The one test here about versions: a 1.21.6 client is shown the same world
// through the transformers, which renumber the player's entity type, move the
// velocity back to the packet's end, and rename the pose serializer.
func TestSyncCarriesThePacketsDownToAnOlderClient(t *testing.T) {
	ps := NewPlayerSync()

	a, _ := newSyncClient(ps, 1, "Alice", uuidA)

	b, bOut := newTestClientOn(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_21_6)
	b.entityId = 2
	b.playerSync = ps
	b.profile = types.GameProfile{Uuid: uuidB, Username: "Bob"}

	a.SyncInput(true, false)
	a.JoinPlayerSync()
	b.JoinPlayerSync()

	got := frames(t, bOut)
	if len(got) != 3 {
		t.Fatalf("b was sent %d packets, want the entry, the entity and the stance", len(got))
	}

	// 1.21.6 numbers these packets differently, which is the registry's doing
	// rather than the transformers'.
	if got[0][0] != 0x3F || got[1][0] != 0x01 || got[2][0] != 0x5C {
		t.Fatalf("b was sent packet ids %#x %#x %#x, want 1.21.6's 0x3f 0x01 0x5c", got[0][0], got[1][0], got[2][0])
	}

	addEntity := got[1]

	// The entity id, the uuid, and then the type: 149, the player before three
	// versions of registry additions, where the latest says 156.
	if addEntity[18] != 0x95 || addEntity[19] != 0x01 {
		t.Errorf("entity type bytes = %#x %#x, want 1.21.6's player 149", addEntity[18], addEntity[19])
	}

	// The tail is the data var int and the three velocity shorts 1.21.6 reads
	// where the newer versions put a quantized vector up front.
	tail := addEntity[len(addEntity)-7:]
	if !bytes.Equal(tail, []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}) {
		t.Errorf("add entity tail = % x, want the data and three empty velocity shorts", tail)
	}

	// The stance names the pose serializer by 1.21.6's number for it: the
	// entity, the flag entry's three bytes, and then the pose entry.
	stance := got[2]
	if got, want := stance[6], byte(0x15); got != want {
		t.Errorf("pose serializer = %#x, want %#x: 1.21.6 still has the compound tag serializer before it", got, want)
	}
}

// The mode in the entry is the shown player's own, not anything of the
// recipient's or a constant: the fake below is put in a mode no default picks.
func TestShowPlayerCarriesThePlayersOwnGameMode(t *testing.T) {
	ps := NewPlayerSync()

	a, _ := newSyncClient(ps, 1, "Alice", uuidA)
	a.gameMode = types.GameModeAdventure

	b, bOut := newSyncClient(ps, 2, "Bob", uuidB)

	a.JoinPlayerSync()
	b.JoinPlayerSync()

	entry := frames(t, bOut)[0]
	if entry[0] != playerInfoUpdateId26_2 {
		t.Fatalf("b's first packet = %#02x, want the player info update %#02x", entry[0], playerInfoUpdateId26_2)
	}

	// The entry ends with the game mode, the listed flag and the hat flag.
	if gameMode := entry[len(entry)-3]; gameMode != byte(types.GameModeAdventure) {
		t.Errorf("game mode = %d, want alice's own adventure %d", gameMode, types.GameModeAdventure)
	}
}
