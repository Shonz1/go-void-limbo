package handlers

import (
	"fmt"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
	"github.com/Shonz1/go-void-limbo/types"
)

// HandlePlayerInputServerboundPacket records the movement keys a client is
// holding. One of them is a stance the other players can see: the shift key is
// sneaking. The sprint key is not sprinting -- a player sprints without it on
// a double tap forward, and holds it without sprinting while standing still --
// so that stance is left to the player command packet, which reports it as
// what it is.
func HandlePlayerInputServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*serverboundPlay.PlayerInputServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *play.PlayerInputServerboundPacket, got %T", packet)
	}

	client.SyncSneaking(p.Sneak)

	return nil
}
