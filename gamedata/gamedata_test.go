package gamedata

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/configuration"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"strings"
	"testing"
)

func TestRegistryEncode(t *testing.T) {
	registry := Registry{
		Name: "minecraft:test",
		Entries: []Entry{
			{Name: "minecraft:a", Data: nbt.Compound{"x": nbt.Byte(1)}},
			{Name: "minecraft:b"},
			{Name: "minecraft:c", Data: nbt.End{}},
		},
	}

	got, err := registry.encode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{
		0x0E, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 't', 'e', 's', 't', // registry name
		0x03,                                                        // entry count
		0x0B, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 'a', // first entry name
		0x01,                                    // has data
		0x0A, 0x01, 0x00, 0x01, 'x', 0x01, 0x00, // unnamed root compound {x:1b}
		0x0B, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 'b', // second entry name
		0x00,                                                        // no data
		0x0B, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 'c', // third entry name
		0x00, // an End tag means no data too
	}

	if !bytes.Equal(got, want) {
		t.Errorf("encoding mismatch.\n got: % X\nwant: % X", got, want)
	}
}

func TestRegistryEncodeEmpty(t *testing.T) {
	got, err := Registry{Name: "minecraft:enchantment"}.encode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A registry with no entries still has to be sent, so the body is the name
	// and a zero count rather than nothing at all.
	if len(got) == 0 || got[len(got)-1] != 0x00 {
		t.Errorf("expected a name followed by a zero entry count, got % X", got)
	}
}

func newTestProvider(t *testing.T, sets ...Set) *Provider {
	t.Helper()

	provider, err := NewProvider(sets...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return provider
}

// mustLoadRegistries and mustLoadTags unwrap one version's content for the
// tests that inspect it, failing the test on a data file that cannot be read.
func mustLoadRegistries(t *testing.T, load func() ([]Registry, error)) []Registry {
	t.Helper()

	registries, err := load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return registries
}

func mustLoadTags(t *testing.T, load func() ([]TagSet, error)) []TagSet {
	t.Helper()

	tags, err := load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	return tags
}

// registryNames collects the registries in the order they are sent. Anything
// that is not registry data (the tags packet) is skipped rather than rejected.
func registryNames(t *testing.T, packets []types.ClientboundPacket) []string {
	t.Helper()

	names := make([]string, 0, len(packets))
	for _, packet := range packets {
		switch typed := packet.(type) {
		case *configuration.RegistryDataClientboundPacket:
			names = append(names, typed.RegistryName())
		case *configuration.UpdateTagsClientboundPacket:
		default:
			t.Fatalf("unexpected packet type %T", packet)
		}
	}

	return names
}

// TestPacketsForResolvesTheNewestSetReached covers the point of bucketing by
// version: a set applies to every version from where it starts until the next
// one begins, so a version that changed nothing needs no set.
func TestPacketsForResolvesTheNewestSetReached(t *testing.T) {
	provider := newTestProvider(t,
		Set{MinProtocol: 300, Registries: []Registry{{Name: "newer"}}},
		Set{MinProtocol: 100, Registries: []Registry{{Name: "older"}}},
	)

	tests := []struct {
		protocol types.ProtocolId
		want     []string
	}{
		{99, nil},
		{100, []string{"older"}},
		{299, []string{"older"}},
		{300, []string{"newer"}},
		{9000, []string{"newer"}},
	}

	for _, test := range tests {
		packets := provider.PacketsFor(types.ProtocolVersion{ID: test.protocol})

		if test.want == nil {
			if packets != nil {
				t.Errorf("protocol %d: expected no packets, got %v", test.protocol, registryNames(t, packets))
			}
			continue
		}

		got := registryNames(t, packets)
		if len(got) != 1 || got[0] != test.want[0] {
			t.Errorf("protocol %d: registries = %v, want %v", test.protocol, got, test.want)
		}
	}
}

func TestPacketsForPreservesRegistryOrder(t *testing.T) {
	want := []string{"minecraft:dimension_type", "minecraft:worldgen/biome", "minecraft:damage_type"}

	registries := make([]Registry, 0, len(want))
	for _, name := range want {
		registries = append(registries, Registry{Name: name})
	}

	provider := newTestProvider(t, Set{MinProtocol: 1, Registries: registries})
	got := registryNames(t, provider.PacketsFor(types.ProtocolVersion{ID: 1}))

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("registry order = %v, want %v", got, want)
		}
	}
}

// TestPacketsForReusesEncodedPackets pins the reason bodies are encoded up
// front: a connection costs a lookup, not an encode.
func TestPacketsForReusesEncodedPackets(t *testing.T) {
	provider := newTestProvider(t, Set{MinProtocol: 1, Registries: []Registry{{Name: "minecraft:test"}}})

	first := provider.PacketsFor(types.ProtocolVersion{ID: 1})
	second := provider.PacketsFor(types.ProtocolVersion{ID: 1})

	if first[0] != second[0] {
		t.Error("expected the same packet instance to be handed to every connection")
	}
}

func TestNewProviderRejectsOverlappingSets(t *testing.T) {
	_, err := NewProvider(
		Set{MinProtocol: 100, Registries: []Registry{{Name: "a"}}},
		Set{MinProtocol: 100, Registries: []Registry{{Name: "b"}}},
	)

	if err == nil {
		t.Fatal("expected an error for two sets starting at the same protocol")
	}
}

func TestDefaultProviderCoversTheSupportedVersion(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	packets := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_26_2)
	if len(packets) == 0 {
		t.Fatal("expected registry packets for the supported protocol version")
	}

	// dimension_type and worldgen/biome are the two the play phase cannot do
	// without: the login packet names a dimension and every chunk names a biome.
	required := []string{"minecraft:dimension_type", "minecraft:worldgen/biome"}

	names := registryNames(t, packets)
	for _, want := range required {
		found := false
		for _, name := range names {
			if name == want {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("registry %s is missing from %v", want, names)
		}
	}
}

// TestCrossRegistryReferencesResolve covers entries that name an entry in a
// different registry. Those references are only resolvable once every registry
// has arrived, so nothing catches a dangling one until the client is already
// mid-join and reports it as a missing element.
func TestCrossRegistryReferencesResolve(t *testing.T) {
	entries := map[string]map[string]bool{}
	for _, registry := range mustLoadRegistries(t, registriesMinecraft26_2) {
		names := make(map[string]bool, len(registry.Entries))
		for _, entry := range registry.Entries {
			names[entry.Name] = true
		}

		entries[registry.Name] = names
	}

	references := []struct{ registry, entry, referencedBy string }{
		{"minecraft:damage_type", "minecraft:thorns", "enchantment/thorns"},
		{"minecraft:damage_type", "minecraft:sulfur_cube_hot", "sulfur_cube_archetype/hot"},
		{"minecraft:world_clock", "minecraft:overworld", "timeline/day"},
		{"minecraft:test_environment", "minecraft:default", "test_instance/always_pass"},
	}

	for _, reference := range references {
		if !entries[reference.registry][reference.entry] {
			t.Errorf("%s is missing %s, which %s points at",
				reference.registry, reference.entry, reference.referencedBy)
		}
	}
}

// TestDefaultProviderSendsTagsLast pins the ordering the client needs: a tag
// names its entries by registry id, so the registries have to be in place
// before the tags that point into them.
func TestDefaultProviderSendsTagsLast(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	packets := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_26_2)
	if len(packets) < 2 {
		t.Fatalf("expected registries and a tags packet, got %d packets", len(packets))
	}

	for i, packet := range packets[:len(packets)-1] {
		if _, ok := packet.(*configuration.RegistryDataClientboundPacket); !ok {
			t.Errorf("packet %d is %T, want registry data", i, packet)
		}
	}

	if _, ok := packets[len(packets)-1].(*configuration.UpdateTagsClientboundPacket); !ok {
		t.Errorf("last packet is %T, want update tags", packets[len(packets)-1])
	}
}

func TestEncodeTags(t *testing.T) {
	got, err := encodeTags([]TagSet{
		{Registry: "minecraft:damage_type", Tags: []NamedTag{
			{Name: "minecraft:is_fire"},
			{Name: "minecraft:bypasses_armor", Entries: []int32{0, 3}},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []byte{
		0x01, // one registry
		0x15, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 'd', 'a', 'm', 'a', 'g', 'e', '_', 't', 'y', 'p', 'e',
		0x02, // two tags
		0x11, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 'i', 's', '_', 'f', 'i', 'r', 'e',
		0x00, // no entries
		0x18, 'm', 'i', 'n', 'e', 'c', 'r', 'a', 'f', 't', ':', 'b', 'y', 'p', 'a', 's', 's', 'e', 's', '_', 'a', 'r', 'm', 'o', 'r',
		0x02, 0x00, 0x03, // two entries, by registry id
	}

	if !bytes.Equal(got, want) {
		t.Errorf("encoding mismatch.\n got: % X\nwant: % X", got, want)
	}
}

// TestDefaultProviderEntriesRoundTrip decodes what the provider produced, so a
// definition that encodes but is structurally wrong does not pass unnoticed.
func TestDefaultProviderEntriesRoundTrip(t *testing.T) {
	for _, registry := range mustLoadRegistries(t, registriesMinecraft26_2) {
		for _, entry := range registry.Entries {
			if entry.Data == nil {
				continue
			}

			t.Run(entry.Name, func(t *testing.T) {
				buf := new(bytes.Buffer)
				out := streams.NewMinecraftStreamFromBuffer(buf)

				if err := nbt.Write(out, entry.Data); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if err := out.Flush(); err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				got, err := nbt.Read(streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(buf.Bytes())))
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if got.String() != entry.Data.String() {
					t.Errorf("round trip changed the definition.\n got: %v\nwant: %v", got, entry.Data)
				}
			})
		}
	}
}

// What 26.2 adds is what 26.1 must not be sent. A registry it has never heard
// of is not one it skips, and an entry naming something it does not have fails
// to parse and takes its whole registry with it.
func TestRegistriesFor26_1LeaveOutWhat26_2Added(t *testing.T) {
	registries := map[string]map[string]bool{}
	for _, registry := range mustLoadRegistries(t, registriesMinecraft26_1) {
		names := make(map[string]bool, len(registry.Entries))
		for _, entry := range registry.Entries {
			names[entry.Name] = true
		}

		registries[registry.Name] = names
	}

	if _, ok := registries["minecraft:sulfur_cube_archetype"]; ok {
		t.Error("26.1 is sent minecraft:sulfur_cube_archetype, which it does not synchronize")
	}

	// bounce names minecraft:music_disc.bounce, a sound event 26.1 has no entry
	// for, which is what a 26.1 client reported as a jukebox song it could not
	// parse.
	if registries["minecraft:jukebox_song"]["minecraft:bounce"] {
		t.Error("26.1 is sent the bounce jukebox song, which names a sound it does not have")
	}

	if registries["minecraft:damage_type"]["minecraft:sulfur_cube_hot"] {
		t.Error("26.1 is sent the sulfur_cube_hot damage type, which 26.2 added")
	}

	// The registries themselves are still all there: leaving out an entry is not
	// leaving out the registry it belonged to.
	if len(registries["minecraft:jukebox_song"]) == 0 || len(registries["minecraft:damage_type"]) == 0 {
		t.Error("26.1 was sent an empty jukebox_song or damage_type")
	}
}

// 26.2 rewrote the entity predicates inside eleven enchantments. An entry in the
// new shape is one 26.1 reads as a predicate with no type constraint at best, so
// the 26.1 content has to come from 26.1's own data rather than from 26.2's.
func TestEnchantmentsFor26_1UseTheOlderPredicateShape(t *testing.T) {
	rendered := func(registries []Registry) string {
		for _, registry := range registries {
			if registry.Name != "minecraft:enchantment" {
				continue
			}

			var sb strings.Builder
			for _, entry := range registry.Entries {
				sb.WriteString(entry.Data.String())
			}

			return sb.String()
		}

		return ""
	}

	older := rendered(mustLoadRegistries(t, registriesMinecraft26_1))
	newer := rendered(mustLoadRegistries(t, registriesMinecraft26_2))

	if older == "" || newer == "" {
		t.Fatal("no enchantment registry in one of the sets")
	}

	if strings.Contains(older, "minecraft:entity_type") {
		t.Error("26.1 enchantments carry 26.2's namespaced entity predicate keys")
	}

	if !strings.Contains(newer, "minecraft:entity_type") {
		t.Error("26.2 enchantments no longer carry the namespaced keys they were generated with")
	}
}

// The provider picks the set by version, so each one has to get its own.
func TestProviderGivesEachVersionItsOwnRegistries(t *testing.T) {
	provider, err := NewDefaultProvider()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	older := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_26_1)
	newer := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_26_2)

	if len(older) == 0 {
		t.Fatal("26.1 was sent no packets at all, which is a client that never reaches the world")
	}

	if len(older) >= len(newer) {
		t.Errorf("26.1 was sent %d packets and 26.2 %d, want fewer for 26.1", len(older), len(newer))
	}
}

// 26.1 reads infiniburn as the name of a block tag and 26.2 as a set of blocks.
// Sending 26.2's empty list to a 26.1 client is a dimension type it cannot
// decode, which is the whole registry rejected.
func TestDimensionTypeFor26_1NamesTheInfiniburnTag(t *testing.T) {
	var dimensionType nbt.Compound

	for _, registry := range mustLoadRegistries(t, registriesMinecraft26_1) {
		if registry.Name != "minecraft:dimension_type" {
			continue
		}

		if len(registry.Entries) != 1 {
			t.Fatalf("expected one dimension type, got %d", len(registry.Entries))
		}

		compound, ok := registry.Entries[0].Data.(nbt.Compound)
		if !ok {
			t.Fatalf("expected a compound, got %T", registry.Entries[0].Data)
		}

		dimensionType = compound
	}

	if dimensionType == nil {
		t.Fatal("26.1 is sent no dimension type at all")
	}

	infiniburn, ok := dimensionType["infiniburn"].(nbt.String)
	if !ok {
		t.Fatalf("26.1 infiniburn is %T, want a tag name", dimensionType["infiniburn"])
	}

	if infiniburn != "#minecraft:infiniburn_overworld" {
		t.Errorf("26.1 infiniburn is %q, want the block tag the client's own overworld names", infiniburn)
	}

	// 26.2 keeps the set form, and the two must not have been crossed over.
	if _, ok := overworldDimensionType["infiniburn"].(nbt.List); !ok {
		t.Errorf("26.2 infiniburn is %T, want a set of blocks", overworldDimensionType["infiniburn"])
	}
}

// A tag the client asks for and was never sent throws rather than defaulting to
// empty, so each version is sent the tag names its own jar declares. 26.2
// renamed one block tag, which is the case that would otherwise go missing.
func TestTagsFor26_1DeclareTheNamesThatVersionAsksFor(t *testing.T) {
	names := func(sets []TagSet, registry string) map[string]bool {
		for _, set := range sets {
			if set.Registry != registry {
				continue
			}

			out := make(map[string]bool, len(set.Tags))
			for _, tag := range set.Tags {
				out[tag.Name] = true
			}

			return out
		}

		return nil
	}

	older := names(mustLoadTags(t, tagsMinecraft26_1), "minecraft:block")
	newer := names(mustLoadTags(t, tagsMinecraft26_2), "minecraft:block")

	if older == nil || newer == nil {
		t.Fatal("no block tags in one of the sets")
	}

	if !older["minecraft:concrete_powder"] {
		t.Error("26.1 is not sent minecraft:concrete_powder, which it asks for")
	}

	if newer["minecraft:concrete_powder"] {
		t.Error("26.2 is sent minecraft:concrete_powder, which it renamed")
	}

	if !newer["minecraft:concrete_powders"] {
		t.Error("26.2 is not sent minecraft:concrete_powders, which it asks for")
	}
}
