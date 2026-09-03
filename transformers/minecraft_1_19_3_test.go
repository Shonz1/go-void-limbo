package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// firstRegistryCodec stands in for what package gamedata hands the
// transformer for 1.19.1, told apart from 1.19.3's and the others' by its
// content.
func firstRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:dimension_type": nbt.Compound{"type": nbt.String("minecraft:dimension_type")}})
	})
}

// The login 1.19.1 reads is the login 1.19.3 reads with one thing changed:
// the registries in the middle are 1.19.1's own rather than 1.19.3's.
// Everything else, on both sides of the registries and with or without a
// death location, is copied, and nothing comes off the end.
func TestDowngradePlayLoginTo1_19_1SwapsTheRegistries(t *testing.T) {
	newer, older := oldestRegistryCodec(t), firstRegistryCodec(t)

	for _, deathLocation := range []bool{false, true} {
		sent := playLogin1_19_4(t, newer, deathLocation)
		got := runTransformer(t, DowngradePlayLoginTo1_19_1(older), sent)
		want := playLogin1_19_4(t, older, deathLocation)

		if !bytes.Equal(got, want) {
			t.Errorf("death location %t: to 1.19.1 = % x\nwant = % x", deathLocation, got, want)
		}

		if bytes.Contains(got, newer) {
			t.Errorf("death location %t: the 1.19.1 login still carries 1.19.3's registries", deathLocation)
		}

		// The registries are the whole of the difference.
		if len(got)-len(older) != len(sent)-len(newer) {
			t.Errorf("death location %t: the body is %d bytes around the registries, want the %d sent", deathLocation, len(got)-len(older), len(sent)-len(newer))
		}
	}
}

// A 1.19.1 client reads the registries out of this packet and nothing else,
// so a transformer with none to write refuses the login rather than send a
// client into a world it cannot make sense of.
func TestDowngradePlayLoginTo1_19_1RefusesToSendNoRegistries(t *testing.T) {
	if err := failingTransformer(t, DowngradePlayLoginTo1_19_1(nil), playLogin1_19_4(t, oldestRegistryCodec(t), false)); err == nil {
		t.Error("error = nil, want a refusal for a login with no registries to carry")
	}
}

// The entry this server sends about a joining player, walked down to
// 1.19.3: the add, the game mode and the listed flag, the hat having come
// off at the 1.21.4 step.
func playerInfoJoin1_19_3(t *testing.T, entry play.PlayerInfoEntry) []byte {
	t.Helper()

	sent := encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{
		Actions: play.PlayerInfoAddPlayer | play.PlayerInfoUpdateGameMode | play.PlayerInfoUpdateListed | play.PlayerInfoUpdateHat,
		Entries: []play.PlayerInfoEntry{entry},
	})

	return runTransformer(t, DowngradePlayerInfoUpdateTo1_21_2, sent)
}

// A packet that adds a player goes out under 1.19.1's add action, with every
// field that action carries: the profile, the game mode the mask named, the
// latency it did not -- none -- and the display name and the profile key
// absent. The listed flag has nowhere to go and comes off.
func TestDowngradePlayerInfoUpdateTo1_19_1IsAnAddPlayer(t *testing.T) {
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

	got := runTransformer(t, DowngradePlayerInfoUpdateTo1_19_1, playerInfoJoin1_19_3(t, entry))

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteVarInt(playerInfoAction1_19_1AddPlayer); err != nil {
			return err
		}

		if err := ms.WriteVarInt(1); err != nil {
			return err
		}

		if err := ms.WriteUuid(entry.Profile.Uuid); err != nil {
			return err
		}

		if err := ms.WriteString(entry.Profile.Username); err != nil {
			return err
		}

		if err := ms.WriteVarInt(1); err != nil {
			return err
		}

		for _, value := range []string{"textures", "dGV4dHVyZXM="} {
			if err := ms.WriteString(value); err != nil {
				return err
			}
		}

		if err := ms.WriteBoolean(true); err != nil {
			return err
		}

		if err := ms.WriteString(signature); err != nil {
			return err
		}

		// The game mode, creative.
		if err := ms.WriteVarInt(1); err != nil {
			return err
		}

		// The latency, none.
		if err := ms.WriteVarInt(0); err != nil {
			return err
		}

		// The display name and the profile key, absent.
		if err := ms.WriteBoolean(false); err != nil {
			return err
		}

		return ms.WriteBoolean(false)
	})

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.19.1 = % x\nwant = % x", got, want)
	}
}

// A packet that changes one thing about a player goes out under that one
// action, with the uuid and the one field.
func TestDowngradePlayerInfoUpdateTo1_19_1PicksTheOneAction(t *testing.T) {
	uuid := "01020304-0506-0708-090a-0b0c0d0e0f10"

	tests := []struct {
		name    string
		actions play.PlayerInfoAction
		entry   play.PlayerInfoEntry
		want    []byte
	}{
		{
			name:    "the game mode",
			actions: play.PlayerInfoUpdateGameMode,
			entry:   play.PlayerInfoEntry{Profile: types.GameProfile{Uuid: uuid}, GameMode: types.GameModeSpectator},
			want: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteVarInt(playerInfoAction1_19_1UpdateGameMode); err != nil {
					return err
				}

				if err := ms.WriteVarInt(1); err != nil {
					return err
				}

				if err := ms.WriteUuid(uuid); err != nil {
					return err
				}

				return ms.WriteVarInt(3)
			}),
		},
		{
			name:    "the latency",
			actions: play.PlayerInfoUpdateLatency,
			entry:   play.PlayerInfoEntry{Profile: types.GameProfile{Uuid: uuid}, Latency: 300},
			want: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteVarInt(playerInfoAction1_19_1UpdateLatency); err != nil {
					return err
				}

				if err := ms.WriteVarInt(1); err != nil {
					return err
				}

				if err := ms.WriteUuid(uuid); err != nil {
					return err
				}

				return ms.WriteVarInt(300)
			}),
		},
		{
			name:    "the display name",
			actions: play.PlayerInfoUpdateDisplayName,
			entry:   play.PlayerInfoEntry{Profile: types.GameProfile{Uuid: uuid}},
			want: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteVarInt(playerInfoAction1_19_1UpdateDisplayName); err != nil {
					return err
				}

				if err := ms.WriteVarInt(1); err != nil {
					return err
				}

				if err := ms.WriteUuid(uuid); err != nil {
					return err
				}

				return ms.WriteBoolean(false)
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sent := encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{Actions: test.actions, Entries: []play.PlayerInfoEntry{test.entry}})
			got := runTransformer(t, DowngradePlayerInfoUpdateTo1_19_1, sent)

			if !bytes.Equal(got, test.want) {
				t.Errorf("to 1.19.1 = % x\nwant = % x", got, test.want)
			}
		})
	}
}

// A packet 1.19.1 has no single action for is refused: two changes at once,
// a change it has no action for, or an optional this rewrite does not walk.
func TestDowngradePlayerInfoUpdateTo1_19_1RefusesWhatItCannotSay(t *testing.T) {
	uuid := "01020304-0506-0708-090a-0b0c0d0e0f10"
	entry := play.PlayerInfoEntry{Profile: types.GameProfile{Uuid: uuid, Username: "Notch"}, Listed: true}

	tests := []struct {
		name string
		sent []byte
	}{
		{
			name: "two changes at once",
			sent: encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{Actions: play.PlayerInfoUpdateGameMode | play.PlayerInfoUpdateLatency, Entries: []play.PlayerInfoEntry{entry}}),
		},
		{
			name: "the listed flag alone",
			sent: encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{Actions: play.PlayerInfoUpdateListed, Entries: []play.PlayerInfoEntry{entry}}),
		},
		{
			name: "a chat session alone",
			sent: encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{Actions: play.PlayerInfoInitializeChat, Entries: []play.PlayerInfoEntry{entry}}),
		},
		{
			name: "a chat session present",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteByte(byte(play.PlayerInfoAddPlayer | play.PlayerInfoInitializeChat)); err != nil {
					return err
				}

				if err := ms.WriteVarInt(1); err != nil {
					return err
				}

				if err := ms.WriteUuid(uuid); err != nil {
					return err
				}

				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				if err := ms.WriteVarInt(0); err != nil {
					return err
				}

				return ms.WriteBoolean(true)
			}),
		},
		{
			name: "a display name present",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteByte(byte(play.PlayerInfoUpdateDisplayName)); err != nil {
					return err
				}

				if err := ms.WriteVarInt(1); err != nil {
					return err
				}

				if err := ms.WriteUuid(uuid); err != nil {
					return err
				}

				if err := ms.WriteBoolean(true); err != nil {
					return err
				}

				return ms.WriteString(`{"text":"Notch"}`)
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := failingTransformer(t, DowngradePlayerInfoUpdateTo1_19_1, test.sent); err == nil {
				t.Error("error = nil, want the packet refused")
			}
		})
	}
}

// The remove goes out as the same packet under its remove action, the uuids
// laid out as 1.19.3 lays them out.
func TestDowngradePlayerInfoRemoveTo1_19_1IsARemovePlayer(t *testing.T) {
	uuids := []string{"01020304-0506-0708-090a-0b0c0d0e0f10", "10203040-5060-7080-90a0-b0c0d0e0f001"}

	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return (&play.PlayerInfoRemoveClientboundPacket{Uuids: uuids}).Encode(ms)
	})

	got := runTransformer(t, DowngradePlayerInfoRemoveTo1_19_1, sent)

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteVarInt(playerInfoAction1_19_1RemovePlayer); err != nil {
			return err
		}

		if err := ms.WriteVarInt(int32(len(uuids))); err != nil {
			return err
		}

		for _, uuid := range uuids {
			if err := ms.WriteUuid(uuid); err != nil {
				return err
			}
		}

		return nil
	})

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.19.1 = % x\nwant = % x", got, want)
	}
}

// The pose serializer is 19 on 1.19.3 and 18 on 1.19.1. Walking a packet
// down through the 1.21.9, 1.20.5, 1.19.4 and 1.19.3 steps lands it there,
// with the flags entry, whose byte serializer sits at 0 throughout,
// untouched.
func TestDowngradeSetEntityDataTo1_19_1RenamesThePoseSerializer(t *testing.T) {
	stance := &play.SetEntityDataClientboundPacket{EntityId: 2, Sneaking: true, Sprinting: true}

	encoded := encodeAddEntityData(t, stance)
	to1_21_7 := runTransformer(t, DowngradeSetEntityDataTo1_21_7, encoded)
	to1_19_4 := runTransformer(t, DowngradeSetEntityDataTo1_20_3, to1_21_7)
	to1_19_3 := runTransformer(t, DowngradeSetEntityDataTo1_19_3, to1_19_4)

	got := runTransformer(t, DowngradeSetEntityDataTo1_19_1, to1_19_3)

	if len(got) != len(to1_19_3) {
		t.Fatalf("to 1.19.1 is %d bytes, want the %d sent: only a serializer id changes", len(got), len(to1_19_3))
	}

	for i := range got {
		if i == 5 {
			continue
		}

		if got[i] != to1_19_3[i] {
			t.Errorf("byte %d = %#x, want %#x untouched", i, got[i], to1_19_3[i])
		}
	}

	if to1_19_3[5] != poseSerializer1_19_3 {
		t.Fatalf("pose serializer sent = %d, want %d on 1.19.3", to1_19_3[5], poseSerializer1_19_3)
	}

	if got[5] != poseSerializer1_19_1 {
		t.Errorf("pose serializer = %d, want %d", got[5], poseSerializer1_19_1)
	}
}

// A 1.19.1 client sends its profile key as an optional between the name and
// the optional uuid, and what 1.19.3 reads is the name and the optional uuid
// alone: the key comes off, present or not, and the uuid is left to the 1.20
// step.
func TestUpgradeLoginStartFrom1_19_1DropsTheProfileKey(t *testing.T) {
	uuid := "01020304-0506-0708-090a-0b0c0d0e0f10"

	withUuid := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteString("Notch"); err != nil {
			return err
		}

		if err := ms.WriteBoolean(true); err != nil {
			return err
		}

		return ms.WriteUuid(uuid)
	})

	withoutUuid := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteString("Notch"); err != nil {
			return err
		}

		return ms.WriteBoolean(false)
	})

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
		want []byte
	}{
		{
			name: "with a key and a uuid",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				if err := profileKey(ms); err != nil {
					return err
				}

				if err := ms.WriteBoolean(true); err != nil {
					return err
				}

				return ms.WriteUuid(uuid)
			}),
			want: withUuid,
		},
		{
			name: "with a key and no uuid",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				if err := profileKey(ms); err != nil {
					return err
				}

				return ms.WriteBoolean(false)
			}),
			want: withoutUuid,
		},
		{
			name: "with no key and a uuid",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				if err := ms.WriteBoolean(false); err != nil {
					return err
				}

				if err := ms.WriteBoolean(true); err != nil {
					return err
				}

				return ms.WriteUuid(uuid)
			}),
			want: withUuid,
		},
		{
			name: "with neither",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				if err := ms.WriteBoolean(false); err != nil {
					return err
				}

				return ms.WriteBoolean(false)
			}),
			want: withoutUuid,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := runTransformer(t, UpgradeLoginStartFrom1_19_1, test.sent)

			if !bytes.Equal(got, test.want) {
				t.Errorf("to 1.19.3 = % x\nwant = % x", got, test.want)
			}

			// And the 1.20 step reads what this one wrote, as it would from
			// a 1.19.3 client.
			runTransformer(t, UpgradeLoginStartFrom1_20, got)
		})
	}
}

// A 1.19.1 client answers the challenge one of two ways, and says which with
// a flag in front. An encrypted challenge is carried across without the
// flag, which is what 1.19.3 reads; a signed one is dropped, and the
// response goes on with no challenge at all, since the key it was signed
// under is not one this server kept.
func TestUpgradeEncryptionResponseFrom1_19_1(t *testing.T) {
	secret := []byte("the shared secret under the server's key")
	challenge := []byte("the challenge under the server's key")

	tests := []struct {
		name string
		sent []byte
		want []byte
	}{
		{
			name: "an encrypted challenge",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteByteArray(secret); err != nil {
					return err
				}

				if err := ms.WriteBoolean(true); err != nil {
					return err
				}

				return ms.WriteByteArray(challenge)
			}),
			want: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteByteArray(secret); err != nil {
					return err
				}

				return ms.WriteByteArray(challenge)
			}),
		},
		{
			name: "a signed challenge",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteByteArray(secret); err != nil {
					return err
				}

				if err := ms.WriteBoolean(false); err != nil {
					return err
				}

				// The salt and the signature under the profile key.
				if err := ms.WriteLong(0x0102030405060708); err != nil {
					return err
				}

				return ms.WriteByteArray([]byte("signed under the profile key"))
			}),
			want: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteByteArray(secret); err != nil {
					return err
				}

				return ms.WriteByteArray(nil)
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := runTransformer(t, UpgradeEncryptionResponseFrom1_19_1, test.sent)

			if !bytes.Equal(got, test.want) {
				t.Errorf("to 1.19.3 = % x\nwant = % x", got, test.want)
			}
		})
	}
}
