package transformers

import (
	"bytes"
	"encoding/binary"
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.15 step is where the biomes of a chunk took on a height. 1.14.4 below
// it holds one biome for every column of a chunk, and reads them off the end
// of the chunk's sections; 1.15 holds one for every four blocks each way, up
// the whole height of the world, and reads them out of the packet itself.
// Read off the 1.14.4 client's own classes through Mojang's mappings the way
// the steps above were, the two versions lay out alike the whole login phase,
// everything this server reads, and of what it sends the tags, the player
// list, the teleport, the head rotation, the animate, the entity metadata with
// the pose at 18, the keep alive, the player position, the entity removal,
// the chunk cache centre, the spawn position and the light update. They
// number what this server reads alike, and what it sends differently from
// 0x08 on, which the id tables say. Three packets differ, all of them on the
// way down, and are below.

const (
	// biomes1_14_4 is how many biomes a 1.14.4 chunk holds: one for every
	// column.
	biomes1_14_4 = 16 * 16

	// biomeCellsAcross1_15 is how many of 1.15's biomes lie side by side
	// along a chunk, each of them four blocks wide.
	biomeCellsAcross1_15 = 4

	// entityDataEnd is what closes a run of entity metadata: the index no
	// entry has.
	entityDataEnd = 0xFF
)

// DowngradePlayLoginTo1_14_4 rewrites the play phase login packet from what
// 1.15 sends into what 1.14.4 reads.
//
// 1.14.4 lays the packet out as 1.15 does but for two things 1.15 added: the
// seed, behind the dimension, which its biomes are blended by, and the flag
// at the end that says whether a death is followed by its screen.
func DowngradePlayLoginTo1_14_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id, a plain int, the game mode's byte and the dimension,
	// another int.
	if err := copyBytes(in, out, 9); err != nil {
		return err
	}

	// The seed.
	if _, err := in.ReadLong(); err != nil {
		return err
	}

	// The most players, a byte.
	if err := copyBytes(in, out, 1); err != nil {
		return err
	}

	// The level type.
	if err := copyString(in, out); err != nil {
		return err
	}

	// The view distance.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	// Whether the debug screen is cut down.
	if _, err := copyBoolean(in, out); err != nil {
		return err
	}

	// Whether a death is followed by its screen, which 1.14.4 is not told.
	if _, err := in.ReadBoolean(); err != nil {
		return err
	}

	return nil
}

// DowngradeLevelChunkTo1_14_4 rewrites the chunk packet from what 1.15 sends
// into what 1.14.4 reads.
//
// 1.15 holds the biomes of a whole chunk in the packet, between the
// heightmaps and the sections: 1024 plain ints, one for every four blocks
// each way, from the bottom of the world up. 1.14.4 reads the biomes of a
// whole chunk off the end of its sections, inside the same counted run of
// bytes: 256 plain ints, one for every column, row by row. A column is given
// the biome of the lowest of 1.15's cells it stands in, which is the whole
// of what 1.14.4 can say of it. The coordinates, the flag, the mask, the
// heightmaps, the sections and the block entities are laid out alike, and
// the two versions pack their entries the same way.
func DowngradeLevelChunkTo1_14_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	wholeChunk, err := copyBoolean(in, out)
	if err != nil {
		return err
	}

	// The mask of the sections.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	name, heightmaps, err := nbt.ReadNamed(in)
	if err != nil {
		return err
	}

	if err := nbt.WriteNamed(out, name, heightmaps); err != nil {
		return err
	}

	// Only a whole chunk carries its biomes, on both sides of the step.
	var columns []byte

	if wholeChunk {
		cells := make([]int32, biomes1_16_4)
		for i := range cells {
			if cells[i], err = in.ReadInt(); err != nil {
				return err
			}
		}

		columns = make([]byte, 0, biomes1_14_4*4)

		for z := range 16 {
			for x := range 16 {
				cell := cells[z/4*biomeCellsAcross1_15+x/4]

				columns = binary.BigEndian.AppendUint32(columns, uint32(cell))
			}
		}
	}

	sections, err := in.ReadByteArray(streams.MaxPacketSize)
	if err != nil {
		return err
	}

	if len(sections)+len(columns) > streams.MaxPacketSize {
		return fmt.Errorf("%d bytes of sections and %d of biomes, more than a packet holds", len(sections), len(columns))
	}

	if err := out.WriteByteArray(bytes.Join([][]byte{sections, columns}, nil)); err != nil {
		return err
	}

	// The block entities.
	return copyRest(in, out)
}

// DowngradeAddPlayerTo1_14_4 rewrites the add player packet -- which is what
// the add entity packet has been since the 1.20.2 step -- from what 1.15
// sends into what 1.14.4 reads.
//
// 1.14.4 reads the player's entity metadata off the end of the packet, which
// 1.15 took off it and left to the entity metadata packet alone. This server
// says what a player's stance is with that packet on every version, behind
// the add player wherever there is something to say, so the run here is an
// empty one: its end, and nothing in front of it, which 1.14.4 reads as no
// metadata and leaves the player as it made it.
func DowngradeAddPlayerTo1_14_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyRest(in, out); err != nil {
		return err
	}

	return out.WriteByte(entityDataEnd)
}
