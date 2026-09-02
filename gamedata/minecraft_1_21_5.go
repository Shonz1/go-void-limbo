package gamedata

// registriesMinecraft1_21_5 is every registry a 1.21.5 client is sent, which
// is all 20 of SYNCHRONIZED_REGISTRIES from its own sources for that version.
// The list is 1.21.6's minus dialog, the registry 1.21.6 added for its
// dialogs, and the eighteen generated registries come from 1.21.5's own jar:
// the only entry the two disagree about is the tears jukebox song, which
// landed in 771. See data/minecraft_1_21_5.json.
//
// The two below are a limbo's own decision, as they are for the later
// versions. They are 1.21.9's entries by reference rather than new ones,
// because the dimension type and biome codecs read the same shape from 770
// through 773 -- 1.21.6 added cloud_height to the dimension type, but as an
// optional field the entry leaves to its default, so 1.21.5 not knowing it
// changes nothing -- and the schema rework that would separate them happened
// in 1.21.11.
func registriesMinecraft1_21_5() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_21_9},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_21_5.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_21_5 is every tag a 1.21.5 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are: the
// client resolves tags while building item components, and a tag it asks for
// that was never declared throws rather than defaulting to empty.
func tagsMinecraft1_21_5() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_21_5.json")

	return tags, err
}
