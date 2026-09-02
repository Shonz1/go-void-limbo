package transformers

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

// DowngradeAddEntityTo1_21_2 rewrites the add entity packet from what 1.21.4
// sends into what 1.21.2 reads.
//
// The packet's shape is identical in the two, and what moved is the entity
// type registry behind one of its fields -- backwards, for once: 1.21.4
// retired the transient creaking, which sorted before the player, so the
// player's number is one higher on 1.21.2 than on 1.21.4.
func DowngradeAddEntityTo1_21_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType1_21_4, playerEntityType1_21_2)
}

// DowngradePlayerInfoUpdateTo1_21_2 rewrites the player info update packet
// from what 1.21.4 sends into what 1.21.2 reads.
//
// 1.21.4 gave the packet an eighth action, the hat, appended after the list
// order: one more bit in the mask and, for every entry, one boolean at the
// end. 1.21.2 has seven actions and reads the mask as a set of those, so the
// bit comes off the mask and the boolean off each entry, and the seven
// fields in front of it are walked across as they are.
//
// Two of those fields are optionals this server always writes absent -- the
// chat session and the display name -- and this rewrite copies the absence
// and refuses the presence, since a session is a key and a signature and a
// display name is a text component, and walking either is a job it was never
// given.
func DowngradePlayerInfoUpdateTo1_21_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	mask, err := in.ReadByte()
	if err != nil {
		return err
	}

	actions := play.PlayerInfoAction(mask)

	if err := out.WriteByte(byte(actions &^ play.PlayerInfoUpdateHat)); err != nil {
		return err
	}

	entries, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	for i := int32(0); i < entries; i++ {
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
			present, err := copyBoolean(in, out)
			if err != nil {
				return err
			}

			if present {
				return fmt.Errorf("player info update carries a chat session, which this rewrite does not walk")
			}
		}

		if actions&play.PlayerInfoUpdateGameMode != 0 {
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		}

		if actions&play.PlayerInfoUpdateListed != 0 {
			if _, err := copyBoolean(in, out); err != nil {
				return err
			}
		}

		if actions&play.PlayerInfoUpdateLatency != 0 {
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		}

		if actions&play.PlayerInfoUpdateDisplayName != 0 {
			present, err := copyBoolean(in, out)
			if err != nil {
				return err
			}

			if present {
				return fmt.Errorf("player info update carries a display name, which this rewrite does not walk")
			}
		}

		if actions&play.PlayerInfoUpdateListOrder != 0 {
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		}

		// The hat, read so that it is consumed and never written.
		if actions&play.PlayerInfoUpdateHat != 0 {
			if _, err := in.ReadBoolean(); err != nil {
				return err
			}
		}
	}

	return nil
}
