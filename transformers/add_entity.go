package transformers

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The id minecraft:player has in each version's entity type registry. The
// registry gains entries almost every version, and the ones that sort before
// the player shift its number, so the add entity packet -- encoded once, with
// the latest id -- has its type field rewritten at every step where the number
// moved. Each id here is read out of that version's own registry report.
const (
	playerEntityType26_2    = 156
	playerEntityType26_1    = 155
	playerEntityType1_21_11 = 155
	playerEntityType1_21_9  = 151
	playerEntityType1_21_7  = 149
	playerEntityType1_21_6  = 149
	playerEntityType1_21_5  = 148
)

// downgradeAddEntityType rewrites the entity type field of an add entity
// packet, the only field of it that registry renumbering moves. The fields in
// front of the type are fixed width, and everything after it is copied
// untouched.
func downgradeAddEntityType(in *streams.MinecraftStream, out *streams.MinecraftStream, from, to int32) error {
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

	// The player is the only entity this server spawns, so the one id is the
	// whole mapping, and anything else is a packet this transformer was never
	// taught.
	if entityType != from {
		return fmt.Errorf("add entity carries entity type %d, expected the player's %d", entityType, from)
	}

	if err := out.WriteVarInt(to); err != nil {
		return err
	}

	return copyRest(in, out)
}
