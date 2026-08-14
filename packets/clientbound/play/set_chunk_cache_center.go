package play

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

// SetChunkCacheCenterClientboundPacket tells the client which chunk its cache
// is centred on. The client keeps chunks in a square window around this centre
// and quietly drops anything sent outside it, so it is said before the chunks
// are, even for a player standing at the default centre of 0,0: a packet is
// how the window is known rather than guessed at.
type SetChunkCacheCenterClientboundPacket struct {
	// X and Z are in chunks.
	X, Z int32
}

func (p *SetChunkCacheCenterClientboundPacket) String() string {
	return fmt.Sprintf("SetChunkCacheCenterClientboundPacket{X:%d Z:%d}", p.X, p.Z)
}

func (p *SetChunkCacheCenterClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.X); err != nil {
		return err
	}

	return ms.WriteVarInt(p.Z)
}
