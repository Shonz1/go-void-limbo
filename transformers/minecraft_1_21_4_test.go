package transformers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

func encodePlayerInfoUpdate(t *testing.T, packet *play.PlayerInfoUpdateClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

// The downgraded body is what the same packet encodes to without the hat
// action: the bit off the mask and the boolean off the end of every entry,
// with everything in front of it walked across untouched, profile
// properties included.
func TestDowngradePlayerInfoUpdateTo1_21_2DropsTheHat(t *testing.T) {
	signature := "sig"
	entries := []play.PlayerInfoEntry{
		{
			Profile: types.GameProfile{
				Uuid:     "01020304-0506-0708-090a-0b0c0d0e0f10",
				Username: "Steve",
				Properties: []types.ProfileProperty{
					{Name: "textures", Value: "skin", Signature: &signature},
				},
			},
			GameMode:  types.GameModeAdventure,
			Listed:    true,
			Latency:   130,
			ListOrder: 2,
			ShowHat:   true,
		},
		{
			Profile:  types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000002", Username: "Alex"},
			GameMode: types.GameModeSpectator,
			ShowHat:  false,
		},
	}

	for _, actions := range []play.PlayerInfoAction{
		// What this server sends when a player joins.
		play.PlayerInfoAddPlayer | play.PlayerInfoUpdateGameMode | play.PlayerInfoUpdateListed | play.PlayerInfoUpdateHat,
		// Every action there is.
		play.PlayerInfoAddPlayer | play.PlayerInfoInitializeChat | play.PlayerInfoUpdateGameMode | play.PlayerInfoUpdateListed |
			play.PlayerInfoUpdateLatency | play.PlayerInfoUpdateDisplayName | play.PlayerInfoUpdateListOrder | play.PlayerInfoUpdateHat,
		// The hat on its own.
		play.PlayerInfoUpdateHat,
	} {
		body := encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{Actions: actions, Entries: entries})
		got := runTransformer(t, DowngradePlayerInfoUpdateTo1_21_2, body)

		want := encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{Actions: actions &^ play.PlayerInfoUpdateHat, Entries: entries})

		if !bytes.Equal(got, want) {
			t.Errorf("actions %s: to 1.21.2 = % x\nwant = % x", actions, got, want)
		}
	}
}

// A packet without the hat crosses untouched, since 1.21.2 reads every other
// action as 1.21.4 does.
func TestDowngradePlayerInfoUpdateTo1_21_2LeavesAPacketWithoutTheHatAlone(t *testing.T) {
	body := encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{
		Actions: play.PlayerInfoUpdateListed | play.PlayerInfoUpdateLatency,
		Entries: []play.PlayerInfoEntry{
			{Profile: types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001"}, Listed: true, Latency: 7},
		},
	})

	if got := runTransformer(t, DowngradePlayerInfoUpdateTo1_21_2, body); !bytes.Equal(got, body) {
		t.Errorf("to 1.21.2 = % x, want the body untouched, % x", got, body)
	}
}

// The two optionals this server always writes absent are the two this
// rewrite never learned to walk, so a present one is refused rather than
// copied as if it were the next field.
func TestDowngradePlayerInfoUpdateTo1_21_2RefusesWhatItCannotWalk(t *testing.T) {
	uuid := bytes.Repeat([]byte{0x00}, uuidSize)

	for _, tc := range []struct {
		name string
		body []byte
		want string
	}{
		{
			name: "display name",
			body: append(append([]byte{byte(play.PlayerInfoUpdateDisplayName | play.PlayerInfoUpdateHat), 0x01}, uuid...), 0x01),
			want: "display name",
		},
		{
			name: "chat session",
			body: append(append([]byte{byte(play.PlayerInfoInitializeChat | play.PlayerInfoUpdateHat), 0x01}, uuid...), 0x01),
			want: "chat session",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(tc.body))
			out := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))

			err := DowngradePlayerInfoUpdateTo1_21_2(in, out)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want a refusal naming the %s", err, tc.name)
			}
		})
	}
}
