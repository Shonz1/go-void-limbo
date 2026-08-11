package status

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"testing"
)

func TestPingRequestServerboundPacketString(t *testing.T) {
	p := &PingRequestServerboundPacket{Payload: 1234}

	want := "PingRequestServerboundPacket{Payload:1234}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}

func TestDecodePingRequestServerboundPacket(t *testing.T) {
	// A plain big-endian long, so the client is free to send a number of any
	// shape and this end reads it back out as it stands.
	tests := []struct {
		name string
		body []byte
		want int64
	}{
		{name: "small", body: []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, want: 1},
		{name: "millisecond clock", body: []byte{0x00, 0x00, 0x01, 0x98, 0x70, 0xa5, 0xe1, 0xd3}, want: 0x0000019870A5E1D3},
		{name: "negative", body: []byte{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}, want: -1},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(test.body))

			packet, err := DecodePingRequestServerboundPacket(stream)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			pingRequest, ok := packet.(*PingRequestServerboundPacket)
			if !ok {
				t.Fatalf("expected *PingRequestServerboundPacket, got %T", packet)
			}

			if pingRequest.Payload != test.want {
				t.Errorf("Payload = %d, want %d", pingRequest.Payload, test.want)
			}
		})
	}
}

func TestDecodePingRequestServerboundPacketReportsATruncatedPayload(t *testing.T) {
	stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer([]byte{0x00, 0x00, 0x00}))

	if _, err := DecodePingRequestServerboundPacket(stream); err == nil {
		t.Error("error = nil, want a payload that is not eight bytes refused")
	}
}
