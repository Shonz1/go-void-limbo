package gamedata

// registriesMinecraft1_15_2 is what a 1.15.2 client is sent of the
// registries, which is nothing: it is from before the dimension types became
// data, holds every dimension in its own code as it holds every biome, and
// is put into one by number. See the 1.16 step's login transformer.
func registriesMinecraft1_15_2() ([]Registry, error) {
	return nil, nil
}

// tagsMinecraft1_15_2 is the tags of the 1.15.2 jar, in the order that
// version reads them: see encodeTags1_16_4. They are 1.16's less what the
// nether update brought -- no tag of 1.15.2's was retired on the way.
func tagsMinecraft1_15_2() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_15_2.json")

	return tags, err
}
