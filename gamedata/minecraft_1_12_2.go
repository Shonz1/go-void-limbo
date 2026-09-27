package gamedata

// registriesMinecraft1_12_2 is what a 1.12.2 client is sent of the
// registries, which is nothing, as it is for 1.13.2: see
// registriesMinecraft1_13_2.
func registriesMinecraft1_12_2() ([]Registry, error) {
	return nil, nil
}

// tagsMinecraft1_12_2 is what a 1.12.2 client is sent of the tags, which is
// nothing as well: 1.13 is where the tags went on the wire, and 1.12.2 has no
// packet to read them from. Its set is empty, so it sends no packet at all.
func tagsMinecraft1_12_2() ([]TagSet, error) {
	return nil, nil
}
