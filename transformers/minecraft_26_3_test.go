package transformers

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// encodeClientbound is what a packet writes at 26.3, which is the only
// version it knows how to be and the input every downgrade starts from.
func encodeClientbound(t *testing.T, packet types.ClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

func encodePlayLogin(t *testing.T, packet *play.LoginClientboundPacket) []byte {
	t.Helper()

	return encodeClientbound(t, packet)
}

// The previous game mode is the one field the step rewrites, and it sits eight
// bytes from the end of every login: the two flags the debug, the flat, the
// death location, the portal cooldown and the sea level, then the online mode
// and enforces secure chat. Every login here has no death location and a
// cooldown and sea level of one byte each, which is what keeps that offset
// fixed.
const previousGameModeFromEnd = 8

func TestDowngradePlayLoginTo26_2RespellsThePreviousGameMode(t *testing.T) {
	cases := []struct {
		name     string
		previous types.GameMode
		want     byte
	}{
		{"none, which 26.3 spells as zero and 26.2 as minus one", types.GameModeNone, 0xFF},
		{"survival, one above its number on 26.3 and the number itself on 26.2", types.GameModeSurvival, 0x00},
		{"spectator", types.GameModeSpectator, 0x03},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			packet := play.LoginClientboundPacket{
				EntityId:           7,
				Dimensions:         []string{"minecraft:overworld", "minecraft:the_nether"},
				MaxPlayers:         20,
				ViewDistance:       10,
				SimulationDistance: 10,
				OnlineMode:         true,
				EnforcesSecureChat: true,
				SpawnInfo: play.SpawnInfo{
					DimensionTypeId:  3,
					Dimension:        "minecraft:the_nether",
					HashedSeed:       -1,
					GameMode:         types.GameModeCreative,
					PreviousGameMode: c.previous,
					IsFlat:           true,
					SeaLevel:         63,
				},
			}

			sent := encodePlayLogin(t, &packet)
			got := runTransformer(t, DowngradePlayLoginTo26_2, sent)

			if len(got) != len(sent) {
				t.Fatalf("the body is %d bytes, want the %d sent: the step changes a value, not a width", len(got), len(sent))
			}

			want := append([]byte{}, sent...)
			want[len(want)-previousGameModeFromEnd] = c.want

			if !bytes.Equal(got, want) {
				t.Errorf("to 26.2 = % x\nwant = % x", got, want)
			}
		})
	}
}

// The death location is the one optional field in front of the end, so a login
// carrying one is the case where walking the body field by field matters.
func TestDowngradePlayLoginTo26_2WalksPastADeathLocation(t *testing.T) {
	packet := play.LoginClientboundPacket{
		EntityId:   1,
		Dimensions: []string{"minecraft:overworld"},
		SpawnInfo: play.SpawnInfo{
			Dimension:        "minecraft:overworld",
			GameMode:         types.GameModeSurvival,
			PreviousGameMode: types.GameModeAdventure,
			DeathLocation: &play.GlobalPos{
				Dimension: "minecraft:overworld",
				Position:  play.BlockPos{X: 1, Y: 2, Z: 3},
			},
		},
	}

	sent := encodePlayLogin(t, &packet)
	got := runTransformer(t, DowngradePlayLoginTo26_2, sent)

	// The death location is a string of 19 characters behind its length byte
	// and a packed long, so the previous mode sits that much further from
	// the end.
	offset := previousGameModeFromEnd + 1 + 19 + 8

	want := append([]byte{}, sent...)
	if sent[len(sent)-offset] != 0x03 {
		t.Fatalf("26.3 spells adventure as %#02x at the offset, want 0x03: the test has the field wrong", sent[len(sent)-offset])
	}

	want[len(want)-offset] = 0x02

	if !bytes.Equal(got, want) {
		t.Errorf("to 26.2 = % x\nwant = % x", got, want)
	}
}

func TestDowngradePlayLoginTo26_2RefusesATruncatedBody(t *testing.T) {
	sent := encodePlayLogin(t, &play.LoginClientboundPacket{
		EntityId:   1,
		Dimensions: []string{"minecraft:overworld"},
		SpawnInfo:  play.SpawnInfo{Dimension: "minecraft:overworld"},
	})

	if err := failingTransformer(t, DowngradePlayLoginTo26_2, sent[:len(sent)/2]); err == nil {
		t.Error("expected a truncated body to be refused")
	}
}

func TestDowngradeAddEntityTo26_2RenumbersThePlayer(t *testing.T) {
	body := encodeAddEntity(t, testAddEntity)

	got := runTransformer(t, DowngradeAddEntityTo26_2, body)

	// 159 and 156 are both two byte var ints, so the body is the original
	// with the two swapped.
	want := append(append([]byte{}, body[:17]...), 0x9C, 0x01) // 156
	want = append(want, body[19:]...)

	if !bytes.Equal(got, want) {
		t.Errorf("to 26.2 = % x, want % x", got, want)
	}
}

func TestDowngradeEntityPositionSyncTo26_2PutsTheDeltaBack(t *testing.T) {
	body := encodeEntityPositionSync(t, &play.EntityPositionSyncClientboundPacket{
		EntityId: 2,
		X:        0.5,
		Y:        64,
		Z:        -0.5,
		Yaw:      90,
		Pitch:    -90,
		OnGround: true,
	})

	got := runTransformer(t, DowngradeEntityPositionSyncTo26_2, body)

	want := []byte{0x02}
	want = append(want,
		0x3F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x, 0.5
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y, 64
		0xBF, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z, -0.5
	)
	want = append(want, make([]byte, 24)...) // the delta, three zero doubles
	want = append(want,
		0x42, 0xB4, 0x00, 0x00, // yaw, 90
		0xC2, 0xB4, 0x00, 0x00, // pitch, -90
		0x01, // on ground
	)

	if !bytes.Equal(got, want) {
		t.Errorf("to 26.2 = % x\nwant = % x", got, want)
	}
}

func TestDowngradeEntityPositionSyncTo26_2RefusesASteppedPath(t *testing.T) {
	// The packet's own encoder only writes a straight line, so the body with
	// the other kind of path is built by hand: the id, then a type of 1.
	body := []byte{0x02, 0x01}
	body = append(body, make([]byte, 24+8+1)...)

	err := failingTransformer(t, DowngradeEntityPositionSyncTo26_2, body)
	if err == nil || !strings.Contains(err.Error(), "path of type 1") {
		t.Errorf("error = %v, want a refusal naming the path type", err)
	}
}

func encodeSwingAnimation(t *testing.T, packet *play.SwingAnimationClientboundPacket) []byte {
	t.Helper()

	return encodeClientbound(t, packet)
}

// A 26.2 animate is the entity and one byte naming the animation, 0 for the
// main arm and 3 for the offhand; the kind of swing and its duration have
// nowhere to go.
func TestDowngradeSwingAnimationTo26_2IsAnAnimate(t *testing.T) {
	cases := []struct {
		name    string
		offHand bool
		want    []byte
	}{
		{"main arm", false, []byte{0x80, 0x01, 0x00}},
		{"offhand", true, []byte{0x80, 0x01, 0x03}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := encodeSwingAnimation(t, &play.SwingAnimationClientboundPacket{
				EntityId:  128,
				OffHand:   c.offHand,
				Animation: play.SwingAnimationWhack,
				Duration:  play.DefaultSwingDuration,
			})

			if got := runTransformer(t, DowngradeSwingAnimationTo26_2, body); !bytes.Equal(got, c.want) {
				t.Errorf("to 26.2 = % x, want % x", got, c.want)
			}
		})
	}
}

func TestDowngradeSwingAnimationTo26_2RefusesATruncatedBody(t *testing.T) {
	// The entity and the hand, with the kind and the duration missing.
	if err := failingTransformer(t, DowngradeSwingAnimationTo26_2, []byte{0x02, 0x00}); err == nil {
		t.Error("expected a truncated body to be refused")
	}
}

// A 26.2 swing is the hand as a var int; a 26.3 punch is nothing at all.
func TestUpgradePunchFrom26_2DropsTheHand(t *testing.T) {
	for _, hand := range []byte{0x00, 0x01} {
		if got := runTransformer(t, UpgradePunchFrom26_2, []byte{hand}); len(got) != 0 {
			t.Errorf("hand %d came through as % x, want an empty body", hand, got)
		}
	}

	if err := failingTransformer(t, UpgradePunchFrom26_2, nil); err == nil {
		t.Error("expected a swing with no hand to be refused")
	}
}

// A 26.2 acknowledgement is the teleport id alone; 26.3 reads a position and a
// rotation behind it, which go in as zero.
func TestUpgradeAcceptTeleportationFrom26_2AppendsAPosition(t *testing.T) {
	got := runTransformer(t, UpgradeAcceptTeleportationFrom26_2, []byte{0x80, 0x01})

	want := []byte{0x80, 0x01}
	want = append(want, make([]byte, 24+8)...)

	if !bytes.Equal(got, want) {
		t.Errorf("to 26.3 = % x, want % x", got, want)
	}

	// What went in is a zero of each type, not merely zero bytes: a float
	// and a double of zero are all zero bits, which is the point.
	if math.Float64frombits(binary.BigEndian.Uint64(got[2:10])) != 0 || math.Float32frombits(binary.BigEndian.Uint32(got[26:30])) != 0 {
		t.Errorf("the position and rotation read back as something other than zero from % x", got)
	}

	if err := failingTransformer(t, UpgradeAcceptTeleportationFrom26_2, nil); err == nil {
		t.Error("expected an acknowledgement with no id to be refused")
	}
}

// A bit set is counted bytes to 26.3 and counted longs to 26.2, with the
// trailing bytes or longs that hold no bit left off either way.
func TestBitSetLongs(t *testing.T) {
	cases := []struct {
		name  string
		bytes []byte
		want  []int64
	}{
		{"nothing", nil, nil},
		{"one byte", []byte{0x20}, []int64{0x20}},
		{"the twenty-six sections of light", []byte{0xFF, 0xFF, 0xFF, 0x03}, []int64{0x3FFFFFF}},
		{"into a second long", append(make([]byte, 8), 0x01), []int64{0, 1}},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := bitSetLongs(c.bytes); !bytes.Equal(play.BitSetBytes(got), play.BitSetBytes(c.want)) || len(got) != len(c.want) {
				t.Errorf("bitSetLongs(% x) = %x, want %x", c.bytes, got, c.want)
			}

			if got := play.BitSetBytes(c.want); !bytes.Equal(got, c.bytes) {
				t.Errorf("BitSetBytes(%x) = % x, want % x", c.want, got, c.bytes)
			}
		})
	}
}

func encodeLightData() play.LightData {
	return play.LightData{
		SkyLightMask:        []int64{0b100000},
		BlockLightMask:      []int64{0},
		EmptySkyLightMask:   []int64{0b011111},
		EmptyBlockLightMask: []int64{0b111111},
		SkyLight:            [][]byte{{1, 2, 3}},
	}
}

// What 26.2 reads behind the masks is the 26.2 form of every mask -- a count
// of longs and the longs -- and then the light arrays untouched.
var lightDataTo26_2 = []byte{
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, // sky light mask
	0x00,                                                 // block light mask, a long of nothing left off
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1F, // empty sky light mask
	0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3F, // empty block light mask
	0x01, 0x03, 0x01, 0x02, 0x03, // one sky array of three bytes
	0x00, // no block arrays
}

func TestDowngradeLevelChunkWithLightTo26_2RewritesTheMasks(t *testing.T) {
	packet := &play.LevelChunkWithLightClientboundPacket{
		X: 1, Z: -1,
		Heightmaps:  []play.Heightmap{{Type: play.HeightmapMotionBlocking, Data: []int64{7, 8}}},
		SectionData: []byte{0xAA, 0xBB, 0xCC},
		LightData:   encodeLightData(),
	}

	got := runTransformer(t, DowngradeLevelChunkWithLightTo26_2, encodeClientbound(t, packet))

	want := []byte{
		0x00, 0x00, 0x00, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, // x and z
		0x01, 0x04, 0x02, // one heightmap of kind 4, two longs
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x07,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x08,
		0x03, 0xAA, 0xBB, 0xCC, // the sections
		0x00, // no block entities
	}
	want = append(want, lightDataTo26_2...)

	if !bytes.Equal(got, want) {
		t.Errorf("to 26.2 = % x\nwant = % x", got, want)
	}
}

func TestDowngradeLevelChunkWithLightTo26_2RefusesBlockEntities(t *testing.T) {
	body := encodeClientbound(t, &play.LevelChunkWithLightClientboundPacket{LightData: encodeLightData()})

	// The block entity count sits right behind the empty section buffer,
	// which is right behind the empty heightmap map: bytes 8, 9 and 10.
	body[10] = 0x01

	err := failingTransformer(t, DowngradeLevelChunkWithLightTo26_2, body)
	if err == nil || !strings.Contains(err.Error(), "block entities") {
		t.Errorf("error = %v, want a refusal naming the block entities", err)
	}
}

func TestDowngradeLightUpdateTo26_2RewritesTheMasks(t *testing.T) {
	packet := &play.LightUpdateClientboundPacket{X: 1, Z: 2, LightData: encodeLightData()}

	got := runTransformer(t, DowngradeLightUpdateTo26_2, encodeClientbound(t, packet))

	want := []byte{0x01, 0x02, 0x01} // x, z, trust edges
	want = append(want, lightDataTo26_2...)

	if !bytes.Equal(got, want) {
		t.Errorf("to 26.2 = % x\nwant = % x", got, want)
	}
}
