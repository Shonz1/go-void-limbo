package status

import (
	"fmt"
	"go-void-limbo/streams"
)

// PongResponseClientboundPacket sends a ping request's number back untouched.
// The round trip is what the client draws as latency, so the only thing this
// packet has to get right is being the same number as fast as possible.
type PongResponseClientboundPacket struct {
	Payload int64
}

func (p *PongResponseClientboundPacket) String() string {
	return fmt.Sprintf("PongResponseClientboundPacket{Payload:%d}", p.Payload)
}

func (p *PongResponseClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	return ms.WriteLong(p.Payload)
}
