package gamedata

import (
	"encoding/json"
	"fmt"

	"github.com/Shonz1/go-void-limbo/types"
)

// A chunk packet names blocks by number: the index of the exact block state --
// a block together with every property it holds -- in the client's own block
// state registry. That registry is compiled into the client and is never sent,
// so a server that reads a world of names has to hold the same numbering to
// translate with, and hold it per version, because every block a version adds
// shifts the numbers of everything registered after it.
//
// The tables live in data/blockstates_minecraft_*.json, generated from the
// blocks report of each version's own data generator the same way the registry
// files are. They lean on how the client numbers states rather than listing
// every state: a block's states are numbered contiguously from a base, in
// row-major order over its properties -- the last property varying fastest --
// so one entry per block with its properties in that order reproduces every id
// the report holds. The generation verifies that reproduction against the full
// report, so a table that loads is a table that numbers exactly as the client
// does.

// BlockStates is one version's numbering of every block state, as
// BlockStatesFor loads it.
type BlockStates struct {
	table *blockStatesTable

	// renames is every block this version numbers under an older name than
	// the one a newer world stores it by, as blockStateRenames spells it.
	// It sits beside the table rather than in it because the table may be
	// shared with a version that has no such rename.
	renames map[string]string

	// valueRenames is likewise every property value this version knows by
	// an older name, as blockStateValueRenames spells it.
	valueRenames map[string]string

	// ids is, for a version from before the flattening, the number the
	// client knows each of the table's states by, as blockIdsFiles names
	// it; nil for every other version, whose table numbers the states the
	// way the client does itself.
	ids *blockIdsTable
}

// maxLegacyBlockId is the largest number a version from before the
// flattening knows a block by: block 255, variant 15.
const maxLegacyBlockId = 255<<4 | 15

// blockIdsTable is one parsed block id file: for every state of the table it
// sits beside, the number a version from before the flattening knows it by.
type blockIdsTable struct {
	ids []int32

	// stateCount is how many states the version's client holds in all,
	// which sizes an id as a table's own count does.
	stateCount int32
}

// blockStatesTable is one parsed table: what a block state file holds, which
// several versions may share.
type blockStatesTable struct {
	blocks map[string]*blockStatesEntry

	// stateCount is how many states the version numbers in all, which is what
	// sizes an id: a palette that names states directly packs them at however
	// many bits the largest id needs.
	stateCount int32
}

type blockStatesEntry struct {
	base       int32
	defaultOff int32
	properties []blockStateProperty
}

type blockStateProperty struct {
	name   string
	values []string
}

// blockStatesFiles is the table each version loads, keyed the way the registry
// data files are.
//
// 26.2 carries its own table below 26.3's: 777 is where the poplar wood
// set, the wool and concrete slabs and stairs of every colour, the red
// shrub, the shelf mushroom and the straw bed landed, ninety blocks in all
// with no property changed on any older one, so 26.2 numbers 32,366 states
// to 26.3's 35,723. 1.21.9 names the 1.21.11 file rather than a copy because the two versions
// number every state identically: 774 added no block and reordered nothing,
// which the generation checked by producing both tables from the two jars'
// own reports and comparing them byte for byte. 1.21.7 gets no such sharing:
// 773 is where the copper additions landed, so 772 numbers 27,946 states to
// 773's 29,671 and carries its own table. 1.21.6 shares that table the way
// 1.21.9 shares 1.21.11's: 772 added no block -- its jar's blocks report is
// byte-identical to 771's -- so the two versions number every state alike.
// 1.21.5 carries its own table again: 771 is where the dried ghast landed,
// so 770 numbers 27,914 states to 771's 27,946. And 1.21.4 its own below
// that: 770 is where the spring vegetation and the test blocks landed, nine
// blocks in all, so 769 numbers 27,866 states. And 1.21.2 its own again:
// 769 is where the resin blocks and the eyeblossoms landed, eleven blocks in
// all, so 768 numbers 27,318 states. And 1.21 its own at the bottom: 768 is
// where the pale oak wood set, the pale moss and the creaking heart landed,
// twenty-four blocks in all, so 767 numbers 26,684 states. 1.20.5 shares that
// table the way 1.21.9 shares 1.21.11's: 1.21 added no block, and the two
// jars' blocks reports are byte-identical, so 766 numbers every state as 767
// does. And 1.20.3 its own below that: 766 is where the vault and the heavy
// core landed, two blocks in all, so 765 numbers 26,644 states. And 1.20.2
// its own below that: 765 is where the crafter, the trial spawner, the
// copper and tuff sets landed and where grass became short grass, fifty-six
// blocks in for the one name gone, so 764 numbers 24,276 states. And 1.20
// its own below that, with the same blocks as 1.20.2 and fewer states: 764
// is where the heads and skulls gained their powered property and the
// barrier its waterlogged one, so 763 numbers 24,135 states. And 1.19.4 its
// own below that: 763 is where the calibrated sculk sensor, the pitcher
// plant and its crop, the sniffer egg and the suspicious gravel landed, five
// blocks in all, and where the decorated pot gained its cracked property and
// the torchflower crop lost one of its three ages, so 762 numbers 23,725
// states. And 1.19.3 its own below that: 762 is where the cherry wood set,
// the pink petals, the torchflower and its crop, the decorated pot and the
// suspicious sand landed, twenty-six blocks in all, so 761 numbers 23,232
// states. And 1.19.1 its own below that: 761 is where the bamboo
// wood set, the hanging signs of every wood, the chiseled bookshelf and the
// piglin heads landed, thirty-nine blocks in all, and where the note block
// gained its seven mob head instruments, so 760 numbers 21,448 states. 1.19
// shares that table the way 1.21.9 shares 1.21.11's: 1.19.1 added no block
// and no property -- the two jars' blocks reports produce byte-identical
// tables -- so 759 numbers every state as 760 does. And 1.18.2 its own
// below the two: 759 is where the mangrove wood set, the mud and its
// bricks, the sculk set, the froglights, the frogspawn and the reinforced
// deepslate landed, thirty-five blocks in all, and where the leaves of
// every wood gained their waterlogged property, so 758 numbers 20,342
// states. 1.18 shares that table the way 1.21.9 shares 1.21.11's: 1.18.2
// added no block and no property -- the two jars' blocks reports produce
// byte-identical tables -- so 757 numbers every state as 758 does. And so
// does 1.17.1 below it: 1.18 raised the world and added no block to it, its
// jar's report byte-identical to the two above, so 756 numbers every state
// as 758 does as well, and 755 with it: 1.17.1 changed no block. And 1.16.4
// its own at the bottom: 755 is where the caves and cliffs blocks landed --
// the copper, the deepslate, the amethyst, the dripstone, the candles, the
// azalea and the rest, 135 blocks in all -- where the cauldron split by what
// it holds and the grass path became the dirt path, so 754 numbers 17,112
// states, and 753 every one of them alike: 1.16.4 changed no block, its
// jar's report byte-identical to 1.16.3's, and 1.16.3's to 1.16.2's, so 751
// numbers them the same way. 1.16.1 has its own below them with the same 763
// blocks: 751 is where the chain took on its axis and the two lanterns
// their waterlogging, eight states in all, so 736 numbers 17,104 and every
// block registered after the chain sits lower. A chain stored with an axis
// or a lantern with water in it is numbered by the properties 736 knows,
// the way any property a table does not hold is passed over. 1.16.1 changed
// no block, its jar's report byte-identical to 1.16's, so 735 numbers them
// as 736 does. And 1.15.2 its own below everything: 735 is where the nether
// update landed -- the crimson and warped sets, the blackstone, the basalt,
// the soul fire and the rest, eighty-three blocks in all -- and where a
// wall's sides went from being there or not to being low or tall, so 578
// numbers 11,337 states; see blockStateValueRenames for the walls. 1.15.2
// changed no block, its jar's report byte-identical to 1.15.1's, and 1.15.1's
// to 1.15's, so 575 and 573 number them as 578 does. And 1.14.4 its own
// again: 573 is where the bees landed -- their nest and their hive, the honey
// block and the honeycomb block -- and where the bell took on being powered,
// so 498 numbers 11,271 states. 1.14.4 changed no block, its jar's report
// byte-identical to 1.14.3's, so 490 numbers them as 498 does, and 1.14.3
// none either, so 485 does as well, nor 1.14.2, so 480 does too, nor 1.14.1,
// so 477 does as well. And 1.13.2 its own at the very bottom: 477 is where
// the village and pillage blocks landed -- the signs of every wood, the
// bamboo, the workstations, the bell, the lantern, the campfire, the
// scaffolding, the new slabs, stairs and walls of stone and the rest,
// eighty-one blocks in all -- and where the note block took on its six
// instruments past the xylophone, so 404 numbers 8,599 states; see
// blockStateRenames for the three blocks 1.14 renamed. 1.13.2 changed no
// block, its jar's report byte-identical to 1.13.1's, so 401 numbers them as
// 404 does. And 1.13 its own below that: 401 is where the dead coral plants
// landed, five blocks, and where the corals and the conduit took on being
// waterlogged and the TNT being unstable, so 393 numbers 8,582 states. A
// stored state's property 1.13's block lacks is passed over, as for any
// version, so a world's TNT and corals resolve on 393 as well. 1.12.2 names
// 1.13's table too, and goes through it to a number of its own, and 1.12.1,
// 1.12, 1.11.1, 1.11, 1.10 and 1.9.3 with it: see blockIdsFiles.
var blockStatesFiles = map[types.ProtocolId]string{
	types.ProtocolVersions.MINECRAFT_1_9_3.ID:   "blockstates_minecraft_1_13.json",
	types.ProtocolVersions.MINECRAFT_1_10.ID:    "blockstates_minecraft_1_13.json",
	types.ProtocolVersions.MINECRAFT_1_11.ID:    "blockstates_minecraft_1_13.json",
	types.ProtocolVersions.MINECRAFT_1_11_1.ID:  "blockstates_minecraft_1_13.json",
	types.ProtocolVersions.MINECRAFT_1_12.ID:    "blockstates_minecraft_1_13.json",
	types.ProtocolVersions.MINECRAFT_1_12_1.ID:  "blockstates_minecraft_1_13.json",
	types.ProtocolVersions.MINECRAFT_1_12_2.ID:  "blockstates_minecraft_1_13.json",
	types.ProtocolVersions.MINECRAFT_1_13.ID:    "blockstates_minecraft_1_13.json",
	types.ProtocolVersions.MINECRAFT_1_13_1.ID:  "blockstates_minecraft_1_13_2.json",
	types.ProtocolVersions.MINECRAFT_1_13_2.ID:  "blockstates_minecraft_1_13_2.json",
	types.ProtocolVersions.MINECRAFT_1_14.ID:    "blockstates_minecraft_1_14_4.json",
	types.ProtocolVersions.MINECRAFT_1_14_1.ID:  "blockstates_minecraft_1_14_4.json",
	types.ProtocolVersions.MINECRAFT_1_14_2.ID:  "blockstates_minecraft_1_14_4.json",
	types.ProtocolVersions.MINECRAFT_1_14_3.ID:  "blockstates_minecraft_1_14_4.json",
	types.ProtocolVersions.MINECRAFT_1_14_4.ID:  "blockstates_minecraft_1_14_4.json",
	types.ProtocolVersions.MINECRAFT_1_15.ID:    "blockstates_minecraft_1_15_2.json",
	types.ProtocolVersions.MINECRAFT_1_15_1.ID:  "blockstates_minecraft_1_15_2.json",
	types.ProtocolVersions.MINECRAFT_1_15_2.ID:  "blockstates_minecraft_1_15_2.json",
	types.ProtocolVersions.MINECRAFT_1_16.ID:    "blockstates_minecraft_1_16_1.json",
	types.ProtocolVersions.MINECRAFT_1_16_1.ID:  "blockstates_minecraft_1_16_1.json",
	types.ProtocolVersions.MINECRAFT_1_16_2.ID:  "blockstates_minecraft_1_16_4.json",
	types.ProtocolVersions.MINECRAFT_1_16_3.ID:  "blockstates_minecraft_1_16_4.json",
	types.ProtocolVersions.MINECRAFT_1_16_4.ID:  "blockstates_minecraft_1_16_4.json",
	types.ProtocolVersions.MINECRAFT_1_17.ID:    "blockstates_minecraft_1_18_2.json",
	types.ProtocolVersions.MINECRAFT_1_17_1.ID:  "blockstates_minecraft_1_18_2.json",
	types.ProtocolVersions.MINECRAFT_1_18.ID:    "blockstates_minecraft_1_18_2.json",
	types.ProtocolVersions.MINECRAFT_1_18_2.ID:  "blockstates_minecraft_1_18_2.json",
	types.ProtocolVersions.MINECRAFT_1_19.ID:    "blockstates_minecraft_1_19_1.json",
	types.ProtocolVersions.MINECRAFT_1_19_1.ID:  "blockstates_minecraft_1_19_1.json",
	types.ProtocolVersions.MINECRAFT_1_19_3.ID:  "blockstates_minecraft_1_19_3.json",
	types.ProtocolVersions.MINECRAFT_1_19_4.ID:  "blockstates_minecraft_1_19_4.json",
	types.ProtocolVersions.MINECRAFT_1_20.ID:    "blockstates_minecraft_1_20.json",
	types.ProtocolVersions.MINECRAFT_1_20_2.ID:  "blockstates_minecraft_1_20_2.json",
	types.ProtocolVersions.MINECRAFT_1_20_3.ID:  "blockstates_minecraft_1_20_3.json",
	types.ProtocolVersions.MINECRAFT_1_20_5.ID:  "blockstates_minecraft_1_21.json",
	types.ProtocolVersions.MINECRAFT_1_21.ID:    "blockstates_minecraft_1_21.json",
	types.ProtocolVersions.MINECRAFT_1_21_2.ID:  "blockstates_minecraft_1_21_2.json",
	types.ProtocolVersions.MINECRAFT_1_21_4.ID:  "blockstates_minecraft_1_21_4.json",
	types.ProtocolVersions.MINECRAFT_1_21_5.ID:  "blockstates_minecraft_1_21_5.json",
	types.ProtocolVersions.MINECRAFT_1_21_6.ID:  "blockstates_minecraft_1_21_7.json",
	types.ProtocolVersions.MINECRAFT_1_21_7.ID:  "blockstates_minecraft_1_21_7.json",
	types.ProtocolVersions.MINECRAFT_1_21_9.ID:  "blockstates_minecraft_1_21_11.json",
	types.ProtocolVersions.MINECRAFT_1_21_11.ID: "blockstates_minecraft_1_21_11.json",
	types.ProtocolVersions.MINECRAFT_26_1.ID:    "blockstates_minecraft_26_1.json",
	types.ProtocolVersions.MINECRAFT_26_2.ID:    "blockstates_minecraft_26_2.json",
	types.ProtocolVersions.MINECRAFT_26_3.ID:    "blockstates_minecraft_26_3.json",
}

// blockIdsFiles is the block id file each version from before the
// flattening loads beside its table: 1.12.2's, which knows a block by
// its number, shifted up four bits, and a variant in the four below, and
// numbers none of 1.13's states, and 1.11.1's below it, which is 1.12.2's
// less what 1.12 added, and which 1.11 shares, and 1.10's below that, which
// is 1.11.1's less what 1.11 added, and 1.9.3's below that, which is 1.10's
// less what 1.10 added. The file maps each of 1.13's states to the
// number 1.12.2 knows it by, and is how 1.12.2 reads a world: a stored state
// is found in 1.13's table, under 1.13's renames, and its number there looked
// up in the file.
//
// The numbers come from 1.13's own data fixers, which read a world saved by
// 1.12.2 and so know every block number and variant 1.12.2 held and the state
// 1.13 made of each. Each number the flattening lists for a state is the
// state's number here, and a state it does not list -- one 1.12.2 worked out
// from the blocks around it rather than keep, a fence's sides or a stair's
// shape among them -- takes the number whose kept properties it shares. The
// numbers are checked against the ones 1.12.2 itself registers, and every
// number maps back from the state it became. Where two of 1.12.2's numbers
// became one state -- the leaves that were waiting to decay and the ones that
// were not, a comparator lit and unlit -- the lower is the state's, but for
// the water and the lava, whose still blocks are what a 1.12.2 world holds
// rather than the flowing ones below them. A few of 1.13's blocks are drawn
// from 1.12.2's block entities rather than its numbers -- a bed's colour, a
// banner's, a skull's kind, a pot's plant -- and take the number the block
// has without one; the woods' own buttons, plates and trapdoors and their
// stripped logs, which 1.13 added, take the oak's and the unstripped log's;
// the blue ice takes the packed ice; and the plants that grow only
// underwater take the water they stand in. The rest of what 1.13 added --
// the corals, the dried kelp block, the conduit, the prismarine slabs and
// stairs, the sea pickle and the turtle egg -- is -1, which is no number at
// all, and a world's block of one of those is substituted as any version's
// unknown block is.
//
// The count beside the numbers is how many states 1.12.2 registers, which is
// what sizes one of its ids: a section too varied for a palette packs them at
// thirteen bits. 1.12.1 registers the same blocks, its block class and its
// block list being the same classes to the byte, so 338 draws the same
// numbers, and 1.12 the same again, its block class and block list being
// 1.12.1's with the names taken out, so 335 draws them too.
//
// 1.11.1 registers every block 1.12 does but the eighteen 1.12 added -- the
// sixteen glazed terracottas, the concrete and the concrete powder -- and
// numbers every other state as 1.12 does, state for state, which the two
// jars' own registries say when read side by side. So 1.11.1's file is
// 1.12.2's with those eighteen blocks' ninety-six states at -1, substituted
// as the corals are, and a count of 5,269, which still packs a
// directly-numbered section at thirteen bits.
//
// 1.11 registers the blocks 1.11.1 does and numbers every state as it does,
// slot for slot -- 1.11.1 added no block, and its registry read off the jar
// is 1.11's to the line -- so 315 draws 1.11.1's numbers.
//
// 1.10 registers every block 1.11 does but the seventeen 1.11 added -- the
// observer and the sixteen shulker boxes -- and numbers every other state as
// 1.11 does, slot for slot, which the two jars' registries say when read
// side by side. So 1.10's file is 1.11.1's with those seventeen blocks' 108
// states at -1 -- the undyed shulker box 1.13 split from the purple one goes
// with them, being the same numbers -- and a count of 5,161, which still
// packs a directly-numbered section at thirteen bits.
//
// 1.9.3 registers every block 1.10 does but the five 1.10 added -- the magma
// block, the nether wart block, the red nether bricks, the bone block and the
// structure void -- and numbers every other state as 1.10 does, slot for
// slot, which the two jars' registries say when read side by side. So 1.9.3's
// file is 1.10's with those five blocks' seven states at -1 and a count of
// 5,154, which still packs a directly-numbered section at thirteen bits.
var blockIdsFiles = map[types.ProtocolId]string{
	types.ProtocolVersions.MINECRAFT_1_9_3.ID:  "blockids_minecraft_1_9_3.json",
	types.ProtocolVersions.MINECRAFT_1_10.ID:   "blockids_minecraft_1_10.json",
	types.ProtocolVersions.MINECRAFT_1_11.ID:   "blockids_minecraft_1_11_1.json",
	types.ProtocolVersions.MINECRAFT_1_11_1.ID: "blockids_minecraft_1_11_1.json",
	types.ProtocolVersions.MINECRAFT_1_12.ID:   "blockids_minecraft_1_12_2.json",
	types.ProtocolVersions.MINECRAFT_1_12_1.ID: "blockids_minecraft_1_12_2.json",
	types.ProtocolVersions.MINECRAFT_1_12_2.ID: "blockids_minecraft_1_12_2.json",
}

// blockStateRenames is every block a version knows under an older name than
// the one a newer world stores it by: the name the world uses, and the name
// this version's table numbers it as. A rename is the same block with the
// same properties under a different name, so a lookup under the newer name
// reads the older name's entry, and a world saved after the rename translates
// to the version before it without a hole. 1.20.3 is where grass became
// short grass, so every version before it answers to both names, and 1.17 is
// where the grass path became the dirt path, which 1.16.4, 1.16.3, 1.16.2,
// 1.16.1, 1.16, 1.15.2, 1.15.1, 1.15, 1.14.4, 1.14.3, 1.14.2, 1.14.1, 1.14, 1.13.2, 1.13.1, 1.13, 1.12.2, 1.12.1, 1.12, 1.11.1, 1.11, 1.10 and 1.9.3 answer to as well. 1.17 is also where the cauldron split by what it holds, and the
// water cauldron of three levels is 1.16.4's cauldron at the same levels,
// which holds nothing else: the one rename that narrows a block rather than
// matching it, since 1.16.4's cauldron has an empty level the water cauldron
// cannot name. 1.14 is where the sign became the oak sign, beside the signs
// of the other woods it added, the wall sign the oak wall sign with them,
// and the stone slab the smooth stone slab, which it looks like, for a
// stone slab of plain stone to take its name: 1.13.2, 1.13.1, 1.13, 1.12.2,
// 1.12.1, 1.12, 1.11.1, 1.11, 1.10 and 1.9.3 answer
// to the three newer names, and a world's slab of plain stone is 1.13.2's stone
// slab by its own name, which draws it smooth, the one block 1.13.2 has for
// either.
var blockStateRenames = map[types.ProtocolId]map[string]string{
	types.ProtocolVersions.MINECRAFT_1_9_3.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_10.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_11.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_11_1.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_12.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_12_1.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_12_2.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_13.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_13_1.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_13_2.ID: {
		"minecraft:short_grass":       "minecraft:grass",
		"minecraft:dirt_path":         "minecraft:grass_path",
		"minecraft:water_cauldron":    "minecraft:cauldron",
		"minecraft:oak_sign":          "minecraft:sign",
		"minecraft:oak_wall_sign":     "minecraft:wall_sign",
		"minecraft:smooth_stone_slab": "minecraft:stone_slab",
	},
	types.ProtocolVersions.MINECRAFT_1_14.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_14_1.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_14_2.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_14_3.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_14_4.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_15.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_15_1.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_15_2.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_16.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_16_1.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_16_2.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_16_3.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_16_4.ID: {
		"minecraft:short_grass":    "minecraft:grass",
		"minecraft:dirt_path":      "minecraft:grass_path",
		"minecraft:water_cauldron": "minecraft:cauldron",
	},
	types.ProtocolVersions.MINECRAFT_1_17.ID:   {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_17_1.ID: {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_18.ID:   {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_18_2.ID: {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_19.ID:   {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_19_1.ID: {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_19_3.ID: {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_19_4.ID: {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_20.ID:   {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_20_2.ID: {"minecraft:short_grass": "minecraft:grass"},
}

// The JSON shape of one version's table.
// blockStateValueRenames is, for a version, the property values later
// versions renamed: the newer value, and what this version calls it. It is
// looked at only for a value the version's block does not have, so a value
// that means something to the version is never turned into another.
//
// 1.16 is where a wall's sides went from being there or not to being low or
// tall: a wall stored with a side of either height has that side on 1.15.2,
// on 1.15.1, on 1.15, on 1.14.4, on 1.14.3, on 1.14.2, on 1.14.1, on 1.14, on
// 1.13.2, on 1.13.1, on 1.13, on 1.12.2, on 1.12.1, on 1.12, on 1.11.1, on
// 1.11, on 1.10 and on 1.9.3, and one stored with none does not.
var blockStateValueRenames = map[types.ProtocolId]map[string]string{
	types.ProtocolVersions.MINECRAFT_1_9_3.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_10.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_11.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_11_1.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_12.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_12_1.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_12_2.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_13.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_13_1.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_13_2.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_14.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_14_1.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_14_2.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_14_3.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_14_4.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_15.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_15_1.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
	types.ProtocolVersions.MINECRAFT_1_15_2.ID: {
		"none": "false",
		"low":  "true",
		"tall": "true",
	},
}

type blockStatesFile struct {
	Blocks []blockStatesFileEntry `json:"blocks"`
}

type blockStatesFileEntry struct {
	Name string `json:"name"`

	// Base is the id of the block's first state. The states that follow it are
	// the row-major walk over Properties in the order given.
	Base int32 `json:"base"`

	// Default is which of the block's states the block is when a property goes
	// unmentioned, as an offset from Base. Zero when absent, which most blocks
	// make true by putting the default first.
	Default    int32                     `json:"default"`
	Properties []blockStatesFileProperty `json:"properties"`
}

type blockStatesFileProperty struct {
	Name   string   `json:"name"`
	Values []string `json:"values"`
}

// BlockStatesFor loads the block state numbering of one version. It reports an
// error for a version this server does not speak, since a table it does not
// hold is not one to guess at.
func BlockStatesFor(version types.ProtocolVersion) (*BlockStates, error) {
	return new(BlockStatesLoader).For(version)
}

// A BlockStatesLoader loads the numbering of several versions and lets those
// that number every state alike share one parsed table, for a caller that
// holds every version's numbering at once. Six of the versions this server
// speaks share a file with another (see blockStatesFiles), and a table is
// hundreds of kilobytes, so loading each version on its own would hold six
// copies of tables already in memory. The zero value is ready to use.
type BlockStatesLoader struct {
	tables map[string]*blockStatesTable
	ids    map[string]*blockIdsTable
}

// For loads the numbering of one version, sharing its table with any version
// loaded before it from the same file.
func (l *BlockStatesLoader) For(version types.ProtocolVersion) (*BlockStates, error) {
	name, ok := blockStatesFiles[version.ID]
	if !ok {
		return nil, fmt.Errorf("gamedata: no block state table for protocol %d", version.ID)
	}

	table, ok := l.tables[name]
	if !ok {
		var err error
		if table, err = loadBlockStatesTable(name); err != nil {
			return nil, err
		}

		if l.tables == nil {
			l.tables = make(map[string]*blockStatesTable)
		}

		l.tables[name] = table
	}

	states := &BlockStates{table: table, renames: blockStateRenames[version.ID], valueRenames: blockStateValueRenames[version.ID]}

	if idsName, ok := blockIdsFiles[version.ID]; ok {
		ids, ok := l.ids[idsName]
		if !ok {
			var err error
			if ids, err = loadBlockIdsTable(idsName, table); err != nil {
				return nil, err
			}

			if l.ids == nil {
				l.ids = make(map[string]*blockIdsTable)
			}

			l.ids[idsName] = ids
		}

		states.ids = ids
	}

	for newer, older := range states.renames {
		if _, ok := table.blocks[older]; !ok {
			return nil, fmt.Errorf("gamedata: %s renames %s to %s, which protocol %d does not number", name, newer, older, version.ID)
		}

		if _, ok := table.blocks[newer]; ok {
			return nil, fmt.Errorf("gamedata: %s renames %s to %s, but protocol %d numbers both", name, newer, older, version.ID)
		}
	}

	return states, nil
}

// loadBlockStatesTable parses one table out of the embedded data directory.
func loadBlockStatesTable(name string) (*blockStatesTable, error) {
	raw, err := dataFiles.ReadFile("data/" + name)
	if err != nil {
		return nil, fmt.Errorf("gamedata: %w", err)
	}

	var file blockStatesFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("gamedata: parsing %s: %w", name, err)
	}

	table := &blockStatesTable{blocks: make(map[string]*blockStatesEntry, len(file.Blocks))}

	for _, block := range file.Blocks {
		entry := &blockStatesEntry{base: block.Base, defaultOff: block.Default}

		count := int32(1)
		for _, property := range block.Properties {
			entry.properties = append(entry.properties, blockStateProperty{name: property.Name, values: property.Values})
			count *= int32(len(property.Values))
		}

		table.blocks[block.Name] = entry

		if end := block.Base + count; end > table.stateCount {
			table.stateCount = end
		}
	}

	return table, nil
}

// loadBlockIdsTable parses one block id file out of the embedded data
// directory, for the table whose states it numbers.
func loadBlockIdsTable(name string, table *blockStatesTable) (*blockIdsTable, error) {
	raw, err := dataFiles.ReadFile("data/" + name)
	if err != nil {
		return nil, fmt.Errorf("gamedata: %w", err)
	}

	var file struct {
		StateCount int32   `json:"stateCount"`
		Ids        []int32 `json:"ids"`
	}

	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("gamedata: parsing %s: %w", name, err)
	}

	if int32(len(file.Ids)) != table.stateCount {
		return nil, fmt.Errorf("gamedata: %s numbers %d states, and its table holds %d", name, len(file.Ids), table.stateCount)
	}

	// A number is a block id of eight bits and a variant of four.
	for state, id := range file.Ids {
		if id < -1 || id > maxLegacyBlockId {
			return nil, fmt.Errorf("gamedata: %s numbers state %d as %d, past the %d a block id and a variant reach", name, state, id, maxLegacyBlockId)
		}
	}

	return &blockIdsTable{ids: file.Ids, stateCount: file.StateCount}, nil
}

// number is the number the client knows a state of the table by: the state's
// own on a version whose table numbers them as its client does, and the one
// the block id file gives it on a version from before the flattening, which
// reports false for a state that version has no number for.
func (s *BlockStates) number(state int32) (int32, bool) {
	if s.ids == nil {
		return state, true
	}

	id := s.ids.ids[state]

	return id, id >= 0
}

// entry finds the block a name numbers, under the name itself or under the
// older name this version knows it by.
func (s *BlockStates) entry(name string) (*blockStatesEntry, bool) {
	if older, renamed := s.renames[name]; renamed {
		name = older
	}

	entry, ok := s.table.blocks[name]

	return entry, ok
}

// StateCount is how many block states the version numbers in all.
func (s *BlockStates) StateCount() int32 {
	if s.ids != nil {
		return s.ids.stateCount
	}

	return s.table.stateCount
}

// Id numbers one block state: a block name and the properties a world's
// palette stored for it. A property the palette does not mention takes the
// value the block defaults to, which is how a world written by a version that
// had fewer properties still resolves.
//
// It reports false for a name this version does not know and for a property
// value it does not know, and decides nothing about what to send instead; the
// caller knows what a hole in a world should look like.
func (s *BlockStates) Id(name string, properties map[string]string) (int32, bool) {
	entry, ok := s.entry(name)
	if !ok {
		return 0, false
	}

	// The id is the row-major position over the property values, refined a
	// property at a time. Missing properties take the default state's value at
	// their own position, which is what peeling the default offset digit by
	// digit recovers.
	id := entry.base
	defaultOff := entry.defaultOff

	for i := len(entry.properties) - 1; i >= 0; i-- {
		property := entry.properties[i]
		count := int32(len(property.values))
		defaultIndex := defaultOff % count
		defaultOff /= count

		index := defaultIndex
		if value, present := properties[property.name]; present {
			index = property.index(value)

			if older, renamed := s.valueRenames[value]; index < 0 && renamed {
				index = property.index(older)
			}

			if index < 0 {
				return 0, false
			}
		}

		id += index * s.stride(entry, i)
	}

	return s.number(id)
}

// index is where value sits among the property's values, or -1.
func (p blockStateProperty) index(value string) int32 {
	for i, candidate := range p.values {
		if candidate == value {
			return int32(i)
		}
	}

	return -1
}

// DefaultId numbers the state a block is in when nothing says otherwise. It
// reports false for a name this version does not know.
func (s *BlockStates) DefaultId(name string) (int32, bool) {
	entry, ok := s.entry(name)
	if !ok {
		return 0, false
	}

	return s.number(entry.base + entry.defaultOff)
}

// stride is how far apart states sit when property i moves one value: the
// product of the value counts of every property that varies faster.
func (s *BlockStates) stride(entry *blockStatesEntry, i int) int32 {
	stride := int32(1)
	for _, property := range entry.properties[i+1:] {
		stride *= int32(len(property.values))
	}

	return stride
}
