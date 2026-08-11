// Package play holds the serverbound packets of the play phase.
package play

import (
	"fmt"
	"go-void-limbo/streams"
	"go-void-limbo/types"
)

// AcceptTeleportationServerboundPacket confirms the client applied a player
// position packet. TeleportId is the one that packet carried, which is how a
// server tells the acknowledgement of one teleport from another when several
// are in flight.
type AcceptTeleportationServerboundPacket struct {
	TeleportId int32
}

func (p *AcceptTeleportationServerboundPacket) String() string {
	return fmt.Sprintf("AcceptTeleportationServerboundPacket{TeleportId:%d}", p.TeleportId)
}

func DecodeAcceptTeleportationServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	teleportId, err := ms.ReadVarInt()
	if err != nil {
		return nil, err
	}

	return &AcceptTeleportationServerboundPacket{TeleportId: teleportId}, nil
}
