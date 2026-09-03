package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// bottomRegistryCodec and bottomDimensionType stand in for what package
// gamedata hands the transformer for 1.18: the registries and the dimension
// type, told apart from 1.18.2's by their content.
func bottomRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:worldgen/biome": nbt.Compound{"type": nbt.String("minecraft:worldgen/biome"), "value": nbt.List{ElementType: nbt.TagCompound}}})
	})
}

func bottomDimensionType(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"infiniburn": nbt.String("minecraft:infiniburn_overworld")})
	})
}

// The login 1.18 reads is the login 1.18.2 reads with the registries and
// the spelled-out dimension type in the middle swapped for 1.18's own.
// Everything else is copied.
func TestDowngradePlayLoginTo1_18SwapsTheRegistriesAndTheDimensionType(t *testing.T) {
	newerCodec, newerDimensionType := lowestRegistryCodec(t), lowestDimensionType(t)
	olderCodec, olderDimensionType := bottomRegistryCodec(t), bottomDimensionType(t)

	sent := playLogin1_18_2(t, newerCodec, newerDimensionType)
	got := runTransformer(t, DowngradePlayLoginTo1_18(olderCodec, olderDimensionType), sent)
	want := playLogin1_18_2(t, olderCodec, olderDimensionType)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.18 = % x\nwant = % x", got, want)
	}

	if bytes.Contains(got, newerCodec) {
		t.Error("the 1.18 login still carries 1.18.2's registries")
	}

	if bytes.Contains(got, newerDimensionType) {
		t.Error("the 1.18 login still spells out 1.18.2's dimension type")
	}
}

// The whole chain from 1.19 down: a 1.19 login walked to 1.18.2 and then to
// 1.18 carries 1.18's registries and dimension type, and neither of the
// other two versions'.
func TestDowngradePlayLoginTo1_18FollowsTheStepAbove(t *testing.T) {
	codec1_19, codec1_18_2, codec1_18 := earliestRegistryCodec(t), lowestRegistryCodec(t), bottomRegistryCodec(t)
	dimensionType1_18_2, dimensionType1_18 := lowestDimensionType(t), bottomDimensionType(t)

	sent := playLogin1_19_4(t, codec1_19, true)
	at1_18_2 := runTransformer(t, DowngradePlayLoginTo1_18_2(codec1_18_2, dimensionType1_18_2), sent)
	got := runTransformer(t, DowngradePlayLoginTo1_18(codec1_18, dimensionType1_18), at1_18_2)
	want := playLogin1_18_2(t, codec1_18, dimensionType1_18)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.18 = % x\nwant = % x", got, want)
	}

	for name, other := range map[string][]byte{"1.19's registries": codec1_19, "1.18.2's registries": codec1_18_2, "1.18.2's dimension type": dimensionType1_18_2} {
		if bytes.Contains(got, other) {
			t.Errorf("the 1.18 login still carries %s", name)
		}
	}
}

// A 1.18 client reads the registries and the dimension type out of this
// packet and nothing else, so a transformer with either missing refuses the
// login rather than send a client into a world it cannot make sense of.
func TestDowngradePlayLoginTo1_18RefusesToSendNoRegistries(t *testing.T) {
	sent := playLogin1_18_2(t, lowestRegistryCodec(t), lowestDimensionType(t))

	if err := failingTransformer(t, DowngradePlayLoginTo1_18(nil, bottomDimensionType(t)), sent); err == nil {
		t.Error("error = nil, want a refusal for a login with no registries to carry")
	}

	if err := failingTransformer(t, DowngradePlayLoginTo1_18(bottomRegistryCodec(t), nil), sent); err == nil {
		t.Error("error = nil, want a refusal for a login with no dimension type to spell out")
	}
}

// A login cut short of its dimension type is refused rather than written
// with 1.18's own on the end of nothing.
func TestDowngradePlayLoginTo1_18RefusesWhatItCannotWalk(t *testing.T) {
	sent := playLogin1_18_2(t, lowestRegistryCodec(t), lowestDimensionType(t))
	truncated := sent[:len(sent)-len(lowestDimensionType(t))-len("minecraft:the_end")-1-8-3-4]

	if err := failingTransformer(t, DowngradePlayLoginTo1_18(bottomRegistryCodec(t), bottomDimensionType(t)), truncated); err == nil {
		t.Error("error = nil, want a refusal for a login cut short of its dimension type")
	}
}
