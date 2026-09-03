package gamedata

// registriesMinecraft1_19_4 is every registry a 1.19.4 client is sent, which
// is all 6 of NETWORKABLE_REGISTRIES from its own sources for that version:
// the same six 1.20 synchronizes, read off the two jars. Two of them go out
// empty, which is not a shortcut but what a vanilla 1.19.4 server sends: the
// armor trims were 1.20's, and 1.19.4 keeps their patterns and materials
// behind the feature flag of its update pack, which a server declares the
// registries for and fills only when the flag is on. The two generated
// registries with entries come from 1.19.4's own jar: its chat types are
// 1.20's entry for entry, and its damage types are 1.20's minus the two 1.20
// introduced, for a kill by command and for the world border. See
// data/minecraft_1_19_4.json.
//
// The registries go out the way 1.20's do, inside the play login as the one
// compound a client before 1.20.2 reads them from, and the 1.20 step's login
// transformer swaps it in for 1.20's.
//
// The two below are the same entries 1.20 is sent, by reference: the
// dimension type and biome codecs are field-for-field identical in 1.19.4
// and 1.20, read off the two jars, and the uniform int provider is the plain
// codec in both, so the nested monster spawn light level 1.20.3 needs is what
// 1.19.4 needs too.
func registriesMinecraft1_19_4() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_20_3},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_19_4.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_19_4 is every tag a 1.19.4 client's jar declares outside
// its feature-flagged packs, generated alongside the registries, for the
// same reason the later versions' are. The sets cover the same eleven
// registries as 1.20's, and differ from them by what 1.20 brought out from
// behind the flag and added: the hanging signs, the cherry and bamboo wood,
// the sniffer, the decorated pots, the trims and the trail ruins have their
// block, item and biome tags in 1.20 and not here, and the one tag 1.20
// retired, the replaceable plants, is here and not there. 1.19.4's jar keeps
// its tag directories under the plural names 1.21 dropped, as 1.20's does,
// which only the generation sees. The tags are sent as a play packet right
// after the login, since there is no configuration phase to send them in.
func tagsMinecraft1_19_4() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_19_4.json")

	return tags, err
}
