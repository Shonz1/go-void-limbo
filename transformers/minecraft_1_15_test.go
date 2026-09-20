package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The login 1.14.4 reads is the login 1.15 reads without the seed behind the
// dimension and without the death screen flag at its end.
func TestDowngradePlayLoginTo1_14_4DropsTheSeedAndTheDeathScreenFlag(t *testing.T) {
	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteByte(2 | 8),
			ms.WriteInt(1),
			ms.WriteByte(20),
			ms.WriteString("flat"),
			ms.WriteVarInt(8),
			ms.WriteBoolean(true),
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})

	if got := runTransformer(t, DowngradePlayLoginTo1_14_4, playLogin1_15_2(t, 2|8, 1, 20, "flat")); !bytes.Equal(got, want) {
		t.Errorf("to 1.14.4 = % x\nwant = % x", got, want)
	}
}

func TestDowngradePlayLoginTo1_14_4RefusesALoginCutShort(t *testing.T) {
	sent := playLogin1_15_2(t, 2, 0, 20, "default")

	if err := failingTransformer(t, DowngradePlayLoginTo1_14_4, sent[:len(sent)-1]); err == nil {
		t.Error("expected a login with no death screen flag to be refused")
	}
}

// chunk1_15Step is a chunk of one section as 1.15 lays it out, or as 1.14.4
// does. Its biomes differ from one of 1.15's cells to the next, and from one
// layer of them to the next, so that where a column's biome came from shows.
func chunk1_15Step(t *testing.T, whole bool, older bool) []byte {
	t.Helper()

	sections := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return chunkSection(ms, 14, nil, span(testEntries(4096, 14), 14))
	})

	heightmaps := nbt.Compound{"MOTION_BLOCKING": nbt.LongArray(span(testEntries(256, 9), 9))}

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteInt(3); err != nil {
			return err
		}

		if err := ms.WriteInt(-4); err != nil {
			return err
		}

		if err := ms.WriteBoolean(whole); err != nil {
			return err
		}

		if err := ms.WriteVarInt(0b100); err != nil {
			return err
		}

		if err := nbt.WriteNamed(ms, "", heightmaps); err != nil {
			return err
		}

		if whole && !older {
			for i := range int32(biomes1_16_4) {
				if err := ms.WriteInt(100 + i); err != nil {
					return err
				}
			}
		}

		data := sections

		if whole && older {
			columns := encodeBody(t, func(ms *streams.MinecraftStream) error {
				for z := range int32(16) {
					for x := range int32(16) {
						// The cell of the lowest layer the column stands in.
						if err := ms.WriteInt(100 + z/4*4 + x/4); err != nil {
							return err
						}
					}
				}

				return nil
			})

			data = append(append([]byte(nil), sections...), columns...)
		}

		if err := ms.WriteByteArray(data); err != nil {
			return err
		}

		// One block entity, which is copied as it stands.
		if err := ms.WriteVarInt(1); err != nil {
			return err
		}

		return nbt.WriteNamed(ms, "", nbt.Compound{"id": nbt.String("minecraft:sign")})
	})
}

// The chunk 1.14.4 reads is the chunk 1.15 reads with its biomes moved onto
// the end of its sections, one to a column, each column given the biome of the
// lowest of 1.15's cells it stands in.
func TestDowngradeLevelChunkTo1_14_4MovesTheBiomesBehindTheSections(t *testing.T) {
	got := runTransformer(t, DowngradeLevelChunkTo1_14_4, chunk1_15Step(t, true, false))

	if want := chunk1_15Step(t, true, true); !bytes.Equal(got, want) {
		t.Errorf("to 1.14.4 = %d bytes, want %d\n got = % x\nwant = % x", len(got), len(want), got[:64], want[:64])
	}
}

// A chunk that is not a whole one carries no biomes on either side of the
// step, and is laid out alike.
func TestDowngradeLevelChunkTo1_14_4LeavesAPartialChunkAsItIs(t *testing.T) {
	sent := chunk1_15Step(t, false, false)

	if got := runTransformer(t, DowngradeLevelChunkTo1_14_4, sent); !bytes.Equal(got, sent) {
		t.Errorf("to 1.14.4 = % x\nwant = % x", got, sent)
	}
}

func TestDowngradeLevelChunkTo1_14_4RefusesBiomesCutShort(t *testing.T) {
	sent := chunk1_15Step(t, true, false)

	// Far enough in to be past the heightmaps and short of the last biome.
	if err := failingTransformer(t, DowngradeLevelChunkTo1_14_4, sent[:len(sent)/2]); err == nil {
		t.Error("expected a whole chunk with fewer biomes than it holds cells to be refused")
	}
}

// The add player 1.14.4 reads is the one 1.15 reads with a run of entity
// metadata behind it, which is an empty one here: its end alone.
func TestDowngradeAddPlayerTo1_14_4ClosesAnEmptyRunOfMetadata(t *testing.T) {
	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteVarInt(300),
			ms.WriteUuid("069a79f4-44e9-4726-a5be-fca90e38aaf5"),
			ms.WriteDouble(1.5),
			ms.WriteDouble(64),
			ms.WriteDouble(-2.5),
			ms.WriteByte(0x40),
			ms.WriteByte(0xF0),
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})

	want := append(append([]byte(nil), sent...), 0xFF)

	if got := runTransformer(t, DowngradeAddPlayerTo1_14_4, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.14.4 = % x\nwant = % x", got, want)
	}
}
