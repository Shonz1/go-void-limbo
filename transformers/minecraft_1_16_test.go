package transformers

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// 1.15.2 reads the uuid of the login success as text, the way a uuid spells
// itself out, where 1.16 reads its sixteen bytes; the name is copied.
func TestDowngradeLoginSuccessTo1_15_2SpellsTheUuidOut(t *testing.T) {
	const uuid = "069a79f4-44e9-4726-a5be-fca90e38aaf5"

	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteUuid(uuid); err != nil {
			return err
		}

		return ms.WriteString("Notch")
	})

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteString(uuid); err != nil {
			return err
		}

		return ms.WriteString("Notch")
	})

	if got := runTransformer(t, DowngradeLoginSuccessTo1_15_2, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.15.2 = % x\nwant = % x", got, want)
	}
}

// playLogin1_15_2 is the login as 1.15.2 lays it out, saying what
// playLogin1_16_1 says: no previous game mode, no dimension names, no
// dimension types, the dimension a number, the level type a name, and two
// flags where 1.16.1 has four.
func playLogin1_15_2(t *testing.T, gameMode byte, dimension int32, maxPlayers byte, levelType string) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteByte(gameMode),
			ms.WriteInt(dimension),
			ms.WriteLong(0x1122334455667788),
			ms.WriteByte(maxPlayers),
			ms.WriteString(levelType),
			ms.WriteVarInt(8),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// login1_16 is a 1.16 login put into the dimension the test asks for, in a
// world that is flat or not.
func login1_16(t *testing.T, dimension string, flat bool) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteByte(2 | 8),
			ms.WriteByte(0xFF),
			ms.WriteVarInt(2),
			ms.WriteString("minecraft:overworld"),
			ms.WriteString("minecraft:the_end"),
			ms.WriteBytes(dimensionList(t)),
			ms.WriteBytes(dimensionTypeName(t)),
			ms.WriteString(dimension),
			ms.WriteLong(0x1122334455667788),
			ms.WriteByte(20),
			ms.WriteVarInt(8),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
			ms.WriteBoolean(true),
			ms.WriteBoolean(flat),
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// The login 1.15.2 reads is the login 1.16 reads with everything of a
// dimension but its number taken out, and the flat flag said as a level type.
func TestDowngradePlayLoginTo1_15_2NumbersTheDimension(t *testing.T) {
	sent := playLogin1_16_1(t, 2|8, dimensionList(t), dimensionTypeName(t), 20)

	if got, want := runTransformer(t, DowngradePlayLoginTo1_15_2, sent), playLogin1_15_2(t, 2|8, 1, 20, "default"); !bytes.Equal(got, want) {
		t.Errorf("to 1.15.2 = % x\nwant = % x", got, want)
	}

	cases := []struct {
		dimension string
		flat      bool
		number    int32
		levelType string
	}{
		{"minecraft:overworld", false, 0, "default"},
		{"minecraft:overworld", true, 0, "flat"},
		{"minecraft:the_nether", false, -1, "default"},
		{"minecraft:the_end", true, 1, "flat"},
	}

	for _, c := range cases {
		got := runTransformer(t, DowngradePlayLoginTo1_15_2, login1_16(t, c.dimension, c.flat))

		if want := playLogin1_15_2(t, 2|8, c.number, 20, c.levelType); !bytes.Equal(got, want) {
			t.Errorf("%s, flat %t: to 1.15.2 = % x\nwant = % x", c.dimension, c.flat, got, want)
		}
	}
}

func TestDowngradePlayLoginTo1_15_2RefusesADimensionItHasNoNumberFor(t *testing.T) {
	if err := failingTransformer(t, DowngradePlayLoginTo1_15_2, login1_16(t, "minecraft:lobby", false)); err == nil {
		t.Error("expected a dimension 1.15.2 does not hold to be refused")
	}
}

// span packs entries end to end the way 1.15.2's own bit storage does, one
// entry at a time, as what spanEntries is checked against.
func span(entries []uint64, bits int) []int64 {
	out := make([]int64, (len(entries)*bits+63)/64)

	for i, entry := range entries {
		for bit := range bits {
			if entry&(1<<bit) != 0 {
				at := i*bits + bit
				out[at/64] |= 1 << (at % 64)
			}
		}
	}

	return out
}

// pack packs entries the way 1.16 does, with no entry crossing a long.
func pack(entries []uint64, bits int) []int64 {
	perLong := 64 / bits

	out := make([]int64, (len(entries)+perLong-1)/perLong)
	for i, entry := range entries {
		out[i/perLong] |= int64(entry << (i % perLong * bits))
	}

	return out
}

func testEntries(count, bits int) []uint64 {
	entries := make([]uint64, count)
	for i := range entries {
		entries[i] = uint64(i*2654435761) & (1<<bits - 1)
	}

	return entries
}

// Entries that do not divide a long are where the two packings part: 1.16
// leaves the rest of the long, and 1.15.2 carries on into the next. At every
// width a section or a heightmap is packed at, what comes out is what 1.15.2's
// own storage would hold, in as many longs as the bits come to.
func TestSpanEntriesPacksEndToEnd(t *testing.T) {
	for _, c := range []struct{ bits, count int }{{4, 4096}, {5, 4096}, {6, 4096}, {7, 4096}, {8, 4096}, {9, 256}, {14, 4096}, {15, 4096}} {
		entries := testEntries(c.count, c.bits)

		got, err := spanEntries(pack(entries, c.bits), c.bits, c.count)
		if err != nil {
			t.Fatalf("%d bits: spanEntries() error = %v", c.bits, err)
		}

		if want := span(entries, c.bits); !reflect.DeepEqual(got, want) {
			t.Errorf("%d bits: spanEntries() is not the entries end to end", c.bits)
		}

		if want := (c.count*c.bits + 63) / 64; len(got) != want {
			t.Errorf("%d bits: %d longs, want %d", c.bits, len(got), want)
		}
	}

	if _, err := spanEntries(make([]int64, 3), 9, 256); err == nil {
		t.Error("expected a heightmap of the wrong length to be refused")
	}
}

// chunkSection is one section as both sides of the step lay it out, given
// its data as that side packs it.
func chunkSection(ms *streams.MinecraftStream, bits byte, palette []int32, data []int64) error {
	if err := ms.WriteShort(100); err != nil {
		return err
	}

	if err := ms.WriteByte(bits); err != nil {
		return err
	}

	if palette != nil {
		if err := ms.WriteVarInt(int32(len(palette))); err != nil {
			return err
		}

		for _, id := range palette {
			if err := ms.WriteVarInt(id); err != nil {
				return err
			}
		}
	}

	if err := ms.WriteVarInt(int32(len(data))); err != nil {
		return err
	}

	for _, long := range data {
		if err := ms.WriteLong(long); err != nil {
			return err
		}
	}

	return nil
}

// chunk1_15_2Step is a whole chunk of two sections -- one behind a palette
// of five bits, one of ids at fourteen -- as 1.16 lays it out, or as 1.15.2
// does.
func chunk1_15_2Step(t *testing.T, older bool) []byte {
	t.Helper()

	packing := pack
	if older {
		packing = span
	}

	palette := make([]int32, 20)
	for i := range palette {
		palette[i] = int32(i * 3)
	}

	sections := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := chunkSection(ms, 5, palette, packing(testEntries(4096, 5), 5)); err != nil {
			return err
		}

		return chunkSection(ms, 14, nil, packing(testEntries(4096, 14), 14))
	})

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteInt(3); err != nil {
			return err
		}

		if err := ms.WriteInt(-4); err != nil {
			return err
		}

		if err := ms.WriteBoolean(true); err != nil {
			return err
		}

		if !older {
			if err := ms.WriteBoolean(true); err != nil {
				return err
			}
		}

		if err := ms.WriteVarInt(0b101); err != nil {
			return err
		}

		heightmaps := nbt.Compound{"MOTION_BLOCKING": nbt.LongArray(packing(testEntries(256, 9), 9))}
		if err := nbt.WriteNamed(ms, "", heightmaps); err != nil {
			return err
		}

		for range 1024 {
			if err := ms.WriteInt(1); err != nil {
				return err
			}
		}

		if err := ms.WriteByteArray(sections); err != nil {
			return err
		}

		return ms.WriteVarInt(0)
	})
}

// The chunk 1.15.2 reads is the chunk 1.16 reads without its second flag,
// and with its heightmaps and the blocks of its sections packed end to end.
func TestDowngradeLevelChunkTo1_15_2DropsTheFlagAndPacksEndToEnd(t *testing.T) {
	got := runTransformer(t, DowngradeLevelChunkTo1_15_2, chunk1_15_2Step(t, false))

	if want := chunk1_15_2Step(t, true); !bytes.Equal(got, want) {
		t.Errorf("to 1.15.2 = %d bytes, want the %d of the chunk as 1.15.2 lays it out", len(got), len(want))
	}
}

func TestDowngradeLevelChunkTo1_15_2RefusesSectionsTheMaskDoesNotName(t *testing.T) {
	sent := chunk1_15_2Step(t, false)

	// The mask, behind the coordinates and the two flags, names one section
	// fewer than the buffer holds.
	sent[10] = 0b001

	if err := failingTransformer(t, DowngradeLevelChunkTo1_15_2, sent); err == nil {
		t.Error("expected a section buffer longer than its mask to be refused")
	}
}

// The light 1.15.2 reads is the light 1.16 reads without the trust edges
// flag behind the coordinates.
func TestDowngradeLightUpdateTo1_15_2DropsTheTrustEdgesFlag(t *testing.T) {
	light := func(older bool) []byte {
		return encodeBody(t, func(ms *streams.MinecraftStream) error {
			if err := ms.WriteVarInt(300); err != nil {
				return err
			}

			if err := ms.WriteVarInt(-2); err != nil {
				return err
			}

			if !older {
				if err := ms.WriteBoolean(true); err != nil {
					return err
				}
			}

			for _, mask := range []int32{0b10, 0, 0x3FFFD, 0x3FFFF} {
				if err := ms.WriteVarInt(mask); err != nil {
					return err
				}
			}

			return ms.WriteByteArray(bytes.Repeat([]byte{0xFF}, 2048))
		})
	}

	if got, want := runTransformer(t, DowngradeLightUpdateTo1_15_2, light(false)), light(true); !bytes.Equal(got, want) {
		t.Errorf("to 1.15.2 = %d bytes, want %d", len(got), len(want))
	}
}
