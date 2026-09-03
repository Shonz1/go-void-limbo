package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// lowestRegistryCodec and lowestDimensionType stand in for what package
// gamedata hands the transformer for 1.18.2: the registries, told apart
// from 1.19's and the others' by their content, and the dimension type the
// login spells out.
func lowestRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:dimension_type": nbt.Compound{"type": nbt.String("minecraft:dimension_type"), "value": nbt.List{ElementType: nbt.TagCompound}}})
	})
}

func lowestDimensionType(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"min_y": nbt.Int(-64), "height": nbt.Int(384)})
	})
}

// playLogin1_18_2 is the login as 1.18.2 lays it out: 1.19's from the front
// up to the registries, the dimension type spelled out where 1.19 names it,
// and nothing after the flat flag, where 1.19 puts the death location.
func playLogin1_18_2(t *testing.T, registryCodec []byte, dimensionType []byte) []byte {
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
			ms.WriteBytes(dimensionType),
			ms.WriteString("minecraft:the_end"),
			ms.WriteLong(0x1122334455667788),
			ms.WriteVarInt(20),
			ms.WriteVarInt(8),
			ms.WriteVarInt(2),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
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

// The login 1.18.2 reads is the login 1.19 reads with three things changed:
// the registries in the middle are 1.18.2's own rather than 1.19's, the
// dimension type behind them is spelled out where 1.19 names it, and the
// death location on the end is gone, absent or present. Everything else is
// copied.
func TestDowngradePlayLoginTo1_18_2SpellsOutTheDimensionType(t *testing.T) {
	newer, older, dimensionType := earliestRegistryCodec(t), lowestRegistryCodec(t), lowestDimensionType(t)

	for _, deathLocation := range []bool{false, true} {
		sent := playLogin1_19_4(t, newer, deathLocation)
		got := runTransformer(t, DowngradePlayLoginTo1_18_2(older, dimensionType), sent)
		want := playLogin1_18_2(t, older, dimensionType)

		if !bytes.Equal(got, want) {
			t.Errorf("death location %t: to 1.18.2 = % x\nwant = % x", deathLocation, got, want)
		}

		if bytes.Contains(got, newer) {
			t.Errorf("death location %t: the 1.18.2 login still carries 1.19's registries", deathLocation)
		}
	}
}

// The one dimension type this server ever names is the overworld, first
// among the entries it sends, and the one entry the transformer is handed.
// A login naming another is refused rather than sent with the wrong world
// spelled out.
func TestDowngradePlayLoginTo1_18_2RefusesADimensionTypeItCannotSpellOut(t *testing.T) {
	registryCodec := earliestRegistryCodec(t)

	// The login as 1.19 lays it out, naming the nether's dimension type.
	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteBoolean(false),
			ms.WriteByte(0),
			ms.WriteByte(0xFF),
			ms.WriteVarInt(1),
			ms.WriteString("minecraft:the_nether"),
			ms.WriteBytes(registryCodec),
			ms.WriteString("minecraft:the_nether"),
			ms.WriteString("minecraft:the_nether"),
			ms.WriteLong(0),
			ms.WriteVarInt(20),
			ms.WriteVarInt(8),
			ms.WriteVarInt(2),
			ms.WriteBoolean(false),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
			ms.WriteBoolean(false),
			ms.WriteBoolean(false),
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err := failingTransformer(t, DowngradePlayLoginTo1_18_2(lowestRegistryCodec(t), lowestDimensionType(t)), sent); err == nil {
		t.Error("error = nil, want a refusal for a login naming a dimension type the transformer has no definition of")
	}
}

// A 1.18.2 client reads the registries and the dimension type out of this
// packet and nothing else, so a transformer with either missing refuses the
// login rather than send a client into a world it cannot make sense of.
func TestDowngradePlayLoginTo1_18_2RefusesToSendNoRegistries(t *testing.T) {
	sent := playLogin1_19_4(t, earliestRegistryCodec(t), false)

	if err := failingTransformer(t, DowngradePlayLoginTo1_18_2(nil, lowestDimensionType(t)), sent); err == nil {
		t.Error("error = nil, want a refusal for a login with no registries to carry")
	}

	if err := failingTransformer(t, DowngradePlayLoginTo1_18_2(lowestRegistryCodec(t), nil), sent); err == nil {
		t.Error("error = nil, want a refusal for a login with no dimension type to spell out")
	}
}

// The entry this server sends about a joining player, walked down to 1.19:
// 1.19.1's add, which 1.19 reads alike, with the display name and the
// profile key absent on the end.
func playerInfoJoin1_19(t *testing.T, entry play.PlayerInfoEntry) []byte {
	t.Helper()

	return runTransformer(t, DowngradePlayerInfoUpdateTo1_19_1, playerInfoJoin1_19_3(t, entry))
}

// An add goes across as it is but for the profile key, which comes off the
// end of every entry.
func TestDowngradePlayerInfoUpdateTo1_18_2DropsTheProfileKey(t *testing.T) {
	signature := "signed"
	entry := play.PlayerInfoEntry{
		Profile: types.GameProfile{
			Uuid:     "01020304-0506-0708-090a-0b0c0d0e0f10",
			Username: "Notch",
			Properties: []types.ProfileProperty{
				{Name: "textures", Value: "dGV4dHVyZXM=", Signature: &signature},
			},
		},
		GameMode: types.GameModeCreative,
		Listed:   true,
		ShowHat:  true,
	}

	sent := playerInfoJoin1_19(t, entry)
	got := runTransformer(t, DowngradePlayerInfoUpdateTo1_18_2, sent)

	// The key is the one trailing byte, and everything in front of it is
	// as 1.19 laid it out.
	want := sent[:len(sent)-1]
	if !bytes.Equal(got, want) {
		t.Errorf("to 1.18.2 = % x\nwant = % x", got, want)
	}

	if sent[len(sent)-1] != 0x00 {
		t.Errorf("the 1.19 entry ends in % x, want the absent profile key this server sends", sent[len(sent)-1])
	}

	// Two players at once come off alike.
	two := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteVarInt(playerInfoAction1_19_1AddPlayer); err != nil {
			return err
		}

		if err := ms.WriteVarInt(2); err != nil {
			return err
		}

		// The entries as 1.19 lays them out, less the action and the count
		// in front of the first.
		for range 2 {
			if err := ms.WriteBytes(sent[2:]); err != nil {
				return err
			}
		}

		return nil
	})

	got = runTransformer(t, DowngradePlayerInfoUpdateTo1_18_2, two)

	want = encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteVarInt(playerInfoAction1_19_1AddPlayer); err != nil {
			return err
		}

		if err := ms.WriteVarInt(2); err != nil {
			return err
		}

		for range 2 {
			if err := ms.WriteBytes(sent[2 : len(sent)-1]); err != nil {
				return err
			}
		}

		return nil
	})

	if !bytes.Equal(got, want) {
		t.Errorf("two entries to 1.18.2 = % x\nwant = % x", got, want)
	}
}

// Every other action is laid out alike in the two and goes across as it is.
func TestDowngradePlayerInfoUpdateTo1_18_2LeavesTheOtherActionsAlone(t *testing.T) {
	uuid := "01020304-0506-0708-090a-0b0c0d0e0f10"

	for _, actions := range []play.PlayerInfoAction{play.PlayerInfoUpdateGameMode, play.PlayerInfoUpdateLatency} {
		sent := runTransformer(t, DowngradePlayerInfoUpdateTo1_19_1, encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{
			Actions: actions,
			Entries: []play.PlayerInfoEntry{{Profile: types.GameProfile{Uuid: uuid}, GameMode: types.GameModeSpectator, Latency: 42}},
		}))

		if got := runTransformer(t, DowngradePlayerInfoUpdateTo1_18_2, sent); !bytes.Equal(got, sent) {
			t.Errorf("%s to 1.18.2 = % x\nwant it untouched: % x", actions, got, sent)
		}
	}

	remove := runTransformer(t, DowngradePlayerInfoRemoveTo1_19_1, encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteVarInt(1); err != nil {
			return err
		}

		return ms.WriteUuid(uuid)
	}))

	if got := runTransformer(t, DowngradePlayerInfoUpdateTo1_18_2, remove); !bytes.Equal(got, remove) {
		t.Errorf("remove to 1.18.2 = % x\nwant it untouched: % x", got, remove)
	}
}

// A key this server never sends, and a display name it never sends either,
// are refused rather than walked.
func TestDowngradePlayerInfoUpdateTo1_18_2RefusesWhatItCannotWalk(t *testing.T) {
	entry := play.PlayerInfoEntry{Profile: types.GameProfile{Uuid: "01020304-0506-0708-090a-0b0c0d0e0f10", Username: "Notch"}}
	sent := playerInfoJoin1_19(t, entry)

	withKey := append(append([]byte{}, sent[:len(sent)-1]...), 0x01)
	if err := failingTransformer(t, DowngradePlayerInfoUpdateTo1_18_2, withKey); err == nil {
		t.Error("error = nil, want a refusal for an entry carrying a profile key")
	}

	withDisplayName := append(append([]byte{}, sent[:len(sent)-2]...), 0x01, 0x00)
	if err := failingTransformer(t, DowngradePlayerInfoUpdateTo1_18_2, withDisplayName); err == nil {
		t.Error("error = nil, want a refusal for an entry carrying a display name")
	}
}

// The login success 1.18.2 reads is the profile's uuid and name: the
// properties 1.19 put behind them come off, however many there are.
func TestDowngradeLoginSuccessTo1_18_2DropsTheProperties(t *testing.T) {
	signature := "signed"

	for _, profile := range []types.GameProfile{
		{Uuid: "01020304-0506-0708-090a-0b0c0d0e0f10", Username: "Steve"},
		{
			Uuid:     "01020304-0506-0708-090a-0b0c0d0e0f10",
			Username: "Steve",
			Properties: []types.ProfileProperty{
				{Name: "textures", Value: "skin", Signature: &signature},
				{Name: "other", Value: "value"},
			},
		},
	} {
		sent := encodeLoginSuccess(t, &login.LoginSuccessClientboundPacket{Profile: profile, SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20"})
		to1_21 := runTransformer(t, DowngradeLoginSuccessTo1_21, downgradeLoginSuccess(t, sent))
		to1_20_3 := runTransformer(t, DowngradeLoginSuccessTo1_20_3, to1_21)

		got := runTransformer(t, DowngradeLoginSuccessTo1_18_2, to1_20_3)

		want := encodeBody(t, func(ms *streams.MinecraftStream) error {
			if err := ms.WriteUuid(profile.Uuid); err != nil {
				return err
			}

			return ms.WriteString(profile.Username)
		})

		if !bytes.Equal(got, want) {
			t.Errorf("to 1.18.2 = % x\nwant = % x", got, want)
		}
	}
}

// A 1.18.2 hello is the name alone, and 1.19 reads an optional profile key
// behind it: the name goes across with the key absent on the end, which is
// what the steps above read from a 1.19 client holding no key.
func TestUpgradeLoginStartFrom1_18_2AppendsNoProfileKey(t *testing.T) {
	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return ms.WriteString("Notch")
	})

	got := runTransformer(t, UpgradeLoginStartFrom1_18_2, sent)
	want := append(append([]byte{}, sent...), 0x00)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.19 = % x\nwant = % x", got, want)
	}

	// And the steps above read what this one wrote as they would from a
	// 1.19 client: the 1.19.1 step puts the uuid on absent, the 1.19.3 step
	// finds no key to take off, and the 1.20 step settles the uuid.
	nameAlone := append(append([]byte{}, sent...), 0x00)
	upgraded := runTransformer(t, UpgradeLoginStartFrom1_19_1, runTransformer(t, UpgradeLoginStartFrom1_19, got))
	if !bytes.Equal(upgraded, nameAlone) {
		t.Errorf("to 1.19.3 = % x\nwant = % x", upgraded, nameAlone)
	}

	runTransformer(t, UpgradeLoginStartFrom1_20, upgraded)
}

// A 1.18.2 encryption response is the secret and the encrypted challenge,
// and 1.19 reads a flag between the two saying the challenge is encrypted
// rather than signed: the flag goes in set, and the 1.19.3 step reads what
// this one wrote as it would from a 1.19 client that encrypted, carrying
// the challenge across as it is.
func TestUpgradeEncryptionResponseFrom1_18_2FlagsTheChallengeEncrypted(t *testing.T) {
	secret := []byte("the shared secret under the server's key")
	challenge := []byte("the challenge under the server's key")

	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteByteArray(secret); err != nil {
			return err
		}

		return ms.WriteByteArray(challenge)
	})

	got := runTransformer(t, UpgradeEncryptionResponseFrom1_18_2, sent)

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteByteArray(secret); err != nil {
			return err
		}

		if err := ms.WriteBoolean(true); err != nil {
			return err
		}

		return ms.WriteByteArray(challenge)
	})

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.19 = % x\nwant = % x", got, want)
	}

	if upgraded := runTransformer(t, UpgradeEncryptionResponseFrom1_19_1, got); !bytes.Equal(upgraded, sent) {
		t.Errorf("to 1.19.3 = % x\nwant the challenge carried across as sent: % x", upgraded, sent)
	}
}
