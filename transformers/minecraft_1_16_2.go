package transformers

import (
	"errors"
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.16.2 step is where the biomes became data. 1.16.1 below it holds
// every biome in its own code and numbers them there, is sent the dimension
// types alone, and is put into one of them by name; 1.16.2 reads the biomes
// and the dimension types out of its play login as one compound, and the
// dimension type it is put into spelled out behind it. Read off the 1.16.1
// client's own classes through Mojang's mappings the way the steps above
// were, the two versions lay out alike the whole login phase, everything
// this server reads, and of what it sends the light, the tags, the player
// list, the add player, the teleport, the head rotation, the animate, the
// entity metadata, the keep alive, the player position, the entity removal,
// the chunk cache centre and the spawn position, whose angle 1.16.2 holds
// and does not write. They number the play phase differently on both sides,
// which the id tables say. Two packets differ, both on the way down, and
// are below.

const (
	// hardcoreBit1_16_1 is where a 1.16.1 play login holds the hardcore flag:
	// the fourth bit of the game mode's byte.
	hardcoreBit1_16_1 = 8

	// maxPlayers1_16_1 is the most a 1.16.1 play login says of how many
	// players a server holds, which it says in one byte.
	maxPlayers1_16_1 = 255

	// plainsBiome1_16_1 is the plains among the biomes a 1.16.1 client
	// numbers for itself, behind the ocean at zero.
	plainsBiome1_16_1 = 1
)

// DowngradePlayLoginTo1_16_1 rewrites the play phase login packet from what
// 1.16.2 sends into what 1.16.1 reads, given the dimension types that version
// reads out of it and the name of the one it is put into, as the string the
// packet holds.
//
// 1.16.1 folds the hardcore flag into the game mode's byte, where 1.16.2
// holds a flag of its own in front of it, and says how many players the
// server holds in a byte, where 1.16.2 says it in a varint. The registries in
// the middle are 1.16.2's, put there by the 1.17 step, and the dimension type
// spelled out behind them 1.16.2's as well; both are read so that they are
// consumed and never written, and 1.16.1's list and its name go in their
// place. The rest is copied. As with the steps above, a transformer built
// with no dimension types or no name to write refuses every login rather
// than send a client into a world it cannot make sense of.
func DowngradePlayLoginTo1_16_1(dimensionTypes []byte, dimensionTypeName []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(dimensionTypes) == 0 {
			return errors.New("a 1.16.1 play login carries the dimension types, and this transformer was given none")
		}

		if len(dimensionTypeName) == 0 {
			return errors.New("a 1.16.1 play login names the dimension type, and this transformer was given no name")
		}

		// The entity id, a plain int.
		if err := copyBytes(in, out, 4); err != nil {
			return err
		}

		hardcore, err := in.ReadBoolean()
		if err != nil {
			return err
		}

		gameMode, err := in.ReadByte()
		if err != nil {
			return err
		}

		if gameMode >= hardcoreBit1_16_1 {
			return fmt.Errorf("game mode %d does not fit under the hardcore bit of a 1.16.1 play login", gameMode)
		}

		if hardcore {
			gameMode |= hardcoreBit1_16_1
		}

		if err := out.WriteByte(gameMode); err != nil {
			return err
		}

		// The previous game mode, a byte on both sides.
		if err := copyBytes(in, out, 1); err != nil {
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

		// 1.16.2's registries, and 1.16.1's dimension types in their place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(dimensionTypes); err != nil {
			return err
		}

		// 1.16.2's dimension type, and the name of 1.16.1's in its place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(dimensionTypeName); err != nil {
			return err
		}

		// The dimension.
		if err := copyString(in, out); err != nil {
			return err
		}

		// The seed, a long.
		if err := copyBytes(in, out, 8); err != nil {
			return err
		}

		maxPlayers, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		if err := out.WriteByte(byte(min(max(maxPlayers, 0), maxPlayers1_16_1))); err != nil {
			return err
		}

		// The view distance and the four flags, laid out alike on both sides
		// of the step.
		return copyRest(in, out)
	}
}

// DowngradeLevelChunkTo1_16_1 rewrites the chunk packet from what 1.16.2
// sends into what 1.16.1 reads.
//
// 1.16.1 holds a second flag behind the whole chunk flag, saying whether the
// client forgets what it held of the chunk before, which a vanilla server
// sets on every whole chunk it sends and 1.16.2 took off the packet. Its
// biomes are as many plain ints as a chunk holds, with no count in front,
// where 1.16.2 counts them and writes each as a varint; and they are numbers
// out of the client's own biomes rather than out of a registry it was sent,
// so the one biome this server registers goes out as the client's plains.
// Everything else is copied: the coordinates, the mask, the heightmaps, the
// sections and the block entities are laid out alike.
func DowngradeLevelChunkTo1_16_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	wholeChunk, err := copyBoolean(in, out)
	if err != nil {
		return err
	}

	// The forget old data flag, which goes with a whole chunk.
	if err := out.WriteBoolean(wholeChunk); err != nil {
		return err
	}

	// The mask of the sections that follow.
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
	if wholeChunk {
		biomeCount, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		if biomeCount != biomes1_16_4 {
			return fmt.Errorf("the chunk carries %d biomes, and 1.16.1 reads %d", biomeCount, biomes1_16_4)
		}

		for range biomeCount {
			biome, err := in.ReadVarInt()
			if err != nil {
				return err
			}

			if biome != 0 {
				return fmt.Errorf("biome %d is not the one this server registers, and 1.16.1 has no number for it", biome)
			}

			if err := out.WriteInt(plainsBiome1_16_1); err != nil {
				return err
			}
		}
	}

	// The sections and the block entities.
	return copyRest(in, out)
}
