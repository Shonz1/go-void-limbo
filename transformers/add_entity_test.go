package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

// encodeAddEntity produces the body the server actually sends, at the latest
// version, which is what every downgrade below is fed.
func encodeAddEntity(t *testing.T, packet *play.AddEntityClientboundPacket) []byte {
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

func runTransformer(t *testing.T, transformer func(in, out *streams.MinecraftStream) error, body []byte) []byte {
	t.Helper()

	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := transformer(in, out); err != nil {
		t.Fatalf("transformer error = %v", err)
	}

	if err := out.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

var testAddEntity = &play.AddEntityClientboundPacket{
	EntityId:     2,
	Uuid:         "11111111-2222-3333-4444-555555555555",
	EntityTypeId: play.PlayerEntityTypeId,
	X:            0.5,
	Y:            64,
	Z:            -0.5,
	Yaw:          90,
	Pitch:        -90,
	HeadYaw:      180,
}

// The registry steps rewrite the one var int and touch nothing else, so the
// downgraded body is the original with two bytes swapped: 156 and 155 are both
// two byte var ints, and 151 is one.
func TestDowngradeAddEntityRenumbersThePlayer(t *testing.T) {
	body := encodeAddEntity(t, testAddEntity)

	to26_1 := runTransformer(t, DowngradeAddEntityTo26_1, body)

	want := append(append([]byte{}, body[:17]...), 0x9B, 0x01) // 155 as a var int
	want = append(want, body[19:]...)

	if !bytes.Equal(to26_1, want) {
		t.Errorf("to 26.1 = % x, want % x", to26_1, want)
	}

	// 26.1 and 1.21.11 number the player alike, so the next step down starts
	// from the same body and only then does the number move again.
	to1_21_9 := runTransformer(t, DowngradeAddEntityTo1_21_9, want)

	want9 := append(append([]byte{}, body[:17]...), 0x97, 0x01) // 151
	want9 = append(want9, body[19:]...)

	if !bytes.Equal(to1_21_9, want9) {
		t.Errorf("to 1.21.9 = % x, want % x", to1_21_9, want9)
	}
}

func TestDowngradeAddEntityTo1_21_7MovesTheVelocityBack(t *testing.T) {
	body := runTransformer(t, DowngradeAddEntityTo1_21_9,
		runTransformer(t, DowngradeAddEntityTo26_1, encodeAddEntity(t, testAddEntity)))

	got := runTransformer(t, DowngradeAddEntityTo1_21_7, body)

	want := []byte{0x02}
	want = append(want,
		0x11, 0x11, 0x11, 0x11, 0x22, 0x22, 0x33, 0x33,
		0x44, 0x44, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55,
	)
	want = append(want,
		0x95, 0x01, // the player's entity type, 149
		0x3F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x, 0.5
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y, 64
		0xBF, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z, -0.5
		// No quantized velocity here: 1.21.7 reads it as shorts at the end.
		0xC0,       // pitch, -90 degrees as an angle byte
		0x40,       // yaw, 90
		0x80,       // head yaw, 180
		0x00,       // data
		0x00, 0x00, // velocity x, the zero the vector carried
		0x00, 0x00, // velocity y
		0x00, 0x00, // velocity z
	)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.21.7 = % x, want % x", got, want)
	}

	// 1.21.5 reads the same shape; only the player's number moves, down to
	// 148, which is still a two byte var int.
	to1_21_5 := runTransformer(t, DowngradeAddEntityTo1_21_5, got)

	want5 := append(append([]byte{}, want[:17]...), 0x94, 0x01)
	want5 = append(want5, want[19:]...)

	if !bytes.Equal(to1_21_5, want5) {
		t.Errorf("to 1.21.5 = % x, want % x", to1_21_5, want5)
	}

	// And 1.21.4 the same shape once more, with the player at 147.
	to1_21_4 := runTransformer(t, DowngradeAddEntityTo1_21_4, to1_21_5)

	want4 := append(append([]byte{}, want[:17]...), 0x93, 0x01)
	want4 = append(want4, want[19:]...)

	if !bytes.Equal(to1_21_4, want4) {
		t.Errorf("to 1.21.4 = % x, want % x", to1_21_4, want4)
	}

	// And 1.21.2 the same shape, with the player back up at 148: the
	// transient creaking still sat in front of it there.
	to1_21_2 := runTransformer(t, DowngradeAddEntityTo1_21_2, to1_21_4)

	if !bytes.Equal(to1_21_2, want5) {
		t.Errorf("to 1.21.2 = % x, want % x", to1_21_2, want5)
	}

	// And 1.21 the same shape at the bottom of the chain, with the player
	// down at 128 -- 1.21.2 added twenty entities in front of it -- which is
	// still a two byte var int.
	to1_21 := runTransformer(t, DowngradeAddEntityTo1_21, to1_21_2)

	want1 := append(append([]byte{}, want[:17]...), 0x80, 0x01)
	want1 = append(want1, want[19:]...)

	if !bytes.Equal(to1_21, want1) {
		t.Errorf("to 1.21 = % x, want % x", to1_21, want1)
	}
}

func TestDowngradeAddEntityRefusesAnEntityItDoesNotKnow(t *testing.T) {
	notAPlayer := &play.AddEntityClientboundPacket{
		EntityId:     2,
		Uuid:         "11111111-2222-3333-4444-555555555555",
		EntityTypeId: 1,
	}

	body := encodeAddEntity(t, notAPlayer)

	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))
	out := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))

	if err := DowngradeAddEntityTo26_1(in, out); err == nil {
		t.Error("error = nil, want an error for an entity type the mapping does not cover")
	}
}

func TestDowngradeSetEntityDataRenamesThePoseSerializer(t *testing.T) {
	stance := &play.SetEntityDataClientboundPacket{EntityId: 2, Sneaking: true}

	body := encodeAddEntityData(t, stance)

	got := runTransformer(t, DowngradeSetEntityDataTo1_21_7, body)

	want := []byte{
		0x02,             // the entity
		0x00, 0x00, 0x02, // the flag byte: index, serializer, sneaking
		0x06, 0x15, 0x05, // the pose: index, serializer 21 rather than 20, crouching
		0xFF, // no more entries
	}

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.21.7 = % x, want % x", got, want)
	}
}

func TestDowngradeSetEntityDataRefusesASerializerItCannotWalk(t *testing.T) {
	// An entry naming serializer 1, an int, whose value this rewrite has no
	// idea how to get past.
	body := []byte{0x02, 0x00, 0x01, 0x00, 0xFF}

	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))
	out := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))

	if err := DowngradeSetEntityDataTo1_21_7(in, out); err == nil {
		t.Error("error = nil, want an error for a serializer the rewrite does not know the shape of")
	}
}

func encodeAddEntityData(t *testing.T, packet *play.SetEntityDataClientboundPacket) []byte {
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
