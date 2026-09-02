package transformers

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.21.2 step is the widest this server crosses: 1.21.2 reworked how a
// player and an entity are moved, and touched the login on both sides of the
// configuration phase. Each rewrite below is one of those changes, checked
// against the client's own classes through Mojang's mappings; everything else
// this server speaks -- the chunk packet, registry synchronization, the tags,
// the entity metadata -- is wire-identical in 767 and 768.

// DowngradeAddEntityTo1_21 rewrites the add entity packet from what 1.21.2
// sends into what 1.21 reads.
//
// The packet's shape is identical in the two, and what moved is the entity
// type registry behind one of its fields, by more than at any other step:
// 1.21.2 gave every wood its own boat and chest boat where 1.21 had one of
// each, and added the creaking, all of which sort before the player.
func DowngradeAddEntityTo1_21(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType1_21_2, playerEntityType1_21)
}

// strictErrorHandling is what the login success packet tells a 1.21 client
// to do about a packet it cannot handle: disconnect, as a vanilla server tells
// it to. The flag exists for servers that would rather the client logged and
// carried on, and this server would rather hear about a packet it got wrong.
const strictErrorHandling = true

// DowngradeLoginSuccessTo1_21 rewrites the login success packet from what
// 1.21.2 sends into what 1.21 reads.
//
// 1.21 ends the packet with a boolean 1.21.2 removed, the strict error
// handling flag, so the profile is walked across and the flag written after
// it. The profile has to be walked rather than copied as a block, since its
// properties are an array of strings with no fixed width.
func DowngradeLoginSuccessTo1_21(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The profile's uuid.
	if err := copyBytes(in, out, uuidSize); err != nil {
		return err
	}

	// The username.
	if err := copyString(in, out); err != nil {
		return err
	}

	if err := copyProfileProperties(in, out); err != nil {
		return err
	}

	return out.WriteBoolean(strictErrorHandling)
}

// DowngradePlayLoginTo1_21 rewrites the play phase login packet from what
// 1.21.2 sends into what 1.21 reads.
//
// 1.21.2 added the sea level to the spawn info, at its end, right in front of
// the enforces secure chat flag that closes the packet. 1.21 has no field for
// it and reads on into the flag, so the var int comes out and everything
// around it is walked across as it is.
func DowngradePlayLoginTo1_21(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyPlayLoginHead(in, out); err != nil {
		return err
	}

	if err := copySpawnInfo(in, out); err != nil {
		return err
	}

	// The sea level, read so that it is consumed and never written.
	if _, err := in.ReadVarInt(); err != nil {
		return err
	}

	// Enforces secure chat, which is all that is left.
	return copyRest(in, out)
}

// DowngradePlayerInfoUpdateTo1_21 rewrites the player info update packet from
// what 1.21.2 sends into what 1.21 reads.
//
// 1.21.2 gave the packet a seventh action, the list order, appended after the
// display name: one more bit in the mask and, for every entry, one var int at
// the end. 1.21 has six actions, so the bit comes off the mask and the var
// int off each entry. This server never sends the action, so the rewrite
// mostly finds nothing to drop, and copies the packet as it is.
func DowngradePlayerInfoUpdateTo1_21(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return dropPlayerInfoAction(in, out, play.PlayerInfoUpdateListOrder)
}

// The relative flags a 1.21 player position packet has bits for: x, y, z,
// yaw and pitch, the low five of 1.21.2's nine. The four 1.21.2 added are
// about the delta movement, which 1.21's packet does not carry.
const relativeFlags1_21 = 0x1F

// DowngradePlayerPositionTo1_21 rewrites the player position packet from what
// 1.21.2 sends into what 1.21 reads.
//
// 1.21.2 rebuilt the packet around a delta movement: the teleport id moved
// from the end to the front, three doubles of velocity landed after the
// position, and the relative flags grew from a byte of five bits to an int of
// nine, the four new ones about the velocity. 1.21's packet is the position,
// the rotation, the byte of flags and then the id, so the fields are read in
// one order and written in the other.
//
// The velocity has no field on the 1.21 side, and the rewrite carries only
// the zero this server sends. That loses nothing: a 1.21 client zeroes its
// own velocity on every axis the packet moves it absolutely on, which is what
// a zero delta sent absolutely means. A velocity that is not zero, or a flag
// among the four 1.21 has no bit for, is refused rather than dropped.
func DowngradePlayerPositionTo1_21(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	teleportId, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	// The position, three doubles.
	position, err := in.ReadBytes(24)
	if err != nil {
		return err
	}

	for _, axis := range []string{"x", "y", "z"} {
		delta, err := in.ReadDouble()
		if err != nil {
			return err
		}

		if delta != 0 {
			return fmt.Errorf("player position carries a delta %s of %g, which 1.21 has no field for", axis, delta)
		}
	}

	// The rotation, two floats.
	rotation, err := in.ReadBytes(8)
	if err != nil {
		return err
	}

	relatives, err := in.ReadInt()
	if err != nil {
		return err
	}

	if relatives&^relativeFlags1_21 != 0 {
		return fmt.Errorf("player position carries relative flags %#x, of which 1.21 has bits for %#x only", relatives, relativeFlags1_21)
	}

	if err := out.WriteBytes(position); err != nil {
		return err
	}

	if err := out.WriteBytes(rotation); err != nil {
		return err
	}

	if err := out.WriteByte(byte(relatives)); err != nil {
		return err
	}

	return out.WriteVarInt(teleportId)
}

// DowngradeEntityPositionSyncTo1_21 rewrites the entity position sync packet
// from what 1.21.2 sends into what 1.21 reads, which is a different packet:
// 1.21 has no position sync, and the teleport entity packet is how an entity
// is put where the server says. The id table names the teleport for 1.21, and
// this is the body that goes under it.
//
// The two agree on the entity id in front and the on ground flag at the end.
// Between them 1.21.2 sends the position, a delta movement and the rotation as
// floats, and 1.21 reads the position and the rotation as angle bytes, the
// 256 steps around the circle every other entity packet uses. The delta has
// no field on the 1.21 side, and the rewrite carries only the zero this
// server sends -- the next position arrives as another one of these rather
// than being extrapolated -- refusing anything else rather than dropping it.
func DowngradeEntityPositionSyncTo1_21(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	// The position, three doubles.
	if err := copyBytes(in, out, 24); err != nil {
		return err
	}

	for _, axis := range []string{"x", "y", "z"} {
		delta, err := in.ReadDouble()
		if err != nil {
			return err
		}

		if delta != 0 {
			return fmt.Errorf("entity position sync carries a delta %s of %g, which 1.21's teleport has no field for", axis, delta)
		}
	}

	// The yaw and the pitch, each a float here and an angle byte there.
	for range 2 {
		degrees, err := in.ReadFloat()
		if err != nil {
			return err
		}

		if err := out.WriteByte(play.Angle(degrees)); err != nil {
			return err
		}
	}

	// On ground, which is all that is left.
	return copyRest(in, out)
}

// The two bits of the flag byte a 1.21 player input packet ends in.
const (
	playerInputFlagJump1_21  = 0x01
	playerInputFlagSneak1_21 = 0x02
)

// The bits of the one byte a 1.21.2 player input packet is, as the packet's
// own decoder reads them.
const (
	playerInputForward  = 0x01
	playerInputBackward = 0x02
	playerInputLeft     = 0x04
	playerInputRight    = 0x08
	playerInputJump     = 0x10
	playerInputSneak    = 0x20
)

// UpgradePlayerInputFrom1_21 rewrites the player input packet from what 1.21
// sends into what 1.21.2 reads.
//
// 1.21 reports the movement keys as two floats and a flag byte: a sideways
// impulse, positive to the left, a forward impulse, positive forward, and a
// byte with a bit for jumping and a bit for sneaking. 1.21.2 reports them as
// one byte with a bit per key. An impulse becomes the key on the side its
// sign says, zero becomes neither, and the two flag bits move to their new
// spots. Sprinting has no bit on the 1.21 side, so it comes out unset.
//
// A 1.21 client sends this packet only while riding something, which nothing
// in a limbo is, so what the rewrite mostly changes is that a stray one
// decodes as itself rather than as the first byte of a float.
func UpgradePlayerInputFrom1_21(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	sideways, err := in.ReadFloat()
	if err != nil {
		return err
	}

	forward, err := in.ReadFloat()
	if err != nil {
		return err
	}

	flags, err := in.ReadByte()
	if err != nil {
		return err
	}

	var input byte

	if forward > 0 {
		input |= playerInputForward
	} else if forward < 0 {
		input |= playerInputBackward
	}

	if sideways > 0 {
		input |= playerInputLeft
	} else if sideways < 0 {
		input |= playerInputRight
	}

	if flags&playerInputFlagJump1_21 != 0 {
		input |= playerInputJump
	}

	if flags&playerInputFlagSneak1_21 != 0 {
		input |= playerInputSneak
	}

	return out.WriteByte(input)
}
