package login

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"testing"
)

func TestEncodeEncryptionRequestClientboundPacket(t *testing.T) {
	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	packet := &EncryptionRequestClientboundPacket{
		PublicKey:          []byte{0xAA, 0xBB, 0xCC},
		VerifyToken:        []byte{0x01, 0x02, 0x03, 0x04},
		ShouldAuthenticate: true,
	}

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("encode error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("flush error = %v", err)
	}

	// An empty server id, then each of the two fields behind its own length, and
	// the flag that asks the client to tell Mojang it joined.
	want := []byte{0x00, 0x03, 0xAA, 0xBB, 0xCC, 0x04, 0x01, 0x02, 0x03, 0x04, 0x01}
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("encoded % x, want % x", buf.Bytes(), want)
	}
}

// The key is a few hundred bytes of nothing anyone can read, and the token is
// the one thing on the packet worth keeping to itself.
func TestEncryptionRequestClientboundPacketStringKeepsTheFieldsToItself(t *testing.T) {
	packet := &EncryptionRequestClientboundPacket{PublicKey: make([]byte, 162), VerifyToken: []byte{0x01, 0x02, 0x03, 0x04}}

	want := "EncryptionRequestClientboundPacket{ServerId: PublicKey:162 bytes VerifyToken:4 bytes ShouldAuthenticate:false}"
	if got := packet.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
