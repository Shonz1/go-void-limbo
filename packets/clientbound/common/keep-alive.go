// Package common holds the clientbound packets that are the same packet in more
// than one phase. Keep alive is sent in configuration and in play, with a
// different id in each, and the client reads both with one codec.
package common

import (
	"fmt"
	"go-void-limbo/streams"
)

// KeepAliveClientboundPacket asks the client to prove it is still there. The
// client answers with a serverbound keep alive carrying the same id back, and
// answers nothing else about it.
//
// Both ends drop a connection that has been silent for thirty seconds, so a
// limbo with nothing to say still has to say this. Id is what ties an answer to
// the packet that asked for it; the value itself means nothing to the client.
type KeepAliveClientboundPacket struct {
	Id int64
}

func (p *KeepAliveClientboundPacket) String() string {
	return fmt.Sprintf("KeepAliveClientboundPacket{Id:%d}", p.Id)
}

func (p *KeepAliveClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	return ms.WriteLong(p.Id)
}
