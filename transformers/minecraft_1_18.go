package transformers

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.18 step is where the world grew: 1.18 is the version that raised the
// build height, and with it reworked how a chunk travels. Read off the 1.17.1
// client's own classes through Mojang's mappings the way the steps above
// were, 1.17.1 and 1.18 lay all but two of the packets this server speaks
// out alike -- the login phase whole, the player list, the player position,
// the movement on both sides, the entity metadata with the pose at 18, the
// tags, the spawn position -- and number them alike but for the one packet
// 1.18 added at the end of the play phase, which the id tables say. The two
// that differ are below.
//
// The play login is 1.18's with one field fewer: 1.18 is where the
// simulation distance joined it, behind the view distance. The registries it
// carries and the dimension type it spells out are 1.17.1's own as well,
// since 1.18 is where the biome codec lost its depth and its scale, which a
// 1.17.1 client reads as required fields.
//
// The chunk packet is where the two versions part ways. 1.18 folded the
// light into it, put a biome container into every section, let a section of
// one block state say so with a palette of one, and sends every section of
// the chunk; 1.17.1 reads the light in a packet of its own, ahead of the
// chunk, the biomes as one array for the whole chunk, a section of one block
// state as a four bit palette of one entry over a full array of zeros, and
// only the sections that hold a block, behind a mask saying which. The
// rewrite below turns 1.18's chunk packet into 1.17.1's, dropping the light
// it carries: package world sends that ahead of the chunk in the packet
// 1.17.1 reads it in, built from the same light.

// DowngradePlayLoginTo1_17_1 rewrites the play phase login packet from what
// 1.18 sends into what 1.17.1 reads, given the registries that version
// reads out of it and the dimension type it spells out in it.
//
// The two versions lay the packet out alike from front to back but for the
// simulation distance, which 1.18 put behind the view distance and 1.17.1
// has no field for, so it comes off. The registries in the middle are
// 1.18's, put there by the 1.18.2 step, and the dimension type spelled out
// behind them 1.18's as well; 1.17.1 has its own of each, so both are read
// so that they are consumed and never written, and 1.17.1's go in their
// place. The rest is copied. As with the steps above, a transformer built
// with no registries or no dimension type to write refuses every login
// rather than send a client into a world it cannot make sense of.
func DowngradePlayLoginTo1_17_1(registryCodec []byte, dimensionType []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.17.1 play login carries the registries, and this transformer was given none")
		}

		if len(dimensionType) == 0 {
			return errors.New("a 1.17.1 play login spells out the dimension type, and this transformer was given none")
		}

		// The entity id, a plain int, the hardcore flag, and the game mode
		// and the previous game mode, a byte each.
		if err := copyBytes(in, out, 7); err != nil {
			return err
		}

		// The dimension names.
		dimensionCount, err := copyVarInt(in, out)
		if err != nil {
			return err
		}

		for range dimensionCount {
			if err := copyString(in, out); err != nil {
				return err
			}
		}

		// 1.18's registries, and 1.17.1's in their place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(registryCodec); err != nil {
			return err
		}

		// 1.18's dimension type, and 1.17.1's in its place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(dimensionType); err != nil {
			return err
		}

		// The dimension.
		if err := copyString(in, out); err != nil {
			return err
		}

		// The seed.
		if err := copyBytes(in, out, 8); err != nil {
			return err
		}

		// Max players and the view distance.
		for range 2 {
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		}

		// The simulation distance, read so that it is consumed and never
		// written.
		if _, err := in.ReadVarInt(); err != nil {
			return err
		}

		// The four flags, laid out alike on both sides of the step.
		return copyRest(in, out)
	}
}

// DowngradeLevelChunkWithLightTo1_17_1 rewrites the chunk packet from what
// 1.18 sends into what 1.17.1 reads, which is the chunk without its light.
//
// 1.17.1 lays the packet out as the coordinates, a mask of the sections that
// follow, the heightmaps, one biome for every four blocks of the whole
// chunk, the sections the mask names, and the block entities. The rewrite
// reads 1.18's sections one by one, since the mask has to be known before
// the sections are written and the sections have to be known before the
// mask is: a section holding no block is left out and its bit clear, as a
// vanilla server of that version leaves it, and the client fills it with
// air; a section of one state, which 1.18 sends as a palette of one and no
// data, is written back out the way 1.17.1 reads a palette that small, four
// bits over an array of zeros; every other section travels as it was, since
// 1.18 packs a palette of two to two hundred and fifty-six entries as
// 1.17.1 does, and the ids past that as well. The biome container 1.18 puts
// into every section comes off, and the biome array goes in its place: the
// first and only biome this server registers, for every four blocks of every
// section, as many sections as the packet holds. The light at the end is
// read so that it is consumed and never written; package world sends it
// ahead of the chunk in the packet 1.17.1 reads it in.
//
// The block entities are refused rather than carried: 1.17.1 reads each as
// a compound naming its type and its position, which 1.18 sends as a type
// id and a packed position beside a compound that names neither, and this
// server sends none, so there is nothing to spell out and no way to spell
// it. A chunk carrying one is a chunk this server did not build.
func DowngradeLevelChunkWithLightTo1_17_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	// The heightmaps, a compound named as a root on both sides of this step,
	// kept for after the mask.
	name, heightmaps, err := nbt.ReadNamed(in)
	if err != nil {
		return err
	}

	// 1.18's section buffer, walked section by section into 1.17.1's.
	sectionData, err := in.ReadByteArray(streams.MaxPacketSize)
	if err != nil {
		return err
	}

	sections, mask, err := rewriteSectionsTo1_17_1(sectionData)
	if err != nil {
		return err
	}

	// The block entities, of which this server sends none: see above.
	blockEntityCount, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if blockEntityCount != 0 {
		return fmt.Errorf("a 1.17.1 chunk names each block entity's type and position inside its compound, which this server cannot spell out: the packet carries %d", blockEntityCount)
	}

	// The light: the trust edges flag the 1.20 step put in, the masks and
	// the arrays, read so that they are consumed and never written.
	if _, err := in.ReadRest(); err != nil {
		return err
	}

	// The mask, a counted long array: one long covers the sixty-four
	// sections no dimension type reaches.
	if err := out.WriteVarInt(1); err != nil {
		return err
	}

	if err := out.WriteLong(mask); err != nil {
		return err
	}

	if err := nbt.WriteNamed(out, name, heightmaps); err != nil {
		return err
	}

	// The biomes, one var int per four blocks of every section, all of them
	// the one biome this server registers: see package gamedata.
	biomeCount := int32(len(sections)) * biomesPerSection

	if err := out.WriteVarInt(biomeCount); err != nil {
		return err
	}

	for range biomeCount {
		if err := out.WriteVarInt(0); err != nil {
			return err
		}
	}

	// The sections the mask names, and the block entities.
	buffer := make([]byte, 0, len(sectionData))
	for _, section := range sections {
		buffer = append(buffer, section...)
	}

	if err := out.WriteByteArray(buffer); err != nil {
		return err
	}

	return out.WriteVarInt(0)
}

// rewriteSectionsTo1_17_1 walks a 1.18 section buffer and returns each
// section as 1.17.1 reads it -- nil for one the mask leaves out -- and the
// mask itself, bit i set for the ith section from the bottom of the world.
func rewriteSectionsTo1_17_1(data []byte) ([][]byte, int64, error) {
	in := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(data))

	var sections [][]byte
	var mask int64

	for index := 0; ; index++ {
		blockCount, err := in.ReadShort()
		if errors.Is(err, errEndOfSections) {
			return sections, mask, nil
		}

		if err != nil {
			return nil, 0, fmt.Errorf("section %d: %w", index, err)
		}

		if index >= sectionsPerMask {
			return nil, 0, fmt.Errorf("section %d is past the %d a 1.17.1 mask names", index, sectionsPerMask)
		}

		blocks, err := readContainer(in, blockIndirectBits)
		if err != nil {
			return nil, 0, fmt.Errorf("section %d: %w", index, err)
		}

		// The biome container, read so that it is consumed and never
		// written.
		if _, err := readContainer(in, biomeIndirectBits); err != nil {
			return nil, 0, fmt.Errorf("section %d: biomes: %w", index, err)
		}

		if blockCount == 0 {
			sections = append(sections, nil)

			continue
		}

		mask |= 1 << index

		section := make([]byte, 0, 2+len(blocks.raw)+len(singleValueSectionData))
		section = append(section, byte(blockCount>>8), byte(blockCount))

		if blocks.bits == 0 {
			// A palette of one over no data, which 1.17.1 has no form for:
			// the smallest palette it reads is four bits over a full array,
			// so the one entry goes into that, with every index naming it.
			section = append(section, linearPaletteBits, 1)
			section = streams.AppendVarInt(section, blocks.singleValue)
			section = append(section, singleValueSectionData...)
		} else {
			section = append(section, blocks.raw...)
		}

		sections = append(sections, section)
	}
}

// A container as read off a 1.18 section: its declared bits, the one value
// a palette of none names, and the bytes it was read from.
type container struct {
	bits        byte
	singleValue int32
	raw         []byte
}

// readContainer reads one paletted container the way 1.18 writes it: the
// bits, the palette -- one value for no bits, a counted list of values up
// to indirectBits, and nothing past that, the data being ids -- and the
// counted longs, which 1.18 still counts, 1.21.5 being where the count
// left. It returns the container with the bytes it was read from, so that a
// section 1.17.1 reads as 1.18 wrote it can travel as it was.
func readContainer(in *streams.MinecraftStream, indirectBits byte) (container, error) {
	var raw []byte

	bits, err := in.ReadByte()
	if err != nil {
		return container{}, err
	}

	raw = append(raw, bits)

	var singleValue int32

	switch {
	case bits == 0:
		if singleValue, err = in.ReadVarInt(); err != nil {
			return container{}, err
		}

		raw = streams.AppendVarInt(raw, singleValue)
	case bits <= indirectBits:
		count, err := in.ReadVarInt()
		if err != nil {
			return container{}, err
		}

		raw = streams.AppendVarInt(raw, count)

		for range count {
			id, err := in.ReadVarInt()
			if err != nil {
				return container{}, err
			}

			raw = streams.AppendVarInt(raw, id)
		}
	}

	longCount, err := in.ReadVarInt()
	if err != nil {
		return container{}, err
	}

	if longCount < 0 || longCount > streams.MaxPacketSize/8 {
		return container{}, fmt.Errorf("a container of %d longs", longCount)
	}

	raw = streams.AppendVarInt(raw, longCount)

	longs, err := in.ReadBytes(longCount * 8)
	if err != nil {
		return container{}, err
	}

	return container{bits: bits, singleValue: singleValue, raw: append(raw, longs...)}, nil
}

// errEndOfSections is what reading the next section's block count returns
// when the buffer holds no next section, which is the one place the buffer
// may end: a buffer that ends inside a section is a short read of a field,
// which is a different error.
var errEndOfSections = io.EOF

const (
	// blockIndirectBits and biomeIndirectBits are the most bits a container
	// of each kind declares while still carrying a palette: past them the
	// data is ids. 1.18 and 1.17.1 agree on the blocks' eight.
	blockIndirectBits = 8
	biomeIndirectBits = 3

	// linearPaletteBits is the fewest bits a 1.17.1 section reads its
	// palette at, whatever fewer the wire declares.
	linearPaletteBits = 4

	// biomesPerSection is how many biomes a section of 1.17.1's biome array
	// holds: one per four blocks along each axis.
	biomesPerSection = 64

	// sectionsPerMask is how many sections one long of mask names, which is
	// more than any dimension type reaches.
	sectionsPerMask = 64
)

// singleValueSectionData is the data of a section whose every block names
// palette entry zero at four bits: a count of the longs four bits over 4096
// blocks pack into, then those longs, all zero.
var singleValueSectionData = append(streams.AppendVarInt(nil, singleValueLongs), make([]byte, singleValueLongs*8)...)

const singleValueLongs = 4096 * linearPaletteBits / 64
