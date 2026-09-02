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

	// A set above 1.20.5, where each registry is a packet of its own and the
	// order is the order of the packets; below it the one packet holds them
	// all, which TestProviderSendsOneCompoundBelow1_20_5 covers.
	provider := newTestProvider(t, Set{MinProtocol: 1000, Registries: registries})
	got := registryNames(t, provider.PacketsFor(types.ProtocolVersion{ID: 1000}))

	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("registry order = %v, want %v", got, want)
		}
	}
}

// A client before 1.20.5 reads every registry from one packet, as one
// compound keyed by registry name, each holding its name again under type and
// its entries under value with explicit ids. The ids are the positions the
// per-registry shape numbers entries by, so the dimension type the play login
// names by index is the same entry on either side of 1.20.5.
func TestProviderSendsOneCompoundBelow1_20_5(t *testing.T) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: nbt.Compound{"height": nbt.Int(384)}},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: nbt.Compound{"temperature": nbt.Float(0.8)}},
			{Name: "minecraft:desert", Data: nbt.Compound{"temperature": nbt.Float(2)}},
		}},
	}

	provider := newTestProvider(t, Set{MinProtocol: types.ProtocolVersions.MINECRAFT_1_20_3.ID, Registries: registries})
	packets := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_20_3)

	if len(packets) != 1 {
		t.Fatalf("1.20.3 was sent %d registry packets, want the one holding every registry", len(packets))
	}

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packets[0].Encode(out); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := out.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	decoded, err := nbt.Read(streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(buf.Bytes())))
	if err != nil {
		t.Fatalf("the body is not one nameless compound: %v", err)
	}

	want := nbt.Compound{
		"minecraft:dimension_type": nbt.Compound{
			"type": nbt.String("minecraft:dimension_type"),
			"value": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
				nbt.Compound{"name": nbt.String("minecraft:overworld"), "id": nbt.Int(0), "element": nbt.Compound{"height": nbt.Int(384)}},
			}},
		},
		"minecraft:worldgen/biome": nbt.Compound{
			"type": nbt.String("minecraft:worldgen/biome"),
			"value": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
				nbt.Compound{"name": nbt.String("minecraft:plains"), "id": nbt.Int(0), "element": nbt.Compound{"temperature": nbt.Float(0.8)}},
				nbt.Compound{"name": nbt.String("minecraft:desert"), "id": nbt.Int(1), "element": nbt.Compound{"temperature": nbt.Float(2)}},
			}},
		},
	}

	if decoded.String() != want.String() {
		t.Errorf("1.20.3's registry data decodes as\n%v\nwant\n%v", decoded, want)
	}
}

// The combined shape has no way to send a name without its definition, since
// nothing before 1.20.5 knows a pack to fall back on, so a set that tries is
// refused when the provider is built rather than sent to a client that cannot
// fill it in.
func TestProviderRefusesAnEntryWithoutDataBelow1_20_5(t *testing.T) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{{Name: "minecraft:overworld"}}},
	}

	if _, err := NewProvider(Set{MinProtocol: types.ProtocolVersions.MINECRAFT_1_20_3.ID, Registries: registries}); err == nil {
		t.Error("error = nil, want a refusal for an entry with no definition")
	}

	// The same set one version up is fine: the per-registry shape sends the
	// name with a flag saying no definition follows.
	if _, err := NewProvider(Set{MinProtocol: types.ProtocolVersions.MINECRAFT_1_20_5.ID, Registries: registries}); err != nil {
		t.Errorf("unexpected error: %v", err)
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

func TestRegistriesFor1_21_11LeaveOutWhat26_1Added(t *testing.T) {
	registries := map[string]int{}
	for _, registry := range mustLoadRegistries(t, registriesMinecraft1_21_11) {
		registries[registry.Name] = len(registry.Entries)
	}

	// 26.1 synchronizes the four animal sound variant registries and
	// world_clock, none of which 1.21.11 has ever heard of. A registry the
	// client is sent and does not expect is rejected along with everything
	// after it.
	for _, added := range []string{
		"minecraft:cat_sound_variant", "minecraft:chicken_sound_variant",
		"minecraft:cow_sound_variant", "minecraft:pig_sound_variant",
		"minecraft:world_clock",
	} {
		if _, ok := registries[added]; ok {
			t.Errorf("1.21.11 is sent %s, which it does not synchronize", added)
		}
	}

	// All 23 of its own SYNCHRONIZED_REGISTRIES are there: the twenty-one
	// generated from its jar and the two written here.
	if len(registries) != 23 {
		t.Errorf("1.21.11 is sent %d registries, want 23", len(registries))
	}

	if registries["minecraft:damage_type"] == 0 || registries["minecraft:enchantment"] == 0 {
		t.Error("1.21.11 was sent an empty damage_type or enchantment")
	}
}

// 26.1 added has_ender_dragon_fight to the dimension type codec, and requires
// it. 1.21.11's codec has no such field -- an unknown field would merely be
// ignored, but the entry documents itself as exactly what the codec requires,
// so the field 1.21.11 cannot read is the field it is not sent.
func TestDimensionTypeFor1_21_11DropsTheEnderDragonFight(t *testing.T) {
	var dimensionType nbt.Compound

	for _, registry := range mustLoadRegistries(t, registriesMinecraft1_21_11) {
		if registry.Name != "minecraft:dimension_type" {
			continue
		}

		if len(registry.Entries) != 1 {
			t.Fatalf("expected one dimension type, got %d", len(registry.Entries))
		}

		dimensionType = registry.Entries[0].Data.(nbt.Compound)
	}

	if dimensionType == nil {
		t.Fatal("1.21.11 is sent no dimension type at all")
	}

	if _, ok := dimensionType["has_ender_dragon_fight"]; ok {
		t.Error("1.21.11 is sent has_ender_dragon_fight, which its codec has no field for")
	}

	// Everything else reads as 26.1 does, infiniburn as a tag name included:
	// the reworked schema predates the 26.x versions.
	if infiniburn, ok := dimensionType["infiniburn"].(nbt.String); !ok || infiniburn != "#minecraft:infiniburn_overworld" {
		t.Errorf("1.21.11 infiniburn is %v, want the block tag the client's own overworld names", dimensionType["infiniburn"])
	}

	if _, ok := overworldDimensionType26_1["has_ender_dragon_fight"]; !ok {
		t.Error("26.1 lost has_ender_dragon_fight, which its codec requires")
	}
}

func TestRegistriesFor1_21_9LeaveOutWhat1_21_11Added(t *testing.T) {
	registries := map[string]int{}
	for _, registry := range mustLoadRegistries(t, registriesMinecraft1_21_9) {
		registries[registry.Name] = len(registry.Entries)
	}

	// 1.21.11 synchronizes zombie_nautilus_variant and timeline, neither of
	// which 1.21.9 has ever heard of. A registry the client is sent and does
	// not expect is rejected along with everything after it.
	for _, added := range []string{
		"minecraft:zombie_nautilus_variant", "minecraft:timeline",
	} {
		if _, ok := registries[added]; ok {
			t.Errorf("1.21.9 is sent %s, which it does not synchronize", added)
		}
	}

	// All 21 of its own SYNCHRONIZED_REGISTRIES are there: the nineteen
	// generated from its jar and the two written here.
	if len(registries) != 21 {
		t.Errorf("1.21.9 is sent %d registries, want 21", len(registries))
	}

	if registries["minecraft:damage_type"] == 0 || registries["minecraft:enchantment"] == 0 {
		t.Error("1.21.9 was sent an empty damage_type or enchantment")
	}
}

// 1.21.7 synchronizes the same 21 registries as 1.21.9 -- 773 added no
// registry -- but their contents are each version's own: the copper additions
// 773 landed reach into trim_material and enchantment, and into the block,
// item and entity tags. A tag the client asks for that was never declared
// throws, and the reverse -- declaring what 1.21.9 has and 1.21.7 does not --
// would ship content no 1.21.7 client was built against.
func TestRegistriesFor1_21_7AreItsOwnAndNot1_21_9s(t *testing.T) {
	registries := map[string]int{}
	for _, registry := range mustLoadRegistries(t, registriesMinecraft1_21_7) {
		registries[registry.Name] = len(registry.Entries)
	}

	if len(registries) != 21 {
		t.Errorf("1.21.7 is sent %d registries, want 21", len(registries))
	}

	if registries["minecraft:damage_type"] == 0 || registries["minecraft:enchantment"] == 0 {
		t.Error("1.21.7 was sent an empty damage_type or enchantment")
	}

	blockTags := func(load func() ([]TagSet, error)) map[string]bool {
		sets, err := load()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		names := map[string]bool{}
		for _, set := range sets {
			if set.Registry != "minecraft:block" {
				continue
			}

			for _, tag := range set.Tags {
				names[tag.Name] = true
			}
		}

		return names
	}

	older, newer := blockTags(tagsMinecraft1_21_7), blockTags(tagsMinecraft1_21_9)

	if older["minecraft:copper_chests"] {
		t.Error("1.21.7 declares copper_chests, a block tag 773 introduced")
	}

	if !newer["minecraft:copper_chests"] {
		t.Error("1.21.9 no longer declares copper_chests, which its jar does")
	}
}

// 1.21.6 synchronizes the same 21 registries as 1.21.7 -- 772 added no
// registry, and their jars declare byte-identical tags -- but the contents
// are each version's own: 772 is where the lava_chicken jukebox song and the
// dennis painting variant landed. Sending either to a 1.21.6 client would
// ship content no such client was built against.
func TestRegistriesFor1_21_6AreItsOwnAndNot1_21_7s(t *testing.T) {
	entries := func(load func() ([]Registry, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, registry := range mustLoadRegistries(t, load) {
			names[registry.Name] = map[string]bool{}
			for _, entry := range registry.Entries {
				names[registry.Name][entry.Name] = true
			}
		}

		return names
	}

	older, newer := entries(registriesMinecraft1_21_6), entries(registriesMinecraft1_21_7)

	if len(older) != 21 {
		t.Errorf("1.21.6 is sent %d registries, want 21", len(older))
	}

	if len(older["minecraft:damage_type"]) == 0 || len(older["minecraft:enchantment"]) == 0 {
		t.Error("1.21.6 was sent an empty damage_type or enchantment")
	}

	if older["minecraft:jukebox_song"]["minecraft:lava_chicken"] {
		t.Error("1.21.6 is sent the lava_chicken jukebox song, which 772 introduced")
	}

	if !newer["minecraft:jukebox_song"]["minecraft:lava_chicken"] {
		t.Error("1.21.7 no longer has the lava_chicken jukebox song, which its jar does")
	}

	if older["minecraft:painting_variant"]["minecraft:dennis"] {
		t.Error("1.21.6 is sent the dennis painting variant, which 772 introduced")
	}

	if !newer["minecraft:painting_variant"]["minecraft:dennis"] {
		t.Error("1.21.7 no longer has the dennis painting variant, which its jar does")
	}
}

// 1.21.5 synchronizes one registry fewer than 1.21.6: dialog is 771's, added
// for the dialogs it introduced, and a registry the client does not expect is
// as much a failed configuration as one it expects and never receives. The
// generated content is 1.21.5's own as well, because the tears jukebox song
// landed in 771, and so did the happy ghast's tags.
func TestRegistriesFor1_21_5AreItsOwnAndNot1_21_6s(t *testing.T) {
	entries := func(load func() ([]Registry, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, registry := range mustLoadRegistries(t, load) {
			names[registry.Name] = map[string]bool{}
			for _, entry := range registry.Entries {
				names[registry.Name][entry.Name] = true
			}
		}

		return names
	}

	older, newer := entries(registriesMinecraft1_21_5), entries(registriesMinecraft1_21_6)

	if len(older) != 20 {
		t.Errorf("1.21.5 is sent %d registries, want 20", len(older))
	}

	if _, ok := older["minecraft:dialog"]; ok {
		t.Error("1.21.5 is sent the dialog registry, which 771 introduced")
	}

	if _, ok := newer["minecraft:dialog"]; !ok {
		t.Error("1.21.6 is no longer sent the dialog registry, which its jar synchronizes")
	}

	if len(older["minecraft:damage_type"]) == 0 || len(older["minecraft:enchantment"]) == 0 {
		t.Error("1.21.5 was sent an empty damage_type or enchantment")
	}

	if older["minecraft:jukebox_song"]["minecraft:tears"] {
		t.Error("1.21.5 is sent the tears jukebox song, which 771 introduced")
	}

	if !newer["minecraft:jukebox_song"]["minecraft:tears"] {
		t.Error("1.21.6 no longer has the tears jukebox song, which its jar does")
	}

	tags := func(load func() ([]TagSet, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, set := range mustLoadTags(t, load) {
			names[set.Registry] = map[string]bool{}
			for _, tag := range set.Tags {
				names[set.Registry][tag.Name] = true
			}
		}

		return names
	}

	olderTags, newerTags := tags(tagsMinecraft1_21_5), tags(tagsMinecraft1_21_6)

	if _, ok := olderTags["minecraft:dialog"]; ok {
		t.Error("1.21.5 declares dialog tags, for a registry it does not have")
	}

	if olderTags["minecraft:item"]["minecraft:happy_ghast_food"] {
		t.Error("1.21.5 declares happy_ghast_food, an item tag 771 introduced")
	}

	if !newerTags["minecraft:item"]["minecraft:happy_ghast_food"] {
		t.Error("1.21.6 no longer declares happy_ghast_food, which its jar does")
	}

	if !olderTags["minecraft:block"]["minecraft:plays_ambient_desert_block_sounds"] {
		t.Error("1.21.5 no longer declares plays_ambient_desert_block_sounds, which its jar does and 771 renamed")
	}
}

// 1.21.4 synchronizes eight registries fewer than 1.21.5, which made the
// remaining animal variants data-driven and added the test registries, and a
// registry the client does not expect is as much a failed configuration as
// one it expects and never receives. The shared ten hold the same entries in
// both, but 1.21.4's content is its own: the wolf variant still carries its
// textures by name and its biomes, and the trim material its ingredient.
func TestRegistriesFor1_21_4AreItsOwnAndNot1_21_5s(t *testing.T) {
	registries := func(load func() ([]Registry, error)) map[string]map[string]nbt.Tag {
		data := map[string]map[string]nbt.Tag{}
		for _, registry := range mustLoadRegistries(t, load) {
			data[registry.Name] = map[string]nbt.Tag{}
			for _, entry := range registry.Entries {
				data[registry.Name][entry.Name] = entry.Data
			}
		}

		return data
	}

	older, newer := registries(registriesMinecraft1_21_4), registries(registriesMinecraft1_21_5)

	if len(older) != 12 {
		t.Errorf("1.21.4 is sent %d registries, want 12", len(older))
	}

	for _, name := range []string{"minecraft:cat_variant", "minecraft:frog_variant", "minecraft:pig_variant", "minecraft:cow_variant", "minecraft:chicken_variant", "minecraft:wolf_sound_variant", "minecraft:test_environment", "minecraft:test_instance"} {
		if _, ok := older[name]; ok {
			t.Errorf("1.21.4 is sent %s, which 770 introduced", name)
		}

		if _, ok := newer[name]; !ok {
			t.Errorf("1.21.5 is no longer sent %s, which its jar synchronizes", name)
		}
	}

	for name, entries := range older {
		if len(entries) == 0 {
			t.Errorf("1.21.4 was sent an empty %s", name)
		}

		if name == "minecraft:dimension_type" || name == "minecraft:worldgen/biome" {
			continue
		}

		if len(entries) != len(newer[name]) {
			t.Errorf("1.21.4's %s holds %d entries and 1.21.5's %d, want the same entries in both", name, len(entries), len(newer[name]))
		}
	}

	wolf, ok := older["minecraft:wolf_variant"]["minecraft:pale"].(nbt.Compound)
	if !ok {
		t.Fatal("1.21.4 has no pale wolf variant")
	}

	if _, ok := wolf["wild_texture"]; !ok {
		t.Error("1.21.4's wolf variant lost the wild_texture its codec requires")
	}

	if _, ok := wolf["assets"]; ok {
		t.Error("1.21.4's wolf variant carries 770's assets compound")
	}

	// The biomes are a set of references into a registry this server sends
	// one entry of, so the set goes out empty rather than naming a biome the
	// client would fail to bind.
	if biomes, ok := wolf["biomes"].(nbt.List); !ok || len(biomes.Elements) != 0 {
		t.Errorf("1.21.4's wolf variant biomes = %v, want an empty list", wolf["biomes"])
	}

	if _, ok := older["minecraft:trim_material"]["minecraft:amethyst"].(nbt.Compound)["ingredient"]; !ok {
		t.Error("1.21.4's trim material lost the ingredient its codec requires")
	}

	if _, ok := newer["minecraft:trim_material"]["minecraft:amethyst"].(nbt.Compound)["ingredient"]; ok {
		t.Error("1.21.5's trim material carries an ingredient, which 770 moved to the item")
	}

	tags := func(load func() ([]TagSet, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, set := range mustLoadTags(t, load) {
			names[set.Registry] = map[string]bool{}
			for _, tag := range set.Tags {
				names[set.Registry][tag.Name] = true
			}
		}

		return names
	}

	olderTags, newerTags := tags(tagsMinecraft1_21_4), tags(tagsMinecraft1_21_5)

	if len(olderTags) != len(newerTags) {
		t.Errorf("1.21.4 declares tags for %d registries and 1.21.5 for %d, want the same registries", len(olderTags), len(newerTags))
	}

	if olderTags["minecraft:item"]["minecraft:eggs"] {
		t.Error("1.21.4 declares eggs, an item tag 770 introduced")
	}

	if !newerTags["minecraft:item"]["minecraft:eggs"] {
		t.Error("1.21.5 no longer declares eggs, which its jar does")
	}

	if !olderTags["minecraft:block"]["minecraft:dead_bush_may_place_on"] {
		t.Error("1.21.4 no longer declares dead_bush_may_place_on, which its jar does and 770 renamed")
	}
}

// 1.21.2 synchronizes the same twelve registries as 1.21.4, holding the same
// entries but one, and the content is its own: 1.21.4 is where the resin
// trim material landed, where the trim material lost the item model index
// that 1.21.4's item model rework made redundant, and where the power
// enchantment's damage was retuned. The tag sets are the same twelve
// registries too, and each version declares the tags of its own jar.
func TestRegistriesFor1_21_2AreItsOwnAndNot1_21_4s(t *testing.T) {
	registries := func(load func() ([]Registry, error)) map[string]map[string]nbt.Tag {
		data := map[string]map[string]nbt.Tag{}
		for _, registry := range mustLoadRegistries(t, load) {
			data[registry.Name] = map[string]nbt.Tag{}
			for _, entry := range registry.Entries {
				data[registry.Name][entry.Name] = entry.Data
			}
		}

		return data
	}

	older, newer := registries(registriesMinecraft1_21_2), registries(registriesMinecraft1_21_4)

	if len(older) != 12 {
		t.Errorf("1.21.2 is sent %d registries, want 12", len(older))
	}

	for name, entries := range older {
		if len(entries) == 0 {
			t.Errorf("1.21.2 was sent an empty %s", name)
		}

		if _, ok := newer[name]; !ok {
			t.Errorf("1.21.2 is sent %s, which 1.21.4 is not", name)
		}
	}

	for name := range newer {
		if _, ok := older[name]; !ok {
			t.Errorf("1.21.4 is sent %s, which 1.21.2 is not", name)
		}
	}

	if _, ok := older["minecraft:trim_material"]["minecraft:resin"]; ok {
		t.Error("1.21.2 is sent the resin trim material, which 769 introduced")
	}

	if _, ok := newer["minecraft:trim_material"]["minecraft:resin"]; !ok {
		t.Error("1.21.4 is no longer sent the resin trim material, which its jar holds")
	}

	amethyst, ok := older["minecraft:trim_material"]["minecraft:amethyst"].(nbt.Compound)
	if !ok {
		t.Fatal("1.21.2 has no amethyst trim material")
	}

	if _, ok := amethyst["item_model_index"]; !ok {
		t.Error("1.21.2's trim material lost the item_model_index its codec requires")
	}

	if _, ok := newer["minecraft:trim_material"]["minecraft:amethyst"].(nbt.Compound)["item_model_index"]; ok {
		t.Error("1.21.4's trim material carries an item_model_index, which 769 dropped")
	}

	// The wolf variant's biomes go out empty on 1.21.2 for the reason they do
	// on 1.21.4: its codec is the same set of biome references.
	wolf, ok := older["minecraft:wolf_variant"]["minecraft:pale"].(nbt.Compound)
	if !ok {
		t.Fatal("1.21.2 has no pale wolf variant")
	}

	if biomes, ok := wolf["biomes"].(nbt.List); !ok || len(biomes.Elements) != 0 {
		t.Errorf("1.21.2's wolf variant biomes = %v, want an empty list", wolf["biomes"])
	}

	tags := func(load func() ([]TagSet, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, set := range mustLoadTags(t, load) {
			names[set.Registry] = map[string]bool{}
			for _, tag := range set.Tags {
				names[set.Registry][tag.Name] = true
			}
		}

		return names
	}

	olderTags, newerTags := tags(tagsMinecraft1_21_2), tags(tagsMinecraft1_21_4)

	if len(olderTags) != len(newerTags) {
		t.Errorf("1.21.2 declares tags for %d registries and 1.21.4 for %d, want the same registries", len(olderTags), len(newerTags))
	}

	if olderTags["minecraft:block"]["minecraft:pale_oak_logs"] {
		t.Error("1.21.2 declares pale_oak_logs, a block tag 769 introduced")
	}

	if !newerTags["minecraft:block"]["minecraft:pale_oak_logs"] {
		t.Error("1.21.4 no longer declares pale_oak_logs, which its jar does")
	}

	if !olderTags["minecraft:item"]["minecraft:tall_flowers"] {
		t.Error("1.21.2 no longer declares tall_flowers, which its jar does and 769 retired")
	}
}

// 1.21 synchronizes eleven registries to 1.21.2's twelve -- the instrument
// became data-driven in 1.21.2 -- and the content is its own: 1.21.2 is where
// the mace smash and ender pearl damage types landed, where every painting
// gained a title and an author, and where ten enchantments were retuned. The
// tag sets cover the same twelve registries, and each version declares the
// tags of its own jar.
func TestRegistriesFor1_21AreItsOwnAndNot1_21_2s(t *testing.T) {
	registries := func(load func() ([]Registry, error)) map[string]map[string]nbt.Tag {
		data := map[string]map[string]nbt.Tag{}
		for _, registry := range mustLoadRegistries(t, load) {
			data[registry.Name] = map[string]nbt.Tag{}
			for _, entry := range registry.Entries {
				data[registry.Name][entry.Name] = entry.Data
			}
		}

		return data
	}

	older, newer := registries(registriesMinecraft1_21), registries(registriesMinecraft1_21_2)

	if len(older) != 11 {
		t.Errorf("1.21 is sent %d registries, want 11", len(older))
	}

	for name, entries := range older {
		if len(entries) == 0 {
			t.Errorf("1.21 was sent an empty %s", name)
		}

		if _, ok := newer[name]; !ok {
			t.Errorf("1.21 is sent %s, which 1.21.2 is not", name)
		}
	}

	if _, ok := older["minecraft:instrument"]; ok {
		t.Error("1.21 is sent the instrument registry, which 768 is where it became data-driven")
	}

	if _, ok := newer["minecraft:instrument"]; !ok {
		t.Error("1.21.2 is no longer sent the instrument registry, which its jar holds")
	}

	if _, ok := older["minecraft:damage_type"]["minecraft:mace_smash"]; ok {
		t.Error("1.21 is sent the mace smash damage type, which 768 introduced")
	}

	if _, ok := newer["minecraft:damage_type"]["minecraft:mace_smash"]; !ok {
		t.Error("1.21.2 is no longer sent the mace smash damage type, which its jar holds")
	}

	kebab, ok := older["minecraft:painting_variant"]["minecraft:kebab"].(nbt.Compound)
	if !ok {
		t.Fatal("1.21 has no kebab painting")
	}

	if _, ok := kebab["title"]; ok {
		t.Error("1.21's painting carries a title, which 768 introduced")
	}

	if _, ok := newer["minecraft:painting_variant"]["minecraft:kebab"].(nbt.Compound)["title"]; !ok {
		t.Error("1.21.2's painting lost the title its jar gives it")
	}

	// The wolf variant's biomes go out empty on 1.21 for the reason they do
	// on the versions above: its codec is the same set of biome references.
	wolf, ok := older["minecraft:wolf_variant"]["minecraft:pale"].(nbt.Compound)
	if !ok {
		t.Fatal("1.21 has no pale wolf variant")
	}

	if biomes, ok := wolf["biomes"].(nbt.List); !ok || len(biomes.Elements) != 0 {
		t.Errorf("1.21's wolf variant biomes = %v, want an empty list", wolf["biomes"])
	}

	tags := func(load func() ([]TagSet, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, set := range mustLoadTags(t, load) {
			names[set.Registry] = map[string]bool{}
			for _, tag := range set.Tags {
				names[set.Registry][tag.Name] = true
			}
		}

		return names
	}

	olderTags, newerTags := tags(tagsMinecraft1_21), tags(tagsMinecraft1_21_2)

	if len(olderTags) != len(newerTags) {
		t.Errorf("1.21 declares tags for %d registries and 1.21.2 for %d, want the same registries", len(olderTags), len(newerTags))
	}

	if olderTags["minecraft:item"]["minecraft:bundles"] {
		t.Error("1.21 declares bundles, an item tag 768 introduced")
	}

	if !newerTags["minecraft:item"]["minecraft:bundles"] {
		t.Error("1.21.2 no longer declares bundles, which its jar does")
	}

	if !olderTags["minecraft:instrument"]["minecraft:goat_horns"] {
		t.Error("1.21 no longer declares goat_horns, an instrument tag its jar does declare for the registry that was still built in")
	}
}

// 1.21.11 reworked the dimension type and biome schema; 1.21.9 is the last
// supported version that reads the shape before it. The entries document
// themselves as exactly what each codec requires, so what the rework removed
// is what 1.21.9 must still be sent, and what it added is what 1.21.9 must
// not be.
// 1.20.5 synchronizes eight registries to 1.21's eleven -- the painting
// variant, the enchantment and the jukebox song became data-driven in 1.21 --
// and the content is its own: 1.21 is where the flow and guster banner
// patterns and the campfire and wind charge damage types landed. The tag sets
// cover the same twelve registries, and each version declares the tags of its
// own jar: 1.21 retired the music discs item tag for the jukebox songs.
func TestRegistriesFor1_20_5AreItsOwnAndNot1_21s(t *testing.T) {
	registries := func(load func() ([]Registry, error)) map[string]map[string]nbt.Tag {
		data := map[string]map[string]nbt.Tag{}
		for _, registry := range mustLoadRegistries(t, load) {
			data[registry.Name] = map[string]nbt.Tag{}
			for _, entry := range registry.Entries {
				data[registry.Name][entry.Name] = entry.Data
			}
		}

		return data
	}

	older, newer := registries(registriesMinecraft1_20_5), registries(registriesMinecraft1_21)

	if len(older) != 8 {
		t.Errorf("1.20.5 is sent %d registries, want 8", len(older))
	}

	for name, entries := range older {
		if len(entries) == 0 {
			t.Errorf("1.20.5 was sent an empty %s", name)
		}

		if _, ok := newer[name]; !ok {
			t.Errorf("1.20.5 is sent %s, which 1.21 is not", name)
		}
	}

	for _, name := range []string{"minecraft:painting_variant", "minecraft:enchantment", "minecraft:jukebox_song"} {
		if _, ok := older[name]; ok {
			t.Errorf("1.20.5 is sent %s, which 767 is where it became data-driven", name)
		}

		if _, ok := newer[name]; !ok {
			t.Errorf("1.21 is no longer sent %s, which its jar holds", name)
		}
	}

	if _, ok := older["minecraft:damage_type"]["minecraft:wind_charge"]; ok {
		t.Error("1.20.5 is sent the wind charge damage type, which 767 introduced")
	}

	if _, ok := newer["minecraft:damage_type"]["minecraft:wind_charge"]; !ok {
		t.Error("1.21 is no longer sent the wind charge damage type, which its jar holds")
	}

	if _, ok := older["minecraft:banner_pattern"]["minecraft:flow"]; ok {
		t.Error("1.20.5 is sent the flow banner pattern, which 767 introduced")
	}

	if _, ok := newer["minecraft:banner_pattern"]["minecraft:flow"]; !ok {
		t.Error("1.21 is no longer sent the flow banner pattern, which its jar holds")
	}

	// The wolf variant's biomes go out empty on 1.20.5 for the reason they
	// do on the versions above: its codec is the same set of biome references.
	wolf, ok := older["minecraft:wolf_variant"]["minecraft:pale"].(nbt.Compound)
	if !ok {
		t.Fatal("1.20.5 has no pale wolf variant")
	}

	if biomes, ok := wolf["biomes"].(nbt.List); !ok || len(biomes.Elements) != 0 {
		t.Errorf("1.20.5's wolf variant biomes = %v, want an empty list", wolf["biomes"])
	}

	tags := func(load func() ([]TagSet, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, set := range mustLoadTags(t, load) {
			names[set.Registry] = map[string]bool{}
			for _, tag := range set.Tags {
				names[set.Registry][tag.Name] = true
			}
		}

		return names
	}

	olderTags, newerTags := tags(tagsMinecraft1_20_5), tags(tagsMinecraft1_21)

	if len(olderTags) != len(newerTags) {
		t.Errorf("1.20.5 declares tags for %d registries and 1.21 for %d, want the same registries", len(olderTags), len(newerTags))
	}

	if !olderTags["minecraft:item"]["minecraft:music_discs"] {
		t.Error("1.20.5 no longer declares music_discs, an item tag its jar does declare")
	}

	if newerTags["minecraft:item"]["minecraft:music_discs"] {
		t.Error("1.21 declares music_discs, an item tag 767 retired")
	}

	if olderTags["minecraft:item"]["minecraft:enchantable/mace"] {
		t.Error("1.20.5 declares enchantable/mace, an item tag 767 introduced")
	}

	if !olderTags["minecraft:block"]["minecraft:enchantment_power_provider"] {
		t.Error("1.20.5 no longer declares enchantment_power_provider, a block tag its jar does declare")
	}
}

// 1.20.3 synchronizes six registries to 1.20.5's eight -- the wolf variant
// and the banner pattern became data-driven in 1.20.5 -- and the content is
// its own: 1.20.5 is where the spit damage type landed. The tag sets cover
// eleven registries to 1.20.5's twelve, since 1.20.5 is where enchantment
// tags appeared, and each version declares the tags of its own jar: 1.20.5
// retired the tools item tag, and gave every animal a food tag.
func TestRegistriesFor1_20_3AreItsOwnAndNot1_20_5s(t *testing.T) {
	registries := func(load func() ([]Registry, error)) map[string]map[string]nbt.Tag {
		data := map[string]map[string]nbt.Tag{}
		for _, registry := range mustLoadRegistries(t, load) {
			data[registry.Name] = map[string]nbt.Tag{}
			for _, entry := range registry.Entries {
				data[registry.Name][entry.Name] = entry.Data
			}
		}

		return data
	}

	older, newer := registries(registriesMinecraft1_20_3), registries(registriesMinecraft1_20_5)

	if len(older) != 6 {
		t.Errorf("1.20.3 is sent %d registries, want 6", len(older))
	}

	for name, entries := range older {
		if len(entries) == 0 {
			t.Errorf("1.20.3 was sent an empty %s", name)
		}

		if _, ok := newer[name]; !ok {
			t.Errorf("1.20.3 is sent %s, which 1.20.5 is not", name)
		}

		for entryName, data := range entries {
			if data == nil {
				t.Errorf("1.20.3's %s entry %s has no definition, which the one-compound shape cannot send", name, entryName)
			}
		}
	}

	for _, name := range []string{"minecraft:wolf_variant", "minecraft:banner_pattern"} {
		if _, ok := older[name]; ok {
			t.Errorf("1.20.3 is sent %s, which 766 is where it became data-driven", name)
		}

		if _, ok := newer[name]; !ok {
			t.Errorf("1.20.5 is no longer sent %s, which its jar holds", name)
		}
	}

	if _, ok := older["minecraft:damage_type"]["minecraft:spit"]; ok {
		t.Error("1.20.3 is sent the spit damage type, which 766 introduced")
	}

	if _, ok := newer["minecraft:damage_type"]["minecraft:spit"]; !ok {
		t.Error("1.20.5 is no longer sent the spit damage type, which its jar holds")
	}

	// The dimension type the play login names by index on 1.20.5 is named
	// outright on 1.20.3, and the 1.20.5 step's rewrite writes the name of
	// the entry package gamedata puts first, so first is where it has to be.
	dimensionTypes := mustLoadRegistries(t, registriesMinecraft1_20_3)[0]
	if dimensionTypes.Name != "minecraft:dimension_type" || dimensionTypes.Entries[0].Name != "minecraft:overworld" {
		t.Errorf("1.20.3's first registry is %s starting with %s, want minecraft:dimension_type starting with minecraft:overworld", dimensionTypes.Name, dimensionTypes.Entries[0].Name)
	}

	tags := func(load func() ([]TagSet, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, set := range mustLoadTags(t, load) {
			names[set.Registry] = map[string]bool{}
			for _, tag := range set.Tags {
				names[set.Registry][tag.Name] = true
			}
		}

		return names
	}

	olderTags, newerTags := tags(tagsMinecraft1_20_3), tags(tagsMinecraft1_20_5)

	if len(olderTags) != len(newerTags)-1 {
		t.Errorf("1.20.3 declares tags for %d registries and 1.20.5 for %d, want one fewer for 1.20.3", len(olderTags), len(newerTags))
	}

	if _, ok := olderTags["minecraft:enchantment"]; ok {
		t.Error("1.20.3 declares enchantment tags, which 766 introduced")
	}

	if !olderTags["minecraft:item"]["minecraft:tools"] {
		t.Error("1.20.3 no longer declares tools, an item tag its jar does declare")
	}

	if newerTags["minecraft:item"]["minecraft:tools"] {
		t.Error("1.20.5 declares tools, an item tag 766 retired")
	}

	if olderTags["minecraft:item"]["minecraft:armadillo_food"] {
		t.Error("1.20.3 declares armadillo_food, an item tag 766 introduced")
	}

	if !olderTags["minecraft:block"]["minecraft:enchantment_power_provider"] {
		t.Error("1.20.3 no longer declares enchantment_power_provider, a block tag its jar does declare")
	}
}

// 1.20.2 synchronizes the same six registries as 1.20.3, and the same
// content in each of them, read off the two jars: 1.20.3 added no trim, no
// damage type and no chat type. The tag sets cover the same eleven
// registries, and differ by the four tags 1.20.3 introduced -- three entity
// type tags and one damage type tag -- since each version declares the tags
// of its own jar.
func TestRegistriesFor1_20_2AreItsOwnAndNot1_20_3s(t *testing.T) {
	registries := func(load func() ([]Registry, error)) map[string]map[string]nbt.Tag {
		data := map[string]map[string]nbt.Tag{}
		for _, registry := range mustLoadRegistries(t, load) {
			data[registry.Name] = map[string]nbt.Tag{}
			for _, entry := range registry.Entries {
				data[registry.Name][entry.Name] = entry.Data
			}
		}

		return data
	}

	older, newer := registries(registriesMinecraft1_20_2), registries(registriesMinecraft1_20_3)

	if len(older) != 6 {
		t.Errorf("1.20.2 is sent %d registries, want 6", len(older))
	}

	if len(older) != len(newer) {
		t.Errorf("1.20.2 is sent %d registries and 1.20.3 %d, want the same six", len(older), len(newer))
	}

	for name, entries := range older {
		if len(entries) == 0 {
			t.Errorf("1.20.2 was sent an empty %s", name)
		}

		newerEntries, ok := newer[name]
		if !ok {
			t.Errorf("1.20.2 is sent %s, which 1.20.3 is not", name)

			continue
		}

		if len(entries) != len(newerEntries) {
			t.Errorf("1.20.2's %s holds %d entries and 1.20.3's %d, want the same: 1.20.3 added nothing to it", name, len(entries), len(newerEntries))
		}

		for entryName, data := range entries {
			if data == nil {
				t.Errorf("1.20.2's %s entry %s has no definition, which the one-compound shape cannot send", name, entryName)
			}

			if _, ok := newerEntries[entryName]; !ok {
				t.Errorf("1.20.2's %s holds %s, which 1.20.3's does not", name, entryName)
			}
		}
	}

	// The dimension type the play login names outright on both sides of the
	// 1.20.3 step, so first is where it has to be here as well.
	dimensionTypes := mustLoadRegistries(t, registriesMinecraft1_20_2)[0]
	if dimensionTypes.Name != "minecraft:dimension_type" || dimensionTypes.Entries[0].Name != "minecraft:overworld" {
		t.Errorf("1.20.2's first registry is %s starting with %s, want minecraft:dimension_type starting with minecraft:overworld", dimensionTypes.Name, dimensionTypes.Entries[0].Name)
	}

	tags := func(load func() ([]TagSet, error)) map[string]map[string]bool {
		names := map[string]map[string]bool{}
		for _, set := range mustLoadTags(t, load) {
			names[set.Registry] = map[string]bool{}
			for _, tag := range set.Tags {
				names[set.Registry][tag.Name] = true
			}
		}

		return names
	}

	olderTags, newerTags := tags(tagsMinecraft1_20_2), tags(tagsMinecraft1_20_3)

	if len(olderTags) != len(newerTags) {
		t.Errorf("1.20.2 declares tags for %d registries and 1.20.3 for %d, want the same eleven", len(olderTags), len(newerTags))
	}

	for _, name := range []string{"minecraft:undead", "minecraft:zombies", "minecraft:can_breathe_under_water"} {
		if olderTags["minecraft:entity_type"][name] {
			t.Errorf("1.20.2 declares %s, an entity type tag 765 introduced", name)
		}

		if !newerTags["minecraft:entity_type"][name] {
			t.Errorf("1.20.3 no longer declares %s, an entity type tag its jar does declare", name)
		}
	}

	if olderTags["minecraft:damage_type"]["minecraft:can_break_armor_stand"] {
		t.Error("1.20.2 declares can_break_armor_stand, a damage type tag 765 introduced")
	}

	if !olderTags["minecraft:item"]["minecraft:tools"] {
		t.Error("1.20.2 no longer declares tools, an item tag its jar does declare")
	}

	if !olderTags["minecraft:block"]["minecraft:enchantment_power_provider"] {
		t.Error("1.20.2 no longer declares enchantment_power_provider, a block tag its jar does declare")
	}
}

// 1.20.3's dimension type is 1.21.9's but for the monster spawn light level,
// whose int provider 1.20.4 reads with its fields nested under value, the
// way a dispatch lays out a plain codec; the flat shape is 1.20.5's, where
// the provider's codec became a map codec. Everything else has to stay
// 1.21.9's, so that the one difference is the one the codecs have.
func TestDimensionTypeFor1_20_3NestsTheIntProvider(t *testing.T) {
	var dimensionType nbt.Compound

	for _, registry := range mustLoadRegistries(t, registriesMinecraft1_20_3) {
		if registry.Name == "minecraft:dimension_type" {
			dimensionType = registry.Entries[0].Data.(nbt.Compound)
		}
	}

	if dimensionType == nil {
		t.Fatal("1.20.3 has no dimension type")
	}

	provider, ok := dimensionType["monster_spawn_light_level"].(nbt.Compound)
	if !ok {
		t.Fatalf("monster_spawn_light_level = %v, want a compound", dimensionType["monster_spawn_light_level"])
	}

	if provider["type"] != nbt.String("minecraft:uniform") {
		t.Errorf("monster_spawn_light_level type = %v, want minecraft:uniform", provider["type"])
	}

	value, ok := provider["value"].(nbt.Compound)
	if !ok {
		t.Fatalf("monster_spawn_light_level value = %v, want the provider's fields nested under it", provider["value"])
	}

	if value["min_inclusive"] != nbt.Int(0) || value["max_inclusive"] != nbt.Int(7) {
		t.Errorf("monster_spawn_light_level value = %v, want min_inclusive 0 and max_inclusive 7", value)
	}

	for _, key := range []string{"min_inclusive", "max_inclusive"} {
		if _, ok := provider[key]; ok {
			t.Errorf("monster_spawn_light_level carries %s beside the type, which is 1.20.5's flat shape", key)
		}
	}

	for key, want := range overworldDimensionType1_21_9 {
		if key == "monster_spawn_light_level" {
			continue
		}

		if got := dimensionType[key]; got.String() != want.String() {
			t.Errorf("%s = %v, want 1.21.9's %v", key, got, want)
		}
	}

	if len(dimensionType) != len(overworldDimensionType1_21_9) {
		t.Errorf("1.20.3's dimension type has %d fields and 1.21.9's %d, want the same fields", len(dimensionType), len(overworldDimensionType1_21_9))
	}
}

func TestDimensionTypeFor1_21_9KeepsThePreReworkSchema(t *testing.T) {
	var dimensionType nbt.Compound

	for _, registry := range mustLoadRegistries(t, registriesMinecraft1_21_9) {
		if registry.Name != "minecraft:dimension_type" {
			continue
		}

		if len(registry.Entries) != 1 {
			t.Fatalf("expected one dimension type, got %d", len(registry.Entries))
		}

		dimensionType = registry.Entries[0].Data.(nbt.Compound)
	}

	if dimensionType == nil {
		t.Fatal("1.21.9 is sent no dimension type at all")
	}

	// The behaviour booleans and the effects identifier the rework removed are
	// fields 1.21.9's codec requires.
	for _, required := range []string{
		"ultrawarm", "natural", "bed_works", "respawn_anchor_works",
		"piglin_safe", "has_raids", "effects",
	} {
		if _, ok := dimensionType[required]; !ok {
			t.Errorf("1.21.9 is not sent %s, which its codec requires", required)
		}
	}

	// Infiniburn is the name of a block tag here too: the rework changed it to
	// a set of blocks only in 26.2.
	if infiniburn, ok := dimensionType["infiniburn"].(nbt.String); !ok || infiniburn != "#minecraft:infiniburn_overworld" {
		t.Errorf("1.21.9 infiniburn is %v, want the block tag the client's own overworld names", dimensionType["infiniburn"])
	}

	// The vertical bounds still have to agree with the chunks the play phase
	// sends, whatever else the schema does.
	if dimensionType["min_y"] != nbt.Int(-64) || dimensionType["height"] != nbt.Int(384) {
		t.Errorf("1.21.9 bounds are %v..%v, want -64 and 384", dimensionType["min_y"], dimensionType["height"])
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

	oldest000 := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_20_2)
	oldest00 := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_20_3)
	oldest0 := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_20_5)
	oldest1 := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_21)
	oldest2 := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_21_2)
	earliest := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_21_4)
	first := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_21_5)
	oldest := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_21_6)
	old7 := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_21_7)
	old9 := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_21_9)
	old := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_1_21_11)
	older := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_26_1)
	newer := provider.PacketsFor(types.ProtocolVersions.MINECRAFT_26_2)

	if len(oldest000) == 0 || len(oldest00) == 0 || len(oldest0) == 0 || len(oldest1) == 0 || len(oldest2) == 0 || len(earliest) == 0 || len(first) == 0 || len(oldest) == 0 || len(old7) == 0 || len(old9) == 0 || len(old) == 0 || len(older) == 0 {
		t.Fatal("a version was sent no packets at all, which is a client that never reaches the world")
	}

	// 1.20.3 takes every registry in one packet, so it is sent that and the
	// tags: two packets, however many registries it synchronizes.
	if len(oldest00) != 2 {
		t.Errorf("1.20.3 was sent %d packets, want the one holding every registry and the tags", len(oldest00))
	}

	// And so does 1.20.2, which is where that shape came from.
	if len(oldest000) != 2 {
		t.Errorf("1.20.2 was sent %d packets, want the one holding every registry and the tags", len(oldest000))
	}

	// 1.20.5 synchronizes three registries fewer than 1.21, which the count sees.
	if len(oldest0) >= len(oldest1) {
		t.Errorf("1.20.5 was sent %d packets and 1.21 %d, want fewer for 1.20.5", len(oldest0), len(oldest1))
	}

	// 1.21 synchronizes one registry fewer than 1.21.2, which the count sees.
	if len(oldest1) >= len(oldest2) {
		t.Errorf("1.21 was sent %d packets and 1.21.2 %d, want fewer for 1.21", len(oldest1), len(oldest2))
	}

	// 1.21.2 and 1.21.4 synchronize the same 12 registries, so the count
	// cannot tell them apart; the content test above does.
	if len(oldest2) != len(earliest) {
		t.Errorf("1.21.2 was sent %d packets and 1.21.4 %d, want the same count for the same registry list", len(oldest2), len(earliest))
	}

	// 1.21.4 synchronizes eight registries fewer than 1.21.5, which the count
	// sees.
	if len(earliest) >= len(first) {
		t.Errorf("1.21.4 was sent %d packets and 1.21.5 %d, want fewer for 1.21.4", len(earliest), len(first))
	}

	// 1.21.5 synchronizes one registry fewer than 1.21.6, which the count sees.
	if len(first) >= len(oldest) {
		t.Errorf("1.21.5 was sent %d packets and 1.21.6 %d, want fewer for 1.21.5", len(first), len(oldest))
	}

	// 1.21.6, 1.21.7 and 1.21.9 synchronize the same 21 registries, so the
	// count cannot tell them apart; the content tests above do.
	if len(oldest) != len(old7) {
		t.Errorf("1.21.6 was sent %d packets and 1.21.7 %d, want the same count for the same registry list", len(oldest), len(old7))
	}

	if len(old7) != len(old9) {
		t.Errorf("1.21.7 was sent %d packets and 1.21.9 %d, want the same count for the same registry list", len(old7), len(old9))
	}

	if len(old9) >= len(old) {
		t.Errorf("1.21.9 was sent %d packets and 1.21.11 %d, want fewer for 1.21.9", len(old9), len(old))
	}

	if len(old) >= len(older) {
		t.Errorf("1.21.11 was sent %d packets and 26.1 %d, want fewer for 1.21.11", len(old), len(older))
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
