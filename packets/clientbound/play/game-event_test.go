package play

import (
	"bytes"
	"testing"
)

func TestGameEventClientboundPacketEncode(t *testing.T) {
	p := &GameEventClientboundPacket{Event: GameEventStartWaitingForChunks}

	want := []byte{
		0x0d,                   // start waiting for chunks
		0x00, 0x00, 0x00, 0x00, // its unused value
	}

	if got := encode(t, p); !bytes.Equal(got, want) {
		t.Errorf("Encode() wrote %v, want %v", got, want)
	}
}

func TestGameEventClientboundPacketString(t *testing.T) {
	p := &GameEventClientboundPacket{Event: GameEventStartWaitingForChunks}

	want := "GameEventClientboundPacket{Event:start_waiting_for_chunks Value:0}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
