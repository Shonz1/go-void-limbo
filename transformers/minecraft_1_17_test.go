package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// flatRegistryCodec and flatDimensionType stand in for what package gamedata
// hands the transformer for 1.16.4: the registries and the dimension type,
// told apart from 1.17's by their content.
func flatRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:dimension_type": nbt.Compound{"type": nbt.String("minecraft:dimension_type")}})
	})
}

func flatDimensionType(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"logical_height": nbt.Int(256)})
	})
}

// The login 1.16.4 reads is the login 1.17 reads with the registries and the
// spelled-out dimension type in the middle swapped for 1.16.4's own.
// Everything else is copied.
func TestDowngradePlayLoginTo1_16_4SwapsTheRegistriesAndTheDimensionType(t *testing.T) {
	newerCodec, newerDimensionType := floorRegistryCodec(t), floorDimensionType(t)
	olderCodec, olderDimensionType := flatRegistryCodec(t), flatDimensionType(t)

	sent := playLogin1_17_1(t, newerCodec, newerDimensionType)
	got := runTransformer(t, DowngradePlayLoginTo1_16_4(olderCodec, olderDimensionType), sent)
	want := playLogin1_17_1(t, olderCodec, olderDimensionType)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.16.4 = % x\nwant = % x", got, want)
	}
}

func TestDowngradePlayLoginTo1_16_4RefusesWithoutItsData(t *testing.T) {
	sent := playLogin1_17_1(t, floorRegistryCodec(t), floorDimensionType(t))

	if err := failingTransformer(t, DowngradePlayLoginTo1_16_4(nil, flatDimensionType(t)), sent); err == nil {
		t.Error("expected a login with no registries to write to be refused")
	}

	if err := failingTransformer(t, DowngradePlayLoginTo1_16_4(flatRegistryCodec(t), nil), sent); err == nil {
		t.Error("expected a login with no dimension type to write to be refused")
	}
}

// section1_17 is one section as both sides of the step lay it out: a block
// count, four bits over a palette of one entry, and 256 longs of zeros.
func section1_17(blockCount int16, state int32) []byte {
	section := []byte{byte(blockCount >> 8), byte(blockCount), linearPaletteBits, 1}
	section = streams.AppendVarInt(section, state)

	return append(section, singleValueSectionData...)
}

// heightmap packs one height for every column, nine bits an entry, seven
// entries a long.
func heightmap(height int64) nbt.LongArray {
	packed := make(nbt.LongArray, 37)
	for i := range 256 {
		packed[i/7] |= height << (i % 7 * 9)
	}

	return packed
}

func chunk1_17(t *testing.T, mask int64, sections []byte, heightmaps nbt.Compound, blockEntities int32) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{ms.WriteInt(3), ms.WriteInt(-2), ms.WriteVarInt(1), ms.WriteLong(mask), nbt.WriteNamed(ms, "", heightmaps), ms.WriteVarInt(24 * 64)}
		for range 24 * 64 {
			steps = append(steps, ms.WriteVarInt(0))
		}

		steps = append(steps, ms.WriteByteArray(sections), ms.WriteVarInt(blockEntities))

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// A 1.16.4 chunk is the sixteen sections from zero up: the ones below and
// above come off, the mask moves down to match and turns into a plain number
// behind the whole chunk flag, the heightmaps come down by the sixty-four
// blocks the world lost at the bottom, and the biomes are 1.16.4's 1024.
func TestDowngradeLevelChunkTo1_16_4CutsTheWorldToSixteenSections(t *testing.T) {
	// Sections 3, 4, 19 and 20 of the twenty-four: the two in the middle are
	// 1.16.4's first and last.
	var sections []byte
	for _, state := range []int32{33, 44, 199, 200} {
		sections = append(sections, section1_17(4096, state)...)
	}

	sent := chunk1_17(t, 1<<3|1<<4|1<<19|1<<20, sections, nbt.Compound{
		"MOTION_BLOCKING": heightmap(129),
		"WORLD_SURFACE":   heightmap(10),
	}, 0)

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(3), ms.WriteInt(-2), ms.WriteBoolean(true), ms.WriteVarInt(1<<0 | 1<<15),
			nbt.WriteNamed(ms, "", nbt.Compound{"MOTION_BLOCKING": heightmap(65), "WORLD_SURFACE": heightmap(0)}),
			ms.WriteVarInt(1024),
		}
		for range 1024 {
			steps = append(steps, ms.WriteVarInt(0))
		}

		steps = append(steps, ms.WriteByteArray(append(section1_17(4096, 44), section1_17(4096, 199)...)), ms.WriteVarInt(0))

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})

	if got := runTransformer(t, DowngradeLevelChunkTo1_16_4, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.16.4 = %d bytes, want %d\ngot  % x\nwant % x", len(got), len(want), got[:64], want[:64])
	}
}

// A column whose first free block is above 1.16.4's world says the top of it.
func TestDowngradeLevelChunkTo1_16_4StopsTheHeightmapsAtTheTop(t *testing.T) {
	sent := chunk1_17(t, 0, nil, nbt.Compound{"MOTION_BLOCKING": heightmap(384)}, 0)
	got := runTransformer(t, DowngradeLevelChunkTo1_16_4, sent)

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"MOTION_BLOCKING": heightmap(256)})
	})

	if !bytes.Contains(got, want) {
		t.Error("the heightmap does not stop at 256")
	}
}

func TestDowngradeLevelChunkTo1_16_4Refuses(t *testing.T) {
	cases := map[string][]byte{
		"block entities":         chunk1_17(t, 0, nil, nbt.Compound{}, 1),
		"a short heightmap":      chunk1_17(t, 0, nil, nbt.Compound{"MOTION_BLOCKING": nbt.LongArray{1}}, 0),
		"a section unnamed":      chunk1_17(t, 0, section1_17(1, 1), nbt.Compound{}, 0),
		"a named section absent": chunk1_17(t, 1<<4, nil, nbt.Compound{}, 0),
		"empty body":             nil,
	}

	for name, body := range cases {
		if err := failingTransformer(t, DowngradeLevelChunkTo1_16_4, body); err == nil {
			t.Errorf("%s: expected the packet to be refused", name)
		}
	}
}

func lightArray(fill byte) []byte {
	return bytes.Repeat([]byte{fill}, lightArrayBytes)
}

// The light 1.16.4 reads names its eighteen sections by plain numbers and
// counts no arrays: the masks move down four bits, and the arrays of the
// sections that fall off either end come off with them.
func TestDowngradeLightUpdateTo1_16_4CutsTheLightToEighteenSections(t *testing.T) {
	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteVarInt(3), ms.WriteVarInt(-2), ms.WriteBoolean(true),
			// Sky light on sections 3, 4, 21 and 22 of the twenty-six, block
			// light on 5, and nothing at all in the empty block mask.
			ms.WriteVarInt(1), ms.WriteLong(1<<3 | 1<<4 | 1<<21 | 1<<22),
			ms.WriteVarInt(1), ms.WriteLong(1 << 5),
			ms.WriteVarInt(1), ms.WriteLong((1<<26 - 1) &^ (1<<3 | 1<<4 | 1<<21 | 1<<22)),
			ms.WriteVarInt(0),
			ms.WriteVarInt(4),
			ms.WriteByteArray(lightArray(3)), ms.WriteByteArray(lightArray(4)), ms.WriteByteArray(lightArray(21)), ms.WriteByteArray(lightArray(22)),
			ms.WriteVarInt(1),
			ms.WriteByteArray(lightArray(5)),
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteVarInt(3), ms.WriteVarInt(-2), ms.WriteBoolean(true),
			ms.WriteVarInt(1<<0 | 1<<17),
			ms.WriteVarInt(1 << 1),
			ms.WriteVarInt((1<<18 - 1) &^ (1<<0 | 1<<17)),
			ms.WriteVarInt(0),
			ms.WriteByteArray(lightArray(4)), ms.WriteByteArray(lightArray(21)),
			ms.WriteByteArray(lightArray(5)),
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})

	if got := runTransformer(t, DowngradeLightUpdateTo1_16_4, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.16.4 = %d bytes, want %d\ngot  % x\nwant % x", len(got), len(want), got[:16], want[:16])
	}
}

func TestDowngradeLightUpdateTo1_16_4RefusesArraysTheMasksDoNotCount(t *testing.T) {
	light := func(mask int64, arrays int) []byte {
		return encodeBody(t, func(ms *streams.MinecraftStream) error {
			steps := []error{ms.WriteVarInt(0), ms.WriteVarInt(0), ms.WriteBoolean(true), ms.WriteVarInt(1), ms.WriteLong(mask), ms.WriteVarInt(0), ms.WriteVarInt(0), ms.WriteVarInt(0), ms.WriteVarInt(int32(arrays))}
			for range arrays {
				steps = append(steps, ms.WriteByteArray(lightArray(1)))
			}

			steps = append(steps, ms.WriteVarInt(0))

			for _, err := range steps {
				if err != nil {
					return err
				}
			}

			return nil
		})
	}

	for name, body := range map[string][]byte{"too few": light(1<<5|1<<6, 1), "too many": light(1<<5, 2), "empty body": nil} {
		if err := failingTransformer(t, DowngradeLightUpdateTo1_16_4, body); err == nil {
			t.Errorf("%s: expected the packet to be refused", name)
		}
	}
}

// 1.17 put the dismount vehicle flag onto the end of the player position,
// and it comes off on the way to 1.16.4.
func TestDowngradePlayerPositionTo1_16_4DropsTheDismountFlag(t *testing.T) {
	position := append(bytes.Repeat([]byte{0x11}, 3*8+2*4), 0x1F, 0x80, 0x01)

	if got := runTransformer(t, DowngradePlayerPositionTo1_16_4, append(bytes.Clone(position), 0x00)); !bytes.Equal(got, position) {
		t.Errorf("to 1.16.4 = % x, want % x", got, position)
	}

	if err := failingTransformer(t, DowngradePlayerPositionTo1_16_4, position); err == nil {
		t.Error("expected a position with no dismount flag to be refused")
	}
}

// 1.17 put the angle onto the end of the default spawn position, and it
// comes off on the way to 1.16.4.
func TestDowngradeSetDefaultSpawnPositionTo1_16_4DropsTheAngle(t *testing.T) {
	position := []byte{1, 2, 3, 4, 5, 6, 7, 8}

	if got := runTransformer(t, DowngradeSetDefaultSpawnPositionTo1_16_4, append(bytes.Clone(position), 0x42, 0xB4, 0x00, 0x00)); !bytes.Equal(got, position) {
		t.Errorf("to 1.16.4 = % x, want % x", got, position)
	}

	if err := failingTransformer(t, DowngradeSetDefaultSpawnPositionTo1_16_4, position); err == nil {
		t.Error("expected a spawn position with no angle to be refused")
	}
}

// A 1.17 removal is the id alone; a 1.16.4 one is a count and the ids.
func TestDowngradeRemoveEntitiesTo1_16_4PutsTheCountBack(t *testing.T) {
	got := runTransformer(t, DowngradeRemoveEntitiesTo1_16_4, []byte{0x80, 0x01})

	if want := []byte{0x01, 0x80, 0x01}; !bytes.Equal(got, want) {
		t.Errorf("to 1.16.4 = % x, want % x", got, want)
	}

	if err := failingTransformer(t, DowngradeRemoveEntitiesTo1_16_4, nil); err == nil {
		t.Error("expected a removal with no id to be refused")
	}
}
