package handlers

import (
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	clientboundPlay "go-void-limbo/packets/clientbound/play"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
	"slices"
	"testing"
)

type fakeClient struct {
	protocolVersion types.ProtocolVersion
	phase           types.Phase
	profile         types.GameProfile
	written         []types.ClientboundPacket
	// writePhases records the phase the client was in as each packet was
	// written, since that phase is what the real client resolves a clientbound
	// packet id from.
	writePhases     []types.Phase
	registryPackets []types.ClientboundPacket
}

func (c *fakeClient) RegistryPackets() []types.ClientboundPacket { return c.registryPackets }

func (c *fakeClient) ProtocolVersion() types.ProtocolVersion { return c.protocolVersion }

func (c *fakeClient) SetProtocolVersion(protocolVersion types.ProtocolVersion) {
	c.protocolVersion = protocolVersion
}

func (c *fakeClient) Phase() types.Phase { return c.phase }

func (c *fakeClient) SetPhase(phase types.Phase) { c.phase = phase }

func (c *fakeClient) Profile() types.GameProfile { return c.profile }

func (c *fakeClient) SetProfile(profile types.GameProfile) { c.profile = profile }

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

	// Nothing later in the connection asks the client who it is, so the play
	// phase can only tell it about itself from what was kept here.
	if client.Profile().Username != packet.Name || client.Profile().Uuid != packet.Uuid {
		t.Errorf("kept profile %s, want the one the client logged in with", client.Profile())
	}
}

func TestHandleLoginAcknowledgedServerboundPacketFinishesConfiguration(t *testing.T) {
	registries := []types.ClientboundPacket{
		clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:dimension_type", []byte{0x01}),
		clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:worldgen/biome", []byte{0x02}),
	}
	client := &fakeClient{phase: types.PhaseLogin, registryPackets: registries}

	if err := HandleLoginAcknowledgedServerboundPacket(client, &login.LoginAcknowledgedServerboundPacket{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Phase() != types.PhaseConfiguration {
		t.Errorf("expected phase %d, got %d", types.PhaseConfiguration, client.Phase())
	}

	if len(client.written) != len(registries)+1 {
		t.Fatalf("expected %d written packets, got %d", len(registries)+1, len(client.written))
	}

	// The client cannot resolve anything the play phase refers to by id until
	// the registries arrive, so they have to precede finish configuration.
	for i, want := range registries {
		if client.written[i] != want {
			t.Errorf("packet %d = %v, want %v", i, client.written[i], want)
		}
	}

	last := client.written[len(client.written)-1]
	if _, ok := last.(*clientboundConfiguration.FinishConfigurationClientboundPacket); !ok {
		t.Errorf("expected *configuration.FinishConfigurationClientboundPacket, got %T", last)
	}

	// Writing before the phase moves would resolve the packet ids in the login
	// phase, where neither packet is registered.
	for i, phase := range client.writePhases {
		if phase != types.PhaseConfiguration {
			t.Errorf("expected packet %d to be written in phase %d, got %d", i, types.PhaseConfiguration, phase)
		}
	}
}

func TestHandleAcknowledgeFinishConfigurationServerboundPacketEntersPlay(t *testing.T) {
	profile := types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001", Username: "Notch"}
	client := &fakeClient{phase: types.PhaseConfiguration, profile: profile}

	if err := HandleAcknowledgeFinishConfigurationServerboundPacket(client, &configuration.AcknowledgeFinishConfigurationServerboundPacket{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.Phase() != types.PhasePlay {
		t.Errorf("expected phase %d, got %d", types.PhasePlay, client.Phase())
	}

	if len(client.written) != 4 {
		t.Fatalf("expected 4 written packets, got %d", len(client.written))
	}

	// Writing before the phase moves would resolve the packet ids in the
	// configuration phase, where none of these are registered.
	for i, phase := range client.writePhases {
		if phase != types.PhasePlay {
			t.Errorf("expected packet %d to be written in phase %d, got %d", i, types.PhasePlay, phase)
		}
	}

	login, ok := client.written[0].(*clientboundPlay.LoginClientboundPacket)
	if !ok {
		t.Fatalf("expected *play.LoginClientboundPacket first, got %T", client.written[0])
	}

	// A dimension the login packet does not list is one the client refuses to
	// be spawned in.
	if !slices.Contains(login.Dimensions, login.SpawnInfo.Dimension) {
		t.Errorf("spawn dimension %q is not among the listed dimensions %v", login.SpawnInfo.Dimension, login.Dimensions)
	}

	// Any other mode leaves the client waiting for the chunk it stands in,
	// which a limbo that sends no chunks never provides.
	if login.SpawnInfo.GameMode != clientboundPlay.GameModeSpectator {
		t.Errorf("game mode = %s, want spectator", login.SpawnInfo.GameMode)
	}

	if login.SpawnInfo.PreviousGameMode != clientboundPlay.GameModeNone {
		t.Errorf("previous game mode = %s, want none", login.SpawnInfo.PreviousGameMode)
	}

	position, ok := client.written[1].(*clientboundPlay.PlayerPositionClientboundPacket)
	if !ok {
		t.Fatalf("expected *play.PlayerPositionClientboundPacket second, got %T", client.written[1])
	}

	// The client answers with this id, so a teleport it can only report as zero
	// cannot be told from one that was never sent.
	if position.TeleportId == 0 {
		t.Error("expected a non-zero teleport id")
	}

	playerInfo, ok := client.written[2].(*clientboundPlay.PlayerInfoUpdateClientboundPacket)
	if !ok {
		t.Fatalf("expected *play.PlayerInfoUpdateClientboundPacket third, got %T", client.written[2])
	}

	if len(playerInfo.Entries) != 1 {
		t.Fatalf("expected 1 player list entry, got %d", len(playerInfo.Entries))
	}

	entry := playerInfo.Entries[0]

	// The entry is the client's own, so it has to carry the profile the client
	// logged in with rather than a fresh one.
	if entry.Profile.String() != client.Profile().String() {
		t.Errorf("player list entry profile = %s, want %s", entry.Profile, client.Profile())
	}

	// An entry the client is never told about is one it ignores every later
	// update for.
	if playerInfo.Actions&clientboundPlay.PlayerInfoAddPlayer == 0 {
		t.Errorf("actions = %s, want the entry to be added", playerInfo.Actions)
	}

	if playerInfo.Actions&clientboundPlay.PlayerInfoUpdateListed == 0 || !entry.Listed {
		t.Errorf("actions = %s listed = %t, want the player listed", playerInfo.Actions, entry.Listed)
	}

	// The client reads its own mode from both packets, so disagreeing would
	// leave it in one mode holding a list entry that says another.
	if playerInfo.Actions&clientboundPlay.PlayerInfoUpdateGameMode == 0 || entry.GameMode != login.SpawnInfo.GameMode {
		t.Errorf("player list game mode = %s, want the login packet's %s", entry.GameMode, login.SpawnInfo.GameMode)
	}

	// The client sits on its loading screen until this arrives, so it has to be
	// the packet that ends the join rather than one in the middle of it.
	chunksNext, ok := client.written[3].(*clientboundPlay.GameEventClientboundPacket)
	if !ok {
		t.Fatalf("expected *play.GameEventClientboundPacket last, got %T", client.written[3])
	}

	if chunksNext.Event != clientboundPlay.GameEventStartWaitingForChunks {
		t.Errorf("game event = %s, want start_waiting_for_chunks", chunksNext.Event)
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
