package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// PlayerCommandAction is what a player command packet says the player did.
type PlayerCommandAction int32

// The actions as 26.3 numbers them, which is how 1.21.6 left them when it took
// the two shift actions off the front of the list and gave the player input
// packet their job.
//
// The last two are not 26.3's: they are the shift actions of the versions
// below 1.21.6, which the latest version has no number for. The 1.21.5 step
// moves them past the end of the list rather than dropping them, because
// below 1.21.2 this packet is the only place a player on foot says it is
// sneaking. No 26.3 client sends them.
const (
	PlayerCommandStopSleeping PlayerCommandAction = iota
	PlayerCommandStartSprinting
	PlayerCommandStopSprinting
	PlayerCommandStartRidingJump
	PlayerCommandStopRidingJump
	PlayerCommandOpenInventory
	PlayerCommandStartFallFlying
	PlayerCommandPressShiftKey
	PlayerCommandReleaseShiftKey
)

// PlayerCommandServerboundPacket reports something the player started or
// stopped doing. Sprinting is the one a limbo cares about: the player input
// packet has a bit for the sprint key, but the key is not the stance -- a
// double tap forward sprints without it -- and this packet is where every
// version reports the stance itself.
type PlayerCommandServerboundPacket struct {
	EntityId int32
	Action   PlayerCommandAction
	Data     int32
}

func (p *PlayerCommandServerboundPacket) String() string {
	return fmt.Sprintf("PlayerCommandServerboundPacket{EntityId:%d Action:%d Data:%d}", p.EntityId, p.Action, p.Data)
}

func DecodePlayerCommandServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	entityId, err := ms.ReadVarInt()
	if err != nil {
		return nil, err
	}

	action, err := ms.ReadVarInt()
	if err != nil {
		return nil, err
	}

	data, err := ms.ReadVarInt()
	if err != nil {
		return nil, err
	}

	return &PlayerCommandServerboundPacket{
		EntityId: entityId,
		Action:   PlayerCommandAction(action),
		Data:     data,
	}, nil
}
