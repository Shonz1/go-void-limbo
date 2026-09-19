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
//
// 26.3 describes the position as a path: a type, and for the one type this
// server sends -- a straight line -- the point it ends at. The other type
// spells out the steps in between, which is for an entity the server moved
// through several points in one tick, and nothing here does. 26.2 and every
// version before it read the point and then the delta movement the entity is
// left with, which the 26.3 step puts back as zero.
type EntityPositionSyncClientboundPacket struct {
	EntityId int32

	X float64
	Y float64
	Z float64

	Yaw   float32
	Pitch float32

	OnGround bool
}

// positionPathLinear is the type of a path that is one straight line to its
// end, the first of the two types the client knows.
const positionPathLinear int32 = 0

func (p *EntityPositionSyncClientboundPacket) String() string {
	return fmt.Sprintf("EntityPositionSyncClientboundPacket{EntityId:%d X:%g Y:%g Z:%g Yaw:%g Pitch:%g OnGround:%t}",
		p.EntityId, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.OnGround)
}

func (p *EntityPositionSyncClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.EntityId); err != nil {
		return err
	}

	if err := ms.WriteVarInt(positionPathLinear); err != nil {
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

	if err := ms.WriteFloat(p.Yaw); err != nil {
		return err
	}

	if err := ms.WriteFloat(p.Pitch); err != nil {
		return err
	}

	return ms.WriteBoolean(p.OnGround)
}
