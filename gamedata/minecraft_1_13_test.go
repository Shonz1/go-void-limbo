package gamedata

import (
	"slices"
	"testing"

	"github.com/Shonz1/go-void-limbo/types"
)

// 1.13's tags are 1.13.1's less the two block tags 1.13.1 added, in the same
// three runs: its items and fluids are the same names, and no tag of 1.13's
// was retired on the way.
func TestProviderSends1_13TheTagsOfItsOwnJar(t *testing.T) {
	older, newer := mustLoadTags(t, tagsMinecraft1_13), mustLoadTags(t, tagsMinecraft1_13_2)
	if len(older) != len(newer) {
		t.Fatalf("1.13 has tags for %d registries, want %d as 1.13.1", len(older), len(newer))
	}

	added := []string{"minecraft:coral_plants", "minecraft:underwater_bonemeals"}

	for i := range older {
		if older[i].Registry != newer[i].Registry {
			t.Fatalf("set %d is for %s, want %s", i, older[i].Registry, newer[i].Registry)
		}

		var olderNames, newerNames []string
		for _, tag := range older[i].Tags {
			olderNames = append(olderNames, tag.Name)
		}

		for _, tag := range newer[i].Tags {
			if older[i].Registry == "minecraft:block" && slices.Contains(added, tag.Name) {
				continue
			}

			newerNames = append(newerNames, tag.Name)
		}

		if !slices.Equal(olderNames, newerNames) {
			t.Errorf("%s: 1.13 has %v, want 1.13.1's %v", older[i].Registry, olderNames, newerNames)
		}
	}

	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	version := types.ProtocolVersions.MINECRAFT_1_13

	if codec := provider.RegistryCodecFor(version); len(codec) != 0 {
		t.Errorf("1.13 has a registry codec of %d bytes, want none: its login holds a dimension's number", len(codec))
	}

	if packets := provider.PacketsFor(version); len(packets) != 1 {
		t.Errorf("1.13 is sent %d packets, want the tags alone", len(packets))
	}
}

// 1.13 numbers its states without the dead coral plants 1.13.1 added, and a
// world's TNT, stored with the instability 1.13.1 gave it, is 1.13's TNT all
// the same.
func TestBlockStatesFor1_13LackTheDeadCoralPlants(t *testing.T) {
	older, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_1_13)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	newer, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_1_13_1)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	coral := map[string]string{"waterlogged": "true"}

	if id, ok := older.Id("minecraft:dead_tube_coral", coral); ok {
		t.Errorf("a dead tube coral = %d on 1.13, want none: 1.13.1 added it", id)
	}

	if _, ok := newer.Id("minecraft:dead_tube_coral", coral); !ok {
		t.Error("1.13.1 has no dead tube coral, want the plant its jar added")
	}

	if id, ok := older.Id("minecraft:tnt", map[string]string{"unstable": "false"}); !ok || id <= 0 {
		t.Errorf("a stored TNT = %d, %t on 1.13, want 1.13's TNT", id, ok)
	}
}
