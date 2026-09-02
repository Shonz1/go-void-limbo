package transformers

import (
	"errors"
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.20.2 step is where the configuration phase appeared, and with it the
// nameless network NBT and the registry packets. A 1.20 client has none of
// those: it is in play the moment its login succeeds, reads the registries out
// of the play login itself as one compound named as a root, reads the tags as
// a play packet right after it, and reads every other NBT on the wire named
// too. It also has no add entity packet for a player, which 1.20.2 is where
// it was folded in, and its hello carries its uuid as an optional. Each
// rewrite below is one of those differences, read off the 1.20 client's own
// classes through Mojang's mappings the way the 1.20.3 step was. Everything
// else this server speaks is wire-identical in 763 and 764: the login on both
// sides, the player list, the movement on both sides, the entity metadata --
// whose serializers are registered in the same order, so the pose still sits
// at 20 -- the tags and the sections inside the chunk packet.

// DowngradePlayLoginTo1_20 rewrites the play phase login packet from what
// 1.20.2 sends into what 1.20 reads, given the registries that version reads
// out of it.
//
// 1.20.2 rebuilt the packet around the spawn info block it shares with the
// respawn packet, and moved the registries out of it into the configuration
// phase. 1.20 lays the same fields out in its own order -- the game modes
// right after the hardcore flag, the dimension type and name, the seed and
// the counts after the registries -- and has no limited crafting flag, which
// 1.20.2 added. So the packet is read whole and written over in the older
// order, with the registry compound in the middle of it. That compound is the
// one thing here the body does not carry: package gamedata encodes it once,
// and a transformer built without it refuses every login rather than send a
// client a world it cannot make sense of.
func DowngradePlayLoginTo1_20(registryCodec []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.20 play login carries the registries, and this transformer was given none")
		}

		// The entity id is a plain int here rather than a var int, unlike
		// everywhere else the protocol names an entity.
		entityId, err := in.ReadInt()
		if err != nil {
			return err
		}

		hardcore, err := in.ReadBoolean()
		if err != nil {
			return err
		}

		dimensionCount, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		dimensions := make([]string, dimensionCount)
		for i := range dimensions {
			if dimensions[i], err = in.ReadString(); err != nil {
				return err
			}
		}

		// Max players, view distance, simulation distance.
		var counts [3]int32
		for i := range counts {
			if counts[i], err = in.ReadVarInt(); err != nil {
				return err
			}
		}

		// Reduced debug info, show death screen, do limited crafting. The
		// last is 1.20.2's and 1.20 has no field for it.
		var flags [3]bool
		for i := range flags {
			if flags[i], err = in.ReadBoolean(); err != nil {
				return err
			}
		}

		dimensionType, err := in.ReadString()
		if err != nil {
			return err
		}

		dimension, err := in.ReadString()
		if err != nil {
			return err
		}

		seed, err := in.ReadLong()
		if err != nil {
			return err
		}

		// Game mode and previous game mode, a byte each.
		gameMode, err := in.ReadByte()
		if err != nil {
			return err
		}

		previousGameMode, err := in.ReadByte()
		if err != nil {
			return err
		}

		// Is debug, is flat, the optional death location and the portal
		// cooldown close both shapes alike, so they are copied over once
		// everything in front of them is in place.
		if err := out.WriteInt(entityId); err != nil {
			return err
		}

		if err := out.WriteBoolean(hardcore); err != nil {
			return err
		}

		if err := out.WriteByte(gameMode); err != nil {
			return err
		}

		if err := out.WriteByte(previousGameMode); err != nil {
			return err
		}

		if err := out.WriteVarInt(dimensionCount); err != nil {
			return err
		}

		for _, name := range dimensions {
			if err := out.WriteString(name); err != nil {
				return err
			}
		}

		if err := out.WriteBytes(registryCodec); err != nil {
			return err
		}

		if err := out.WriteString(dimensionType); err != nil {
			return err
		}

		if err := out.WriteString(dimension); err != nil {
			return err
		}

		if err := out.WriteLong(seed); err != nil {
			return err
		}

		for _, count := range counts {
			if err := out.WriteVarInt(count); err != nil {
				return err
			}
		}

		for _, flag := range flags[:2] {
			if err := out.WriteBoolean(flag); err != nil {
				return err
			}
		}

		return copyRest(in, out)
	}
}

// DowngradeAddEntityTo1_20 rewrites the add entity packet from what 1.20.2
// sends into what 1.20 reads, which is a different packet.
//
// 1.20.2 is where a player came to be spawned by the add entity packet like
// everything else; before it, the client only spawned one from the add player
// packet, which the id table names for 1.20 and whose body this is: the
// entity id, the uuid, the position, and the yaw and pitch in that order,
// which is the reverse of the add entity packet's. The entity type has to be
// the player's, since the add player packet spawns nothing else, and the head
// yaw, the data and the velocity have no field to go to: the head follows in
// a rotate head packet of its own, the data means nothing for a player, and
// the velocity this server sends is always zero, so anything else is refused
// rather than quietly dropped.
func DowngradeAddEntityTo1_20(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
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

	if entityType != playerEntityType1_20_2 {
		return fmt.Errorf("add entity carries entity type %d, and 1.20's add player packet spawns only the player's %d", entityType, playerEntityType1_20_2)
	}

	// The position, three doubles.
	if err := copyBytes(in, out, 24); err != nil {
		return err
	}

	// The pitch, the yaw and the head yaw, an angle byte each, of which the
	// first two go out in the other order.
	pitch, err := in.ReadByte()
	if err != nil {
		return err
	}

	yaw, err := in.ReadByte()
	if err != nil {
		return err
	}

	if _, err := in.ReadByte(); err != nil {
		return err
	}

	if err := out.WriteByte(yaw); err != nil {
		return err
	}

	if err := out.WriteByte(pitch); err != nil {
		return err
	}

	// The data, read so that it is consumed and never written.
	if _, err := in.ReadVarInt(); err != nil {
		return err
	}

	// The velocity, three shorts.
	for _, axis := range []string{"x", "y", "z"} {
		velocity, err := in.ReadShort()
		if err != nil {
			return err
		}

		if velocity != 0 {
			return fmt.Errorf("add entity carries a velocity %s of %d, which 1.20's add player packet has no field for", axis, velocity)
		}
	}

	return nil
}

// DowngradeLevelChunkWithLightTo1_20 rewrites the chunk packet from what
// 1.20.2 sends into what 1.20 reads.
//
// 1.20.2 is where NBT on the wire lost its root name. Before it, every
// compound went out the way it does in a file, with a name after the type
// byte, and the one the chunk packet carries -- the heightmaps, right after
// the chunk coordinates -- is named with the empty string, as a vanilla
// server names it. The rewrite reads the nameless compound the 1.21.5 step
// produced, writes it named, and copies everything after it untouched.
//
// The section data behind it is not this transformer's concern: sections are
// built per version by package world before the packet exists, and no block
// entity goes out for its NBT to need naming.
func DowngradeLevelChunkWithLightTo1_20(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk coordinates, two ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	heightmaps, err := nbt.Read(in)
	if err != nil {
		return err
	}

	if err := nbt.WriteNamed(out, "", heightmaps); err != nil {
		return err
	}

	return copyRest(in, out)
}

// UpgradeLoginStartFrom1_20 rewrites the login start packet from what 1.20
// sends into what 1.20.2 reads.
//
// 1.20.2 made the uuid a field of its own; 1.20 sends it as an optional, a
// flag and then the uuid when the flag is set. A 1.20 client always sets it,
// but the shape allows for one that does not, and what such a client is
// taken to have sent is the nil uuid: what the client says about itself here
// is worth nothing anyway, since the login settles on Mojang's word or on the
// name alone.
func UpgradeLoginStartFrom1_20(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The username.
	if err := copyString(in, out); err != nil {
		return err
	}

	hasUuid, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	if hasUuid {
		return copyBytes(in, out, uuidSize)
	}

	return out.WriteBytes(make([]byte, uuidSize))
}
