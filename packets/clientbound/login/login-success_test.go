package login

import (
	"bytes"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"testing"
)

func encodeLoginSuccess(t *testing.T, p *LoginSuccessClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := p.Encode(stream); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

func TestLoginSuccessClientboundPacketEncode(t *testing.T) {
	signature := "sig"
	p := &LoginSuccessClientboundPacket{
		Profile: types.GameProfile{
			Uuid:     "01020304-0506-0708-090a-0b0c0d0e0f10",
			Username: "Steve",
			Properties: []types.ProfileProperty{
				{Name: "textures", Value: "skin", Signature: &signature},
				{Name: "unsigned", Value: "value"},
			},
		},
		SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20",
	}

	want := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x05, 'S', 't', 'e', 'v', 'e',
		0x02,
		0x08, 't', 'e', 'x', 't', 'u', 'r', 'e', 's',
		0x04, 's', 'k', 'i', 'n',
		0x01,
		0x03, 's', 'i', 'g',
		0x08, 'u', 'n', 's', 'i', 'g', 'n', 'e', 'd',
		0x05, 'v', 'a', 'l', 'u', 'e',
		0x00,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}

	if got := encodeLoginSuccess(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

func TestLoginSuccessClientboundPacketEncodeRejectsInvalidUuid(t *testing.T) {
	p := &LoginSuccessClientboundPacket{
		Profile:   types.GameProfile{Uuid: "not-a-uuid", Username: "Steve"},
		SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20",
	}

	stream := streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))
	if err := p.Encode(stream); err == nil {
		t.Error("Encode() error = nil, want an error for a malformed profile uuid")
	}
}

func TestLoginSuccessClientboundPacketString(t *testing.T) {
	p := &LoginSuccessClientboundPacket{
		Profile: types.GameProfile{
			Uuid:       "01020304-0506-0708-090a-0b0c0d0e0f10",
			Username:   "Steve",
			Properties: []types.ProfileProperty{{Name: "textures", Value: "skin"}},
		},
		SessionId: "11121314-1516-1718-191a-1b1c1d1e1f20",
	}

	want := "LoginSuccessClientboundPacket{Profile:GameProfile{Uuid:01020304-0506-0708-090a-0b0c0d0e0f10 Username:Steve Properties:[ProfileProperty{Name:textures Value:skin}]} SessionId:11121314-1516-1718-191a-1b1c1d1e1f20}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
