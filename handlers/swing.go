package handlers

import (
	"fmt"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
	"github.com/Shonz1/go-void-limbo/types"
)

// HandleSwingServerboundPacket plays a client's arm swing on everyone else's
// view of it, which is all a swing means on a server with nothing to hit.
func HandleSwingServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*serverboundPlay.SwingServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *play.SwingServerboundPacket, got %T", packet)
	}

	client.SyncSwing(p.OffHand)

	return nil
}
