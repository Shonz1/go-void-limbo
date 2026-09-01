// Package play holds the clientbound packets of the play phase.
package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"strings"
)

// BlockPos is a block coordinate.
type BlockPos struct {
	X int32
	Y int32
	Z int32
}

func (p BlockPos) String() string {
	return fmt.Sprintf("BlockPos{X:%d Y:%d Z:%d}", p.X, p.Y, p.Z)
}

// pack lays the coordinates out the way the protocol sends them, as one long
// holding x in the top 26 bits, z in the next 26 and y in the low 12.
// Coordinates outside those ranges wrap rather than fail, and the client reads
// the wrapped value back as a different block.
func (p BlockPos) pack() int64 {
	return (int64(p.X)&0x3FFFFFF)<<38 | (int64(p.Z)&0x3FFFFFF)<<12 | int64(p.Y)&0xFFF
}

// GlobalPos is a block coordinate together with the dimension it is in.
type GlobalPos struct {
	Dimension string
	Position  BlockPos
}

func (p GlobalPos) String() string {
	return fmt.Sprintf("GlobalPos{Dimension:%s Position:%s}", p.Dimension, p.Position)
}

func (p GlobalPos) encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteString(p.Dimension); err != nil {
		return err
	}

	return ms.WriteLong(p.Position.pack())
}

// SpawnInfo describes the world the client is being placed in. The same block
// is sent again by respawn, which is why the protocol keeps it separate from
// the login packet that carries it here.
type SpawnInfo struct {
	// DimensionTypeId is an index into the minecraft:dimension_type registry as
	// it was sent during configuration, not a name. Package gamedata decides
	// what that registry holds, so the two have to agree: entry 0 there is the
	// dimension the client is told it is in.
	DimensionTypeId int32

	// Dimension is the name of the world itself, which is a different thing
	// from its type. It has to be one of the names the login packet lists.
	Dimension string

	// HashedSeed is used by the client only to place biome noise, which a
	// limbo without terrain has none of.
	HashedSeed int64

	GameMode types.GameMode

	// PreviousGameMode is GameModeNone on a fresh join, since there is no
	// earlier mode to return to.
	PreviousGameMode types.GameMode

	IsDebug bool
	IsFlat  bool

	// DeathLocation is where the client last died, nil when it has not. The
	// client points its recovery compass at it.
	DeathLocation *GlobalPos

	PortalCooldown int32
	SeaLevel       int32
}

func (s SpawnInfo) String() string {
	deathLocation := "none"
	if s.DeathLocation != nil {
		deathLocation = s.DeathLocation.String()
	}

	return fmt.Sprintf("SpawnInfo{DimensionTypeId:%d Dimension:%s HashedSeed:%d GameMode:%s PreviousGameMode:%s IsDebug:%t IsFlat:%t DeathLocation:%s PortalCooldown:%d SeaLevel:%d}",
		s.DimensionTypeId, s.Dimension, s.HashedSeed, s.GameMode, s.PreviousGameMode, s.IsDebug, s.IsFlat, deathLocation, s.PortalCooldown, s.SeaLevel)
}

func (s SpawnInfo) encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(s.DimensionTypeId); err != nil {
		return err
	}

	if err := ms.WriteString(s.Dimension); err != nil {
		return err
	}

	if err := ms.WriteLong(s.HashedSeed); err != nil {
		return err
	}

	if err := ms.WriteByte(byte(s.GameMode)); err != nil {
		return err
	}

	if err := ms.WriteByte(byte(s.PreviousGameMode)); err != nil {
		return err
	}

	if err := ms.WriteBoolean(s.IsDebug); err != nil {
		return err
	}

	if err := ms.WriteBoolean(s.IsFlat); err != nil {
		return err
	}

	if err := ms.WriteBoolean(s.DeathLocation != nil); err != nil {
		return err
	}

	if s.DeathLocation != nil {
		if err := s.DeathLocation.encode(ms); err != nil {
			return err
		}
	}

	if err := ms.WriteVarInt(s.PortalCooldown); err != nil {
		return err
	}

	return ms.WriteVarInt(s.SeaLevel)
}

// LoginClientboundPacket puts the client into a world. It is the first packet
// of the play phase and the client builds its level, its own player entity and
// its game mode from it, so nothing else in play means anything until it
// arrives.
type LoginClientboundPacket struct {
	// EntityId is the id of the client's own player entity. Everything the play
	// phase later says about that entity refers to it by this number.
	EntityId int32

	Hardcore bool

	// Dimensions is every world the client may be sent to without reconnecting.
	// It has to contain SpawnInfo.Dimension.
	Dimensions []string

	// MaxPlayers is only the number the client shows on the player list.
	MaxPlayers int32

	// ViewDistance is how far the client keeps chunks for. The client raises
	// anything below 2 to 2, since that is the smallest cache it can hold.
	ViewDistance int32

	SimulationDistance int32
	ReducedDebugInfo   bool
	ShowDeathScreen    bool
	DoLimitedCrafting  bool
	SpawnInfo          SpawnInfo
	OnlineMode         bool
	EnforcesSecureChat bool
}

func (p *LoginClientboundPacket) String() string {
	return fmt.Sprintf("LoginClientboundPacket{EntityId:%d Hardcore:%t Dimensions:[%s] MaxPlayers:%d ViewDistance:%d SimulationDistance:%d ReducedDebugInfo:%t ShowDeathScreen:%t DoLimitedCrafting:%t SpawnInfo:%s OnlineMode:%t EnforcesSecureChat:%t}",
		p.EntityId, p.Hardcore, strings.Join(p.Dimensions, " "), p.MaxPlayers, p.ViewDistance, p.SimulationDistance,
		p.ReducedDebugInfo, p.ShowDeathScreen, p.DoLimitedCrafting, p.SpawnInfo, p.OnlineMode, p.EnforcesSecureChat)
}

func (p *LoginClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	// The entity id is a plain int rather than a VarInt here, unlike everywhere
	// else the protocol names an entity.
	if err := ms.WriteInt(p.EntityId); err != nil {
		return err
	}

	if err := ms.WriteBoolean(p.Hardcore); err != nil {
		return err
	}

	if err := ms.WriteVarInt(int32(len(p.Dimensions))); err != nil {
		return err
	}

	for _, dimension := range p.Dimensions {
		if err := ms.WriteString(dimension); err != nil {
			return err
		}
	}

	if err := ms.WriteVarInt(p.MaxPlayers); err != nil {
		return err
	}

	if err := ms.WriteVarInt(p.ViewDistance); err != nil {
		return err
	}

	if err := ms.WriteVarInt(p.SimulationDistance); err != nil {
		return err
	}

	if err := ms.WriteBoolean(p.ReducedDebugInfo); err != nil {
		return err
	}

	if err := ms.WriteBoolean(p.ShowDeathScreen); err != nil {
		return err
	}

	if err := ms.WriteBoolean(p.DoLimitedCrafting); err != nil {
		return err
	}

	if err := p.SpawnInfo.encode(ms); err != nil {
		return err
	}

	if err := ms.WriteBoolean(p.OnlineMode); err != nil {
		return err
	}

	return ms.WriteBoolean(p.EnforcesSecureChat)
}
