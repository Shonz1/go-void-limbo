package transformers

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// DowngradeAddEntityTo1_21_4 rewrites the add entity packet from what 1.21.5
// sends into what 1.21.4 reads.
//
// The packet's shape is identical in the two, and what moved is the entity
// type registry behind one of its fields: 1.21.5 split the potion entity into
// a splash and a lingering one, and both sort before the player, so its
// number went up by one.
func DowngradeAddEntityTo1_21_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType1_21_5, playerEntityType1_21_4)
}

// heightmapNames is what 1.21.4 calls each heightmap kind: the key it is
// stored under in the compound, which is the name of the client's own enum
// constant. Only the two kinds this server sends are named, since a kind it
// does not send is one this rewrite was never taught.
var heightmapNames = map[int32]string{
	1: "WORLD_SURFACE",
	4: "MOTION_BLOCKING",
}

// DowngradeLevelChunkWithLightTo1_21_4 rewrites the chunk packet from what
// 1.21.5 sends into what 1.21.4 reads.
//
// 1.21.5 turned the chunk's heightmaps from an NBT compound -- one long array
// per kind, keyed by the kind's name -- into a counted map of kind number to
// long array. That is the one field of the packet whose shape moved, and it
// sits right after the chunk coordinates, so the rewrite reads the map,
// writes the compound, and copies everything after it untouched.
//
// The section data behind it is not this transformer's concern even though
// 1.21.4 reads it differently too: sections are built per version by package
// world before the packet exists, the way every difference inside them is.
func DowngradeLevelChunkWithLightTo1_21_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	count, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	heightmaps := make(nbt.Compound, count)

	for range count {
		kind, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		name, ok := heightmapNames[kind]
		if !ok {
			return fmt.Errorf("chunk carries heightmap kind %d, which this rewrite does not know the name of", kind)
		}

		length, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		data := make(nbt.LongArray, length)
		for i := range data {
			if data[i], err = in.ReadLong(); err != nil {
				return err
			}
		}

		heightmaps[name] = data
	}

	if err := nbt.Write(out, heightmaps); err != nil {
		return err
	}

	return copyRest(in, out)
}
