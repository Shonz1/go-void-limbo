package transformers

import (
	"errors"
	"fmt"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.19.3 step is where the player list and the login were reworked, read
// off the 1.19.1 client's own classes through Mojang's mappings the way the
// 1.19.4 step was. 1.19.3 and 1.19.1 lay all but five of the packets this
// server speaks out alike -- the two player list packets, which 1.19.1 reads
// as one packet under an action of its own; the entity metadata, whose pose
// serializer 1.19.3's long serializer pushed up one; the hello, which 1.19.1
// sends with the profile key 1.19.3 took off; and the encryption response,
// which a 1.19.1 client holding that key signs rather than encrypts -- and
// each rewrite below is one of those, with one more for the play login,
// laid out alike but carrying registries that are 1.19.1's own. Everything
// else is wire-identical in 760 and 761: the encryption request, the login
// on both sides of the success packet, the add player packet, the player
// position with its dismount flag, the movement on both sides, the tags, and
// the chunk packet with its named heightmaps and its trust edges flag. 1.19.3 did renumber the play phase, in both directions,
// which the id tables say; this file is only about the bodies.

// The actions a 1.19.1 player info packet is one of, in the order its enum
// declares them, since the packet names its action by ordinal.
const (
	playerInfoAction1_19_1AddPlayer         int32 = 0
	playerInfoAction1_19_1UpdateGameMode    int32 = 1
	playerInfoAction1_19_1UpdateLatency     int32 = 2
	playerInfoAction1_19_1UpdateDisplayName int32 = 3
	playerInfoAction1_19_1RemovePlayer      int32 = 4
)

// DowngradePlayerInfoUpdateTo1_19_1 rewrites the player info update packet
// from what 1.19.3 sends into what 1.19.1 reads, which is a different packet.
//
// 1.19.3 is where the player list packet was split in two and given a mask
// of actions, every entry carrying whichever the mask names. Before it, one
// packet did the whole job under a single action, and every entry carried
// that action's fields alone: an added player came with its profile, its
// game mode, its latency, an optional display name and an optional profile
// key, all at once, and any other action changed one thing about a player
// the client already knew.
//
// So a packet that adds players goes out as an add, with the game mode and
// the latency the mask names or the defaults the client would have given a
// player it was told nothing about -- survival and none -- and with the
// display name and the profile key absent, the two things this server never
// sends. Whether the player is listed is a thing 1.19.1 has no field for,
// since it lists everyone it is told about, so the flag comes off. A packet
// that changes one thing about a player goes out as that one action, and a
// packet that would change two, or something 1.19.1 has no action for, is
// refused rather than sent as the wrong one.
//
// Two of the fields are optionals this server always writes absent -- the
// chat session and the display name -- and this rewrite copies the absence
// and refuses the presence, as the steps above do.
func DowngradePlayerInfoUpdateTo1_19_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	mask, err := in.ReadByte()
	if err != nil {
		return err
	}

	actions := play.PlayerInfoAction(mask)

	action, err := playerInfoAction1_19_1(actions)
	if err != nil {
		return err
	}

	if err := out.WriteVarInt(action); err != nil {
		return err
	}

	entries, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	for range entries {
		// The entry's uuid.
		if err := copyBytes(in, out, uuidSize); err != nil {
			return err
		}

		if actions&play.PlayerInfoAddPlayer != 0 {
			// The username.
			if err := copyString(in, out); err != nil {
				return err
			}

			if err := copyProfileProperties(in, out); err != nil {
				return err
			}
		}

		if actions&play.PlayerInfoInitializeChat != 0 {
			present, err := in.ReadBoolean()
			if err != nil {
				return err
			}

			if present {
				return errors.New("player info update carries a chat session, which this rewrite does not walk")
			}
		}

		// The game mode and the latency: read when the mask names them,
		// written when the action wants them, and the client's own defaults
		// when the action wants one the mask did not name.
		gameMode, err := readVarIntIf(in, actions&play.PlayerInfoUpdateGameMode != 0, defaultGameMode1_19_1)
		if err != nil {
			return err
		}

		if action == playerInfoAction1_19_1AddPlayer || action == playerInfoAction1_19_1UpdateGameMode {
			if err := out.WriteVarInt(gameMode); err != nil {
				return err
			}
		}

		if actions&play.PlayerInfoUpdateListed != 0 {
			if _, err := in.ReadBoolean(); err != nil {
				return err
			}
		}

		latency, err := readVarIntIf(in, actions&play.PlayerInfoUpdateLatency != 0, defaultLatency1_19_1)
		if err != nil {
			return err
		}

		if action == playerInfoAction1_19_1AddPlayer || action == playerInfoAction1_19_1UpdateLatency {
			if err := out.WriteVarInt(latency); err != nil {
				return err
			}
		}

		if actions&play.PlayerInfoUpdateDisplayName != 0 {
			present, err := in.ReadBoolean()
			if err != nil {
				return err
			}

			if present {
				return errors.New("player info update carries a display name, which this rewrite does not walk")
			}
		}

		// The display name, absent, and for an added player the profile key
		// after it, absent as well: a key is the client's to announce, and
		// this server keeps none.
		if action == playerInfoAction1_19_1AddPlayer || action == playerInfoAction1_19_1UpdateDisplayName {
			if err := out.WriteBoolean(false); err != nil {
				return err
			}
		}

		if action == playerInfoAction1_19_1AddPlayer {
			if err := out.WriteBoolean(false); err != nil {
				return err
			}
		}
	}

	return nil
}

// What a 1.19.1 client holds for a player it was told nothing about: the
// survival game mode, and no latency measured.
const (
	defaultGameMode1_19_1 int32 = 0
	defaultLatency1_19_1  int32 = 0
)

// playerInfoAction1_19_1 picks the one action a 1.19.1 packet can carry for a
// mask of 1.19.3's. An add is an add whatever else the mask names, since the
// add carries everything; otherwise the mask has to name exactly one thing
// 1.19.1 has an action for.
func playerInfoAction1_19_1(actions play.PlayerInfoAction) (int32, error) {
	if actions&play.PlayerInfoAddPlayer != 0 {
		return playerInfoAction1_19_1AddPlayer, nil
	}

	switch actions {
	case play.PlayerInfoUpdateGameMode:
		return playerInfoAction1_19_1UpdateGameMode, nil
	case play.PlayerInfoUpdateLatency:
		return playerInfoAction1_19_1UpdateLatency, nil
	case play.PlayerInfoUpdateDisplayName:
		return playerInfoAction1_19_1UpdateDisplayName, nil
	default:
		return 0, fmt.Errorf("player info update carries the actions %s, which 1.19.1 has no single action for", actions)
	}
}

// readVarIntIf reads a var int when present says one is there, and hands back
// fallback when it does not.
func readVarIntIf(in *streams.MinecraftStream, present bool, fallback int32) (int32, error) {
	if !present {
		return fallback, nil
	}

	return in.ReadVarInt()
}

// DowngradePlayerInfoRemoveTo1_19_1 rewrites the player info remove packet
// from what 1.19.3 sends into what 1.19.1 reads, which is the same packet
// the update goes out as, under its remove action: the action, and then the
// uuids as 1.19.3 lays them out, a count and then each one.
func DowngradePlayerInfoRemoveTo1_19_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := out.WriteVarInt(playerInfoAction1_19_1RemovePlayer); err != nil {
		return err
	}

	return copyRest(in, out)
}

// DowngradePlayLoginTo1_19_1 rewrites the play phase login packet from what
// 1.19.3 sends into what 1.19.1 reads, given the registries that version
// reads out of it.
//
// The two versions lay the packet out alike from front to back, and read the
// same three registries out of it -- the dimension types, the biomes and the
// chat types -- so the one thing to change is the compound in the middle,
// which is 1.19.3's, put there by the 1.19.4 step, and which 1.19.1 has its
// own of: the same three registries with its own content, encoded once by
// package gamedata. As with the steps above, a transformer built with no
// registries to write refuses every login rather than send a client into a
// world it cannot make sense of.
func DowngradePlayLoginTo1_19_1(registryCodec []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.19.1 play login carries the registries, and this transformer was given none")
		}

		return swapPlayLoginRegistries(in, out, registryCodec)
	}
}

// The pose serializer's id on 1.19.1: 1.19.3 is where the long serializer
// appeared, registered right after the int one and so in front of the pose,
// which moved up one, to the poseSerializer1_19_3 of the step above.
const poseSerializer1_19_1 = 18

// DowngradeSetEntityDataTo1_19_1 rewrites the set entity data packet from
// what 1.19.3 sends into what 1.19.1 reads: the same entries with the pose
// under the serializer id 1.19.1 gives it. The byte serializer sits at 0 in
// both, so the flags entry is copied as it is.
func DowngradeSetEntityDataTo1_19_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return renamePoseSerializer(in, out, poseSerializer1_19_3, poseSerializer1_19_1)
}

// The bounds on the two byte arrays of a profile key: the key is an encoded
// RSA public key, and the signature is Mojang's over it. Both are what the
// 1.19.1 client itself allows for.
const (
	maxProfileKeyLength          = 512
	maxProfileKeySignatureLength = 4096
)

// UpgradeLoginStartFrom1_19_1 rewrites the login start packet from what
// 1.19.1 sends into what 1.19.3 reads.
//
// 1.19 gave the client a profile key to sign its chat under, and the hello
// announced it: after the name, an optional holding the key's expiry, the
// key itself and Mojang's signature over the two. 1.19.3 moved the key into
// a packet of its own in the play phase and took it off the hello, so it
// comes off here: the name is copied, the key is read so that it is consumed
// and never written, and the optional uuid behind it is copied for the 1.20
// step to settle.
//
// The key is the one thing a signed encryption response could be checked
// against, and nothing above this step has a field to carry it in, so it is
// gone by the time the response arrives: see UpgradeEncryptionResponseFrom1_19_1.
func UpgradeLoginStartFrom1_19_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The username.
	if err := copyString(in, out); err != nil {
		return err
	}

	hasProfileKey, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	if hasProfileKey {
		// The expiry, a long of milliseconds.
		if _, err := in.ReadLong(); err != nil {
			return err
		}

		if _, err := in.ReadByteArray(maxProfileKeyLength); err != nil {
			return err
		}

		if _, err := in.ReadByteArray(maxProfileKeySignatureLength); err != nil {
			return err
		}
	}

	// The optional uuid.
	return copyRest(in, out)
}

// The bound on the signature a 1.19.1 client answers the challenge with,
// which is what the client itself allows for.
const maxChallengeSignatureLength = 4096

// UpgradeEncryptionResponseFrom1_19_1 rewrites the encryption response from
// what 1.19.1 sends into what 1.19.3 reads.
//
// Both open with the shared secret under the server's key. 1.19.3 follows it
// with the challenge under the same key, and 1.19.1 with one of two things:
// the challenge encrypted the same way, from a client without a profile key,
// or a salt and a signature over the challenge and the salt under the
// profile key the client announced in its hello, from a client holding one
// -- which is every client logged into an account. A flag in front says
// which.
//
// The encrypted challenge is carried across as it is, flag off. The
// signature is one this server cannot check: the key it was made under left
// the connection with the hello, since nothing above the 1.19.3 step has a
// field for it, and a signature checked against nothing is worth nothing. So
// it is read and dropped, and the response goes on with no challenge at all,
// which is what tells the connection the client signed instead: a version
// that may sign is let through on the session server's word alone, which is
// what settles the login on every version anyway, and a version that cannot
// sign is refused an empty challenge as it would be any wrong one. Nothing is
// written into the challenge's place that this server did not see the
// client prove, since an empty challenge is refused everywhere a real one is
// expected.
func UpgradeEncryptionResponseFrom1_19_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The shared secret.
	if err := copyByteArray(in, out); err != nil {
		return err
	}

	encrypted, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	if encrypted {
		return copyByteArray(in, out)
	}

	// The salt, a long, and the signature, read so that they are consumed and
	// never written.
	if _, err := in.ReadLong(); err != nil {
		return err
	}

	if _, err := in.ReadByteArray(maxChallengeSignatureLength); err != nil {
		return err
	}

	return out.WriteByteArray(nil)
}
