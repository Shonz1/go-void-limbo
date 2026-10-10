package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// chunk1_9_3 lays a chunk packet out as 1.9.3 reads it, around the section
// buffer given, with the block entities given behind it.
func chunk1_9_3(t *testing.T, whole bool, mask int32, sections []byte, blockEntities []nbt.Compound) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := writeAll(
			ms.WriteInt(3),
			ms.WriteInt(-4),
			ms.WriteBoolean(whole),
			ms.WriteVarInt(mask),
			ms.WriteByteArray(sections),
			ms.WriteVarInt(int32(len(blockEntities))),
		); err != nil {
			return err
		}

		for _, blockEntity := range blockEntities {
			if err := nbt.WriteNamed(ms, "", blockEntity); err != nil {
				return err
			}
		}

		return nil
	})
}

// chunk1_9_2 lays a chunk packet out as 1.9.2 reads it, which is 1.9.3's cut
// off behind its sections.
func chunk1_9_2(t *testing.T, whole bool, mask int32, sections []byte) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteInt(3),
			ms.WriteInt(-4),
			ms.WriteBoolean(whole),
			ms.WriteVarInt(mask),
			ms.WriteByteArray(sections),
		)
	})
}

// 1.9.2 reads the chunk 1.9.3 does up to the end of its sections, and nothing
// behind them: the count of block entities goes, and the block entities with
// it.
func TestDowngradeLevelChunkWithSectionLightTo1_9_2(t *testing.T) {
	sections := section1_13(4, []int32{0, 2 << 4}, make([]int64, 256), true)

	sent := chunk1_9_3(t, true, 0b1, sections, nil)
	want := chunk1_9_2(t, true, 0b1, sections)

	if got := runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_9_2, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.9.2 = % x, want % x", got, want)
	}

	chest := nbt.Compound{"id": nbt.String("minecraft:chest"), "x": nbt.Int(48), "y": nbt.Int(64), "z": nbt.Int(-60)}
	sign := nbt.Compound{"id": nbt.String("minecraft:sign"), "Text1": nbt.String(`{"text":"welcome"}`)}

	withEntities := chunk1_9_3(t, false, 0b1, sections, []nbt.Compound{chest, sign})
	wantPart := chunk1_9_2(t, false, 0b1, sections)

	if got := runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_9_2, withEntities); !bytes.Equal(got, wantPart) {
		t.Errorf("to 1.9.2 with block entities = % x, want % x", got, wantPart)
	}
}

// A chunk that is not what its count of block entities says it is -- one cut
// short of them, or one with bytes past the last -- is refused rather than
// sent on.
func TestDowngradeLevelChunkWithSectionLightTo1_9_2Refuses(t *testing.T) {
	sections := section1_13(4, []int32{0}, make([]int64, 256), true)

	short := chunk1_9_3(t, false, 0b1, sections, []nbt.Compound{{"id": nbt.String("minecraft:chest")}})
	short = short[:len(short)-3]

	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_9_2, short); err == nil {
		t.Error("expected a block entity cut short to be refused")
	}

	long := append(chunk1_9_3(t, false, 0b1, sections, nil), 0x00)
	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_9_2, long); err == nil {
		t.Error("expected a byte past the block entities to be refused")
	}

	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_9_2, chunk1_9_2(t, false, 0b1, sections)); err == nil {
		t.Error("expected a chunk with no count of block entities to be refused")
	}
}
