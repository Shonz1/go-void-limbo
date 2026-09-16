package transformers

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// DowngradePlayLoginTo26_2 rewrites the play phase login packet from what 26.3
// sends into what 26.2 reads.
//
// The one thing 26.3 changed is how the spawn info spells the two game modes.
// The mode itself went from a byte to a var int, which for the four modes there
// are is the same byte either way. The previous mode went from a byte that is
// -1 for none to an optional var int that is zero for none and one more than
// the mode otherwise, so that one is read and written back the older way.
// Everything around the two is copied across a field at a time, because the
// fields before them are not a fixed width: the dimension names are an array.
func DowngradePlayLoginTo26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyPlayLoginHead(in, out); err != nil {
		return err
	}

	// Dimension type id, dimension name, hashed seed.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	if err := copyString(in, out); err != nil {
		return err
	}

	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	gameMode, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if err := out.WriteByte(byte(gameMode)); err != nil {
		return err
	}

	previous, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	// Zero is none, which 26.2 spells as -1; anything else is the mode plus
	// one.
	if err := out.WriteByte(byte(previous - 1)); err != nil {
		return err
	}

	// Is debug, is flat, the death location, the portal cooldown, the sea
	// level and the two flags at the end, none of which moved.
	return copyRest(in, out)
}

// DowngradeAddEntityTo26_2 rewrites the add entity packet from what 26.3 sends
// into what 26.2 reads.
//
// The packet's shape is identical in the two; what moved is the entity type
// registry behind one of its fields, where 26.3's additions renumbered the
// player. The only entity this server ever spawns is a player, so the one id
// is the whole mapping.
func DowngradeAddEntityTo26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType26_3, playerEntityType26_2)
}

// DowngradeEntityPositionSyncTo26_2 rewrites the entity position sync packet
// from what 26.3 sends into what 26.2 reads.
//
// 26.3 describes the position as a path: a var int saying which kind, and for
// the straight line this server sends the point it ends at. 26.2 reads the
// point, and after it the delta movement the entity is left with, three
// doubles 26.3 has no field for. The path type comes off and the delta goes in
// as zero, which is what the packet's own encoder wrote there before 26.3: the
// next position comes as another one of these rather than being extrapolated
// from a velocity. The other kind of path, which spells out the steps in
// between, has no shape 26.2 can be handed, so it is refused rather than
// guessed at.
func DowngradeEntityPositionSyncTo26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	pathType, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if pathType != positionPathLinear26_3 {
		return fmt.Errorf("entity position sync carries a path of type %d, which 26.2 has no shape for; only a straight line is", pathType)
	}

	// The position, three doubles.
	if err := copyBytes(in, out, 24); err != nil {
		return err
	}

	// The delta movement, three doubles of zero.
	for range 3 {
		if err := out.WriteDouble(0); err != nil {
			return err
		}
	}

	// The yaw, the pitch and on ground, which are all that is left.
	return copyRest(in, out)
}

// positionPathLinear26_3 is the type of a position path that is one straight
// line to its end, the first of the two types 26.3 knows and the one the
// packet's encoder writes.
const positionPathLinear26_3 = 0

// The animate packet's numbers for the two arm swings, which 26.3 retired
// along with the swing packet that asked for them.
const (
	animateSwingMainArm26_2 = 0
	animateSwingOffhand26_2 = 3
)

// DowngradeSwingAnimationTo26_2 rewrites the swing animation packet from what
// 26.3 sends into the animate packet 26.2 reads it as.
//
// 26.3 gave the arm swing a packet of its own: the entity, the hand, the kind
// of swing and how many ticks it takes. 26.2 has no such packet, and plays a
// swing from the animate packet, which is the entity and one byte naming the
// animation among the handful of one-shot ones: 0 for the main arm, 3 for the
// offhand. The kind and the duration have nowhere to go, since 26.2 swings the
// one way it knows, and come off.
func DowngradeSwingAnimationTo26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	hand, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	animation := byte(animateSwingMainArm26_2)
	if hand == 1 {
		animation = animateSwingOffhand26_2
	}

	if err := out.WriteByte(animation); err != nil {
		return err
	}

	// The kind of swing and its duration, read so that they are consumed and
	// never written.
	if _, err := in.ReadVarInt(); err != nil {
		return err
	}

	_, err = in.ReadVarInt()

	return err
}

// UpgradePunchFrom26_2 rewrites the swing packet a 26.2 client sends into the
// punch packet 26.3 reads it as.
//
// A 26.2 swing names the hand that swung, as a var int; a 26.3 punch names
// nothing, since it is only ever the main arm swinging at nothing, and a swing
// at something or with an item reaches a server through the packet that says
// so. The hand comes off, and with it the one thing a 26.2 client could say
// that a 26.3 client cannot: an offhand swing is played on the others as a
// main arm one, the way a 26.3 client's own would be.
func UpgradePunchFrom26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	_, err := in.ReadVarInt()

	return err
}

// UpgradeAcceptTeleportationFrom26_2 rewrites the accept teleportation packet a
// 26.2 client sends into what 26.3 reads.
//
// 26.3 appended the position and the rotation the client put itself at for the
// teleport: three doubles and two floats a 26.2 client never sends. They go in
// as zero, which is no position a client is at but the one value that says so,
// and nothing here reads them: a limbo takes the id as the acknowledgement it
// is and puts the player nowhere on the strength of it.
func UpgradeAcceptTeleportationFrom26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The teleport id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	for range 3 {
		if err := out.WriteDouble(0); err != nil {
			return err
		}
	}

	for range 2 {
		if err := out.WriteFloat(0); err != nil {
			return err
		}
	}

	return nil
}

// DowngradeLevelChunkWithLightTo26_2 rewrites the chunk packet from what 26.3
// sends into what 26.2 reads.
//
// The one thing that moved is inside the light data at the end: the four masks
// are bit sets, and 26.3 sends a bit set as a counted array of bytes where
// 26.2 reads a counted array of longs. Everything in front of them is copied
// across a field at a time, since the heightmaps, the section buffer and the
// block entities are not a fixed width. This server sends no block entities,
// and a body that carries some is refused rather than walked: their shape is
// one this rewrite was never taught.
func DowngradeLevelChunkWithLightTo26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	heightmaps, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	for range heightmaps {
		// The kind, then the counted longs.
		if _, err := copyVarInt(in, out); err != nil {
			return err
		}

		length, err := copyVarInt(in, out)
		if err != nil {
			return err
		}

		if err := copyBytes(in, out, 8*length); err != nil {
			return err
		}
	}

	// The section buffer.
	if err := copyByteArray(in, out); err != nil {
		return err
	}

	blockEntities, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	if blockEntities != 0 {
		return fmt.Errorf("chunk carries %d block entities, which this rewrite cannot walk past", blockEntities)
	}

	return downgradeLightDataTo26_2(in, out)
}

// DowngradeLightUpdateTo26_2 rewrites the light update packet from what 26.3
// sends into what 26.2 reads: the same four masks as in the chunk packet,
// behind the coordinates and the trust edges flag the packet opens with.
func DowngradeLightUpdateTo26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two var ints, and the trust edges flag.
	for range 2 {
		if _, err := copyVarInt(in, out); err != nil {
			return err
		}
	}

	if _, err := copyBoolean(in, out); err != nil {
		return err
	}

	return downgradeLightDataTo26_2(in, out)
}

// downgradeLightDataTo26_2 rewrites the four light masks from bytes into
// longs and copies the light arrays behind them across untouched.
func downgradeLightDataTo26_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	for range 4 {
		bytes, err := in.ReadByteArray(lightMaskBytes)
		if err != nil {
			return err
		}

		longs := bitSetLongs(bytes)

		if err := out.WriteVarInt(int32(len(longs))); err != nil {
			return err
		}

		for _, long := range longs {
			if err := out.WriteLong(long); err != nil {
				return err
			}
		}
	}

	// The sky arrays and the block arrays, which are all that is left.
	return copyRest(in, out)
}

// lightMaskBytes is the most bytes a light mask can hold: a bit per section
// of light, of which a world this server speaks has twenty-six, though a mask
// is refused only past what a wildly taller world would need.
const lightMaskBytes = 64

// bitSetLongs lays a bit set sent as bytes out as the longs 26.2 reads it as:
// eight bytes to a long, lowest first, with the trailing longs that hold no
// bit left off, which is what the client's own bit set writes.
func bitSetLongs(bytes []byte) []int64 {
	longs := make([]int64, (len(bytes)+7)/8)
	for i, b := range bytes {
		longs[i/8] |= int64(b) << (8 * (i % 8))
	}

	for len(longs) > 0 && longs[len(longs)-1] == 0 {
		longs = longs[:len(longs)-1]
	}

	return longs
}
