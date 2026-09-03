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
// 1.21.9 names the 1.21.11 file rather than a copy because the two versions
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
// states. And 1.19.1 its own at the very bottom: 761 is where the bamboo
// wood set, the hanging signs of every wood, the chiseled bookshelf and the
// piglin heads landed, thirty-nine blocks in all, and where the note block
// gained its seven mob head instruments, so 760 numbers 21,448 states.
var blockStatesFiles = map[types.ProtocolId]string{
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
}

// blockStateRenames is every block a version knows under an older name than
// the one a newer world stores it by: the name the world uses, and the name
// this version's table numbers it as. A rename is the same block with the
// same properties under a different name, so a lookup under the newer name
// reads the older name's entry, and a world saved after the rename translates
// to the version before it without a hole. 1.20.3 is where grass became
// short grass, the one rename among the versions this server speaks, so
// every version before it answers to both names.
var blockStateRenames = map[types.ProtocolId]map[string]string{
	types.ProtocolVersions.MINECRAFT_1_19_1.ID: {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_19_3.ID: {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_19_4.ID: {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_20.ID:   {"minecraft:short_grass": "minecraft:grass"},
	types.ProtocolVersions.MINECRAFT_1_20_2.ID: {"minecraft:short_grass": "minecraft:grass"},
}

// The JSON shape of one version's table.
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
// holds every version's numbering at once. Three of the versions this server
// speaks share a file with another (see blockStatesFiles), and a table is
// hundreds of kilobytes, so loading each version on its own would hold three
// copies of tables already in memory. The zero value is ready to use.
type BlockStatesLoader struct {
	tables map[string]*blockStatesTable
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

	states := &BlockStates{table: table, renames: blockStateRenames[version.ID]}

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
			index = -1
			for j, candidate := range property.values {
				if candidate == value {
					index = int32(j)
					break
				}
			}

			if index < 0 {
				return 0, false
			}
		}

		id += index * s.stride(entry, i)
	}

	return id, true
}

// DefaultId numbers the state a block is in when nothing says otherwise. It
// reports false for a name this version does not know.
func (s *BlockStates) DefaultId(name string) (int32, bool) {
	entry, ok := s.entry(name)
	if !ok {
		return 0, false
	}

	return entry.base + entry.defaultOff, true
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
