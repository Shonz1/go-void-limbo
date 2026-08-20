package gamedata

// registriesMinecraft1_21_7 is every registry a 1.21.7 client is sent, which
// is all 21 of SYNCHRONIZED_REGISTRIES from its own sources for that version.
// The list is the same as 1.21.9's -- 773 added no synchronized registry --
// but the contents are not shared: 1.21.9's copper additions reach into
// trim_material and enchantment, and into the block, item and entity tags, so
// the nineteen generated registries come from 1.21.7's own jar. See
// data/minecraft_1_21_7.json.
//
// The two below are a limbo's own decision, as they are for the later
// versions. They are 1.21.9's entries by reference rather than new ones,
// because the dimension type and biome codecs are identical in 772 and 773 --
// the schema rework that would separate them happened in 1.21.11.
func registriesMinecraft1_21_7() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_21_9},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_21_7.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_21_7 is every tag a 1.21.7 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are: the
// client resolves tags while building item components, and a tag it asks for
// that was never declared throws rather than defaulting to empty.
func tagsMinecraft1_21_7() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_21_7.json")

	return tags, err
}
