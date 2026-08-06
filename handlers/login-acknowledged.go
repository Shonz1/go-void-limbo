package handlers

import (
	"fmt"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
)

func HandleLoginAcknowledgedServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	_, ok := packet.(*login.LoginAcknowledgedServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *login.LoginAcknowledgedServerboundPacket, got %T", packet)
	}

	client.SetPhase(types.PhaseConfiguration)

	return nil
}
