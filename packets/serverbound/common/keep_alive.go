// Package common holds the serverbound packets that are the same packet in more
// than one phase. Keep alive arrives in configuration and in play, with a
// different id in each, and one decoder reads both.
package common

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// KeepAliveServerboundPacket answers a clientbound keep alive. Id is the one
// that packet carried, and the client sends it back unchanged, which is the
// whole of what it says: an answer that does not match what was asked is an
// answer to nothing.
type KeepAliveServerboundPacket struct {
	Id int64
}

func (p *KeepAliveServerboundPacket) String() string {
	return fmt.Sprintf("KeepAliveServerboundPacket{Id:%d}", p.Id)
}

func DecodeKeepAliveServerboundPacket(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	id, err := ms.ReadLong()
	if err != nil {
		return nil, err
	}

	return &KeepAliveServerboundPacket{Id: id}, nil
}
