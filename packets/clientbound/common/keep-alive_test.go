package common

import (
	"bytes"
	"go-void-limbo/streams"
	"testing"
)

func encode(t *testing.T, packet *KeepAliveClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("encode error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("flush error = %v", err)
	}

	return buf.Bytes()
}

func TestEncodeKeepAliveClientboundPacket(t *testing.T) {
	// The id is a plain big-endian long, not a var int, so every id is eight
	// bytes wide and a negative one is written as it stands.
	tests := []struct {
		name string
		id   int64
		want []byte
	}{
		{name: "small", id: 1, want: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{name: "millisecond clock", id: 0x0000019870A5E1D3, want: []byte{0x00, 0x00, 0x01, 0x98, 0x70, 0xa5, 0xe1, 0xd3}},
		{name: "negative", id: -1, want: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := encode(t, &KeepAliveClientboundPacket{Id: test.id})

			if !bytes.Equal(got, test.want) {
				t.Errorf("encoded % x, want % x", got, test.want)
			}
		})
	}
}
