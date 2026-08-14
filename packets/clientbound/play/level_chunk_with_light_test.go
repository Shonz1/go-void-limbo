package play

import (
	"bytes"
	"testing"
)

func TestLevelChunkWithLightClientboundPacketEncode(t *testing.T) {
	p := &LevelChunkWithLightClientboundPacket{
		X: 1,
		Z: -2,
		Heightmaps: []Heightmap{
			{Type: HeightmapMotionBlocking, Data: []int64{65}},
		},
		SectionData:         []byte{0xAA, 0xBB, 0xCC},
		SkyLightMask:        []int64{0b100000},
		BlockLightMask:      nil,
		EmptySkyLightMask:   []int64{0b011111},
		EmptyBlockLightMask: []int64{0b111111},
		SkyLight:            [][]byte{{0xFF, 0xEE}},
	}

	want := []byte{
		0x00, 0x00, 0x00, 0x01, // x
		0xff, 0xff, 0xff, 0xfe, // z
		0x01,                                           // one heightmap
		0x04,                                           // of the motion blocking kind
		0x01,                                           // one long of it
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x41, // 65
		0x03, 0xaa, 0xbb, 0xcc, // the section buffer, length first
		0x00,                                                 // no block entities
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, // sky light mask
		0x00,                                                 // block light mask, empty
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x1f, // empty sky light mask
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x3f, // empty block light mask
		0x01,             // one sky light array
		0x02, 0xff, 0xee, // of two bytes (a real one holds 2048)
		0x00, // no block light arrays
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

func TestSetChunkCacheCenterClientboundPacketEncode(t *testing.T) {
	p := &SetChunkCacheCenterClientboundPacket{X: 5, Z: -1}

	// Two var ints, the negative one at its full five bytes.
	want := []byte{0x05, 0xff, 0xff, 0xff, 0xff, 0x0f}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}
