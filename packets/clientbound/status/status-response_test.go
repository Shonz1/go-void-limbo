package status

import (
	"bytes"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"testing"
)

func encodeStatus(t *testing.T, packet *StatusResponseClientboundPacket) *streams.MinecraftStream {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("encode error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("flush error = %v", err)
	}

	return streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(buf.Bytes()))
}

// The document the client actually parses, field for field. The names are the
// client's and the nesting is the client's, so this is the one packet where a
// test that only checked the fields went in would be checking nothing.
func TestEncodeStatusResponseClientboundPacket(t *testing.T) {
	packet := &StatusResponseClientboundPacket{
		Status: types.ServerStatus{
			Version:     types.ServerVersion{Name: "26.2", Protocol: 776},
			Players:     types.ServerPlayers{Online: 3, Max: 4},
			Description: types.TextComponent{Text: "A void limbo"},
		},
	}

	// One string, length first, which is what makes a document the client cannot
	// parse a packet it reads to the end of anyway.
	document, err := encodeStatus(t, packet).ReadString()
	if err != nil {
		t.Fatalf("reading the document: %v", err)
	}

	want := `{"version":{"name":"26.2","protocol":776},"players":{"max":4,"online":3},"description":{"text":"A void limbo"}}`
	if document != want {
		t.Errorf("encoded %s, want %s", document, want)
	}
}

func TestEncodeStatusResponseClientboundPacketEscapesTheDescription(t *testing.T) {
	// A description is whatever the operator typed, and it lands inside a JSON
	// string. A quote or a line break that went in as it stands would be a
	// document the client gives up on, and the server list would show it as
	// unreachable rather than as badly described.
	packet := &StatusResponseClientboundPacket{
		Status: types.ServerStatus{Description: types.TextComponent{Text: "a \"void\"\nlimbo"}},
	}

	document, err := encodeStatus(t, packet).ReadString()
	if err != nil {
		t.Fatalf("reading the document: %v", err)
	}

	want := `{"version":{"name":"","protocol":0},"players":{"max":0,"online":0},"description":{"text":"a \"void\"\nlimbo"}}`
	if document != want {
		t.Errorf("encoded %s, want %s", document, want)
	}
}

func TestStatusResponseClientboundPacketString(t *testing.T) {
	p := &StatusResponseClientboundPacket{
		Status: types.ServerStatus{
			Version:     types.ServerVersion{Name: "26.2", Protocol: 776},
			Players:     types.ServerPlayers{Online: 3, Max: 4},
			Description: types.TextComponent{Text: "A void limbo"},
		},
	}

	want := "StatusResponseClientboundPacket{Version:26.2 Protocol:776 Players:3/4 Description:A void limbo}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
