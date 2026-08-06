package handlers

import (
	"fmt"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/types"
)

func HandleAcknowledgeFinishConfigurationServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	_, ok := packet.(*configuration.AcknowledgeFinishConfigurationServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *configuration.AcknowledgeFinishConfigurationServerboundPacket, got %T", packet)
	}

	client.SetPhase(types.PhasePlay)

	return nil
}
