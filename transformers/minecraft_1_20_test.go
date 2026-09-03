package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// playLogin1_19_4 is the login as 1.19.4 lays it out: 1.20's from front to
// back, with the registries the transformer was handed in the middle and
// nothing after the death location, where 1.20 puts the portal cooldown.
func playLogin1_19_4(t *testing.T, registryCodec []byte, deathLocation bool) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteBoolean(true),
			ms.WriteByte(2),
			ms.WriteByte(0xFF),
			ms.WriteVarInt(2),
			ms.WriteString("minecraft:overworld"),
			ms.WriteString("minecraft:the_end"),
			ms.WriteBytes(registryCodec),
			ms.WriteString("minecraft:overworld"),
			ms.WriteString("minecraft:the_end"),
			ms.WriteLong(0x1122334455667788),
			ms.WriteVarInt(20),
			ms.WriteVarInt(8),
			ms.WriteVarInt(2),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
			ms.WriteBoolean(deathLocation),
		}

		if deathLocation {
			steps = append(steps, ms.WriteString("minecraft:the_nether"), ms.WriteLong(0x0102030405060708))
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// olderRegistryCodec stands in for what package gamedata hands the
// transformer for 1.19.4, told apart from 1.20's by its content.
func olderRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:damage_type": nbt.Compound{"type": nbt.String("minecraft:damage_type")}})
	})
}

// The login 1.20 reads is the login 1.19.4 reads with two things changed:
// the registries in the middle are 1.19.4's own rather than 1.20's, and the
// portal cooldown at the end is gone. Everything else, on both sides of the
// registries and with or without a death location, is copied.
func TestDowngradePlayLoginTo1_19_4SwapsTheRegistriesAndDropsTheCooldown(t *testing.T) {
	newer, older := testRegistryCodec(t), olderRegistryCodec(t)

	for _, deathLocation := range []bool{false, true} {
		got := runTransformer(t, DowngradePlayLoginTo1_19_4(older), playLogin1_20(t, newer, deathLocation))
		want := playLogin1_19_4(t, older, deathLocation)

		if !bytes.Equal(got, want) {
			t.Errorf("death location %t: to 1.19.4 = % x\nwant = % x", deathLocation, got, want)
		}

		if bytes.Contains(got, newer) {
			t.Errorf("death location %t: the 1.19.4 login still carries 1.20's registries", deathLocation)
		}
	}
}

// A 1.19.4 client reads the registries out of this packet and nothing else,
// so a transformer with none to write refuses the login rather than send a
// client into a world it cannot make sense of.
func TestDowngradePlayLoginTo1_19_4RefusesToSendNoRegistries(t *testing.T) {
	if err := failingTransformer(t, DowngradePlayLoginTo1_19_4(nil), playLogin1_20(t, testRegistryCodec(t), false)); err == nil {
		t.Error("error = nil, want a refusal for a login with no registries to carry")
	}
}

// The chunk packet 1.19.4 reads is the one 1.20 reads with the trust edges
// flag put back in front of the light data: after the coordinates, the named
// heightmaps, the sections and the block entities, and before the four
// masks. The flag is set, as a vanilla server of that version sets it.
func TestDowngradeLevelChunkWithLightTo1_19_4PutsTheTrustEdgesFlagInFrontOfTheLight(t *testing.T) {
	heightmaps := nbt.Compound{"MOTION_BLOCKING": nbt.LongArray{1, 2, 3}, "WORLD_SURFACE": nbt.LongArray{4}}
	sections := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	light := []byte{0x01, 0x02, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00}

	for _, blockEntities := range []bool{false, true} {
		head := encodeBody(t, func(ms *streams.MinecraftStream) error {
			steps := []error{
				ms.WriteInt(-3),
				ms.WriteInt(9),
				nbt.WriteNamed(ms, "", heightmaps),
				ms.WriteByteArray(sections),
			}

			if blockEntities {
				steps = append(steps,
					ms.WriteVarInt(1),
					ms.WriteByte(0x35),
					ms.WriteShort(64),
					ms.WriteVarInt(7),
					nbt.WriteNamed(ms, "", nbt.Compound{"id": nbt.String("minecraft:chest")}),
				)
			} else {
				steps = append(steps, ms.WriteVarInt(0))
			}

			for _, err := range steps {
				if err != nil {
					return err
				}
			}

			return nil
		})

		sent := append(append([]byte{}, head...), light...)
		got := runTransformer(t, DowngradeLevelChunkWithLightTo1_19_4, sent)
		want := append(append(append([]byte{}, head...), 0x01), light...)

		if !bytes.Equal(got, want) {
			t.Errorf("block entities %t: to 1.19.4 = % x\nwant = % x", blockEntities, got, want)
		}

		// The flag is one byte, and that is the whole of the difference.
		if len(got) != len(sent)+1 {
			t.Errorf("block entities %t: the body is %d bytes, want the %d sent plus the flag", blockEntities, len(got), len(sent))
		}
	}
}
