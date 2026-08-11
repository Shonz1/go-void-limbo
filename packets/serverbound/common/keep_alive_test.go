package common

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"testing"
)

func TestDecodeKeepAliveServerboundPacket(t *testing.T) {
	body := []byte{0x00, 0x00, 0x01, 0x98, 0x70, 0xa5, 0xe1, 0xd3}
	stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	packet, err := DecodeKeepAliveServerboundPacket(stream)
	if err != nil {
		t.Fatalf("decode error = %v", err)
	}

	keepAlive, ok := packet.(*KeepAliveServerboundPacket)
	if !ok {
		t.Fatalf("expected *KeepAliveServerboundPacket, got %T", packet)
	}

	// The id has to come back exactly as it went out, since matching it against
	// what was sent is the only thing the packet is for.
	if keepAlive.Id != 0x0000019870A5E1D3 {
		t.Errorf("Id = %d, want %d", keepAlive.Id, 0x0000019870A5E1D3)
	}
}

func TestDecodeKeepAliveServerboundPacketRejectsTruncatedBody(t *testing.T) {
	// Seven bytes of an eight byte id. Taking it for a whole one would answer a
	// keep alive that was never sent.
	stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}))

	if _, err := DecodeKeepAliveServerboundPacket(stream); err == nil {
		t.Error("error = nil, want an error for a body with no whole id in it")
	}
}
