package gamedata

// registriesMinecraft1_19_1 is every registry a 1.19.1 client is sent, which
// is the same 3 a 1.19.3 client is: the dimension types, the biomes and the
// chat types, the three the version's own RegistryAccess marks as sent to
// the client. 1.19.2 shares the protocol and reads the same. The one
// generated registry comes from 1.19.1's own jar, by way of its data
// generator, since 1.19.3 is where the chat types moved from code into data
// files: its chat types are 1.19.3's entry for entry. See
// data/minecraft_1_19_1.json.
//
// The registries go out the way 1.19.3's do, inside the play login as the
// one compound a client before 1.20.2 reads them from, and the 1.19.3 step's
// login transformer swaps it in for 1.19.3's.
//
// The dimension type below is the same entry 1.19.3 is sent, by reference:
// the codec is field-for-field identical in 1.19.1 and 1.19.3, read off the
// two jars, and the uniform int provider is the plain codec in both, so the
// nested monster spawn light level 1.20.3 needs is what 1.19.1 needs too.
// The biome is 1.19.3's own as well, by reference: the climate is spelled
// the same way in the two, with the precipitation as a name, since 1.19.4 is
// where it changed shape.
func registriesMinecraft1_19_1() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_20_3},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_19_3},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_19_1.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_19_1 is every tag a 1.19.1 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are.
// The sets cover the same ten registries 1.19.3's do, and differ from them
// by what each jar declares: 1.19.1 still has the non-flammable wood block
// tag and the overworld natural logs item tag, and lacks the sign, spawn
// and fence gate tags 1.19.3 added. 1.19.1's jar keeps its tag directories
// under the plural names 1.21 dropped, as 1.19.3's does, which only the
// generation sees. The tags are sent as a play packet right after the
// login, since there is no configuration phase to send them in.
func tagsMinecraft1_19_1() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_19_1.json")

	return tags, err
}
