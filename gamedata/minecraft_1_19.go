package gamedata

// registriesMinecraft1_19 is every registry a 1.19 client is sent, which is
// the same 3 a 1.19.1 client is: the dimension types, the biomes and the
// chat types, the three the version's own RegistryAccess marks as sent to
// the client. The one generated registry comes from 1.19's own jar, by way
// of its data generator, since the chat types lived in code until 1.19.3,
// and it is the one thing 1.19 reads differently from 1.19.1: 1.19.1 is
// where the chat types were reworked, from a chat, an overlay and a
// narration each optional and each with its own decoration, into the two
// decorations every version from it on reads, and from eight entries into
// seven. See data/minecraft_1_19.json.
//
// The registries go out the way 1.19.1's do, inside the play login as the
// one compound a client before 1.20.2 reads them from, and the 1.19.1 step's
// login transformer swaps it in for 1.19.1's.
//
// The dimension type below is the same entry 1.19.1 is sent, by reference:
// the codec is field-for-field identical in 1.19 and 1.19.1, read off the
// two jars, and the uniform int provider is the plain codec in both, so the
// nested monster spawn light level 1.20.3 needs is what 1.19 needs too. The
// biome is 1.19.3's own as well, by reference: the climate is spelled the
// same way in the three, with the precipitation as a name, since 1.19.4 is
// where it changed shape.
func registriesMinecraft1_19() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_20_3},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_19_3},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_19.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_19 is every tag a 1.19 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are.
// The sets cover the same ten registries 1.19.1's do and declare the same
// tags, name for name: 1.19.1 added none and retired none, which the
// generation checked by producing both files from the two jars. 1.19's jar
// keeps its tag directories under the plural names 1.21 dropped, as
// 1.19.1's does, which only the generation sees. The tags are sent as a
// play packet right after the login, since there is no configuration phase
// to send them in.
func tagsMinecraft1_19() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_19.json")

	return tags, err
}
