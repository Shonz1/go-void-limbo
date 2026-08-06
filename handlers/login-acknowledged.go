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

	// The client assembles its registries from these packets and needs them
	// before it can make sense of anything the play phase refers to by id, so
	// they go out before it is told configuration is over.
	for _, registry := range client.RegistryPackets() {
		if err := client.WritePacket(registry); err != nil {
			return err
		}
	}

	// The limbo has nothing left to configure once the registries are sent.
	finishConfiguration := clientboundConfiguration.FinishConfigurationClientboundPacket{}

	return client.WritePacket(&finishConfiguration)
}
