package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

func encodeEncryptionRequest(t *testing.T, packet *login.EncryptionRequestClientboundPacket) []byte {
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

// 1.20.3 reads the server id, the key and the token and nothing after them,
// so the downgraded body is what 1.20.5 reads with its last byte cut off.
func TestDowngradeEncryptionRequestTo1_20_3DropsShouldAuthenticate(t *testing.T) {
	sent := encodeEncryptionRequest(t, &login.EncryptionRequestClientboundPacket{
		PublicKey:          bytes.Repeat([]byte{0xAB}, 162),
		VerifyToken:        []byte{1, 2, 3, 4},
		ShouldAuthenticate: true,
	})

	got := runTransformer(t, DowngradeEncryptionRequestTo1_20_3, sent)
	want := sent[:len(sent)-1]

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.20.3 = % x\nwant = % x", got, want)
	}
}

// A 1.20.3 client always authenticates, so a request that asks it not to is
// one the rewrite cannot carry the meaning of.
func TestDowngradeEncryptionRequestTo1_20_3RefusesAnUnauthenticatedRequest(t *testing.T) {
	sent := encodeEncryptionRequest(t, &login.EncryptionRequestClientboundPacket{
		PublicKey:   []byte{1},
		VerifyToken: []byte{2},
	})

	if err := failingTransformer(t, DowngradeEncryptionRequestTo1_20_3, sent); err == nil {
		t.Error("error = nil, want a refusal for a request that skips authentication")
	}
}

// The strict error handling flag is what the 1.21.2 step appends and this
// step removes, so what 1.20.3 reads is exactly what 26.1 reads: the profile
// alone, with no session id after it either.
func TestDowngradeLoginSuccessTo1_20_3DropsStrictErrorHandling(t *testing.T) {
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
		to1_21 := runTransformer(t, DowngradeLoginSuccessTo1_21, to26_1)

		got := runTransformer(t, DowngradeLoginSuccessTo1_20_3, to1_21)

		if !bytes.Equal(got, to26_1) {
			t.Errorf("to 1.20.3 = % x\nwant = % x", got, to26_1)
		}
	}
}

// The play login reaches this step as 1.20.5 reads it: the head, the spawn
// info with the dimension type as an index, and the enforces secure chat
// flag. 1.20.3 reads the type as a name and stops after the spawn info.
func TestDowngradePlayLoginTo1_20_3NamesTheDimensionType(t *testing.T) {
	cases := []struct {
		name   string
		packet play.LoginClientboundPacket
	}{
		{
			name: "what a join sends",
			packet: play.LoginClientboundPacket{
				EntityId:           1,
				Dimensions:         []string{"minecraft:overworld"},
				ViewDistance:       8,
				SimulationDistance: 2,
				ShowDeathScreen:    true,
				OnlineMode:         true,
				EnforcesSecureChat: true,
				SpawnInfo: play.SpawnInfo{
					Dimension:        "minecraft:overworld",
					GameMode:         types.GameModeCreative,
					PreviousGameMode: types.GameModeNone,
					SeaLevel:         63,
				},
			},
		},
		{
			name: "with a death location",
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
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// The body every step below 1.21.2 starts from: the online mode
			// flag and the sea level gone.
			to1_21 := runTransformer(t, DowngradePlayLoginTo1_21, withoutOnlineMode(encode(t, &c.packet)))

			got := runTransformer(t, DowngradePlayLoginTo1_20_3, to1_21)

			// The head as 1.20.5 has it, up to the spawn info.
			head := encodePlayLoginHead(t, &c.packet)

			// Then the name in place of the index, then the spawn info from the
			// dimension name on, which is everything between the index and the
			// flag that closed the 1.20.5 body.
			want := append([]byte{}, head...)
			want = append(want, encodeString(t, "minecraft:overworld")...)
			want = append(want, to1_21[len(head)+1:len(to1_21)-1]...)

			if !bytes.Equal(got, want) {
				t.Errorf("to 1.20.3 = % x\nwant = % x", got, want)
			}
		})
	}
}

// The rewrite knows one name, for the one index this server sends. Anything
// else is a dimension type it cannot name, and is refused.
func TestDowngradePlayLoginTo1_20_3RefusesADimensionTypeItCannotName(t *testing.T) {
	packet := play.LoginClientboundPacket{
		EntityId:   1,
		Dimensions: []string{"minecraft:overworld"},
		SpawnInfo:  play.SpawnInfo{Dimension: "minecraft:overworld", DimensionTypeId: 1},
	}

	to1_21 := runTransformer(t, DowngradePlayLoginTo1_21, withoutOnlineMode(encode(t, &packet)))

	if err := failingTransformer(t, DowngradePlayLoginTo1_20_3, to1_21); err == nil {
		t.Error("error = nil, want a refusal for a dimension type index the rewrite has no name for")
	}
}

// encodePlayLoginHead is the play login up to the spawn info, which no step
// touches: everything a 1.20.3 body starts with.
func encodePlayLoginHead(t *testing.T, packet *play.LoginClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	headless := *packet
	headless.SpawnInfo = play.SpawnInfo{}

	if err := headless.Encode(stream); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	// An empty spawn info encodes as a var int 0, an empty string (one byte
	// of length), a long, two bytes, three booleans and two var ints; then
	// the two flags after it. None of it is the head.
	spawnInfoAndFlags := 1 + 1 + 8 + 2 + 3 + 2 + 2

	return buf.Bytes()[:buf.Len()-spawnInfoAndFlags]
}

func encodeString(t *testing.T, value string) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := stream.WriteString(value); err != nil {
		t.Fatalf("WriteString() error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

// The pose serializer is 20 as 26.2 encodes it, 21 after the 1.21.9 step
// puts it where 1.21.7 through 1.20.5 have it, and 20 again for 1.20.3: the
// body that reaches this step comes out as the body that was first encoded.
func TestDowngradeSetEntityDataTo1_20_3RenamesThePoseSerializerBack(t *testing.T) {
	stance := &play.SetEntityDataClientboundPacket{EntityId: 2, Sneaking: true, Sprinting: true}

	encoded := encodeAddEntityData(t, stance)
	to1_21_7 := runTransformer(t, DowngradeSetEntityDataTo1_21_7, encoded)

	got := runTransformer(t, DowngradeSetEntityDataTo1_20_3, to1_21_7)

	if !bytes.Equal(got, encoded) {
		t.Errorf("to 1.20.3 = % x, want % x", got, encoded)
	}

	if got[5] != 0x14 {
		t.Errorf("pose serializer = %d, want 20", got[5])
	}
}

// 128 is a var int of two bytes and 124 one of one, so the body shrinks by
// a byte on the way down, and the result is what encoding the packet with
// 1.20.3's own number gives.
func TestDowngradeAddEntityTo1_20_3RenumbersThePlayer(t *testing.T) {
	packet := play.AddEntityClientboundPacket{
		EntityId:     5,
		Uuid:         "01020304-0506-0708-090a-0b0c0d0e0f10",
		EntityTypeId: playerEntityType1_20_5,
		X:            1,
		Y:            2,
		Z:            3,
	}

	sent := encodeAddEntity(t, &packet)
	got := runTransformer(t, DowngradeAddEntityTo1_20_3, sent)

	packet.EntityTypeId = playerEntityType1_20_3
	want := encodeAddEntity(t, &packet)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.20.3 = % x\nwant = % x", got, want)
	}

	if len(got) != len(sent)-1 {
		t.Errorf("the body is %d bytes, want one fewer than the %d sent", len(got), len(sent))
	}
}
