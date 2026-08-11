package login

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"testing"
)

func TestLoginAcknowledgedServerboundPacketString(t *testing.T) {
	p := &LoginAcknowledgedServerboundPacket{}

	want := "LoginAcknowledgedServerboundPacket{}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestDecodeLoginAcknowledgedServerboundPacketConsumesNothing(t *testing.T) {
	buf := bytes.NewBuffer([]byte{0xff})
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	packet, err := DecodeLoginAcknowledgedServerboundPacket(stream)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := packet.(*LoginAcknowledgedServerboundPacket); !ok {
		t.Fatalf("expected *LoginAcknowledgedServerboundPacket, got %T", packet)
	}

	// The packet has no body, so the trailing byte must still be readable.
	got, err := stream.ReadByte()
	if err != nil {
		t.Fatalf("unexpected error reading after decode: %v", err)
	}

	if got != 0xff {
		t.Errorf("expected the decoder to leave the stream untouched, read %#x", got)
	}
}
