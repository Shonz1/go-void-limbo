package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
)

// chunk1_13 lays a chunk packet out as 1.13 reads it, around the section
// buffer given, with no block entities.
func chunk1_13(t *testing.T, whole bool, mask int32, sections []byte) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteInt(3),
			ms.WriteInt(-4),
			ms.WriteBoolean(whole),
			ms.WriteVarInt(mask),
			ms.WriteByteArray(sections),
			ms.WriteVarInt(0),
		)
	})
}

// section1_13 lays one section out as 1.13 reads it: the bits, the palette
// when there is one, the counted longs and the two light arrays. A nil
// palette is a container that names its ids directly and has none, and
// emptyPalette writes the palette of none 1.12.2 reads in its place.
func section1_13(bits byte, palette []int32, longs []int64, emptyPalette bool) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(bits)

	if palette != nil {
		buf.Write(streams.AppendVarInt(nil, int32(len(palette))))
		for _, id := range palette {
			buf.Write(streams.AppendVarInt(nil, id))
		}
	} else if emptyPalette {
		buf.WriteByte(0x00)
	}

	buf.Write(streams.AppendVarInt(nil, int32(len(longs))))
	for _, long := range longs {
		buf.Write(appendLong(nil, long))
	}

	buf.Write(bytes.Repeat([]byte{0x0F}, lightArrayBytes))
	buf.Write(bytes.Repeat([]byte{0xF0}, lightArrayBytes))

	return buf.Bytes()
}

func appendLong(data []byte, value int64) []byte {
	for shift := 56; shift >= 0; shift -= 8 {
		data = append(data, byte(value>>shift))
	}

	return data
}

// 1.12.2 reads a section with a palette as 1.13 does, and a section whose
// ids go on the wire directly with a palette of none in front of them; the
// biomes of a whole chunk are a byte each.
func TestDowngradeLevelChunkWithSectionLightTo1_12_2(t *testing.T) {
	paletted := make([]int64, 256)
	paletted[0] = 0x1

	direct := make([]int64, 832)
	direct[0] = 2 << 4

	biomes := new(bytes.Buffer)
	for range biomes1_12_2 {
		biomes.Write([]byte{0x00, 0x00, 0x00, 0x01})
	}

	sent := chunk1_13(t, true, 0b101, append(append(
		section1_13(4, []int32{0, 2 << 4}, paletted, false),
		section1_13(13, nil, direct, false)...),
		biomes.Bytes()...))

	want := chunk1_13(t, true, 0b101, append(append(
		section1_13(4, []int32{0, 2 << 4}, paletted, false),
		section1_13(13, nil, direct, true)...),
		bytes.Repeat([]byte{0x01}, biomes1_12_2)...))

	if got := runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_12_2, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.12.2 = %d bytes, want %d", len(got), len(want))
	}

	// A chunk that is not whole carries no biomes on either side.
	part := chunk1_13(t, false, 0b1, section1_13(4, []int32{0}, paletted, false))

	if got := runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_12_2, part); !bytes.Equal(got, part) {
		t.Errorf("a part of a chunk to 1.12.2 = % x, want it as it stands", got)
	}
}

// What 1.12.2 cannot hold is refused rather than cut down: a biome past the
// byte it is read as, and a section above the sixteen its chunk holds.
func TestDowngradeLevelChunkWithSectionLightTo1_12_2Refuses(t *testing.T) {
	biomes := new(bytes.Buffer)
	for range biomes1_12_2 {
		biomes.Write([]byte{0x00, 0x00, 0x01, 0x00})
	}

	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_12_2, chunk1_13(t, true, 0, biomes.Bytes())); err == nil {
		t.Error("expected biome 256 to be refused")
	}

	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_12_2, chunk1_13(t, false, 1<<16, nil)); err == nil {
		t.Error("expected a seventeenth section to be refused")
	}

	short := chunk1_13(t, false, 0b1, section1_13(4, []int32{0}, make([]int64, 256), false)[:100])
	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_12_2, short); err == nil {
		t.Error("expected a section cut short to be refused")
	}
}
