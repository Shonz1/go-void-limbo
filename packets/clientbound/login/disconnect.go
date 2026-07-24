package login

import (
	"fmt"
	"go-void-limbo/streams"
)

type DisconnectClientboundPacket struct {
	Reason string
}

func (p *DisconnectClientboundPacket) String() string {
	return fmt.Sprintf("DisconnectClientboundPacket{Reason:%s}", p.Reason)
}

func (p *DisconnectClientboundPacket) Encode(minecraftStream *streams.MinecraftStream) error {
	return minecraftStream.WriteString(p.Reason)
}
