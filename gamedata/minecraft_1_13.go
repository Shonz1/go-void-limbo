package gamedata

// tagsMinecraft1_13 is the tags of the 1.13 jar, in the order that version
// reads them: see encodeTags1_13_2. They are 1.13.1's less the two tags 1.13.1
// added with its dead coral plants -- the coral plants, which 1.13's corals
// list by name, and the underwater bonemeals. The set is 393's alone, and is
// sent no registry, as 401's is not: see registriesMinecraft1_13_2.
func tagsMinecraft1_13() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_13.json")

	return tags, err
}
