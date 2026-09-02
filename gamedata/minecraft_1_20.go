package gamedata

// registriesMinecraft1_20 is every registry a 1.20 client is sent, which is
// all 6 of NETWORKABLE_REGISTRIES from its own sources for that version: the
// same six 1.20.2 synchronizes, read off the two jars. The four generated
// registries come from 1.20's own jar and hold, entry for entry, what
// 1.20.2's do, but for one field: 1.20.2 is where the trim patterns gained
// their decal flag, which 1.20 has no field for and so is not sent. See
// data/minecraft_1_20.json.
//
// The shape the registries go out in is not a packet at all: a client before
// 1.20.2 has no configuration phase to be sent them in, and reads them out of
// its play login instead, as the one compound 1.20.2 went on to send in a
// packet of its own. The provider encodes a set that starts below 1.20.2 that
// way, and the 1.20.2 step's login transformer writes the compound into the
// packet.
//
// The two below are the same entries 1.20.2 is sent, by reference: the
// dimension type and biome codecs are field-for-field identical in 1.20 and
// 1.20.2, read off the two jars, and the uniform int provider is the plain
// codec in both, so the nested monster spawn light level 1.20.3 needs is what
// 1.20 needs too.
func registriesMinecraft1_20() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_20_3},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_20.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_20 is every tag a 1.20 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are. The
// sets cover the same eleven registries as 1.20.2's, and differ from them by
// five tags 1.20.2 introduced: two block tags -- the concrete powders and
// what the camel steps on quietly -- two damage type tags -- what always
// kills armor stands and what deals no knockback -- and the entity type tag
// for the riders that do not steer. 1.20's jar keeps its tag directories
// under the plural names 1.21 dropped, as 1.20.2's does, which only the
// generation sees. The tags are sent as a play packet right after the login,
// since there is no configuration phase to send them in.
func tagsMinecraft1_20() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_20.json")

	return tags, err
}
