package gamedata

import "github.com/Shonz1/go-void-limbo/nbt"

// registriesMinecraft1_17_1 is every registry a 1.17.1 client is sent, which
// is the same 2 as 1.18's: the dimension types and the biomes, the two its
// own RegistryAccess marks as sent to the client, 1.18 having added none.
// Nothing here is generated, for the reason 1.18's are not: see
// data/minecraft_1_17_1.json, which holds the tags alone.
//
// The registries go out the way 1.18's do, inside the play login with the
// dimension type spelled out alongside them, and the 1.18 step's login
// transformer swaps both in for 1.18's: see Provider.DimensionTypeFor.
//
// The dimension type is 1.18's by reference, since the codec is field for
// field the same in the two, read off the jars, the infiniburn tag named
// plainly in both. The biome is 1.17.1's own, for two fields: see
// plainsBiome1_17_1.
func registriesMinecraft1_17_1() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_18},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_17_1},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_17_1.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// plainsBiome1_17_1 is what every chunk a limbo sends is made of, as 1.17.1
// reads it: the 1.18.2 compound that serves 1.18, with the biome's depth
// and scale on top. 1.18 is where the two left the codec, along with the
// terrain shaping they drove, and 1.17.1 reads both as required fields: a
// compound without them is a biome the client refuses. The values are the
// ones the client's own plains carries, read off its jar's report, and
// every other field has the same codec in the two.
var plainsBiome1_17_1 = nbt.Compound{
	"category":      nbt.String("plains"),
	"depth":         nbt.Float(0.125),
	"scale":         nbt.Float(0.05),
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

// tagsMinecraft1_17_1 is every tag a 1.17.1 client's jar declares for a
// registry a vanilla 1.17.1 server sends tags for, generated alongside the
// registries, for the same reason the later versions' are. The sets cover
// the same five registries as 1.18's: the blocks, the items, the entity
// types, the fluids and the game events, the five static tag helpers a
// server of either version serializes. Within the five the sets differ by
// what each jar declares: 1.18's additions, the blocks its mobs spawn on and
// its azaleas grow on among them, the terracotta and dirt tags, and the one
// block tag 1.18 renamed, from the lava pool's stone replaceables to what it
// cannot replace. The five matter as much on 1.17.1 as on 1.18: a 1.17.1
// client checks the payload for every tag its own helpers name and leaves
// over a missing one, which the generation covers by listing every tag the
// jar declares, the helpers' among them. 1.17.1's jar keeps its tag
// directories under the plural names 1.21 dropped, as 1.18's does, which
// only the generation sees. The tags are sent as a play packet right after
// the login, since there is no configuration phase to send them in.
func tagsMinecraft1_17_1() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_17_1.json")

	return tags, err
}
