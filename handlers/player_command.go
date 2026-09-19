package handlers

import (
	"fmt"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
	"github.com/Shonz1/go-void-limbo/types"
)

// HandlePlayerCommandServerboundPacket shows the other players the stances a
// client reports starting and stopping: sprinting at every version, and
// sneaking at the ones below 1.21.6, whose shift actions reach here renumbered.
// The rest -- beds, mounts, elytra -- are nothing a limbo has.
func HandlePlayerCommandServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*serverboundPlay.PlayerCommandServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *play.PlayerCommandServerboundPacket, got %T", packet)
	}

	switch p.Action {
	case serverboundPlay.PlayerCommandStartSprinting:
		client.SyncSprinting(true)
	case serverboundPlay.PlayerCommandStopSprinting:
		client.SyncSprinting(false)
	case serverboundPlay.PlayerCommandPressShiftKey:
		client.SyncSneaking(true)
	case serverboundPlay.PlayerCommandReleaseShiftKey:
		client.SyncSneaking(false)
	}

	return nil
}
