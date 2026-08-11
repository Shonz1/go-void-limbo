package play

import (
	"fmt"
	"go-void-limbo/streams"
	"go-void-limbo/types"
)

// The client sends whichever of the four move player packets carries only what
// changed this tick, so a standing player still reports its state through the
// status one. Every variant ends in the same flag byte.
const (
	movePlayerFlagOnGround            = 0x01
	movePlayerFlagHorizontalCollision = 0x02
)

// MovePlayerStatus is the tail every move player packet shares: what the
// client's own movement ran into this tick.
type MovePlayerStatus struct {
	OnGround            bool
	HorizontalCollision bool
}

func (s MovePlayerStatus) String() string {
	return fmt.Sprintf("OnGround:%t HorizontalCollision:%t", s.OnGround, s.HorizontalCollision)
}

func decodeMovePlayerStatus(ms *streams.MinecraftStream) (MovePlayerStatus, error) {
	flags, err := ms.ReadByte()
	if err != nil {
		return MovePlayerStatus{}, err
	}

	return MovePlayerStatus{
		OnGround:            flags&movePlayerFlagOnGround != 0,
		HorizontalCollision: flags&movePlayerFlagHorizontalCollision != 0,
	}, nil
}

// MovePlayerPositionServerboundPacket reports a player that moved without
// turning.
type MovePlayerPositionServerboundPacket struct {
	X float64
	Y float64
	Z float64
	MovePlayerStatus
}

func (p *MovePlayerPositionServerboundPacket) String() string {
	return fmt.Sprintf("MovePlayerPositionServerboundPacket{X:%g Y:%g Z:%g %s}", p.X, p.Y, p.Z, p.MovePlayerStatus)
}

func DecodeMovePlayerPositionServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	x, y, z, err := readPosition(ms)
	if err != nil {
		return nil, err
	}

	status, err := decodeMovePlayerStatus(ms)
	if err != nil {
		return nil, err
	}

	return &MovePlayerPositionServerboundPacket{X: x, Y: y, Z: z, MovePlayerStatus: status}, nil
}

// MovePlayerPositionRotationServerboundPacket reports a player that both moved
// and turned. It is what a client sends back after being teleported.
type MovePlayerPositionRotationServerboundPacket struct {
	X     float64
	Y     float64
	Z     float64
	Yaw   float32
	Pitch float32
	MovePlayerStatus
}

func (p *MovePlayerPositionRotationServerboundPacket) String() string {
	return fmt.Sprintf("MovePlayerPositionRotationServerboundPacket{X:%g Y:%g Z:%g Yaw:%g Pitch:%g %s}",
		p.X, p.Y, p.Z, p.Yaw, p.Pitch, p.MovePlayerStatus)
}

func DecodeMovePlayerPositionRotationServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	x, y, z, err := readPosition(ms)
	if err != nil {
		return nil, err
	}

	yaw, pitch, err := readRotation(ms)
	if err != nil {
		return nil, err
	}

	status, err := decodeMovePlayerStatus(ms)
	if err != nil {
		return nil, err
	}

	return &MovePlayerPositionRotationServerboundPacket{X: x, Y: y, Z: z, Yaw: yaw, Pitch: pitch, MovePlayerStatus: status}, nil
}

// MovePlayerRotationServerboundPacket reports a player that turned on the spot.
type MovePlayerRotationServerboundPacket struct {
	Yaw   float32
	Pitch float32
	MovePlayerStatus
}

func (p *MovePlayerRotationServerboundPacket) String() string {
	return fmt.Sprintf("MovePlayerRotationServerboundPacket{Yaw:%g Pitch:%g %s}", p.Yaw, p.Pitch, p.MovePlayerStatus)
}

func DecodeMovePlayerRotationServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	yaw, pitch, err := readRotation(ms)
	if err != nil {
		return nil, err
	}

	status, err := decodeMovePlayerStatus(ms)
	if err != nil {
		return nil, err
	}

	return &MovePlayerRotationServerboundPacket{Yaw: yaw, Pitch: pitch, MovePlayerStatus: status}, nil
}

// MovePlayerStatusServerboundPacket reports a player that neither moved nor
// turned, which the client still sends so a server sees the flags change.
type MovePlayerStatusServerboundPacket struct {
	MovePlayerStatus
}

func (p *MovePlayerStatusServerboundPacket) String() string {
	return fmt.Sprintf("MovePlayerStatusServerboundPacket{%s}", p.MovePlayerStatus)
}

func DecodeMovePlayerStatusServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	status, err := decodeMovePlayerStatus(ms)
	if err != nil {
		return nil, err
	}

	return &MovePlayerStatusServerboundPacket{MovePlayerStatus: status}, nil
}

func readPosition(ms *streams.MinecraftStream) (float64, float64, float64, error) {
	x, err := ms.ReadDouble()
	if err != nil {
		return 0, 0, 0, err
	}

	y, err := ms.ReadDouble()
	if err != nil {
		return 0, 0, 0, err
	}

	z, err := ms.ReadDouble()
	if err != nil {
		return 0, 0, 0, err
	}

	return x, y, z, nil
}

func readRotation(ms *streams.MinecraftStream) (float32, float32, error) {
	yaw, err := ms.ReadFloat()
	if err != nil {
		return 0, 0, err
	}

	pitch, err := ms.ReadFloat()
	if err != nil {
		return 0, 0, err
	}

	return yaw, pitch, nil
}
