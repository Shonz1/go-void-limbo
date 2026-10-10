package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
)

// playLogin1_9_1 lays the login out as 1.9.1 reads it: the entity id, the
// game mode, the dimension as an int, the difficulty, the most players, the
// level type and the flag for the debug screen.
func playLogin1_9_1(t *testing.T, dimension int32) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteInt(7),
			ms.WriteByte(2|8),
			ms.WriteInt(dimension),
			ms.WriteByte(0),
			ms.WriteByte(20),
			ms.WriteString("flat"),
			ms.WriteBoolean(true),
		)
	})
}

// playLogin1_9 lays the login out as 1.9 reads it, which is 1.9.1's with the
// dimension as a byte.
func playLogin1_9(t *testing.T, dimension int8) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteInt(7),
			ms.WriteByte(2|8),
			ms.WriteByte(byte(dimension)),
			ms.WriteByte(0),
			ms.WriteByte(20),
			ms.WriteString("flat"),
			ms.WriteBoolean(true),
		)
	})
}

// 1.9 reads the login 1.9.1 does with the dimension as a byte, the nether's
// minus one included.
func TestDowngradePlayLoginTo1_9(t *testing.T) {
	for _, dimension := range []int8{0, -1, 1} {
		want := playLogin1_9(t, dimension)

		if got := runTransformer(t, DowngradePlayLoginTo1_9, playLogin1_9_1(t, int32(dimension))); !bytes.Equal(got, want) {
			t.Errorf("dimension %d: to 1.9 = % x, want % x", dimension, got, want)
		}
	}
}

// A dimension no byte holds is refused rather than cut down, and so is a
// login cut short.
func TestDowngradePlayLoginTo1_9Refuses(t *testing.T) {
	for _, dimension := range []int32{128, -129, 1 << 20} {
		if err := failingTransformer(t, DowngradePlayLoginTo1_9, playLogin1_9_1(t, dimension)); err == nil {
			t.Errorf("dimension %d: expected a dimension no byte holds to be refused", dimension)
		}
	}

	sent := playLogin1_9_1(t, 0)
	if err := failingTransformer(t, DowngradePlayLoginTo1_9, sent[:len(sent)-1]); err == nil {
		t.Error("expected a login with no debug screen flag to be refused")
	}
}
