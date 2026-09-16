package gamedata

// registriesMinecraft26_2 is every registry a 26.2 client is sent, which is all
// 29 of SYNCHRONIZED_REGISTRIES from its own sources for that version. 26.3 has
// the same 29 and three on top: the decorated pot patterns, the block
// transformers and the block state providers.
//
// The two below are a limbo's own decision, as they are for 26.3, and the
// compounds are 26.3's, since the two versions' dimension type and biome
// codecs ask for the same fields. The other twenty-seven come from 26.2's own
// data rather than from 26.3's, because the two versions disagree about the
// contents of entries they both have as well as about which entries exist.
// See data/minecraft_26_2.json.
func registriesMinecraft26_2() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome},
		}},
	}

	generated, _, err := loadDataFile("minecraft_26_2.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft26_2 is every tag a 26.2 client is sent, generated from its own
// jar alongside the registries, for the same reason.
func tagsMinecraft26_2() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_26_2.json")

	return tags, err
}
