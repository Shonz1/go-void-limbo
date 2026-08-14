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
var blockStatesFiles = map[types.ProtocolId]string{
	types.ProtocolVersions.MINECRAFT_26_1.ID: "blockstates_minecraft_26_1.json",
	types.ProtocolVersions.MINECRAFT_26_2.ID: "blockstates_minecraft_26_2.json",
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
	name, ok := blockStatesFiles[version.ID]
	if !ok {
		return nil, fmt.Errorf("gamedata: no block state table for protocol %d", version.ID)
	}

	raw, err := dataFiles.ReadFile("data/" + name)
	if err != nil {
		return nil, fmt.Errorf("gamedata: %w", err)
	}

	var file blockStatesFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, fmt.Errorf("gamedata: parsing %s: %w", name, err)
	}

	states := &BlockStates{blocks: make(map[string]*blockStatesEntry, len(file.Blocks))}

	for _, block := range file.Blocks {
		entry := &blockStatesEntry{base: block.Base, defaultOff: block.Default}

		count := int32(1)
		for _, property := range block.Properties {
			entry.properties = append(entry.properties, blockStateProperty{name: property.Name, values: property.Values})
			count *= int32(len(property.Values))
		}

		states.blocks[block.Name] = entry

		if end := block.Base + count; end > states.stateCount {
			states.stateCount = end
		}
	}

	return states, nil
}

// StateCount is how many block states the version numbers in all.
func (s *BlockStates) StateCount() int32 {
	return s.stateCount
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
	entry, ok := s.blocks[name]
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
	entry, ok := s.blocks[name]
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
