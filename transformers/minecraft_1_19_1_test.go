package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// earliestRegistryCodec stands in for what package gamedata hands the
// transformer for 1.19, told apart from 1.19.1's and the others' by its
// content.
func earliestRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:worldgen/biome": nbt.Compound{"type": nbt.String("minecraft:worldgen/biome")}})
	})
}

// The login 1.19 reads is the login 1.19.1 reads with one thing changed:
// the registries in the middle are 1.19's own rather than 1.19.1's.
// Everything else, on both sides of the registries and with or without a
// death location, is copied, and nothing comes off the end.
func TestDowngradePlayLoginTo1_19SwapsTheRegistries(t *testing.T) {
	newer, older := firstRegistryCodec(t), earliestRegistryCodec(t)

	for _, deathLocation := range []bool{false, true} {
		sent := playLogin1_19_4(t, newer, deathLocation)
		got := runTransformer(t, DowngradePlayLoginTo1_19(older), sent)
		want := playLogin1_19_4(t, older, deathLocation)

		if !bytes.Equal(got, want) {
			t.Errorf("death location %t: to 1.19 = % x\nwant = % x", deathLocation, got, want)
		}

		if bytes.Contains(got, newer) {
			t.Errorf("death location %t: the 1.19 login still carries 1.19.1's registries", deathLocation)
		}

		// The registries are the whole of the difference.
		if len(got)-len(older) != len(sent)-len(newer) {
			t.Errorf("death location %t: the body is %d bytes around the registries, want the %d sent", deathLocation, len(got)-len(older), len(sent)-len(newer))
		}
	}
}

// A 1.19 client reads the registries out of this packet and nothing else,
// so a transformer with none to write refuses the login rather than send a
// client into a world it cannot make sense of.
func TestDowngradePlayLoginTo1_19RefusesToSendNoRegistries(t *testing.T) {
	if err := failingTransformer(t, DowngradePlayLoginTo1_19(nil), playLogin1_19_4(t, firstRegistryCodec(t), false)); err == nil {
		t.Error("error = nil, want a refusal for a login with no registries to carry")
	}
}

// A 1.19 hello is the name and the optional profile key, and 1.19.1 reads
// an optional uuid behind the two: the body goes across as it is, key
// included, with the uuid absent on the end. What comes out is what the
// 1.19.3 step reads from a 1.19.1 client that sent no uuid, and the key
// comes off there.
func TestUpgradeLoginStartFrom1_19AppendsNoUuid(t *testing.T) {
	profileKey := func(ms *streams.MinecraftStream) error {
		if err := ms.WriteBoolean(true); err != nil {
			return err
		}

		// The expiry, the key and Mojang's signature over the two.
		if err := ms.WriteLong(1700000000000); err != nil {
			return err
		}

		if err := ms.WriteByteArray([]byte("an encoded rsa public key")); err != nil {
			return err
		}

		return ms.WriteByteArray([]byte("signed by mojang"))
	}

	tests := []struct {
		name string
		sent []byte
	}{
		{
			name: "with a key",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				return profileKey(ms)
			}),
		},
		{
			name: "without a key",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				return ms.WriteBoolean(false)
			}),
		},
	}

	// What every case comes out of the 1.19.3 step as: the name and no uuid.
	nameAlone := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteString("Notch"); err != nil {
			return err
		}

		return ms.WriteBoolean(false)
	})

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := runTransformer(t, UpgradeLoginStartFrom1_19, test.sent)
			want := append(append([]byte{}, test.sent...), 0x00)

			if !bytes.Equal(got, want) {
				t.Errorf("to 1.19.1 = % x\nwant = % x", got, want)
			}

			// And the 1.19.3 step reads what this one wrote, as it would from
			// a 1.19.1 client, and takes the key off; the 1.20 step after it.
			upgraded := runTransformer(t, UpgradeLoginStartFrom1_19_1, got)
			if !bytes.Equal(upgraded, nameAlone) {
				t.Errorf("to 1.19.3 = % x\nwant = % x", upgraded, nameAlone)
			}

			runTransformer(t, UpgradeLoginStartFrom1_20, upgraded)
		})
	}
}
