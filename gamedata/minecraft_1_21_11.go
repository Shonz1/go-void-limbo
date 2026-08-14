package gamedata

import (
	"maps"

	"github.com/Shonz1/go-void-limbo/nbt"
)

// registriesMinecraft1_21_11 is every registry a 1.21.11 client is sent, which
// is all 23 of SYNCHRONIZED_REGISTRIES from its own sources for that version.
// 26.1 has the same list plus the four animal sound variant registries and
// world_clock.
//
// The two below are a limbo's own decision, as they are for the later
// versions. The other twenty-one come from 1.21.11's own data, because
// versions disagree about the contents of entries they share as well as about
// which entries exist. See data/minecraft_1_21_11.json.
func registriesMinecraft1_21_11() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType1_21_11},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome},
		}},
	}

	generated, _, err := loadDataFile("minecraft_1_21_11.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft1_21_11 is every tag a 1.21.11 client's jar declares, generated
// alongside the registries, for the same reason the later versions' are: the
// client resolves tags while building item components, and a tag it asks for
// that was never declared throws rather than defaulting to empty.
func tagsMinecraft1_21_11() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_1_21_11.json")

	return tags, err
}

// overworldDimensionType1_21_11 is the overworld dimension type as 1.21.11
// reads it.
//
// 1.21.11 already reads the reworked schema the 26.x entries are built on --
// the attributes, skybox and timelines fields all exist, and infiniburn is the
// name of a block tag exactly as 26.1 has it. The one field it does not have
// is has_ender_dragon_fight, which 26.1 requires and 1.21.11's codec has no
// field for, so it is the one field removed here.
var overworldDimensionType1_21_11 = func() nbt.Compound {
	dimensionType := maps.Clone(overworldDimensionType26_1)
	delete(dimensionType, "has_ender_dragon_fight")

	return dimensionType
}()
