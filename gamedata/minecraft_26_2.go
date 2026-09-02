package gamedata

import (
	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/types"
)

// NewDefaultProvider builds the provider for the versions this server speaks.
//
// Adding a version means adding a set here. Start it as a copy of the set below
// and apply the differences: registry content changes unpredictably between
// versions, so a chain of shared bases becomes harder to read than the
// duplication it saves.
//
// Each set is encoded as soon as it is loaded, before the next is read. A
// version's registries parse into a tree several times the size of the bytes
// they encode to, and thirteen of those held at once until the last is read
// is what this process would peak at for no reason; one at a time, the peak
// is one tree.
func NewDefaultProvider() (*Provider, error) {
	sets := []struct {
		minProtocol types.ProtocolId
		registries  func() ([]Registry, error)
		tags        func() ([]TagSet, error)
	}{
		{types.ProtocolVersions.MINECRAFT_1_20_2.ID, registriesMinecraft1_20_2, tagsMinecraft1_20_2},
		{types.ProtocolVersions.MINECRAFT_1_20_3.ID, registriesMinecraft1_20_3, tagsMinecraft1_20_3},
		{types.ProtocolVersions.MINECRAFT_1_20_5.ID, registriesMinecraft1_20_5, tagsMinecraft1_20_5},
		{types.ProtocolVersions.MINECRAFT_1_21.ID, registriesMinecraft1_21, tagsMinecraft1_21},
		{types.ProtocolVersions.MINECRAFT_1_21_2.ID, registriesMinecraft1_21_2, tagsMinecraft1_21_2},
		{types.ProtocolVersions.MINECRAFT_1_21_4.ID, registriesMinecraft1_21_4, tagsMinecraft1_21_4},
		{types.ProtocolVersions.MINECRAFT_1_21_5.ID, registriesMinecraft1_21_5, tagsMinecraft1_21_5},
		{types.ProtocolVersions.MINECRAFT_1_21_6.ID, registriesMinecraft1_21_6, tagsMinecraft1_21_6},
		{types.ProtocolVersions.MINECRAFT_1_21_7.ID, registriesMinecraft1_21_7, tagsMinecraft1_21_7},
		{types.ProtocolVersions.MINECRAFT_1_21_9.ID, registriesMinecraft1_21_9, tagsMinecraft1_21_9},
		{types.ProtocolVersions.MINECRAFT_1_21_11.ID, registriesMinecraft1_21_11, tagsMinecraft1_21_11},
		{types.ProtocolVersions.MINECRAFT_26_1.ID, registriesMinecraft26_1, tagsMinecraft26_1},
		{types.ProtocolVersions.MINECRAFT_26_2.ID, registriesMinecraft26_2, tagsMinecraft26_2},
	}

	buckets := make([]bucket, 0, len(sets))

	for _, set := range sets {
		registries, err := set.registries()
		if err != nil {
			return nil, err
		}

		tags, err := set.tags()
		if err != nil {
			return nil, err
		}

		encoded, err := encodeSet(Set{MinProtocol: set.minProtocol, Registries: registries, Tags: tags})
		if err != nil {
			return nil, err
		}

		buckets = append(buckets, encoded)
	}

	return newProvider(buckets)
}

// registriesMinecraft26_2 is every registry the client is sent, which is all 29
// of SYNCHRONIZED_REGISTRIES from its own sources for this version. That list is
// not one to guess at: a registry the client expects and never receives
// disconnects it. Recheck it against the jar whenever a version is added.
//
// Only the two below are a limbo's own decision. The client builds a registry
// from exactly what arrives, so these are not a subset it fills in afterwards --
// one entry means a registry of one, and everything later refers to it by id 0.
//
// The other twenty-seven are generated from the client's own data, because it
// will not have them any other way. See data/minecraft_26_2.json.
func registriesMinecraft26_2() ([]Registry, error) {
	registries := []Registry{
		{Name: "minecraft:dimension_type", Entries: []Entry{
			{Name: "minecraft:overworld", Data: overworldDimensionType},
		}},
		{Name: "minecraft:worldgen/biome", Entries: []Entry{
			{Name: "minecraft:plains", Data: plainsBiome},
		}},
	}

	generated, _, err := loadDataFile("minecraft_26_2.json")
	if err != nil {
		return nil, err
	}

	return append(registries, generated...), nil
}

// tagsMinecraft26_2 is every tag the 26.2 client's jar declares, all of them
// empty. The client resolves tags it needs while building item components at
// the end of the configuration phase, and a tag it asks for that was never
// declared throws rather than defaulting to empty, so declaring all of them is
// what makes that lookup total.
func tagsMinecraft26_2() ([]TagSet, error) {
	_, tags, err := loadDataFile("minecraft_26_2.json")

	return tags, err
}

// overworldDimensionType is the dimension the play login packet points at.
//
// These are exactly the fields DimensionType's codec requires, and nothing else.
// Reading them off the codec beats copying the client's own overworld, which
// carries six optional fields on top: has_fixed_time, skybox, cardinal_light,
// attributes, timelines and default_clock. Each is a value a limbo does not need
// and can only get wrong, and two of them reach into registries that would then
// have to be synced as well, so all six are left to their defaults.
//
// The schema moved a long way from 1.21. The booleans that described how a
// dimension behaves (natural, ultrawarm, piglin_safe, has_raids, bed_works,
// respawn_anchor_works) and the effects identifier are gone; what survives of
// the monster settings is inlined here rather than nested.
//
// The vertical bounds decide where the client accepts chunks and draws the void
// fog, so they have to agree with whatever the play phase sends. The codec
// checks them: height at least 16 and a multiple of 16, min_y a multiple of 16,
// and logical_height no higher than height.
var overworldDimensionType = nbt.Compound{
	"ambient_light":          nbt.Float(0),
	"coordinate_scale":       nbt.Double(1),
	"has_ceiling":            nbt.Byte(0),
	"has_ender_dragon_fight": nbt.Byte(0),
	"has_skylight":           nbt.Byte(1),
	"height":                 nbt.Int(384),
	// The set of blocks that burn forever, empty because nothing here burns.
	//
	// This is a set of blocks, where 1.21 had the name of a block tag. A set
	// still reads the "#minecraft:infiniburn_overworld" the client's own
	// overworld uses, but only as one of two accepted forms, and that form
	// needs the tag bound by an Update Tags packet a limbo does not send. The
	// other form is a plain list of blocks, which when empty resolves without
	// consulting the block registry at all.
	"infiniburn":                      nbt.List{},
	"logical_height":                  nbt.Int(384),
	"min_y":                           nbt.Int(-64),
	"monster_spawn_block_light_limit": nbt.Int(0),
	// An int provider. This is the shape the client's own overworld uses, and
	// nothing here spawns monsters anyway.
	"monster_spawn_light_level": nbt.Compound{
		"type":          nbt.String("minecraft:uniform"),
		"min_inclusive": nbt.Int(0),
		"max_inclusive": nbt.Int(7),
	},
}

// plainsBiome is what every chunk a limbo sends is made of. The colours are the
// only part a player can see, since nothing here generates terrain or weather.
//
// The generation fields the client ships with (features, carvers, spawners) are
// absent because the client is not sent them: biomes cross the wire through a
// reduced codec carrying only what rendering needs.
var plainsBiome = nbt.Compound{
	"has_precipitation": nbt.Byte(0),
	"temperature":       nbt.Float(0.8),
	"downfall":          nbt.Float(0.4),
	"effects": nbt.Compound{
		"water_color": nbt.String("#3f76e4"),
	},
	"attributes": nbt.Compound{
		"minecraft:visual/sky_color": nbt.String("#78a7ff"),
	},
}
