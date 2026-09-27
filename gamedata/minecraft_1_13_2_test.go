package gamedata

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// 1.13.2 is sent no registry, as 1.14.4 is, and the tags alone: in three
// unnamed runs, the blocks, the items and the fluids, since the entity types
// have no tags before 1.14.
func TestProviderSends1_13_2ThreeRunsOfTags(t *testing.T) {
	registries := mustLoadRegistries(t, registriesMinecraft1_13_2)
	if len(registries) != 0 {
		t.Fatalf("1.13.2 is sent %v, want no registry", registries)
	}

	tags := mustLoadTags(t, tagsMinecraft1_13_2)
	if len(tags) != len(unnamedTagRegistries1_13_2) {
		t.Fatalf("1.13.2's jar has tags for %d registries, want %d", len(tags), len(unnamedTagRegistries1_13_2))
	}

	for i, set := range tags {
		if set.Registry != unnamedTagRegistries1_13_2[i] {
			t.Errorf("set %d is for %s, want %s", i, set.Registry, unnamedTagRegistries1_13_2[i])
		}
	}

	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	version := types.ProtocolVersions.MINECRAFT_1_13_2

	if codec := provider.RegistryCodecFor(version); codec != nil {
		t.Errorf("1.13.2 has a registry codec of %d bytes, want none: its login holds a dimension's number", len(codec))
	}

	if dimensionType := provider.DimensionTypeFor(version); dimensionType != nil {
		t.Errorf("1.13.2 has a dimension type of %d bytes, want none", len(dimensionType))
	}

	packets := provider.PacketsFor(version)
	if len(packets) != 1 {
		t.Fatalf("1.13.2 is sent %d packets, want the tags alone", len(packets))
	}

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packets[0].Encode(out); err != nil {
		t.Fatalf("Encode() error: %v", err)
	}

	if err := out.Flush(); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	want, err := encodeTags1_13_2(tags)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), want) {
		t.Error("1.13.2's tags are not in the three unnamed runs it reads")
	}
}

// Three runs are what 1.13.2 reads, so the four 1.14 reads are refused, as
// is a set in another order.
func TestEncodeTags1_13_2RefusesAnythingButTheThreeRuns(t *testing.T) {
	if _, err := encodeTags1_13_2(mustLoadTags(t, tagsMinecraft1_14_4)); err == nil {
		t.Error("expected 1.14.4's four sets of tags to be refused")
	}

	tags := mustLoadTags(t, tagsMinecraft1_13_2)
	tags[0], tags[1] = tags[1], tags[0]

	if _, err := encodeTags1_13_2(tags); err == nil {
		t.Error("expected the items ahead of the blocks to be refused")
	}
}

// 1.13.1 is sent what 1.13.2 is, to the byte: the two jars' tags are the same
// files with the same contents and their blocks reports the same bytes, so
// nothing here is its own, and it is sent no registry either.
func TestProviderSends1_13_1What1_13_2Is(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("NewDefaultProvider() error: %v", err)
	}

	older, newer := types.ProtocolVersions.MINECRAFT_1_13_1, types.ProtocolVersions.MINECRAFT_1_13_2

	sendsTheSetOf(t, provider, older, newer)

	olderStates, err := BlockStatesFor(older)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	newerStates, err := BlockStatesFor(newer)
	if err != nil {
		t.Fatalf("BlockStatesFor() error: %v", err)
	}

	// The oak sign 1.13.2 answers to as its sign, which 1.13.1 does as well.
	stored := map[string]string{"rotation": "4", "waterlogged": "false"}

	olderId, olderOk := olderStates.Id("minecraft:oak_sign", stored)
	newerId, newerOk := newerStates.Id("minecraft:oak_sign", stored)

	if !olderOk || !newerOk || olderId != newerId {
		t.Errorf("an oak sign = %d, %t on 1.13.1 and %d, %t on 1.13.2, want the two alike", olderId, olderOk, newerId, newerOk)
	}
}
