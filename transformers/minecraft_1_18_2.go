package transformers

import (
	"errors"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.18.2 step is where tag keys arrived, and that is the whole of what
// the two versions lay out differently, read off the 1.18 client's own
// classes through Mojang's mappings the way the 1.19 step was. 1.18.2 and
// 1.18 lay every packet this server speaks out alike -- the login phase
// whole, the play login field for field, the player list, the player
// position, the movement on both sides, the entity metadata with the pose
// at 18, the tags, the spawn position and the chunk packet -- and number
// them alike as well, in every phase and both directions. What differs is
// inside one field of the play login: the dimension type it spells out,
// whose infiniburn field 1.18.2 reads as a hashed tag key and 1.18 as a
// plain name, and the registries alongside it, which carry the same
// dimension type again. Both are 1.18's own, encoded once by package
// gamedata, and the one rewrite below swaps them in.

// DowngradePlayLoginTo1_18 rewrites the play phase login packet from what
// 1.18.2 sends into what 1.18 reads, given the registries that version reads
// out of it and the dimension type it spells out in it.
//
// The two versions lay the packet out alike from front to back: the
// registries in the middle are 1.18.2's, put there by the 1.19 step, and
// the dimension type spelled out behind them 1.18.2's as well, and 1.18 has
// its own of each, differing by the one field. So both are read so that
// they are consumed and never written, and 1.18's go in their place; the
// rest is copied. As with the steps above, a transformer built with no
// registries or no dimension type to write refuses every login rather than
// send a client into a world it cannot make sense of.
func DowngradePlayLoginTo1_18(registryCodec []byte, dimensionType []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.18 play login carries the registries, and this transformer was given none")
		}

		if len(dimensionType) == 0 {
			return errors.New("a 1.18 play login spells out the dimension type, and this transformer was given none")
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

		// 1.18.2's registries, and 1.18's in their place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(registryCodec); err != nil {
			return err
		}

		// 1.18.2's dimension type, and 1.18's in its place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(dimensionType); err != nil {
			return err
		}

		// The dimension, the seed, the three distances and the four flags,
		// laid out alike on both sides of the step.
		return copyRest(in, out)
	}
}
