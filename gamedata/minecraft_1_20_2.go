package gamedata

// registriesMinecraft1_20_2 is every registry a 1.20.2 client is sent, which
// is all 6 of NETWORKABLE_REGISTRIES from its own sources for that version:
// the same six 1.20.3 synchronizes, read off the two jars. The four generated
// registries come from 1.20.2's own jar and hold, entry for entry, what
// 1.20.4's do: 1.20.3 added no trim, no damage type and no chat type. See
// data/minecraft_1_20_2.json.
//
// The shape the registries go out in is the one-compound one 1.20.3 reads,
// since 1.20.2 is where that shape came from: the packet and the compound
// inside it are wire-identical in the two, and the provider encodes every set
// that starts below 1.20.5 that way.
//
// The two below are the same entries 1.20.3 is sent, by reference: the
// dimension type and biome codecs are field-for-field identical in 1.20.2 and
// 1.20.4, read off the two jars, and the uniform int provider is the plain
// codec in both, so the nested monster spawn light level 1.20.3 needs is what
// 1.20.2 needs too.
func registriesMinecraft1_20_2() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_20_3},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_20_2.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_20_2 is every tag a 1.20.2 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are. The
// sets cover the same eleven registries as 1.20.3's, and differ from them by
// four tags 1.20.3 introduced: three entity type tags -- the undead, the
// zombies and what can breathe under water -- and the damage type tag for
// what can break an armor stand. 1.20.2's jar keeps its tag directories under
// the plural names 1.21 dropped, as 1.20.4's does, which only the generation
// sees.
func tagsMinecraft1_20_2() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_20_2.json")

	return tags, err
}
