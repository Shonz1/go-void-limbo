package gamedata

import (
	"testing"

	"github.com/Shonz1/go-void-limbo/types"
)

// 1.12.2 is sent nothing: no registry, and no tags, which it has no packet
// for.
func TestProviderSends1_12_2Nothing(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	version := types.ProtocolVersions.MINECRAFT_1_12_2

	if codec := provider.RegistryCodecFor(version); len(codec) != 0 {
		t.Errorf("1.12.2 has a registry codec of %d bytes, want none", len(codec))
	}

	if packets := provider.PacketsFor(version); len(packets) != 0 {
		t.Errorf("1.12.2 is sent %d packets, want none: it has no tags packet", len(packets))
	}
}

// The numbers asserted here are the ones 1.12.2 registers each block and
// variant under -- its block id shifted up four bits, and the variant below --
// as its own block registry lists them.
func TestBlockStatesFor1_12_2AreItsBlockIdsAndVariants(t *testing.T) {
	states, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_1_12_2)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	cases := []struct {
		name       string
		properties map[string]string
		want       int32
	}{
		{"minecraft:air", nil, 0},
		{"minecraft:stone", nil, 1 << 4},
		{"minecraft:granite", nil, 1<<4 | 1},
		// Snow on the grass is worked out from the block above it.
		{"minecraft:grass_block", map[string]string{"snowy": "true"}, 2 << 4},
		{"minecraft:short_grass", nil, 31<<4 | 1},
		{"minecraft:water", map[string]string{"level": "0"}, 9 << 4},
		{"minecraft:spruce_log", map[string]string{"axis": "x"}, 17<<4 | 5},
		{"minecraft:oak_wood", map[string]string{"axis": "z"}, 17<<4 | 12},
		{"minecraft:oak_leaves", map[string]string{"distance": "3", "persistent": "true"}, 18<<4 | 4},
		// A door's lower half keeps where it faces and whether it is open,
		// its upper half its hinge and whether it is powered.
		{"minecraft:oak_door", map[string]string{"facing": "west", "half": "lower", "hinge": "right", "open": "true", "powered": "true"}, 64<<4 | 6},
		{"minecraft:oak_door", map[string]string{"facing": "west", "half": "upper", "hinge": "right", "open": "true", "powered": "true"}, 64<<4 | 11},
		// A stair's shape is worked out from the stairs beside it.
		{"minecraft:oak_stairs", map[string]string{"facing": "east", "half": "top", "shape": "outer_left"}, 53<<4 | 4},
		{"minecraft:smooth_stone_slab", map[string]string{"type": "top"}, 44<<4 | 8},
		{"minecraft:smooth_stone_slab", map[string]string{"type": "double"}, 43 << 4},
		{"minecraft:cobblestone_wall", map[string]string{"north": "low"}, 139 << 4},
		{"minecraft:tall_grass", map[string]string{"half": "lower"}, 175<<4 | 2},
		// Every tall plant's upper half is the same block and variants.
		{"minecraft:rose_bush", map[string]string{"half": "upper"}, 175<<4 | 8},
		// A bed's colour, a skull's kind and a pot's plant are block entities
		// on 1.12.2, and the block is the same for all of them.
		{"minecraft:white_bed", map[string]string{"facing": "north", "part": "head"}, 26<<4 | 10},
		{"minecraft:zombie_wall_head", map[string]string{"facing": "east"}, 144<<4 | 5},
		{"minecraft:player_head", nil, 144<<4 | 1},
		{"minecraft:potted_poppy", nil, 140 << 4},
		{"minecraft:spruce_trapdoor", map[string]string{"facing": "north", "half": "top", "open": "true"}, 96<<4 | 12},
		{"minecraft:stripped_birch_log", map[string]string{"axis": "z"}, 17<<4 | 10},
		{"minecraft:cave_air", nil, 0},
	}

	for _, c := range cases {
		if got, ok := states.Id(c.name, c.properties); !ok || got != c.want {
			t.Errorf("%s%v = %d, %t, want %d (block %d, variant %d)", c.name, c.properties, got, ok, c.want, c.want>>4, c.want&15)
		}
	}

	// What 1.13 added and 1.12.2 has nothing like is no number at all.
	for _, name := range []string{"minecraft:tube_coral_block", "minecraft:conduit", "minecraft:prismarine_stairs", "minecraft:turtle_egg"} {
		if id, ok := states.Id(name, nil); ok {
			t.Errorf("%s = %d on 1.12.2, want none", name, id)
		}

		if id, ok := states.DefaultId(name); ok {
			t.Errorf("%s defaults to %d on 1.12.2, want none", name, id)
		}
	}

	// A section too varied for a palette packs 1.12.2's ids at the bits its
	// 5,365 registered states take.
	if got := states.StateCount(); got != 5365 {
		t.Errorf("StateCount() = %d, want 5365", got)
	}
}
