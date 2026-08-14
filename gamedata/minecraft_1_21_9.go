package gamedata

import (
	"github.com/Shonz1/go-void-limbo/nbt"
)

// registriesMinecraft1_21_9 is every registry a 1.21.9 client is sent, which
// is all 21 of SYNCHRONIZED_REGISTRIES from its own sources for that version.
// 1.21.11 has the same list plus zombie_nautilus_variant and timeline.
//
// The two below are a limbo's own decision, as they are for the later
// versions. The other nineteen come from 1.21.9's own data, because versions
// disagree about the contents of entries they share as well as about which
// entries exist. See data/minecraft_1_21_9.json.
func registriesMinecraft1_21_9() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_21_9},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome1_21_9},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_21_9.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_21_9 is every tag a 1.21.9 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are: the
// client resolves tags while building item components, and a tag it asks for
// that was never declared throws rather than defaulting to empty.
func tagsMinecraft1_21_9() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_21_9.json")

	return tags, err
}

// overworldDimensionType1_21_9 is the overworld dimension type as 1.21.9
// reads it.
//
// 1.21.9 is the last supported version before the schema rework the later
// entries are built on -- it happened in 1.21.11, not in 26.x -- so nothing
// here is shared with them. These are the fields its own codec requires, in
// the shape the client's own overworld uses: the behaviour booleans the rework
// removed (ultrawarm, natural, bed_works, respawn_anchor_works, piglin_safe,
// has_raids), the effects identifier, and infiniburn as the name of a block
// tag. The codec's two optional fields, fixed_time and cloud_height, are left
// to their defaults.
//
// The vertical bounds decide where the client accepts chunks and draws the
// void fog, so they have to agree with whatever the play phase sends, exactly
// as they do in the later entries.
var overworldDimensionType1_21_9 = nbt.Compound{
	"ambient_light":    nbt.Float(0),
	"bed_works":        nbt.Byte(1),
	"coordinate_scale": nbt.Double(1),
	"effects":          nbt.String("minecraft:overworld"),
	"has_ceiling":      nbt.Byte(0),
	"has_raids":        nbt.Byte(1),
	"has_skylight":     nbt.Byte(1),
	"height":           nbt.Int(384),
	// The name of a block tag, as in 26.1: 1.21.9 parses the name into a
	// reference without resolving it, so a tag with no members is not a tag it
	// complains about. Nothing burns either way.
	"infiniburn":                      nbt.String("#minecraft:infiniburn_overworld"),
	"logical_height":                  nbt.Int(384),
	"min_y":                           nbt.Int(-64),
	"monster_spawn_block_light_limit": nbt.Int(0),
	// An int provider, the shape the client's own overworld uses; nothing here
	// spawns monsters anyway.
	"monster_spawn_light_level": nbt.Compound{
		"type":          nbt.String("minecraft:uniform"),
		"min_inclusive": nbt.Int(0),
		"max_inclusive": nbt.Int(7),
	},
	"natural":              nbt.Byte(1),
	"piglin_safe":          nbt.Byte(0),
	"respawn_anchor_works": nbt.Byte(0),
	"ultrawarm":            nbt.Byte(0),
}

// plainsBiome1_21_9 is what every chunk a limbo sends is made of, in the
// pre-rework schema 1.21.9 reads: the colours sit as integers in effects
// rather than as "#rrggbb" strings split between effects and attributes, and
// four of them plus the music volume are fields the codec requires rather
// than optional. The values are the same colours the later entries pick,
// which are the client's own plains.
var plainsBiome1_21_9 = nbt.Compound{
	"has_precipitation": nbt.Byte(0),
	"temperature":       nbt.Float(0.8),
	"downfall":          nbt.Float(0.4),
	"effects": nbt.Compound{
		"fog_color":       nbt.Int(12638463),
		"sky_color":       nbt.Int(7907327),
		"water_color":     nbt.Int(4159204),
		"water_fog_color": nbt.Int(329011),
		"music_volume":    nbt.Float(1),
	},
}
