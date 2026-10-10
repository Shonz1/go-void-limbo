package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
)

// spawnPlayer1_7_6 lays the spawn player out as 1.7.6 reads it: the entity
// id, the uuid as text, the name, the given properties as a name, a value
// and a signature each behind their count, the position as three ints, the
// rotation, the held item and the metadata.
func spawnPlayer1_7_6(t *testing.T, properties [][3]string) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteVarInt(300),
			ms.WriteString(uuidText1_7_6),
			ms.WriteString("Notch"),
			ms.WriteVarInt(int32(len(properties))),
		}

		for _, property := range properties {
			for _, field := range property {
				steps = append(steps, ms.WriteString(field))
			}
		}

		return writeAll(append(steps,
			ms.WriteInt(16),
			ms.WriteInt(64*32),
			ms.WriteInt(16),
			ms.WriteByte(10),
			ms.WriteByte(20),
			ms.WriteShort(0),
			ms.WriteBytes(entityData1_8(0x02, true)),
		)...)
	})
}

// spawnPlayer1_7_2 lays the spawn player out as 1.7.2 reads it, which is
// 1.7.6's with no properties and no count of them.
func spawnPlayer1_7_2(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(300),
			ms.WriteString(uuidText1_7_6),
			ms.WriteString("Notch"),
			ms.WriteInt(16),
			ms.WriteInt(64*32),
			ms.WriteInt(16),
			ms.WriteByte(10),
			ms.WriteByte(20),
			ms.WriteShort(0),
			ms.WriteBytes(entityData1_8(0x02, true)),
		)
	})
}

// 1.7.2 reads the spawn player 1.7.6 does with the properties taken out,
// whether there were two, one unsigned, or none.
func TestDowngradeAddEntityTo1_7_2(t *testing.T) {
	want := spawnPlayer1_7_2(t)

	for name, properties := range map[string][][3]string{
		"two properties": {{"textures", "value", "signature"}, {"other", "plain", ""}},
		"one property":   {{"textures", "value", "signature"}},
		"no property":    nil,
	} {
		if got := runTransformer(t, DowngradeAddEntityTo1_7_2, spawnPlayer1_7_6(t, properties)); !bytes.Equal(got, want) {
			t.Errorf("%s: to 1.7.2 = % x, want % x", name, got, want)
		}
	}
}

// A count below zero, a property cut short and a spawn cut short are
// refused.
func TestDowngradeAddEntityTo1_7_2Refuses(t *testing.T) {
	negative := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(300),
			ms.WriteString(uuidText1_7_6),
			ms.WriteString("Notch"),
			ms.WriteVarInt(-1),
		)
	})

	if err := failingTransformer(t, DowngradeAddEntityTo1_7_2, negative); err == nil {
		t.Error("expected a count below zero to be refused")
	}

	cutProperty := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(300),
			ms.WriteString(uuidText1_7_6),
			ms.WriteString("Notch"),
			ms.WriteVarInt(1),
			ms.WriteString("textures"),
		)
	})

	if err := failingTransformer(t, DowngradeAddEntityTo1_7_2, cutProperty); err == nil {
		t.Error("expected a property cut short to be refused")
	}

	sent := spawnPlayer1_7_6(t, nil)
	if err := failingTransformer(t, DowngradeAddEntityTo1_7_2, sent[:3]); err == nil {
		t.Error("expected a spawn cut short in the uuid to be refused")
	}
}
