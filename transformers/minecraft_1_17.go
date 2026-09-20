package transformers

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.17 step is where the world grew downwards: 1.17 is the version that
// let a dimension type say where its world starts and how tall it is, and
// 1.16.4 below it holds a world of sixteen sections from zero up whatever it
// is told. Read off the 1.16.4 client's own classes through Mojang's
// mappings the way the steps above were, the two versions lay the login
// phase out alike, and of the play phase the login itself, the player list,
// the add player, the teleport, the head rotation, the animate, the entity
// metadata with the pose at 18, the keep alive, the chunk cache centre and
// everything this server reads. They number the play phase differently on
// both sides, which the id tables say. Six packets differ, all of them on
// the way down, and are below; the tags differ as well, and package
// gamedata encodes those as 1.16.4 reads them, since they are encoded per
// version to begin with.
//
// The play login is laid out alike, and carries 1.16.4's own registries and
// dimension type: 1.17 is where the dimension type took on its min_y and its
// height, and 1.16.4's logical height stops at 256.
//
// The chunk and its light are where the two versions part ways. 1.17 names
// the sections a chunk carries, and the sections its light does, by bit sets
// as long as the world is tall; 1.16.4 names them by a plain number, a bit
// for each of its sixteen sections and for each of the eighteen its light
// spans. This server's world is 1.18's, twenty-four sections from sixty-four
// below zero, so on the way down the four sections below zero and the four
// above 255 come off, and everything in between moves down four bits.

// sectionsBelowZero is how many of the sections this server sends sit below
// zero, which is how far every mask moves on its way to 1.16.4: the world is
// the one the dimension type in package gamedata announces, from sixty-four
// below zero, and 1.16.4's starts at zero.
const sectionsBelowZero = 4

const (
	// sections1_16_4 is how many sections a 1.16.4 chunk holds, and
	// lightSections1_16_4 how many its light spans: one more at either end.
	sections1_16_4      = 16
	lightSections1_16_4 = sections1_16_4 + 2

	// biomes1_16_4 is how many biomes a 1.16.4 chunk carries, one per four
	// blocks along each axis of its 256 blocks of height, which the client
	// reads as an array of exactly this many.
	biomes1_16_4 = 1024

	// heightmapBits is how many bits a heightmap entry takes on both sides of
	// the step -- enough for 256 and for 384 alike -- and heightmapEntries
	// how many entries a heightmap holds, one per column.
	heightmapBits    = 9
	heightmapEntries = 256

	// worldHeight1_16_4 is the most a 1.16.4 heightmap entry says.
	worldHeight1_16_4 = 256

	// lightArrayBytes is one section's light: four bits a block.
	lightArrayBytes = 2048
)

// DowngradePlayLoginTo1_16_4 rewrites the play phase login packet from what
// 1.17 sends into what 1.16.4 reads, given the registries that version reads
// out of it and the dimension type it spells out in it.
//
// The two versions lay the packet out alike from front to back. The
// registries in the middle are 1.17's, put there by the 1.18 step, and the
// dimension type spelled out behind them 1.17's as well; 1.16.4 has its own
// dimension type, so both are read so that they are consumed and never
// written, and 1.16.4's go in their place. The rest is copied. As with the
// steps above, a transformer built with no registries or no dimension type to
// write refuses every login rather than send a client into a world it cannot
// make sense of.
func DowngradePlayLoginTo1_16_4(registryCodec []byte, dimensionType []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.16.4 play login carries the registries, and this transformer was given none")
		}

		if len(dimensionType) == 0 {
			return errors.New("a 1.16.4 play login spells out the dimension type, and this transformer was given none")
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

		// 1.17's registries, and 1.16.4's in their place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(registryCodec); err != nil {
			return err
		}

		// 1.17's dimension type, and 1.16.4's in its place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(dimensionType); err != nil {
			return err
		}

		// The dimension, the seed, max players, the view distance and the
		// four flags, laid out alike on both sides of the step.
		return copyRest(in, out)
	}
}

// DowngradeLevelChunkTo1_16_4 rewrites the chunk packet from what 1.17 sends
// into what 1.16.4 reads.
//
// 1.16.4 lays the packet out as the coordinates, a flag saying the chunk is
// whole rather than a change to one the client holds, a mask of the sections
// that follow as a plain number, the heightmaps, the biomes of a whole chunk
// -- which is what the flag promises -- the sections the mask names, and the
// block entities. The sections themselves are laid out alike in the two
// versions, so the ones 1.16.4 has room for travel as they were, and the
// ones below zero and above 255 come off; the mask moves down with them. The
// heightmaps count from the bottom of the world, which is sixty-four lower
// on 1.17, so every entry comes down by that much, and stops at either end
// of what 1.16.4 holds. The biomes are the first and only biome this server
// registers, as on 1.17, for as many entries as 1.16.4 reads.
//
// The block entities are refused rather than carried, as the 1.18 step
// refuses them: this server sends none.
func DowngradeLevelChunkTo1_16_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	mask, err := readBitSetLong(in)
	if err != nil {
		return err
	}

	name, heightmaps, err := nbt.ReadNamed(in)
	if err != nil {
		return err
	}

	// 1.17's biomes, read so that they are consumed and never written.
	biomeCount, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	for range biomeCount {
		if _, err := in.ReadVarInt(); err != nil {
			return err
		}
	}

	sectionData, err := in.ReadByteArray(streams.MaxPacketSize)
	if err != nil {
		return err
	}

	sections, err := cutSectionsTo1_16_4(sectionData, mask)
	if err != nil {
		return err
	}

	blockEntityCount, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if blockEntityCount != 0 {
		return fmt.Errorf("this server sends no block entities, and the packet carries %d", blockEntityCount)
	}

	lowered, err := lowerHeightmapsTo1_16_4(heightmaps)
	if err != nil {
		return err
	}

	// The whole chunk flag: every chunk this server sends is one.
	if err := out.WriteBoolean(true); err != nil {
		return err
	}

	if err := out.WriteVarInt(int32(mask >> sectionsBelowZero & (1<<sections1_16_4 - 1))); err != nil {
		return err
	}

	if err := nbt.WriteNamed(out, name, lowered); err != nil {
		return err
	}

	if err := out.WriteVarInt(biomes1_16_4); err != nil {
		return err
	}

	for range biomes1_16_4 {
		if err := out.WriteVarInt(0); err != nil {
			return err
		}
	}

	if err := out.WriteByteArray(sections); err != nil {
		return err
	}

	return out.WriteVarInt(0)
}

// cutSectionsTo1_16_4 walks a 1.17 section buffer, which holds the sections
// its mask names from the bottom of the world up, and returns the ones
// 1.16.4 has room for as they were.
func cutSectionsTo1_16_4(data []byte, mask int64) ([]byte, error) {
	in := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(data))

	kept := make([]byte, 0, len(data))

	for index := 0; index < sectionsPerMask; index++ {
		if mask&(1<<index) == 0 {
			continue
		}

		blockCount, err := in.ReadShort()
		if err != nil {
			return nil, fmt.Errorf("section %d: %w", index, err)
		}

		blocks, err := readContainer(in, blockIndirectBits)
		if err != nil {
			return nil, fmt.Errorf("section %d: %w", index, err)
		}

		if index < sectionsBelowZero || index >= sectionsBelowZero+sections1_16_4 {
			continue
		}

		kept = append(kept, byte(blockCount>>8), byte(blockCount))
		kept = append(kept, blocks.raw...)
	}

	if rest, err := in.ReadRest(); err != nil {
		return nil, err
	} else if len(rest) != 0 {
		return nil, fmt.Errorf("%d bytes of sections past the ones the mask names", len(rest))
	}

	return kept, nil
}

// lowerHeightmapsTo1_16_4 brings every entry of every heightmap down to where
// 1.16.4 counts from. An entry is how far above the bottom of the world the
// first free block of its column is, packed nine bits to an entry with no
// entry crossing a long, on both sides of the step.
func lowerHeightmapsTo1_16_4(heightmaps nbt.Tag) (nbt.Tag, error) {
	compound, ok := heightmaps.(nbt.Compound)
	if !ok {
		return nil, fmt.Errorf("the heightmaps are a %s, not a compound", heightmaps.Type())
	}

	const entriesPerLong = 64 / heightmapBits
	const longs = (heightmapEntries + entriesPerLong - 1) / entriesPerLong
	const entryMask = 1<<heightmapBits - 1

	lowered := make(nbt.Compound, len(compound))

	for kind, tag := range compound {
		packed, ok := tag.(nbt.LongArray)
		if !ok {
			return nil, fmt.Errorf("the %s heightmap is a %s, not a long array", kind, tag.Type())
		}

		if len(packed) != longs {
			return nil, fmt.Errorf("the %s heightmap packs %d longs, want %d", kind, len(packed), longs)
		}

		out := make(nbt.LongArray, longs)

		for i := range heightmapEntries {
			shift := i % entriesPerLong * heightmapBits

			height := packed[i/entriesPerLong]>>shift&entryMask - sectionsBelowZero*16
			height = min(max(height, 0), worldHeight1_16_4)

			out[i/entriesPerLong] |= height << shift
		}

		lowered[kind] = out
	}

	return lowered, nil
}

// DowngradeLightUpdateTo1_16_4 rewrites the light update packet from what
// 1.17 sends into what 1.16.4 reads.
//
// The two lay it out alike up to the masks: the chunk coordinates and the
// trust edges flag. 1.17 writes each of the four masks as a bit set and each
// of the two runs of arrays behind a count; 1.16.4 writes each mask as a
// plain number of eighteen bits, from the section below its world to the one
// above it, and the arrays with no count, since the masks say how many
// follow. The masks move down by the sections this server sends below zero,
// and the arrays of the sections that fall off either end come off with
// them.
func DowngradeLightUpdateTo1_16_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates.
	for range 2 {
		if _, err := copyVarInt(in, out); err != nil {
			return err
		}
	}

	// The trust edges flag.
	if _, err := copyBoolean(in, out); err != nil {
		return err
	}

	var masks [4]int64

	for i := range masks {
		mask, err := readBitSetLong(in)
		if err != nil {
			return err
		}

		masks[i] = mask

		if err := out.WriteVarInt(int32(mask >> sectionsBelowZero & (1<<lightSections1_16_4 - 1))); err != nil {
			return err
		}
	}

	// The sky light's arrays, then the block light's, each run named by the
	// first two masks in that order.
	for _, mask := range masks[:2] {
		if err := cutLightArraysTo1_16_4(in, out, mask); err != nil {
			return err
		}
	}

	return nil
}

// cutLightArraysTo1_16_4 reads one counted run of light arrays, one for each
// bit of mask from the bottom up, and writes the ones 1.16.4 has a section
// for.
func cutLightArraysTo1_16_4(in *streams.MinecraftStream, out *streams.MinecraftStream, mask int64) error {
	count, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	for index := 0; index < sectionsPerMask; index++ {
		if mask&(1<<index) == 0 {
			continue
		}

		if count--; count < 0 {
			return errors.New("the light mask names more sections than the packet carries arrays")
		}

		array, err := in.ReadByteArray(lightArrayBytes)
		if err != nil {
			return err
		}

		if index < sectionsBelowZero || index >= sectionsBelowZero+lightSections1_16_4 {
			continue
		}

		if err := out.WriteByteArray(array); err != nil {
			return err
		}
	}

	if count != 0 {
		return fmt.Errorf("the packet carries %d light arrays the mask does not name", count)
	}

	return nil
}

// readBitSetLong reads a bit set as every version from 1.17 to 26.2 writes
// it, a count of longs and the longs, and returns the first, which covers
// the sixty-four sections no dimension type reaches. An empty bit set is
// written as no longs at all.
func readBitSetLong(in *streams.MinecraftStream) (int64, error) {
	count, err := in.ReadVarInt()
	if err != nil {
		return 0, err
	}

	if count < 0 || count > 1 {
		return 0, fmt.Errorf("a section mask of %d longs", count)
	}

	if count == 0 {
		return 0, nil
	}

	return in.ReadLong()
}

// DowngradePlayerPositionTo1_16_4 rewrites the player position packet from
// what 1.17 sends into what 1.16.4 reads: 1.17 is where the dismount vehicle
// flag went onto its end, which the 1.19.4 step puts back for the versions
// that read it, and it comes off again here.
func DowngradePlayerPositionTo1_16_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The position, three doubles, the rotation, two floats, and the
	// relative flags, a byte.
	if err := copyBytes(in, out, 3*8+2*4+1); err != nil {
		return err
	}

	// The teleport id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	// The dismount vehicle flag, read so that it is consumed and never
	// written.
	_, err := in.ReadBoolean()

	return err
}

// DowngradeSetDefaultSpawnPositionTo1_16_4 rewrites the default spawn
// position packet -- which is what this server's game event has been since
// the 1.20.3 step -- from what 1.17 sends into what 1.16.4 reads: 1.17 is
// where the angle went onto its end, behind the position, and 1.16.4 reads
// the position alone.
func DowngradeSetDefaultSpawnPositionTo1_16_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The position, packed into a long.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	// The angle, read so that it is consumed and never written.
	_, err := in.ReadFloat()

	return err
}

// DowngradeRemoveEntitiesTo1_16_4 rewrites the remove entities packet from
// what 1.17 sends into what 1.16.4 reads. 1.17 is the one version that takes
// an entity id alone, which the 1.17.1 step cut the packet down to; 1.16.4
// reads a count and that many ids, as every version from 1.17.1 on does, so
// the count goes back in front.
func DowngradeRemoveEntitiesTo1_16_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := out.WriteVarInt(1); err != nil {
		return err
	}

	_, err := copyVarInt(in, out)

	return err
}
