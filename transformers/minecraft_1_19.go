package transformers

import (
	"errors"
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.19 step is where the profile key arrived, and with it most of what
// the two versions lay out differently, read off the 1.18.2 client's own
// classes through Mojang's mappings the way the 1.19.1 step was. 1.19 and
// 1.18.2 lay all but five of the packets this server speaks out alike --
// the hello, onto which 1.19 put the optional profile key; the encryption
// response, whose challenge 1.19 lets a client sign; the login success,
// onto which 1.19 put the profile's properties; the player list, whose
// added player 1.19 gave the key; and the play login, which 1.18.2 reads
// with registries of its own, with the dimension type spelled out in it
// rather than named, and with no death location on the end -- and each
// rewrite below is one of those five. Everything else is wire-identical in
// 758 and 759: the encryption request, the login disconnect and the
// compression, the add player packet, the player position with its
// dismount flag, the movement on both sides, the entity metadata with the
// pose at 18, the tags, the spawn position and the chunk packet with its
// named heightmaps and its trust edges flag. 1.19 did renumber the play
// phase, in both directions, which the id tables say; this file is only
// about the bodies.

// DowngradePlayLoginTo1_18_2 rewrites the play phase login packet from what
// 1.19 sends into what 1.18.2 reads, given the registries that version
// reads out of it and the dimension type it spells out in it.
//
// The two versions lay the packet out alike from the front up to the
// registries in the middle, which are 1.19's, put there by the 1.19.1 step,
// and which 1.18.2 has its own of: two registries to 1.19's three, since
// 1.19 is where the chat types joined them, encoded once by package
// gamedata. Behind the registries 1.19 names the dimension type the player
// is put into, one of the entries the registries hold, and 1.18.2 reads the
// entry itself there, spelled out in full a second time: so the name is
// read and the entry written in its place, the one entry this server ever
// names, and a login naming another is refused rather than sent with the
// wrong world. The rest is laid out alike up to the end, where 1.19 puts
// the optional death location 1.18.2 has no field for, which comes off. As
// with the steps above, a transformer built with no registries or no
// dimension type to write refuses every login rather than send a client
// into a world it cannot make sense of.
func DowngradePlayLoginTo1_18_2(registryCodec []byte, dimensionType []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.18.2 play login carries the registries, and this transformer was given none")
		}

		if len(dimensionType) == 0 {
			return errors.New("a 1.18.2 play login spells out the dimension type, and this transformer was given none")
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

		// 1.19's registries, read so that they are consumed and never
		// written, and 1.18.2's in their place.
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return err
		}

		if err := out.WriteBytes(registryCodec); err != nil {
			return err
		}

		// The dimension type's name, and the entry it names in its place.
		name, err := in.ReadString()
		if err != nil {
			return err
		}

		if name != overworldDimensionType1_20_3 {
			return fmt.Errorf("play login puts the player into the %s dimension type, which this rewrite has no definition of", name)
		}

		if err := out.WriteBytes(dimensionType); err != nil {
			return err
		}

		// The dimension.
		if err := copyString(in, out); err != nil {
			return err
		}

		// The seed, a long.
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
		if err := copyBytes(in, out, 4); err != nil {
			return err
		}

		// The death location, read so that it is consumed and never written:
		// the dimension it was in and the packed position.
		hasDeathLocation, err := in.ReadBoolean()
		if err != nil {
			return err
		}

		if !hasDeathLocation {
			return nil
		}

		if _, err := in.ReadString(); err != nil {
			return err
		}

		_, err = in.ReadLong()

		return err
	}
}

// DowngradePlayerInfoUpdateTo1_18_2 rewrites the player info packet from what
// 1.19 sends into what 1.18.2 reads.
//
// The packet reaches this step as the 1.19.3 step wrote it: one action, and
// every entry carrying that action's fields. 1.19 is where an added player
// came to carry a profile key, an optional on the end of the entry behind
// the display name, and 1.18.2 has no field for it: so an add goes across
// with the key read off the end of every entry, absent as this server
// always sends it, and refused present, since a key is the client's own to
// announce and this server keeps none. Every other action is laid out
// alike in the two and copied as it is.
func DowngradePlayerInfoUpdateTo1_18_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	action, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	if action != playerInfoAction1_19_1AddPlayer {
		return copyRest(in, out)
	}

	entries, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	for range entries {
		// The entry's uuid and the username.
		if err := copyBytes(in, out, uuidSize); err != nil {
			return err
		}

		if err := copyString(in, out); err != nil {
			return err
		}

		if err := copyProfileProperties(in, out); err != nil {
			return err
		}

		// The game mode and the latency.
		for range 2 {
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		}

		// The display name, absent as this server sends it; a present one is
		// a component this rewrite does not walk.
		hasDisplayName, err := copyBoolean(in, out)
		if err != nil {
			return err
		}

		if hasDisplayName {
			return errors.New("player info carries a display name, which this rewrite does not walk")
		}

		// The profile key, read so that it is consumed and never written.
		hasProfileKey, err := in.ReadBoolean()
		if err != nil {
			return err
		}

		if hasProfileKey {
			return errors.New("player info carries a profile key, which 1.18.2 has no field for")
		}
	}

	return nil
}

// DowngradeLoginSuccessTo1_18_2 rewrites the login success packet from what
// 1.19 sends into what 1.18.2 reads.
//
// 1.19 is where the profile's properties joined the packet, behind the uuid
// and the name; 1.18.2 reads the two and nothing after them. So the
// properties come off the end, read so that they are consumed and never
// written. The client is not left without them: the player list entry that
// adds the player carries the same properties on every version, and that
// is where a 1.18.2 client takes its own skin from.
func DowngradeLoginSuccessTo1_18_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The profile's uuid.
	if err := copyBytes(in, out, uuidSize); err != nil {
		return err
	}

	// The username.
	if err := copyString(in, out); err != nil {
		return err
	}

	return skipProfileProperties(in)
}

// skipProfileProperties reads a game profile's property block -- a counted
// array of name, value and an optional signature -- so that it is consumed,
// and writes nothing.
func skipProfileProperties(in *streams.MinecraftStream) error {
	properties, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	for range properties {
		// The property's name and value.
		for range 2 {
			if _, err := in.ReadString(); err != nil {
				return err
			}
		}

		signed, err := in.ReadBoolean()
		if err != nil {
			return err
		}

		if signed {
			if _, err := in.ReadString(); err != nil {
				return err
			}
		}
	}

	return nil
}

// UpgradeLoginStartFrom1_18_2 rewrites the login start packet from what
// 1.18.2 sends into what 1.19 reads.
//
// 1.18.2's hello is the name and nothing else, and 1.19 reads an optional
// profile key behind it, which 1.18.2 has nothing to put there: so the name
// is copied and the key goes on the end absent. What comes out is what the
// 1.19.1 step reads from a 1.19 client that holds no key, and the steps
// above settle the rest.
func UpgradeLoginStartFrom1_18_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyRest(in, out); err != nil {
		return err
	}

	return out.WriteBoolean(false)
}

// UpgradeEncryptionResponseFrom1_18_2 rewrites the encryption response from
// what 1.18.2 sends into what 1.19 reads.
//
// Both open with the shared secret under the server's key. 1.18.2 follows
// it with the challenge under the same key, and 1.19 with one of two things
// -- that, or a salt and a signature from a client holding a profile key --
// with a flag in front saying which. A 1.18.2 client has no key and always
// encrypts, so the flag goes in set and the challenge after it as it came.
// What comes out is what the 1.19.3 step reads from a 1.19 client that
// encrypted its challenge, and it carries the challenge across as it is.
func UpgradeEncryptionResponseFrom1_18_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The shared secret.
	if err := copyByteArray(in, out); err != nil {
		return err
	}

	if err := out.WriteBoolean(true); err != nil {
		return err
	}

	// The encrypted challenge.
	return copyByteArray(in, out)
}
