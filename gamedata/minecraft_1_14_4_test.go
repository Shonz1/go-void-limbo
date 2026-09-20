package gamedata

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// 1.14.4 knows its dimensions for itself as it knows its biomes, as 1.15.2
// does, so it is sent no registry: nothing in its play login, and the tags
// alone behind it, in the four unnamed runs 1.15.2 reads.
func TestProviderSends1_14_4TheTagsAlone(t *testing.T) {
	registries := mustLoadRegistries(t, registriesMinecraft1_14_4)
	if len(registries) != 0 {
		t.Fatalf("1.14.4 is sent %v, want no registry", registries)
	}

	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	version := types.ProtocolVersions.MINECRAFT_1_14_4

	if codec := provider.RegistryCodecFor(version); codec != nil {
		t.Errorf("1.14.4 has a registry codec of %d bytes, want none: its login holds a dimension's number", len(codec))
	}

	if dimensionType := provider.DimensionTypeFor(version); dimensionType != nil {
		t.Errorf("1.14.4 has a dimension type of %d bytes, want none", len(dimensionType))
	}

	packets := provider.PacketsFor(version)
	if len(packets) != 1 {
		t.Fatalf("1.14.4 is sent %d packets, want the tags alone", len(packets))
	}

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packets[0].Encode(out); err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	if err := out.Flush(); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	want, err := encodeTags1_16_4(mustLoadTags(t, tagsMinecraft1_14_4))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Error("1.14.4's tags are not in the unnamed shape it reads")
	}

	newer, err := encodeTags1_16_4(mustLoadTags(t, tagsMinecraft1_15_2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if bytes.Equal(buf.Bytes(), newer) {
		t.Error("1.14.4 is sent 1.15.2's tags, want its own jar's")
	}
}

// 1.14.3, 1.14.2, 1.14.1 and 1.14 are sent what 1.14.4 is, to the byte: the five
// jars' tags are files of the same names and their blocks reports the same
// bytes, so nothing here is their own, and they are sent no registry either.
func TestProviderSendsTheOlder1_14sWhat1_14_4Is(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	for _, older := range []types.ProtocolVersion{types.ProtocolVersions.MINECRAFT_1_14, types.ProtocolVersions.MINECRAFT_1_14_1, types.ProtocolVersions.MINECRAFT_1_14_2, types.ProtocolVersions.MINECRAFT_1_14_3} {
		t.Run(older.Names[0], func(t *testing.T) {
			sendsTheSetOf(t, provider, older, types.ProtocolVersions.MINECRAFT_1_14_4)

			states, err := BlockStatesFor(older)
			if err != nil {
				t.Fatalf("BlockStatesFor() error: %v", err)
			}

			// The bell 1.14.4 numbers without its power, where 1.15 has 11209.
			stored := map[string]string{"attachment": "ceiling", "facing": "south", "powered": "false"}

			if id, ok := states.Id("minecraft:bell", stored); !ok || id != 11203 {
				t.Errorf("a bell hung from the ceiling = %d, %t, want 1.14.4's 11203", id, ok)
			}
		})
	}
}

// 1.14.4's tags are its own jar's: 1.15's less what came with the bees and a
// handful beside, in the same four runs, and with the one tag 1.15 retired.
func TestTagsFor1_14_4AreItsOwnJars(t *testing.T) {
	older, newer := mustLoadTags(t, tagsMinecraft1_14_4), mustLoadTags(t, tagsMinecraft1_15_2)

	if len(older) != len(newer) {
		t.Fatalf("1.14.4 declares %d tag sets and 1.15.2 %d, want the same four", len(older), len(newer))
	}

	wantCounts := []int{50, 39, 2, 2}

	// The tags of 1.14.4's that did not last into 1.15.
	retired := map[string]bool{"minecraft:block minecraft:dirt_like": true}

	for i, set := range older {
		if set.Registry != newer[i].Registry {
			t.Errorf("set %d is %s on 1.14.4 and %s on 1.15.2, want the same order", i, set.Registry, newer[i].Registry)
		}

		if len(set.Tags) != wantCounts[i] {
			t.Errorf("1.14.4 declares %d %s tags, want %d", len(set.Tags), set.Registry, wantCounts[i])
		}

		newerNames := map[string]bool{}
		for _, tag := range newer[i].Tags {
			newerNames[tag.Name] = true
		}

		for _, tag := range set.Tags {
			if !newerNames[tag.Name] && !retired[set.Registry+" "+tag.Name] {
				t.Errorf("only 1.14.4 declares %s %s, want every tag of its but the dirt-like to have lasted into 1.15", set.Registry, tag.Name)
			}
		}
	}
}

// 1.14.4 is from before the bees and from before a bell could be powered: it
// numbers 11,271 states, and a bell stored the way 1.15 stores one is the bell
// it was stored as.
func TestBlockStatesFor1_14_4NumberABellWithoutItsPower(t *testing.T) {
	states, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_1_14_4)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	if states.StateCount() != 11271 {
		t.Errorf("1.14.4 numbers %d states, want 11271", states.StateCount())
	}

	// From the jar's own blocks report.
	if id, ok := states.Id("minecraft:dirt_path", nil); !ok || id != 8687 {
		t.Errorf("the dirt path = %d, %t, want the grass path's 8687", id, ok)
	}

	if _, ok := states.Id("minecraft:honey_block", nil); ok {
		t.Error("1.14.4 numbers the honey block, which 1.15 added")
	}

	if id, ok := states.DefaultId("minecraft:bell"); !ok || id != 11198 {
		t.Errorf("the default bell = %d, %t, want 11198", id, ok)
	}

	stored := map[string]string{"attachment": "ceiling", "facing": "south", "powered": "false"}

	if id, ok := states.Id("minecraft:bell", stored); !ok || id != 11203 {
		t.Errorf("a bell hung from the ceiling = %d, %t, want 11203", id, ok)
	}

	newer, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_1_15)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	if id, ok := newer.Id("minecraft:bell", stored); !ok || id != 11209 {
		t.Errorf("the same bell on 1.15 = %d, %t, want 11209", id, ok)
	}

	// A wall keeps the sides it was stored with, as it does on 1.15.2.
	sides := map[string]string{"east": "low", "north": "none", "south": "tall", "up": "true", "waterlogged": "false", "west": "none"}

	if id, ok := states.Id("minecraft:cobblestone_wall", sides); !ok || id == 5700 {
		t.Errorf("a wall with two sides = %d, %t, want a state of its own and not the bare post", id, ok)
	}
}
