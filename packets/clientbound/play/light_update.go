package play

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

// LightUpdateClientboundPacket carries one chunk's light on its own, which is
// how 1.17.1 reads it: 1.18 is where the light joined the chunk packet, and
// before it a vanilla server sent a chunk as two packets, this one first and
// the chunk itself after it. Package world builds one beside every chunk for
// a version that reads it, from the same light the chunk packet carries, and
// no version from 1.18 on has an id for it: the chunk packet says everything
// this does.
//
// The layout is the chunk packet's light data behind the chunk coordinates
// and the trust edges flag, which a vanilla server of that version sets for
// every chunk it sends: the light at the chunk's edges is settled, and the
// client need not recompute it against the neighbours.
type LightUpdateClientboundPacket struct {
	// X and Z are in chunks.
	X, Z int32

	LightData
}

func (p *LightUpdateClientboundPacket) String() string {
	return fmt.Sprintf("LightUpdateClientboundPacket{X:%d Z:%d Sky:%d Block:%d}", p.X, p.Z, len(p.SkyLight), len(p.BlockLight))
}

func (p *LightUpdateClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.X); err != nil {
		return err
	}

	if err := ms.WriteVarInt(p.Z); err != nil {
		return err
	}

	if err := ms.WriteBoolean(lightTrustsEdges); err != nil {
		return err
	}

	return p.LightData.encode(ms)
}

// lightTrustsEdges is what a vanilla server before 1.20 says about the light
// of every chunk it sends. 1.20 is where the flag left the wire.
const lightTrustsEdges = true
