package gamedata

// registriesMinecraft1_21_4 is every registry a 1.21.4 client is sent, which
// is all 12 of SYNCHRONIZED_REGISTRIES from its own sources for that version.
// 1.21.5 is where the list grew by eight: it made the cat, frog, pig, cow and
// chicken variants and the wolf sound variants data-driven, and added the
// two test registries. The ten generated registries come from 1.21.4's own
// jar. They hold exactly the entries 1.21.5's do, but not the same content:
// 1.21.5 reworked the wolf variant into a set of assets, and moved the trim
// material's ingredient out to the item. See data/minecraft_1_21_4.json.
//
// The wolf variant is the one place the generated content is not the jar's
// verbatim. Its 1.21.4 codec still carries the biomes the variant spawns in,
// as a set of biome references, and the client resolves those against the
// biome registry it was sent -- which holds one plains and nothing else, so
// a reference to a taiga is a holder that never binds, and a registry with
// one of those in it fails to freeze. The set is sent empty instead, which
// is what a tag reference becomes everywhere in these files, for the same
// reason.
//
// The two below are a limbo's own decision, as they are for the later
// versions. They are 1.21.9's entries by reference rather than new ones,
// because the dimension type codec is identical from 769 through 770 and the
// biome codec only gained an optional field along the way, so the entries
// 1.21.5 reads serve 1.21.4 as they are.
func registriesMinecraft1_21_4() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_21_9},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_21_4.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_21_4 is every tag a 1.21.4 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are: the
// client resolves tags while building item components, and a tag it asks for
// that was never declared throws rather than defaulting to empty.
func tagsMinecraft1_21_4() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_21_4.json")

	return tags, err
}
