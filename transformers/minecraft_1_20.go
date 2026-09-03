package transformers

import (
	"errors"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.20 step is the smallest of the ones below the configuration phase.
// 1.20 and 1.19.4 number every packet this server speaks alike, in every
// phase, and lay all but two of them out alike, read off the 1.19.4 client's
// own classes through Mojang's mappings the way the 1.20.2 step was: the
// play login, to whose end 1.20 appended the portal cooldown, and the chunk
// packet, whose light data 1.19.4 opens with a trust edges flag that 1.20
// took off. Each rewrite below is one of those two. Everything else is
// wire-identical in 762 and 763 -- the hello with its optional uuid, the
// login on both sides of the success packet, the add player packet, the
// player list, the movement on both sides, the tags, the entity metadata
// with the pose at 20, the sections inside the chunk packet -- and 1.19.4
// numbers the player the same 122 in its entity type registry, so the add
// player rewrite of the 1.20.2 step goes out as it is.
//
// The registries a 1.19.4 client reads out of its play login are the same
// compound 1.20 reads, with 1.19.4's own content in it, and the shape of the
// login around them is 1.20's but for the cooldown at the end. So the login
// rewrite below swaps the compound for the one package gamedata encodes for
// 1.19.4 rather than laying the packet out afresh: everything in front of
// the compound and behind it is copied.

// DowngradePlayLoginTo1_19_4 rewrites the play phase login packet from what
// 1.20 sends into what 1.19.4 reads, given the registries that version reads
// out of it.
//
// The packet is 1.20's from front to back, but for two things. The
// registries in the middle are 1.20's, put there by the 1.20.2 step, and
// 1.19.4 has its own: the compound is read whole so that it is consumed, and
// the one package gamedata encodes for 1.19.4 goes out in its place. And the
// portal cooldown at the end is 1.20's alone, which 1.19.4 has no field for,
// so it comes off. As with the 1.20.2 step's rewrite, a transformer built
// with no registries to write refuses every login rather than send a client
// into a world it cannot make sense of.
func DowngradePlayLoginTo1_19_4(registryCodec []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.19.4 play login carries the registries, and this transformer was given none")
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

		// 1.20's registries, read so that they are consumed and never
		// written, and 1.19.4's in their place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(registryCodec); err != nil {
			return err
		}

		// The dimension type and the dimension.
		for range 2 {
			if err := copyString(in, out); err != nil {
				return err
			}
		}

		// The seed.
		if err := copyBytes(in, out, 8); err != nil {
			return err
		}

		// Max players, view distance, simulation distance.
		for range 3 {
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		}

		// Reduced debug info, show death screen, is debug, is flat.
		for range 4 {
			if _, err := copyBoolean(in, out); err != nil {
				return err
			}
		}

		// The death location, a flag and then the dimension name and the
		// packed position when the flag is set.
		hasDeathLocation, err := copyBoolean(in, out)
		if err != nil {
			return err
		}

		if hasDeathLocation {
			if err := copyString(in, out); err != nil {
				return err
			}

			if err := copyBytes(in, out, 8); err != nil {
				return err
			}
		}

		// The portal cooldown, read so that it is consumed and never written.
		_, err = in.ReadVarInt()

		return err
	}
}

// DowngradeLevelChunkWithLightTo1_19_4 rewrites the chunk packet from what
// 1.20 sends into what 1.19.4 reads.
//
// 1.20 is where the light data lost its trust edges flag, the first thing it
// used to be written with. A 1.19.4 client reads that flag in front of the
// four masks, and a vanilla server of that version sets it for every chunk it
// sends, so the rewrite copies everything up to the light data -- the
// coordinates, the heightmaps, the sections, the block entities, whose NBT
// is named in both versions -- puts the flag in, and copies the light data
// after it untouched.
func DowngradeLevelChunkWithLightTo1_19_4(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	// The heightmaps, a compound named as a root on both sides of this step.
	name, heightmaps, err := nbt.ReadNamed(in)
	if err != nil {
		return err
	}

	if err := nbt.WriteNamed(out, name, heightmaps); err != nil {
		return err
	}

	// The section data, a counted byte array.
	if err := copyByteArray(in, out); err != nil {
		return err
	}

	// The block entities: a count, then for each a packed position byte,
	// a y short, a type and a named compound.
	blockEntityCount, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	for range blockEntityCount {
		if err := copyBytes(in, out, 3); err != nil {
			return err
		}

		if _, err := copyVarInt(in, out); err != nil {
			return err
		}

		name, data, err := nbt.ReadNamed(in)
		if err != nil {
			return err
		}

		if err := nbt.WriteNamed(out, name, data); err != nil {
			return err
		}
	}

	if err := out.WriteBoolean(trustEdges); err != nil {
		return err
	}

	return copyRest(in, out)
}

// trustEdges is what a vanilla 1.19.4 server says about the light data of
// every chunk it sends: that the light at the chunk's edges is settled, and
// the client need not recompute it against the neighbours.
const trustEdges = true
