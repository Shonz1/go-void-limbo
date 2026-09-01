package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// PlayerEntityTypeId is minecraft:player in the entity type registry at the
// latest version. The number is the registry's, not the protocol's, so it
// moves whenever a version adds an entity type that sorts before it; the
// versions that number it differently get a transformer rewriting it on the
// way down.
const PlayerEntityTypeId = 156

// Angle converts degrees to the single byte the protocol packs an angle into:
// 256 steps around the circle. Every rotation an entity packet carries is one
// of these, except the head rotation the client is told about itself, which
// stays a float.
func Angle(degrees float32) byte {
	return byte(degrees * 256 / 360)
}

// AddEntityClientboundPacket puts an entity into the client's world. It is how
// one player learns another exists: the entry the player list holds is just a
// name and a skin, and this is what makes a body appear at a position.
//
// The client refuses to spawn a player entity whose uuid it has no player list
// entry for, so the info update announcing the player has to arrive before
// this does.
type AddEntityClientboundPacket struct {
	EntityId int32

	// Uuid ties the entity to its player list entry, which is where the client
	// finds the skin to draw on it.
	Uuid string

	EntityTypeId int32

	X float64
	Y float64
	Z float64

	Yaw     float32
	Pitch   float32
	HeadYaw float32

	// Data parameterises a few entity types -- the block a falling block is,
	// the direction a painting hangs -- and means nothing for a player.
	Data int32
}

func (p *AddEntityClientboundPacket) String() string {
	return fmt.Sprintf("AddEntityClientboundPacket{EntityId:%d Uuid:%s EntityTypeId:%d X:%g Y:%g Z:%g Yaw:%g Pitch:%g HeadYaw:%g Data:%d}",
		p.EntityId, p.Uuid, p.EntityTypeId, p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.HeadYaw, p.Data)
}

func (p *AddEntityClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.EntityId); err != nil {
		return err
	}

	if err := ms.WriteUuid(p.Uuid); err != nil {
		return err
	}

	if err := ms.WriteVarInt(p.EntityTypeId); err != nil {
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

	// The velocity, as the quantized vector 1.21.11 introduced, whose zero is
	// the one byte written here. A player spawned by another player's join is
	// standing still, and the moves that follow carry their own deltas, so
	// nothing more of the encoding is implemented.
	if err := ms.WriteByte(0x00); err != nil {
		return err
	}

	if err := ms.WriteByte(Angle(p.Pitch)); err != nil {
		return err
	}

	if err := ms.WriteByte(Angle(p.Yaw)); err != nil {
		return err
	}

	if err := ms.WriteByte(Angle(p.HeadYaw)); err != nil {
		return err
	}

	return ms.WriteVarInt(p.Data)
}
