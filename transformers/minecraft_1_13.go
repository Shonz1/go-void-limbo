package transformers

import (
	"bytes"
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.13 step is where the flattening landed. 1.12.2 below it knows a block
// by its number and a four-bit variant rather than by a state of its own, and
// has no tags packet and no login plugin message. Mojang published no
// mappings for either version, so the two were read off their own obfuscated
// classes, packet by packet, by the calls each makes on the byte buffer: the
// handshake, the status and the login phases lay out alike, and of the play
// phase so do the add player with its metadata, the animate, the keep alive,
// the play login -- the difficulty and the level type included -- the player
// list, the player position, the entity removal, the head rotation, the
// entity metadata, whose byte serializer is the first either registers, the
// spawn position, the teleport and everything this server reads. The two
// number the play phase differently on both sides, which the id tables say.
//
// What differs is the chunk, below, and the numbers its blocks go by, which
// package world packs as 1.12.2's own out of package gamedata's table.

const (
	// biomes1_12_2 is how many biomes a whole 1.12.2 chunk carries: one for
	// every column, as on 1.13.
	biomes1_12_2 = 16 * 16

	// biomeMax1_12_2 is the largest biome a 1.12.2 chunk can name, a biome
	// being one unsigned byte there.
	biomeMax1_12_2 = 0xFF
)

// DowngradeLevelChunkWithSectionLightTo1_12_2 rewrites the chunk packet --
// the one packet that carries a chunk and its light on 1.13 -- from what 1.13
// sends into what 1.12.2 reads.
//
// The two lay the packet out alike up to its sections, and the block
// entities after them. A section is alike too -- the bits, the palette, the
// counted longs packed across one another, and the block light and the sky
// light behind them -- but for a section whose ids go on the wire without a
// palette: 1.12.2's container reads a palette of its own there all the same,
// which says nothing and is a count of none, and 1.13's reads no palette at
// all. The biomes behind the last section are a byte each on 1.12.2 and an
// int each on 1.13. The ids themselves are 1.12.2's already: a version's chunk
// is built in its own numbering, and the ids here are the block numbers and
// variants 1.12.2 draws.
func DowngradeLevelChunkWithSectionLightTo1_12_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk's coordinates, two plain ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	whole, err := copyBoolean(in, out)
	if err != nil {
		return err
	}

	mask, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	sectionData, err := in.ReadByteArray(streams.MaxPacketSize)
	if err != nil {
		return err
	}

	sections, err := sectionsTo1_12_2(sectionData, mask, whole)
	if err != nil {
		return err
	}

	if err := out.WriteByteArray(sections); err != nil {
		return err
	}

	// The block entities.
	return copyRest(in, out)
}

// sectionsTo1_12_2 rewrites 1.13's sections, and the biomes of a whole chunk
// behind them, into 1.12.2's.
func sectionsTo1_12_2(data []byte, mask int32, whole bool) ([]byte, error) {
	in := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(data))

	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if mask>>sections1_13_2 != 0 {
		return nil, fmt.Errorf("the mask %b names sections past the %d a 1.12.2 chunk holds", mask, sections1_13_2)
	}

	for index := 0; index < sections1_13_2; index++ {
		if mask&(1<<index) == 0 {
			continue
		}

		if err := sectionTo1_12_2(in, out); err != nil {
			return nil, fmt.Errorf("section %d: %w", index, err)
		}
	}

	if whole {
		for range biomes1_12_2 {
			biome, err := in.ReadInt()
			if err != nil {
				return nil, fmt.Errorf("the biomes: %w", err)
			}

			if biome < 0 || biome > biomeMax1_12_2 {
				return nil, fmt.Errorf("biome %d, which a 1.12.2 chunk has no byte for", biome)
			}

			if err := out.WriteByte(byte(biome)); err != nil {
				return nil, err
			}
		}
	}

	if rest, err := in.ReadRest(); err != nil {
		return nil, err
	} else if len(rest) != 0 {
		return nil, fmt.Errorf("%d bytes past the sections and biomes the mask names", len(rest))
	}

	if err := out.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// sectionTo1_12_2 rewrites one of 1.13's sections into 1.12.2's: the same
// container and light, and a palette of none in front of the longs of a
// container that has no palette.
func sectionTo1_12_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	bits, err := in.ReadByte()
	if err != nil {
		return err
	}

	if bits == 0 || bits > 32 {
		return fmt.Errorf("a container of %d bits to an entry, which 1.12.2 has no form for", bits)
	}

	if err := out.WriteByte(bits); err != nil {
		return err
	}

	if bits <= blockIndirectBits {
		paletteSize, err := copyVarInt(in, out)
		if err != nil {
			return err
		}

		for range paletteSize {
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		}
	} else if err := out.WriteVarInt(0); err != nil {
		return err
	}

	longCount, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	if longCount < 0 || longCount > streams.MaxPacketSize/8 {
		return fmt.Errorf("a container of %d longs", longCount)
	}

	if err := copyBytes(in, out, 8*longCount); err != nil {
		return err
	}

	// The block light and the sky light, whole arrays with nothing in front.
	return copyBytes(in, out, 2*lightArrayBytes)
}
