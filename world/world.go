// Package world turns a saved world into the packets that show it: the chunks
// around its spawn, prebuilt for every protocol version this server speaks.
//
// Everything here happens once, at startup. A saved world names its blocks and
// the network numbers them, differently per version, so the work of a chunk
// packet is translation -- and with the world read-only and the same for every
// client, there is no reason to translate later than load or more than once
// per version.
package world

import (
	"errors"
	"fmt"
	"log/slog"
	"math/bits"
	"path/filepath"

	"github.com/Shonz1/go-void-limbo/anvil"
	"github.com/Shonz1/go-void-limbo/gamedata"
	clientboundPlay "github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// The world's vertical bounds, in sections. These are the bounds the
// dimension type in package gamedata announces (min_y -64, height 384), and
// they have to be: the client sizes a chunk by the dimension it was told it is
// in, and a chunk of any other size is a chunk it refuses.
const (
	minSectionY  = -4
	sectionCount = 24
)

// Light spans one section beyond the blocks in both directions.
const lightSectionCount = sectionCount + 2

// chunkRadius is how far out from the spawn chunk the world is sent, in
// chunks. One more than the view distance the join announces (see package
// handlers), because the client only renders a chunk whose neighbours it also
// holds: the last ring is sent to be leaned on, not seen.
const chunkRadius = 9

// blocksPerSection is the entries a section's block container packs, 16x16x16.
const blocksPerSection = 4096

const fullChunkStatus = "minecraft:full"

// airBlock is what a section that was never stored is made of, and what an
// unresolvable block state falls back to: a hole shaped like the world's own
// nothing.
const airBlock = "minecraft:air"

// airBlocks are the states that count as no block at all when a section's
// blocks are counted. The count is what tells the client a section is worth
// rendering, so air of any flavour stays out of it.
var airBlocks = map[string]bool{
	"minecraft:air":      true,
	"minecraft:cave_air": true,
	"minecraft:void_air": true,
}

// fluidBlocks are the states that hold fluid by what they are rather than by a
// waterlogged property: the fluids themselves, and the plants that only exist
// underwater. The section's fluid count wants every block whose fluid is not
// empty, and for everything else the waterlogged property answers that.
var fluidBlocks = map[string]bool{
	"minecraft:water":         true,
	"minecraft:lava":          true,
	"minecraft:bubble_column": true,
	"minecraft:kelp":          true,
	"minecraft:kelp_plant":    true,
	"minecraft:seagrass":      true,
	"minecraft:tall_seagrass": true,
}

// heightmaps is which stored heightmaps a chunk packet carries and the number
// the network knows each by. The client keeps exactly these two and computes
// the rest; a vanilla server sends no others.
var heightmaps = map[string]int32{
	"WORLD_SURFACE":   clientboundPlay.HeightmapWorldSurface,
	"MOTION_BLOCKING": clientboundPlay.HeightmapMotionBlocking,
}

// World is one loaded world: where its spawn is, and the packets that put its
// chunks on a wire, per protocol version.
type World struct {
	spawn anvil.Spawn

	// packets starts with the chunk cache centre and carries the chunks after
	// it, for each version this server speaks. The slices are shared by every
	// client on the version and must not be modified.
	packets map[types.ProtocolId][]types.ClientboundPacket
}

// Load reads the world saved in dir and prebuilds its packets for every
// supported protocol version.
//
// A chunk that was never generated, or that generation never finished, is
// skipped: the client shows void there, which is what this server used to show
// everywhere. A chunk that exists but cannot be translated is an error, since
// a world that loads with holes torn by mistranslation would look like damage
// and read like success.
func Load(dir string) (*World, error) {
	spawn, err := anvil.ReadSpawn(dir)
	if err != nil {
		return nil, err
	}

	centerX, centerZ := spawn.X>>4, spawn.Z>>4

	regionDir := filepath.Join(dir, "region")

	var chunks []*anvil.Chunk
	for z := centerZ - chunkRadius; z <= centerZ+chunkRadius; z++ {
		for x := centerX - chunkRadius; x <= centerX+chunkRadius; x++ {
			chunk, err := anvil.ReadChunk(regionDir, x, z)
			if errors.Is(err, anvil.ErrChunkNotFound) {
				continue
			}

			if err != nil {
				return nil, err
			}

			if chunk.Status != fullChunkStatus {
				continue
			}

			chunks = append(chunks, chunk)
		}
	}

	world := &World{spawn: spawn, packets: make(map[types.ProtocolId][]types.ClientboundPacket)}

	for _, version := range types.SupportedProtocolVersions {
		blockStates, err := gamedata.BlockStatesFor(version)
		if err != nil {
			return nil, err
		}

		builder := &chunkBuilder{
			blockStates: blockStates,
			version:     version,
			fluidCounts: version.ID >= types.ProtocolVersions.MINECRAFT_26_1.ID,
			dataLengths: version.ID < types.ProtocolVersions.MINECRAFT_1_21_5.ID,
		}

		packets := make([]types.ClientboundPacket, 0, len(chunks)+1)
		packets = append(packets, &clientboundPlay.SetChunkCacheCenterClientboundPacket{X: centerX, Z: centerZ})

		for _, chunk := range chunks {
			packet, err := builder.build(chunk)
			if err != nil {
				return nil, fmt.Errorf("world: chunk %d,%d for protocol %d: %w", chunk.X, chunk.Z, version.ID, err)
			}

			packets = append(packets, packet)
		}

		world.packets[version.ID] = packets
	}

	slog.Info("world loaded", "dir", dir, "chunks", len(chunks),
		"spawn", fmt.Sprintf("%d,%d,%d", spawn.X, spawn.Y, spawn.Z))

	return world, nil
}

// PacketsFor returns the packets that put this world on the wire of a client
// speaking version, the chunk cache centre first. The slice is shared across
// connections and must not be modified. It is empty for a version the world
// was not built for, which is no version a connection can reach the play
// phase on.
func (w *World) PacketsFor(version types.ProtocolVersion) []types.ClientboundPacket {
	return w.packets[version.ID]
}

// Spawn is where the world puts a joining player: the centre of the block the
// level.dat names.
func (w *World) Spawn() (x, y, z float64) {
	return float64(w.spawn.X) + 0.5, float64(w.spawn.Y), float64(w.spawn.Z) + 0.5
}

// chunkBuilder translates chunks for one protocol version.
type chunkBuilder struct {
	blockStates *gamedata.BlockStates
	version     types.ProtocolVersion

	// fluidCounts is whether sections carry a fluid count after the block
	// count, which 26.1 added: a 1.21.x section is the block count and the
	// two containers, and a 26.x section has the fluid count between them.
	fluidCounts bool

	// dataLengths is whether a paletted container's data comes with its
	// length in front, which 1.21.5 dropped: a container on 1.21.4 or before
	// is the bits, the palette, a var int count of longs and the longs, and a
	// later one leaves the count out because the bits already say how many
	// longs follow. A single value container has no longs, and before 1.21.5
	// says so with a count of zero.
	dataLengths bool

	// substituted is every stored state this version had no number for, warned
	// about once each rather than once per block.
	substituted map[string]bool
}

func (b *chunkBuilder) build(chunk *anvil.Chunk) (*clientboundPlay.LevelChunkWithLightClientboundPacket, error) {
	if chunk.MinSectionY != minSectionY {
		return nil, fmt.Errorf("world starts at section %d, want %d", chunk.MinSectionY, minSectionY)
	}

	packet := &clientboundPlay.LevelChunkWithLightClientboundPacket{X: chunk.X, Z: chunk.Z}

	for name, packed := range chunk.Heightmaps {
		if kind, wanted := heightmaps[name]; wanted {
			packet.Heightmaps = append(packet.Heightmaps, clientboundPlay.Heightmap{Type: kind, Data: packed})
		}
	}

	// The stored sections by height, blocks and light alike.
	stored := make(map[int32]*anvil.Section, len(chunk.Sections))
	for i := range chunk.Sections {
		stored[chunk.Sections[i].Y] = &chunk.Sections[i]
	}

	var sectionData []byte
	for y := int32(minSectionY); y < minSectionY+sectionCount; y++ {
		var err error

		if section := stored[y]; section != nil && section.BlockPalette != nil {
			sectionData, err = b.appendSection(sectionData, section)
		} else {
			sectionData, err = b.appendEmptySection(sectionData)
		}

		if err != nil {
			return nil, fmt.Errorf("section %d: %w", y, err)
		}
	}

	packet.SectionData = sectionData

	// The light masks index sections from one below the world's blocks. A
	// stored array travels; a section without one is declared empty, which for
	// a saved world it is: the game stores every array that holds any light.
	skyMask, blockMask := make([]int64, 1), make([]int64, 1)
	emptySkyMask, emptyBlockMask := make([]int64, 1), make([]int64, 1)

	for bit := 0; bit < lightSectionCount; bit++ {
		section := stored[minSectionY-1+int32(bit)]

		if section != nil && len(section.SkyLight) > 0 {
			skyMask[0] |= 1 << bit
			packet.SkyLight = append(packet.SkyLight, section.SkyLight)
		} else {
			emptySkyMask[0] |= 1 << bit
		}

		if section != nil && len(section.BlockLight) > 0 {
			blockMask[0] |= 1 << bit
			packet.BlockLight = append(packet.BlockLight, section.BlockLight)
		} else {
			emptyBlockMask[0] |= 1 << bit
		}
	}

	packet.SkyLightMask, packet.BlockLightMask = skyMask, blockMask
	packet.EmptySkyLightMask, packet.EmptyBlockLightMask = emptySkyMask, emptyBlockMask

	return packet, nil
}

// appendSection appends one stored section in wire form: block count, fluid
// count on the versions that carry one, the block state container, the biome
// container.
func (b *chunkBuilder) appendSection(data []byte, section *anvil.Section) ([]byte, error) {
	// The palette, translated. What this version cannot name becomes the
	// block's own default state when the block exists and air when it does
	// not, because a lobby missing one block reads better than a server that
	// will not start over it -- and the log says what was lost.
	ids := make([]int32, len(section.BlockPalette))
	air := make([]bool, len(section.BlockPalette))
	fluid := make([]bool, len(section.BlockPalette))

	for i, state := range section.BlockPalette {
		id, ok := b.blockStates.Id(state.Name, state.Properties)
		if !ok {
			id, ok = b.blockStates.DefaultId(state.Name)
			if !ok {
				id = 0
			}

			b.warnSubstituted(state)
		}

		ids[i] = id
		air[i] = airBlocks[state.Name]
		fluid[i] = fluidBlocks[state.Name] || state.Properties["waterlogged"] == "true"
	}

	if len(ids) == 1 {
		return b.appendSingleValueSection(data, ids[0], air[0], fluid[0]), nil
	}

	// The stored packing: at least four bits, indices never crossing a long.
	storedBits := max(4, bits.Len(uint(len(ids)-1)))

	entriesPerLong := 64 / storedBits
	if want := (blocksPerSection + entriesPerLong - 1) / entriesPerLong; len(section.BlockData) != want {
		return nil, fmt.Errorf("%d palette entries pack %d longs, stored %d", len(ids), want, len(section.BlockData))
	}

	// The counts the wire wants and the world does not store.
	blockCount, fluidCount := int32(0), int32(0)
	mask := int64(1<<storedBits) - 1

	for i := 0; i < blocksPerSection; i++ {
		index := (section.BlockData[i/entriesPerLong] >> ((i % entriesPerLong) * storedBits)) & mask
		if index >= int64(len(ids)) {
			return nil, fmt.Errorf("block %d names palette entry %d of %d", i, index, len(ids))
		}

		if !air[index] {
			blockCount++
		}

		if fluid[index] {
			fluidCount++
		}
	}

	data = appendShort(data, blockCount)
	if b.fluidCounts {
		data = appendShort(data, fluidCount)
	}

	// The block container. Up to 256 entries the stored packing is exactly the
	// wire packing -- the same bit widths over the same indices -- so the
	// palette is written translated and the data travels as stored. Past 256
	// the wire has no palette form and wants ids packed directly, which no
	// section of a plausible lobby reaches, but worlds are under no obligation
	// to be plausible.
	if len(ids) <= 256 {
		data = append(data, byte(storedBits))
		data = streams.AppendVarInt(data, int32(len(ids)))
		for _, id := range ids {
			data = streams.AppendVarInt(data, id)
		}

		data = b.appendData(data, section.BlockData)
	} else {
		data = b.appendGlobalBlockData(data, section.BlockData, ids, storedBits)
	}

	return b.appendBiomes(data), nil
}

// appendEmptySection appends the wire form of a section the world never
// stored: all air, no fluid.
func (b *chunkBuilder) appendEmptySection(data []byte) ([]byte, error) {
	id, ok := b.blockStates.Id(airBlock, nil)
	if !ok {
		return nil, fmt.Errorf("this version has no %s", airBlock)
	}

	return b.appendSingleValueSection(data, id, true, false), nil
}

// appendSingleValueSection appends a section that is 4096 of the same state:
// counts for all of it or none of it, and a container of one entry and no
// data.
func (b *chunkBuilder) appendSingleValueSection(data []byte, id int32, air, fluid bool) []byte {
	blockCount, fluidCount := int32(blocksPerSection), int32(0)
	if air {
		blockCount = 0
	}

	if fluid {
		fluidCount = blocksPerSection
	}

	data = appendShort(data, blockCount)
	if b.fluidCounts {
		data = appendShort(data, fluidCount)
	}

	data = append(data, 0)
	data = streams.AppendVarInt(data, id)

	return b.appendBiomes(b.appendData(data, nil))
}

// appendBiomes appends the section's biome container: every block the first
// and only biome the registry this server sends holds (see package gamedata).
// What the world stored cannot be said more faithfully than that, because a
// biome is whatever the client was told the numbers mean.
func (b *chunkBuilder) appendBiomes(data []byte) []byte {
	data = append(data, 0)
	data = streams.AppendVarInt(data, 0)

	return b.appendData(data, nil)
}

// appendData appends a container's packed longs, with the count in front on
// the versions that want it.
func (b *chunkBuilder) appendData(data []byte, longs []int64) []byte {
	if b.dataLengths {
		data = streams.AppendVarInt(data, int32(len(longs)))
	}

	return appendLongs(data, longs)
}

// appendGlobalBlockData repacks a section too varied for a palette: the wire
// form is the ids themselves, at however many bits the version's largest id
// needs.
func (b *chunkBuilder) appendGlobalBlockData(data []byte, packed []int64, ids []int32, storedBits int) []byte {
	wireBits := bits.Len(uint(b.blockStates.StateCount() - 1))
	data = append(data, byte(wireBits))

	storedPerLong := 64 / storedBits
	storedMask := int64(1<<storedBits) - 1

	wirePerLong := 64 / wireBits
	longs := make([]int64, (blocksPerSection+wirePerLong-1)/wirePerLong)

	for i := 0; i < blocksPerSection; i++ {
		index := (packed[i/storedPerLong] >> ((i % storedPerLong) * storedBits)) & storedMask
		longs[i/wirePerLong] |= int64(ids[index]) << ((i % wirePerLong) * wireBits)
	}

	return b.appendData(data, longs)
}

func (b *chunkBuilder) warnSubstituted(state anvil.BlockState) {
	key := fmt.Sprintf("%s%v", state.Name, state.Properties)
	if b.substituted[key] {
		return
	}

	if b.substituted == nil {
		b.substituted = map[string]bool{}
	}
	b.substituted[key] = true

	slog.Warn("the world stores a block state this version cannot name, substituting",
		"block", state.Name, "properties", fmt.Sprintf("%v", state.Properties), "protocol", b.version.ID)
}

func appendShort(data []byte, value int32) []byte {
	return append(data, byte(value>>8), byte(value))
}

func appendLongs(data []byte, values []int64) []byte {
	for _, value := range values {
		data = append(data,
			byte(value>>56), byte(value>>48), byte(value>>40), byte(value>>32),
			byte(value>>24), byte(value>>16), byte(value>>8), byte(value))
	}

	return data
}
