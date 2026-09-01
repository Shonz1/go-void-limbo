package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"strings"
)

// PlayerInfoRemoveClientboundPacket takes players off the client's player
// list. It is the list half of a player leaving; the remove entities packet
// takes the body with it.
type PlayerInfoRemoveClientboundPacket struct {
	Uuids []string
}

func (p *PlayerInfoRemoveClientboundPacket) String() string {
	return fmt.Sprintf("PlayerInfoRemoveClientboundPacket{Uuids:[%s]}", strings.Join(p.Uuids, " "))
}

func (p *PlayerInfoRemoveClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(int32(len(p.Uuids))); err != nil {
		return err
	}

	for _, uuid := range p.Uuids {
		if err := ms.WriteUuid(uuid); err != nil {
			return err
		}
	}

	return nil
}
