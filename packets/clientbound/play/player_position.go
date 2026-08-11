package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// PlayerPositionClientboundPacket moves the client's own player. It is what
// places a joining client at its spawn, since the login packet says nothing
// about where the player is.
//
// The client replies with an accept teleportation packet carrying TeleportId
// back, followed by its own position. A server that tracks positions is
// expected to ignore what the client reports until that acknowledgement
// arrives, because everything in flight was measured before the move.
type PlayerPositionClientboundPacket struct {
	TeleportId int32

	X float64
	Y float64
	Z float64

	// The delta movement the player is left with. Zero stops it dead, which is
	// what a teleport onto a spawn point wants.
	DeltaX float64
	DeltaY float64
	DeltaZ float64

	Yaw   float32
	Pitch float32

	// Relatives marks fields the client should add to its current values
	// instead of replacing them, as a bitmask in the order x, y, z, yaw, pitch,
	// delta x, delta y, delta z, rotate delta. Zero makes every field absolute.
	Relatives int32
}

func (p *PlayerPositionClientboundPacket) String() string {
	return fmt.Sprintf("PlayerPositionClientboundPacket{TeleportId:%d X:%g Y:%g Z:%g DeltaX:%g DeltaY:%g DeltaZ:%g Yaw:%g Pitch:%g Relatives:%#x}",
		p.TeleportId, p.X, p.Y, p.Z, p.DeltaX, p.DeltaY, p.DeltaZ, p.Yaw, p.Pitch, p.Relatives)
}

func (p *PlayerPositionClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.TeleportId); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.X); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.Y); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.Z); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.DeltaX); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.DeltaY); err != nil {
		return err
	}

	if err := ms.WriteDouble(p.DeltaZ); err != nil {
		return err
	}

	if err := ms.WriteFloat(p.Yaw); err != nil {
		return err
	}

	if err := ms.WriteFloat(p.Pitch); err != nil {
		return err
	}

	// The mask is a plain int, not a VarInt.
	return ms.WriteInt(p.Relatives)
}
