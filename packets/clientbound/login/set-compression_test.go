package login

import (
	"bytes"
	"go-void-limbo/streams"
	"testing"
)

func TestEncodeSetCompressionClientboundPacket(t *testing.T) {
	// The threshold is a var int, so the vanilla 256 is two bytes rather than
	// the four an int would be.
	tests := []struct {
		name      string
		threshold int32
		want      []byte
	}{
		{name: "everything", threshold: 0, want: []byte{0x00}},
		{name: "vanilla", threshold: 256, want: []byte{0x80, 0x02}},
		{name: "single byte", threshold: 127, want: []byte{0x7f}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			stream := streams.NewMinecraftStreamFromBuffer(buf)

			packet := &SetCompressionClientboundPacket{Threshold: test.threshold}

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
