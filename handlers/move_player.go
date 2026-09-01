package handlers

import (
	"fmt"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
	"github.com/Shonz1/go-void-limbo/types"
)

// The move player handlers pass what a client reports about its own movement
// to the connection's player sync, which is how the other players see it move.
// The four packets carry four subsets of the same state, and each handler
// hands over exactly the fields its packet carried, leaving the rest at their
// last reported values.
//
// A limbo takes the client entirely at its word here. There is no terrain to
// cheat through and nothing to gain by teleporting, so validating movement
// would be simulating a world this server does not have.

func HandleMovePlayerPositionServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*serverboundPlay.MovePlayerPositionServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *play.MovePlayerPositionServerboundPacket, got %T", packet)
	}

	client.SyncPosition(p.X, p.Y, p.Z, p.OnGround)

	return nil
}

func HandleMovePlayerPositionRotationServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*serverboundPlay.MovePlayerPositionRotationServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *play.MovePlayerPositionRotationServerboundPacket, got %T", packet)
	}

	client.SyncPositionRotation(p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.OnGround)

	return nil
}

func HandleMovePlayerRotationServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*serverboundPlay.MovePlayerRotationServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *play.MovePlayerRotationServerboundPacket, got %T", packet)
	}

	client.SyncRotation(p.Yaw, p.Pitch, p.OnGround)

	return nil
}

func HandleMovePlayerStatusServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*serverboundPlay.MovePlayerStatusServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *play.MovePlayerStatusServerboundPacket, got %T", packet)
	}

	client.SyncGround(p.OnGround)

	return nil
}
