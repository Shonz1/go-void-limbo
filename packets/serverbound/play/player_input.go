package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// The bits of the one flag byte a player input packet is. The movement keys
// are decoded along with the rest for the sake of a complete packet, but only
// the last two matter here: sneaking and sprinting are the two stances other
// players can see, and this packet is where the client reports them -- the
// player command packet lost its shift actions when this packet gained the
// bits.
const (
	playerInputFlagForward  = 0x01
	playerInputFlagBackward = 0x02
	playerInputFlagLeft     = 0x04
	playerInputFlagRight    = 0x08
	playerInputFlagJump     = 0x10
	playerInputFlagSneak    = 0x20
	playerInputFlagSprint   = 0x40
)

// PlayerInputServerboundPacket reports which movement keys the player is
// holding, sent whenever the set of them changes.
type PlayerInputServerboundPacket struct {
	Forward  bool
	Backward bool
	Left     bool
	Right    bool
	Jump     bool
	Sneak    bool
	Sprint   bool
}

func (p *PlayerInputServerboundPacket) String() string {
	return fmt.Sprintf("PlayerInputServerboundPacket{Forward:%t Backward:%t Left:%t Right:%t Jump:%t Sneak:%t Sprint:%t}",
		p.Forward, p.Backward, p.Left, p.Right, p.Jump, p.Sneak, p.Sprint)
}

func DecodePlayerInputServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	flags, err := ms.ReadByte()
	if err != nil {
		return nil, err
	}

	return &PlayerInputServerboundPacket{
		Forward:  flags&playerInputFlagForward != 0,
		Backward: flags&playerInputFlagBackward != 0,
		Left:     flags&playerInputFlagLeft != 0,
		Right:    flags&playerInputFlagRight != 0,
		Jump:     flags&playerInputFlagJump != 0,
		Sneak:    flags&playerInputFlagSneak != 0,
		Sprint:   flags&playerInputFlagSprint != 0,
	}, nil
}
