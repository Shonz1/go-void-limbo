package gamedata

import "go-void-limbo/nbt"

// Code generated from the 26.2 client jar. DO NOT EDIT BY HAND.
//
// Every synced registry except the two a limbo actually decides for itself,
// dimension_type and worldgen/biome, which are written by hand next door.
//
// Sending these at all is not optional, and neither is filling them. The client
// refuses some registries outright when empty, and reaches into others by name
// while building item components at the end of the configuration phase, where a
// key that resolves to nothing throws: a goat horn names an instrument, a spawn
// egg a mob variant, a spear its damage type. Which names it will ask for is not
// knowable from the outside, so the answer to both demands is the same, and it
// is the reason this file is generated rather than written: carry the client's
// own entries verbatim and there is nothing left for it to miss.
//
// Two kinds of value are rewritten on the way in, because both would otherwise
// be resolved against things that are not there while registries load:
//
//   - Tag references become empty sets. A tag is bound by Update Tags, which
//     arrives after this, so naming one here fails the whole entry. This is the
//     same reason infiniburn is sent as a plain empty list.
//   - spawn_conditions is dropped from the mob variants. It is optional and its
//     biome references are tags.
//
// Nothing in a limbo reads any of this, so what the rewrites cost is behaviour
// nothing here has: an enchantment that applies to no items, a variant that
// spawns nowhere.
func generatedRegistriesMinecraft26_2() []Registry {
	return []Registry{
		{Name: "minecraft:banner_pattern", Entries: []Entry{
			{Name: "minecraft:base", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:base"),
				"translation_key": nbt.String("block.minecraft.banner.base"),
			}},
			{Name: "minecraft:border", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:border"),
				"translation_key": nbt.String("block.minecraft.banner.border"),
			}},
			{Name: "minecraft:bricks", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:bricks"),
				"translation_key": nbt.String("block.minecraft.banner.bricks"),
			}},
			{Name: "minecraft:circle", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:circle"),
				"translation_key": nbt.String("block.minecraft.banner.circle"),
			}},
			{Name: "minecraft:creeper", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:creeper"),
				"translation_key": nbt.String("block.minecraft.banner.creeper"),
			}},
			{Name: "minecraft:cross", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:cross"),
				"translation_key": nbt.String("block.minecraft.banner.cross"),
			}},
			{Name: "minecraft:curly_border", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:curly_border"),
				"translation_key": nbt.String("block.minecraft.banner.curly_border"),
			}},
			{Name: "minecraft:diagonal_left", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:diagonal_left"),
				"translation_key": nbt.String("block.minecraft.banner.diagonal_left"),
			}},
			{Name: "minecraft:diagonal_right", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:diagonal_right"),
				"translation_key": nbt.String("block.minecraft.banner.diagonal_right"),
			}},
			{Name: "minecraft:diagonal_up_left", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:diagonal_up_left"),
				"translation_key": nbt.String("block.minecraft.banner.diagonal_up_left"),
			}},
			{Name: "minecraft:diagonal_up_right", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:diagonal_up_right"),
				"translation_key": nbt.String("block.minecraft.banner.diagonal_up_right"),
			}},
			{Name: "minecraft:flow", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:flow"),
				"translation_key": nbt.String("block.minecraft.banner.flow"),
			}},
			{Name: "minecraft:flower", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:flower"),
				"translation_key": nbt.String("block.minecraft.banner.flower"),
			}},
			{Name: "minecraft:globe", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:globe"),
				"translation_key": nbt.String("block.minecraft.banner.globe"),
			}},
			{Name: "minecraft:gradient", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:gradient"),
				"translation_key": nbt.String("block.minecraft.banner.gradient"),
			}},
			{Name: "minecraft:gradient_up", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:gradient_up"),
				"translation_key": nbt.String("block.minecraft.banner.gradient_up"),
			}},
			{Name: "minecraft:guster", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:guster"),
				"translation_key": nbt.String("block.minecraft.banner.guster"),
			}},
			{Name: "minecraft:half_horizontal", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:half_horizontal"),
				"translation_key": nbt.String("block.minecraft.banner.half_horizontal"),
			}},
			{Name: "minecraft:half_horizontal_bottom", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:half_horizontal_bottom"),
				"translation_key": nbt.String("block.minecraft.banner.half_horizontal_bottom"),
			}},
			{Name: "minecraft:half_vertical", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:half_vertical"),
				"translation_key": nbt.String("block.minecraft.banner.half_vertical"),
			}},
			{Name: "minecraft:half_vertical_right", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:half_vertical_right"),
				"translation_key": nbt.String("block.minecraft.banner.half_vertical_right"),
			}},
			{Name: "minecraft:mojang", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:mojang"),
				"translation_key": nbt.String("block.minecraft.banner.mojang"),
			}},
			{Name: "minecraft:piglin", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:piglin"),
				"translation_key": nbt.String("block.minecraft.banner.piglin"),
			}},
			{Name: "minecraft:rhombus", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:rhombus"),
				"translation_key": nbt.String("block.minecraft.banner.rhombus"),
			}},
			{Name: "minecraft:skull", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:skull"),
				"translation_key": nbt.String("block.minecraft.banner.skull"),
			}},
			{Name: "minecraft:small_stripes", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:small_stripes"),
				"translation_key": nbt.String("block.minecraft.banner.small_stripes"),
			}},
			{Name: "minecraft:square_bottom_left", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:square_bottom_left"),
				"translation_key": nbt.String("block.minecraft.banner.square_bottom_left"),
			}},
			{Name: "minecraft:square_bottom_right", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:square_bottom_right"),
				"translation_key": nbt.String("block.minecraft.banner.square_bottom_right"),
			}},
			{Name: "minecraft:square_top_left", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:square_top_left"),
				"translation_key": nbt.String("block.minecraft.banner.square_top_left"),
			}},
			{Name: "minecraft:square_top_right", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:square_top_right"),
				"translation_key": nbt.String("block.minecraft.banner.square_top_right"),
			}},
			{Name: "minecraft:straight_cross", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:straight_cross"),
				"translation_key": nbt.String("block.minecraft.banner.straight_cross"),
			}},
			{Name: "minecraft:stripe_bottom", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:stripe_bottom"),
				"translation_key": nbt.String("block.minecraft.banner.stripe_bottom"),
			}},
			{Name: "minecraft:stripe_center", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:stripe_center"),
				"translation_key": nbt.String("block.minecraft.banner.stripe_center"),
			}},
			{Name: "minecraft:stripe_downleft", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:stripe_downleft"),
				"translation_key": nbt.String("block.minecraft.banner.stripe_downleft"),
			}},
			{Name: "minecraft:stripe_downright", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:stripe_downright"),
				"translation_key": nbt.String("block.minecraft.banner.stripe_downright"),
			}},
			{Name: "minecraft:stripe_left", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:stripe_left"),
				"translation_key": nbt.String("block.minecraft.banner.stripe_left"),
			}},
			{Name: "minecraft:stripe_middle", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:stripe_middle"),
				"translation_key": nbt.String("block.minecraft.banner.stripe_middle"),
			}},
			{Name: "minecraft:stripe_right", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:stripe_right"),
				"translation_key": nbt.String("block.minecraft.banner.stripe_right"),
			}},
			{Name: "minecraft:stripe_top", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:stripe_top"),
				"translation_key": nbt.String("block.minecraft.banner.stripe_top"),
			}},
			{Name: "minecraft:triangle_bottom", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:triangle_bottom"),
				"translation_key": nbt.String("block.minecraft.banner.triangle_bottom"),
			}},
			{Name: "minecraft:triangle_top", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:triangle_top"),
				"translation_key": nbt.String("block.minecraft.banner.triangle_top"),
			}},
			{Name: "minecraft:triangles_bottom", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:triangles_bottom"),
				"translation_key": nbt.String("block.minecraft.banner.triangles_bottom"),
			}},
			{Name: "minecraft:triangles_top", Data: nbt.Compound{
				"asset_id":        nbt.String("minecraft:triangles_top"),
				"translation_key": nbt.String("block.minecraft.banner.triangles_top"),
			}},
		}},
		{Name: "minecraft:instrument", Entries: []Entry{
			{Name: "minecraft:admire_goat_horn", Data: nbt.Compound{
				"description": nbt.Compound{
					"translate": nbt.String("instrument.minecraft.admire_goat_horn"),
				},
				"range":        nbt.Double(256.0),
				"sound_event":  nbt.String("minecraft:item.goat_horn.sound.4"),
				"use_duration": nbt.Double(7.0),
			}},
			{Name: "minecraft:call_goat_horn", Data: nbt.Compound{
				"description": nbt.Compound{
					"translate": nbt.String("instrument.minecraft.call_goat_horn"),
				},
				"range":        nbt.Double(256.0),
				"sound_event":  nbt.String("minecraft:item.goat_horn.sound.5"),
				"use_duration": nbt.Double(7.0),
			}},
			{Name: "minecraft:dream_goat_horn", Data: nbt.Compound{
				"description": nbt.Compound{
					"translate": nbt.String("instrument.minecraft.dream_goat_horn"),
				},
				"range":        nbt.Double(256.0),
				"sound_event":  nbt.String("minecraft:item.goat_horn.sound.7"),
				"use_duration": nbt.Double(7.0),
			}},
			{Name: "minecraft:feel_goat_horn", Data: nbt.Compound{
				"description": nbt.Compound{
					"translate": nbt.String("instrument.minecraft.feel_goat_horn"),
				},
				"range":        nbt.Double(256.0),
				"sound_event":  nbt.String("minecraft:item.goat_horn.sound.3"),
				"use_duration": nbt.Double(7.0),
			}},
			{Name: "minecraft:ponder_goat_horn", Data: nbt.Compound{
				"description": nbt.Compound{
					"translate": nbt.String("instrument.minecraft.ponder_goat_horn"),
				},
				"range":        nbt.Double(256.0),
				"sound_event":  nbt.String("minecraft:item.goat_horn.sound.0"),
				"use_duration": nbt.Double(7.0),
			}},
			{Name: "minecraft:seek_goat_horn", Data: nbt.Compound{
				"description": nbt.Compound{
					"translate": nbt.String("instrument.minecraft.seek_goat_horn"),
				},
				"range":        nbt.Double(256.0),
				"sound_event":  nbt.String("minecraft:item.goat_horn.sound.2"),
				"use_duration": nbt.Double(7.0),
			}},
			{Name: "minecraft:sing_goat_horn", Data: nbt.Compound{
				"description": nbt.Compound{
					"translate": nbt.String("instrument.minecraft.sing_goat_horn"),
				},
				"range":        nbt.Double(256.0),
				"sound_event":  nbt.String("minecraft:item.goat_horn.sound.1"),
				"use_duration": nbt.Double(7.0),
			}},
			{Name: "minecraft:yearn_goat_horn", Data: nbt.Compound{
				"description": nbt.Compound{
					"translate": nbt.String("instrument.minecraft.yearn_goat_horn"),
				},
				"range":        nbt.Double(256.0),
				"sound_event":  nbt.String("minecraft:item.goat_horn.sound.6"),
				"use_duration": nbt.Double(7.0),
			}},
		}},
		{Name: "minecraft:jukebox_song", Entries: []Entry{
			{Name: "minecraft:11", Data: nbt.Compound{
				"comparator_output": nbt.Int(11),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.11"),
				},
				"length_in_seconds": nbt.Double(71.0),
				"sound_event":       nbt.String("minecraft:music_disc.11"),
			}},
			{Name: "minecraft:13", Data: nbt.Compound{
				"comparator_output": nbt.Int(1),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.13"),
				},
				"length_in_seconds": nbt.Double(178.0),
				"sound_event":       nbt.String("minecraft:music_disc.13"),
			}},
			{Name: "minecraft:5", Data: nbt.Compound{
				"comparator_output": nbt.Int(15),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.5"),
				},
				"length_in_seconds": nbt.Double(178.0),
				"sound_event":       nbt.String("minecraft:music_disc.5"),
			}},
			{Name: "minecraft:blocks", Data: nbt.Compound{
				"comparator_output": nbt.Int(3),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.blocks"),
				},
				"length_in_seconds": nbt.Double(345.0),
				"sound_event":       nbt.String("minecraft:music_disc.blocks"),
			}},
			{Name: "minecraft:bounce", Data: nbt.Compound{
				"comparator_output": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.bounce"),
				},
				"length_in_seconds": nbt.Double(234.0),
				"sound_event":       nbt.String("minecraft:music_disc.bounce"),
			}},
			{Name: "minecraft:cat", Data: nbt.Compound{
				"comparator_output": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.cat"),
				},
				"length_in_seconds": nbt.Double(185.0),
				"sound_event":       nbt.String("minecraft:music_disc.cat"),
			}},
			{Name: "minecraft:chirp", Data: nbt.Compound{
				"comparator_output": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.chirp"),
				},
				"length_in_seconds": nbt.Double(185.0),
				"sound_event":       nbt.String("minecraft:music_disc.chirp"),
			}},
			{Name: "minecraft:creator", Data: nbt.Compound{
				"comparator_output": nbt.Int(12),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.creator"),
				},
				"length_in_seconds": nbt.Double(176.0),
				"sound_event":       nbt.String("minecraft:music_disc.creator"),
			}},
			{Name: "minecraft:creator_music_box", Data: nbt.Compound{
				"comparator_output": nbt.Int(11),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.creator_music_box"),
				},
				"length_in_seconds": nbt.Double(73.0),
				"sound_event":       nbt.String("minecraft:music_disc.creator_music_box"),
			}},
			{Name: "minecraft:far", Data: nbt.Compound{
				"comparator_output": nbt.Int(5),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.far"),
				},
				"length_in_seconds": nbt.Double(174.0),
				"sound_event":       nbt.String("minecraft:music_disc.far"),
			}},
			{Name: "minecraft:lava_chicken", Data: nbt.Compound{
				"comparator_output": nbt.Int(9),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.lava_chicken"),
				},
				"length_in_seconds": nbt.Double(134.0),
				"sound_event":       nbt.String("minecraft:music_disc.lava_chicken"),
			}},
			{Name: "minecraft:mall", Data: nbt.Compound{
				"comparator_output": nbt.Int(6),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.mall"),
				},
				"length_in_seconds": nbt.Double(197.0),
				"sound_event":       nbt.String("minecraft:music_disc.mall"),
			}},
			{Name: "minecraft:mellohi", Data: nbt.Compound{
				"comparator_output": nbt.Int(7),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.mellohi"),
				},
				"length_in_seconds": nbt.Double(96.0),
				"sound_event":       nbt.String("minecraft:music_disc.mellohi"),
			}},
			{Name: "minecraft:otherside", Data: nbt.Compound{
				"comparator_output": nbt.Int(14),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.otherside"),
				},
				"length_in_seconds": nbt.Double(195.0),
				"sound_event":       nbt.String("minecraft:music_disc.otherside"),
			}},
			{Name: "minecraft:pigstep", Data: nbt.Compound{
				"comparator_output": nbt.Int(13),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.pigstep"),
				},
				"length_in_seconds": nbt.Double(149.0),
				"sound_event":       nbt.String("minecraft:music_disc.pigstep"),
			}},
			{Name: "minecraft:precipice", Data: nbt.Compound{
				"comparator_output": nbt.Int(13),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.precipice"),
				},
				"length_in_seconds": nbt.Double(299.0),
				"sound_event":       nbt.String("minecraft:music_disc.precipice"),
			}},
			{Name: "minecraft:relic", Data: nbt.Compound{
				"comparator_output": nbt.Int(14),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.relic"),
				},
				"length_in_seconds": nbt.Double(218.0),
				"sound_event":       nbt.String("minecraft:music_disc.relic"),
			}},
			{Name: "minecraft:stal", Data: nbt.Compound{
				"comparator_output": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.stal"),
				},
				"length_in_seconds": nbt.Double(150.0),
				"sound_event":       nbt.String("minecraft:music_disc.stal"),
			}},
			{Name: "minecraft:strad", Data: nbt.Compound{
				"comparator_output": nbt.Int(9),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.strad"),
				},
				"length_in_seconds": nbt.Double(188.0),
				"sound_event":       nbt.String("minecraft:music_disc.strad"),
			}},
			{Name: "minecraft:tears", Data: nbt.Compound{
				"comparator_output": nbt.Int(10),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.tears"),
				},
				"length_in_seconds": nbt.Double(175.0),
				"sound_event":       nbt.String("minecraft:music_disc.tears"),
			}},
			{Name: "minecraft:wait", Data: nbt.Compound{
				"comparator_output": nbt.Int(12),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.wait"),
				},
				"length_in_seconds": nbt.Double(238.0),
				"sound_event":       nbt.String("minecraft:music_disc.wait"),
			}},
			{Name: "minecraft:ward", Data: nbt.Compound{
				"comparator_output": nbt.Int(10),
				"description": nbt.Compound{
					"translate": nbt.String("jukebox_song.minecraft.ward"),
				},
				"length_in_seconds": nbt.Double(251.0),
				"sound_event":       nbt.String("minecraft:music_disc.ward"),
			}},
		}},
		{Name: "minecraft:trim_material", Entries: []Entry{
			{Name: "minecraft:amethyst", Data: nbt.Compound{
				"asset_name": nbt.String("amethyst"),
				"description": nbt.Compound{
					"color":     nbt.String("#9A5CC6"),
					"translate": nbt.String("trim_material.minecraft.amethyst"),
				},
			}},
			{Name: "minecraft:copper", Data: nbt.Compound{
				"asset_name": nbt.String("copper"),
				"description": nbt.Compound{
					"color":     nbt.String("#B4684D"),
					"translate": nbt.String("trim_material.minecraft.copper"),
				},
				"override_armor_assets": nbt.Compound{
					"minecraft:copper": nbt.String("copper_darker"),
				},
			}},
			{Name: "minecraft:diamond", Data: nbt.Compound{
				"asset_name": nbt.String("diamond"),
				"description": nbt.Compound{
					"color":     nbt.String("#6EECD2"),
					"translate": nbt.String("trim_material.minecraft.diamond"),
				},
				"override_armor_assets": nbt.Compound{
					"minecraft:diamond": nbt.String("diamond_darker"),
				},
			}},
			{Name: "minecraft:emerald", Data: nbt.Compound{
				"asset_name": nbt.String("emerald"),
				"description": nbt.Compound{
					"color":     nbt.String("#11A036"),
					"translate": nbt.String("trim_material.minecraft.emerald"),
				},
			}},
			{Name: "minecraft:gold", Data: nbt.Compound{
				"asset_name": nbt.String("gold"),
				"description": nbt.Compound{
					"color":     nbt.String("#DEB12D"),
					"translate": nbt.String("trim_material.minecraft.gold"),
				},
				"override_armor_assets": nbt.Compound{
					"minecraft:gold": nbt.String("gold_darker"),
				},
			}},
			{Name: "minecraft:iron", Data: nbt.Compound{
				"asset_name": nbt.String("iron"),
				"description": nbt.Compound{
					"color":     nbt.String("#ECECEC"),
					"translate": nbt.String("trim_material.minecraft.iron"),
				},
				"override_armor_assets": nbt.Compound{
					"minecraft:iron": nbt.String("iron_darker"),
				},
			}},
			{Name: "minecraft:lapis", Data: nbt.Compound{
				"asset_name": nbt.String("lapis"),
				"description": nbt.Compound{
					"color":     nbt.String("#416E97"),
					"translate": nbt.String("trim_material.minecraft.lapis"),
				},
			}},
			{Name: "minecraft:netherite", Data: nbt.Compound{
				"asset_name": nbt.String("netherite"),
				"description": nbt.Compound{
					"color":     nbt.String("#625859"),
					"translate": nbt.String("trim_material.minecraft.netherite"),
				},
				"override_armor_assets": nbt.Compound{
					"minecraft:netherite": nbt.String("netherite_darker"),
				},
			}},
			{Name: "minecraft:quartz", Data: nbt.Compound{
				"asset_name": nbt.String("quartz"),
				"description": nbt.Compound{
					"color":     nbt.String("#E3D4C4"),
					"translate": nbt.String("trim_material.minecraft.quartz"),
				},
			}},
			{Name: "minecraft:redstone", Data: nbt.Compound{
				"asset_name": nbt.String("redstone"),
				"description": nbt.Compound{
					"color":     nbt.String("#971607"),
					"translate": nbt.String("trim_material.minecraft.redstone"),
				},
			}},
			{Name: "minecraft:resin", Data: nbt.Compound{
				"asset_name": nbt.String("resin"),
				"description": nbt.Compound{
					"color":     nbt.String("#FC7812"),
					"translate": nbt.String("trim_material.minecraft.resin"),
				},
			}},
		}},
		{Name: "minecraft:trim_pattern", Entries: []Entry{
			{Name: "minecraft:bolt", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:bolt"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.bolt"),
				},
			}},
			{Name: "minecraft:coast", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:coast"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.coast"),
				},
			}},
			{Name: "minecraft:dune", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:dune"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.dune"),
				},
			}},
			{Name: "minecraft:eye", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:eye"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.eye"),
				},
			}},
			{Name: "minecraft:flow", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:flow"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.flow"),
				},
			}},
			{Name: "minecraft:host", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:host"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.host"),
				},
			}},
			{Name: "minecraft:raiser", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:raiser"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.raiser"),
				},
			}},
			{Name: "minecraft:rib", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:rib"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.rib"),
				},
			}},
			{Name: "minecraft:sentry", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:sentry"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.sentry"),
				},
			}},
			{Name: "minecraft:shaper", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:shaper"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.shaper"),
				},
			}},
			{Name: "minecraft:silence", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:silence"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.silence"),
				},
			}},
			{Name: "minecraft:snout", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:snout"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.snout"),
				},
			}},
			{Name: "minecraft:spire", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:spire"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.spire"),
				},
			}},
			{Name: "minecraft:tide", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:tide"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.tide"),
				},
			}},
			{Name: "minecraft:vex", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:vex"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.vex"),
				},
			}},
			{Name: "minecraft:ward", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:ward"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.ward"),
				},
			}},
			{Name: "minecraft:wayfinder", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:wayfinder"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.wayfinder"),
				},
			}},
			{Name: "minecraft:wild", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:wild"),
				"decal":    nbt.Byte(0),
				"description": nbt.Compound{
					"translate": nbt.String("trim_pattern.minecraft.wild"),
				},
			}},
		}},
		{Name: "minecraft:cat_variant", Entries: []Entry{
			{Name: "minecraft:all_black", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_all_black"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_all_black_baby"),
			}},
			{Name: "minecraft:black", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_black"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_black_baby"),
			}},
			{Name: "minecraft:british_shorthair", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_british_shorthair"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_british_shorthair_baby"),
			}},
			{Name: "minecraft:calico", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_calico"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_calico_baby"),
			}},
			{Name: "minecraft:jellie", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_jellie"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_jellie_baby"),
			}},
			{Name: "minecraft:persian", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_persian"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_persian_baby"),
			}},
			{Name: "minecraft:ragdoll", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_ragdoll"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_ragdoll_baby"),
			}},
			{Name: "minecraft:red", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_red"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_red_baby"),
			}},
			{Name: "minecraft:siamese", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_siamese"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_siamese_baby"),
			}},
			{Name: "minecraft:tabby", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_tabby"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_tabby_baby"),
			}},
			{Name: "minecraft:white", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cat/cat_white"),
				"baby_asset_id": nbt.String("minecraft:entity/cat/cat_white_baby"),
			}},
		}},
		{Name: "minecraft:cat_sound_variant", Entries: []Entry{
			{Name: "minecraft:classic", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound":       nbt.String("minecraft:entity.cat.ambient"),
					"beg_for_food_sound":  nbt.String("minecraft:entity.cat.beg_for_food"),
					"death_sound":         nbt.String("minecraft:entity.cat.death"),
					"eat_sound":           nbt.String("minecraft:entity.cat.eat"),
					"hiss_sound":          nbt.String("minecraft:entity.cat.hiss"),
					"hurt_sound":          nbt.String("minecraft:entity.cat.hurt"),
					"purr_sound":          nbt.String("minecraft:entity.cat.purr"),
					"purreow_sound":       nbt.String("minecraft:entity.cat.purreow"),
					"stray_ambient_sound": nbt.String("minecraft:entity.cat.stray_ambient"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound":       nbt.String("minecraft:entity.baby_cat.ambient"),
					"beg_for_food_sound":  nbt.String("minecraft:entity.baby_cat.beg_for_food"),
					"death_sound":         nbt.String("minecraft:entity.baby_cat.death"),
					"eat_sound":           nbt.String("minecraft:entity.baby_cat.eat"),
					"hiss_sound":          nbt.String("minecraft:entity.baby_cat.hiss"),
					"hurt_sound":          nbt.String("minecraft:entity.baby_cat.hurt"),
					"purr_sound":          nbt.String("minecraft:entity.baby_cat.purr"),
					"purreow_sound":       nbt.String("minecraft:entity.baby_cat.purreow"),
					"stray_ambient_sound": nbt.String("minecraft:entity.baby_cat.stray_ambient"),
				},
			}},
			{Name: "minecraft:royal", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound":       nbt.String("minecraft:entity.cat_royal.ambient"),
					"beg_for_food_sound":  nbt.String("minecraft:entity.cat_royal.beg_for_food"),
					"death_sound":         nbt.String("minecraft:entity.cat_royal.death"),
					"eat_sound":           nbt.String("minecraft:entity.cat_royal.eat"),
					"hiss_sound":          nbt.String("minecraft:entity.cat_royal.hiss"),
					"hurt_sound":          nbt.String("minecraft:entity.cat_royal.hurt"),
					"purr_sound":          nbt.String("minecraft:entity.cat_royal.purr"),
					"purreow_sound":       nbt.String("minecraft:entity.cat_royal.purreow"),
					"stray_ambient_sound": nbt.String("minecraft:entity.cat_royal.stray_ambient"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound":       nbt.String("minecraft:entity.baby_cat.ambient"),
					"beg_for_food_sound":  nbt.String("minecraft:entity.baby_cat.beg_for_food"),
					"death_sound":         nbt.String("minecraft:entity.baby_cat.death"),
					"eat_sound":           nbt.String("minecraft:entity.baby_cat.eat"),
					"hiss_sound":          nbt.String("minecraft:entity.baby_cat.hiss"),
					"hurt_sound":          nbt.String("minecraft:entity.baby_cat.hurt"),
					"purr_sound":          nbt.String("minecraft:entity.baby_cat.purr"),
					"purreow_sound":       nbt.String("minecraft:entity.baby_cat.purreow"),
					"stray_ambient_sound": nbt.String("minecraft:entity.baby_cat.stray_ambient"),
				},
			}},
		}},
		{Name: "minecraft:chicken_variant", Entries: []Entry{
			{Name: "minecraft:cold", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/chicken/chicken_cold"),
				"baby_asset_id": nbt.String("minecraft:entity/chicken/chicken_cold_baby"),
				"model":         nbt.String("cold"),
			}},
			{Name: "minecraft:temperate", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/chicken/chicken_temperate"),
				"baby_asset_id": nbt.String("minecraft:entity/chicken/chicken_temperate_baby"),
			}},
			{Name: "minecraft:warm", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/chicken/chicken_warm"),
				"baby_asset_id": nbt.String("minecraft:entity/chicken/chicken_warm_baby"),
			}},
		}},
		{Name: "minecraft:chicken_sound_variant", Entries: []Entry{
			{Name: "minecraft:classic", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.chicken.ambient"),
					"death_sound":   nbt.String("minecraft:entity.chicken.death"),
					"hurt_sound":    nbt.String("minecraft:entity.chicken.hurt"),
					"step_sound":    nbt.String("minecraft:entity.chicken.step"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_chicken.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_chicken.death"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_chicken.hurt"),
					"step_sound":    nbt.String("minecraft:entity.baby_chicken.step"),
				},
			}},
			{Name: "minecraft:picky", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.chicken_picky.ambient"),
					"death_sound":   nbt.String("minecraft:entity.chicken_picky.death"),
					"hurt_sound":    nbt.String("minecraft:entity.chicken_picky.hurt"),
					"step_sound":    nbt.String("minecraft:entity.chicken.step"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_chicken.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_chicken.death"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_chicken.hurt"),
					"step_sound":    nbt.String("minecraft:entity.baby_chicken.step"),
				},
			}},
		}},
		{Name: "minecraft:cow_variant", Entries: []Entry{
			{Name: "minecraft:cold", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cow/cow_cold"),
				"baby_asset_id": nbt.String("minecraft:entity/cow/cow_cold_baby"),
				"model":         nbt.String("cold"),
			}},
			{Name: "minecraft:temperate", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cow/cow_temperate"),
				"baby_asset_id": nbt.String("minecraft:entity/cow/cow_temperate_baby"),
			}},
			{Name: "minecraft:warm", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/cow/cow_warm"),
				"baby_asset_id": nbt.String("minecraft:entity/cow/cow_warm_baby"),
				"model":         nbt.String("warm"),
			}},
		}},
		{Name: "minecraft:cow_sound_variant", Entries: []Entry{
			{Name: "minecraft:classic", Data: nbt.Compound{
				"ambient_sound": nbt.String("minecraft:entity.cow.ambient"),
				"death_sound":   nbt.String("minecraft:entity.cow.death"),
				"hurt_sound":    nbt.String("minecraft:entity.cow.hurt"),
				"step_sound":    nbt.String("minecraft:entity.cow.step"),
			}},
			{Name: "minecraft:moody", Data: nbt.Compound{
				"ambient_sound": nbt.String("minecraft:entity.cow_moody.ambient"),
				"death_sound":   nbt.String("minecraft:entity.cow_moody.death"),
				"hurt_sound":    nbt.String("minecraft:entity.cow_moody.hurt"),
				"step_sound":    nbt.String("minecraft:entity.cow_moody.step"),
			}},
		}},
		{Name: "minecraft:frog_variant", Entries: []Entry{
			{Name: "minecraft:cold", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:entity/frog/frog_cold"),
			}},
			{Name: "minecraft:temperate", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:entity/frog/frog_temperate"),
			}},
			{Name: "minecraft:warm", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:entity/frog/frog_warm"),
			}},
		}},
		{Name: "minecraft:painting_variant", Entries: []Entry{
			{Name: "minecraft:alban", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:alban"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.alban.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.alban.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:aztec", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:aztec"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.aztec.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.aztec.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:aztec2", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:aztec2"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.aztec2.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.aztec2.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:backyard", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:backyard"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.backyard.author"),
				},
				"height": nbt.Int(4),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.backyard.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:baroque", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:baroque"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.baroque.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.baroque.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:bomb", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:bomb"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.bomb.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.bomb.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:bouquet", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:bouquet"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.bouquet.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.bouquet.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:burning_skull", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:burning_skull"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.burning_skull.author"),
				},
				"height": nbt.Int(4),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.burning_skull.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:bust", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:bust"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.bust.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.bust.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:cavebird", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:cavebird"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.cavebird.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.cavebird.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:changing", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:changing"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.changing.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.changing.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:cotan", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:cotan"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.cotan.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.cotan.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:courbet", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:courbet"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.courbet.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.courbet.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:creebet", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:creebet"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.creebet.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.creebet.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:dennis", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:dennis"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.dennis.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.dennis.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:donkey_kong", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:donkey_kong"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.donkey_kong.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.donkey_kong.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:earth", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:earth"),
				"height":   nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.earth.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:endboss", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:endboss"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.endboss.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.endboss.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:fern", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:fern"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.fern.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.fern.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:fighters", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:fighters"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.fighters.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.fighters.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:finding", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:finding"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.finding.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.finding.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:fire", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:fire"),
				"height":   nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.fire.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:graham", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:graham"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.graham.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.graham.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:humble", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:humble"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.humble.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.humble.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:kebab", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:kebab"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.kebab.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.kebab.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:lowmist", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:lowmist"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.lowmist.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.lowmist.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:match", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:match"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.match.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.match.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:meditative", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:meditative"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.meditative.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.meditative.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:orb", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:orb"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.orb.author"),
				},
				"height": nbt.Int(4),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.orb.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:owlemons", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:owlemons"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.owlemons.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.owlemons.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:passage", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:passage"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.passage.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.passage.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:pigscene", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:pigscene"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.pigscene.author"),
				},
				"height": nbt.Int(4),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.pigscene.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:plant", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:plant"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.plant.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.plant.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:pointer", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:pointer"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.pointer.author"),
				},
				"height": nbt.Int(4),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.pointer.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:pond", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:pond"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.pond.author"),
				},
				"height": nbt.Int(4),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.pond.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:pool", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:pool"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.pool.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.pool.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:prairie_ride", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:prairie_ride"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.prairie_ride.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.prairie_ride.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:sea", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:sea"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.sea.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.sea.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:skeleton", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:skeleton"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.skeleton.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.skeleton.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:skull_and_roses", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:skull_and_roses"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.skull_and_roses.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.skull_and_roses.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:stage", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:stage"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.stage.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.stage.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:sunflowers", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:sunflowers"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.sunflowers.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.sunflowers.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:sunset", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:sunset"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.sunset.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.sunset.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:tides", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:tides"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.tides.author"),
				},
				"height": nbt.Int(3),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.tides.title"),
				},
				"width": nbt.Int(3),
			}},
			{Name: "minecraft:unpacked", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:unpacked"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.unpacked.author"),
				},
				"height": nbt.Int(4),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.unpacked.title"),
				},
				"width": nbt.Int(4),
			}},
			{Name: "minecraft:void", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:void"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.void.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.void.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:wanderer", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:wanderer"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.wanderer.author"),
				},
				"height": nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.wanderer.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:wasteland", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:wasteland"),
				"author": nbt.Compound{
					"color":     nbt.String("gray"),
					"translate": nbt.String("painting.minecraft.wasteland.author"),
				},
				"height": nbt.Int(1),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.wasteland.title"),
				},
				"width": nbt.Int(1),
			}},
			{Name: "minecraft:water", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:water"),
				"height":   nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.water.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:wind", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:wind"),
				"height":   nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.wind.title"),
				},
				"width": nbt.Int(2),
			}},
			{Name: "minecraft:wither", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:wither"),
				"height":   nbt.Int(2),
				"title": nbt.Compound{
					"color":     nbt.String("yellow"),
					"translate": nbt.String("painting.minecraft.wither.title"),
				},
				"width": nbt.Int(2),
			}},
		}},
		{Name: "minecraft:pig_variant", Entries: []Entry{
			{Name: "minecraft:cold", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/pig/pig_cold"),
				"baby_asset_id": nbt.String("minecraft:entity/pig/pig_cold_baby"),
				"model":         nbt.String("cold"),
			}},
			{Name: "minecraft:temperate", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/pig/pig_temperate"),
				"baby_asset_id": nbt.String("minecraft:entity/pig/pig_temperate_baby"),
			}},
			{Name: "minecraft:warm", Data: nbt.Compound{
				"asset_id":      nbt.String("minecraft:entity/pig/pig_warm"),
				"baby_asset_id": nbt.String("minecraft:entity/pig/pig_warm_baby"),
			}},
		}},
		{Name: "minecraft:pig_sound_variant", Entries: []Entry{
			{Name: "minecraft:big", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.pig_big.ambient"),
					"death_sound":   nbt.String("minecraft:entity.pig_big.death"),
					"eat_sound":     nbt.String("minecraft:entity.pig_big.eat"),
					"hurt_sound":    nbt.String("minecraft:entity.pig_big.hurt"),
					"step_sound":    nbt.String("minecraft:entity.pig.step"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_pig.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_pig.death"),
					"eat_sound":     nbt.String("minecraft:entity.baby_pig.eat"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_pig.hurt"),
					"step_sound":    nbt.String("minecraft:entity.baby_pig.step"),
				},
			}},
			{Name: "minecraft:classic", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.pig.ambient"),
					"death_sound":   nbt.String("minecraft:entity.pig.death"),
					"eat_sound":     nbt.String("minecraft:entity.pig.eat"),
					"hurt_sound":    nbt.String("minecraft:entity.pig.hurt"),
					"step_sound":    nbt.String("minecraft:entity.pig.step"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_pig.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_pig.death"),
					"eat_sound":     nbt.String("minecraft:entity.baby_pig.eat"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_pig.hurt"),
					"step_sound":    nbt.String("minecraft:entity.baby_pig.step"),
				},
			}},
			{Name: "minecraft:mini", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.pig_mini.ambient"),
					"death_sound":   nbt.String("minecraft:entity.pig_mini.death"),
					"eat_sound":     nbt.String("minecraft:entity.pig_mini.eat"),
					"hurt_sound":    nbt.String("minecraft:entity.pig_mini.hurt"),
					"step_sound":    nbt.String("minecraft:entity.pig.step"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_pig.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_pig.death"),
					"eat_sound":     nbt.String("minecraft:entity.baby_pig.eat"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_pig.hurt"),
					"step_sound":    nbt.String("minecraft:entity.baby_pig.step"),
				},
			}},
		}},
		{Name: "minecraft:wolf_variant", Entries: []Entry{
			{Name: "minecraft:ashen", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_ashen_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_ashen_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_ashen"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_ashen_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_ashen_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_ashen_baby"),
				},
			}},
			{Name: "minecraft:black", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_black_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_black_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_black"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_black_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_black_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_black_baby"),
				},
			}},
			{Name: "minecraft:chestnut", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_chestnut_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_chestnut_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_chestnut"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_chestnut_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_chestnut_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_chestnut_baby"),
				},
			}},
			{Name: "minecraft:pale", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_baby"),
				},
			}},
			{Name: "minecraft:rusty", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_rusty_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_rusty_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_rusty"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_rusty_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_rusty_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_rusty_baby"),
				},
			}},
			{Name: "minecraft:snowy", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_snowy_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_snowy_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_snowy"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_snowy_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_snowy_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_snowy_baby"),
				},
			}},
			{Name: "minecraft:spotted", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_spotted_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_spotted_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_spotted"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_spotted_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_spotted_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_spotted_baby"),
				},
			}},
			{Name: "minecraft:striped", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_striped_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_striped_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_striped"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_striped_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_striped_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_striped_baby"),
				},
			}},
			{Name: "minecraft:woods", Data: nbt.Compound{
				"assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_woods_angry"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_woods_tame"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_woods"),
				},
				"baby_assets": nbt.Compound{
					"angry": nbt.String("minecraft:entity/wolf/wolf_woods_angry_baby"),
					"tame":  nbt.String("minecraft:entity/wolf/wolf_woods_tame_baby"),
					"wild":  nbt.String("minecraft:entity/wolf/wolf_woods_baby"),
				},
			}},
		}},
		{Name: "minecraft:wolf_sound_variant", Entries: []Entry{
			{Name: "minecraft:angry", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.wolf_angry.ambient"),
					"death_sound":   nbt.String("minecraft:entity.wolf_angry.death"),
					"growl_sound":   nbt.String("minecraft:entity.wolf_angry.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.wolf_angry.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.wolf_angry.pant"),
					"step_sound":    nbt.String("minecraft:entity.wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.wolf_angry.whine"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_wolf.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_wolf.death"),
					"growl_sound":   nbt.String("minecraft:entity.baby_wolf.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_wolf.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.baby_wolf.pant"),
					"step_sound":    nbt.String("minecraft:entity.baby_wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.baby_wolf.whine"),
				},
			}},
			{Name: "minecraft:big", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.wolf_big.ambient"),
					"death_sound":   nbt.String("minecraft:entity.wolf_big.death"),
					"growl_sound":   nbt.String("minecraft:entity.wolf_big.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.wolf_big.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.wolf_big.pant"),
					"step_sound":    nbt.String("minecraft:entity.wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.wolf_big.whine"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_wolf.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_wolf.death"),
					"growl_sound":   nbt.String("minecraft:entity.baby_wolf.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_wolf.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.baby_wolf.pant"),
					"step_sound":    nbt.String("minecraft:entity.baby_wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.baby_wolf.whine"),
				},
			}},
			{Name: "minecraft:classic", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.wolf.ambient"),
					"death_sound":   nbt.String("minecraft:entity.wolf.death"),
					"growl_sound":   nbt.String("minecraft:entity.wolf.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.wolf.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.wolf.pant"),
					"step_sound":    nbt.String("minecraft:entity.wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.wolf.whine"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_wolf.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_wolf.death"),
					"growl_sound":   nbt.String("minecraft:entity.baby_wolf.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_wolf.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.baby_wolf.pant"),
					"step_sound":    nbt.String("minecraft:entity.baby_wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.baby_wolf.whine"),
				},
			}},
			{Name: "minecraft:cute", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.wolf_cute.ambient"),
					"death_sound":   nbt.String("minecraft:entity.wolf_cute.death"),
					"growl_sound":   nbt.String("minecraft:entity.wolf_cute.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.wolf_cute.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.wolf_cute.pant"),
					"step_sound":    nbt.String("minecraft:entity.wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.wolf_cute.whine"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_wolf.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_wolf.death"),
					"growl_sound":   nbt.String("minecraft:entity.baby_wolf.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_wolf.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.baby_wolf.pant"),
					"step_sound":    nbt.String("minecraft:entity.baby_wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.baby_wolf.whine"),
				},
			}},
			{Name: "minecraft:grumpy", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.wolf_grumpy.ambient"),
					"death_sound":   nbt.String("minecraft:entity.wolf_grumpy.death"),
					"growl_sound":   nbt.String("minecraft:entity.wolf_grumpy.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.wolf_grumpy.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.wolf_grumpy.pant"),
					"step_sound":    nbt.String("minecraft:entity.wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.wolf_grumpy.whine"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_wolf.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_wolf.death"),
					"growl_sound":   nbt.String("minecraft:entity.baby_wolf.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_wolf.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.baby_wolf.pant"),
					"step_sound":    nbt.String("minecraft:entity.baby_wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.baby_wolf.whine"),
				},
			}},
			{Name: "minecraft:puglin", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.wolf_puglin.ambient"),
					"death_sound":   nbt.String("minecraft:entity.wolf_puglin.death"),
					"growl_sound":   nbt.String("minecraft:entity.wolf_puglin.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.wolf_puglin.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.wolf_puglin.pant"),
					"step_sound":    nbt.String("minecraft:entity.wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.wolf_puglin.whine"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_wolf.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_wolf.death"),
					"growl_sound":   nbt.String("minecraft:entity.baby_wolf.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_wolf.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.baby_wolf.pant"),
					"step_sound":    nbt.String("minecraft:entity.baby_wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.baby_wolf.whine"),
				},
			}},
			{Name: "minecraft:sad", Data: nbt.Compound{
				"adult_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.wolf_sad.ambient"),
					"death_sound":   nbt.String("minecraft:entity.wolf_sad.death"),
					"growl_sound":   nbt.String("minecraft:entity.wolf_sad.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.wolf_sad.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.wolf_sad.pant"),
					"step_sound":    nbt.String("minecraft:entity.wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.wolf_sad.whine"),
				},
				"baby_sounds": nbt.Compound{
					"ambient_sound": nbt.String("minecraft:entity.baby_wolf.ambient"),
					"death_sound":   nbt.String("minecraft:entity.baby_wolf.death"),
					"growl_sound":   nbt.String("minecraft:entity.baby_wolf.growl"),
					"hurt_sound":    nbt.String("minecraft:entity.baby_wolf.hurt"),
					"pant_sound":    nbt.String("minecraft:entity.baby_wolf.pant"),
					"step_sound":    nbt.String("minecraft:entity.baby_wolf.step"),
					"whine_sound":   nbt.String("minecraft:entity.baby_wolf.whine"),
				},
			}},
		}},
		{Name: "minecraft:zombie_nautilus_variant", Entries: []Entry{
			{Name: "minecraft:temperate", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:entity/nautilus/zombie_nautilus"),
			}},
			{Name: "minecraft:warm", Data: nbt.Compound{
				"asset_id": nbt.String("minecraft:entity/nautilus/zombie_nautilus_coral"),
				"model":    nbt.String("warm"),
			}},
		}},
		{Name: "minecraft:damage_type", Entries: []Entry{
			{Name: "minecraft:arrow", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("arrow"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:bad_respawn_point", Data: nbt.Compound{
				"death_message_type": nbt.String("intentional_game_design"),
				"exhaustion":         nbt.Double(0.1),
				"message_id":         nbt.String("badRespawnPoint"),
				"scaling":            nbt.String("always"),
			}},
			{Name: "minecraft:cactus", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("cactus"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:campfire", Data: nbt.Compound{
				"effects":    nbt.String("burning"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("inFire"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:cramming", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("cramming"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:dragon_breath", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("dragonBreath"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:drown", Data: nbt.Compound{
				"effects":    nbt.String("drowning"),
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("drown"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:dry_out", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("dryout"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:ender_pearl", Data: nbt.Compound{
				"death_message_type": nbt.String("fall_variants"),
				"exhaustion":         nbt.Double(0.0),
				"message_id":         nbt.String("fall"),
				"scaling":            nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:explosion", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("explosion"),
				"scaling":    nbt.String("always"),
			}},
			{Name: "minecraft:fall", Data: nbt.Compound{
				"death_message_type": nbt.String("fall_variants"),
				"exhaustion":         nbt.Double(0.0),
				"message_id":         nbt.String("fall"),
				"scaling":            nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:falling_anvil", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("anvil"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:falling_block", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("fallingBlock"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:falling_stalactite", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("fallingStalactite"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:fireball", Data: nbt.Compound{
				"effects":    nbt.String("burning"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("fireball"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:fireworks", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("fireworks"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:fly_into_wall", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("flyIntoWall"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:freeze", Data: nbt.Compound{
				"effects":    nbt.String("freezing"),
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("freeze"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:generic", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("generic"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:generic_kill", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("genericKill"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:hot_floor", Data: nbt.Compound{
				"effects":    nbt.String("burning"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("hotFloor"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:in_fire", Data: nbt.Compound{
				"effects":    nbt.String("burning"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("inFire"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:in_wall", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("inWall"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:indirect_magic", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("indirectMagic"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:lava", Data: nbt.Compound{
				"effects":    nbt.String("burning"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("lava"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:lightning_bolt", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("lightningBolt"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:mace_smash", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("mace_smash"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:magic", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("magic"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:mob_attack", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("mob"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:mob_attack_no_aggro", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("mob"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:mob_projectile", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("mob"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:on_fire", Data: nbt.Compound{
				"effects":    nbt.String("burning"),
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("onFire"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:out_of_world", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("outOfWorld"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:outside_border", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("outsideBorder"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:player_attack", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("player"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:player_explosion", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("explosion.player"),
				"scaling":    nbt.String("always"),
			}},
			{Name: "minecraft:sonic_boom", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("sonic_boom"),
				"scaling":    nbt.String("always"),
			}},
			{Name: "minecraft:spear", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("spear"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:spit", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("mob"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:stalagmite", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("stalagmite"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:starve", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("starve"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:sting", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("sting"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:sulfur_cube_hot", Data: nbt.Compound{
				"effects":    nbt.String("burning"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("sulfurCubeHot"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:sweet_berry_bush", Data: nbt.Compound{
				"effects":    nbt.String("poking"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("sweetBerryBush"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:thorns", Data: nbt.Compound{
				"effects":    nbt.String("thorns"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("thorns"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:thrown", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("thrown"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:trident", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("trident"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:unattributed_fireball", Data: nbt.Compound{
				"effects":    nbt.String("burning"),
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("onFire"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:wind_charge", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("mob"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:wither", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.0),
				"message_id": nbt.String("wither"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
			{Name: "minecraft:wither_skull", Data: nbt.Compound{
				"exhaustion": nbt.Double(0.1),
				"message_id": nbt.String("witherSkull"),
				"scaling":    nbt.String("when_caused_by_living_non_player"),
			}},
		}},
		{Name: "minecraft:chat_type", Entries: []Entry{
			{Name: "minecraft:chat", Data: nbt.Compound{
				"chat": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.text"),
				},
				"narration": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.text.narrate"),
				},
			}},
			{Name: "minecraft:emote_command", Data: nbt.Compound{
				"chat": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.emote"),
				},
				"narration": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.emote"),
				},
			}},
			{Name: "minecraft:msg_command_incoming", Data: nbt.Compound{
				"chat": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"style": nbt.Compound{
						"color":  nbt.String("gray"),
						"italic": nbt.Byte(1),
					},
					"translation_key": nbt.String("commands.message.display.incoming"),
				},
				"narration": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.text.narrate"),
				},
			}},
			{Name: "minecraft:msg_command_outgoing", Data: nbt.Compound{
				"chat": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("target"),
						nbt.String("content"),
					}},
					"style": nbt.Compound{
						"color":  nbt.String("gray"),
						"italic": nbt.Byte(1),
					},
					"translation_key": nbt.String("commands.message.display.outgoing"),
				},
				"narration": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.text.narrate"),
				},
			}},
			{Name: "minecraft:say_command", Data: nbt.Compound{
				"chat": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.announcement"),
				},
				"narration": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.text.narrate"),
				},
			}},
			{Name: "minecraft:team_msg_command_incoming", Data: nbt.Compound{
				"chat": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("target"),
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.team.text"),
				},
				"narration": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.text.narrate"),
				},
			}},
			{Name: "minecraft:team_msg_command_outgoing", Data: nbt.Compound{
				"chat": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("target"),
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.team.sent"),
				},
				"narration": nbt.Compound{
					"parameters": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("sender"),
						nbt.String("content"),
					}},
					"translation_key": nbt.String("chat.type.text.narrate"),
				},
			}},
		}},
		{Name: "minecraft:enchantment", Entries: []Entry{
			{Name: "minecraft:aqua_affinity", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.aqua_affinity"),
				},
				"effects": nbt.Compound{
					"minecraft:attributes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"amount": nbt.Compound{
								"base":                  nbt.Double(4.0),
								"per_level_above_first": nbt.Double(4.0),
								"type":                  nbt.String("minecraft:linear"),
							},
							"attribute": nbt.String("minecraft:submerged_mining_speed"),
							"id":        nbt.String("minecraft:enchantment.aqua_affinity"),
							"operation": nbt.String("add_multiplied_total"),
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(41),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(1),
					"per_level_above_first": nbt.Int(0),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("head"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:bane_of_arthropods", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.bane_of_arthropods"),
				},
				"effects": nbt.Compound{
					"minecraft:damage": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(2.5),
									"per_level_above_first": nbt.Double(2.5),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:entity_properties"),
								"entity":    nbt.String("this"),
								"predicate": nbt.Compound{
									"minecraft:entity_type": nbt.List{},
								},
							},
						},
					}},
					"minecraft:post_attack": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"affected": nbt.String("victim"),
							"effect": nbt.Compound{
								"max_amplifier": nbt.Double(3.0),
								"max_duration": nbt.Compound{
									"base":                  nbt.Double(1.5),
									"per_level_above_first": nbt.Double(0.5),
									"type":                  nbt.String("minecraft:linear"),
								},
								"min_amplifier": nbt.Double(3.0),
								"min_duration":  nbt.Double(1.5),
								"to_apply":      nbt.String("minecraft:slowness"),
								"type":          nbt.String("minecraft:apply_mob_effect"),
							},
							"enchanted": nbt.String("attacker"),
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("this"),
										"predicate": nbt.Compound{
											"minecraft:entity_type": nbt.List{},
										},
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:damage_source_properties"),
										"predicate": nbt.Compound{
											"is_direct": nbt.Byte(1),
										},
									},
								}},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(8),
				},
				"max_level": nbt.Int(5),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(8),
				},
				"primary_items": nbt.List{},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:binding_curse", Data: nbt.Compound{
				"anvil_cost": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.binding_curse"),
				},
				"effects": nbt.Compound{
					"minecraft:prevent_armor_change": nbt.Compound{},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(0),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("armor"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(1),
			}},
			{Name: "minecraft:blast_protection", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.blast_protection"),
				},
				"effects": nbt.Compound{
					"minecraft:attributes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"amount": nbt.Compound{
								"base":                  nbt.Double(0.15),
								"per_level_above_first": nbt.Double(0.15),
								"type":                  nbt.String("minecraft:linear"),
							},
							"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
							"id":        nbt.String("minecraft:enchantment.blast_protection"),
							"operation": nbt.String("add_value"),
						},
					}},
					"minecraft:damage_protection": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(2.0),
									"per_level_above_first": nbt.Double(2.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:damage_source_properties"),
								"predicate": nbt.Compound{
									"tags": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
										nbt.Compound{
											"expected": nbt.Byte(1),
											"id":       nbt.String("minecraft:is_explosion"),
										},
										nbt.Compound{
											"expected": nbt.Byte(0),
											"id":       nbt.String("minecraft:bypasses_invulnerability"),
										},
									}},
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(13),
					"per_level_above_first": nbt.Int(8),
				},
				"max_level": nbt.Int(4),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(8),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("armor"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:breach", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.breach"),
				},
				"effects": nbt.Compound{
					"minecraft:armor_effectiveness": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(-0.15),
									"per_level_above_first": nbt.Double(-0.15),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(65),
					"per_level_above_first": nbt.Int(9),
				},
				"max_level": nbt.Int(4),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(15),
					"per_level_above_first": nbt.Int(9),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:channeling", Data: nbt.Compound{
				"anvil_cost": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.channeling"),
				},
				"effects": nbt.Compound{
					"minecraft:hit_block": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"effects": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"entity": nbt.String("minecraft:lightning_bolt"),
										"type":   nbt.String("minecraft:summon_entity"),
									},
									nbt.Compound{
										"pitch":  nbt.Double(1.0),
										"sound":  nbt.String("minecraft:item.trident.thunder"),
										"type":   nbt.String("minecraft:play_sound"),
										"volume": nbt.Double(5.0),
									},
								}},
								"type": nbt.String("minecraft:all_of"),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"condition":  nbt.String("minecraft:weather_check"),
										"thundering": nbt.Byte(1),
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("this"),
										"predicate": nbt.Compound{
											"minecraft:entity_type": nbt.String("minecraft:trident"),
										},
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:location_check"),
										"predicate": nbt.Compound{
											"block": nbt.Compound{
												"blocks": nbt.List{},
											},
											"can_see_sky": nbt.Byte(1),
										},
									},
								}},
							},
						},
					}},
					"minecraft:post_attack": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"affected": nbt.String("victim"),
							"effect": nbt.Compound{
								"effects": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"entity": nbt.String("minecraft:lightning_bolt"),
										"type":   nbt.String("minecraft:summon_entity"),
									},
									nbt.Compound{
										"pitch":  nbt.Double(1.0),
										"sound":  nbt.String("minecraft:item.trident.thunder"),
										"type":   nbt.String("minecraft:play_sound"),
										"volume": nbt.Double(5.0),
									},
								}},
								"type": nbt.String("minecraft:all_of"),
							},
							"enchanted": nbt.String("attacker"),
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"condition":  nbt.String("minecraft:weather_check"),
										"thundering": nbt.Byte(1),
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("this"),
										"predicate": nbt.Compound{
											"minecraft:location": nbt.Compound{
												"can_see_sky": nbt.Byte(1),
											},
										},
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("direct_attacker"),
										"predicate": nbt.Compound{
											"minecraft:entity_type": nbt.String("minecraft:trident"),
										},
									},
								}},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(0),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(1),
			}},
			{Name: "minecraft:density", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.density"),
				},
				"effects": nbt.Compound{
					"minecraft:smash_damage_per_fallen_block": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(0.5),
									"per_level_above_first": nbt.Double(0.5),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(8),
				},
				"max_level": nbt.Int(5),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(8),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:depth_strider", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.depth_strider"),
				},
				"effects": nbt.Compound{
					"minecraft:attributes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"amount": nbt.Compound{
								"base":                  nbt.Double(0.33333334),
								"per_level_above_first": nbt.Double(0.33333334),
								"type":                  nbt.String("minecraft:linear"),
							},
							"attribute": nbt.String("minecraft:water_movement_efficiency"),
							"id":        nbt.String("minecraft:enchantment.depth_strider"),
							"operation": nbt.String("add_value"),
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(10),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(10),
					"per_level_above_first": nbt.Int(10),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("feet"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:efficiency", Data: nbt.Compound{
				"anvil_cost": nbt.Int(1),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.efficiency"),
				},
				"effects": nbt.Compound{
					"minecraft:attributes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"amount": nbt.Compound{
								"added": nbt.Double(1.0),
								"type":  nbt.String("minecraft:levels_squared"),
							},
							"attribute": nbt.String("minecraft:mining_efficiency"),
							"id":        nbt.String("minecraft:enchantment.efficiency"),
							"operation": nbt.String("add_value"),
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(51),
					"per_level_above_first": nbt.Int(10),
				},
				"max_level": nbt.Int(5),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(1),
					"per_level_above_first": nbt.Int(10),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(10),
			}},
			{Name: "minecraft:feather_falling", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.feather_falling"),
				},
				"effects": nbt.Compound{
					"minecraft:damage_protection": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(3.0),
									"per_level_above_first": nbt.Double(3.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:damage_source_properties"),
								"predicate": nbt.Compound{
									"tags": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
										nbt.Compound{
											"expected": nbt.Byte(1),
											"id":       nbt.String("minecraft:is_fall"),
										},
										nbt.Compound{
											"expected": nbt.Byte(0),
											"id":       nbt.String("minecraft:bypasses_invulnerability"),
										},
									}},
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(11),
					"per_level_above_first": nbt.Int(6),
				},
				"max_level": nbt.Int(4),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(6),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("armor"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:fire_aspect", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.fire_aspect"),
				},
				"effects": nbt.Compound{
					"minecraft:post_attack": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"affected": nbt.String("victim"),
							"effect": nbt.Compound{
								"duration": nbt.Compound{
									"base":                  nbt.Double(4.0),
									"per_level_above_first": nbt.Double(4.0),
									"type":                  nbt.String("minecraft:linear"),
								},
								"type": nbt.String("minecraft:ignite"),
							},
							"enchanted": nbt.String("attacker"),
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:damage_source_properties"),
								"predicate": nbt.Compound{
									"is_direct": nbt.Byte(1),
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(60),
					"per_level_above_first": nbt.Int(20),
				},
				"max_level": nbt.Int(2),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(10),
					"per_level_above_first": nbt.Int(20),
				},
				"primary_items": nbt.List{},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:fire_protection", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.fire_protection"),
				},
				"effects": nbt.Compound{
					"minecraft:attributes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"amount": nbt.Compound{
								"base":                  nbt.Double(-0.15),
								"per_level_above_first": nbt.Double(-0.15),
								"type":                  nbt.String("minecraft:linear"),
							},
							"attribute": nbt.String("minecraft:burning_time"),
							"id":        nbt.String("minecraft:enchantment.fire_protection"),
							"operation": nbt.String("add_multiplied_base"),
						},
					}},
					"minecraft:damage_protection": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(2.0),
									"per_level_above_first": nbt.Double(2.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"condition": nbt.String("minecraft:damage_source_properties"),
										"predicate": nbt.Compound{
											"tags": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
												nbt.Compound{
													"expected": nbt.Byte(1),
													"id":       nbt.String("minecraft:is_fire"),
												},
												nbt.Compound{
													"expected": nbt.Byte(0),
													"id":       nbt.String("minecraft:bypasses_invulnerability"),
												},
											}},
										},
									},
								}},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(18),
					"per_level_above_first": nbt.Int(8),
				},
				"max_level": nbt.Int(4),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(10),
					"per_level_above_first": nbt.Int(8),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("armor"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:flame", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.flame"),
				},
				"effects": nbt.Compound{
					"minecraft:projectile_spawned": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"duration": nbt.Double(100.0),
								"type":     nbt.String("minecraft:ignite"),
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(20),
					"per_level_above_first": nbt.Int(0),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:fortune", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.fortune"),
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(65),
					"per_level_above_first": nbt.Int(9),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(15),
					"per_level_above_first": nbt.Int(9),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:frost_walker", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.frost_walker"),
				},
				"effects": nbt.Compound{
					"minecraft:damage_immunity": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:damage_source_properties"),
								"predicate": nbt.Compound{
									"tags": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
										nbt.Compound{
											"expected": nbt.Byte(1),
											"id":       nbt.String("minecraft:burn_from_stepping"),
										},
										nbt.Compound{
											"expected": nbt.Byte(0),
											"id":       nbt.String("minecraft:bypasses_invulnerability"),
										},
									}},
								},
							},
						},
					}},
					"minecraft:location_changed": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"block_state": nbt.Compound{
									"state": nbt.Compound{
										"Name": nbt.String("minecraft:frosted_ice"),
										"Properties": nbt.Compound{
											"age": nbt.String("0"),
										},
									},
									"type": nbt.String("minecraft:simple_state_provider"),
								},
								"height": nbt.Double(1.0),
								"offset": nbt.List{ElementType: nbt.TagInt, Elements: []nbt.Tag{
									nbt.Int(0),
									nbt.Int(-1),
									nbt.Int(0),
								}},
								"predicate": nbt.Compound{
									"predicates": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
										nbt.Compound{
											"offset": nbt.List{ElementType: nbt.TagInt, Elements: []nbt.Tag{
												nbt.Int(0),
												nbt.Int(1),
												nbt.Int(0),
											}},
											"tag":  nbt.String("minecraft:air"),
											"type": nbt.String("minecraft:matching_block_tag"),
										},
										nbt.Compound{
											"blocks": nbt.String("minecraft:water"),
											"type":   nbt.String("minecraft:matching_blocks"),
										},
										nbt.Compound{
											"fluids": nbt.String("minecraft:water"),
											"type":   nbt.String("minecraft:matching_fluids"),
										},
										nbt.Compound{
											"type": nbt.String("minecraft:unobstructed"),
										},
									}},
									"type": nbt.String("minecraft:all_of"),
								},
								"radius": nbt.Compound{
									"max":  nbt.Double(16.0),
									"min":  nbt.Double(0.0),
									"type": nbt.String("minecraft:clamped"),
									"value": nbt.Compound{
										"base":                  nbt.Double(3.0),
										"per_level_above_first": nbt.Double(1.0),
										"type":                  nbt.String("minecraft:linear"),
									},
								},
								"trigger_game_event": nbt.String("minecraft:block_place"),
								"type":               nbt.String("minecraft:replace_disk"),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("this"),
										"predicate": nbt.Compound{
											"minecraft:flags": nbt.Compound{
												"is_on_ground": nbt.Byte(1),
											},
										},
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:inverted"),
										"term": nbt.Compound{
											"condition": nbt.String("minecraft:entity_properties"),
											"entity":    nbt.String("this"),
											"predicate": nbt.Compound{
												"minecraft:vehicle": nbt.Compound{},
											},
										},
									},
								}},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(10),
				},
				"max_level": nbt.Int(2),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(10),
					"per_level_above_first": nbt.Int(10),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("feet"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:impaling", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.impaling"),
				},
				"effects": nbt.Compound{
					"minecraft:damage": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(2.5),
									"per_level_above_first": nbt.Double(2.5),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:entity_properties"),
								"entity":    nbt.String("this"),
								"predicate": nbt.Compound{
									"minecraft:entity_type": nbt.List{},
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(21),
					"per_level_above_first": nbt.Int(8),
				},
				"max_level": nbt.Int(5),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(1),
					"per_level_above_first": nbt.Int(8),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:infinity", Data: nbt.Compound{
				"anvil_cost": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.infinity"),
				},
				"effects": nbt.Compound{
					"minecraft:ammo_use": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type":  nbt.String("minecraft:set"),
								"value": nbt.Double(0.0),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:match_tool"),
								"predicate": nbt.Compound{
									"items": nbt.String("minecraft:arrow"),
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(20),
					"per_level_above_first": nbt.Int(0),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(1),
			}},
			{Name: "minecraft:knockback", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.knockback"),
				},
				"effects": nbt.Compound{
					"minecraft:knockback": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(1.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(55),
					"per_level_above_first": nbt.Int(20),
				},
				"max_level": nbt.Int(2),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(20),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:looting", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.looting"),
				},
				"effects": nbt.Compound{
					"minecraft:equipment_drops": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(0.01),
									"per_level_above_first": nbt.Double(0.01),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"enchanted": nbt.String("attacker"),
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:entity_properties"),
								"entity":    nbt.String("attacker"),
								"predicate": nbt.Compound{
									"minecraft:entity_type": nbt.String("minecraft:player"),
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(65),
					"per_level_above_first": nbt.Int(9),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(15),
					"per_level_above_first": nbt.Int(9),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:loyalty", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.loyalty"),
				},
				"effects": nbt.Compound{
					"minecraft:trident_return_acceleration": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(1.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(12),
					"per_level_above_first": nbt.Int(7),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:luck_of_the_sea", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.luck_of_the_sea"),
				},
				"effects": nbt.Compound{
					"minecraft:fishing_luck_bonus": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(1.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(65),
					"per_level_above_first": nbt.Int(9),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(15),
					"per_level_above_first": nbt.Int(9),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:lunge", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.lunge"),
				},
				"effects": nbt.Compound{
					"minecraft:post_piercing_attack": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"effects": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"amount": nbt.Double(1.0),
										"type":   nbt.String("minecraft:change_item_damage"),
									},
									nbt.Compound{
										"amount": nbt.Compound{
											"base":                  nbt.Double(4.0),
											"per_level_above_first": nbt.Double(4.0),
											"type":                  nbt.String("minecraft:linear"),
										},
										"type": nbt.String("minecraft:apply_exhaustion"),
									},
									nbt.Compound{
										"coordinate_scale": nbt.List{ElementType: nbt.TagDouble, Elements: []nbt.Tag{
											nbt.Double(1.0),
											nbt.Double(0.0),
											nbt.Double(1.0),
										}},
										"direction": nbt.List{ElementType: nbt.TagDouble, Elements: []nbt.Tag{
											nbt.Double(0.0),
											nbt.Double(0.0),
											nbt.Double(1.0),
										}},
										"magnitude": nbt.Compound{
											"base":                  nbt.Double(0.458),
											"per_level_above_first": nbt.Double(0.458),
											"type":                  nbt.String("minecraft:linear"),
										},
										"type": nbt.String("minecraft:apply_impulse"),
									},
									nbt.Compound{
										"pitch": nbt.Double(1.0),
										"sound": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
											nbt.String("minecraft:item.spear.lunge_1"),
											nbt.String("minecraft:item.spear.lunge_2"),
											nbt.String("minecraft:item.spear.lunge_3"),
										}},
										"type":   nbt.String("minecraft:play_sound"),
										"volume": nbt.Double(1.0),
									},
								}},
								"type": nbt.String("minecraft:all_of"),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"condition": nbt.String("minecraft:inverted"),
										"term": nbt.Compound{
											"condition": nbt.String("minecraft:entity_properties"),
											"entity":    nbt.String("this"),
											"predicate": nbt.Compound{
												"minecraft:vehicle": nbt.Compound{},
											},
										},
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("this"),
										"predicate": nbt.Compound{
											"minecraft:flags": nbt.Compound{
												"is_fall_flying": nbt.Byte(0),
											},
										},
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("this"),
										"predicate": nbt.Compound{
											"minecraft:flags": nbt.Compound{
												"is_in_water": nbt.Byte(0),
											},
										},
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:any_of"),
										"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
											nbt.Compound{
												"condition": nbt.String("minecraft:inverted"),
												"term": nbt.Compound{
													"condition": nbt.String("minecraft:entity_properties"),
													"entity":    nbt.String("this"),
													"predicate": nbt.Compound{
														"minecraft:type_specific/player": nbt.Compound{},
													},
												},
											},
											nbt.Compound{
												"condition": nbt.String("minecraft:entity_properties"),
												"entity":    nbt.String("this"),
												"predicate": nbt.Compound{
													"minecraft:type_specific/player": nbt.Compound{
														"gamemode": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
															nbt.String("creative"),
														}},
													},
												},
											},
											nbt.Compound{
												"condition": nbt.String("minecraft:entity_properties"),
												"entity":    nbt.String("this"),
												"predicate": nbt.Compound{
													"minecraft:type_specific/player": nbt.Compound{
														"food": nbt.Compound{
															"level": nbt.Compound{
																"min": nbt.Int(7),
															},
														},
													},
												},
											},
										}},
									},
								}},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(8),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(8),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("hand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:lure", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.lure"),
				},
				"effects": nbt.Compound{
					"minecraft:fishing_time_reduction": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(5.0),
									"per_level_above_first": nbt.Double(5.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(65),
					"per_level_above_first": nbt.Int(9),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(15),
					"per_level_above_first": nbt.Int(9),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:mending", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.mending"),
				},
				"effects": nbt.Compound{
					"minecraft:repair_with_xp": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"factor": nbt.Double(2.0),
								"type":   nbt.String("minecraft:multiply"),
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(75),
					"per_level_above_first": nbt.Int(25),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(25),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("any"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:multishot", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.multishot"),
				},
				"effects": nbt.Compound{
					"minecraft:projectile_count": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(2.0),
									"per_level_above_first": nbt.Double(2.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
					"minecraft:projectile_spread": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(10.0),
									"per_level_above_first": nbt.Double(10.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(20),
					"per_level_above_first": nbt.Int(0),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:piercing", Data: nbt.Compound{
				"anvil_cost": nbt.Int(1),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.piercing"),
				},
				"effects": nbt.Compound{
					"minecraft:projectile_piercing": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(1.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(4),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(1),
					"per_level_above_first": nbt.Int(10),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(10),
			}},
			{Name: "minecraft:power", Data: nbt.Compound{
				"anvil_cost": nbt.Int(1),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.power"),
				},
				"effects": nbt.Compound{
					"minecraft:damage": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(0.5),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:entity_properties"),
								"entity":    nbt.String("direct_attacker"),
								"predicate": nbt.Compound{
									"minecraft:entity_type": nbt.List{},
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(16),
					"per_level_above_first": nbt.Int(10),
				},
				"max_level": nbt.Int(5),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(1),
					"per_level_above_first": nbt.Int(10),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(10),
			}},
			{Name: "minecraft:projectile_protection", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.projectile_protection"),
				},
				"effects": nbt.Compound{
					"minecraft:damage_protection": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(2.0),
									"per_level_above_first": nbt.Double(2.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:damage_source_properties"),
								"predicate": nbt.Compound{
									"tags": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
										nbt.Compound{
											"expected": nbt.Byte(1),
											"id":       nbt.String("minecraft:is_projectile"),
										},
										nbt.Compound{
											"expected": nbt.Byte(0),
											"id":       nbt.String("minecraft:bypasses_invulnerability"),
										},
									}},
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(9),
					"per_level_above_first": nbt.Int(6),
				},
				"max_level": nbt.Int(4),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(3),
					"per_level_above_first": nbt.Int(6),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("armor"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:protection", Data: nbt.Compound{
				"anvil_cost": nbt.Int(1),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.protection"),
				},
				"effects": nbt.Compound{
					"minecraft:damage_protection": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(1.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:damage_source_properties"),
								"predicate": nbt.Compound{
									"tags": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
										nbt.Compound{
											"expected": nbt.Byte(0),
											"id":       nbt.String("minecraft:bypasses_invulnerability"),
										},
									}},
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(12),
					"per_level_above_first": nbt.Int(11),
				},
				"max_level": nbt.Int(4),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(1),
					"per_level_above_first": nbt.Int(11),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("armor"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(10),
			}},
			{Name: "minecraft:punch", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.punch"),
				},
				"effects": nbt.Compound{
					"minecraft:knockback": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(1.0),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:entity_properties"),
								"entity":    nbt.String("direct_attacker"),
								"predicate": nbt.Compound{
									"minecraft:entity_type": nbt.List{},
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(37),
					"per_level_above_first": nbt.Int(20),
				},
				"max_level": nbt.Int(2),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(12),
					"per_level_above_first": nbt.Int(20),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:quick_charge", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.quick_charge"),
				},
				"effects": nbt.Compound{
					"minecraft:crossbow_charge_time": nbt.Compound{
						"type": nbt.String("minecraft:add"),
						"value": nbt.Compound{
							"base":                  nbt.Double(-0.25),
							"per_level_above_first": nbt.Double(-0.25),
							"type":                  nbt.String("minecraft:linear"),
						},
					},
					"minecraft:crossbow_charging_sounds": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"end":   nbt.String("minecraft:item.crossbow.loading_end"),
							"start": nbt.String("minecraft:item.crossbow.quick_charge_1"),
						},
						nbt.Compound{
							"end":   nbt.String("minecraft:item.crossbow.loading_end"),
							"start": nbt.String("minecraft:item.crossbow.quick_charge_2"),
						},
						nbt.Compound{
							"end":   nbt.String("minecraft:item.crossbow.loading_end"),
							"start": nbt.String("minecraft:item.crossbow.quick_charge_3"),
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(12),
					"per_level_above_first": nbt.Int(20),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
					nbt.String("offhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:respiration", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.respiration"),
				},
				"effects": nbt.Compound{
					"minecraft:attributes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"amount": nbt.Compound{
								"base":                  nbt.Double(1.0),
								"per_level_above_first": nbt.Double(1.0),
								"type":                  nbt.String("minecraft:linear"),
							},
							"attribute": nbt.String("minecraft:oxygen_bonus"),
							"id":        nbt.String("minecraft:enchantment.respiration"),
							"operation": nbt.String("add_value"),
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(40),
					"per_level_above_first": nbt.Int(10),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(10),
					"per_level_above_first": nbt.Int(10),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("head"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:riptide", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.riptide"),
				},
				"effects": nbt.Compound{
					"minecraft:trident_sound": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
						nbt.String("minecraft:item.trident.riptide_1"),
						nbt.String("minecraft:item.trident.riptide_2"),
						nbt.String("minecraft:item.trident.riptide_3"),
					}},
					"minecraft:trident_spin_attack_strength": nbt.Compound{
						"type": nbt.String("minecraft:add"),
						"value": nbt.Compound{
							"base":                  nbt.Double(1.5),
							"per_level_above_first": nbt.Double(0.75),
							"type":                  nbt.String("minecraft:linear"),
						},
					},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(17),
					"per_level_above_first": nbt.Int(7),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("hand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:sharpness", Data: nbt.Compound{
				"anvil_cost": nbt.Int(1),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.sharpness"),
				},
				"effects": nbt.Compound{
					"minecraft:damage": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(0.5),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(21),
					"per_level_above_first": nbt.Int(11),
				},
				"max_level": nbt.Int(5),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(1),
					"per_level_above_first": nbt.Int(11),
				},
				"primary_items": nbt.List{},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(10),
			}},
			{Name: "minecraft:silk_touch", Data: nbt.Compound{
				"anvil_cost": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.silk_touch"),
				},
				"effects": nbt.Compound{
					"minecraft:block_experience": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type":  nbt.String("minecraft:set"),
								"value": nbt.Double(0.0),
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(65),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(15),
					"per_level_above_first": nbt.Int(0),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(1),
			}},
			{Name: "minecraft:smite", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.smite"),
				},
				"effects": nbt.Compound{
					"minecraft:damage": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"type": nbt.String("minecraft:add"),
								"value": nbt.Compound{
									"base":                  nbt.Double(2.5),
									"per_level_above_first": nbt.Double(2.5),
									"type":                  nbt.String("minecraft:linear"),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:entity_properties"),
								"entity":    nbt.String("this"),
								"predicate": nbt.Compound{
									"minecraft:entity_type": nbt.List{},
								},
							},
						},
					}},
				},
				"exclusive_set": nbt.List{},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(8),
				},
				"max_level": nbt.Int(5),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(8),
				},
				"primary_items": nbt.List{},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:soul_speed", Data: nbt.Compound{
				"anvil_cost": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.soul_speed"),
				},
				"effects": nbt.Compound{
					"minecraft:location_changed": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"effects": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"amount": nbt.Compound{
											"base":                  nbt.Double(0.0405),
											"per_level_above_first": nbt.Double(0.0105),
											"type":                  nbt.String("minecraft:linear"),
										},
										"attribute": nbt.String("minecraft:movement_speed"),
										"id":        nbt.String("minecraft:enchantment.soul_speed"),
										"operation": nbt.String("add_value"),
										"type":      nbt.String("minecraft:attribute"),
									},
									nbt.Compound{
										"amount":    nbt.Double(1.0),
										"attribute": nbt.String("minecraft:movement_efficiency"),
										"id":        nbt.String("minecraft:enchantment.soul_speed"),
										"operation": nbt.String("add_value"),
										"type":      nbt.String("minecraft:attribute"),
									},
								}},
								"type": nbt.String("minecraft:all_of"),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"condition": nbt.String("minecraft:inverted"),
										"term": nbt.Compound{
											"condition": nbt.String("minecraft:entity_properties"),
											"entity":    nbt.String("this"),
											"predicate": nbt.Compound{
												"minecraft:vehicle": nbt.Compound{},
											},
										},
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:any_of"),
										"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
											nbt.Compound{
												"condition": nbt.String("minecraft:all_of"),
												"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
													nbt.Compound{
														"active":    nbt.Byte(1),
														"condition": nbt.String("minecraft:enchantment_active_check"),
													},
													nbt.Compound{
														"condition": nbt.String("minecraft:entity_properties"),
														"entity":    nbt.String("this"),
														"predicate": nbt.Compound{
															"minecraft:flags": nbt.Compound{
																"is_flying": nbt.Byte(0),
															},
														},
													},
													nbt.Compound{
														"condition": nbt.String("minecraft:any_of"),
														"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
															nbt.Compound{
																"condition": nbt.String("minecraft:entity_properties"),
																"entity":    nbt.String("this"),
																"predicate": nbt.Compound{
																	"minecraft:movement_affected_by": nbt.Compound{
																		"block": nbt.Compound{
																			"blocks": nbt.List{},
																		},
																	},
																},
															},
															nbt.Compound{
																"condition": nbt.String("minecraft:entity_properties"),
																"entity":    nbt.String("this"),
																"predicate": nbt.Compound{
																	"minecraft:flags": nbt.Compound{
																		"is_on_ground": nbt.Byte(0),
																	},
																},
															},
														}},
													},
												}},
											},
											nbt.Compound{
												"condition": nbt.String("minecraft:all_of"),
												"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
													nbt.Compound{
														"active":    nbt.Byte(0),
														"condition": nbt.String("minecraft:enchantment_active_check"),
													},
													nbt.Compound{
														"condition": nbt.String("minecraft:entity_properties"),
														"entity":    nbt.String("this"),
														"predicate": nbt.Compound{
															"minecraft:flags": nbt.Compound{
																"is_flying": nbt.Byte(0),
															},
															"minecraft:movement_affected_by": nbt.Compound{
																"block": nbt.Compound{
																	"blocks": nbt.List{},
																},
															},
														},
													},
												}},
											},
										}},
									},
								}},
							},
						},
						nbt.Compound{
							"effect": nbt.Compound{
								"amount": nbt.Double(1.0),
								"type":   nbt.String("minecraft:change_item_damage"),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"chance": nbt.Compound{
											"amount": nbt.Double(0.04),
											"type":   nbt.String("minecraft:enchantment_level"),
										},
										"condition": nbt.String("minecraft:random_chance"),
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("this"),
										"predicate": nbt.Compound{
											"minecraft:flags": nbt.Compound{
												"is_on_ground": nbt.Byte(1),
											},
											"minecraft:movement_affected_by": nbt.Compound{
												"block": nbt.Compound{
													"blocks": nbt.List{},
												},
											},
										},
									},
								}},
							},
						},
					}},
					"minecraft:tick": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"horizontal_position": nbt.Compound{
									"type": nbt.String("in_bounding_box"),
								},
								"horizontal_velocity": nbt.Compound{
									"movement_scale": nbt.Double(-0.2),
								},
								"particle": nbt.Compound{
									"type": nbt.String("minecraft:soul"),
								},
								"speed": nbt.Double(1.0),
								"type":  nbt.String("minecraft:spawn_particles"),
								"vertical_position": nbt.Compound{
									"offset": nbt.Double(0.1),
									"type":   nbt.String("entity_position"),
								},
								"vertical_velocity": nbt.Compound{
									"base": nbt.Double(0.1),
								},
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:entity_properties"),
								"entity":    nbt.String("this"),
								"predicate": nbt.Compound{
									"minecraft:flags": nbt.Compound{
										"is_flying":    nbt.Byte(0),
										"is_on_ground": nbt.Byte(1),
									},
									"minecraft:movement": nbt.Compound{
										"horizontal_speed": nbt.Compound{
											"min": nbt.Double(9.999999747378752e-06),
										},
									},
									"minecraft:movement_affected_by": nbt.Compound{
										"block": nbt.Compound{
											"blocks": nbt.List{},
										},
									},
									"minecraft:periodic_tick": nbt.Int(5),
								},
							},
						},
						nbt.Compound{
							"effect": nbt.Compound{
								"pitch": nbt.Compound{
									"max_exclusive": nbt.Double(1.0),
									"min_inclusive": nbt.Double(0.6),
									"type":          nbt.String("minecraft:uniform"),
								},
								"sound":  nbt.String("minecraft:particle.soul_escape"),
								"type":   nbt.String("minecraft:play_sound"),
								"volume": nbt.Double(0.6),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:all_of"),
								"terms": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"chance":    nbt.Double(0.35),
										"condition": nbt.String("minecraft:random_chance"),
									},
									nbt.Compound{
										"condition": nbt.String("minecraft:entity_properties"),
										"entity":    nbt.String("this"),
										"predicate": nbt.Compound{
											"minecraft:flags": nbt.Compound{
												"is_flying":    nbt.Byte(0),
												"is_on_ground": nbt.Byte(1),
											},
											"minecraft:movement": nbt.Compound{
												"horizontal_speed": nbt.Compound{
													"min": nbt.Double(9.999999747378752e-06),
												},
											},
											"minecraft:movement_affected_by": nbt.Compound{
												"block": nbt.Compound{
													"blocks": nbt.List{},
												},
											},
											"minecraft:periodic_tick": nbt.Int(5),
										},
									},
								}},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(10),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(10),
					"per_level_above_first": nbt.Int(10),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("feet"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(1),
			}},
			{Name: "minecraft:sweeping_edge", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.sweeping_edge"),
				},
				"effects": nbt.Compound{
					"minecraft:attributes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"amount": nbt.Compound{
								"denominator": nbt.Compound{
									"base":                  nbt.Double(2.0),
									"per_level_above_first": nbt.Double(1.0),
									"type":                  nbt.String("minecraft:linear"),
								},
								"numerator": nbt.Compound{
									"base":                  nbt.Double(1.0),
									"per_level_above_first": nbt.Double(1.0),
									"type":                  nbt.String("minecraft:linear"),
								},
								"type": nbt.String("minecraft:fraction"),
							},
							"attribute": nbt.String("minecraft:sweeping_damage_ratio"),
							"id":        nbt.String("minecraft:enchantment.sweeping_edge"),
							"operation": nbt.String("add_value"),
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(20),
					"per_level_above_first": nbt.Int(9),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(9),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
			{Name: "minecraft:swift_sneak", Data: nbt.Compound{
				"anvil_cost": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.swift_sneak"),
				},
				"effects": nbt.Compound{
					"minecraft:attributes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"amount": nbt.Compound{
								"base":                  nbt.Double(0.15),
								"per_level_above_first": nbt.Double(0.15),
								"type":                  nbt.String("minecraft:linear"),
							},
							"attribute": nbt.String("minecraft:sneaking_speed"),
							"id":        nbt.String("minecraft:enchantment.swift_sneak"),
							"operation": nbt.String("add_value"),
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(75),
					"per_level_above_first": nbt.Int(25),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(25),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("legs"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(1),
			}},
			{Name: "minecraft:thorns", Data: nbt.Compound{
				"anvil_cost": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.thorns"),
				},
				"effects": nbt.Compound{
					"minecraft:post_attack": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"affected": nbt.String("attacker"),
							"effect": nbt.Compound{
								"effects": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
									nbt.Compound{
										"damage_type": nbt.String("minecraft:thorns"),
										"max_damage":  nbt.Double(5.0),
										"min_damage":  nbt.Double(1.0),
										"type":        nbt.String("minecraft:damage_entity"),
									},
									nbt.Compound{
										"amount": nbt.Double(2.0),
										"type":   nbt.String("minecraft:change_item_damage"),
									},
								}},
								"type": nbt.String("minecraft:all_of"),
							},
							"enchanted": nbt.String("victim"),
							"requirements": nbt.Compound{
								"chance": nbt.Compound{
									"amount": nbt.Compound{
										"base":                  nbt.Double(0.15),
										"per_level_above_first": nbt.Double(0.15),
										"type":                  nbt.String("minecraft:linear"),
									},
									"type": nbt.String("minecraft:enchantment_level"),
								},
								"condition": nbt.String("minecraft:random_chance"),
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(60),
					"per_level_above_first": nbt.Int(20),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(10),
					"per_level_above_first": nbt.Int(20),
				},
				"primary_items": nbt.List{},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("any"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(1),
			}},
			{Name: "minecraft:unbreaking", Data: nbt.Compound{
				"anvil_cost": nbt.Int(2),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.unbreaking"),
				},
				"effects": nbt.Compound{
					"minecraft:item_damage": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"effect": nbt.Compound{
								"chance": nbt.Compound{
									"denominator": nbt.Compound{
										"base":                  nbt.Double(10.0),
										"per_level_above_first": nbt.Double(5.0),
										"type":                  nbt.String("minecraft:linear"),
									},
									"numerator": nbt.Compound{
										"base":                  nbt.Double(2.0),
										"per_level_above_first": nbt.Double(2.0),
										"type":                  nbt.String("minecraft:linear"),
									},
									"type": nbt.String("minecraft:fraction"),
								},
								"type": nbt.String("minecraft:remove_binomial"),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:match_tool"),
								"predicate": nbt.Compound{
									"items": nbt.List{},
								},
							},
						},
						nbt.Compound{
							"effect": nbt.Compound{
								"chance": nbt.Compound{
									"denominator": nbt.Compound{
										"base":                  nbt.Double(2.0),
										"per_level_above_first": nbt.Double(1.0),
										"type":                  nbt.String("minecraft:linear"),
									},
									"numerator": nbt.Compound{
										"base":                  nbt.Double(1.0),
										"per_level_above_first": nbt.Double(1.0),
										"type":                  nbt.String("minecraft:linear"),
									},
									"type": nbt.String("minecraft:fraction"),
								},
								"type": nbt.String("minecraft:remove_binomial"),
							},
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:inverted"),
								"term": nbt.Compound{
									"condition": nbt.String("minecraft:match_tool"),
									"predicate": nbt.Compound{
										"items": nbt.List{},
									},
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(55),
					"per_level_above_first": nbt.Int(8),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(5),
					"per_level_above_first": nbt.Int(8),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("any"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(5),
			}},
			{Name: "minecraft:vanishing_curse", Data: nbt.Compound{
				"anvil_cost": nbt.Int(8),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.vanishing_curse"),
				},
				"effects": nbt.Compound{
					"minecraft:prevent_equipment_drop": nbt.Compound{},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(50),
					"per_level_above_first": nbt.Int(0),
				},
				"max_level": nbt.Int(1),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(25),
					"per_level_above_first": nbt.Int(0),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("any"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(1),
			}},
			{Name: "minecraft:wind_burst", Data: nbt.Compound{
				"anvil_cost": nbt.Int(4),
				"description": nbt.Compound{
					"translate": nbt.String("enchantment.minecraft.wind_burst"),
				},
				"effects": nbt.Compound{
					"minecraft:post_attack": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
						nbt.Compound{
							"affected": nbt.String("attacker"),
							"effect": nbt.Compound{
								"block_interaction": nbt.String("trigger"),
								"immune_blocks":     nbt.List{},
								"knockback_multiplier": nbt.Compound{
									"fallback": nbt.Compound{
										"base":                  nbt.Double(1.5),
										"per_level_above_first": nbt.Double(0.35),
										"type":                  nbt.String("minecraft:linear"),
									},
									"type": nbt.String("minecraft:lookup"),
									"values": nbt.List{ElementType: nbt.TagDouble, Elements: []nbt.Tag{
										nbt.Double(1.2),
										nbt.Double(1.75),
										nbt.Double(2.2),
									}},
								},
								"large_particle": nbt.Compound{
									"type": nbt.String("minecraft:gust_emitter_large"),
								},
								"radius": nbt.Double(3.5),
								"small_particle": nbt.Compound{
									"type": nbt.String("minecraft:gust_emitter_small"),
								},
								"sound": nbt.String("minecraft:entity.wind_charge.wind_burst"),
								"type":  nbt.String("minecraft:explode"),
							},
							"enchanted": nbt.String("attacker"),
							"requirements": nbt.Compound{
								"condition": nbt.String("minecraft:entity_properties"),
								"entity":    nbt.String("direct_attacker"),
								"predicate": nbt.Compound{
									"minecraft:flags": nbt.Compound{
										"is_flying": nbt.Byte(0),
									},
									"minecraft:movement": nbt.Compound{
										"fall_distance": nbt.Compound{
											"min": nbt.Double(1.5),
										},
									},
								},
							},
						},
					}},
				},
				"max_cost": nbt.Compound{
					"base":                  nbt.Int(65),
					"per_level_above_first": nbt.Int(9),
				},
				"max_level": nbt.Int(3),
				"min_cost": nbt.Compound{
					"base":                  nbt.Int(15),
					"per_level_above_first": nbt.Int(9),
				},
				"slots": nbt.List{ElementType: nbt.TagString, Elements: []nbt.Tag{
					nbt.String("mainhand"),
				}},
				"supported_items": nbt.List{},
				"weight":          nbt.Int(2),
			}},
		}},
		{Name: "minecraft:sulfur_cube_archetype", Entries: []Entry{
			{Name: "minecraft:bouncy", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(-2.0),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:bouncy_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-2.0),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:bouncy_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.8999999761581421),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:bouncy_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.699999988079071),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:bouncy_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9900000002235174),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:bouncy_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"buoyant": nbt.Byte(1),
				"items":   nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.105),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.bouncy.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.bouncy.push"),
					"push_sound_cooldown":          nbt.Double(0.7),
					"push_sound_impulse_threshold": nbt.Double(0.3),
				},
			}},
			{Name: "minecraft:explosive", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:explosive_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:explosive_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.5),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:explosive_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.699999988079071),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:explosive_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.699999988079071),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:explosive_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"buoyant": nbt.Byte(1),
				"explosion": nbt.Compound{
					"causes_fire": nbt.Byte(0),
					"fuse":        nbt.Int(120),
					"power":       nbt.Int(3),
				},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.09),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.explosive.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.explosive.push"),
					"push_sound_cooldown":          nbt.Double(0.7),
					"push_sound_impulse_threshold": nbt.Double(0.1),
				},
			}},
			{Name: "minecraft:fast_flat", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:fast_flat_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:fast_flat_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.5),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:fast_flat_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.7999999970197678),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:fast_flat_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9900000002235174),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:fast_flat_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.9125),
					"vertical_power":   nbt.Double(0.09),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.fast_flat.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.fast_flat.push"),
					"push_sound_cooldown":          nbt.Double(0.9),
					"push_sound_impulse_threshold": nbt.Double(0.03),
				},
			}},
			{Name: "minecraft:fast_sliding", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(0.5),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:fast_sliding_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.5),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:fast_sliding_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.10000000149011612),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:fast_sliding_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9499999992549419),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:fast_sliding_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9900000002235174),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:fast_sliding_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.6625),
					"vertical_power":   nbt.Double(0.09),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.fast_sliding.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.fast_sliding.push"),
					"push_sound_cooldown":          nbt.Double(1.0),
					"push_sound_impulse_threshold": nbt.Double(0.05),
				},
			}},
			{Name: "minecraft:high_resistance", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(0.699999988079071),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:high_resistance_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.699999988079071),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:high_resistance_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.20000000298023224),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:high_resistance_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.0),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:high_resistance_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9900000002235174),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:high_resistance_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.09),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.high_resistance.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.high_resistance.push"),
					"push_sound_cooldown":          nbt.Double(0.7),
					"push_sound_impulse_threshold": nbt.Double(0.03),
				},
			}},
			{Name: "minecraft:hot", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:hot_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:hot_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.5),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:hot_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.699999988079071),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:hot_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.8999999985098839),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:hot_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"buoyant": nbt.Byte(1),
				"contact_damage": nbt.Compound{
					"amount":              nbt.Double(1.0),
					"attribute_to_source": nbt.Byte(0),
					"damage_type":         nbt.String("minecraft:sulfur_cube_hot"),
				},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.09),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.hot.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.hot.push"),
					"push_sound_cooldown":          nbt.Double(0.7),
					"push_sound_impulse_threshold": nbt.Double(0.2),
				},
			}},
			{Name: "minecraft:light", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:light_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:light_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(1.0),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:light_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.699999988079071),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:light_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.7999999523162842),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:light_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"buoyant": nbt.Byte(1),
				"items":   nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.18),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.light.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.light.push"),
					"push_sound_cooldown":          nbt.Double(0.7),
					"push_sound_impulse_threshold": nbt.Double(0.2),
				},
			}},
			{Name: "minecraft:regular", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:regular_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-1.0),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:regular_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.5),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:regular_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.699999988079071),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:regular_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.8999999985098839),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:regular_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"buoyant": nbt.Byte(1),
				"items":   nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.09),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.regular.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.regular.push"),
					"push_sound_cooldown":          nbt.Double(0.5),
					"push_sound_impulse_threshold": nbt.Double(0.2),
				},
			}},
			{Name: "minecraft:slow_bouncy", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(0.4000000059604645),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:slow_bouncy_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.4000000059604645),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:slow_bouncy_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.6000000238418579),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:slow_bouncy_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.699999988079071),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:slow_bouncy_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9499999992549419),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:slow_bouncy_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.24),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.slow_bouncy.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.slow_bouncy.push"),
					"push_sound_cooldown":          nbt.Double(0.5),
					"push_sound_impulse_threshold": nbt.Double(0.05),
				},
			}},
			{Name: "minecraft:slow_flat", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(0.5),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:slow_flat_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.5),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:slow_flat_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.4000000059604645),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:slow_flat_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.5999999940395355),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:slow_flat_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.8999999985098839),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:slow_flat_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.105),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.slow_flat.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.slow_flat.push"),
					"push_sound_cooldown":          nbt.Double(0.9),
					"push_sound_impulse_threshold": nbt.Double(0.03),
				},
			}},
			{Name: "minecraft:slow_sliding", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(0.800000011920929),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:slow_sliding_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.800000011920929),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:slow_sliding_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.10000000149011612),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:slow_sliding_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9499999992549419),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:slow_sliding_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9900000002235174),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:slow_sliding_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.09),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.slow_sliding.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.slow_sliding.push"),
					"push_sound_cooldown":          nbt.Double(1.0),
					"push_sound_impulse_threshold": nbt.Double(0.02),
				},
			}},
			{Name: "minecraft:sticky", Data: nbt.Compound{
				"attribute_modifiers": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
					nbt.Compound{
						"amount":    nbt.Double(-2.0),
						"attribute": nbt.String("minecraft:knockback_resistance"),
						"id":        nbt.String("minecraft:sticky_add_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-2.0),
						"attribute": nbt.String("minecraft:explosion_knockback_resistance"),
						"id":        nbt.String("minecraft:sticky_add_explosion_knockback_resistance"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(0.0),
						"attribute": nbt.String("minecraft:bounciness"),
						"id":        nbt.String("minecraft:sticky_add_bounciness"),
						"operation": nbt.String("add_value"),
					},
					nbt.Compound{
						"amount":    nbt.Double(1.0),
						"attribute": nbt.String("minecraft:friction_modifier"),
						"id":        nbt.String("minecraft:sticky_mul_friction_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
					nbt.Compound{
						"amount":    nbt.Double(-0.9900000002235174),
						"attribute": nbt.String("minecraft:air_drag_modifier"),
						"id":        nbt.String("minecraft:sticky_mul_air_drag_modifier"),
						"operation": nbt.String("add_multiplied_total"),
					},
				}},
				"items": nbt.List{},
				"knockback_modifiers": nbt.Compound{
					"horizontal_power": nbt.Double(0.4125),
					"vertical_power":   nbt.Double(0.09),
				},
				"sound_settings": nbt.Compound{
					"hit_sound":                    nbt.String("minecraft:entity.sulfur_cube.sticky.hit"),
					"push_sound":                   nbt.String("minecraft:entity.sulfur_cube.sticky.push"),
					"push_sound_cooldown":          nbt.Double(0.5),
					"push_sound_impulse_threshold": nbt.Double(0.05),
				},
			}},
		}},
		{Name: "minecraft:dialog", Entries: []Entry{
			{Name: "minecraft:custom_options", Data: nbt.Compound{
				"button_width": nbt.Int(310),
				"columns":      nbt.Int(1),
				"dialogs":      nbt.List{},
				"exit_action": nbt.Compound{
					"label": nbt.Compound{
						"translate": nbt.String("gui.back"),
					},
					"width": nbt.Int(200),
				},
				"external_title": nbt.Compound{
					"translate": nbt.String("menu.custom_options"),
				},
				"title": nbt.Compound{
					"translate": nbt.String("menu.custom_options.title"),
				},
				"type": nbt.String("minecraft:dialog_list"),
			}},
			{Name: "minecraft:quick_actions", Data: nbt.Compound{
				"button_width": nbt.Int(310),
				"columns":      nbt.Int(1),
				"dialogs":      nbt.List{},
				"exit_action": nbt.Compound{
					"label": nbt.Compound{
						"translate": nbt.String("gui.back"),
					},
					"width": nbt.Int(200),
				},
				"external_title": nbt.Compound{
					"translate": nbt.String("menu.quick_actions"),
				},
				"title": nbt.Compound{
					"translate": nbt.String("menu.quick_actions.title"),
				},
				"type": nbt.String("minecraft:dialog_list"),
			}},
			{Name: "minecraft:server_links", Data: nbt.Compound{
				"button_width": nbt.Int(310),
				"columns":      nbt.Int(1),
				"exit_action": nbt.Compound{
					"label": nbt.Compound{
						"translate": nbt.String("gui.back"),
					},
					"width": nbt.Int(200),
				},
				"external_title": nbt.Compound{
					"translate": nbt.String("menu.server_links"),
				},
				"title": nbt.Compound{
					"translate": nbt.String("menu.server_links.title"),
				},
				"type": nbt.String("minecraft:server_links"),
			}},
		}},
		{Name: "minecraft:world_clock", Entries: []Entry{
			{Name: "minecraft:overworld", Data: nbt.Compound{}},
			{Name: "minecraft:the_end", Data: nbt.Compound{}},
		}},
		{Name: "minecraft:timeline", Entries: []Entry{
			{Name: "minecraft:day", Data: nbt.Compound{
				"clock":        nbt.String("minecraft:overworld"),
				"period_ticks": nbt.Int(24000),
				"time_markers": nbt.Compound{
					"minecraft:day": nbt.Compound{
						"show_in_commands": nbt.Byte(1),
						"ticks":            nbt.Int(1000),
					},
					"minecraft:midnight": nbt.Compound{
						"show_in_commands": nbt.Byte(1),
						"ticks":            nbt.Int(18000),
					},
					"minecraft:night": nbt.Compound{
						"show_in_commands": nbt.Byte(1),
						"ticks":            nbt.Int(13000),
					},
					"minecraft:noon": nbt.Compound{
						"show_in_commands": nbt.Byte(1),
						"ticks":            nbt.Int(6000),
					},
					"minecraft:roll_village_siege": nbt.Int(18000),
					"minecraft:wake_up_from_sleep": nbt.Int(0),
				},
				"tracks": nbt.Compound{
					"minecraft:audio/firefly_bush_sounds": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(12600),
								"value": nbt.Byte(1),
							},
							nbt.Compound{
								"ticks": nbt.Int(23401),
								"value": nbt.Byte(0),
							},
						}},
						"modifier": nbt.String("or"),
					},
					"minecraft:gameplay/bees_stay_in_hive": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(12542),
								"value": nbt.Byte(1),
							},
							nbt.Compound{
								"ticks": nbt.Int(23460),
								"value": nbt.Byte(0),
							},
						}},
						"modifier": nbt.String("or"),
					},
					"minecraft:gameplay/cat_waking_up_gift_chance": nbt.Compound{
						"ease": nbt.String("constant"),
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(362),
								"value": nbt.Double(0.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(23667),
								"value": nbt.Double(0.7),
							},
						}},
						"modifier": nbt.String("maximum"),
					},
					"minecraft:gameplay/creaking_active": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(12600),
								"value": nbt.Byte(1),
							},
							nbt.Compound{
								"ticks": nbt.Int(23401),
								"value": nbt.Byte(0),
							},
						}},
						"modifier": nbt.String("or"),
					},
					"minecraft:gameplay/eyeblossom_open": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(12600),
								"value": nbt.Byte(1),
							},
							nbt.Compound{
								"ticks": nbt.Int(23401),
								"value": nbt.Byte(0),
							},
						}},
					},
					"minecraft:gameplay/monsters_burn": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(12542),
								"value": nbt.Byte(0),
							},
							nbt.Compound{
								"ticks": nbt.Int(23460),
								"value": nbt.Byte(1),
							},
						}},
						"modifier": nbt.String("or"),
					},
					"minecraft:gameplay/sky_light_level": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(133),
								"value": nbt.Double(1.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(11867),
								"value": nbt.Double(1.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(13670),
								"value": nbt.Double(0.26666668),
							},
							nbt.Compound{
								"ticks": nbt.Int(22330),
								"value": nbt.Double(0.26666668),
							},
						}},
						"modifier": nbt.String("multiply"),
					},
					"minecraft:gameplay/turtle_egg_hatch_chance": nbt.Compound{
						"ease": nbt.String("constant"),
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(21062),
								"value": nbt.Double(1.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(21905),
								"value": nbt.Double(0.002),
							},
						}},
						"modifier": nbt.String("maximum"),
					},
					"minecraft:visual/cloud_color": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(133),
								"value": nbt.Int(-1),
							},
							nbt.Compound{
								"ticks": nbt.Int(11867),
								"value": nbt.Int(-1),
							},
							nbt.Compound{
								"ticks": nbt.Int(13670),
								"value": nbt.Int(-15132378),
							},
							nbt.Compound{
								"ticks": nbt.Int(22330),
								"value": nbt.Int(-15132378),
							},
						}},
						"modifier": nbt.String("multiply"),
					},
					"minecraft:visual/fog_color": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(133),
								"value": nbt.String("#ffffff"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11867),
								"value": nbt.String("#ffffff"),
							},
							nbt.Compound{
								"ticks": nbt.Int(13670),
								"value": nbt.String("#0c0c16"),
							},
							nbt.Compound{
								"ticks": nbt.Int(22330),
								"value": nbt.String("#161616"),
							},
						}},
						"modifier": nbt.String("multiply"),
					},
					"minecraft:visual/moon_angle": nbt.Compound{
						"ease": nbt.Compound{
							"cubic_bezier": nbt.List{ElementType: nbt.TagDouble, Elements: []nbt.Tag{
								nbt.Double(0.362),
								nbt.Double(0.241),
								nbt.Double(0.638),
								nbt.Double(0.759),
							}},
						},
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(6000),
								"value": nbt.Double(540.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(6000),
								"value": nbt.Double(180.0),
							},
						}},
					},
					"minecraft:visual/sky_color": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(133),
								"value": nbt.String("#ffffff"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11867),
								"value": nbt.String("#ffffff"),
							},
							nbt.Compound{
								"ticks": nbt.Int(13670),
								"value": nbt.String("#000000"),
							},
							nbt.Compound{
								"ticks": nbt.Int(22330),
								"value": nbt.String("#000000"),
							},
						}},
						"modifier": nbt.String("multiply"),
					},
					"minecraft:visual/sky_light_color": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(730),
								"value": nbt.String("#ffffff"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11270),
								"value": nbt.String("#ffffff"),
							},
							nbt.Compound{
								"ticks": nbt.Int(13140),
								"value": nbt.String("#7a7aff"),
							},
							nbt.Compound{
								"ticks": nbt.Int(22860),
								"value": nbt.String("#7a7aff"),
							},
						}},
						"modifier": nbt.String("multiply"),
					},
					"minecraft:visual/sky_light_factor": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(730),
								"value": nbt.Double(1.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(11270),
								"value": nbt.Double(1.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(13140),
								"value": nbt.Double(0.24),
							},
							nbt.Compound{
								"ticks": nbt.Int(22860),
								"value": nbt.Double(0.24),
							},
						}},
						"modifier": nbt.String("multiply"),
					},
					"minecraft:visual/star_angle": nbt.Compound{
						"ease": nbt.Compound{
							"cubic_bezier": nbt.List{ElementType: nbt.TagDouble, Elements: []nbt.Tag{
								nbt.Double(0.362),
								nbt.Double(0.241),
								nbt.Double(0.638),
								nbt.Double(0.759),
							}},
						},
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(6000),
								"value": nbt.Double(360.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(6000),
								"value": nbt.Double(0.0),
							},
						}},
					},
					"minecraft:visual/star_brightness": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(92),
								"value": nbt.Double(0.037),
							},
							nbt.Compound{
								"ticks": nbt.Int(627),
								"value": nbt.Double(0.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(11373),
								"value": nbt.Double(0.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(11732),
								"value": nbt.Double(0.016),
							},
							nbt.Compound{
								"ticks": nbt.Int(11959),
								"value": nbt.Double(0.044),
							},
							nbt.Compound{
								"ticks": nbt.Int(12399),
								"value": nbt.Double(0.143),
							},
							nbt.Compound{
								"ticks": nbt.Int(12729),
								"value": nbt.Double(0.258),
							},
							nbt.Compound{
								"ticks": nbt.Int(13228),
								"value": nbt.Double(0.5),
							},
							nbt.Compound{
								"ticks": nbt.Int(22772),
								"value": nbt.Double(0.5),
							},
							nbt.Compound{
								"ticks": nbt.Int(23032),
								"value": nbt.Double(0.364),
							},
							nbt.Compound{
								"ticks": nbt.Int(23356),
								"value": nbt.Double(0.225),
							},
							nbt.Compound{
								"ticks": nbt.Int(23758),
								"value": nbt.Double(0.101),
							},
						}},
						"modifier": nbt.String("maximum"),
					},
					"minecraft:visual/sun_angle": nbt.Compound{
						"ease": nbt.Compound{
							"cubic_bezier": nbt.List{ElementType: nbt.TagDouble, Elements: []nbt.Tag{
								nbt.Double(0.362),
								nbt.Double(0.241),
								nbt.Double(0.638),
								nbt.Double(0.759),
							}},
						},
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(6000),
								"value": nbt.Double(360.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(6000),
								"value": nbt.Double(0.0),
							},
						}},
					},
					"minecraft:visual/sunrise_sunset_color": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(71),
								"value": nbt.String("#5fefa333"),
							},
							nbt.Compound{
								"ticks": nbt.Int(310),
								"value": nbt.String("#29f5ba33"),
							},
							nbt.Compound{
								"ticks": nbt.Int(565),
								"value": nbt.String("#06fbd433"),
							},
							nbt.Compound{
								"ticks": nbt.Int(730),
								"value": nbt.String("#00ffe533"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11270),
								"value": nbt.String("#00ffe533"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11397),
								"value": nbt.String("#04fcd833"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11522),
								"value": nbt.String("#0ff9cb33"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11690),
								"value": nbt.String("#29f5ba33"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11929),
								"value": nbt.String("#5fefa333"),
							},
							nbt.Compound{
								"ticks": nbt.Int(12243),
								"value": nbt.String("#b1e78733"),
							},
							nbt.Compound{
								"ticks": nbt.Int(12358),
								"value": nbt.String("#cce47e33"),
							},
							nbt.Compound{
								"ticks": nbt.Int(12512),
								"value": nbt.String("#e9e07233"),
							},
							nbt.Compound{
								"ticks": nbt.Int(12613),
								"value": nbt.String("#f6dd6b33"),
							},
							nbt.Compound{
								"ticks": nbt.Int(12732),
								"value": nbt.String("#feda6333"),
							},
							nbt.Compound{
								"ticks": nbt.Int(12841),
								"value": nbt.String("#fed75c33"),
							},
							nbt.Compound{
								"ticks": nbt.Int(13035),
								"value": nbt.String("#ecd25133"),
							},
							nbt.Compound{
								"ticks": nbt.Int(13252),
								"value": nbt.String("#c1cc4733"),
							},
							nbt.Compound{
								"ticks": nbt.Int(13775),
								"value": nbt.String("#36be3733"),
							},
							nbt.Compound{
								"ticks": nbt.Int(13888),
								"value": nbt.String("#1fbb3533"),
							},
							nbt.Compound{
								"ticks": nbt.Int(14039),
								"value": nbt.String("#09b73333"),
							},
							nbt.Compound{
								"ticks": nbt.Int(14192),
								"value": nbt.String("#00b33333"),
							},
							nbt.Compound{
								"ticks": nbt.Int(21807),
								"value": nbt.String("#00b23333"),
							},
							nbt.Compound{
								"ticks": nbt.Int(21961),
								"value": nbt.String("#09b73333"),
							},
							nbt.Compound{
								"ticks": nbt.Int(22112),
								"value": nbt.String("#1fbb3533"),
							},
							nbt.Compound{
								"ticks": nbt.Int(22225),
								"value": nbt.String("#36be3733"),
							},
							nbt.Compound{
								"ticks": nbt.Int(22748),
								"value": nbt.String("#c1cc4733"),
							},
							nbt.Compound{
								"ticks": nbt.Int(22965),
								"value": nbt.String("#ecd25133"),
							},
							nbt.Compound{
								"ticks": nbt.Int(23159),
								"value": nbt.String("#fed75c33"),
							},
							nbt.Compound{
								"ticks": nbt.Int(23272),
								"value": nbt.String("#feda6333"),
							},
							nbt.Compound{
								"ticks": nbt.Int(23488),
								"value": nbt.String("#e9e07233"),
							},
							nbt.Compound{
								"ticks": nbt.Int(23642),
								"value": nbt.String("#cce47e33"),
							},
							nbt.Compound{
								"ticks": nbt.Int(23757),
								"value": nbt.String("#b1e78733"),
							},
						}},
					},
				},
			}},
			{Name: "minecraft:early_game", Data: nbt.Compound{
				"clock": nbt.String("minecraft:overworld"),
				"tracks": nbt.Compound{
					"minecraft:gameplay/can_pillager_patrol_spawn": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(0),
								"value": nbt.Byte(0),
							},
							nbt.Compound{
								"ticks": nbt.Int(120000),
								"value": nbt.Byte(1),
							},
						}},
						"modifier": nbt.String("and"),
					},
				},
			}},
			{Name: "minecraft:moon", Data: nbt.Compound{
				"clock":        nbt.String("minecraft:overworld"),
				"period_ticks": nbt.Int(192000),
				"tracks": nbt.Compound{
					"minecraft:gameplay/surface_slime_spawn_chance": nbt.Compound{
						"ease": nbt.String("constant"),
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(0),
								"value": nbt.Double(0.5),
							},
							nbt.Compound{
								"ticks": nbt.Int(24000),
								"value": nbt.Double(0.375),
							},
							nbt.Compound{
								"ticks": nbt.Int(48000),
								"value": nbt.Double(0.25),
							},
							nbt.Compound{
								"ticks": nbt.Int(72000),
								"value": nbt.Double(0.125),
							},
							nbt.Compound{
								"ticks": nbt.Int(96000),
								"value": nbt.Double(0.0),
							},
							nbt.Compound{
								"ticks": nbt.Int(120000),
								"value": nbt.Double(0.125),
							},
							nbt.Compound{
								"ticks": nbt.Int(144000),
								"value": nbt.Double(0.25),
							},
							nbt.Compound{
								"ticks": nbt.Int(168000),
								"value": nbt.Double(0.375),
							},
						}},
						"modifier": nbt.String("maximum"),
					},
					"minecraft:visual/moon_phase": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(0),
								"value": nbt.String("full_moon"),
							},
							nbt.Compound{
								"ticks": nbt.Int(24000),
								"value": nbt.String("waning_gibbous"),
							},
							nbt.Compound{
								"ticks": nbt.Int(48000),
								"value": nbt.String("third_quarter"),
							},
							nbt.Compound{
								"ticks": nbt.Int(72000),
								"value": nbt.String("waning_crescent"),
							},
							nbt.Compound{
								"ticks": nbt.Int(96000),
								"value": nbt.String("new_moon"),
							},
							nbt.Compound{
								"ticks": nbt.Int(120000),
								"value": nbt.String("waxing_crescent"),
							},
							nbt.Compound{
								"ticks": nbt.Int(144000),
								"value": nbt.String("first_quarter"),
							},
							nbt.Compound{
								"ticks": nbt.Int(168000),
								"value": nbt.String("waxing_gibbous"),
							},
						}},
					},
				},
			}},
			{Name: "minecraft:villager_schedule", Data: nbt.Compound{
				"clock":        nbt.String("minecraft:overworld"),
				"period_ticks": nbt.Int(24000),
				"tracks": nbt.Compound{
					"minecraft:gameplay/baby_villager_activity": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(10),
								"value": nbt.String("minecraft:idle"),
							},
							nbt.Compound{
								"ticks": nbt.Int(3000),
								"value": nbt.String("minecraft:play"),
							},
							nbt.Compound{
								"ticks": nbt.Int(6000),
								"value": nbt.String("minecraft:idle"),
							},
							nbt.Compound{
								"ticks": nbt.Int(10000),
								"value": nbt.String("minecraft:play"),
							},
							nbt.Compound{
								"ticks": nbt.Int(12000),
								"value": nbt.String("minecraft:rest"),
							},
						}},
					},
					"minecraft:gameplay/villager_activity": nbt.Compound{
						"keyframes": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{
								"ticks": nbt.Int(10),
								"value": nbt.String("minecraft:idle"),
							},
							nbt.Compound{
								"ticks": nbt.Int(2000),
								"value": nbt.String("minecraft:work"),
							},
							nbt.Compound{
								"ticks": nbt.Int(9000),
								"value": nbt.String("minecraft:meet"),
							},
							nbt.Compound{
								"ticks": nbt.Int(11000),
								"value": nbt.String("minecraft:idle"),
							},
							nbt.Compound{
								"ticks": nbt.Int(12000),
								"value": nbt.String("minecraft:rest"),
							},
						}},
					},
				},
			}},
		}},
		{Name: "minecraft:test_environment", Entries: []Entry{
			{Name: "minecraft:default", Data: nbt.Compound{
				"definitions": nbt.List{},
				"type":        nbt.String("minecraft:all_of"),
			}},
		}},
		{Name: "minecraft:test_instance", Entries: []Entry{
			{Name: "minecraft:always_pass", Data: nbt.Compound{
				"environment": nbt.String("minecraft:default"),
				"function":    nbt.String("minecraft:always_pass"),
				"max_ticks":   nbt.Int(1),
				"required":    nbt.Byte(0),
				"setup_ticks": nbt.Int(1),
				"structure":   nbt.String("minecraft:empty"),
				"type":        nbt.String("minecraft:function"),
			}},
		}},
	}
}
