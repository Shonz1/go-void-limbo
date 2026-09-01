package play

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"testing"
)

func encode(t *testing.T, p interface {
	Encode(ms *streams.MinecraftStream) error
}) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := p.Encode(stream); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

func TestBlockPosPack(t *testing.T) {
	tests := []struct {
		name string
		pos  BlockPos
		want int64
	}{
		{name: "origin", pos: BlockPos{}, want: 0},
		{name: "one of each", pos: BlockPos{X: 1, Y: 2, Z: 3}, want: 1<<38 | 3<<12 | 2},
		// Every field is signed and packed on its own, so all ones stays all ones.
		{name: "negative", pos: BlockPos{X: -1, Y: -1, Z: -1}, want: -1},
		{name: "below zero y", pos: BlockPos{X: 0, Y: -64, Z: 0}, want: 0xFC0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.pos.pack(); got != test.want {
				t.Errorf("pack() = %#016x, want %#016x", got, test.want)
			}
		})
	}
}

func TestLoginClientboundPacketEncode(t *testing.T) {
	p := &LoginClientboundPacket{
		EntityId:           0x01020304,
		Hardcore:           true,
		Dimensions:         []string{"a:b", "c:d"},
		MaxPlayers:         20,
		ViewDistance:       2,
		SimulationDistance: 3,
		ReducedDebugInfo:   false,
		ShowDeathScreen:    true,
		DoLimitedCrafting:  false,
		SpawnInfo: SpawnInfo{
			DimensionTypeId:  1,
			Dimension:        "a:b",
			HashedSeed:       0x0102030405060708,
			GameMode:         types.GameModeSpectator,
			PreviousGameMode: types.GameModeAdventure,
			IsDebug:          true,
			IsFlat:           false,
			DeathLocation:    &GlobalPos{Dimension: "c:d", Position: BlockPos{X: 1, Y: 2, Z: 3}},
			PortalCooldown:   5,
			SeaLevel:         63,
		},
		OnlineMode:         false,
		EnforcesSecureChat: true,
	}

	want := []byte{
		0x01, 0x02, 0x03, 0x04, // entity id, a plain int
		0x01,                // hardcore
		0x02,                // two dimensions
		0x03, 'a', ':', 'b', //
		0x03, 'c', ':', 'd', //
		0x14,                // max players
		0x02,                // view distance
		0x03,                // simulation distance
		0x00,                // reduced debug info
		0x01,                // show death screen
		0x00,                // do limited crafting
		0x01,                // dimension type id
		0x03, 'a', ':', 'b', // dimension
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, // hashed seed
		0x03,                // spectator
		0x02,                // previous mode, adventure
		0x01,                // is debug
		0x00,                // is flat
		0x01,                // a death location follows
		0x03, 'c', ':', 'd', // its dimension
		0x00, 0x00, 0x00, 0x40, 0x00, 0x00, 0x30, 0x02, // its packed position
		0x05, // portal cooldown
		0x3f, // sea level
		0x00, // online mode
		0x01, // enforces secure chat
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

func TestLoginClientboundPacketEncodeWithoutDeathLocation(t *testing.T) {
	p := &LoginClientboundPacket{
		Dimensions: []string{"a:b"},
		SpawnInfo: SpawnInfo{
			Dimension:        "a:b",
			GameMode:         types.GameModeSpectator,
			PreviousGameMode: types.GameModeNone,
		},
	}

	want := []byte{
		0x00, 0x00, 0x00, 0x00, // entity id
		0x00,                      // hardcore
		0x01, 0x03, 'a', ':', 'b', // one dimension
		0x00,                // max players
		0x00,                // view distance
		0x00,                // simulation distance
		0x00,                // reduced debug info
		0x00,                // show death screen
		0x00,                // do limited crafting
		0x00,                // dimension type id
		0x03, 'a', ':', 'b', // dimension
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // hashed seed
		0x03, // spectator
		// The absent previous mode is a signed byte the client reads as none,
		// not a mode of its own.
		0xff,
		0x00, // is debug
		0x00, // is flat
		0x00, // no death location, and so nothing follows it
		0x00, // portal cooldown
		0x00, // sea level
		0x00, // online mode
		0x00, // enforces secure chat
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

func TestLoginClientboundPacketString(t *testing.T) {
	p := &LoginClientboundPacket{
		EntityId:     1,
		Dimensions:   []string{"minecraft:overworld"},
		ViewDistance: 2,
		SpawnInfo: SpawnInfo{
			Dimension:        "minecraft:overworld",
			GameMode:         types.GameModeSpectator,
			PreviousGameMode: types.GameModeNone,
		},
	}

	want := "LoginClientboundPacket{EntityId:1 Hardcore:false Dimensions:[minecraft:overworld] MaxPlayers:0 ViewDistance:2 SimulationDistance:0 ReducedDebugInfo:false ShowDeathScreen:false DoLimitedCrafting:false SpawnInfo:SpawnInfo{DimensionTypeId:0 Dimension:minecraft:overworld HashedSeed:0 GameMode:spectator PreviousGameMode:none IsDebug:false IsFlat:false DeathLocation:none PortalCooldown:0 SeaLevel:0} OnlineMode:false EnforcesSecureChat:false}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
