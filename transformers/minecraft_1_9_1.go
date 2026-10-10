package transformers

import (
	"fmt"
	"math"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.9.1 step is where the play login's dimension became an int. 1.9
// below it reads a byte there, and nothing else in the login moved: see the
// 1.9 step in package types. The two jars were compared class by class with
// the names taken out: every phase registers the same packets in the same
// order in both, so the ids are 1.9.1's throughout, and of the registered
// packets the login below is the one that differs on the wire. The keep
// alive, the chunk, the player position, the spawn player, the entity
// metadata and everything this server reads are the same classes with the
// names taken out, and so are the section and its container, the palettes,
// the bit array and the nibble arrays the chunk is made of. The packet
// buffer differs in the caps 1.9.1 put on its array readers, which read the
// same bytes, and the data watcher reads a serializer's id as a byte where
// 1.9.1 reads a var int, which is the same byte for every id there is. The
// blocks are 1.9.1's to the last state: see package gamedata.

// DowngradePlayLoginTo1_9 rewrites the play phase login packet from what
// 1.9.1 reads into what 1.9 reads: the same login, with the dimension, an
// int behind the game mode, as a byte.
//
// A dimension is one of three numbers, each of which a byte holds, so one
// outside a byte is refused rather than cut down to a number that names a
// dimension the server did not send.
func DowngradePlayLoginTo1_9(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id, a plain int, and the game mode's byte.
	if err := copyBytes(in, out, 5); err != nil {
		return err
	}

	dimension, err := in.ReadInt()
	if err != nil {
		return err
	}

	if dimension < math.MinInt8 || dimension > math.MaxInt8 {
		return fmt.Errorf("dimension %d, which no byte holds", dimension)
	}

	if err := out.WriteByte(byte(int8(dimension))); err != nil {
		return err
	}

	// The difficulty and the most players, a byte each.
	if err := copyBytes(in, out, 2); err != nil {
		return err
	}

	// The level type.
	if err := copyString(in, out); err != nil {
		return err
	}

	// Whether the debug screen is cut down.
	if _, err := copyBoolean(in, out); err != nil {
		return err
	}

	return nil
}
