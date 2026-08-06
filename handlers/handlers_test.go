package handlers

import (
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
	"testing"
)

type fakeClient struct {
	protocolVersion types.ProtocolVersion
	phase           types.Phase
	written         []types.ClientboundPacket
}

func (c *fakeClient) ProtocolVersion() types.ProtocolVersion { return c.protocolVersion }

func (c *fakeClient) SetProtocolVersion(protocolVersion types.ProtocolVersion) {
	c.protocolVersion = protocolVersion
}

func (c *fakeClient) Phase() types.Phase { return c.phase }

func (c *fakeClient) SetPhase(phase types.Phase) { c.phase = phase }

func (c *fakeClient) WritePacket(packet types.ClientboundPacket) error {
	c.written = append(c.written, packet)
	return nil
}

func TestHandleHandshakeServerboundPacket(t *testing.T) {
	client := &fakeClient{}
	packet := &handshake.HandshakeServerboundPacket{
		ProtocolVersion: int32(types.ProtocolVersions.MINECRAFT_26_2.ID),
		ServerAddress:   "localhost",
		ServerPort:      25565,
		Intent:          int32(types.PhaseLogin),
	}

	if err := HandleHandshakeServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.ProtocolVersion().ID != types.ProtocolVersions.MINECRAFT_26_2.ID {
		t.Errorf("expected protocol version %d, got %d", types.ProtocolVersions.MINECRAFT_26_2.ID, client.ProtocolVersion().ID)
	}

	if client.Phase() != types.PhaseLogin {
		t.Errorf("expected phase %d, got %d", types.PhaseLogin, client.Phase())
	}
}

func TestHandleLoginStartServerboundPacketWritesLoginSuccess(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin}
	packet := &login.LoginStartServerboundPacket{Name: "Notch", Uuid: "00000000-0000-0000-0000-000000000001"}

	if err := HandleLoginStartServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	loginSuccess, ok := client.written[0].(*clientboundLogin.LoginSuccessClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginSuccessClientboundPacket, got %T", client.written[0])
	}

	if loginSuccess.Profile.Username != packet.Name {
		t.Errorf("expected username %q, got %q", packet.Name, loginSuccess.Profile.Username)
	}

	if loginSuccess.Profile.Uuid != packet.Uuid {
		t.Errorf("expected uuid %q, got %q", packet.Uuid, loginSuccess.Profile.Uuid)
	}

	if loginSuccess.SessionId == "" {
		t.Error("expected a generated session id, got an empty string")
	}

	if loginSuccess.SessionId == loginSuccess.Profile.Uuid {
		t.Error("expected the session id to differ from the profile uuid")
	}

	// The client stays in the login phase until it acknowledges the success packet.
	if client.Phase() != types.PhaseLogin {
		t.Errorf("expected phase %d, got %d", types.PhaseLogin, client.Phase())
	}
}

func TestHandlersRejectUnexpectedPacketType(t *testing.T) {
	client := &fakeClient{}

	if err := HandleHandshakeServerboundPacket(client, &login.LoginStartServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleLoginStartServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}
}
