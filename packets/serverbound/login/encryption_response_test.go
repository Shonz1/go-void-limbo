package login

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"testing"
)

func TestDecodeEncryptionResponseServerboundPacket(t *testing.T) {
	body := []byte{0x03, 0xAA, 0xBB, 0xCC, 0x04, 0x01, 0x02, 0x03, 0x04}

	stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	packet, err := DecodeEncryptionResponseServerboundPacket(stream)
	if err != nil {
		t.Fatalf("decode error = %v", err)
	}

	response, ok := packet.(*EncryptionResponseServerboundPacket)
	if !ok {
		t.Fatalf("decoded %T, want an encryption response", packet)
	}

	if !bytes.Equal(response.SharedSecret, []byte{0xAA, 0xBB, 0xCC}) {
		t.Errorf("shared secret = % x, want aa bb cc", response.SharedSecret)
	}

	if !bytes.Equal(response.VerifyToken, []byte{0x01, 0x02, 0x03, 0x04}) {
		t.Errorf("verify token = % x, want 01 02 03 04", response.VerifyToken)
	}
}

// Both lengths arrive from the connection and are what the buffers for them are
// allocated from, before a byte of either has been read.
func TestDecodeEncryptionResponseServerboundPacketRefusesALengthItCannotHold(t *testing.T) {
	tests := []struct {
		name string
		body []byte
	}{
		{name: "a shared secret longer than a ciphertext", body: []byte{0x80, 0x80, 0x80, 0x01}},
		{name: "a verify token longer than a ciphertext", body: []byte{0x01, 0xAA, 0x80, 0x80, 0x80, 0x01}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(test.body))

			if _, err := DecodeEncryptionResponseServerboundPacket(stream); err == nil {
				t.Error("error = nil, want a length no ciphertext could be refused")
			}
		})
	}
}

// The ciphertext says nothing worth reading, and what is inside it is the key to
// the connection.
func TestEncryptionResponseServerboundPacketStringKeepsTheFieldsToItself(t *testing.T) {
	packet := &EncryptionResponseServerboundPacket{SharedSecret: make([]byte, 128), VerifyToken: make([]byte, 128)}

	want := "EncryptionResponseServerboundPacket{SharedSecret:128 bytes VerifyToken:128 bytes}"
	if got := packet.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
