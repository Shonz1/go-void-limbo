package handlers

import (
	"fmt"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
)

func HandleLoginStartServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	_, ok := packet.(*login.LoginStartServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *login.LoginStartServerboundPacket, got %T", packet)
	}

	disconnect := clientboundLogin.DisconnectClientboundPacket{Reason: `{"text": "TODO"}`}

	return client.WritePacket(&disconnect)
}
