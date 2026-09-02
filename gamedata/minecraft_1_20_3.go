package gamedata

import "github.com/Shonz1/go-void-limbo/nbt"

// registriesMinecraft1_20_3 is every registry a 1.20.3 client is sent, which
// is all 6 of NETWORKABLE_REGISTRIES from its own sources for that version:
// 1.20.5's eight minus the two 1.20.5 is where they became data-driven -- the
// wolf variant and the banner pattern. The four generated registries come
// from 1.20.4's own jar, and what 1.20.5 has that 1.20.3 does not is what
// 1.20.5 added: the spit damage type, and the armor materials the four metal
// trim materials override written under their namespaced names. The trim
// patterns and the chat types are the same in both. See
// data/minecraft_1_20_3.json.
//
// The shape the registries go out in is the older one: a client before 1.20.5
// takes every registry in one packet, as one compound, and the provider
// encodes a set that starts below 1.20.5 that way. Package gamedata is where
// that difference lives because it is not a rewrite of a packet but a
// different packet altogether, one that holds what would be several.
//
// The two below are a limbo's own decision, as they are for the later
// versions. The biome is 1.21.9's entry by reference: its codec is
// field-for-field identical in 1.20.4 and 1.20.6, read off the two jars, so
// what serves 1.20.5 serves 1.20.3. The dimension type is this version's own,
// for one field: see overworldDimensionType1_20_3.
func registriesMinecraft1_20_3() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_20_3},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_20_3.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// overworldDimensionType1_20_3 is the overworld dimension type as 1.20.3
// reads it: 1.21.9's entry, field for field, but for the monster spawn light
// level. The field is an int provider, which every version reads through a
// dispatch on its type -- and 1.20.5 is where the uniform provider's codec
// became a map codec, which the dispatch lays out flat beside the type.
// Before that it was a plain codec, which the dispatch nests under value, and
// 1.20.4's client has no other way to read it: the flat shape fails to parse
// and takes the whole registry with it, which the client answers by leaving
// during configuration. Every other field has the same codec in the two, read
// off the jars.
var overworldDimensionType1_20_3 = nbt.Compound{
	"ambient_light":                   nbt.Float(0),
	"bed_works":                       nbt.Byte(1),
	"coordinate_scale":                nbt.Double(1),
	"effects":                         nbt.String("minecraft:overworld"),
	"has_ceiling":                     nbt.Byte(0),
	"has_raids":                       nbt.Byte(1),
	"has_skylight":                    nbt.Byte(1),
	"height":                          nbt.Int(384),
	"infiniburn":                      nbt.String("#minecraft:infiniburn_overworld"),
	"logical_height":                  nbt.Int(384),
	"min_y":                           nbt.Int(-64),
	"monster_spawn_block_light_limit": nbt.Int(0),
	"monster_spawn_light_level": nbt.Compound{
		"type": nbt.String("minecraft:uniform"),
		"value": nbt.Compound{
			"min_inclusive": nbt.Int(0),
			"max_inclusive": nbt.Int(7),
		},
	},
	"natural":              nbt.Byte(1),
	"piglin_safe":          nbt.Byte(0),
	"respawn_anchor_works": nbt.Byte(0),
	"ultrawarm":            nbt.Byte(0),
}

// tagsMinecraft1_20_3 is every tag a 1.20.3 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are. The
// sets cover eleven registries to 1.20.5's twelve: 1.20.5 is where the
// enchantment tags appeared. 1.20.4's jar keeps its tag directories under the
// plural names 1.21 dropped, as 1.20.6's does, which only the generation sees.
func tagsMinecraft1_20_3() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_20_3.json")

	return tags, err
}
