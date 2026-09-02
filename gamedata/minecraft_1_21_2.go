package gamedata

// registriesMinecraft1_21_2 is every registry a 1.21.2 client is sent, which
// is all 12 of SYNCHRONIZED_REGISTRIES from its own sources for that version
// -- the same twelve 1.21.4 synchronizes, since 1.21.4 added no registry and
// retired none. The ten generated registries come from 1.21.2's own jar. They
// hold the entries 1.21.4's do but one -- 1.21.4 is where the resin trim
// material landed -- and not the same content: the trim material still
// carries the item model index that 1.21.4's item model rework made
// redundant, and 1.21.4 retuned two enchantments. See
// data/minecraft_1_21_2.json.
//
// The wolf variant's biomes go out empty, as they do for 1.21.4 and for the
// same reason: the codec is the same set of biome references, and a reference
// to a biome the client was never sent is a holder that never binds.
//
// The two below are a limbo's own decision, as they are for the later
// versions. They are 1.21.9's entries by reference rather than new ones: the
// dimension type codec is identical from 768 through 770, and the biome
// codec differs only in what 1.21.4 added to it -- an optional music volume
// and a weighted music list -- neither of which the plains entry carries a
// value for that 1.21.2 would read. The music volume it does carry is a key
// 1.21.2's codec does not name, and a key a codec does not name is one it
// skips.
func registriesMinecraft1_21_2() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_21_9},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_21_2.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_21_2 is every tag a 1.21.2 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are: the
// client resolves tags while building item components, and a tag it asks for
// that was never declared throws rather than defaulting to empty.
func tagsMinecraft1_21_2() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_21_2.json")

	return tags, err
}
