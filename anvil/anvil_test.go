package anvil

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// encodeNamedNBT renders a tag the way files store one: a named root.
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

// writeRegionFile writes a region file holding the given chunks, each already
// encoded as NBT, at the slots their coordinates pick.
func writeRegionFile(t *testing.T, dir string, regionX, regionZ int32, chunks map[[2]int32][]byte) {
	t.Helper()

	header := make([]byte, 2*sectorSize)
	var body []byte

	sector := int64(2)
	for pos, payload := range chunks {
		compressed := new(bytes.Buffer)
		writer := zlib.NewWriter(compressed)
		if _, err := writer.Write(payload); err != nil {
			t.Fatalf("compressing chunk: %v", err)
		}
		if err := writer.Close(); err != nil {
			t.Fatalf("compressing chunk: %v", err)
		}

		chunk := make([]byte, 5+compressed.Len())
		binary.BigEndian.PutUint32(chunk, uint32(compressed.Len()+1))
		chunk[4] = compressionZlib
		copy(chunk[5:], compressed.Bytes())

		sectors := (len(chunk) + sectorSize - 1) / sectorSize
		padded := make([]byte, sectors*sectorSize)
		copy(padded, chunk)
		body = append(body, padded...)

		slot := (pos[0]&31 + (pos[1]&31)*32) * 4
		header[slot] = byte(sector >> 16)
		header[slot+1] = byte(sector >> 8)
		header[slot+2] = byte(sector)
		header[slot+3] = byte(sectors)

		sector += int64(sectors)
	}

	name := filepath.Join(dir, "r."+itoa(regionX)+"."+itoa(regionZ)+".mca")
	if err := os.WriteFile(name, append(header, body...), 0o644); err != nil {
		t.Fatalf("writing region file: %v", err)
	}
}

func itoa(v int32) string {
	if v < 0 {
		return "-" + itoa(-v)
	}

	if v < 10 {
		return string(rune('0' + v))
	}

	return itoa(v/10) + string(rune('0'+v%10))
}

// testChunkNBT is a small but complete chunk: one section of blocks with a
// two-entry palette, light on the section, and a heightmap.
func testChunkNBT(t *testing.T, x, z int32) []byte {
	t.Helper()

	// 4096 indices at four bits, the first block stone and the rest air.
	data := make(nbt.LongArray, 256)
	data[0] = 0x1

	skyLight := make(nbt.ByteArray, 2048)
	for i := range skyLight {
		skyLight[i] = 0xFF
	}

	heightmap := make(nbt.LongArray, 37)
	heightmap[0] = 65

	return encodeNamedNBT(t, nbt.Compound{
		"xPos":   nbt.Int(x),
		"zPos":   nbt.Int(z),
		"yPos":   nbt.Int(-4),
		"Status": nbt.String("minecraft:full"),
		"Heightmaps": nbt.Compound{
			"MOTION_BLOCKING": heightmap,
		},
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
	})
}

func TestReadChunk(t *testing.T) {
	dir := t.TempDir()
	writeRegionFile(t, dir, 0, 0, map[[2]int32][]byte{{5, 3}: testChunkNBT(t, 5, 3)})

	chunk, err := ReadChunk(dir, 5, 3)
	if err != nil {
		t.Fatalf("ReadChunk() error: %v", err)
	}

	if chunk.X != 5 || chunk.Z != 3 || chunk.MinSectionY != -4 {
		t.Errorf("ReadChunk() = chunk %d,%d from section %d, want 5,3 from -4", chunk.X, chunk.Z, chunk.MinSectionY)
	}

	if chunk.Status != "minecraft:full" {
		t.Errorf("Status = %q, want minecraft:full", chunk.Status)
	}

	if len(chunk.Sections) != 1 {
		t.Fatalf("got %d sections, want 1", len(chunk.Sections))
	}

	section := chunk.Sections[0]
	if section.Y != 0 {
		t.Errorf("section Y = %d, want 0", section.Y)
	}

	if len(section.BlockPalette) != 2 {
		t.Fatalf("palette holds %d entries, want 2", len(section.BlockPalette))
	}

	if got := section.BlockPalette[0]; got.Name != "minecraft:air" || got.Properties != nil {
		t.Errorf("palette[0] = %+v, want plain minecraft:air", got)
	}

	if got := section.BlockPalette[1]; got.Name != "minecraft:grass_block" || got.Properties["snowy"] != "false" {
		t.Errorf("palette[1] = %+v, want minecraft:grass_block snowy=false", got)
	}

	if len(section.BlockData) != 256 || section.BlockData[0] != 1 {
		t.Errorf("data holds %d longs starting %d, want 256 starting 1", len(section.BlockData), section.BlockData[0])
	}

	if len(section.SkyLight) != 2048 || section.SkyLight[0] != 0xFF {
		t.Errorf("sky light holds %d bytes starting %#x, want 2048 starting 0xff", len(section.SkyLight), section.SkyLight[0])
	}

	if section.BlockLight != nil {
		t.Errorf("block light = %d bytes, want none stored", len(section.BlockLight))
	}

	heightmap, ok := chunk.Heightmaps["MOTION_BLOCKING"]
	if !ok || len(heightmap) != 37 || heightmap[0] != 65 {
		t.Errorf("MOTION_BLOCKING = %d longs starting %v, %t; want 37 starting 65", len(heightmap), heightmap, ok)
	}
}

func TestReadChunkNotFound(t *testing.T) {
	dir := t.TempDir()
	writeRegionFile(t, dir, 0, 0, map[[2]int32][]byte{{5, 3}: testChunkNBT(t, 5, 3)})

	// An empty slot in a region file that exists.
	if _, err := ReadChunk(dir, 1, 1); !errors.Is(err, ErrChunkNotFound) {
		t.Errorf("ReadChunk(1,1) error = %v, want ErrChunkNotFound", err)
	}

	// A region file that does not.
	if _, err := ReadChunk(dir, -1, -1); !errors.Is(err, ErrChunkNotFound) {
		t.Errorf("ReadChunk(-1,-1) error = %v, want ErrChunkNotFound", err)
	}
}

func TestReadChunkWrongCoordinates(t *testing.T) {
	dir := t.TempDir()

	// A chunk stored in the slot for 5,3 that says it is 6,3.
	writeRegionFile(t, dir, 0, 0, map[[2]int32][]byte{{5, 3}: testChunkNBT(t, 6, 3)})

	if _, err := ReadChunk(dir, 5, 3); err == nil {
		t.Error("ReadChunk() accepted a chunk stored in the wrong slot")
	}
}

func TestReadSpawn(t *testing.T) {
	dir := t.TempDir()

	payload := encodeNamedNBT(t, nbt.Compound{
		"Data": nbt.Compound{
			"spawn": nbt.Compound{
				"pos":       nbt.IntArray{1, 65, -3},
				"dimension": nbt.String("minecraft:overworld"),
			},
		},
	})

	writeLevelDat(t, dir, payload)

	spawn, err := ReadSpawn(dir)
	if err != nil {
		t.Fatalf("ReadSpawn() error: %v", err)
	}

	if spawn != (Spawn{X: 1, Y: 65, Z: -3}) {
		t.Errorf("ReadSpawn() = %+v, want {1 65 -3}", spawn)
	}
}

func TestReadSpawnOldForm(t *testing.T) {
	dir := t.TempDir()

	payload := encodeNamedNBT(t, nbt.Compound{
		"Data": nbt.Compound{
			"SpawnX": nbt.Int(7),
			"SpawnY": nbt.Int(70),
			"SpawnZ": nbt.Int(-7),
		},
	})

	writeLevelDat(t, dir, payload)

	spawn, err := ReadSpawn(dir)
	if err != nil {
		t.Fatalf("ReadSpawn() error: %v", err)
	}

	if spawn != (Spawn{X: 7, Y: 70, Z: -7}) {
		t.Errorf("ReadSpawn() = %+v, want {7 70 -7}", spawn)
	}
}

func TestReadSpawnMissing(t *testing.T) {
	dir := t.TempDir()
	writeLevelDat(t, dir, encodeNamedNBT(t, nbt.Compound{"Data": nbt.Compound{}}))

	if _, err := ReadSpawn(dir); err == nil {
		t.Error("ReadSpawn() invented a spawn for a level.dat holding none")
	}
}

func writeLevelDat(t *testing.T, dir string, payload []byte) {
	t.Helper()

	compressed := new(bytes.Buffer)
	writer := gzip.NewWriter(compressed)
	if _, err := writer.Write(payload); err != nil {
		t.Fatalf("compressing level.dat: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("compressing level.dat: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "level.dat"), compressed.Bytes(), 0o644); err != nil {
		t.Fatalf("writing level.dat: %v", err)
	}
}
