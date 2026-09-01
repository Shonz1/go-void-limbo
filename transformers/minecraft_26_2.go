package transformers

import "github.com/Shonz1/go-void-limbo/streams"

// uuidSize is what a uuid takes on the wire, as the two longs the protocol
// writes it as.
const uuidSize = 16

// DowngradeLoginSuccessTo26_1 rewrites the login success packet from what 26.2
// sends into what 26.1 reads.
//
// 26.2 appended a session id to it, which 26.1 has no field for: its packet is
// the game profile and nothing else. The sixteen bytes are the whole difference,
// and a 26.1 client that is sent them reads a login success longer than the one
// it expects and drops the connection.
//
// The profile in front of the session id has to be walked rather than skipped,
// since the properties are an array of strings and there is no fixed width at
// which the session id starts.
func DowngradeLoginSuccessTo26_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The profile's uuid.
	if err := copyBytes(in, out, uuidSize); err != nil {
		return err
	}

	// The username.
	if err := copyString(in, out); err != nil {
		return err
	}

	properties, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	for i := int32(0); i < properties; i++ {
		// The property's name and value.
		for j := 0; j < 2; j++ {
			if err := copyString(in, out); err != nil {
				return err
			}
		}

		signed, err := copyBoolean(in, out)
		if err != nil {
			return err
		}

		if signed {
			if err := copyString(in, out); err != nil {
				return err
			}
		}
	}

	// The session id, read so that it is consumed and never written.
	_, err = in.ReadBytes(uuidSize)

	return err
}

// DowngradeAddEntityTo26_1 rewrites the add entity packet from what 26.2 sends
// into what 26.1 reads.
//
// The packet's shape is identical in the two; what moved is the entity type
// registry behind one of its fields, where 26.2's additions renumbered the
// player. The only entity this server ever spawns is a player, so the one id
// is the whole mapping.
func DowngradeAddEntityTo26_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType26_2, playerEntityType26_1)
}

// DowngradePlayLoginTo26_1 rewrites the play phase login packet from what 26.2
// sends into what 26.1 reads.
//
// 26.2 added the online mode flag, which sits between the spawn info block and
// enforces secure chat. The two around it are booleans of one byte each, so a
// 26.1 client reads the flag as enforces secure chat and then runs off the end
// of the body looking for a field that is no longer there. Dropping it is the
// whole difference: writing it false instead would move every byte after it by
// one just the same.
//
// Everything before it is copied across a field at a time rather than as a
// block, because the fields before it are not a fixed width: the dimension
// names are an array, and the spawn info holds a string and an optional death
// location.
func DowngradePlayLoginTo26_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id is a plain int here rather than a var int, unlike everywhere
	// else the protocol names an entity.
	if err := copyBytes(in, out, 4); err != nil {
		return err
	}

	// Is hardcore.
	if _, err := copyBoolean(in, out); err != nil {
		return err
	}

	dimensions, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	for i := int32(0); i < dimensions; i++ {
		if err := copyString(in, out); err != nil {
			return err
		}
	}

	// Max players, view distance, simulation distance.
	for i := 0; i < 3; i++ {
		if _, err := copyVarInt(in, out); err != nil {
			return err
		}
	}

	// Reduced debug info, show death screen, do limited crafting.
	for i := 0; i < 3; i++ {
		if _, err := copyBoolean(in, out); err != nil {
			return err
		}
	}

	if err := copySpawnInfo(in, out); err != nil {
		return err
	}

	// The online mode flag, read so that it is consumed and never written.
	if _, err := in.ReadBoolean(); err != nil {
		return err
	}

	// Enforces secure chat, which is all that is left.
	return copyRest(in, out)
}

// copySpawnInfo moves the block describing the world the client is placed in
// across. 26.1 and 26.2 lay it out the same way.
func copySpawnInfo(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// Dimension type id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	// Dimension name.
	if err := copyString(in, out); err != nil {
		return err
	}

	// Hashed seed.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	// Game mode and previous game mode, a byte each.
	if err := copyBytes(in, out, 2); err != nil {
		return err
	}

	// Is debug, is flat.
	for i := 0; i < 2; i++ {
		if _, err := copyBoolean(in, out); err != nil {
			return err
		}
	}

	hasDeathLocation, err := copyBoolean(in, out)
	if err != nil {
		return err
	}

	if hasDeathLocation {
		// The dimension it is in, and the block itself packed into one long.
		if err := copyString(in, out); err != nil {
			return err
		}

		if err := copyBytes(in, out, 8); err != nil {
			return err
		}
	}

	// Portal cooldown.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	// Sea level.
	_, err = copyVarInt(in, out)

	return err
}
