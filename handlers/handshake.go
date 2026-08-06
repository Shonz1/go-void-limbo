package handlers

import (
	"fmt"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/types"
)

func HandleHandshakeServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*handshake.HandshakeServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *handshake.HandshakeServerboundPacket, got %T", packet)
	}

	client.SetProtocolVersion(types.GetProtocolVersionById(types.ProtocolId(p.ProtocolVersion)))
	client.SetPhase(types.Phase(p.Intent))

	return nil
}
