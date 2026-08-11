package play

import (
	"bytes"
	"testing"
)

func TestPlayerPositionClientboundPacketEncode(t *testing.T) {
	p := &PlayerPositionClientboundPacket{
		TeleportId: 1,
		X:          0.5,
		Y:          64,
		Z:          -0.5,
		DeltaX:     0,
		DeltaY:     0,
		DeltaZ:     0,
		Yaw:        180,
		Pitch:      -90,
		Relatives:  0,
	}

	want := []byte{
		0x01,                                           // teleport id
		0x3f, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x
		0x40, 0x50, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // y
		0xbf, 0xe0, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // z
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // delta x
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // delta y
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // delta z
		0x43, 0x34, 0x00, 0x00, // yaw
		0xc2, 0xb4, 0x00, 0x00, // pitch
		0x00, 0x00, 0x00, 0x00, // relatives, a plain int
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

func TestPlayerPositionClientboundPacketEncodeRelatives(t *testing.T) {
	// The mask is sent as it is given, so only its width is this packet's
	// concern; which bit means what is the client's.
	p := &PlayerPositionClientboundPacket{Relatives: 0x0000FF01}

	got := encode(t, p)
	if want := []byte{0x00, 0x00, 0xff, 0x01}; !bytes.Equal(got[len(got)-4:], want) {
		t.Errorf("Encode() wrote relatives %v, want %v", got[len(got)-4:], want)
	}
}
