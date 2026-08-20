package gamedata

import (
	"testing"

	"github.com/Shonz1/go-void-limbo/types"
)

// The ids asserted here are read out of each version's own blocks report, the
// same report the tables are generated from, so what they really check is that
// the compact form reproduces the report it came from.

func TestBlockStatesForEverySupportedVersion(t *testing.T) {
	// How many states each version numbers, from its report. The count pins
	// the table to the version: a table missing a block, or holding another
	// version's, lands somewhere else.
	stateCounts := map[types.ProtocolId]int32{
		types.ProtocolVersions.MINECRAFT_1_21_6.ID:  27946,
		types.ProtocolVersions.MINECRAFT_1_21_7.ID:  27946,
		types.ProtocolVersions.MINECRAFT_1_21_9.ID:  29671,
		types.ProtocolVersions.MINECRAFT_1_21_11.ID: 29671,
		types.ProtocolVersions.MINECRAFT_26_1.ID:    29873,
		types.ProtocolVersions.MINECRAFT_26_2.ID:    32366,
	}

	for _, version := range types.SupportedProtocolVersions {
		states, err := BlockStatesFor(version)
		if err != nil {
			t.Fatalf("BlockStatesFor(%d) error: %v", version.ID, err)
		}

		if got, want := states.StateCount(), stateCounts[version.ID]; got != want {
			t.Errorf("protocol %d: StateCount() = %d, want %d", version.ID, got, want)
		}

		if id, ok := states.Id("minecraft:air", nil); !ok || id != 0 {
			t.Errorf("protocol %d: air = %d, %t, want 0, true", version.ID, id, ok)
		}
	}
}

func TestBlockStatesForUnknownVersion(t *testing.T) {
	if _, err := BlockStatesFor(types.ProtocolVersions.ZERO); err == nil {
		t.Fatal("BlockStatesFor(ZERO) did not report an error")
	}
}

func TestBlockStatesId(t *testing.T) {
	states, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_26_2)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	tests := []struct {
		name       string
		block      string
		properties map[string]string
		want       int32
	}{
		{name: "no properties", block: "minecraft:stone", want: 1},
		{name: "one property", block: "minecraft:grass_block", properties: map[string]string{"snowy": "false"}, want: 9},
		{name: "one property other value", block: "minecraft:grass_block", properties: map[string]string{"snowy": "true"}, want: 8},

		// The chest is one of the few blocks whose states do not walk its
		// properties in the order the report lists them, so it is the case
		// that catches a table generated in the wrong order.
		{
			name:       "report order differs from state order",
			block:      "minecraft:chest",
			properties: map[string]string{"type": "single", "facing": "north", "waterlogged": "true"},
			want:       3987,
		},

		{
			name:       "several properties",
			block:      "minecraft:spruce_leaves",
			properties: map[string]string{"distance": "1", "persistent": "true", "waterlogged": "false"},
			want:       281,
		},

		// A property the palette does not mention takes the default state's
		// value: an unmentioned type is single and unmentioned waterlogged is
		// false, which is the chest's default state.
		{name: "missing properties default", block: "minecraft:chest", properties: map[string]string{"facing": "north"}, want: 3988},
		{name: "no properties at all defaults", block: "minecraft:grass_block", want: 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := states.Id(tt.block, tt.properties)
			if !ok {
				t.Fatalf("Id(%s, %v) not found", tt.block, tt.properties)
			}

			if got != tt.want {
				t.Errorf("Id(%s, %v) = %d, want %d", tt.block, tt.properties, got, tt.want)
			}
		})
	}
}

func TestBlockStatesIdUnknown(t *testing.T) {
	states, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_26_2)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	if _, ok := states.Id("minecraft:no_such_block", nil); ok {
		t.Error("Id() resolved a block that does not exist")
	}

	if _, ok := states.Id("minecraft:grass_block", map[string]string{"snowy": "maybe"}); ok {
		t.Error("Id() resolved a property value that does not exist")
	}
}

func TestBlockStatesDefaultId(t *testing.T) {
	states, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_26_2)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	if id, ok := states.DefaultId("minecraft:chest"); !ok || id != 3988 {
		t.Errorf("DefaultId(chest) = %d, %t, want 3988, true", id, ok)
	}

	if _, ok := states.DefaultId("minecraft:no_such_block"); ok {
		t.Error("DefaultId() resolved a block that does not exist")
	}
}
