package gamedata

// registriesMinecraft1_21 is every registry a 1.21 client is sent, which is
// all 11 of SYNCHRONIZED_REGISTRIES from its own sources for that version:
// 1.21.2's twelve minus the instrument, which 1.21.2 is where it became
// data-driven. The nine generated registries come from 1.21.1's own jar, and
// the two 1.21.2 has that 1.21 does not are the two 1.21.2 added -- two
// damage types, the mace smash and the ender pearl -- while 1.21.2 also gave
// every painting a title and an author and retuned ten enchantments, none of
// which 1.21's codecs would read. See data/minecraft_1_21.json.
//
// The wolf variant's biomes go out empty, as they do for the versions above
// and for the same reason: the codec is the same set of biome references, and
// a reference to a biome the client was never sent is a holder that never
// binds.
//
// The two below are a limbo's own decision, as they are for the later
// versions. They are 1.21.9's entries by reference rather than new ones: the
// dimension type and biome codecs are byte-identical in 1.21 and 1.21.2, and
// what the versions between there and 1.21.9 added to them the plains entry
// carries nothing for that 1.21 would read, as minecraft_1_21_2.go sets out.
func registriesMinecraft1_21() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_21_9},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_21.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_21 is every tag a 1.21 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are: the
// client resolves tags while building item components, and a tag it asks for
// that was never declared throws rather than defaulting to empty.
func tagsMinecraft1_21() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_21.json")

	return tags, err
}
