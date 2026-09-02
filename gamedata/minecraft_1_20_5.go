package gamedata

// registriesMinecraft1_20_5 is every registry a 1.20.5 client is sent, which
// is all 8 of SYNCHRONIZED_REGISTRIES from its own sources for that version:
// 1.21's eleven minus the three 1.21 is where they became data-driven -- the
// painting variant, the enchantment and the jukebox song. The six generated
// registries come from 1.20.6's own jar, and what 1.21 has that 1.20.5 does
// not is what 1.21 added: the flow and guster banner patterns and trim
// patterns, the bolt trim pattern, and the campfire and wind charge damage
// types. The trim materials, the wolf variants and the chat types are the
// same in both. See data/minecraft_1_20_5.json.
//
// The wolf variant's biomes go out empty, as they do for the versions above
// and for the same reason: the codec is the same set of biome references, and
// a reference to a biome the client was never sent is a holder that never
// binds.
//
// The two below are a limbo's own decision, as they are for the later
// versions. They are 1.21.9's entries by reference rather than new ones: the
// dimension type and biome codecs are byte-identical in 1.20.6 and 1.21.1,
// read off the two jars field by field, so what serves 1.21 serves 1.20.5.
func registriesMinecraft1_20_5() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_21_9},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_20_5.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_20_5 is every tag a 1.20.5 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are: the
// client resolves tags while building item components, and a tag it asks for
// that was never declared throws rather than defaulting to empty. The sets
// cover the same twelve registries as 1.21's; 1.20.6's jar keeps its tag
// directories under the plural names 1.21 dropped, which only the generation
// sees, since a set is named on the wire for the registry it belongs to.
func tagsMinecraft1_20_5() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_20_5.json")

	return tags, err
}
