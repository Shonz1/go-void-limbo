package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// SwingAnimationType is how an arm swings: the whack of a punch, the stab of
// a spear, or nothing at all. It is a var int of the type's place in the
// client's own list, which is the order below.
type SwingAnimationType int32

const (
	SwingAnimationNone  SwingAnimationType = 0
	SwingAnimationWhack SwingAnimationType = 1
	SwingAnimationStab  SwingAnimationType = 2
)

func (t SwingAnimationType) String() string {
	switch t {
	case SwingAnimationNone:
		return "none"
	case SwingAnimationWhack:
		return "whack"
	case SwingAnimationStab:
		return "stab"
	}

	return fmt.Sprintf("SwingAnimationType(%d)", int32(t))
}

// DefaultSwingDuration is how many ticks the client's own default swing
// animation takes, which is the one an empty hand plays.
const DefaultSwingDuration int32 = 6

// SwingAnimationClientboundPacket plays an arm swing on an entity. It is the
// far side of the punch packet: one client says it swung, and this is the
// swing everyone else sees. 26.3 is where the swing left the animate packet,
// which named it by number among the one-shot animations, for a packet of its
// own that says which arm, how, and for how long; every version before it
// reads the animate packet, which the 26.3 step rewrites this one into.
type SwingAnimationClientboundPacket struct {
	EntityId int32

	// OffHand says which arm swings. The wire carries the hand enum, whose two
	// values make it a boolean with a longer name.
	OffHand bool

	Animation SwingAnimationType

	// Duration is how many ticks the swing takes.
	Duration int32
}

func (p *SwingAnimationClientboundPacket) String() string {
	return fmt.Sprintf("SwingAnimationClientboundPacket{EntityId:%d OffHand:%t Animation:%s Duration:%d}",
		p.EntityId, p.OffHand, p.Animation, p.Duration)
}

func (p *SwingAnimationClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.EntityId); err != nil {
		return err
	}

	hand := int32(0)
	if p.OffHand {
		hand = 1
	}

	if err := ms.WriteVarInt(hand); err != nil {
		return err
	}

	if err := ms.WriteVarInt(int32(p.Animation)); err != nil {
		return err
	}

	return ms.WriteVarInt(p.Duration)
}
