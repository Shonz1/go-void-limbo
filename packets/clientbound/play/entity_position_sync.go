package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// EntityPositionSyncClientboundPacket puts an entity exactly where the server
// says it is, rotation and all. It is how one player's movement reaches the
// others here: the client walks the entity over in a few ticks rather than
// snapping it, so a stream of these -- one per move packet the moving client
// sends -- plays back as motion.
//
// Vanilla servers mostly send relative move packets and keep this for
// teleports, but the relative ones top out at eight blocks and accumulate
// rounding, which a server would have to track and correct. Absolute positions
// need neither, and a limbo relaying a handful of players is not the situation
// the deltas' smaller wire size was for.
type EntityPositionSyncClientboundPacket struct {
	EntityId int32

	X float64
	Y float64
	Z float64

	Yaw   float32
	Pitch float32

	OnGround bool
}

func (p *EntityPositionSyncClientboundPacket) String() string {
	return fmt.Sprintf("EntityPositionSyncClientboundPacket{EntityId:%d X:%g Y:%g Z:%g Yaw:%g Pitch:%g OnGround:%t}",
		p.EntityId, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.OnGround)
}

func (p *EntityPositionSyncClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.EntityId); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.X); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.Y); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.Z); err != nil {
		return err
	}

	// The delta movement the entity is left with, zero because the next
	// position is coming as another one of these rather than being
	// extrapolated from a velocity.
	for range 3 {
		if err := ms.WriteDouble(0); err != nil {
			return err
		}
	}

	if err := ms.WriteFloat(p.Yaw); err != nil {
		return err
	}

	if err := ms.WriteFloat(p.Pitch); err != nil {
		return err
	}

	return ms.WriteBoolean(p.OnGround)
}
