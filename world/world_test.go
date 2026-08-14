package world

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"fmt"
	"math/bits"
	"os"
	"path/filepath"
	"testing"

	"github.com/Shonz1/go-void-limbo/gamedata"
	"github.com/Shonz1/go-void-limbo/nbt"
	clientboundPlay "github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// The tests here build a world on disk and read the packets back the way the
// client reads them, with a decoder of their own, so an encoding mistake has
// to be made identically twice to slip through.

func TestLoad(t *testing.T) {
	dir := t.TempDir()

	// A spawn inside chunk 0,0.
	writeLevelDat(t, dir, nbt.Compound{
		"Data": nbt.Compound{
			"spawn": nbt.Compound{"pos": nbt.IntArray{8, 65, 8}},
		},
	})

	// One chunk: a single grass block in an otherwise empty section at y 0,
	// with sky light stored on that section and a heightmap on the chunk.
	data := make(nbt.LongArray, 256)
	data[0] = 0x1

	skyLight := make(nbt.ByteArray, 2048)
	for i := range skyLight {
		skyLight[i] = 0xFF
	}

	heightmap := make(nbt.LongArray, 37)
	heightmap[0] = 65

	writeRegion(t, dir, 0, 0, map[[2]int32]nbt.Compound{
		{0, 0}: {
			"xPos":       nbt.Int(0),
			"zPos":       nbt.Int(0),
			"yPos":       nbt.Int(-4),
			"Status":     nbt.String("minecraft:full"),
			"Heightmaps": nbt.Compound{"MOTION_BLOCKING": heightmap},
			"sections": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
				nbt.Compound{
					"Y": nbt.Byte(0),
					"block_states": nbt.Compound{
						"palette": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
							nbt.Compound{"Name": nbt.String("minecraft:air")},
							nbt.Compound{
								"Name":       nbt.String("minecraft:grass_block"),
								"Properties": nbt.Compound{"snowy": nbt.String("false")},
							},
						}},
						"data": data,
					},
					"SkyLight": skyLight,
				},
			}},
		},
	})

	world, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if x, y, z := world.Spawn(); x != 8.5 || y != 65 || z != 8.5 {
		t.Errorf("Spawn() = %v,%v,%v, want 8.5,65,8.5", x, y, z)
	}

	for _, version := range types.SupportedProtocolVersions {
		t.Run(fmt.Sprintf("protocol %d", version.ID), func(t *testing.T) {
			packets := world.PacketsFor(version)
			if len(packets) != 2 {
				t.Fatalf("PacketsFor() returned %d packets, want centre and one chunk", len(packets))
			}

			center, ok := packets[0].(*clientboundPlay.SetChunkCacheCenterClientboundPacket)
			if !ok || center.X != 0 || center.Z != 0 {
				t.Fatalf("packets[0] = %v, want the cache centred on 0,0", packets[0])
			}

			chunk, ok := packets[1].(*clientboundPlay.LevelChunkWithLightClientboundPacket)
			if !ok || chunk.X != 0 || chunk.Z != 0 {
				t.Fatalf("packets[1] = %v, want chunk 0,0", packets[1])
			}

			if len(chunk.Heightmaps) != 1 || chunk.Heightmaps[0].Type != clientboundPlay.HeightmapMotionBlocking || chunk.Heightmaps[0].Data[0] != 65 {
				t.Errorf("Heightmaps = %v, want the stored motion blocking map", chunk.Heightmaps)
			}

			blockStates, err := gamedata.BlockStatesFor(version)
			if err != nil {
				t.Fatalf("BlockStatesFor() error: %v", err)
			}

			grass, ok := blockStates.Id("minecraft:grass_block", map[string]string{"snowy": "false"})
			if !ok {
				t.Fatal("this version cannot name a grass block")
			}

			sections := decodeSections(t, version, chunk.SectionData, blockStates.StateCount())

			for i, section := range sections {
				wantCount := int32(0)
				wantBlock := int32(0)
				if i == 4 { // section y 0 in a world from section -4
					wantCount = 1
					wantBlock = grass
				}

				if section.blockCount != wantCount {
					t.Errorf("section %d holds %d blocks, want %d", i, section.blockCount, wantCount)
				}

				if section.fluidCount != 0 {
					t.Errorf("section %d holds %d fluids, want none", i, section.fluidCount)
				}

				if got := section.blocks[0]; got != wantBlock {
					t.Errorf("section %d block 0 = state %d, want %d", i, got, wantBlock)
				}

				for _, id := range section.blocks[1:] {
					if id != 0 {
						t.Fatalf("section %d holds state %d where only air was stored", i, id)
					}
				}
			}

			// The stored sky light is on section y 0: light bit 5, counting
			// from one section below the world. Everything else was never
			// stored, which the packet declares empty.
			if chunk.SkyLightMask[0] != 1<<5 || len(chunk.SkyLight) != 1 || !bytes.Equal(chunk.SkyLight[0], skyLight) {
				t.Errorf("sky light mask %b with %d arrays, want bit 5 and the stored array", chunk.SkyLightMask[0], len(chunk.SkyLight))
			}

			if chunk.EmptySkyLightMask[0] != (1<<lightSectionCount-1)&^(1<<5) {
				t.Errorf("empty sky light mask = %b, want every section but bit 5", chunk.EmptySkyLightMask[0])
			}

			if chunk.BlockLightMask[0] != 0 || chunk.EmptyBlockLightMask[0] != 1<<lightSectionCount-1 {
				t.Errorf("block light masks = %b and %b, want none stored and all empty", chunk.BlockLightMask[0], chunk.EmptyBlockLightMask[0])
			}
		})
	}
}

// TestLoadRepacksLargePalettes drives a section past the 256 palette entries
// the wire can name indirectly, which is the point where the stored indices
// have to be repacked into ids.
func TestLoadRepacksLargePalettes(t *testing.T) {
	dir := t.TempDir()

	writeLevelDat(t, dir, nbt.Compound{
		"Data": nbt.Compound{
			"spawn": nbt.Compound{"pos": nbt.IntArray{8, 65, 8}},
		},
	})

	// Redstone wire has thousands of states, so its power and connections make
	// as many distinct palette entries as the test needs.
	var palette []nbt.Tag
	var stored []map[string]string
	for _, east := range []string{"up", "side", "none"} {
		for _, north := range []string{"up", "side", "none"} {
			for _, south := range []string{"up", "side", "none"} {
				for power := 0; power < 16 && len(palette) < 300; power++ {
					properties := map[string]string{
						"east": east, "north": north, "south": south, "west": "none",
						"power": fmt.Sprint(power),
					}

					compound := nbt.Compound{}
					for name, value := range properties {
						compound[name] = nbt.String(value)
					}

					palette = append(palette, nbt.Compound{"Name": nbt.String("minecraft:redstone_wire"), "Properties": compound})
					stored = append(stored, properties)
				}
			}
		}
	}

	if len(palette) != 300 {
		t.Fatalf("built %d palette entries, want 300", len(palette))
	}

	// Every block names palette entry i%300, packed at the 9 bits 300 entries
	// store at.
	storedBits := bits.Len(uint(len(palette) - 1))
	perLong := 64 / storedBits
	data := make(nbt.LongArray, (4096+perLong-1)/perLong)
	for i := 0; i < 4096; i++ {
		data[i/perLong] |= int64(i%300) << ((i % perLong) * storedBits)
	}

	writeRegion(t, dir, 0, 0, map[[2]int32]nbt.Compound{
		{0, 0}: {
			"xPos":   nbt.Int(0),
			"zPos":   nbt.Int(0),
			"yPos":   nbt.Int(-4),
			"Status": nbt.String("minecraft:full"),
			"sections": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
				nbt.Compound{
					"Y": nbt.Byte(0),
					"block_states": nbt.Compound{
						"palette": nbt.List{ElementType: nbt.TagCompound, Elements: palette},
						"data":    data,
					},
				},
			}},
		},
	})

	world, err := Load(dir)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	for _, version := range types.SupportedProtocolVersions {
		t.Run(fmt.Sprintf("protocol %d", version.ID), func(t *testing.T) {
			blockStates, err := gamedata.BlockStatesFor(version)
			if err != nil {
				t.Fatalf("BlockStatesFor() error: %v", err)
			}

			want := make([]int32, len(stored))
			for i, properties := range stored {
				id, ok := blockStates.Id("minecraft:redstone_wire", properties)
				if !ok {
					t.Fatalf("this version cannot name redstone wire %v", properties)
				}

				want[i] = id
			}

			chunk := world.PacketsFor(version)[1].(*clientboundPlay.LevelChunkWithLightClientboundPacket)
			sections := decodeSections(t, version, chunk.SectionData, blockStates.StateCount())

			section := sections[4]
			if section.blockCount != 4096 {
				t.Errorf("section holds %d blocks, want all 4096", section.blockCount)
			}

			for i, id := range section.blocks {
				if id != want[i%300] {
					t.Fatalf("block %d = state %d, want %d", i, id, want[i%300])
				}
			}
		})
	}
}

// decodedSection is one section as the test's own decoder reads it back.
type decodedSection struct {
	blockCount, fluidCount int32
	blocks                 []int32
}

// decodeSections reads a section buffer by the client's rules: a count, a
// fluid count on the versions whose sections carry one, a block container
// whose declared bits pick the palette form, and a biome container.
func decodeSections(t *testing.T, version types.ProtocolVersion, data []byte, stateCount int32) []decodedSection {
	t.Helper()

	fluidCounts := version.ID >= types.ProtocolVersions.MINECRAFT_26_1.ID

	r := &sectionReader{t: t, data: data}

	var sections []decodedSection
	for i := 0; i < sectionCount; i++ {
		section := decodedSection{blockCount: r.short()}
		if fluidCounts {
			section.fluidCount = r.short()
		}
		section.blocks = r.container(4096, stateCount)

		// The biome container, which this server always writes as a single
		// value naming the one biome it registers.
		if biomes := r.container(64, 1); biomes[0] != 0 {
			t.Fatalf("section %d biome = %d, want 0", i, biomes[0])
		}

		sections = append(sections, section)
	}

	if r.pos != len(r.data) {
		t.Fatalf("section buffer holds %d bytes past its sections", len(r.data)-r.pos)
	}

	return sections
}

type sectionReader struct {
	t    *testing.T
	data []byte
	pos  int
}

func (r *sectionReader) short() int32 {
	value := int32(r.data[r.pos])<<8 | int32(r.data[r.pos+1])
	r.pos += 2
	return value
}

func (r *sectionReader) varInt() int32 {
	value, size, err := streams.ReadVarIntFrom(r.data[r.pos:])
	if err != nil {
		r.t.Fatalf("reading var int: %v", err)
	}

	r.pos += size
	return value
}

// container reads one paletted container of entries values, normalizing the
// declared bits the way the client does: one to four bits read as a four bit
// linear palette, five to eight as a map palette, and anything above as ids
// packed directly at however many bits the registry's size needs.
func (r *sectionReader) container(entries int, registrySize int32) []int32 {
	declared := int(r.data[r.pos])
	r.pos++

	if declared == 0 {
		value := r.varInt()
		values := make([]int32, entries)
		for i := range values {
			values[i] = value
		}

		return values
	}

	bitsPerEntry := declared
	var palette []int32

	maxIndirect := 8
	if entries == 64 { // biomes read indirect palettes up to three bits
		maxIndirect = 3
	}

	if declared <= maxIndirect {
		if entries == 4096 && declared <= 4 {
			bitsPerEntry = 4
		}

		palette = make([]int32, r.varInt())
		for i := range palette {
			palette[i] = r.varInt()
		}
	} else {
		bitsPerEntry = bits.Len(uint(registrySize - 1))
	}

	perLong := 64 / bitsPerEntry
	longs := (entries + perLong - 1) / perLong
	mask := int64(1)<<bitsPerEntry - 1

	values := make([]int32, entries)
	for i := range values {
		long := binary.BigEndian.Uint64(r.data[r.pos+(i/perLong)*8:])
		value := int32((int64(long) >> ((i % perLong) * bitsPerEntry)) & mask)

		if palette != nil {
			if value >= int32(len(palette)) {
				r.t.Fatalf("entry %d names palette index %d of %d", i, value, len(palette))
			}

			value = palette[value]
		}

		values[i] = value
	}

	r.pos += longs * 8

	return values
}

func writeLevelDat(t *testing.T, dir string, root nbt.Compound) {
	t.Helper()

	compressed := new(bytes.Buffer)
	writer := gzip.NewWriter(compressed)
	if _, err := writer.Write(encodeNamedNBT(t, root)); err != nil {
		t.Fatalf("compressing level.dat: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("compressing level.dat: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "level.dat"), compressed.Bytes(), 0o644); err != nil {
		t.Fatalf("writing level.dat: %v", err)
	}
}

func writeRegion(t *testing.T, dir string, regionX, regionZ int32, chunks map[[2]int32]nbt.Compound) {
	t.Helper()

	regionDir := filepath.Join(dir, "region")
	if err := os.MkdirAll(regionDir, 0o755); err != nil {
		t.Fatalf("creating region directory: %v", err)
	}

	header := make([]byte, 8192)
	var body []byte

	sector := int64(2)
	for pos, root := range chunks {
		compressed := new(bytes.Buffer)
		writer := zlib.NewWriter(compressed)
		if _, err := writer.Write(encodeNamedNBT(t, root)); err != nil {
			t.Fatalf("compressing chunk: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("compressing chunk: %v", err)
		}

		chunk := make([]byte, 5+compressed.Len())
		binary.BigEndian.PutUint32(chunk, uint32(compressed.Len()+1))
		chunk[4] = 2 // zlib
		copy(chunk[5:], compressed.Bytes())

		sectors := (len(chunk) + 4095) / 4096
		padded := make([]byte, sectors*4096)
		copy(padded, chunk)
		body = append(body, padded...)

		slot := (pos[0]&31 + (pos[1]&31)*32) * 4
		header[slot] = byte(sector >> 16)
		header[slot+1] = byte(sector >> 8)
		header[slot+2] = byte(sector)
		header[slot+3] = byte(sectors)

		sector += int64(sectors)
	}

	name := filepath.Join(regionDir, fmt.Sprintf("r.%d.%d.mca", regionX, regionZ))
	if err := os.WriteFile(name, append(header, body...), 0o644); err != nil {
		t.Fatalf("writing region file: %v", err)
	}
}

func encodeNamedNBT(t *testing.T, tag nbt.Tag) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	ms := streams.NewMinecraftStreamFromBuffer(buf)

	if err := nbt.WriteNamed(ms, "", tag); err != nil {
		t.Fatalf("WriteNamed() error: %v", err)
	}

	if err := ms.Flush(); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	return buf.Bytes()
}
