package transformers

import (
	"bytes"
	"math"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
)

// playerPosition1_9 lays the player position out as 1.9 reads it: the
// position, the rotation, the flags and the teleport id.
func playerPosition1_9(t *testing.T, teleportId int32) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteDouble(8.5),
			ms.WriteDouble(65),
			ms.WriteDouble(-8.5),
			ms.WriteFloat(90),
			ms.WriteFloat(-10),
			ms.WriteByte(0x00),
			ms.WriteVarInt(teleportId),
		)
	})
}

// 1.8 reads the player position 1.9 does, cut off before the teleport id.
func TestDowngradePlayerPositionTo1_8(t *testing.T) {
	sent := playerPosition1_9(t, 300)

	want := sent[:playerPositionHead1_9]

	if got := runTransformer(t, DowngradePlayerPositionTo1_8, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.8 = % x, want % x", got, want)
	}

	if err := failingTransformer(t, DowngradePlayerPositionTo1_8, sent[:len(sent)-1]); err == nil {
		t.Error("expected a position with no teleport id to be refused")
	}

	if err := failingTransformer(t, DowngradePlayerPositionTo1_8, append(sent, 0x00)); err == nil {
		t.Error("expected a position with bytes past the teleport id to be refused")
	}
}

// fixedPoint1_8 is a coordinate as 1.8 carries one: thirty-seconds of a
// block, rounded down.
func fixedPoint1_8(coordinate float64) int32 {
	return int32(math.Floor(coordinate * 32))
}

// 1.8 reads the teleport 1.9 does with the position as three ints of
// thirty-seconds of a block, rounded down, the negative ones included.
func TestDowngradeEntityPositionSyncTo1_8(t *testing.T) {
	x, y, z := 8.53, -1.5, -0.01

	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(7),
			ms.WriteDouble(x),
			ms.WriteDouble(y),
			ms.WriteDouble(z),
			ms.WriteByte(64),
			ms.WriteByte(0xF0),
			ms.WriteBoolean(true),
		)
	})

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(7),
			ms.WriteInt(fixedPoint1_8(x)),
			ms.WriteInt(fixedPoint1_8(y)),
			ms.WriteInt(fixedPoint1_8(z)),
			ms.WriteByte(64),
			ms.WriteByte(0xF0),
			ms.WriteBoolean(true),
		)
	})

	if fixedPoint1_8(y) != -48 || fixedPoint1_8(z) != -1 {
		t.Fatalf("the test's own rounding is off: %d and %d, want -48 and -1", fixedPoint1_8(y), fixedPoint1_8(z))
	}

	if got := runTransformer(t, DowngradeEntityPositionSyncTo1_8, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.8 = % x, want % x", got, want)
	}

	// A coordinate whose thirty-seconds no int holds, and a teleport cut
	// short, are refused.
	farOut := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(7),
			ms.WriteDouble(1<<30),
			ms.WriteDouble(0),
			ms.WriteDouble(0),
			ms.WriteByte(0),
			ms.WriteByte(0),
			ms.WriteBoolean(false),
		)
	})

	if err := failingTransformer(t, DowngradeEntityPositionSyncTo1_8, farOut); err == nil {
		t.Error("expected a coordinate no int holds to be refused")
	}

	if err := failingTransformer(t, DowngradeEntityPositionSyncTo1_8, sent[:len(sent)-1]); err == nil {
		t.Error("expected a teleport with no on ground flag to be refused")
	}
}

// entityData1_9 lays a run of entity metadata out as 1.9 reads it, holding
// the flag byte alone, or nothing when it has no flags to hold.
func entityData1_9(flags byte, withFlags bool) []byte {
	if !withFlags {
		return []byte{0xFF}
	}

	return []byte{0x00, 0x00, flags, 0xFF}
}

// entityData1_8 lays the same run out as 1.8 reads it.
func entityData1_8(flags byte, withFlags bool) []byte {
	if !withFlags {
		return []byte{0x7F}
	}

	return []byte{0x00, flags, 0x7F}
}

var uuidBytes = bytes.Repeat([]byte{0xAB}, 16)

// 1.8 reads the spawn player 1.9 does with the position as three ints of
// thirty-seconds of a block, a held item of none in front of the metadata,
// and the metadata in its own form, with or without the flag byte.
func TestDowngradeAddEntityTo1_8(t *testing.T) {
	for _, withFlags := range []bool{true, false} {
		sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(
				ms.WriteVarInt(300),
				ms.WriteBytes(uuidBytes),
				ms.WriteDouble(0.5),
				ms.WriteDouble(64),
				ms.WriteDouble(0.5),
				ms.WriteByte(10),
				ms.WriteByte(20),
				ms.WriteBytes(entityData1_9(0x02, withFlags)),
			)
		})

		want := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(
				ms.WriteVarInt(300),
				ms.WriteBytes(uuidBytes),
				ms.WriteInt(16),
				ms.WriteInt(64*32),
				ms.WriteInt(16),
				ms.WriteByte(10),
				ms.WriteByte(20),
				ms.WriteShort(0),
				ms.WriteBytes(entityData1_8(0x02, withFlags)),
			)
		})

		if got := runTransformer(t, DowngradeAddEntityTo1_8, sent); !bytes.Equal(got, want) {
			t.Errorf("with flags %t: to 1.8 = % x, want % x", withFlags, got, want)
		}
	}
}

// 1.8 reads the entity metadata packet 1.9 does with the metadata in its
// own form: the flag byte under a kind and index byte of zero, and the run
// ended with 0x7F. An entry other than the flag byte, which the two versions
// file differently, is refused, as is a run that does not end.
func TestDowngradeSetEntityDataTo1_8(t *testing.T) {
	sent := append([]byte{0x07}, entityData1_9(0x0A, true)...)
	want := append([]byte{0x07}, entityData1_8(0x0A, true)...)

	if got := runTransformer(t, DowngradeSetEntityDataTo1_8, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.8 = % x, want % x", got, want)
	}

	refused := map[string][]byte{
		"another index":      {0x07, 0x01, 0x00, 0x00, 0xFF},
		"another serializer": {0x07, 0x00, 0x01, 0x00, 0xFF},
		"no end":             {0x07, 0x00, 0x00, 0x00},
		"bytes past the end": {0x07, 0x00, 0x00, 0x00, 0xFF, 0x00},
	}

	for name, body := range refused {
		if err := failingTransformer(t, DowngradeSetEntityDataTo1_8, body); err == nil {
			t.Errorf("%s: expected the metadata to be refused", name)
		}
	}
}

// chunk1_9 lays a chunk packet out as 1.9 reads it, around the section
// buffer given: the coordinates, whether the chunk is whole, the mask and
// the bytes, with nothing behind them.
func chunk1_9(t *testing.T, whole bool, mask int32, sections []byte) []byte {
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

// spanLongs packs values at width bits each end to end across the longs,
// the way the versions before 1.16 pack a container, and returns the longs.
func spanLongs(values []int32, width int) []int64 {
	longs := make([]int64, (len(values)*width+63)/64)

	for i, value := range values {
		at := i * width
		long, shift := at/64, at%64

		longs[long] |= int64(uint64(value) << shift)
		if shift+width > 64 {
			longs[long+1] |= int64(uint64(value) >> (64 - shift))
		}
	}

	return longs
}

// blocks1_8 lays 4096 block numbers out as a 1.8 section carries them:
// little-endian shorts.
func blocks1_8(ids []int32) []byte {
	out := make([]byte, 0, 2*len(ids))
	for _, id := range ids {
		out = append(out, byte(id), byte(id>>8))
	}

	return out
}

// bulk1_8 lays a map chunk bulk of one chunk out as 1.8 reads it: the sky
// light flag, the count, the coordinates, the mask, and the bytes.
func bulk1_8(t *testing.T, mask int16, data []byte) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteBoolean(true),
			ms.WriteVarInt(1),
			ms.WriteInt(3),
			ms.WriteInt(-4),
			ms.WriteShort(mask),
			ms.WriteBytes(data),
		)
	})
}

// 1.8 reads a chunk as a map chunk bulk of one: every section's blocks as
// little-endian shorts of the number 1.9 keeps in its palette, then every
// section's block light, then every section's sky light, then the biomes.
// A section packing its ids directly is read at the width it says, behind
// the palette of none 1.9 reads there.
func TestDowngradeLevelChunkWithSectionLightTo1_8(t *testing.T) {
	// A section of a four-bit palette, with stone at block 0 and dirt at
	// block 4095, and a section of ids packed directly at sixteen bits,
	// with a state near the top of a short at block 1.
	paletted := make([]int32, 4096)
	paletted[0], paletted[4095] = 1, 2

	direct := make([]int32, 4096)
	direct[1] = 0xFFF0

	sections := append(
		section1_13(4, []int32{0, 16, 48}, spanLongs(paletted, 4), false),
		section1_13(16, nil, spanLongs(direct, 16), true)...,
	)

	biomes := bytes.Repeat([]byte{0x01}, 256)

	sent := chunk1_9(t, true, 1<<0|1<<5, append(sections, biomes...))

	palettedIds := make([]int32, 4096)
	palettedIds[0], palettedIds[4095] = 16, 48

	var data []byte
	data = append(data, blocks1_8(palettedIds)...)
	data = append(data, blocks1_8(direct)...)
	data = append(data, bytes.Repeat([]byte{0x0F}, 2048)...)
	data = append(data, bytes.Repeat([]byte{0x0F}, 2048)...)
	data = append(data, bytes.Repeat([]byte{0xF0}, 2048)...)
	data = append(data, bytes.Repeat([]byte{0xF0}, 2048)...)
	data = append(data, biomes...)

	want := bulk1_8(t, 1<<0|1<<5, data)

	if got := runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_8, sent); !bytes.Equal(got, want) {
		t.Errorf("to 1.8 = %d bytes, want %d; first difference at %d", len(got), len(want), firstDifference(got, want))
	}

	// A chunk with no section in it is a bulk of one chunk with nothing but
	// its biomes, which a 1.8 client loads all the same.
	empty := chunk1_9(t, true, 0, biomes)

	if got, want := runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_8, empty), bulk1_8(t, 0, biomes); !bytes.Equal(got, want) {
		t.Errorf("an empty chunk to 1.8 = % x, want % x", got, want)
	}
}

// A chunk that is not whole, a mask past sixteen sections, a state no short
// holds, a palette index past the palette, a section cut short and bytes
// past the biomes are refused.
func TestDowngradeLevelChunkWithSectionLightTo1_8Refuses(t *testing.T) {
	air := make([]int32, 4096)
	biomes := bytes.Repeat([]byte{0x01}, 256)

	tooWide := make([]int32, 4096)
	tooWide[0] = 0x10000

	outOfPalette := make([]int32, 4096)
	outOfPalette[0] = 3

	section := section1_13(4, []int32{0, 16, 48}, spanLongs(air, 4), false)

	refused := map[string][]byte{
		"not whole":         chunk1_9(t, false, 1, append(section, biomes...)),
		"mask past sixteen": chunk1_9(t, true, 1<<16, append(section, biomes...)),
		"state past a short": chunk1_9(t, true, 1, append(
			section1_13(17, nil, spanLongs(tooWide, 17), true), biomes...)),
		"index past the palette": chunk1_9(t, true, 1, append(
			section1_13(4, []int32{0, 16, 48}, spanLongs(outOfPalette, 4), false), biomes...)),
		"section cut short":     chunk1_9(t, true, 1, append(section[:len(section)-1], biomes...)),
		"bytes past the biomes": chunk1_9(t, true, 1, append(append(section, biomes...), 0x00)),
		"no biomes":             chunk1_9(t, true, 0, nil),
	}

	for name, body := range refused {
		if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_8, body); err == nil {
			t.Errorf("%s: expected the chunk to be refused", name)
		}
	}
}

func firstDifference(a, b []byte) int {
	for i := range min(len(a), len(b)) {
		if a[i] != b[i] {
			return i
		}
	}

	return min(len(a), len(b))
}

// 1.9 reads the hand that swung off a swing 1.8 sends with nothing in it:
// the main hand, the one 1.8 has. A swing with something in it is refused.
func TestUpgradePunchFrom1_8(t *testing.T) {
	if got := runTransformer(t, UpgradePunchFrom1_8, nil); !bytes.Equal(got, []byte{0x00}) {
		t.Errorf("from 1.8 = % x, want the main hand", got)
	}

	if err := failingTransformer(t, UpgradePunchFrom1_8, []byte{0x00}); err == nil {
		t.Error("expected a swing with a hand in it to be refused")
	}
}

// 1.9 reads the player command 1.8 sends with the actions the two number
// alike as they are, and opening the inventory one higher; an action 1.8
// has no number for is refused.
func TestUpgradePlayerCommandFrom1_8(t *testing.T) {
	command := func(action int32) []byte {
		return encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(ms.WriteVarInt(7), ms.WriteVarInt(action), ms.WriteVarInt(0))
		})
	}

	for action := int32(0); action <= 5; action++ {
		if got := runTransformer(t, UpgradePlayerCommandFrom1_8, command(action)); !bytes.Equal(got, command(action)) {
			t.Errorf("action %d: from 1.8 = % x, want it as sent", action, got)
		}
	}

	if got := runTransformer(t, UpgradePlayerCommandFrom1_8, command(6)); !bytes.Equal(got, command(7)) {
		t.Errorf("opening the inventory: from 1.8 = % x, want action 7", got)
	}

	for _, action := range []int32{7, 8, -1} {
		if err := failingTransformer(t, UpgradePlayerCommandFrom1_8, command(action)); err == nil {
			t.Errorf("action %d: expected an action 1.8 has no number for to be refused", action)
		}
	}

	if err := failingTransformer(t, UpgradePlayerCommandFrom1_8, append(command(3), 0x00)); err == nil {
		t.Error("expected a command with bytes past its jump strength to be refused")
	}
}
