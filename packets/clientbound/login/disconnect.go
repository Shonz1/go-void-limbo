package login

import (
	"fmt"
	"go-void-limbo/streams"
)

type DisconnectClientboundPacket struct {
	Reason string
}

func (p *DisconnectClientboundPacket) ToString() string {
	return fmt.Sprintf("DisconnectClientboundPacket %v", p)
}

func (p *DisconnectClientboundPacket) Encode(minecraftStream *streams.MinecraftStream) error {
	return minecraftStream.WriteString(p.Reason)
}
