package transformers

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.20.3 step is a small one, read off the 1.20.2 client's own classes
// through Mojang's mappings the way the 1.20.5 step was: every packet this
// server speaks is wire-identical in 764 and 765 -- the chunk packet, the
// one-compound registry data, the tags, the login on both sides of the
// configuration phase, the player list, the movement -- and the entity
// metadata numbers its serializers alike. Text components are the one
// serialization 1.20.3 changed, from JSON to NBT, and the only component this
// server sends is the login disconnect's, which stayed JSON. What is left is
// the registry renumbering below, and one packet 1.20.2 has no event for.

// DowngradeAddEntityTo1_20_2 rewrites the add entity packet from what 1.20.3
// sends into what 1.20.2 reads.
//
// The packet's shape is identical in the two, and what moved is the entity
// type registry behind one of its fields: 1.20.3 added the breeze and its
// wind charge, both of which sort before the player.
func DowngradeAddEntityTo1_20_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType1_20_3, playerEntityType1_20_2)
}

// DowngradeGameEventTo1_20_2 rewrites the game event packet from what 1.20.3
// sends into what 1.20.2 reads, which is a different packet.
//
// The one game event this server sends is the one that lets a joining client
// off its loading screen, and 1.20.3 is where that event appeared. A 1.20.2
// client reads an event it has no name for as nothing at all, and what its
// loading screen waits for instead is the default spawn position packet,
// which a vanilla server of either version sends at the same point of the
// join. So the id table names that packet for 1.20.2, and this is the body
// that goes under it: a block position packed into one long, then the spawn
// angle as a float. Neither is anything a limbo has a value for -- the client
// uses them to point its compasses and nothing else -- so both are zero, and
// an event other than the one this server sends is refused rather than
// turned into a spawn position it does not mean.
func DowngradeGameEventTo1_20_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	event, err := in.ReadByte()
	if err != nil {
		return err
	}

	if play.GameEvent(event) != play.GameEventStartWaitingForChunks {
		return fmt.Errorf("game event %s is not the one 1.20.2 reads as a default spawn position", play.GameEvent(event))
	}

	// The event's value, read so that it is consumed and never written.
	if _, err := in.ReadFloat(); err != nil {
		return err
	}

	// The position, packed the way every block position is on the wire, and
	// the angle.
	if err := out.WriteLong(0); err != nil {
		return err
	}

	return out.WriteFloat(0)
}
