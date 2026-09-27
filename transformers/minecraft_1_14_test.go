package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// writeAll runs the writes of a body in order and stops at the first that
// fails.
func writeAll(steps ...error) error {
	for _, err := range steps {
		if err != nil {
			return err
		}
	}

	return nil
}

// The login 1.13.2 reads is the login 1.14 reads with the difficulty behind
// the dimension and without the view distance behind the level type.
func TestDowngradePlayLoginTo1_13_2PutsTheDifficultyInAndTheViewDistanceOut(t *testing.T) {
	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteInt(7),
			ms.WriteByte(2|8),
			ms.WriteInt(-1),
			ms.WriteByte(20),
			ms.WriteString("flat"),
			ms.WriteVarInt(8),
			ms.WriteBoolean(true),
		)
	})

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteInt(7),
			ms.WriteByte(2|8),
			ms.WriteInt(-1),
			ms.WriteByte(peaceful1_13_2),
			ms.WriteByte(20),
			ms.WriteString("flat"),
			ms.WriteBoolean(true),
		)
	})

	if got := runTransformer(t, DowngradePlayLoginTo1_13_2, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.13.2 = % x\nwant = % x", got, want)
	}

	if err := failingTransformer(t, DowngradePlayLoginTo1_13_2, sent[:len(sent)-1]); err == nil {
		t.Error("expected a login with no debug screen flag to be refused")
	}
}

// packBlockPos1_14 and packBlockPos1_13_2 pack a block position the way each
// side of the step does.
func packBlockPos1_14(x, y, z int64) int64 {
	return (x&0x3FFFFFF)<<38 | (z&0x3FFFFFF)<<12 | y&0xFFF
}

func packBlockPos1_13_2(x, y, z int64) int64 {
	return (x&0x3FFFFFF)<<38 | (y&0xFFF)<<26 | z&0x3FFFFFF
}

// The spawn position 1.13.2 reads is the same one long, with the height moved
// from the lowest bits to between the other two, the signs kept.
func TestDowngradeSetDefaultSpawnPositionTo1_13_2RepacksThePosition(t *testing.T) {
	for _, pos := range [][3]int64{{0, 0, 0}, {8, 65, 8}, {-1, -1, -1}, {-30000000, 2047, 29999999}, {123456, -64, -654321}} {
		sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return ms.WriteLong(packBlockPos1_14(pos[0], pos[1], pos[2]))
		})

		want := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return ms.WriteLong(packBlockPos1_13_2(pos[0], pos[1], pos[2]))
		})

		if got := runTransformer(t, DowngradeSetDefaultSpawnPositionTo1_13_2, sent); !bytes.Equal(got, want) {
			t.Errorf("%v to 1.13.2 = % x\nwant = % x", pos, got, want)
		}
	}
}

// The metadata 1.13.2 reads of a player's stance is the flag byte alone: the
// pose has no field and no serializer there.
func TestDowngradeSetEntityDataTo1_13_2DropsThePose(t *testing.T) {
	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(42),
			ms.WriteByte(0),
			ms.WriteVarInt(byteSerializer),
			ms.WriteByte(0x02),
			ms.WriteByte(6),
			ms.WriteVarInt(poseSerializer1_19_1),
			ms.WriteVarInt(5),
			ms.WriteByte(entityDataTerminator),
		)
	})

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(42),
			ms.WriteByte(0),
			ms.WriteVarInt(byteSerializer),
			ms.WriteByte(0x02),
			ms.WriteByte(entityDataTerminator),
		)
	})

	if got := runTransformer(t, DowngradeSetEntityDataTo1_13_2, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.13.2 = % x\nwant = % x", got, want)
	}

	other := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteVarInt(42), ms.WriteByte(7), ms.WriteVarInt(2), ms.WriteFloat(20), ms.WriteByte(entityDataTerminator))
	})

	if err := failingTransformer(t, DowngradeSetEntityDataTo1_13_2, other); err == nil {
		t.Error("expected an entry of a serializer the rewrite does not know to be refused")
	}
}

// Paired carries each half of the body by its own transformer, and a half
// with none as it stands.
func TestPairedCarriesEachHalfByItsOwnTransformer(t *testing.T) {
	first := []byte{1, 2, 3}
	second := []byte{4, 5}

	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteByteArray(first), ms.WriteBytes(second))
	})

	dropFirstByte := func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if _, err := in.ReadByte(); err != nil {
			return err
		}

		return copyRest(in, out)
	}

	for _, c := range []struct {
		name          string
		first, second func(in, out *streams.MinecraftStream) error
		want          []byte
	}{
		{"neither", nil, nil, sent},
		{"the first", dropFirstByte, nil, []byte{2, 2, 3, 4, 5}},
		{"the second", nil, dropFirstByte, []byte{3, 1, 2, 3, 5}},
		{"both", dropFirstByte, dropFirstByte, []byte{2, 2, 3, 5}},
	} {
		if got := runTransformer(t, Paired(c.first, c.second), sent); !bytes.Equal(got, c.want) {
			t.Errorf("%s: Paired() = % x, want % x", c.name, got, c.want)
		}
	}
}

// chunkWithSectionLight1_14 is a chunk of two sections, at the bottom of the
// world and two above it, as 1.14 sends it and its light: the chunk behind its
// length, then the light update. The lower section is paletted, the upper
// one packs ids directly. The light names both sections' sky light and the
// upper one's block light, and the empty masks everything else.
func chunkWithSectionLight1_14(t *testing.T, whole bool, lightX int32) []byte {
	t.Helper()

	sections := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := chunkSection(ms, 4, []int32{0, 9}, span(testEntries(4096, 1), 4)); err != nil {
			return err
		}

		if err := chunkSection(ms, 14, nil, span(testEntries(4096, 14), 14)); err != nil {
			return err
		}

		if whole {
			for i := range int32(256) {
				if err := ms.WriteInt(i % 3); err != nil {
					return err
				}
			}
		}

		return nil
	})

	chunk := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteInt(3),
			ms.WriteInt(-4),
			ms.WriteBoolean(whole),
			ms.WriteVarInt(0b101),
			nbt.WriteNamed(ms, "", nbt.Compound{"MOTION_BLOCKING": nbt.LongArray(span(testEntries(256, 9), 9))}),
			ms.WriteByteArray(sections),
			ms.WriteVarInt(0),
		)
	})

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteByteArray(chunk),
			ms.WriteVarInt(lightX),
			ms.WriteVarInt(-4),
			ms.WriteVarInt(0b1010),
			ms.WriteVarInt(0b1000),
			ms.WriteVarInt(0b111111111111110101),
			ms.WriteVarInt(0b111111111111110111),
			ms.WriteByteArray(lightArray(0xF0)),
			ms.WriteByteArray(lightArray(0xFF)),
			ms.WriteByteArray(lightArray(0x11)),
		)
	})
}

// chunkWithSectionLight1_13_2 is the same chunk as 1.13.2 reads it: each
// section without its block count, with its block light and then its sky
// light behind it, and the lower section's block light, which the light
// says nothing of, as an array of none.
func chunkWithSectionLight1_13_2(t *testing.T, whole bool) []byte {
	t.Helper()

	sections := encodeBody(t, func(ms *streams.MinecraftStream) error {
		lower := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return chunkSection(ms, 4, []int32{0, 9}, span(testEntries(4096, 1), 4))
		})

		upper := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return chunkSection(ms, 14, nil, span(testEntries(4096, 14), 14))
		})

		steps := []error{
			ms.WriteBytes(lower[2:]),
			ms.WriteBytes(lightArray(0)),
			ms.WriteBytes(lightArray(0xF0)),
			ms.WriteBytes(upper[2:]),
			ms.WriteBytes(lightArray(0x11)),
			ms.WriteBytes(lightArray(0xFF)),
		}

		if err := writeAll(steps...); err != nil {
			return err
		}

		if whole {
			for i := range int32(256) {
				if err := ms.WriteInt(i % 3); err != nil {
					return err
				}
			}
		}

		return nil
	})

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteInt(3),
			ms.WriteInt(-4),
			ms.WriteBoolean(whole),
			ms.WriteVarInt(0b101),
			ms.WriteByteArray(sections),
			ms.WriteVarInt(0),
		)
	})
}

// The chunk 1.13.2 reads is 1.14's without its heightmaps, each section's
// light from the light update put behind its blocks.
func TestDowngradeLevelChunkWithSectionLightTo1_13_2PutsTheLightInTheSections(t *testing.T) {
	for _, whole := range []bool{true, false} {
		got := runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_13_2, chunkWithSectionLight1_14(t, whole, 3))

		if want := chunkWithSectionLight1_13_2(t, whole); !bytes.Equal(got, want) {
			t.Errorf("whole %t: to 1.13.2 = %d bytes, want %d", whole, len(got), len(want))
		}
	}
}

func TestDowngradeLevelChunkWithSectionLightTo1_13_2RefusesTheLightOfAnotherChunk(t *testing.T) {
	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_13_2, chunkWithSectionLight1_14(t, true, 4)); err == nil {
		t.Error("expected a chunk paired with another chunk's light to be refused")
	}
}

func TestDowngradeLevelChunkWithSectionLightTo1_13_2RefusesLightCutShort(t *testing.T) {
	sent := chunkWithSectionLight1_14(t, true, 3)

	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_13_2, sent[:len(sent)-1]); err == nil {
		t.Error("expected light with its last array cut short to be refused")
	}
}
