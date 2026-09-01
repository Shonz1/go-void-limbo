package transformers

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"testing"
)

// encode is what the packet writes at 26.2, which is the only version it knows
// how to be and the input every downgrade starts from.
func encode(t *testing.T, packet *play.LoginClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("failed to flush: %v", err)
	}

	return buf.Bytes()
}

func downgrade(t *testing.T, body []byte) []byte {
	t.Helper()

	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := DowngradePlayLoginTo26_1(in, out); err != nil {
		t.Fatalf("failed to downgrade: %v", err)
	}

	if err := out.Flush(); err != nil {
		t.Fatalf("failed to flush: %v", err)
	}

	return buf.Bytes()
}

// The online mode flag and enforces secure chat are the last two fields and a
// byte each, so what 26.1 reads is what 26.2 sends without the second to last
// byte. Checking it that way pins the transformer against the encoder rather
// than against a copy of the layout written out twice.
func withoutOnlineMode(body []byte) []byte {
	trimmed := make([]byte, 0, len(body)-1)
	trimmed = append(trimmed, body[:len(body)-2]...)

	return append(trimmed, body[len(body)-1])
}

func TestDowngradePlayLoginTo26_1(t *testing.T) {
	cases := []struct {
		name   string
		packet play.LoginClientboundPacket
	}{
		{
			name: "what a join sends",
			packet: play.LoginClientboundPacket{
				EntityId:           1,
				Dimensions:         []string{"minecraft:overworld"},
				ViewDistance:       2,
				SimulationDistance: 2,
				ShowDeathScreen:    true,
				OnlineMode:         true,
				SpawnInfo: play.SpawnInfo{
					Dimension:        "minecraft:overworld",
					GameMode:         types.GameModeSpectator,
					PreviousGameMode: types.GameModeNone,
				},
			},
		},
		{
			name: "with a death location, which is the one optional field in the way",
			packet: play.LoginClientboundPacket{
				EntityId:           7,
				Hardcore:           true,
				Dimensions:         []string{"minecraft:overworld", "minecraft:the_nether"},
				MaxPlayers:         20,
				ViewDistance:       10,
				SimulationDistance: 10,
				ReducedDebugInfo:   true,
				DoLimitedCrafting:  true,
				OnlineMode:         true,
				EnforcesSecureChat: true,
				SpawnInfo: play.SpawnInfo{
					DimensionTypeId: 3,
					Dimension:       "minecraft:the_nether",
					HashedSeed:      -1,
					GameMode:        types.GameModeCreative,
					IsFlat:          true,
					DeathLocation: &play.GlobalPos{
						Dimension: "minecraft:overworld",
						Position:  play.BlockPos{X: 1, Y: 2, Z: 3},
					},
					PortalCooldown: 400,
					SeaLevel:       63,
				},
			},
		},
		{
			name: "online mode false, which still has to come out rather than be left",
			packet: play.LoginClientboundPacket{
				EntityId:           1,
				Dimensions:         []string{"minecraft:overworld"},
				OnlineMode:         false,
				EnforcesSecureChat: true,
				SpawnInfo: play.SpawnInfo{
					Dimension:        "minecraft:overworld",
					PreviousGameMode: types.GameModeNone,
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sent := encode(t, &c.packet)
			got := downgrade(t, sent)
			want := withoutOnlineMode(sent)

			if !bytes.Equal(got, want) {
				t.Errorf("downgraded body is\n%v\nwant\n%v", got, want)
			}

			if len(got) != len(sent)-1 {
				t.Errorf("expected exactly one byte to come out, got %d fewer", len(sent)-len(got))
			}

			// Enforces secure chat is what a 26.1 client reads where 26.2 put
			// the online mode flag, so the value that ends up there is the whole
			// point of the transformer.
			if got[len(got)-1] != sent[len(sent)-1] {
				t.Errorf("enforces secure chat came out as %d, want %d", got[len(got)-1], sent[len(sent)-1])
			}
		})
	}
}

// A body that stops short is a body the transformer has to refuse rather than
// pass on half rewritten.
func TestDowngradePlayLoginTo26_1RefusesATruncatedBody(t *testing.T) {
	packet := play.LoginClientboundPacket{
		EntityId:   1,
		Dimensions: []string{"minecraft:overworld"},
		SpawnInfo:  play.SpawnInfo{Dimension: "minecraft:overworld"},
	}

	sent := encode(t, &packet)

	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(sent[:len(sent)/2]))
	out := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))

	if err := DowngradePlayLoginTo26_1(in, out); err == nil {
		t.Error("expected a truncated body to be refused")
	}
}

// encodeLoginSuccess is what the login success packet writes at 26.2.
func encodeLoginSuccess(t *testing.T, packet *login.LoginSuccessClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("failed to encode: %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("failed to flush: %v", err)
	}

	return buf.Bytes()
}

func downgradeLoginSuccess(t *testing.T, body []byte) []byte {
	t.Helper()

	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := DowngradeLoginSuccessTo26_1(in, out); err != nil {
		t.Fatalf("failed to downgrade: %v", err)
	}

	if err := out.Flush(); err != nil {
		t.Fatalf("failed to flush: %v", err)
	}

	return buf.Bytes()
}

// The session id is the last field and a uuid, so what 26.1 reads is what 26.2
// sends with the last sixteen bytes taken off.
func TestDowngradeLoginSuccessTo26_1(t *testing.T) {
	signature := "signed"

	cases := []struct {
		name   string
		packet login.LoginSuccessClientboundPacket
	}{
		{
			name: "no properties",
			packet: login.LoginSuccessClientboundPacket{
				Profile:   types.GameProfile{Uuid: "01020304-0506-0708-090a-0b0c0d0e0f10", Username: "Steve"},
				SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20",
			},
		},
		{
			name: "an unsigned property",
			packet: login.LoginSuccessClientboundPacket{
				Profile: types.GameProfile{
					Uuid:       "01020304-0506-0708-090a-0b0c0d0e0f10",
					Username:   "Steve",
					Properties: []types.ProfileProperty{{Name: "textures", Value: "skin"}},
				},
				SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20",
			},
		},
		{
			name: "a signed property, which is what an authenticated login carries",
			packet: login.LoginSuccessClientboundPacket{
				Profile: types.GameProfile{
					Uuid:     "01020304-0506-0708-090a-0b0c0d0e0f10",
					Username: "Steve",
					Properties: []types.ProfileProperty{
						{Name: "textures", Value: "skin", Signature: &signature},
						{Name: "other", Value: "value"},
					},
				},
				SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20",
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sent := encodeLoginSuccess(t, &c.packet)
			got := downgradeLoginSuccess(t, sent)
			want := sent[:len(sent)-16]

			if !bytes.Equal(got, want) {
				t.Errorf("downgraded body is\n%v\nwant\n%v", got, want)
			}

			if len(got) != len(sent)-16 {
				t.Errorf("expected the sixteen byte session id to come out, got %d fewer", len(sent)-len(got))
			}
		})
	}
}

func TestDowngradeLoginSuccessTo26_1RefusesATruncatedBody(t *testing.T) {
	packet := login.LoginSuccessClientboundPacket{
		Profile:   types.GameProfile{Uuid: "01020304-0506-0708-090a-0b0c0d0e0f10", Username: "Steve"},
		SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20",
	}

	sent := encodeLoginSuccess(t, &packet)

	// The body without its session id is what 26.1 reads, and is exactly the
	// body this transformer must not accept as its input: there is nothing left
	// to take off.
	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(sent[:len(sent)-16]))
	out := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))

	if err := DowngradeLoginSuccessTo26_1(in, out); err == nil {
		t.Error("expected a body with no session id on the end to be refused")
	}
}
