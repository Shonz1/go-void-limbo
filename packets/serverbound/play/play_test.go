package play

import (
	"bytes"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"testing"
)

func decode(t *testing.T, decoder func(*streams.MinecraftStream) (types.ServerboundPacket, error), body []byte) types.ServerboundPacket {
	t.Helper()

	stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	packet, err := decoder(stream)
	if err != nil {
		t.Fatalf("decode error = %v", err)
	}

	return packet
}

func TestDecodeAcceptTeleportationServerboundPacket(t *testing.T) {
	packet := decode(t, DecodeAcceptTeleportationServerboundPacket, []byte{0x80, 0x01})

	acceptTeleportation, ok := packet.(*AcceptTeleportationServerboundPacket)
	if !ok {
		t.Fatalf("expected *AcceptTeleportationServerboundPacket, got %T", packet)
	}

	if acceptTeleportation.TeleportId != 128 {
		t.Errorf("TeleportId = %d, want 128", acceptTeleportation.TeleportId)
	}
}

func TestDecodeAcceptTeleportationServerboundPacketRejectsTruncatedBody(t *testing.T) {
	stream := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))

	if _, err := DecodeAcceptTeleportationServerboundPacket(stream); err == nil {
		t.Error("error = nil, want an error for a body with no teleport id in it")
	}
}

func TestDecodeEmptyBodiedPacketsConsumeNothing(t *testing.T) {
	tests := []struct {
		name    string
		decoder func(*streams.MinecraftStream) (types.ServerboundPacket, error)
		want    types.ServerboundPacket
	}{
		{name: "player loaded", decoder: DecodePlayerLoadedServerboundPacket, want: &PlayerLoadedServerboundPacket{}},
		{name: "client tick end", decoder: DecodeClientTickEndServerboundPacket, want: &ClientTickEndServerboundPacket{}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer([]byte{0xff}))

			packet, err := test.decoder(stream)
			if err != nil {
				t.Fatalf("decode error = %v", err)
			}

			if packet.String() != test.want.String() {
				t.Errorf("decoded %s, want %s", packet, test.want)
			}

			got, err := stream.ReadByte()
			if err != nil {
				t.Fatalf("unexpected error reading after decode: %v", err)
			}

			if got != 0xff {
				t.Errorf("expected the decoder to leave the stream untouched, read %#x", got)
			}
		})
	}
}

// The four move player packets differ only in which of position and rotation
// they carry, and all end in the same flag byte.
func TestDecodeMovePlayerServerboundPackets(t *testing.T) {
	position := []byte{
		0x3f, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x, 0.5
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y, 64
		0xbf, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z, -0.5
	}
	rotation := []byte{
		0x43, 0x34, 0x00, 0x00, // yaw, 180
		0xc2, 0xb4, 0x00, 0x00, // pitch, -90
	}
	// On the ground and against a wall, so both flags are set.
	flags := []byte{0x03}

	tests := []struct {
		name    string
		decoder func(*streams.MinecraftStream) (types.ServerboundPacket, error)
		body    []byte
		want    types.ServerboundPacket
	}{
		{
			name:    "position",
			decoder: DecodeMovePlayerPositionServerboundPacket,
			body:    append(append([]byte{}, position...), flags...),
			want: &MovePlayerPositionServerboundPacket{
				X: 0.5, Y: 64, Z: -0.5,
				MovePlayerStatus: MovePlayerStatus{OnGround: true, HorizontalCollision: true},
			},
		},
		{
			name:    "position and rotation",
			decoder: DecodeMovePlayerPositionRotationServerboundPacket,
			body:    append(append(append([]byte{}, position...), rotation...), flags...),
			want: &MovePlayerPositionRotationServerboundPacket{
				X: 0.5, Y: 64, Z: -0.5, Yaw: 180, Pitch: -90,
				MovePlayerStatus: MovePlayerStatus{OnGround: true, HorizontalCollision: true},
			},
		},
		{
			name:    "rotation",
			decoder: DecodeMovePlayerRotationServerboundPacket,
			body:    append(append([]byte{}, rotation...), flags...),
			want: &MovePlayerRotationServerboundPacket{
				Yaw: 180, Pitch: -90,
				MovePlayerStatus: MovePlayerStatus{OnGround: true, HorizontalCollision: true},
			},
		},
		{
			name:    "status only",
			decoder: DecodeMovePlayerStatusServerboundPacket,
			body:    flags,
			want:    &MovePlayerStatusServerboundPacket{MovePlayerStatus: MovePlayerStatus{OnGround: true, HorizontalCollision: true}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packet := decode(t, test.decoder, test.body)

			if packet.String() != test.want.String() {
				t.Errorf("decoded %s, want %s", packet, test.want)
			}
		})
	}
}

func TestDecodeMovePlayerStatusServerboundPacketReadsEachFlagOnItsOwn(t *testing.T) {
	tests := []struct {
		flags               byte
		onGround            bool
		horizontalCollision bool
	}{
		{flags: 0x00},
		{flags: 0x01, onGround: true},
		{flags: 0x02, horizontalCollision: true},
		{flags: 0x03, onGround: true, horizontalCollision: true},
	}

	for _, test := range tests {
		packet := decode(t, DecodeMovePlayerStatusServerboundPacket, []byte{test.flags})

		status, ok := packet.(*MovePlayerStatusServerboundPacket)
		if !ok {
			t.Fatalf("expected *MovePlayerStatusServerboundPacket, got %T", packet)
		}

		if status.OnGround != test.onGround {
			t.Errorf("flags %#02x: OnGround = %t, want %t", test.flags, status.OnGround, test.onGround)
		}

		if status.HorizontalCollision != test.horizontalCollision {
			t.Errorf("flags %#02x: HorizontalCollision = %t, want %t", test.flags, status.HorizontalCollision, test.horizontalCollision)
		}
	}
}
