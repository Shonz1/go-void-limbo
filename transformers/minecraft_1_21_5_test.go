package transformers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

func encodeChunk(t *testing.T, packet *play.LevelChunkWithLightClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

// The heightmaps sit between the chunk coordinates and everything else, so
// the downgraded body is the coordinates, the compound in place of the map,
// and the original's tail from the section buffer on, byte for byte.
func TestDowngradeLevelChunkWithLightTo1_21_4RewritesTheHeightmaps(t *testing.T) {
	packet := &play.LevelChunkWithLightClientboundPacket{
		X: 1,
		Z: -2,
		Heightmaps: []play.Heightmap{
			{Type: play.HeightmapMotionBlocking, Data: []int64{65, -1}},
			{Type: play.HeightmapWorldSurface, Data: []int64{66}},
		},
		SectionData:         []byte{0xAA, 0xBB, 0xCC},
		SkyLightMask:        []int64{0b100000},
		EmptySkyLightMask:   []int64{0b011111},
		EmptyBlockLightMask: []int64{0b111111},
		SkyLight:            [][]byte{{0xFF, 0xEE}},
	}

	body := encodeChunk(t, packet)
	got := runTransformer(t, DowngradeLevelChunkWithLightTo1_21_4, body)

	want := []byte{
		0x00, 0x00, 0x00, 0x01, // x
		0xff, 0xff, 0xff, 0xfe, // z
		0x0a, // an unnamed root compound
		// The entries in name order, each a long array with its name and int
		// length in front.
		0x0c, 0x00, 0x0f,
	}
	want = append(want, "MOTION_BLOCKING"...)
	want = append(want,
		0x00, 0x00, 0x00, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x41, // 65
		0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, // -1
		0x0c, 0x00, 0x0d,
	)
	want = append(want, "WORLD_SURFACE"...)
	want = append(want,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x42, // 66
		0x00, // the end of the compound
	)

	// Everything from the section buffer on is the original's.
	tail := bytes.Index(body, []byte{0x03, 0xaa, 0xbb, 0xcc})
	if tail < 0 {
		t.Fatal("the encoded body does not hold the section buffer where expected")
	}
	want = append(want, body[tail:]...)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.21.4 = % x\nwant        = % x", got, want)
	}
}

func TestDowngradeLevelChunkWithLightTo1_21_4RefusesAHeightmapItCannotName(t *testing.T) {
	packet := &play.LevelChunkWithLightClientboundPacket{
		Heightmaps: []play.Heightmap{{Type: 5, Data: []int64{1}}},
	}

	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(encodeChunk(t, packet)))
	out := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))

	err := DowngradeLevelChunkWithLightTo1_21_4(in, out)
	if err == nil || !strings.Contains(err.Error(), "heightmap kind 5") {
		t.Errorf("error = %v, want a refusal naming heightmap kind 5", err)
	}
}
