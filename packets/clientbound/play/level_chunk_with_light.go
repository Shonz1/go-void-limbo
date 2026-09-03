package play

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The heightmap kinds a chunk packet may carry, numbered as the client's
// Heightmap.Types enum orders them. Only the two the client keeps after
// loading are sent, which is also all a vanilla server sends.
const (
	// HeightmapWorldSurface is the highest non-air block per column.
	HeightmapWorldSurface int32 = 1

	// HeightmapMotionBlocking is the highest block per column that stops
	// movement or holds fluid.
	HeightmapMotionBlocking int32 = 4
)

// Heightmap is one heightmap of a chunk: 256 columns in ZX order, packed at
// however many bits the world height needs, values never crossing a long
// boundary.
type Heightmap struct {
	Type int32
	Data []int64
}

// LevelChunkWithLightClientboundPacket carries one whole chunk: its sections,
// its heightmaps, and all of its light. The client holds the chunk exactly as
// sent and asks for nothing more about it.
//
// The sections travel as one opaque byte field, already in wire form. They are
// built once when the world is loaded rather than on every send, because their
// encoding is the bulk of the packet and identical for every client on the
// same version -- and it is per version, since sections name block states by
// number and the versions number them differently. Whoever builds this packet
// has already chosen the numbering the receiving client reads by, which is why
// no transformer carries the sections between versions: everything that
// differs inside them is settled before the packet exists. The packet's own
// fields are another matter -- 1.21.4 reads the heightmaps in an older shape,
// and a transformer carries them there the way it would for any packet.
type LevelChunkWithLightClientboundPacket struct {
	// X and Z are in chunks.
	X, Z int32

	Heightmaps []Heightmap

	// SectionData is every section of the chunk in wire form, lowest first:
	// block count, fluid count, block states, biomes.
	SectionData []byte

	// The light, which 1.18 folded into this packet: a version before it
	// reads the same data in a packet of its own, sent ahead of the chunk.
	LightData
}

// LightData is all of a chunk's light as the wire carries it, which is the
// same data whether it travels inside the chunk packet, as it does from 1.18
// on, or in the light update packet of its own that 1.17.1 reads it in.
type LightData struct {
	// The light arrays, half a byte per block, 2048 bytes each, for the
	// sections that have any. Light spans one section beyond the blocks in
	// both directions, so the masks index sections from one below the world to
	// one above it: bit 0 is the section under the lowest blocks. A mask names
	// the sections whose arrays follow, in bit order; an empty mask names
	// sections known to hold no light at all, and a section in neither is one
	// nothing is said about.
	SkyLightMask, BlockLightMask           []int64
	EmptySkyLightMask, EmptyBlockLightMask []int64
	SkyLight, BlockLight                   [][]byte
}

func (p *LevelChunkWithLightClientboundPacket) String() string {
	return fmt.Sprintf("LevelChunkWithLightClientboundPacket{X:%d Z:%d Sections:%dB Sky:%d Block:%d}",
		p.X, p.Z, len(p.SectionData), len(p.SkyLight), len(p.BlockLight))
}

func (p *LevelChunkWithLightClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteInt(p.X); err != nil {
		return err
	}

	if err := ms.WriteInt(p.Z); err != nil {
		return err
	}

	if err := ms.WriteVarInt(int32(len(p.Heightmaps))); err != nil {
		return err
	}

	for _, heightmap := range p.Heightmaps {
		if err := ms.WriteVarInt(heightmap.Type); err != nil {
			return err
		}

		if err := writeLongArray(ms, heightmap.Data); err != nil {
			return err
		}
	}

	if err := ms.WriteByteArray(p.SectionData); err != nil {
		return err
	}

	// The block entities, of which none are sent. What a lobby's chests and
	// signs held is world data this server does not read, and a block entity
	// the client is not told about renders as the block alone.
	if err := ms.WriteVarInt(0); err != nil {
		return err
	}

	return p.LightData.encode(ms)
}

// encode writes the light the way both packets carry it: the four masks,
// then the sky arrays and the block arrays, each counted.
func (l *LightData) encode(ms *streams.MinecraftStream) error {
	for _, mask := range [][]int64{l.SkyLightMask, l.BlockLightMask, l.EmptySkyLightMask, l.EmptyBlockLightMask} {
		if err := writeLongArray(ms, mask); err != nil {
			return err
		}
	}

	for _, arrays := range [][][]byte{l.SkyLight, l.BlockLight} {
		if err := ms.WriteVarInt(int32(len(arrays))); err != nil {
			return err
		}

		for _, array := range arrays {
			if err := ms.WriteByteArray(array); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeLongArray writes a var int count and then the longs themselves, the
// form every counted long array of this packet takes.
func writeLongArray(ms *streams.MinecraftStream, values []int64) error {
	if err := ms.WriteVarInt(int32(len(values))); err != nil {
		return err
	}

	for _, value := range values {
		if err := ms.WriteLong(value); err != nil {
			return err
		}
	}

	return nil
}
