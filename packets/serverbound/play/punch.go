package play

import (
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// PunchServerboundPacket reports the client swinging its main arm at nothing,
// which it does for every click that had nothing in reach to hit. It carries
// no fields: 26.3 is where the swing packet, which named the arm, became
// this one, and a swing that hits something or uses an item reaches a server
// through the packets that say so instead. Every version before 26.3 sends
// the swing with its hand, which the 26.2 step's upgrade takes off.
type PunchServerboundPacket struct{}

func (p *PunchServerboundPacket) String() string {
	return "PunchServerboundPacket{}"
}

func DecodePunchServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	return &PunchServerboundPacket{}, nil
}
