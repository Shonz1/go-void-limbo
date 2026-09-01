package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// SwingServerboundPacket reports the client swinging an arm, which it does for
// every click whether or not anything was in reach to hit.
type SwingServerboundPacket struct {
	// OffHand says which arm swung. The wire carries the hand enum, whose two
	// values make it a boolean with a longer name.
	OffHand bool
}

func (p *SwingServerboundPacket) String() string {
	return fmt.Sprintf("SwingServerboundPacket{OffHand:%t}", p.OffHand)
}

func DecodeSwingServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	hand, err := ms.ReadVarInt()
	if err != nil {
		return nil, err
	}

	return &SwingServerboundPacket{OffHand: hand == 1}, nil
}
