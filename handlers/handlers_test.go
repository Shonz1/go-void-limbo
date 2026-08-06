package handlers

import (
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
	"testing"
)

type fakeClient struct {
	protocolVersion types.ProtocolVersion
	phase           types.Phase
	written         []types.ClientboundPacket
	// writePhases records the phase the client was in as each packet was
	// written, since that phase is what the real client resolves a clientbound
	// packet id from.
	writePhases []types.Phase
}

func (c *fakeClient) ProtocolVersion() types.ProtocolVersion { return c.protocolVersion }

func (c *fakeClient) SetProtocolVersion(protocolVersion types.ProtocolVersion) {
	c.protocolVersion = protocolVersion
}

func (c *fakeClient) Phase() types.Phase { return c.phase }

func (c *fakeClient) SetPhase(phase types.Phase) { c.phase = phase }

func (c *fakeClient) WritePacket(packet types.ClientboundPacket) error {
	c.written = append(c.written, packet)
	c.writePhases = append(c.writePhases, c.phase)
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

func TestHandleLoginAcknowledgedServerboundPacketFinishesConfiguration(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin}

	if err := HandleLoginAcknowledgedServerboundPacket(client, &login.LoginAcknowledgedServerboundPacket{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Phase() != types.PhaseConfiguration {
		t.Errorf("expected phase %d, got %d", types.PhaseConfiguration, client.Phase())
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	if _, ok := client.written[0].(*clientboundConfiguration.FinishConfigurationClientboundPacket); !ok {
		t.Errorf("expected *configuration.FinishConfigurationClientboundPacket, got %T", client.written[0])
	}

	// Writing before the phase moves would resolve the packet id in the login
	// phase, where finish configuration is not registered.
	if client.writePhases[0] != types.PhaseConfiguration {
		t.Errorf("expected the packet to be written in phase %d, got %d", types.PhaseConfiguration, client.writePhases[0])
	}
}

func TestHandleAcknowledgeFinishConfigurationServerboundPacketEntersPlay(t *testing.T) {
	client := &fakeClient{phase: types.PhaseConfiguration}

	if err := HandleAcknowledgeFinishConfigurationServerboundPacket(client, &configuration.AcknowledgeFinishConfigurationServerboundPacket{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Phase() != types.PhasePlay {
		t.Errorf("expected phase %d, got %d", types.PhasePlay, client.Phase())
	}

	if len(client.written) != 0 {
		t.Errorf("expected no written packets, got %d", len(client.written))
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

	if err := HandleLoginAcknowledgedServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleAcknowledgeFinishConfigurationServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}
}
