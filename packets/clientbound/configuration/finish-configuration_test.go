package configuration

import (
	"bytes"
	"go-void-limbo/streams"
	"testing"
)

func TestFinishConfigurationClientboundPacketString(t *testing.T) {
	p := &FinishConfigurationClientboundPacket{}

	want := "FinishConfigurationClientboundPacket{}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestFinishConfigurationClientboundPacketEncodeWritesNothing(t *testing.T) {
	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	p := &FinishConfigurationClientboundPacket{}
	if err := p.Encode(stream); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	if buf.Len() != 0 {
		t.Errorf("Encode() wrote %v, want an empty body", buf.Bytes())
	}
}
