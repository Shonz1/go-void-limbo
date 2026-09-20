package gamedata

import "github.com/Shonz1/go-void-limbo/nbt"

// registriesMinecraft1_16_1 is every registry a 1.16.1 client is sent, which
// is one: the dimension types. 1.16.2 is where the biomes became data and
// joined them in the play login; a 1.16.1 client holds every biome in its
// own code, numbered there, and its RegistryAccess holds the dimension types
// and nothing else.
//
// They go out inside the play login in a shape of their own, a list under
// "dimension" with each entry's name beside its fields, and the login names
// the one the player is put into rather than spelling it out a second time.
// The 1.16.2 step's login transformer swaps both in for 1.16.2's: see
// encodeDimensionList and Provider.DimensionTypeFor.
func registriesMinecraft1_16_1() ([]Registry, error) {
	return []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_16_1},
		}},
	}, nil
}

// overworldDimensionType1_16_1 is the overworld dimension type as 1.16.1
// reads it, read off its codec: 1.16.4's without the effects, which 1.16.2
// added, and with 1.16.2's coordinate scale still the flag it replaced --
// shrunk, which is what the nether is and the overworld is not.
var overworldDimensionType1_16_1 = nbt.Compound{
	"ambient_light":        nbt.Float(0),
	"bed_works":            nbt.Byte(1),
	"has_ceiling":          nbt.Byte(0),
	"has_raids":            nbt.Byte(1),
	"has_skylight":         nbt.Byte(1),
	"infiniburn":           nbt.String("minecraft:infiniburn_overworld"),
	"logical_height":       nbt.Int(256),
	"natural":              nbt.Byte(1),
	"piglin_safe":          nbt.Byte(0),
	"respawn_anchor_works": nbt.Byte(0),
	"shrunk":               nbt.Byte(0),
	"ultrawarm":            nbt.Byte(0),
}

// tagsMinecraft1_16_1 is every tag a 1.16.1 client's jar declares, in the
// four runs and the order 1.16.4 reads them in: see tagsMinecraft1_16_4. The
// sets are 1.16.2's but for five tags: 1.16.2 added the two base stones and
// the mushroom grow block among the blocks, and among the items traded the
// furnace materials for the stone crafting materials. A 1.16.1 client does
// not check the payload for the tags its helpers name, as the versions above
// it do, and is sent its own jar's all the same.
func tagsMinecraft1_16_1() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_16_1.json")

	return tags, err
}
