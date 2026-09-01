package transformers

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The two ends of the velocity rework this step crosses: 1.21.9 introduced the
// quantized velocity vector in the middle of the add entity packet, and 1.21.7
// still reads the three shorts it replaced at the packet's end. The zero this
// server always sends is one byte on the new side -- a leading byte of zero
// means the zero vector, with nothing behind it -- and three empty shorts on
// the old one.
const lpVec3Zero = 0x00

// DowngradeAddEntityTo1_21_7 rewrites the add entity packet from what 1.21.9
// sends into what 1.21.7 reads.
//
// Two things moved at this step. The entity type registry renumbered the
// player, as at the steps above. And the velocity changed shape and place: a
// quantized vector between the position and the rotations in 1.21.9, three
// shorts after the data field in 1.21.7. The velocity this server sends is
// always zero -- a spawned player is standing still -- and the rewrite only
// carries that one value, refusing anything else rather than quietly
// misplacing it.
func DowngradeAddEntityTo1_21_7(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	// The entity's uuid.
	if err := copyBytes(in, out, uuidSize); err != nil {
		return err
	}

	entityType, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if entityType != playerEntityType1_21_9 {
		return fmt.Errorf("add entity carries entity type %d, expected the player's %d", entityType, playerEntityType1_21_9)
	}

	if err := out.WriteVarInt(playerEntityType1_21_7); err != nil {
		return err
	}

	// The position, three doubles.
	if err := copyBytes(in, out, 24); err != nil {
		return err
	}

	// The velocity, consumed here and written as the shorts at the end.
	velocity, err := in.ReadByte()
	if err != nil {
		return err
	}

	if velocity != lpVec3Zero {
		return fmt.Errorf("add entity carries velocity byte %#x, expected the zero vector", velocity)
	}

	// The three rotation bytes and the data var int.
	if err := copyBytes(in, out, 3); err != nil {
		return err
	}

	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	for range 3 {
		if err := out.WriteShort(0); err != nil {
			return err
		}
	}

	return nil
}

// The serializer the pose field names on each side of this step: 1.21.9
// retired the compound tag serializer, and every serializer registered after
// it moved down one.
const (
	poseSerializer1_21_9 = 20
	poseSerializer1_21_7 = 21
)

// The two serializers a set entity data packet from this server ever names,
// at the version its entries are written at. Everything about an entry after
// its serializer is shaped by the serializer, so the rewrite can only carry
// entries it knows the shape of.
const (
	byteSerializer1_21_9 = 0

	// entityDataTerminator ends the entry list, sitting where no real index
	// can.
	entityDataTerminator = 0xff
)

// DowngradeSetEntityDataTo1_21_7 rewrites the set entity data packet from what
// 1.21.9 sends into what 1.21.7 reads.
//
// Metadata entries name their serializer by its spot in the client's
// serializer registry, and 1.21.9 removed a serializer that sat before the
// pose, moving it from 21 to 20. The entries themselves are laid out the same
// way in the two, so renaming the serializer is the whole rewrite -- but the
// walk has to understand each entry to get past its value, so only the two
// serializers this server sends are accepted.
func DowngradeSetEntityDataTo1_21_7(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	for {
		index, err := in.ReadByte()
		if err != nil {
			return err
		}

		if err := out.WriteByte(index); err != nil {
			return err
		}

		if index == entityDataTerminator {
			return nil
		}

		serializer, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		switch serializer {
		case byteSerializer1_21_9:
			if err := out.WriteVarInt(serializer); err != nil {
				return err
			}

			if err := copyBytes(in, out, 1); err != nil {
				return err
			}
		case poseSerializer1_21_9:
			if err := out.WriteVarInt(poseSerializer1_21_7); err != nil {
				return err
			}

			// The pose ordinal, identical in the two versions.
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		default:
			return fmt.Errorf("set entity data carries serializer %d, which this rewrite does not know the shape of", serializer)
		}
	}
}
