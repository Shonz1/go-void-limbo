package gamedata

// registriesMinecraft1_12_2 is what a 1.12.2 client is sent of the
// registries, which is nothing, as it is for 1.13.2: see
// registriesMinecraft1_13_2. A 1.12.1 client is sent the same nothing, having
// the same packets, and a 1.12, a 1.11.1, a 1.11, a 1.10, a 1.9.3, a 1.9.2,
// a 1.9.1, a 1.9 and a 1.8 client too; the set starts at 47.
func registriesMinecraft1_12_2() ([]Registry, error) {
	return nil, nil
}

// tagsMinecraft1_12_2 is what a 1.12.2 client is sent of the tags, which is
// nothing as well: 1.13 is where the tags went on the wire, and 1.12.2 has no
// packet to read them from, nor has 1.12.1, 1.12, 1.11.1, 1.11, 1.10, 1.9.3,
// 1.9.2, 1.9.1, 1.9 or 1.8. Its set is empty, so it sends no packet at all.
func tagsMinecraft1_12_2() ([]TagSet, error) {
	return nil, nil
}
