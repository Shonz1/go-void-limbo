package login

import "testing"

func TestDisconnectClientboundPacketString(t *testing.T) {
	p := &DisconnectClientboundPacket{Reason: `{"text": "TODO"}`}

	want := `DisconnectClientboundPacket{Reason:{"text": "TODO"}}`
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
