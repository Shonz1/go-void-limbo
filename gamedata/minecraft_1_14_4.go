package gamedata

// registriesMinecraft1_14_4 is what a 1.14.4 client is sent of the
// registries, which is nothing, as it is for 1.15.2: see
// registriesMinecraft1_15_2.
func registriesMinecraft1_14_4() ([]Registry, error) {
	return nil, nil
}

// tagsMinecraft1_14_4 is the tags of the 1.14.4 jar, in the order that
// version reads them, which is 1.15.2's: see encodeTags1_16_4. They are
// 1.15's less what came with the bees and a handful beside -- the crops, the
// flowers, the portals, the shulker boxes, the lectern books and the arrows
// among them -- and with the one block tag 1.15 retired, the dirt-like.
//
// A 1.14.3 client is sent this set as well, and a 1.14.2 client, their jars
// holding the same tags, and a 1.14.1 client, whose jar holds tags of the
// same names: what 1.14.2 changed is which slabs the slabs are, and a tag is
// sent here by its name alone. And a 1.14 client, whose jar holds the same
// tags as 1.14.1's, contents and all. The set starts at 477.
func tagsMinecraft1_14_4() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_14_4.json")

	return tags, err
}
