package gamedata

import (
	"github.com/Shonz1/go-void-limbo/nbt"
	"maps"
)

// registriesMinecraft26_1 is every registry a 26.1 client is sent, which is all
// 28 of SYNCHRONIZED_REGISTRIES from its own sources for that version. 26.2 has
// the same 28 and sulfur_cube_archetype on top.
//
// The two below are a limbo's own decision, as they are for 26.2. The other
// twenty-six come from 26.1's own data rather than from 26.2's, because the two
// versions disagree about the contents of entries they both have as well as
// about which entries exist. See data/minecraft_26_1.json.
func registriesMinecraft26_1() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType26_1},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome},
		}},
	}

	generated, _, err := loadDataFile("minecraft_26_1.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft26_1 is every tag a 26.1 client is sent, generated from its own
// jar alongside the registries, for the same reason.
func tagsMinecraft26_1() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_26_1.json")

	return tags, err
}

// overworldDimensionType26_1 is the overworld dimension type as 26.1 reads it.
//
// The one field that differs is infiniburn. 26.2 made it a set of blocks, which
// is what the 26.2 entry sends as an empty list; 26.1 has the name of a block
// tag there and its codec accepts nothing else, so an empty list is a dimension
// type it cannot read and a registry it rejects.
//
// The tag named here is the client's own, and naming it costs nothing: 26.1
// parses the name into a reference without resolving it, so a tag with no
// members -- which is every tag a limbo declares -- is not a tag it complains
// about. Nothing burns either way.
var overworldDimensionType26_1 = func() nbt.Compound {
	dimensionType := maps.Clone(overworldDimensionType)
	dimensionType["infiniburn"] = nbt.String("#minecraft:infiniburn_overworld")

	return dimensionType
}()
