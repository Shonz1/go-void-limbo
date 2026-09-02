package transformers

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.20.5 step is where the login and the registry synchronization were
// reworked. Each rewrite below is one of those changes, checked against the
// 1.20.4 client's own classes through Mojang's mappings; the registries
// themselves are no packet body rewrite at all but a different packet, which
// package gamedata encodes for 1.20.3 in the one-compound shape it reads.
// Everything else this server speaks -- the chunk packet, the tags, the
// movement on both sides, the player list -- is wire-identical in 765 and 766.

// DowngradeEncryptionRequestTo1_20_3 rewrites the encryption request from
// what 1.20.5 sends into what 1.20.3 reads.
//
// 1.20.5 appended the should authenticate flag, for a server that wants the
// cipher without the account behind it. 1.20.3 has no such server: it always
// tells Mojang about the join, which is what this server asks for anyway, so
// the flag comes off the end and nothing is lost.
func DowngradeEncryptionRequestTo1_20_3(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The server id.
	if err := copyString(in, out); err != nil {
		return err
	}

	// The public key and the verify token, each a counted byte array.
	for range 2 {
		if err := copyByteArray(in, out); err != nil {
			return err
		}
	}

	shouldAuthenticate, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	if !shouldAuthenticate {
		return fmt.Errorf("encryption request asks the client not to authenticate, which 1.20.3 has no way to be told")
	}

	return nil
}

// DowngradeLoginSuccessTo1_20_3 rewrites the login success packet from what
// 1.20.5 sends into what 1.20.3 reads.
//
// 1.20.5 is where the strict error handling flag appeared at the end of the
// packet, the same flag 1.21.2 removed again, so the 1.21.2 step puts it on
// and this one takes it off: 1.20.3's packet is the profile and nothing after
// it.
func DowngradeLoginSuccessTo1_20_3(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
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

	// Strict error handling, read so that it is consumed and never written.
	_, err := in.ReadBoolean()

	return err
}

// The dimension type the play login names, as 1.20.3 reads it. 1.20.5 turned
// the field into an index into the dimension type registry; before that it
// was the entry's own name. Package gamedata sends one dimension type, the
// overworld, first and alone, so the one index this server ever sends is the
// one name below.
const (
	overworldDimensionTypeId     = 0
	overworldDimensionType1_20_3 = "minecraft:overworld"
)

// DowngradePlayLoginTo1_20_3 rewrites the play phase login packet from what
// 1.20.5 sends into what 1.20.3 reads.
//
// Two fields differ. The spawn info opens with the dimension type, which
// 1.20.5 sends as a registry index and 1.20.3 reads as a name, so the index
// is looked up and the name written in its place. And 1.20.5 is where the
// enforces secure chat flag moved into this packet from the server data one;
// 1.20.3 reads nothing after the spawn info, so the flag comes off the end.
func DowngradePlayLoginTo1_20_3(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyPlayLoginHead(in, out); err != nil {
		return err
	}

	dimensionType, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if dimensionType != overworldDimensionTypeId {
		return fmt.Errorf("play login names dimension type %d, and %d is the only one 1.20.3 has a name for", dimensionType, overworldDimensionTypeId)
	}

	if err := out.WriteString(overworldDimensionType1_20_3); err != nil {
		return err
	}

	if err := copySpawnInfoAfterDimensionType(in, out); err != nil {
		return err
	}

	// Enforces secure chat, read so that it is consumed and never written.
	_, err = in.ReadBoolean()

	return err
}

// DowngradeAddEntityTo1_20_3 rewrites the add entity packet from what 1.20.5
// sends into what 1.20.3 reads.
//
// The packet's shape is identical in the two, and what moved is the entity
// type registry behind one of its fields: 1.20.5 added the armadillo, the
// bogged, the breeze's wind charge and the ominous item spawner, all of which
// sort before the player.
func DowngradeAddEntityTo1_20_3(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType1_20_5, playerEntityType1_20_3)
}

// The serializer the pose field names on each side of this step: 1.20.5
// added the particles serializer right after the particle one, and every
// serializer registered after it moved up one.
const (
	poseSerializer1_20_5 = 21
	poseSerializer1_20_3 = 20
)

// DowngradeSetEntityDataTo1_20_3 rewrites the set entity data packet from
// what 1.20.5 sends into what 1.20.3 reads: the pose serializer goes back to
// the number it had before 1.20.5's insertion, which is also the number it
// has again from 1.21.9 on. See DowngradeSetEntityDataTo1_21_7 for the walk.
func DowngradeSetEntityDataTo1_20_3(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return renamePoseSerializer(in, out, poseSerializer1_20_5, poseSerializer1_20_3)
}
