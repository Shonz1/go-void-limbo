package transformers

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.16 step is where the dimensions became data. 1.15.2 below it holds
// every dimension in its own code, as it holds every biome, and is put into
// one of them by number; 1.16 reads the dimension types out of its play login
// and is put into one by name. Read off the 1.15.2 client's own classes
// through Mojang's mappings the way the steps above were, the two versions
// lay out alike the whole login phase but for its last packet, everything
// this server reads, and of what it sends the tags, the player list, the add
// player, the teleport, the head rotation, the animate, the entity metadata
// with the pose at 18, the keep alive, the player position, the entity
// removal, the chunk cache centre and the spawn position. They number the
// play phase differently on both sides, which the id tables say. Four packets
// differ, all of them on the way down, and are below.
//
// 1.16 is also where a packed entry stopped crossing from one long into the
// next: 1.15.2 packs the entries of a section's blocks and of a heightmap end
// to end, the bits of one carrying on into the next long wherever a long runs
// out, where 1.16 starts a new long instead and leaves the bits it could not
// fill. The two pack alike only where the bits of an entry divide a long.

const (
	// defaultLevelType1_15_2 and flatLevelType1_15_2 are the two level types
	// a 1.15.2 play login tells apart for what this server says of a world:
	// 1.16 says a world is flat with a flag, and 1.15.2 by naming its
	// generator.
	defaultLevelType1_15_2 = "default"
	flatLevelType1_15_2    = "flat"

	// sectionBlocks is how many blocks a section holds, which is how many
	// entries its container packs.
	sectionBlocks = 4096
)

// dimensions1_15_2 is the dimensions a 1.15.2 client numbers for itself, by
// the names the versions above it know them by.
var dimensions1_15_2 = map[string]int32{
	"minecraft:the_nether": -1,
	"minecraft:overworld":  0,
	"minecraft:the_end":    1,
}

// DowngradeLoginSuccessTo1_15_2 rewrites the login success packet from what
// 1.16 sends into what 1.15.2 reads.
//
// 1.16 writes the uuid as its sixteen bytes; 1.15.2 reads it as text, the way
// java.util.UUID spells one out, and turns it back into a uuid. The name
// behind it is laid out alike.
func DowngradeLoginSuccessTo1_15_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	uuid, err := in.ReadUuid()
	if err != nil {
		return err
	}

	if err := out.WriteString(uuid); err != nil {
		return err
	}

	return copyRest(in, out)
}

// DowngradePlayLoginTo1_15_2 rewrites the play phase login packet from what
// 1.16 sends into what 1.15.2 reads.
//
// 1.15.2 lays the packet out as the entity id, the game mode with the
// hardcore flag folded into its byte as 1.16 folds it, the dimension as a
// number, the seed, the most players, the level type, the view distance and
// two flags. It has no previous game mode, no dimension names, no dimension
// types and no dimension type name, and its last two flags -- whether the
// world is a debug one and whether it is flat -- are gone, the second of them
// into the level type. The dimension is one of the three 1.15.2 holds or it
// is refused, since there is no number to put a client into another by; its
// type is not told apart from it, the two being one thing to 1.15.2.
func DowngradePlayLoginTo1_15_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id, a plain int, and the game mode's byte.
	if err := copyBytes(in, out, 5); err != nil {
		return err
	}

	// The previous game mode.
	if _, err := in.ReadByte(); err != nil {
		return err
	}

	// The dimension names.
	dimensionCount, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	for range dimensionCount {
		if _, err := in.ReadString(); err != nil {
			return err
		}
	}

	// 1.16's dimension types, read so that they are consumed and never
	// written.
	if _, _, err := nbt.ReadNamed(in); err != nil {
		return err
	}

	// The name of the dimension type, which 1.15.2 does not tell from the
	// dimension.
	if _, err := in.ReadString(); err != nil {
		return err
	}

	dimension, err := in.ReadString()
	if err != nil {
		return err
	}

	number, ok := dimensions1_15_2[dimension]
	if !ok {
		return fmt.Errorf("dimension %s is none of the three 1.15.2 holds, and it has no number for it", dimension)
	}

	if err := out.WriteInt(number); err != nil {
		return err
	}

	// The seed, a long, and the most players, a byte.
	if err := copyBytes(in, out, 9); err != nil {
		return err
	}

	viewDistance, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	reducedDebugInfo, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	showDeathScreen, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	// Whether the world is a debug one, which 1.15.2 is not told.
	if _, err := in.ReadBoolean(); err != nil {
		return err
	}

	flat, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	levelType := defaultLevelType1_15_2
	if flat {
		levelType = flatLevelType1_15_2
	}

	if err := out.WriteString(levelType); err != nil {
		return err
	}

	if err := out.WriteVarInt(viewDistance); err != nil {
		return err
	}

	if err := out.WriteBoolean(reducedDebugInfo); err != nil {
		return err
	}

	return out.WriteBoolean(showDeathScreen)
}

// DowngradeLevelChunkTo1_15_2 rewrites the chunk packet from what 1.16 sends
// into what 1.15.2 reads.
//
// 1.15.2 has no second flag behind the whole chunk flag: 1.16 put it there
// and 1.16.2 took it off again. Its heightmaps and the blocks of its sections
// pack their entries across the ends of their longs, where 1.16 starts a new
// long, so both are packed over again. The biomes are laid out alike, as many
// plain ints as a chunk holds out of the client's own biomes, and so are the
// coordinates, the mask and the block entities.
func DowngradeLevelChunkTo1_15_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	wholeChunk, err := copyBoolean(in, out)
	if err != nil {
		return err
	}

	// The forget old data flag.
	if _, err := in.ReadBoolean(); err != nil {
		return err
	}

	mask, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	name, heightmaps, err := nbt.ReadNamed(in)
	if err != nil {
		return err
	}

	spanned, err := spanHeightmapsTo1_15_2(heightmaps)
	if err != nil {
		return err
	}

	if err := nbt.WriteNamed(out, name, spanned); err != nil {
		return err
	}

	// Only a whole chunk carries its biomes, on both sides of the step.
	if wholeChunk {
		if err := copyBytes(in, out, biomes1_16_4*4); err != nil {
			return err
		}
	}

	sectionData, err := in.ReadByteArray(streams.MaxPacketSize)
	if err != nil {
		return err
	}

	sections, err := spanSectionsTo1_15_2(sectionData, mask)
	if err != nil {
		return err
	}

	if err := out.WriteByteArray(sections); err != nil {
		return err
	}

	// The block entities.
	return copyRest(in, out)
}

// spanHeightmapsTo1_15_2 packs every heightmap over again the way 1.15.2
// reads one: nine bits to an entry on both sides of the step, thirty-seven
// longs with no entry crossing a long on 1.16 and thirty-six with the entries
// end to end on 1.15.2.
func spanHeightmapsTo1_15_2(heightmaps nbt.Tag) (nbt.Tag, error) {
	compound, ok := heightmaps.(nbt.Compound)
	if !ok {
		return nil, fmt.Errorf("the heightmaps are a %s, not a compound", heightmaps.Type())
	}

	spanned := make(nbt.Compound, len(compound))

	for kind, tag := range compound {
		packed, ok := tag.(nbt.LongArray)
		if !ok {
			return nil, fmt.Errorf("the %s heightmap is a %s, not a long array", kind, tag.Type())
		}

		out, err := spanEntries(packed, heightmapBits, heightmapEntries)
		if err != nil {
			return nil, fmt.Errorf("the %s heightmap: %w", kind, err)
		}

		spanned[kind] = nbt.LongArray(out)
	}

	return spanned, nil
}

// spanSectionsTo1_15_2 walks a 1.16 section buffer, which holds the sections
// its mask names from the bottom of the world up, and returns it with the
// blocks of every section packed the way 1.15.2 reads them. A section is laid
// out alike otherwise: the block count, the bits of an entry, the palette
// where the bits call for one, and the longs behind their count.
func spanSectionsTo1_15_2(data []byte, mask int32) ([]byte, error) {
	in := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(data))

	buf := bytes.NewBuffer(make([]byte, 0, len(data)))
	out := streams.NewMinecraftStreamFromBuffer(buf)

	for index := 0; index < sections1_16_4; index++ {
		if mask&(1<<index) == 0 {
			continue
		}

		if err := spanSectionTo1_15_2(in, out); err != nil {
			return nil, fmt.Errorf("section %d: %w", index, err)
		}
	}

	if rest, err := in.ReadRest(); err != nil {
		return nil, fmt.Errorf("past the last section: %w", err)
	} else if len(rest) != 0 {
		return nil, fmt.Errorf("%d bytes of sections past the ones the mask names", len(rest))
	}

	if err := out.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func spanSectionTo1_15_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The block count, a short.
	if err := copyBytes(in, out, 2); err != nil {
		return err
	}

	bits, err := in.ReadByte()
	if err != nil {
		return err
	}

	// A palette of one is a form neither side of this step reads: the 1.18
	// step spells it out before a section gets here.
	if bits == 0 || bits > 32 {
		return fmt.Errorf("a container of %d bits to an entry, which 1.15.2 has no form for", bits)
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
	}

	longCount, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if longCount < 0 || longCount > streams.MaxPacketSize/8 {
		return fmt.Errorf("a container of %d longs", longCount)
	}

	packed := make([]int64, longCount)
	for i := range packed {
		if packed[i], err = in.ReadLong(); err != nil {
			return err
		}
	}

	spanned, err := spanEntries(packed, int(bits), sectionBlocks)
	if err != nil {
		return err
	}

	if err := out.WriteVarInt(int32(len(spanned))); err != nil {
		return err
	}

	for _, long := range spanned {
		if err := out.WriteLong(long); err != nil {
			return err
		}
	}

	return nil
}

// spanEntries packs entries over again from the way 1.16 packs them into the
// way 1.15.2 does. 1.16 fits as many whole entries into a long as it has room
// for and starts the next entry in the next long; 1.15.2 lays the entries end
// to end, so that one may start in one long and end in the next, in as many
// longs as their bits come to.
func spanEntries(packed []int64, bits int, entries int) ([]int64, error) {
	if bits <= 0 || bits > 32 {
		return nil, errors.New("an entry takes between one bit and thirty-two")
	}

	entriesPerLong := 64 / bits

	if want := (entries + entriesPerLong - 1) / entriesPerLong; len(packed) != want {
		return nil, fmt.Errorf("%d entries of %d bits pack %d longs, want %d", entries, bits, len(packed), want)
	}

	entryMask := uint64(1)<<bits - 1

	spanned := make([]int64, (entries*bits+63)/64)

	for i := range entries {
		entry := uint64(packed[i/entriesPerLong]) >> (i % entriesPerLong * bits) & entryMask

		at := i * bits
		long, shift := at/64, at%64

		spanned[long] |= int64(entry << shift)

		if shift+bits > 64 {
			spanned[long+1] |= int64(entry >> (64 - shift))
		}
	}

	return spanned, nil
}

// DowngradeLightUpdateTo1_15_2 rewrites the light update packet from what
// 1.16 sends into what 1.15.2 reads: 1.16 put a trust edges flag behind the
// chunk coordinates, which 1.15.2 does not have. The masks and the arrays
// behind it are laid out alike.
func DowngradeLightUpdateTo1_15_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two varints.
	for range 2 {
		if _, err := copyVarInt(in, out); err != nil {
			return err
		}
	}

	// The trust edges flag.
	if _, err := in.ReadBoolean(); err != nil {
		return err
	}

	return copyRest(in, out)
}
