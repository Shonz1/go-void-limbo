package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

// oldestRegistryCodec stands in for what package gamedata hands the
// transformer for 1.19.3, told apart from 1.19.4's and 1.20's by its content.
func oldestRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:chat_type": nbt.Compound{"type": nbt.String("minecraft:chat_type")}})
	})
}

// The login 1.19.3 reads is the login 1.19.4 reads with one thing changed:
// the registries in the middle are 1.19.3's own rather than 1.19.4's.
// Everything else, on both sides of the registries and with or without a
// death location, is copied, and nothing comes off the end.
func TestDowngradePlayLoginTo1_19_3SwapsTheRegistries(t *testing.T) {
	newer, older := olderRegistryCodec(t), oldestRegistryCodec(t)

	for _, deathLocation := range []bool{false, true} {
		sent := playLogin1_19_4(t, newer, deathLocation)
		got := runTransformer(t, DowngradePlayLoginTo1_19_3(older), sent)
		want := playLogin1_19_4(t, older, deathLocation)

		if !bytes.Equal(got, want) {
			t.Errorf("death location %t: to 1.19.3 = % x\nwant = % x", deathLocation, got, want)
		}

		if bytes.Contains(got, newer) {
			t.Errorf("death location %t: the 1.19.3 login still carries 1.19.4's registries", deathLocation)
		}

		// The registries are the whole of the difference.
		if len(got)-len(older) != len(sent)-len(newer) {
			t.Errorf("death location %t: the body is %d bytes around the registries, want the %d sent", deathLocation, len(got)-len(older), len(sent)-len(newer))
		}
	}
}

// A 1.19.3 client reads the registries out of this packet and nothing else,
// so a transformer with none to write refuses the login rather than send a
// client into a world it cannot make sense of.
func TestDowngradePlayLoginTo1_19_3RefusesToSendNoRegistries(t *testing.T) {
	if err := failingTransformer(t, DowngradePlayLoginTo1_19_3(nil), playLogin1_19_4(t, olderRegistryCodec(t), false)); err == nil {
		t.Error("error = nil, want a refusal for a login with no registries to carry")
	}
}

// 1.19.3 reads the position, the rotation, the byte of flags and the
// teleport id as 1.19.4 does, and then one more thing: the dismount vehicle
// flag, which is clear.
func TestDowngradePlayerPositionTo1_19_3AppendsTheDismountFlag(t *testing.T) {
	body := encodePlayerPosition(t, &play.PlayerPositionClientboundPacket{
		TeleportId: 300,
		X:          0.5,
		Y:          64,
		Z:          -0.5,
		Yaw:        90,
		Pitch:      -45,
		Relatives:  0x18, // yaw and pitch relative
	})

	to1_21 := runTransformer(t, DowngradePlayerPositionTo1_21, body)
	got := runTransformer(t, DowngradePlayerPositionTo1_19_3, to1_21)

	want := []byte{
		0x3F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x, 0.5
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y, 64
		0xBF, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z, -0.5
		0x42, 0xB4, 0x00, 0x00, // yaw, 90
		0xC2, 0x34, 0x00, 0x00, // pitch, -45
		0x18,       // the flags, one byte
		0xAC, 0x02, // the teleport id, 300
		0x00, // dismount vehicle, clear
	}

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.19.3 = % x\nwant = % x", got, want)
	}
}

// The pose serializer is 20 at the latest version, 21 from 1.21.7 down to
// 1.20.5, 20 again from 1.20.3 down to 1.19.4, and 19 on 1.19.3. Walking a
// packet down through the 1.21.9, 1.20.5 and 1.19.4 steps lands it there,
// with the flags entry, whose byte serializer sits at 0 throughout,
// untouched.
func TestDowngradeSetEntityDataTo1_19_3RenamesThePoseSerializer(t *testing.T) {
	stance := &play.SetEntityDataClientboundPacket{EntityId: 2, Sneaking: true, Sprinting: true}

	encoded := encodeAddEntityData(t, stance)
	to1_21_7 := runTransformer(t, DowngradeSetEntityDataTo1_21_7, encoded)
	to1_19_4 := runTransformer(t, DowngradeSetEntityDataTo1_20_3, to1_21_7)

	got := runTransformer(t, DowngradeSetEntityDataTo1_19_3, to1_19_4)

	if len(got) != len(to1_19_4) {
		t.Fatalf("to 1.19.3 is %d bytes, want the %d sent: only a serializer id changes", len(got), len(to1_19_4))
	}

	for i := range got {
		if i == 5 {
			continue
		}

		if got[i] != to1_19_4[i] {
			t.Errorf("byte %d = %#x, want %#x untouched", i, got[i], to1_19_4[i])
		}
	}

	if to1_19_4[5] != poseSerializer1_19_4 {
		t.Fatalf("pose serializer sent = %d, want %d on 1.19.4", to1_19_4[5], poseSerializer1_19_4)
	}

	if got[5] != poseSerializer1_19_3 {
		t.Errorf("pose serializer = %d, want %d", got[5], poseSerializer1_19_3)
	}
}
