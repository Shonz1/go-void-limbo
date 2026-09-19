package transformers

import "github.com/Shonz1/go-void-limbo/streams"

// DowngradeAddEntityTo1_21_5 rewrites the add entity packet from what 1.21.6
// sends into what 1.21.5 reads.
//
// The packet's shape is identical in the two -- 1.21.5 already reads the
// velocity as the three shorts at the end that 1.21.7 does -- and what moved
// is the entity type registry behind one of its fields: 1.21.6 added the happy
// ghast, which sorts before the player and so pushed its number up by one.
func DowngradeAddEntityTo1_21_5(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType1_21_6, playerEntityType1_21_5)
}

// The player command actions as the versions below 1.21.6 number them, the
// two shift actions in front, and where the shift actions go on the way up:
// past the end of 1.21.6's list, which is where the packet's decoder looks
// for them.
const (
	playerCommandPressShiftKey1_21_5   = 0
	playerCommandReleaseShiftKey1_21_5 = 1
	playerCommandShiftActions1_21_5    = 2

	playerCommandPressShiftKey   = 7
	playerCommandReleaseShiftKey = 8
)

// UpgradePlayerCommandFrom1_21_5 rewrites the player command packet from what
// 1.21.5 sends into what 1.21.6 reads.
//
// The shape is the same -- an entity id, an action and a number for the action
// to use -- and what changed is the list of actions: 1.21.6 took the press and
// the release of the shift key off its front, the player input packet having
// a bit for the key since 1.21.2, and every action after them moved down by
// two. Those move down here. The two shift actions have nowhere in 1.21.6 to
// go, and they cannot be dropped: below 1.21.2 the input packet is only sent
// from a vehicle, so they are all a player on foot says about sneaking. They
// go past the end of the list.
func UpgradePlayerCommandFrom1_21_5(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	entityId, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	action, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	data, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	switch action {
	case playerCommandPressShiftKey1_21_5:
		action = playerCommandPressShiftKey
	case playerCommandReleaseShiftKey1_21_5:
		action = playerCommandReleaseShiftKey
	default:
		action -= playerCommandShiftActions1_21_5
	}

	if err := out.WriteVarInt(entityId); err != nil {
		return err
	}

	if err := out.WriteVarInt(action); err != nil {
		return err
	}

	return out.WriteVarInt(data)
}
