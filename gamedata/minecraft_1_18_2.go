package gamedata

import "github.com/Shonz1/go-void-limbo/nbt"

// registriesMinecraft1_18_2 is every registry a 1.18.2 client is sent, which
// is the 2 its own RegistryAccess marks as sent to the client: the dimension
// types and the biomes. 1.19 is where the chat types joined them, so a
// 1.18.2 client reads nothing of those, and a compound naming a registry it
// does not know is one it refuses. Nothing here is generated: the two
// registries are a limbo's own decision, as they are for the later versions,
// and there is no third. See data/minecraft_1_18_2.json, which holds the
// tags alone.
//
// The registries go out the way 1.19's do, inside the play login as the one
// compound a client before 1.20.2 reads them from, and the 1.19 step's login
// transformer swaps it in for 1.19's. 1.18.2 reads one thing more out of
// that login than 1.19 does: the dimension type it is put into, spelled out
// as the entry itself rather than named, which the provider hands the same
// transformer as well: see Provider.DimensionTypeFor.
//
// Both entries are 1.18.2's own, for a field each: see
// overworldDimensionType1_18_2 and plainsBiome1_18_2.
func registriesMinecraft1_18_2() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_18_2},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_18_2},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_18_2.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// overworldDimensionType1_18_2 is the overworld dimension type as 1.18.2
// reads it: the 1.20.3 entry that serves every version from 1.19 to 1.20.3,
// field for field, less the two monster spawn fields. 1.19 is where the
// monster spawn light level and its block light limit joined the codec, and
// 1.18.2 has no field for either; every other field has the same codec in
// the two, read off the jars, the infiniburn tag included, since 1.18.2 is
// where tag keys appeared and the field took its hash.
var overworldDimensionType1_18_2 = nbt.Compound{
	"ambient_light":        nbt.Float(0),
	"bed_works":            nbt.Byte(1),
	"coordinate_scale":     nbt.Double(1),
	"effects":              nbt.String("minecraft:overworld"),
	"has_ceiling":          nbt.Byte(0),
	"has_raids":            nbt.Byte(1),
	"has_skylight":         nbt.Byte(1),
	"height":               nbt.Int(384),
	"infiniburn":           nbt.String("#minecraft:infiniburn_overworld"),
	"logical_height":       nbt.Int(384),
	"min_y":                nbt.Int(-64),
	"natural":              nbt.Byte(1),
	"piglin_safe":          nbt.Byte(0),
	"respawn_anchor_works": nbt.Byte(0),
	"ultrawarm":            nbt.Byte(0),
}

// plainsBiome1_18_2 is what every chunk a limbo sends is made of, as 1.18.2
// reads it: the 1.19.3 compound that serves 1.19 and 1.19.1, with the
// biome's category on top. 1.19 is where the category left the codec, for
// the biome tags to say what it said, and 1.18.2 reads it as a required
// field: a compound without it is a biome the client refuses. The value is
// the one the client's own plains carries, and the climate and the colours
// are as 1.19 reads them, since those fields have the same codec in the two.
var plainsBiome1_18_2 = nbt.Compound{
	"category":      nbt.String("plains"),
	"precipitation": nbt.String("none"),
	"temperature":   nbt.Float(0.8),
	"downfall":      nbt.Float(0.4),
	"effects": nbt.Compound{
		"fog_color":       nbt.Int(12638463),
		"sky_color":       nbt.Int(7907327),
		"water_color":     nbt.Int(4159204),
		"water_fog_color": nbt.Int(329011),
		"music_volume":    nbt.Float(1),
	},
}

// tagsMinecraft1_18_2 is every tag a 1.18.2 client's jar declares for a
// registry a vanilla 1.18.2 server sends tags for, generated alongside the
// registries, for the same reason the later versions' are. The sets cover
// six registries to 1.19's ten: the blocks, the items, the entity types,
// the fluids, the game events and the biomes. 1.19 is where the points of
// interest, the instruments, the painting variants and the banner patterns
// came to have tags. The jar declares tags for the configured structure
// features as well, which are left out: that registry is not one a 1.18.2
// server sends to the client, and a 1.18.2 client answers a tag set for a
// registry it does not hold by leaving. Within the six the sets differ from
// 1.19's by what each jar declares: 1.19's additions, and the four tags
// 1.19 renamed -- the carpets on the blocks and on the items, the frozen
// ocean polar bear tag on the blocks, and the occludes vibration signals
// tag on the items. 1.18.2's jar keeps its
// tag directories under the plural names 1.21 dropped, as 1.19's does,
// which only the generation sees. The tags are sent as a play packet right
// after the login, since there is no configuration phase to send them in.
func tagsMinecraft1_18_2() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_18_2.json")

	return tags, err
}
