package transformers

import (
	"errors"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.19.4 step is a small one, read off the 1.19.3 client's own classes
// through Mojang's mappings the way the 1.20 step was. 1.19.4 and 1.19.3 lay
// all but three of the packets this server speaks out alike -- the play
// login, whose registries are 1.19.3's own; the player position, to whose
// end 1.19.3 puts a flag 1.19.4 took off; and the entity metadata, whose
// pose serializer 1.19.4's optional block state serializer pushed up one --
// and each rewrite below is one of those three. Everything else is
// wire-identical in 761 and 762: the hello with its optional uuid, the login
// on both sides of the success packet, the add player packet, the player
// list with its six actions, the movement on both sides, the tags, the
// chunk packet with its named heightmaps and its trust edges flag. 1.19.4
// did renumber the play phase, in both directions, which the id tables say;
// this file is only about the bodies.

// DowngradePlayLoginTo1_19_3 rewrites the play phase login packet from what
// 1.19.4 sends into what 1.19.3 reads, given the registries that version
// reads out of it.
//
// The two versions lay the packet out alike from front to back, so the one
// thing to change is the registries in the middle, which are 1.19.4's, put
// there by the 1.20 step, and which 1.19.3 has its own of: three registries
// to 1.19.4's six, since 1.19.4 is where the damage types and the armor
// trims appeared. The compound is read whole so that it is consumed, and the
// one package gamedata encodes for 1.19.3 goes out in its place. As with the
// steps above, a transformer built with no registries to write refuses every
// login rather than send a client into a world it cannot make sense of.
func DowngradePlayLoginTo1_19_3(registryCodec []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.19.3 play login carries the registries, and this transformer was given none")
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

		// 1.19.4's registries, read so that they are consumed and never
		// written, and 1.19.3's in their place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(registryCodec); err != nil {
			return err
		}

		// The dimension type and the dimension, the seed, the three
		// distances, the four flags and the death location, laid out alike
		// on both sides of this step.
		return copyRest(in, out)
	}
}

// DowngradePlayerPositionTo1_19_3 rewrites the player position packet from
// what 1.19.4 sends into what 1.19.3 reads.
//
// 1.19.4 is where the packet lost its last field, a flag telling the client
// to dismount whatever it rides. 1.19.3 reads it after the teleport id, and
// nothing in a limbo is ridden, so the body is copied and the flag goes on
// the end unset.
func DowngradePlayerPositionTo1_19_3(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyRest(in, out); err != nil {
		return err
	}

	return out.WriteBoolean(dismountVehicle)
}

// dismountVehicle is what the player position packet tells a 1.19.3 client
// about whatever it rides: nothing, since there is nothing to ride here.
const dismountVehicle = false

// The pose serializer's id on either side of the 1.19.4 step: 1.19.4 is where
// the optional block state serializer appeared, registered right after the
// block state one and so in front of the pose, which moved up one.
const (
	poseSerializer1_19_4 = 20
	poseSerializer1_19_3 = 19
)

// DowngradeSetEntityDataTo1_19_3 rewrites the set entity data packet from
// what 1.19.4 sends into what 1.19.3 reads: the same entries with the pose
// under the serializer id 1.19.3 gives it. The byte serializer sits at 0 in
// both, so the flags entry is copied as it is.
func DowngradeSetEntityDataTo1_19_3(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return renamePoseSerializer(in, out, poseSerializer1_19_4, poseSerializer1_19_3)
}
