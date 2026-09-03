package gamedata

import (
	"github.com/Shonz1/go-void-limbo/nbt"
)

// registriesMinecraft1_19_3 is every registry a 1.19.3 client is sent, which
// is all 3 of NETWORKABLE_REGISTRIES from its own sources for that version:
// the dimension types, the biomes and the chat types. 1.19.4 is where the
// damage types and the armor trims appeared, so a 1.19.3 client reads
// nothing of them, and a compound naming a registry it does not know is one
// it refuses. The one generated registry comes from 1.19.3's own jar: its
// chat types are 1.19.4's entry for entry. See data/minecraft_1_19_3.json.
//
// The registries go out the way 1.19.4's do, inside the play login as the
// one compound a client before 1.20.2 reads them from, and the 1.19.4 step's
// login transformer swaps it in for 1.19.4's.
//
// The dimension type below is the same entry 1.19.4 is sent, by reference:
// the codec is field-for-field identical in 1.19.3 and 1.19.4, read off the
// two jars, and the uniform int provider is the plain codec in both, so the
// nested monster spawn light level 1.20.3 needs is what 1.19.3 needs too.
// The biome is 1.19.3's own, since 1.19.4 is where its climate changed
// shape: see plainsBiome1_19_3.
func registriesMinecraft1_19_3() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_20_3},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_19_3},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_19_3.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_19_3 is every tag a 1.19.3 client's jar declares outside
// its feature-flagged packs, generated alongside the registries, for the
// same reason the later versions' are. The sets cover ten registries to
// 1.19.4's eleven -- there are no damage types to tag -- and differ from
// them by what 1.19.4 added: the tool tags and the smelts to glass tag on
// the items, the fall damage immunity and the underwater dismount on the
// entity types, and the biome tags 1.19.4 split the snow and gold rabbits
// one into. 1.19.3's jar keeps its tag directories under the plural names
// 1.21 dropped, as 1.19.4's does, which only the generation sees. The tags
// are sent as a play packet right after the login, since there is no
// configuration phase to send them in.
func tagsMinecraft1_19_3() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_19_3.json")

	return tags, err
}

// plainsBiome1_19_3 is what every chunk a limbo sends is made of, as 1.19.3
// reads it: the 1.21.9 compound that serves 1.19.4 and 1.20, with the
// climate's precipitation as 1.19.3 spells it. 1.19.4 is where the field
// turned from a name -- none, rain or snow -- into the has precipitation
// flag, so a 1.19.3 client reads the flag as a missing field and refuses the
// biome. The value is the name for the flag the later entries clear, and the
// colours are the same the later entries pick, which are the client's own
// plains.
var plainsBiome1_19_3 = nbt.Compound{
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
