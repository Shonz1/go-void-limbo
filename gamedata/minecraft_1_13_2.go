package gamedata

// registriesMinecraft1_13_2 is what a 1.13.2 client is sent of the
// registries, which is nothing, as it is for 1.15.2: see
// registriesMinecraft1_15_2.
func registriesMinecraft1_13_2() ([]Registry, error) {
	return nil, nil
}

// tagsMinecraft1_13_2 is the tags of the 1.13.2 jar, in the order that
// version reads them: see encodeTags1_13_2. There are three sets of them
// rather than four, the entity types having none before 1.14, and the three
// are 1.14's less what came with the village and pillage update and a
// handful beside -- the beds, the fences, the signs, the walls, the small
// flowers, the arrows and the music discs among them. No tag of 1.13.2's
// was retired on the way.
func tagsMinecraft1_13_2() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_13_2.json")

	return tags, err
}
