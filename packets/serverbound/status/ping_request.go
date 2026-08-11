package status

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// PingRequestServerboundPacket asks the server to send a number straight back,
// which is how a client measures what it draws as the connection's latency.
//
// Payload means nothing to the server. The client picks it, times how long it
// takes to come back, and is the only end that ever reads it, so the one thing
// this end can get wrong about it is changing it.
type PingRequestServerboundPacket struct {
	Payload int64
}

func (p *PingRequestServerboundPacket) String() string {
	return fmt.Sprintf("PingRequestServerboundPacket{Payload:%d}", p.Payload)
}

func DecodePingRequestServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	payload, err := ms.ReadLong()
	if err != nil {
		return nil, err
	}

	return &PingRequestServerboundPacket{Payload: payload}, nil
}
