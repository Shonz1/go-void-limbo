package play

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"testing"
)

func TestPlayerInfoActionBits(t *testing.T) {
	// The bit is the action's position in the client's own list, so these are
	// not free to change.
	tests := []struct {
		action PlayerInfoAction
		want   byte
	}{
		{PlayerInfoAddPlayer, 0x01},
		{PlayerInfoInitializeChat, 0x02},
		{PlayerInfoUpdateGameMode, 0x04},
		{PlayerInfoUpdateListed, 0x08},
		{PlayerInfoUpdateLatency, 0x10},
		{PlayerInfoUpdateDisplayName, 0x20},
		{PlayerInfoUpdateListOrder, 0x40},
		{PlayerInfoUpdateHat, 0x80},
	}

	for _, test := range tests {
		if byte(test.action) != test.want {
			t.Errorf("%s = %#02x, want %#02x", test.action, byte(test.action), test.want)
		}
	}
}

func TestPlayerInfoUpdateClientboundPacketEncodeJoin(t *testing.T) {
	// What the join sends: the entry is created, put in the list, and given a
	// game mode.
	p := &PlayerInfoUpdateClientboundPacket{
		Actions: PlayerInfoAddPlayer | PlayerInfoUpdateGameMode | PlayerInfoUpdateListed,
		Entries: []PlayerInfoEntry{
			{
				Profile:  types.GameProfile{Uuid: "01020304-0506-0708-090a-0b0c0d0e0f10", Username: "Steve"},
				GameMode: GameModeSpectator,
				Listed:   true,
			},
		},
	}

	want := []byte{
		0x0d, // add player, update game mode, update listed
		0x01, // one entry
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x05, 'S', 't', 'e', 'v', 'e',
		0x00, // no profile properties
		0x03, // spectator
		0x01, // listed
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

// Only the fields the actions name are written, and they are written in the
// order the actions are declared in rather than the order of the fields.
func TestPlayerInfoUpdateClientboundPacketEncodeEveryAction(t *testing.T) {
	signature := "sig"
	p := &PlayerInfoUpdateClientboundPacket{
		Actions: PlayerInfoAddPlayer | PlayerInfoInitializeChat | PlayerInfoUpdateGameMode | PlayerInfoUpdateListed |
			PlayerInfoUpdateLatency | PlayerInfoUpdateDisplayName | PlayerInfoUpdateListOrder | PlayerInfoUpdateHat,
		Entries: []PlayerInfoEntry{
			{
				Profile: types.GameProfile{
					Uuid:     "01020304-0506-0708-090a-0b0c0d0e0f10",
					Username: "Steve",
					Properties: []types.ProfileProperty{
						{Name: "textures", Value: "skin", Signature: &signature},
					},
				},
				GameMode:  GameModeAdventure,
				Listed:    true,
				Latency:   130,
				ListOrder: 2,
				ShowHat:   true,
			},
		},
	}

	want := []byte{
		0xff, // every action
		0x01, // one entry
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x05, 'S', 't', 'e', 'v', 'e',
		0x01, // one property
		0x08, 't', 'e', 'x', 't', 'u', 'r', 'e', 's',
		0x04, 's', 'k', 'i', 'n',
		0x01,                // it is signed
		0x03, 's', 'i', 'g', //
		0x00,       // no chat session
		0x02,       // adventure
		0x01,       // listed
		0x82, 0x01, // latency, a VarInt
		0x00, // no display name of its own
		0x02, // list order
		0x01, // show hat
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

func TestPlayerInfoUpdateClientboundPacketEncodeManyEntries(t *testing.T) {
	p := &PlayerInfoUpdateClientboundPacket{
		Actions: PlayerInfoUpdateListed,
		Entries: []PlayerInfoEntry{
			{Profile: types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001"}, Listed: true},
			{Profile: types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000002"}},
		},
	}

	want := []byte{
		0x08, // update listed
		0x02, // two entries
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01,
		0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02,
		0x00,
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

func TestPlayerInfoUpdateClientboundPacketEncodeRejectsInvalidUuid(t *testing.T) {
	p := &PlayerInfoUpdateClientboundPacket{
		Actions: PlayerInfoAddPlayer,
		Entries: []PlayerInfoEntry{{Profile: types.GameProfile{Uuid: "not-a-uuid", Username: "Steve"}}},
	}

	stream := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))
	if err := p.Encode(stream); err == nil {
		t.Error("Encode() error = nil, want an error for a malformed profile uuid")
	}
}

func TestPlayerInfoActionString(t *testing.T) {
	tests := []struct {
		actions PlayerInfoAction
		want    string
	}{
		{0, "none"},
		{PlayerInfoAddPlayer, "add_player"},
		{PlayerInfoAddPlayer | PlayerInfoUpdateGameMode | PlayerInfoUpdateListed, "add_player|update_game_mode|update_listed"},
	}

	for _, test := range tests {
		if got := test.actions.String(); got != test.want {
			t.Errorf("String() = %q, want %q", got, test.want)
		}
	}
}
