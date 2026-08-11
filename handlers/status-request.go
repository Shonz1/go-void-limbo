package handlers

import (
	"fmt"
	clientboundStatus "go-void-limbo/packets/clientbound/status"
	"go-void-limbo/packets/serverbound/status"
	"go-void-limbo/types"
)

// HandleStatusRequestServerboundPacket answers a client asking what this server
// is with everything it says about itself.
//
// What that is belongs to the connection rather than to this packet: the counts
// and the description are the server's, and the version reported depends on
// which one is asking. So there is nothing to decide here, and a ping is
// answered without a word about the client having been kept.
func HandleStatusRequestServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	_, ok := packet.(*status.StatusRequestServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *status.StatusRequestServerboundPacket, got %T", packet)
	}

	response := clientboundStatus.StatusResponseClientboundPacket{Status: client.ServerStatus()}

	return client.WritePacket(&response)
}
