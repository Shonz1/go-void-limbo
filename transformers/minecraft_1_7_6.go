package transformers

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.7.6 step is where the spawn player took on the skin. 1.7.2 below it
// reads the position straight behind the name, and nothing else this server
// sends moved: see the 1.7.6 step in package types. The two jars were
// compared class by class with the names taken out: every phase registers
// the same packets in the same order in both, so the ids are 1.7.6's
// throughout, and of the ninety-nine packets each registers, five differ.
// 1.7.6 is where the uuid became a java.util.UUID: its login success and its
// spawn player parse the text 1.7.2 keeps as text, and its login start keeps
// one where 1.7.2 keeps none, which is the same bytes on the wire for all
// three. The set slot's window id is a short where 1.7.6 reads a byte, and
// the plugin message's payload is capped at a short where 1.7.6 caps it at a
// megabyte, and this server sends neither. The spawn player below is the
// one packet that differs in what this server sends. The keep alive, the
// chunk, the player position, the player list, the entity metadata and
// everything this server reads are the same classes with the names taken
// out. The blocks are 1.7.6's to the last number: see package gamedata.

// DowngradeAddEntityTo1_7_2 rewrites the spawn player -- which is what the
// add entity packet has been since the 1.20.2 step -- from what 1.7.6 reads
// into what 1.7.2 reads: the same entity id, the uuid as text and the name,
// and then the position, the rotation, the held item and the metadata as
// they came, with the skin properties between the name and the position
// taken out. 1.7.6 reads those as a count and then a name, a value and a
// signature for each; 1.7.2 has no skin to read, and takes a name's skin
// from the session server by the name alone.
//
// A count below zero is refused: 1.7.6 reads it as none, and this server
// never sends one.
func DowngradeAddEntityTo1_7_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	// The uuid, spelled out, and the name.
	if err := copyString(in, out); err != nil {
		return err
	}

	if err := copyString(in, out); err != nil {
		return err
	}

	count, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if count < 0 {
		return fmt.Errorf("spawn player carries %d properties, which is fewer than none", count)
	}

	for i := int32(0); i < count; i++ {
		// The property's name, its value and its signature.
		for range 3 {
			if _, err := in.ReadString(); err != nil {
				return err
			}
		}
	}

	// The position as three ints, the yaw, the pitch, the held item and the
	// metadata.
	rest, err := in.ReadRest()
	if err != nil {
		return err
	}

	return out.WriteBytes(rest)
}
