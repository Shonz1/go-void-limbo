package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// RotateHeadClientboundPacket turns an entity's head. For a player entity it
// is the packet that matters most of the three rotations: the client aims the
// head at what this says, and swings the body around to follow it, so a
// player relayed without these stares past everyone no matter what its move
// packets carried.
type RotateHeadClientboundPacket struct {
	EntityId int32
	HeadYaw  float32
}

func (p *RotateHeadClientboundPacket) String() string {
	return fmt.Sprintf("RotateHeadClientboundPacket{EntityId:%d HeadYaw:%g}", p.EntityId, p.HeadYaw)
}

func (p *RotateHeadClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.EntityId); err != nil {
		return err
	}

	return ms.WriteByte(Angle(p.HeadYaw))
}
