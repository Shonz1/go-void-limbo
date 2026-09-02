package transformers

import (
	"bytes"
	"encoding/binary"
	"math"
	"strings"
	"testing"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

func failingTransformer(t *testing.T, transformer func(in, out *streams.MinecraftStream) error, body []byte) error {
	t.Helper()

	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))
	out := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))

	return transformer(in, out)
}

// 1.21 reads the profile and then one boolean 1.21.2 removed, so the
// downgraded body is what 26.1 reads with a byte of true on the end.
func TestDowngradeLoginSuccessTo1_21AppendsStrictErrorHandling(t *testing.T) {
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
		to26_1 := downgradeLoginSuccess(t, sent)

		got := runTransformer(t, DowngradeLoginSuccessTo1_21, to26_1)
		want := append(append([]byte{}, to26_1...), 0x01)

		if !bytes.Equal(got, want) {
			t.Errorf("to 1.21 = % x\nwant = % x", got, want)
		}
	}
}

// The sea level is a var int right in front of the last byte, so what 1.21
// reads is what 1.21.2 sends with that var int cut out, and the enforces
// secure chat flag has to land where 1.21 looks for it.
func TestDowngradePlayLoginTo1_21DropsTheSeaLevel(t *testing.T) {
	cases := []struct {
		name     string
		packet   play.LoginClientboundPacket
		seaLevel int
	}{
		{
			name: "what a join sends",
			packet: play.LoginClientboundPacket{
				EntityId:           1,
				Dimensions:         []string{"minecraft:overworld"},
				ViewDistance:       2,
				SimulationDistance: 2,
				ShowDeathScreen:    true,
				EnforcesSecureChat: true,
				SpawnInfo: play.SpawnInfo{
					Dimension:        "minecraft:overworld",
					GameMode:         types.GameModeSpectator,
					PreviousGameMode: types.GameModeNone,
					SeaLevel:         63,
				},
			},
			seaLevel: 1,
		},
		{
			name: "with a death location and a sea level of two bytes",
			packet: play.LoginClientboundPacket{
				EntityId:   7,
				Dimensions: []string{"minecraft:overworld", "minecraft:the_nether"},
				SpawnInfo: play.SpawnInfo{
					Dimension: "minecraft:the_nether",
					DeathLocation: &play.GlobalPos{
						Dimension: "minecraft:overworld",
						Position:  play.BlockPos{X: 1, Y: 2, Z: 3},
					},
					PortalCooldown: 400,
					SeaLevel:       200,
				},
			},
			seaLevel: 2,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// The body every step below 26.2 starts from.
			sent := withoutOnlineMode(encode(t, &c.packet))
			got := runTransformer(t, DowngradePlayLoginTo1_21, sent)

			want := append(append([]byte{}, sent[:len(sent)-1-c.seaLevel]...), sent[len(sent)-1])

			if !bytes.Equal(got, want) {
				t.Errorf("to 1.21 = % x\nwant = % x", got, want)
			}

			if got[len(got)-1] != sent[len(sent)-1] {
				t.Errorf("enforces secure chat came out as %d, want %d", got[len(got)-1], sent[len(sent)-1])
			}
		})
	}
}

// The list order is the action 1.21.2 appended, so a packet carrying it
// comes out as the same packet encoded without it, and a packet without it
// crosses untouched.
func TestDowngradePlayerInfoUpdateTo1_21DropsTheListOrder(t *testing.T) {
	entries := []play.PlayerInfoEntry{
		{
			Profile:   types.GameProfile{Uuid: "01020304-0506-0708-090a-0b0c0d0e0f10", Username: "Steve"},
			GameMode:  types.GameModeAdventure,
			Listed:    true,
			Latency:   130,
			ListOrder: 300,
		},
		{
			Profile:   types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000002", Username: "Alex"},
			ListOrder: -1,
		},
	}

	for _, actions := range []play.PlayerInfoAction{
		// What this server sends when a player joins, after the 1.21.4 step
		// took the hat off.
		play.PlayerInfoAddPlayer | play.PlayerInfoUpdateGameMode | play.PlayerInfoUpdateListed,
		// Every action 1.21.2 has.
		play.PlayerInfoAddPlayer | play.PlayerInfoInitializeChat | play.PlayerInfoUpdateGameMode | play.PlayerInfoUpdateListed |
			play.PlayerInfoUpdateLatency | play.PlayerInfoUpdateDisplayName | play.PlayerInfoUpdateListOrder,
		// The list order on its own.
		play.PlayerInfoUpdateListOrder,
	} {
		body := encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{Actions: actions, Entries: entries})
		got := runTransformer(t, DowngradePlayerInfoUpdateTo1_21, body)

		want := encodePlayerInfoUpdate(t, &play.PlayerInfoUpdateClientboundPacket{Actions: actions &^ play.PlayerInfoUpdateListOrder, Entries: entries})

		if !bytes.Equal(got, want) {
			t.Errorf("actions %s: to 1.21 = % x\nwant = % x", actions, got, want)
		}
	}
}

func encodePlayerPosition(t *testing.T, packet *play.PlayerPositionClientboundPacket) []byte {
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

// 1.21 reads the position, the rotation, a byte of flags and then the
// teleport id, with no delta anywhere.
func TestDowngradePlayerPositionTo1_21ReordersThePacket(t *testing.T) {
	body := encodePlayerPosition(t, &play.PlayerPositionClientboundPacket{
		TeleportId: 300,
		X:          0.5,
		Y:          64,
		Z:          -0.5,
		Yaw:        90,
		Pitch:      -45,
		Relatives:  0x18, // yaw and pitch relative
	})

	got := runTransformer(t, DowngradePlayerPositionTo1_21, body)

	want := []byte{
		0x3F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x, 0.5
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y, 64
		0xBF, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z, -0.5
		0x42, 0xB4, 0x00, 0x00, // yaw, 90
		0xC2, 0x34, 0x00, 0x00, // pitch, -45
		0x18,       // the flags, one byte now
		0xAC, 0x02, // the teleport id, 300
	}

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.21 = % x\nwant = % x", got, want)
	}
}

// A delta 1.21 has no field for, or a relative flag it has no bit for, is a
// packet this rewrite refuses rather than silently narrows.
func TestDowngradePlayerPositionTo1_21RefusesWhatItCannotCarry(t *testing.T) {
	for _, tc := range []struct {
		name   string
		packet play.PlayerPositionClientboundPacket
		want   string
	}{
		{"a delta", play.PlayerPositionClientboundPacket{TeleportId: 1, DeltaY: -0.5}, "delta y"},
		{"a delta flag", play.PlayerPositionClientboundPacket{TeleportId: 1, Relatives: 0x100}, "relative flags"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := failingTransformer(t, DowngradePlayerPositionTo1_21, encodePlayerPosition(t, &tc.packet))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error = %v, want a refusal naming the %s", err, tc.want)
			}
		})
	}
}

func encodeEntityPositionSync(t *testing.T, packet *play.EntityPositionSyncClientboundPacket) []byte {
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

// What comes out is a 1.21 teleport entity body: id, position, the two
// rotations as angle bytes, on ground.
func TestDowngradeEntityPositionSyncTo1_21IsATeleport(t *testing.T) {
	body := encodeEntityPositionSync(t, &play.EntityPositionSyncClientboundPacket{
		EntityId: 2,
		X:        0.5,
		Y:        64,
		Z:        -0.5,
		Yaw:      90,
		Pitch:    -90,
		OnGround: true,
	})

	got := runTransformer(t, DowngradeEntityPositionSyncTo1_21, body)

	want := []byte{
		0x02,
		0x3F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x, 0.5
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y, 64
		0xBF, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z, -0.5
		0x40, // yaw, 90 degrees as an angle byte
		0xC0, // pitch, -90
		0x01, // on ground
	}

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.21 = % x\nwant = % x", got, want)
	}
}

func TestDowngradeEntityPositionSyncTo1_21RefusesADelta(t *testing.T) {
	// The packet's own encoder always writes a zero delta, so the body with
	// one is built by hand: the id, the position, then a delta x of 1.
	body := []byte{0x02}
	body = append(body, make([]byte, 24)...)
	body = binary.BigEndian.AppendUint64(body, math.Float64bits(1))
	body = append(body, make([]byte, 16+8+1)...)

	err := failingTransformer(t, DowngradeEntityPositionSyncTo1_21, body)
	if err == nil || !strings.Contains(err.Error(), "delta x") {
		t.Errorf("error = %v, want a refusal naming the delta", err)
	}
}

// A 1.21 body is two floats and a flag byte; what 1.21.2 reads is one byte
// with a bit per key.
func TestUpgradePlayerInputFrom1_21(t *testing.T) {
	input := func(sideways, forward float32, flags byte) []byte {
		body := binary.BigEndian.AppendUint32(nil, math.Float32bits(sideways))
		body = binary.BigEndian.AppendUint32(body, math.Float32bits(forward))

		return append(body, flags)
	}

	for _, tc := range []struct {
		name string
		body []byte
		want byte
	}{
		{"nothing held", input(0, 0, 0), 0x00},
		{"forward", input(0, 1, 0), 0x01},
		{"backward", input(0, -1, 0), 0x02},
		{"left", input(1, 0, 0), 0x04},
		{"right", input(-1, 0, 0), 0x08},
		{"jump", input(0, 0, 0x01), 0x10},
		{"sneak", input(0, 0, 0x02), 0x20},
		{"forward and left while sneaking, at a sneaking impulse", input(0.3, 0.3, 0x02), 0x25},
		{"everything at once", input(-0.98, -0.98, 0x03), 0x3A},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := runTransformer(t, UpgradePlayerInputFrom1_21, tc.body)

			if !bytes.Equal(got, []byte{tc.want}) {
				t.Errorf("from 1.21 = % x, want %#02x", got, tc.want)
			}
		})
	}
}
