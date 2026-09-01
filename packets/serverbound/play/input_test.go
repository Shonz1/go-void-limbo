package play

import "testing"

func TestDecodeSwingServerboundPacket(t *testing.T) {
	tests := []struct {
		name    string
		body    []byte
		offHand bool
	}{
		{name: "main hand", body: []byte{0x00}},
		{name: "off hand", body: []byte{0x01}, offHand: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			packet := decode(t, DecodeSwingServerboundPacket, test.body)

			swing, ok := packet.(*SwingServerboundPacket)
			if !ok {
				t.Fatalf("expected *SwingServerboundPacket, got %T", packet)
			}

			if swing.OffHand != test.offHand {
				t.Errorf("OffHand = %t, want %t", swing.OffHand, test.offHand)
			}
		})
	}
}

func TestDecodePlayerInputServerboundPacketReadsEachFlagOnItsOwn(t *testing.T) {
	tests := []struct {
		flags byte
		want  PlayerInputServerboundPacket
	}{
		{flags: 0x00, want: PlayerInputServerboundPacket{}},
		{flags: 0x01, want: PlayerInputServerboundPacket{Forward: true}},
		{flags: 0x02, want: PlayerInputServerboundPacket{Backward: true}},
		{flags: 0x04, want: PlayerInputServerboundPacket{Left: true}},
		{flags: 0x08, want: PlayerInputServerboundPacket{Right: true}},
		{flags: 0x10, want: PlayerInputServerboundPacket{Jump: true}},
		{flags: 0x20, want: PlayerInputServerboundPacket{Sneak: true}},
		{flags: 0x40, want: PlayerInputServerboundPacket{Sprint: true}},
		{flags: 0x7F, want: PlayerInputServerboundPacket{Forward: true, Backward: true, Left: true, Right: true, Jump: true, Sneak: true, Sprint: true}},
	}

	for _, test := range tests {
		packet := decode(t, DecodePlayerInputServerboundPacket, []byte{test.flags})

		if packet.String() != test.want.String() {
			t.Errorf("flags %#02x: decoded %s, want %s", test.flags, packet, test.want.String())
		}
	}
}
