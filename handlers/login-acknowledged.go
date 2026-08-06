package handlers

import (
	"fmt"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
)

func HandleLoginAcknowledgedServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	_, ok := packet.(*login.LoginAcknowledgedServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *login.LoginAcknowledgedServerboundPacket, got %T", packet)
	}

	// The phase has to move first: WritePacket resolves the packet id from the
	// phase the client is currently in, and finish configuration is only
	// registered for the configuration phase.
	client.SetPhase(types.PhaseConfiguration)

	// The limbo has nothing to configure, so the configuration phase is finished
	// as soon as it starts.
	finishConfiguration := clientboundConfiguration.FinishConfigurationClientboundPacket{}

	return client.WritePacket(&finishConfiguration)
}
