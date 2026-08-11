package login

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"testing"
)

func TestDecodeLoginPluginResponseServerboundPacket(t *testing.T) {
	body := append([]byte{0x01, 0x01}, "a signed payload"...)

	packet, err := DecodeLoginPluginResponseServerboundPacket(streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body)))
	if err != nil {
		t.Fatalf("decode error = %v", err)
	}

	response, ok := packet.(*LoginPluginResponseServerboundPacket)
	if !ok {
		t.Fatalf("decoded %T, want *LoginPluginResponseServerboundPacket", packet)
	}

	if response.MessageId != 1 {
		t.Errorf("message id = %d, want the one the request went out under", response.MessageId)
	}

	if !response.Successful {
		t.Error("successful = false, want the answer of somebody who knows the channel")
	}

	// The payload runs to the end of the packet rather than sitting behind a
	// length, so what is read is everything that was left.
	if string(response.Data) != "a signed payload" {
		t.Errorf("data = %q, want everything after the flag", response.Data)
	}
}

// What a client that has never heard of the channel answers with: the flag, and
// nothing behind it.
func TestDecodeLoginPluginResponseServerboundPacketReadsAnAnswerThatCarriesNothing(t *testing.T) {
	packet, err := DecodeLoginPluginResponseServerboundPacket(streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer([]byte{0x01, 0x00})))
	if err != nil {
		t.Fatalf("decode error = %v", err)
	}

	response, ok := packet.(*LoginPluginResponseServerboundPacket)
	if !ok {
		t.Fatalf("decoded %T, want *LoginPluginResponseServerboundPacket", packet)
	}

	if response.Successful {
		t.Error("successful = true, want the answer of somebody who does not know the channel")
	}

	if len(response.Data) != 0 {
		t.Errorf("data = %q, want nothing behind an answer that carried nothing", response.Data)
	}
}

func TestDecodeLoginPluginResponseServerboundPacketReportsABodyThatEndsEarly(t *testing.T) {
	// The message id, and then the flag that never arrived.
	if _, err := DecodeLoginPluginResponseServerboundPacket(streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer([]byte{0x01}))); err == nil {
		t.Error("error = nil, want a body that ended before the answer did reported")
	}
}

// The payload is a login and a signature over it, and neither belongs in a log
// line.
func TestLoginPluginResponseServerboundPacketStringKeepsThePayloadToItself(t *testing.T) {
	packet := &LoginPluginResponseServerboundPacket{MessageId: 1, Successful: true, Data: []byte("a signed payload")}

	want := "LoginPluginResponseServerboundPacket{MessageId:1 Successful:true Data:16 bytes}"
	if got := packet.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
