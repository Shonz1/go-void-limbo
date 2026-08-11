package play

import (
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// PlayerLoadedServerboundPacket says the client is through its loading screen
// and in the world. It is the last packet of a join and carries no fields.
type PlayerLoadedServerboundPacket struct{}

func (p *PlayerLoadedServerboundPacket) String() string {
	return "PlayerLoadedServerboundPacket{}"
}

func DecodePlayerLoadedServerboundPacket(_ *streams.MinecraftStream) (types.ServerboundPacket, error) {
	return &PlayerLoadedServerboundPacket{}, nil
}
