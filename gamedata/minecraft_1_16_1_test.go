package gamedata

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// 1.16.1 is sent the dimension types and nothing else, the biomes being its
// own, and its overworld is 1.16.4's as its codec asks for it: without the
// effects and the coordinate scale 1.16.2 added, and with the shrunk flag
// the scale replaced.
func TestRegistriesFor1_16_1AreTheDimensionTypesAlone(t *testing.T) {
	older, newer := mustLoadRegistries(t, registriesMinecraft1_16_1), mustLoadRegistries(t, registriesMinecraft1_16_4)

	if len(older) != 1 || older[0].Name != "minecraft:dimension_type" {
		t.Fatalf("1.16.1 is sent %v, want the dimension types alone", older)
	}

	overworld, ok := older[0].Entries[0].Data.(nbt.Compound)
	if !ok || older[0].Entries[0].Name != "minecraft:overworld" {
		t.Fatalf("1.16.1's first dimension type is %s, a %T, want the overworld as a compound", older[0].Entries[0].Name, older[0].Entries[0].Data)
	}

	newerOverworld := newer[0].Entries[0].Data.(nbt.Compound)

	for _, field := range []string{"effects", "coordinate_scale"} {
		if _, ok := overworld[field]; ok {
			t.Errorf("1.16.1's overworld carries %s, a field 1.16.2 added to the codec", field)
		}

		if _, ok := newerOverworld[field]; !ok {
			t.Errorf("1.16.4's overworld lacks %s, which this test takes 1.16.2 to have added", field)
		}
	}

	if overworld["shrunk"] != nbt.Byte(0) {
		t.Errorf("1.16.1's overworld is shrunk = %v, want the flag its codec requires, unset", overworld["shrunk"])
	}

	for field, value := range overworld {
		if field == "shrunk" {
			continue
		}

		if !reflect.DeepEqual(newerOverworld[field], value) {
			t.Errorf("1.16.1's overworld has %s = %v and 1.16.4's %v, want the same", field, value, newerOverworld[field])
		}
	}
}

// The play login of a 1.16.1 client carries the dimension types as a list
// under "dimension", each entry its name beside its fields, and names the
// one the player is put into, where 1.16.2 carries the combined compound
// and spells the dimension type out.
func TestProviderHands1_16_1ADimensionListAndAName(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	version := types.ProtocolVersions.MINECRAFT_1_16_1

	name, tag, err := nbt.ReadNamed(streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(provider.RegistryCodecFor(version))))
	if err != nil {
		t.Fatalf("reading 1.16.1's dimension types: %v", err)
	}

	want := nbt.Compound{"dimension": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{nbt.Compound{
		"name":                 nbt.String("minecraft:overworld"),
		"ambient_light":        nbt.Float(0),
		"bed_works":            nbt.Byte(1),
		"has_ceiling":          nbt.Byte(0),
		"has_raids":            nbt.Byte(1),
		"has_skylight":         nbt.Byte(1),
		"infiniburn":           nbt.String("minecraft:infiniburn_overworld"),
		"logical_height":       nbt.Int(256),
		"natural":              nbt.Byte(1),
		"piglin_safe":          nbt.Byte(0),
		"respawn_anchor_works": nbt.Byte(0),
		"shrunk":               nbt.Byte(0),
		"ultrawarm":            nbt.Byte(0),
	}}}}

	if name != "" || !reflect.DeepEqual(tag, want) {
		t.Errorf("1.16.1's dimension types are %q: %v\nwant a root named with the empty string: %v", name, tag, want)
	}

	ms := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(provider.DimensionTypeFor(version)))

	if dimensionType, err := ms.ReadString(); err != nil || dimensionType != "minecraft:overworld" {
		t.Errorf("1.16.1's dimension type is %q, %v, want the overworld's name", dimensionType, err)
	}

	if rest, _ := ms.ReadRest(); len(rest) != 0 {
		t.Errorf("1.16.1's dimension type holds %d bytes past its name", len(rest))
	}

	if bytes.Equal(provider.RegistryCodecFor(version), provider.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_16_2)) {
		t.Error("1.16.1 is handed 1.16.2's registry codec, want its own shape")
	}
}

// The list holds the dimension types and nothing else, so a set that holds
// another registry, or an entry with nothing to lay out, is refused rather
// than sent short.
func TestEncodeDimensionListRefusesWhatItCannotCarry(t *testing.T) {
	overworld := Entry{Name: "minecraft:overworld", Data: nbt.Compound{"shrunk": nbt.Byte(0)}}

	cases := map[string][]Registry{
		"no registry":      {},
		"no entry":         {{Name: "minecraft:dimension_type"}},
		"another registry": {{Name: "minecraft:dimension_type", Entries: []Entry{overworld}}, {Name: "minecraft:worldgen/biome"}},
		"no definition":    {{Name: "minecraft:dimension_type", Entries: []Entry{{Name: "minecraft:overworld"}}}},
		"a name field":     {{Name: "minecraft:dimension_type", Entries: []Entry{{Name: "minecraft:overworld", Data: nbt.Compound{"name": nbt.String("other")}}}}},
	}

	for name, registries := range cases {
		if _, err := encodeDimensionList(registries); err == nil {
			t.Errorf("%s: encodeDimensionList() succeeded, want a refusal", name)
		}
	}

	if _, err := encodeDimensionTypeName(nil); err == nil {
		t.Error("encodeDimensionTypeName() of no registries succeeded, want a refusal")
	}
}

// 1.16.1 reads its tags in the four unnamed runs 1.16.4 does, and is sent
// its own jar's: 1.16.2's but for five.
func TestTagsFor1_16_1AreItsOwnJars(t *testing.T) {
	older, newer := mustLoadTags(t, tagsMinecraft1_16_1), mustLoadTags(t, tagsMinecraft1_16_4)

	names := func(sets []TagSet) map[string]bool {
		all := map[string]bool{}
		for _, set := range sets {
			for _, tag := range set.Tags {
				all[set.Registry+" "+tag.Name] = true
			}
		}

		return all
	}

	olderNames, newerNames := names(older), names(newer)

	var onlyOlder, onlyNewer []string
	for name := range olderNames {
		if !newerNames[name] {
			onlyOlder = append(onlyOlder, name)
		}
	}

	for name := range newerNames {
		if !olderNames[name] {
			onlyNewer = append(onlyNewer, name)
		}
	}

	if len(onlyOlder) != 1 || onlyOlder[0] != "minecraft:item minecraft:furnace_materials" {
		t.Errorf("only 1.16.1 declares %v, want the furnace materials alone", onlyOlder)
	}

	if len(onlyNewer) != 4 {
		t.Errorf("only 1.16.4 declares %v, want the two base stones, the mushroom grow block and the stone crafting materials", onlyNewer)
	}

	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	packets := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_16_1)
	if len(packets) != 1 {
		t.Fatalf("1.16.1 is sent %d packets, want the tags alone", len(packets))
	}

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packets[0].Encode(out); err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	if err := out.Flush(); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	want, err := encodeTags1_16_4(older)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Error("1.16.1's tags are not in the unnamed shape it reads")
	}
}

// 751 is where the chain took on its axis and the lanterns their
// waterlogging, so 1.16.1 numbers eight states fewer and everything
// registered after the chain lower; a chain stored with an axis is numbered
// by the one property 1.16.1 knows.
func TestBlockStatesFor1_16_1NumberTheChainWithoutItsAxis(t *testing.T) {
	var loader BlockStatesLoader

	older, err := loader.For(types.ProtocolVersions.MINECRAFT_1_16_1)
	if err != nil {
		t.Fatalf("For() error: %v", err)
	}

	newer, err := loader.For(types.ProtocolVersions.MINECRAFT_1_16_2)
	if err != nil {
		t.Fatalf("For() error: %v", err)
	}

	if older.StateCount() != 17104 || newer.StateCount() != 17112 {
		t.Errorf("1.16.1 numbers %d states and 1.16.2 %d, want 17104 and 17112", older.StateCount(), newer.StateCount())
	}

	// From the jars' own blocks reports.
	if id, ok := older.Id("minecraft:chain", map[string]string{"axis": "x", "waterlogged": "false"}); !ok || id != 4730 {
		t.Errorf("1.16.1's dry chain = %d, %t, want 4730 whatever its axis", id, ok)
	}

	if olderId, _ := older.Id("minecraft:stone", nil); olderId != 1 {
		t.Errorf("1.16.1's stone = %d, want 1", olderId)
	}

	olderId, olderOk := older.Id("minecraft:dirt_path", nil)
	newerId, newerOk := newer.Id("minecraft:dirt_path", nil)

	// The grass path sits behind the chain and in front of the lanterns, so
	// it is lower by the four states the chain's axis added.
	if !olderOk || !newerOk || olderId != 9223 || newerId != 9227 {
		t.Errorf("the dirt path = %d, %t on 1.16.1 and %d, %t on 1.16.2, want the grass path's 9223 and 9227", olderId, olderOk, newerId, newerOk)
	}
}
