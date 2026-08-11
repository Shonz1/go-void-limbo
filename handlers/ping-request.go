package handlers

import (
	"fmt"
	clientboundStatus "go-void-limbo/packets/clientbound/status"
	"go-void-limbo/packets/serverbound/status"
	"go-void-limbo/types"
)

// HandlePingRequestServerboundPacket sends a ping request's number back, which
// is the last thing said on a connection that only came to ask: the client
// times the answer, draws it as the latency beside the server, and closes.
func HandlePingRequestServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	pingRequest, ok := packet.(*status.PingRequestServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *status.PingRequestServerboundPacket, got %T", packet)
	}

	// Unchanged, and not a number of this end's own: the client matches the
	// answer against what it sent, and one that does not match is a latency it
	// cannot measure.
	pong := clientboundStatus.PongResponseClientboundPacket{Payload: pingRequest.Payload}

	return client.WritePacket(&pong)
}
