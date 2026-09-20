package gamedata

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// 1.15.2 knows its dimensions for itself as it knows its biomes, so it is
// sent no registry: nothing in its play login, and the tags alone behind it.
func TestProviderSends1_15_2TheTagsAlone(t *testing.T) {
	registries := mustLoadRegistries(t, registriesMinecraft1_15_2)
	if len(registries) != 0 {
		t.Fatalf("1.15.2 is sent %v, want no registry", registries)
	}

	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	version := types.ProtocolVersions.MINECRAFT_1_15_2

	if codec := provider.RegistryCodecFor(version); codec != nil {
		t.Errorf("1.15.2 has a registry codec of %d bytes, want none: its login holds a dimension's number", len(codec))
	}

	if dimensionType := provider.DimensionTypeFor(version); dimensionType != nil {
		t.Errorf("1.15.2 has a dimension type of %d bytes, want none", len(dimensionType))
	}

	packets := provider.PacketsFor(version)
	if len(packets) != 1 {
		t.Fatalf("1.15.2 is sent %d packets, want the tags alone", len(packets))
	}

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packets[0].Encode(out); err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	if err := out.Flush(); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	want, err := encodeTags1_16_4(mustLoadTags(t, tagsMinecraft1_15_2))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Error("1.15.2's tags are not in the unnamed shape it reads")
	}
}

// A set from before 1.16 has nowhere to put a registry, and says so rather
// than drop one.
func TestEncodeSetRefusesARegistryBefore1_16(t *testing.T) {
	set := Set{MinProtocol: types.ProtocolVersions.MINECRAFT_1_15_2.ID, Registries: mustLoadRegistries(t, registriesMinecraft1_16_1)}

	if _, err := encodeSet(set); err == nil {
		t.Error("expected a set with a registry for 1.15.2 to be refused")
	}
}

// 1.15.2's tags are its own jar's: 1.16's less what the nether update
// brought, in the same four runs, with none of its own that 1.16 retired.
func TestTagsFor1_15_2AreItsOwnJars(t *testing.T) {
	older, newer := mustLoadTags(t, tagsMinecraft1_15_2), mustLoadTags(t, tagsMinecraft1_16_1)

	if len(older) != len(newer) {
		t.Fatalf("1.15.2 declares %d tag sets and 1.16.1 %d, want the same four", len(older), len(newer))
	}

	wantCounts := []int{56, 42, 2, 4}

	for i, set := range older {
		if set.Registry != newer[i].Registry {
			t.Errorf("set %d is %s on 1.15.2 and %s on 1.16.1, want the same order", i, set.Registry, newer[i].Registry)
		}

		if len(set.Tags) != wantCounts[i] {
			t.Errorf("1.15.2 declares %d %s tags, want %d", len(set.Tags), set.Registry, wantCounts[i])
		}

		newerNames := map[string]bool{}
		for _, tag := range newer[i].Tags {
			newerNames[tag.Name] = true
		}

		for _, tag := range set.Tags {
			if !newerNames[tag.Name] {
				t.Errorf("only 1.15.2 declares %s %s, want every tag of its to have lasted into 1.16", set.Registry, tag.Name)
			}
		}
	}
}

// 1.15.2 is from before the nether update's blocks, and from before a wall's
// sides were low or tall: it numbers 11,337 states, and a wall stored the way
// 1.16 stores one has the sides it was stored with.
func TestBlockStatesFor1_15_2NumberAWallByWhetherItsSidesAreThere(t *testing.T) {
	states, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_1_15_2)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	if states.StateCount() != 11337 {
		t.Errorf("1.15.2 numbers %d states, want 11337", states.StateCount())
	}

	// From the jar's own blocks report.
	if id, ok := states.Id("minecraft:dirt_path", nil); !ok || id != 8687 {
		t.Errorf("the dirt path = %d, %t, want the grass path's 8687", id, ok)
	}

	if _, ok := states.Id("minecraft:soul_fire", nil); ok {
		t.Error("1.15.2 numbers the soul fire, which 1.16 added")
	}

	if id, ok := states.DefaultId("minecraft:cobblestone_wall"); !ok || id != 5700 {
		t.Errorf("the default cobblestone wall = %d, %t, want 5700", id, ok)
	}

	stored := map[string]string{"east": "low", "north": "none", "south": "tall", "up": "true", "waterlogged": "false", "west": "none"}
	named := map[string]string{"east": "true", "north": "false", "south": "true", "up": "true", "waterlogged": "false", "west": "false"}

	storedId, storedOk := states.Id("minecraft:cobblestone_wall", stored)
	namedId, namedOk := states.Id("minecraft:cobblestone_wall", named)

	if !storedOk || !namedOk || storedId != namedId || storedId == 5700 {
		t.Errorf("a wall with two sides = %d, %t as 1.16 stores it and %d, %t as 1.15.2 names it, want the same state and not the bare post", storedId, storedOk, namedId, namedOk)
	}

	// The versions that know the heights are not talked out of them.
	newer, err := BlockStatesFor(types.ProtocolVersions.MINECRAFT_1_16)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	if _, ok := newer.Id("minecraft:cobblestone_wall", named); ok {
		t.Error("1.16 numbers a wall whose sides are true or false, want it refused: its sides are none, low or tall")
	}
}

// 1.15.1 is sent what 1.15.2 is, to the byte: its jar's tags and its blocks
// report are 1.15.2's, so nothing here is its own, and it is sent no
// registry either.
func TestProviderSends1_15_1What1_15_2Is(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	older, newer := types.ProtocolVersions.MINECRAFT_1_15_1, types.ProtocolVersions.MINECRAFT_1_15_2

	if codec := provider.RegistryCodecFor(older); len(codec) != 0 {
		t.Errorf("1.15.1 has a registry codec of %d bytes, want none: its login holds a dimension's number", len(codec))
	}

	if dimensionType := provider.DimensionTypeFor(older); len(dimensionType) != 0 {
		t.Errorf("1.15.1 has a dimension type of %d bytes, want none", len(dimensionType))
	}

	olderPackets, newerPackets := provider.PacketsFor(older), provider.PacketsFor(newer)
	if len(olderPackets) != 1 || len(newerPackets) != 1 {
		t.Fatalf("1.15.1 is sent %d packets and 1.15.2 %d, want the tags alone for both", len(olderPackets), len(newerPackets))
	}

	encode := func(packet types.ClientboundPacket) []byte {
		buf := new(bytes.Buffer)
		out := streams.NewMinecraftStreamFromBuffer(buf)

		if err := packet.Encode(out); err != nil {
			t.Fatalf("Encode() error: %v", err)
		}

		if err := out.Flush(); err != nil {
			t.Fatalf("Flush() error: %v", err)
		}

		return buf.Bytes()
	}

	if !bytes.Equal(encode(olderPackets[0]), encode(newerPackets[0])) {
		t.Error("1.15.1's tags are not 1.15.2's")
	}

	var loader BlockStatesLoader

	olderStates, err := loader.For(older)
	if err != nil {
		t.Fatalf("1.15.1: For() error: %v", err)
	}

	newerStates, err := loader.For(newer)
	if err != nil {
		t.Fatalf("1.15.2: For() error: %v", err)
	}

	if olderStates.StateCount() != newerStates.StateCount() {
		t.Errorf("1.15.1 numbers %d states and 1.15.2 %d, want the same table", olderStates.StateCount(), newerStates.StateCount())
	}

	wall := map[string]string{"east": "low", "north": "none", "south": "tall", "up": "true", "waterlogged": "false", "west": "none"}

	for _, name := range []string{"minecraft:dirt_path", "minecraft:short_grass", "minecraft:water_cauldron", "minecraft:stone"} {
		olderId, olderOk := olderStates.Id(name, nil)
		newerId, newerOk := newerStates.Id(name, nil)

		if !olderOk || !newerOk || olderId != newerId {
			t.Errorf("Id(%s) = %d, %t on 1.15.1 and %d, %t on 1.15.2, want the same", name, olderId, olderOk, newerId, newerOk)
		}
	}

	olderId, olderOk := olderStates.Id("minecraft:cobblestone_wall", wall)
	newerId, newerOk := newerStates.Id("minecraft:cobblestone_wall", wall)

	if !olderOk || !newerOk || olderId != newerId {
		t.Errorf("a wall as 1.16 stores it = %d, %t on 1.15.1 and %d, %t on 1.15.2, want the same state", olderId, olderOk, newerId, newerOk)
	}
}
