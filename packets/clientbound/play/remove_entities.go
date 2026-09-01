package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// RemoveEntitiesClientboundPacket takes entities out of the client's world. It
// is the body half of a player leaving; the player info remove packet takes
// the list entry with it.
type RemoveEntitiesClientboundPacket struct {
	EntityIds []int32
}

func (p *RemoveEntitiesClientboundPacket) String() string {
	return fmt.Sprintf("RemoveEntitiesClientboundPacket{EntityIds:%v}", p.EntityIds)
}

func (p *RemoveEntitiesClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(int32(len(p.EntityIds))); err != nil {
		return err
	}

	for _, id := range p.EntityIds {
		if err := ms.WriteVarInt(id); err != nil {
			return err
		}
	}

	return nil
}
