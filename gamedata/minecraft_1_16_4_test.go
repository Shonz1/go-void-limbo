package gamedata

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// 1.16.4 synchronizes the same two registries as 1.17, read off the two
// jars. The biome is 1.17.1's by reference, the codec being field for field
// the same, depth and scale included. The dimension type is its own: 1.17 is
// where the min_y and the height joined the codec, and 1.16.4 holds the
// logical height to the 256 blocks its world has.
func TestRegistriesFor1_16_4AreItsOwnAndNot1_17s(t *testing.T) {
	older, newer := mustLoadRegistries(t, registriesMinecraft1_16_4), mustLoadRegistries(t, registriesMinecraft1_17_1)

	if len(older) != 2 || older[0].Name != "minecraft:dimension_type" || older[1].Name != "minecraft:worldgen/biome" {
		t.Fatalf("1.16.4 is sent %v, want the dimension types, which the play login spells the first of out, and the biomes", older)
	}

	if !reflect.DeepEqual(older[1], newer[1]) {
		t.Error("1.16.4's biomes differ from 1.17.1's, want the same: the codec is field for field the same in the two")
	}

	overworld, ok := older[0].Entries[0].Data.(nbt.Compound)
	if !ok || older[0].Entries[0].Name != "minecraft:overworld" {
		t.Fatalf("1.16.4's first dimension type is %s, a %T, want the overworld as a compound", older[0].Entries[0].Name, older[0].Entries[0].Data)
	}

	newerOverworld := newer[0].Entries[0].Data.(nbt.Compound)

	for _, field := range []string{"min_y", "height"} {
		if _, ok := overworld[field]; ok {
			t.Errorf("1.16.4's overworld carries %s, a field 1.17 added to the codec", field)
		}

		if _, ok := newerOverworld[field]; !ok {
			t.Errorf("1.17.1's overworld lacks %s, which its client reads as a required field", field)
		}
	}

	if height, ok := overworld["logical_height"].(nbt.Int); !ok || height != 256 {
		t.Errorf("1.16.4's overworld has logical height %v, want 256: its codec refuses more", overworld["logical_height"])
	}

	for field, value := range overworld {
		if field != "logical_height" && !reflect.DeepEqual(value, newerOverworld[field]) {
			t.Errorf("1.16.4's overworld has %s = %v and 1.17.1's %v, want the same: the field has the same codec in the two", field, value, newerOverworld[field])
		}
	}

	if len(overworld) != len(newerOverworld)-2 {
		t.Errorf("1.16.4's overworld has %d fields and 1.17.1's %d, want two fewer for 1.16.4", len(overworld), len(newerOverworld))
	}
}

// 1.16.4 reads tags for four registries in an order it knows, so the set has
// to be those four in that order; within them it declares what its own jar
// does, which is less than 1.17's, and nothing 1.17's jar lacks but the one
// block 1.17 renamed.
func TestTagsFor1_16_4AreTheFourRegistriesInWireOrder(t *testing.T) {
	older, newer := mustLoadTags(t, tagsMinecraft1_16_4), mustLoadTags(t, tagsMinecraft1_17_1)

	if len(older) != len(unnamedTagRegistries) {
		t.Fatalf("1.16.4 declares tags for %d registries, want %d", len(older), len(unnamedTagRegistries))
	}

	newerNames := map[string]map[string]bool{}
	for _, set := range newer {
		newerNames[set.Registry] = map[string]bool{}
		for _, tag := range set.Tags {
			newerNames[set.Registry][tag.Name] = true
		}
	}

	wantCounts := []int{86, 54, 2, 5}

	for i, set := range older {
		if set.Registry != unnamedTagRegistries[i] {
			t.Errorf("1.16.4's tag set %d is for %s, want %s: the client tells the sets apart by position alone", i, set.Registry, unnamedTagRegistries[i])
		}

		if len(set.Tags) != wantCounts[i] {
			t.Errorf("1.16.4 declares %d %s tags, want the %d its jar holds", len(set.Tags), set.Registry, wantCounts[i])
		}

		for _, tag := range set.Tags {
			if !newerNames[set.Registry][tag.Name] {
				t.Errorf("1.16.4 declares the %s tag %s, which 1.17.1 does not: 1.17 retired no tag", set.Registry, tag.Name)
			}
		}
	}
}

// The tags 1.16.4 reads are the runs 1.17 reads with nothing in front of
// them: no count of registries and no registry names.
func TestEncodeTags1_16_4NamesNoRegistry(t *testing.T) {
	sets := []TagSet{
		{Registry: "minecraft:block", Tags: []NamedTag{{Name: "a:b", Entries: []int32{7}}}},
		{Registry: "minecraft:item"},
		{Registry: "minecraft:fluid", Tags: []NamedTag{{Name: "c:d"}}},
		{Registry: "minecraft:entity_type"},
	}

	got, err := encodeTags1_16_4(sets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{
		0x01, 0x03, 'a', ':', 'b', 0x01, 0x07, // the blocks: one tag of one entry
		0x00,                            // the items: none
		0x01, 0x03, 'c', ':', 'd', 0x00, // the fluids: one empty tag
		0x00, // the entity types: none
	}

	if !bytes.Equal(got, want) {
		t.Errorf("encoding mismatch.\n got: % X\nwant: % X", got, want)
	}

	sets[2], sets[3] = sets[3], sets[2]
	if _, err := encodeTags1_16_4(sets); err == nil {
		t.Error("expected the sets out of the order the client reads them in to be refused")
	}

	if _, err := encodeTags1_16_4(sets[:3]); err == nil {
		t.Error("expected three sets to be refused: the client reads four")
	}
}

// The provider hands a 1.16.4 client the tags in the shape it reads, and a
// 1.17 client the named one.
func TestDefaultProviderSends1_16_4TheUnnamedTags(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	body := func(version types.ProtocolVersion) []byte {
		packets := provider.PacketsFor(version)
		if len(packets) != 1 {
			t.Fatalf("protocol %d is sent %d packets, want the tags alone", version.ID, len(packets))
		}

		buf := new(bytes.Buffer)
		out := streams.NewMinecraftStreamFromBuffer(buf)

		if err := packets[0].Encode(out); err != nil {
			t.Fatalf("protocol %d: Encode() error: %v", version.ID, err)
		}

		if err := out.Flush(); err != nil {
			t.Fatalf("protocol %d: Flush() error: %v", version.ID, err)
		}

		return buf.Bytes()
	}

	wantOlder, err := encodeTags1_16_4(mustLoadTags(t, tagsMinecraft1_16_4))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(body(types.ProtocolVersions.MINECRAFT_1_16_4), wantOlder) {
		t.Error("1.16.4's tags are not in the unnamed shape it reads")
	}

	wantNewer, err := encodeTags(mustLoadTags(t, tagsMinecraft1_17_1))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(body(types.ProtocolVersions.MINECRAFT_1_17), wantNewer) {
		t.Error("1.17's tags are not in the named shape it reads")
	}

	if provider.DimensionTypeFor(types.ProtocolVersions.MINECRAFT_1_16_4) == nil || provider.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_16_4) == nil {
		t.Error("1.16.4 is handed no registries or no dimension type for its play login")
	}
}
