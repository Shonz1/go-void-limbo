package login

import (
	"bytes"
	"go-void-limbo/streams"
	"testing"
)

func TestEncodeLoginPluginRequestClientboundPacket(t *testing.T) {
	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	packet := &LoginPluginRequestClientboundPacket{MessageId: 1, Channel: "velocity:player_info", Data: []byte{0x01}}

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("encode error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("flush error = %v", err)
	}

	// The message id, the channel behind its length, and then the data with no
	// length in front of it: it runs to the end of the packet, which is what
	// makes it the last field.
	want := append([]byte{0x01, 0x14}, "velocity:player_info"...)
	want = append(want, 0x01)

	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("encoded % x, want % x", buf.Bytes(), want)
	}
}

func TestLoginPluginRequestClientboundPacketString(t *testing.T) {
	packet := &LoginPluginRequestClientboundPacket{MessageId: 1, Channel: "velocity:player_info", Data: []byte{0x01}}

	want := "LoginPluginRequestClientboundPacket{MessageId:1 Channel:velocity:player_info Data:1 bytes}"
	if got := packet.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
