package play

import (
	"bytes"
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

// LevelChunkWithSectionLightClientboundPacket is the chunk packet of a
// version that reads each section's light inside the chunk, behind the
// section's blocks: 1.14 is where the light left the chunk for a packet of
// its own, and 1.13.2 below it reads the two as one. Package world builds it
// in place of the chunk and the light update for such a version, from the
// same chunk.
//
// No version this server speaks from 1.14 on has an id for it, and the
// versions in between carry its halves apart: a step that changes the chunk
// packet or the light update changes the half that is that packet, the same
// way, so the two arrive at the 1.14 step as 1.14 would send them. That step
// is where they are made one. The body is the chunk packet's, behind a var
// int of how long it is, and then the light update's.
type LevelChunkWithSectionLightClientboundPacket struct {
	Chunk *LevelChunkWithLightClientboundPacket
}

func (p *LevelChunkWithSectionLightClientboundPacket) String() string {
	return fmt.Sprintf("LevelChunkWithSectionLightClientboundPacket{X:%d Z:%d Sections:%dB Sky:%d Block:%d}",
		p.Chunk.X, p.Chunk.Z, len(p.Chunk.SectionData), len(p.Chunk.SkyLight), len(p.Chunk.BlockLight))
}

func (p *LevelChunkWithSectionLightClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	chunk := new(bytes.Buffer)
	chunkStream := streams.NewMinecraftStreamFromBuffer(chunk)

	if err := p.Chunk.Encode(chunkStream); err != nil {
		return err
	}

	if err := chunkStream.Flush(); err != nil {
		return err
	}

	if err := ms.WriteByteArray(chunk.Bytes()); err != nil {
		return err
	}

	light := LightUpdateClientboundPacket{X: p.Chunk.X, Z: p.Chunk.Z, LightData: p.Chunk.LightData}

	return light.Encode(ms)
}
