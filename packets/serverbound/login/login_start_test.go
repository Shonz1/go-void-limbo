package login

import "testing"

func TestLoginStartServerboundPacketString(t *testing.T) {
	p := &LoginStartServerboundPacket{
		Name: "Steve",
		Uuid: "01020304-0506-0708-090a-0b0c0d0e0f10",
	}

	want := "LoginStartServerboundPacket{Name:Steve Uuid:01020304-0506-0708-090a-0b0c0d0e0f10}"
	if got := p.String(); got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
}
