// Package play holds the serverbound packets of the play phase.
package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// AcceptTeleportationServerboundPacket confirms the client applied a player
// position packet. TeleportId is the one that packet carried, which is how a
// server tells the acknowledgement of one teleport from another when several
// are in flight. The position and the rotation are where the client put
// itself for it, which 26.3 added: every version before it acknowledges with
// the id alone, and the 26.2 step's upgrade fills the rest in.
type AcceptTeleportationServerboundPacket struct {
	TeleportId int32

	X float64
	Y float64
	Z float64

	Yaw   float32
	Pitch float32
}

func (p *AcceptTeleportationServerboundPacket) String() string {
	return fmt.Sprintf("AcceptTeleportationServerboundPacket{TeleportId:%d X:%g Y:%g Z:%g Yaw:%g Pitch:%g}",
		p.TeleportId, p.X, p.Y, p.Z, p.Yaw, p.Pitch)
}

func DecodeAcceptTeleportationServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	teleportId, err := ms.ReadVarInt()
	if err != nil {
		return nil, err
	}

	x, err := ms.ReadDouble()
	if err != nil {
		return nil, err
	}

	y, err := ms.ReadDouble()
	if err != nil {
		return nil, err
	}

	z, err := ms.ReadDouble()
	if err != nil {
		return nil, err
	}

	yaw, err := ms.ReadFloat()
	if err != nil {
		return nil, err
	}

	pitch, err := ms.ReadFloat()
	if err != nil {
		return nil, err
	}

	return &AcceptTeleportationServerboundPacket{TeleportId: teleportId, X: x, Y: y, Z: z, Yaw: yaw, Pitch: pitch}, nil
}
