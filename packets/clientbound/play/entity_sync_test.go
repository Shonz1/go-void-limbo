package play

import (
	"bytes"
	"testing"
)

// The uuid the entity packet tests spawn their player under, and the sixteen
// bytes it is on the wire.
const testUuid = "11111111-2222-3333-4444-555555555555"

var testUuidBytes = []byte{
	0x11, 0x11, 0x11, 0x11, 0x22, 0x22, 0x33, 0x33,
	0x44, 0x44, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55,
}

func TestAngle(t *testing.T) {
	tests := []struct {
		degrees float32
		want    byte
	}{
		{degrees: 0, want: 0},
		{degrees: 90, want: 64},
		{degrees: 180, want: 128},
		// A negative angle wraps around the circle rather than clamping.
		{degrees: -90, want: 192},
		{degrees: 360, want: 0},
	}

	for _, test := range tests {
		if got := Angle(test.degrees); got != test.want {
			t.Errorf("Angle(%g) = %d, want %d", test.degrees, got, test.want)
		}
	}
}

func TestAddEntityClientboundPacketEncode(t *testing.T) {
	p := &AddEntityClientboundPacket{
		EntityId:     2,
		Uuid:         testUuid,
		EntityTypeId: PlayerEntityTypeId,
		X:            0.5,
		Y:            64,
		Z:            -0.5,
		Yaw:          90,
		Pitch:        -90,
		HeadYaw:      180,
	}

	want := []byte{0x02}
	want = append(want, testUuidBytes...)
	want = append(want,
		0x9C, 0x01, // the player's entity type, 156 as a var int
		0x3F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x, 0.5
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y, 64
		0xBF, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z, -0.5
		0x00, // the zero velocity, one byte of the quantized vector
		0xC0, // pitch, -90 degrees as an angle byte
		0x40, // yaw, 90
		0x80, // head yaw, 180
		0x00, // data
	)

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote % x, want % x", got, want)
	}
}

func TestEntityPositionSyncClientboundPacketEncode(t *testing.T) {
	p := &EntityPositionSyncClientboundPacket{
		EntityId: 2,
		X:        0.5,
		Y:        64,
		Z:        -0.5,
		Yaw:      180,
		Pitch:    -90,
		OnGround: true,
	}

	want := []byte{
		0x02,
		0x3F, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x, 0.5
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y, 64
		0xBF, 0xE0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z, -0.5
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // delta x, zero
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // delta y
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // delta z
		0x43, 0x34, 0x00, 0x00, // yaw, 180, a float rather than an angle byte
		0xC2, 0xB4, 0x00, 0x00, // pitch, -90
		0x01, // on ground
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote % x, want % x", got, want)
	}
}

func TestRotateHeadClientboundPacketEncode(t *testing.T) {
	p := &RotateHeadClientboundPacket{EntityId: 2, HeadYaw: 180}

	want := []byte{0x02, 0x80}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote % x, want % x", got, want)
	}
}

func TestAnimateClientboundPacketEncode(t *testing.T) {
	p := &AnimateClientboundPacket{EntityId: 2, Animation: AnimationSwingOffhand}

	want := []byte{0x02, 0x03}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote % x, want % x", got, want)
	}
}

func TestRemoveEntitiesClientboundPacketEncode(t *testing.T) {
	p := &RemoveEntitiesClientboundPacket{EntityIds: []int32{2, 128}}

	want := []byte{
		0x02,       // two entities
		0x02,       // the first id
		0x80, 0x01, // the second, 128 as a var int
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote % x, want % x", got, want)
	}
}

func TestPlayerInfoRemoveClientboundPacketEncode(t *testing.T) {
	p := &PlayerInfoRemoveClientboundPacket{Uuids: []string{testUuid}}

	want := append([]byte{0x01}, testUuidBytes...)

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote % x, want % x", got, want)
	}
}

func TestSetEntityDataClientboundPacketEncode(t *testing.T) {
	tests := []struct {
		name      string
		sneaking  bool
		sprinting bool
		flags     byte
		pose      byte
	}{
		{name: "standing", flags: 0x00, pose: 0x00},
		{name: "sneaking", sneaking: true, flags: 0x02, pose: 0x05},
		{name: "sprinting", sprinting: true, flags: 0x08, pose: 0x00},
		{name: "both", sneaking: true, sprinting: true, flags: 0x0A, pose: 0x05},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			p := &SetEntityDataClientboundPacket{EntityId: 2, Sneaking: test.sneaking, Sprinting: test.sprinting}

			want := []byte{
				0x02,                   // the entity
				0x00, 0x00, test.flags, // the flag byte: index, serializer, value
				0x06, 0x14, test.pose, // the pose: index, serializer 20, ordinal
				0xFF, // no more entries
			}

			if got := encode(t, p); !bytes.Equal(got, want) {
				t.Errorf("Encode() wrote % x, want % x", got, want)
			}
		})
	}
}
