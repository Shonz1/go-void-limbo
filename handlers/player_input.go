package handlers

import (
	"fmt"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
	"github.com/Shonz1/go-void-limbo/types"
)

// HandlePlayerInputServerboundPacket records the movement keys a client is
// holding. Two of them are stances the other players can see -- sneak and
// sprint -- and the input packet is where the client reports both.
func HandlePlayerInputServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*serverboundPlay.PlayerInputServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *play.PlayerInputServerboundPacket, got %T", packet)
	}

	client.SyncInput(p.Sneak, p.Sprint)

	return nil
}
