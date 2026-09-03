package gamedata

import "github.com/Shonz1/go-void-limbo/nbt"

// registriesMinecraft1_18 is every registry a 1.18 client is sent, which is
// the same 2 as 1.18.2's: the dimension types and the biomes, the two its
// own RegistryAccess marks as sent to the client, 1.18.2 having added none.
// Nothing here is generated, for the reason 1.18.2's are not: see
// data/minecraft_1_18.json, which holds the tags alone.
//
// The registries go out the way 1.18.2's do, inside the play login with the
// dimension type spelled out alongside them, and the 1.18.2 step's login
// transformer swaps both in for 1.18.2's: see Provider.DimensionTypeFor.
//
// The dimension type is 1.18's own, for one field: see
// overworldDimensionType1_18. The biome is 1.18.2's by reference, since the
// codec is field for field the same in the two, read off the jars.
func registriesMinecraft1_18() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_18},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_18_2},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_18.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// overworldDimensionType1_18 is the overworld dimension type as 1.18 reads
// it: the 1.18.2 entry, field for field, with the infiniburn tag named
// plainly rather than hashed. 1.18.2 is where tag keys appeared and the
// field took the hashed codec, whose value opens with the hash; 1.18 reads
// the field as a plain resource location, which the hash is not a valid
// character of, so a client on it refuses 1.18.2's value. Every other field
// has the same codec in the two, read off the jars.
var overworldDimensionType1_18 = nbt.Compound{
	"ambient_light":        nbt.Float(0),
	"bed_works":            nbt.Byte(1),
	"coordinate_scale":     nbt.Double(1),
	"effects":              nbt.String("minecraft:overworld"),
	"has_ceiling":          nbt.Byte(0),
	"has_raids":            nbt.Byte(1),
	"has_skylight":         nbt.Byte(1),
	"height":               nbt.Int(384),
	"infiniburn":           nbt.String("minecraft:infiniburn_overworld"),
	"logical_height":       nbt.Int(384),
	"min_y":                nbt.Int(-64),
	"natural":              nbt.Byte(1),
	"piglin_safe":          nbt.Byte(0),
	"respawn_anchor_works": nbt.Byte(0),
	"ultrawarm":            nbt.Byte(0),
}

// tagsMinecraft1_18 is every tag a 1.18 client's jar declares for a registry
// a vanilla 1.18 server sends tags for, generated alongside the registries,
// for the same reason the later versions' are. The sets cover five
// registries to 1.18.2's six: the blocks, the items, the entity types, the
// fluids and the game events, which are the five static tag helpers a 1.18
// server serializes. 1.18.2 is where the biomes came to have tags, along
// with the tag keys, and where the fall damage resetting block tag was
// added; within the five the sets otherwise declare the same tags name for
// name. The five matter more on 1.18 than on any version above it: a 1.18
// client checks the payload for every tag its own helpers name and leaves
// over a missing one, which the generation covers by listing every tag the
// jar declares, the helpers' among them. 1.18's jar keeps its tag
// directories under the plural names 1.21 dropped, as 1.18.2's does, which
// only the generation sees. The tags are sent as a play packet right after
// the login, since there is no configuration phase to send them in.
func tagsMinecraft1_18() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_18.json")

	return tags, err
}
