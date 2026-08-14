package anvil

import (
	"bytes"
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// BlockState is one entry of a section's palette as the world stores it: a
// block by name, and the properties that pick which of its states this is. A
// palette that stores a block in its default state may leave Properties empty
// or mention every property; both name the same state.
type BlockState struct {
	Name       string
	Properties map[string]string
}

// Section is one 16x16x16 slice of a chunk. The world stores light for one
// section above and one below the block range, so a Section may carry light
// and no blocks at all, which a nil BlockPalette says.
type Section struct {
	// Y is the section's height in sections: block y from Y*16 up.
	Y int32

	// BlockPalette is every distinct state in the section, and BlockData packs
	// an index into it per block, 4096 of them in YZX order. Data is nil when
	// the palette holds a single entry, because there is nothing to tell
	// apart; it packs at max(4, bits the palette size needs) bits per index
	// otherwise, with indices never crossing a long boundary.
	BlockPalette []BlockState
	BlockData    []int64

	// SkyLight and BlockLight are half a byte per block in YZX order, 2048
	// bytes each, and nil where the world stored none.
	SkyLight   []byte
	BlockLight []byte
}

// Chunk is one column of sections as the world stores it.
type Chunk struct {
	X, Z int32

	// MinSectionY is where the block range starts, in sections.
	MinSectionY int32

	// Status is how far generation got. Only "minecraft:full" chunks are
	// finished terrain; the rest are scaffolding a generating world left
	// behind.
	Status string

	// Sections is every section the world stored, in the order stored,
	// including the light-only ones outside the block range.
	Sections []Section

	// Heightmaps is the packed heightmaps the world stored, by the names it
	// stores them under, 256 values in ZX order at however many bits the
	// world's height needs, indices never crossing a long boundary.
	Heightmaps map[string][]int64
}

// ReadChunk reads one chunk from the region directory, or ErrChunkNotFound
// for one the world never generated.
func ReadChunk(regionDir string, x, z int32) (*Chunk, error) {
	payload, err := ReadChunkPayload(regionDir, x, z)
	if err != nil {
		return nil, err
	}

	// Files write the named root form.
	_, root, err := nbt.ReadNamed(streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(payload)))
	if err != nil {
		return nil, fmt.Errorf("anvil: chunk %d,%d: %w", x, z, err)
	}

	compound, ok := root.(nbt.Compound)
	if !ok {
		return nil, fmt.Errorf("anvil: chunk %d,%d: root is %s, want compound", x, z, root.Type())
	}

	chunk, err := parseChunk(compound)
	if err != nil {
		return nil, fmt.Errorf("anvil: chunk %d,%d: %w", x, z, err)
	}

	if chunk.X != x || chunk.Z != z {
		return nil, fmt.Errorf("anvil: chunk %d,%d: file says it is chunk %d,%d", x, z, chunk.X, chunk.Z)
	}

	return chunk, nil
}

func parseChunk(root nbt.Compound) (*Chunk, error) {
	chunk := &Chunk{
		X:           compoundInt(root, "xPos"),
		Z:           compoundInt(root, "zPos"),
		MinSectionY: compoundInt(root, "yPos"),
		Status:      compoundString(root, "Status"),
		Heightmaps:  map[string][]int64{},
	}

	if heightmaps, ok := root["Heightmaps"].(nbt.Compound); ok {
		for name, tag := range heightmaps {
			if packed, ok := tag.(nbt.LongArray); ok {
				chunk.Heightmaps[name] = packed
			}
		}
	}

	sections, ok := root["sections"].(nbt.List)
	if !ok {
		// A chunk with no sections at all is a shape generation never leaves a
		// full chunk in, but it parses: a column of nothing.
		return chunk, nil
	}

	for _, tag := range sections.Elements {
		compound, ok := tag.(nbt.Compound)
		if !ok {
			return nil, fmt.Errorf("section is %s, want compound", tag.Type())
		}

		section, err := parseSection(compound)
		if err != nil {
			return nil, err
		}

		chunk.Sections = append(chunk.Sections, section)
	}

	return chunk, nil
}

func parseSection(compound nbt.Compound) (Section, error) {
	section := Section{Y: compoundInt(compound, "Y")}

	if light, ok := compound["SkyLight"].(nbt.ByteArray); ok {
		section.SkyLight = light
	}

	if light, ok := compound["BlockLight"].(nbt.ByteArray); ok {
		section.BlockLight = light
	}

	states, ok := compound["block_states"].(nbt.Compound)
	if !ok {
		return section, nil
	}

	palette, ok := states["palette"].(nbt.List)
	if !ok {
		return Section{}, fmt.Errorf("section %d has block states without a palette", section.Y)
	}

	for _, tag := range palette.Elements {
		entry, ok := tag.(nbt.Compound)
		if !ok {
			return Section{}, fmt.Errorf("section %d palette entry is %s, want compound", section.Y, tag.Type())
		}

		state := BlockState{Name: compoundString(entry, "Name")}
		if state.Name == "" {
			return Section{}, fmt.Errorf("section %d palette entry has no name", section.Y)
		}

		if properties, ok := entry["Properties"].(nbt.Compound); ok {
			state.Properties = make(map[string]string, len(properties))
			for name, value := range properties {
				text, ok := value.(nbt.String)
				if !ok {
					return Section{}, fmt.Errorf("section %d property %s is %s, want string", section.Y, name, value.Type())
				}

				state.Properties[name] = string(text)
			}
		}

		section.BlockPalette = append(section.BlockPalette, state)
	}

	if len(section.BlockPalette) == 0 {
		return Section{}, fmt.Errorf("section %d palette is empty", section.Y)
	}

	if data, ok := states["data"].(nbt.LongArray); ok {
		section.BlockData = data
	}

	if len(section.BlockPalette) > 1 && section.BlockData == nil {
		return Section{}, fmt.Errorf("section %d has %d palette entries and no data picking between them", section.Y, len(section.BlockPalette))
	}

	return section, nil
}

// compoundInt reads a number stored under name however narrowly the writer
// stored it. Section heights fit a byte and chunk coordinates an int, and a
// writer is free to have used either.
func compoundInt(compound nbt.Compound, name string) int32 {
	switch value := compound[name].(type) {
	case nbt.Byte:
		return int32(value)
	case nbt.Short:
		return int32(value)
	case nbt.Int:
		return int32(value)
	}

	return 0
}

func compoundString(compound nbt.Compound, name string) string {
	if value, ok := compound[name].(nbt.String); ok {
		return string(value)
	}

	return ""
}
