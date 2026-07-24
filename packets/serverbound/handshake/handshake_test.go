package handshake

import "testing"

func TestHandshakeServerboundPacketString(t *testing.T) {
	p := &HandshakeServerboundPacket{
		ProtocolVersion: 776,
		ServerAddress:   "localhost",
		ServerPort:      25565,
		Intent:          2,
	}

	want := "HandshakeServerboundPacket{ProtocolVersion:776 ServerAddress:localhost ServerPort:25565 Intent:2}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
