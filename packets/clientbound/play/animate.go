package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// Animation is one of the handful of one-shot entity animations the animate
// packet plays. Only the two arm swings are named here, because they are the
// two a client asks for: the rest -- waking up, critical hit sparks -- are
// consequences of combat and sleep this server does not simulate.
type Animation byte

const (
	AnimationSwingMainArm Animation = 0
	AnimationSwingOffhand Animation = 3
)

func (a Animation) String() string {
	switch a {
	case AnimationSwingMainArm:
		return "swing_main_arm"
	case AnimationSwingOffhand:
		return "swing_offhand"
	}

	return fmt.Sprintf("Animation(%d)", byte(a))
}

// AnimateClientboundPacket plays a one-shot animation on an entity. It is the
// far side of the swing packet: one client says it swung, and this is the
// swing everyone else sees.
type AnimateClientboundPacket struct {
	EntityId  int32
	Animation Animation
}

func (p *AnimateClientboundPacket) String() string {
	return fmt.Sprintf("AnimateClientboundPacket{EntityId:%d Animation:%s}", p.EntityId, p.Animation)
}

func (p *AnimateClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.EntityId); err != nil {
		return err
	}

	return ms.WriteByte(byte(p.Animation))
}
