package gamedata

import "github.com/Shonz1/go-void-limbo/nbt"

// registriesMinecraft1_16_4 is every registry a 1.16.4 client is sent, which
// is the same 2 as 1.17's: the dimension types and the biomes, the two its
// own RegistryAccess marks as sent to the client, 1.17 having added none.
// Nothing here is generated, for the reason 1.17.1's are not: see
// data/minecraft_1_16_4.json, which holds the tags alone.
//
// The registries go out the way 1.17's do, inside the play login with the
// dimension type spelled out alongside them, and the 1.17 step's login
// transformer swaps both in for 1.17's: see Provider.DimensionTypeFor.
//
// The biome is 1.17.1's by reference, since the codec is field for field the
// same in the two, read off the jars, depth and scale included. The
// dimension type is 1.16.4's own: see overworldDimensionType1_16_4.
//
// A 1.16.5 client is sent this set as well, since it speaks 1.16.4's
// protocol, and so is a 1.16.3 or a 1.16.2 client, whose jar holds the same
// tags and the same codecs: the set starts at 751.
func registriesMinecraft1_16_4() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_16_4},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_17_1},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_16_4.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// overworldDimensionType1_16_4 is the overworld dimension type as 1.16.4
// reads it: the 1.18 entry that serves 1.17 without the two fields 1.17
// added, the min_y and the height, and with the logical height that goes
// with them. 1.16.4's world is 256 blocks from zero up whatever it is told,
// and its codec holds the logical height to that: 384 is a value the client
// refuses. Every other field has the same codec in the two, read off the
// jars, the infiniburn tag named plainly in both.
var overworldDimensionType1_16_4 = nbt.Compound{
	"ambient_light":        nbt.Float(0),
	"bed_works":            nbt.Byte(1),
	"coordinate_scale":     nbt.Double(1),
	"effects":              nbt.String("minecraft:overworld"),
	"has_ceiling":          nbt.Byte(0),
	"has_raids":            nbt.Byte(1),
	"has_skylight":         nbt.Byte(1),
	"infiniburn":           nbt.String("minecraft:infiniburn_overworld"),
	"logical_height":       nbt.Int(256),
	"natural":              nbt.Byte(1),
	"piglin_safe":          nbt.Byte(0),
	"respawn_anchor_works": nbt.Byte(0),
	"ultrawarm":            nbt.Byte(0),
}

// tagsMinecraft1_16_4 is every tag a 1.16.4 client's jar declares, generated
// for the same reason the later versions' are. The sets cover four
// registries to 1.17's five: the blocks, the items, the fluids and the
// entity types, 1.17 being where the game events and their tags appeared.
// They are listed in that order because it is the order 1.16.4 reads them
// in, with no names in front: see encodeTags1_16_4. Within the four the sets
// are 1.17's minus what 1.17 added -- the caves and cliffs blocks' tags, the
// candles, the ores by metal among them. As on 1.17, the client checks the
// payload for every tag its own helpers name and leaves over a missing one,
// which the generation covers by listing every tag the jar declares. The
// tags are sent as a play packet right after the login, since there is no
// configuration phase to send them in.
func tagsMinecraft1_16_4() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_16_4.json")

	return tags, err
}
