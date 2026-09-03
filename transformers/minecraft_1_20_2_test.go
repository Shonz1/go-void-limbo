package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

// encodeBody runs one write into a buffer and returns what it wrote, for the
// bodies below that are laid out by hand in an older version's shape.
func encodeBody(t *testing.T, write func(ms *streams.MinecraftStream) error) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := write(stream); err != nil {
		t.Fatalf("write error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

// The play login as it stands at 1.20.2, after the steps above have had
// their say: the spawn info names the dimension type outright and ends at
// the portal cooldown, and nothing follows it.
func playLogin1_20_2(t *testing.T, deathLocation bool) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteBoolean(true),
			ms.WriteVarInt(2),
			ms.WriteString("minecraft:overworld"),
			ms.WriteString("minecraft:the_end"),
			ms.WriteVarInt(20),
			ms.WriteVarInt(8),
			ms.WriteVarInt(2),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
			ms.WriteBoolean(true),
			ms.WriteString("minecraft:overworld"),
			ms.WriteString("minecraft:the_end"),
			ms.WriteLong(0x1122334455667788),
			ms.WriteByte(2),
			ms.WriteByte(0xFF),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
			ms.WriteBoolean(deathLocation),
		}

		if deathLocation {
			steps = append(steps, ms.WriteString("minecraft:the_nether"), ms.WriteLong(0x0102030405060708))
		}

		steps = append(steps, ms.WriteVarInt(42))

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// The same login as 1.20 lays it out: the game modes right after the
// hardcore flag, the registries after the dimension list, the dimension type
// and name and the seed after them, no limited crafting flag, and the same
// tail.
func playLogin1_20(t *testing.T, registryCodec []byte, deathLocation bool) []byte {
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

		steps = append(steps, ms.WriteVarInt(42))

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// testRegistryCodec stands in for what package gamedata hands the transformer:
// a compound named as a root, the way 1.20 reads every NBT on the wire.
func testRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:dimension_type": nbt.Compound{"type": nbt.String("minecraft:dimension_type")}})
	})
}

func TestDowngradePlayLoginTo1_20ReordersTheFieldsAroundTheRegistries(t *testing.T) {
	codec := testRegistryCodec(t)

	for _, deathLocation := range []bool{false, true} {
		got := runTransformer(t, DowngradePlayLoginTo1_20(codec), playLogin1_20_2(t, deathLocation))
		want := playLogin1_20(t, codec, deathLocation)

		if !bytes.Equal(got, want) {
			t.Errorf("death location %t: to 1.20 = % x\nwant = % x", deathLocation, got, want)
		}
	}
}

// A 1.20 client reads the registries out of this packet and nothing else,
// so a transformer with none to write refuses the login rather than send a
// client into a world it cannot make sense of.
func TestDowngradePlayLoginTo1_20RefusesToSendNoRegistries(t *testing.T) {
	if err := failingTransformer(t, DowngradePlayLoginTo1_20(nil), playLogin1_20_2(t, false)); err == nil {
		t.Error("error = nil, want a refusal for a login with no registries to carry")
	}
}

// What goes out under 1.20's add player id is that packet's body: the entity
// id, the uuid, the position, then the yaw and the pitch -- the other way
// round from the add entity packet -- and nothing of the type, the head yaw,
// the data or the velocity.
func TestDowngradeAddEntityTo1_20BecomesAnAddPlayer(t *testing.T) {
	packet := *testAddEntity
	packet.EntityTypeId = playerEntityType1_20_2

	got := runTransformer(t, DowngradeAddEntityTo1_20, encodeAddEntity1_20_2(t, &packet))

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		for _, err := range []error{
			ms.WriteVarInt(packet.EntityId),
			ms.WriteUuid(packet.Uuid),
			ms.WriteDouble(packet.X),
			ms.WriteDouble(packet.Y),
			ms.WriteDouble(packet.Z),
			ms.WriteByte(play.Angle(packet.Yaw)),
			ms.WriteByte(play.Angle(packet.Pitch)),
		} {
			if err != nil {
				return err
			}
		}

		return nil
	})

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.20 = % x\nwant = % x", got, want)
	}
}

// The add player packet spawns nothing but a player, so an entity of any
// other type is refused rather than shown as one.
func TestDowngradeAddEntityTo1_20RefusesAnythingButThePlayer(t *testing.T) {
	packet := *testAddEntity
	packet.EntityTypeId = playerEntityType1_20_2 + 1

	if err := failingTransformer(t, DowngradeAddEntityTo1_20, encodeAddEntity1_20_2(t, &packet)); err == nil {
		t.Error("error = nil, want a refusal for an entity type 1.20's add player packet cannot spawn")
	}
}

// The add player packet has no velocity, and the rewrite only carries the
// zero this server sends: anything else is refused rather than dropped.
func TestDowngradeAddEntityTo1_20RefusesAVelocity(t *testing.T) {
	packet := *testAddEntity
	packet.EntityTypeId = playerEntityType1_20_2

	body := encodeAddEntity1_20_2(t, &packet)
	body[len(body)-1] = 1

	if err := failingTransformer(t, DowngradeAddEntityTo1_20, body); err == nil {
		t.Error("error = nil, want a refusal for a velocity 1.20's add player packet has no field for")
	}
}

// encodeAddEntity1_20_2 lays the add entity packet out as it stands at
// 1.20.2, after the steps above have moved the velocity to the end as three
// shorts and renumbered the player.
func encodeAddEntity1_20_2(t *testing.T, packet *play.AddEntityClientboundPacket) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		for _, err := range []error{
			ms.WriteVarInt(packet.EntityId),
			ms.WriteUuid(packet.Uuid),
			ms.WriteVarInt(packet.EntityTypeId),
			ms.WriteDouble(packet.X),
			ms.WriteDouble(packet.Y),
			ms.WriteDouble(packet.Z),
			ms.WriteByte(play.Angle(packet.Pitch)),
			ms.WriteByte(play.Angle(packet.Yaw)),
			ms.WriteByte(play.Angle(packet.HeadYaw)),
			ms.WriteVarInt(packet.Data),
			ms.WriteShort(0),
			ms.WriteShort(0),
			ms.WriteShort(0),
		} {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// The chunk packet reaches this step with its heightmaps as the nameless
// compound the 1.21.5 step wrote, and leaves it with the same compound named
// with the empty string, everything else untouched.
func TestDowngradeLevelChunkWithLightTo1_20NamesTheHeightmaps(t *testing.T) {
	heightmaps := nbt.Compound{"MOTION_BLOCKING": nbt.LongArray{1, 2, 3}, "WORLD_SURFACE": nbt.LongArray{4}}
	rest := []byte{0x05, 0x01, 0x02, 0x03, 0x04, 0x05, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00}

	sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteInt(-3); err != nil {
			return err
		}

		if err := ms.WriteInt(9); err != nil {
			return err
		}

		if err := nbt.Write(ms, heightmaps); err != nil {
			return err
		}

		return ms.WriteBytes(rest)
	})

	got := runTransformer(t, DowngradeLevelChunkWithLightTo1_20, sent)

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		if err := ms.WriteInt(-3); err != nil {
			return err
		}

		if err := ms.WriteInt(9); err != nil {
			return err
		}

		if err := nbt.WriteNamed(ms, "", heightmaps); err != nil {
			return err
		}

		return ms.WriteBytes(rest)
	})

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.20 = % x\nwant = % x", got, want)
	}

	// The name is two bytes of zero after the type byte, and that is the
	// whole of the difference.
	if len(got) != len(sent)+2 {
		t.Errorf("the body is %d bytes, want the %d sent plus the empty name", len(got), len(sent))
	}
}

// A 1.20 client sends its uuid as an optional, and what 1.20.2 reads is the
// uuid alone: the flag comes off, and a login without one carries the nil
// uuid in its place.
func TestUpgradeLoginStartFrom1_20DropsTheOptionalFlag(t *testing.T) {
	uuid := "01020304-0506-0708-090a-0b0c0d0e0f10"

	tests := []struct {
		name string
		sent []byte
		want []byte
	}{
		{
			name: "with a uuid",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				if err := ms.WriteBoolean(true); err != nil {
					return err
				}

				return ms.WriteUuid(uuid)
			}),
			want: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				return ms.WriteUuid(uuid)
			}),
		},
		{
			name: "without one",
			sent: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				return ms.WriteBoolean(false)
			}),
			want: encodeBody(t, func(ms *streams.MinecraftStream) error {
				if err := ms.WriteString("Notch"); err != nil {
					return err
				}

				return ms.WriteBytes(make([]byte, uuidSize))
			}),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := runTransformer(t, UpgradeLoginStartFrom1_20, test.sent)

			if !bytes.Equal(got, test.want) {
				t.Errorf("from 1.20 = % x\nwant = % x", got, test.want)
			}
		})
	}
}
