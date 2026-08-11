package status

import (
	"bytes"
	"go-void-limbo/streams"
	"testing"
)

func TestEncodePongResponseClientboundPacket(t *testing.T) {
	// The payload is a plain big-endian long, and whatever the client sent is
	// what goes back, so every value is eight bytes of exactly what arrived.
	tests := []struct {
		name    string
		payload int64
		want    []byte
	}{
		{name: "small", payload: 1, want: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{name: "millisecond clock", payload: 0x0000019870A5E1D3, want: []byte{0x00, 0x00, 0x01, 0x98, 0x70, 0xa5, 0xe1, 0xd3}},
		{name: "negative", payload: -1, want: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			stream := streams.NewMinecraftStreamFromBuffer(buf)

			packet := &PongResponseClientboundPacket{Payload: test.payload}

			if err := packet.Encode(stream); err != nil {
				t.Fatalf("encode error = %v", err)
			}

			if err := stream.Flush(); err != nil {
				t.Fatalf("flush error = %v", err)
			}

			if !bytes.Equal(buf.Bytes(), test.want) {
				t.Errorf("encoded % x, want % x", buf.Bytes(), test.want)
			}
		})
	}
}

func TestPongResponseClientboundPacketString(t *testing.T) {
	p := &PongResponseClientboundPacket{Payload: 1234}

	want := "PongResponseClientboundPacket{Payload:1234}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
