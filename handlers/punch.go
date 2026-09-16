package handlers

import (
	"fmt"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
	"github.com/Shonz1/go-void-limbo/types"
)

// HandlePunchServerboundPacket plays a client's arm swing on everyone else's
// view of it, which is all a punch means on a server with nothing to hit.
func HandlePunchServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	if _, ok := packet.(*serverboundPlay.PunchServerboundPacket); !ok {
		return fmt.Errorf("expected *play.PunchServerboundPacket, got %T", packet)
	}

	client.SyncSwing()

	return nil
}
