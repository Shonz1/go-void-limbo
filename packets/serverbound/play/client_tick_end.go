package play

import (
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// ClientTickEndServerboundPacket closes a client tick, marking everything sent
// before it as belonging to that tick. It carries no fields and arrives twenty
// times a second for as long as the client is in play.
type ClientTickEndServerboundPacket struct{}

func (p *ClientTickEndServerboundPacket) String() string {
	return "ClientTickEndServerboundPacket{}"
}

func DecodeClientTickEndServerboundPacket(_ *streams.MinecraftStream) (types.ServerboundPacket, error) {
	return &ClientTickEndServerboundPacket{}, nil
}
