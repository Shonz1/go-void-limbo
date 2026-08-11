package handlers

import (
	"fmt"
	"go-void-limbo/packets/serverbound/common"
	"go-void-limbo/types"
)

// HandleKeepAliveServerboundPacket takes the client's answer to a keep alive as
// proof the connection is still live. The same packet arrives in configuration
// and in play, and means the same thing in both.
//
// Nothing is sent back: the server asks and the client answers, never the other
// way round.
func HandleKeepAliveServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	keepAlive, ok := packet.(*common.KeepAliveServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *common.KeepAliveServerboundPacket, got %T", packet)
	}

	return client.ConfirmKeepAlive(keepAlive.Id)
}
