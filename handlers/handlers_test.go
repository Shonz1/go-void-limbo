package handlers

import (
	"bytes"
	"errors"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	clientboundPlay "go-void-limbo/packets/clientbound/play"
	"go-void-limbo/packets/serverbound/common"
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

	// confirmedKeepAlives records the ids answered, and keepAliveErr is what the
	// real client would report for an answer that matches nothing.
	confirmedKeepAlives []int64
	keepAliveErr        error

	// compressionThresholds records the thresholds compression was enabled at,
	// and compressionAfter how many packets had been written by then. The
	// threshold only applies to what follows it, so when it was announced
	// matters as much as that it was.
	compressionThresholds []int32
	compressionAfter      int
	compressionErr        error

	// publicKey and verifyToken are what the real client hands back to put in an
	// encryption request, and beginErr is a connection that could not produce
	// them.
	publicKey   []byte
	verifyToken []byte
	beginErr    error

	// completedSecrets and completedTokens record the two fields of every
	// encryption response that got as far as the connection, and encryptedAfter
	// how many packets had been written by then. Everything the client sends
	// after its response is encrypted, so a reply written before the cipher is
	// on is a reply it cannot read.
	completedSecrets [][]byte
	completedTokens  [][]byte
	encryptedAfter   int
	completeErr      error

	// authenticated is the profile the session server would answer with, and
	// authenticateErr a client it has no record of. authenticateAfter is how
	// many packets had been written when it was asked.
	authenticated     types.GameProfile
	authenticateCalls int
	authenticateAfter int
	authenticateErr   error
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

func (c *fakeClient) EnableCompression(threshold int32) error {
	if c.compressionErr != nil {
		return c.compressionErr
	}

	c.compressionThresholds = append(c.compressionThresholds, threshold)
	c.compressionAfter = len(c.written)

	return nil
}

func (c *fakeClient) ConfirmKeepAlive(id int64) error {
	c.confirmedKeepAlives = append(c.confirmedKeepAlives, id)
	return c.keepAliveErr
}

func (c *fakeClient) BeginEncryption() ([]byte, []byte, error) {
	if c.beginErr != nil {
		return nil, nil, c.beginErr
	}

	return c.publicKey, c.verifyToken, nil
}

func (c *fakeClient) CompleteEncryption(sharedSecret, verifyToken []byte) error {
	if c.completeErr != nil {
		return c.completeErr
	}

	c.completedSecrets = append(c.completedSecrets, sharedSecret)
	c.completedTokens = append(c.completedTokens, verifyToken)
	c.encryptedAfter = len(c.written)

	return nil
}

func (c *fakeClient) Authenticate() (types.GameProfile, error) {
	c.authenticateCalls++
	c.authenticateAfter = len(c.written)

	if c.authenticateErr != nil {
		return types.GameProfile{}, c.authenticateErr
	}

	return c.authenticated, nil
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

func TestHandleLoginStartServerboundPacketAsksForEncryption(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin, publicKey: []byte("a public key"), verifyToken: []byte{0x01, 0x02, 0x03, 0x04}}
	packet := &login.LoginStartServerboundPacket{Name: "Notch", Uuid: "00000000-0000-0000-0000-000000000001"}

	if err := HandleLoginStartServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The client encrypts everything it sends after answering this, so a second
	// packet sent alongside it is one that arrives in a framing the client has
	// already stopped reading for.
	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	request, ok := client.written[0].(*clientboundLogin.EncryptionRequestClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.EncryptionRequestClientboundPacket, got %T", client.written[0])
	}

	if !bytes.Equal(request.PublicKey, client.publicKey) {
		t.Errorf("public key = % x, want the server's % x", request.PublicKey, client.publicKey)
	}

	// A token the client cannot send back leaves nothing to check its answer
	// against.
	if !bytes.Equal(request.VerifyToken, client.verifyToken) {
		t.Errorf("verify token = % x, want the one the connection is waiting on % x", request.VerifyToken, client.verifyToken)
	}

	// A client that is not asked to authenticate never tells Mojang it joined,
	// and the session server then has no record of a login this end is about to
	// ask it about.
	if !request.ShouldAuthenticate {
		t.Error("should authenticate = false, want the client to tell Mojang it joined")
	}

	// Compression is announced on the far side of the cipher, so nothing is
	// framed for a threshold the client has not been told yet.
	if len(client.compressionThresholds) != 0 {
		t.Errorf("enabled compression at %v, want it left until the connection is encrypted", client.compressionThresholds)
	}

	// The client stays in the login phase until it acknowledges the success packet.
	if client.Phase() != types.PhaseLogin {
		t.Errorf("expected phase %d, got %d", types.PhaseLogin, client.Phase())
	}

	// Nothing else on the connection carries the name the client logged in
	// under, and the session server is asked about it by that name.
	if client.Profile().Username != packet.Name || client.Profile().Uuid != packet.Uuid {
		t.Errorf("kept profile %s, want the one the client logged in with", client.Profile())
	}
}

func TestHandleLoginStartServerboundPacketReportsAFailureToBeginEncryption(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin, beginErr: errors.New("no verify token")}

	// A request sent without a token to check the answer against is a login that
	// cannot be finished, so it is not sent.
	if err := HandleLoginStartServerboundPacket(client, &login.LoginStartServerboundPacket{Name: "Notch"}); err == nil {
		t.Error("error = nil, want the failure to begin encryption passed back")
	}

	if len(client.written) != 0 {
		t.Errorf("wrote %d packets, want none after encryption could not be begun", len(client.written))
	}
}

func TestHandleEncryptionResponseServerboundPacketAuthenticatesAndWritesLoginSuccess(t *testing.T) {
	claimed := types.GameProfile{Uuid: "00000000-0000-0000-0000-000000000001", Username: "notch"}
	signature := "a signature"
	authenticated := types.GameProfile{
		Uuid:       "069a79f4-44e9-4726-a5be-fca90e38aaf5",
		Username:   "Notch",
		Properties: []types.ProfileProperty{{Name: "textures", Value: "a skin", Signature: &signature}},
	}

	client := &fakeClient{phase: types.PhaseLogin, profile: claimed, authenticated: authenticated}
	packet := &login.EncryptionResponseServerboundPacket{SharedSecret: []byte("an encrypted secret"), VerifyToken: []byte("an encrypted token")}

	if err := HandleEncryptionResponseServerboundPacket(client, packet); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(client.completedSecrets) != 1 || !bytes.Equal(client.completedSecrets[0], packet.SharedSecret) || !bytes.Equal(client.completedTokens[0], packet.VerifyToken) {
		t.Fatalf("completed encryption with %q and %q, want the two fields the response carried", client.completedSecrets, client.completedTokens)
	}

	// The client turned its own cipher on the moment it sent this, so anything
	// written before this end catches up is a packet it reads as noise.
	if client.encryptedAfter != 0 {
		t.Errorf("encrypted the connection after %d packets, want it encrypted before any reply", client.encryptedAfter)
	}

	if client.authenticateCalls != 1 {
		t.Fatalf("asked the session server %d times, want once", client.authenticateCalls)
	}

	// A login that is only checked after the client has been welcomed in is a
	// login that was not checked.
	if client.authenticateAfter != 0 {
		t.Errorf("authenticated after %d packets, want it settled before any reply", client.authenticateAfter)
	}

	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	// The threshold has to reach the client before anything framed for it does.
	if !slices.Equal(client.compressionThresholds, []int32{compressionThreshold}) {
		t.Errorf("enabled compression at %v, want the threshold announced once at %d", client.compressionThresholds, compressionThreshold)
	}

	if client.compressionAfter != 0 {
		t.Errorf("enabled compression after %d packets, want it announced before any reply", client.compressionAfter)
	}

	loginSuccess, ok := client.written[0].(*clientboundLogin.LoginSuccessClientboundPacket)
	if !ok {
		t.Fatalf("expected *login.LoginSuccessClientboundPacket, got %T", client.written[0])
	}

	// What the client claimed about itself is worth nothing next to what Mojang
	// answered, down to the case of the name and the textures that give it a
	// skin.
	if loginSuccess.Profile.String() != authenticated.String() {
		t.Errorf("login success profile = %s, want the authenticated %s", loginSuccess.Profile, authenticated)
	}

	if client.Profile().String() != authenticated.String() {
		t.Errorf("kept profile = %s, want the authenticated %s", client.Profile(), authenticated)
	}

	if loginSuccess.SessionId == "" {
		t.Error("expected a generated session id, got an empty string")
	}

	if loginSuccess.SessionId == loginSuccess.Profile.Uuid {
		t.Error("expected the session id to differ from the profile uuid")
	}
}

func TestHandleEncryptionResponseServerboundPacketDisconnectsAClientMojangDoesNotVouchFor(t *testing.T) {
	client := &fakeClient{
		phase:           types.PhaseLogin,
		profile:         types.GameProfile{Username: "Notch"},
		authenticateErr: errors.New("no record of this login"),
	}

	err := HandleEncryptionResponseServerboundPacket(client, &login.EncryptionResponseServerboundPacket{})
	if err == nil {
		t.Fatal("error = nil, want the failed authentication reported")
	}

	// The client is sitting on a connection it has no reason to think went
	// wrong, so it is told why it is being let go.
	if len(client.written) != 1 {
		t.Fatalf("expected 1 written packet, got %d", len(client.written))
	}

	if _, ok := client.written[0].(*clientboundLogin.DisconnectClientboundPacket); !ok {
		t.Fatalf("expected *login.DisconnectClientboundPacket, got %T", client.written[0])
	}

	// Nothing about a login that was refused should look like one that worked.
	if len(client.compressionThresholds) != 0 {
		t.Errorf("enabled compression at %v, want a refused login left alone", client.compressionThresholds)
	}
}

func TestHandleEncryptionResponseServerboundPacketStopsWhereEncryptionFails(t *testing.T) {
	client := &fakeClient{phase: types.PhaseLogin, completeErr: errors.New("the verify token is not the one that was sent")}

	if err := HandleEncryptionResponseServerboundPacket(client, &login.EncryptionResponseServerboundPacket{}); err == nil {
		t.Error("error = nil, want the failure to encrypt passed back")
	}

	// The client is reading through a cipher this end could not turn on, so
	// there is nothing worth writing to it, and no login worth asking about.
	if len(client.written) != 0 {
		t.Errorf("wrote %d packets, want none once the connection cannot be encrypted", len(client.written))
	}

	if client.authenticateCalls != 0 {
		t.Errorf("asked the session server %d times, want a login abandoned before it is asked about", client.authenticateCalls)
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

	// Every login here is one Mojang vouched for, and a client told otherwise
	// draws no head beside any name in the player list.
	if !login.OnlineMode {
		t.Error("online mode = false, want the client told its login was authenticated")
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

	// A new entry holds no hat, so an entry that does not say otherwise draws
	// the head as its base skin layer alone.
	if playerInfo.Actions&clientboundPlay.PlayerInfoUpdateHat == 0 || !entry.ShowHat {
		t.Errorf("actions = %s show hat = %t, want the hat shown", playerInfo.Actions, entry.ShowHat)
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

func TestHandleKeepAliveServerboundPacketConfirmsTheIdItCarries(t *testing.T) {
	// The same packet arrives in both phases that have one, and means the same
	// thing in both.
	for _, phase := range []types.Phase{types.PhaseConfiguration, types.PhasePlay} {
		client := &fakeClient{phase: phase}

		if err := HandleKeepAliveServerboundPacket(client, &common.KeepAliveServerboundPacket{Id: 1234}); err != nil {
			t.Fatalf("phase %d: unexpected error: %v", phase, err)
		}

		if !slices.Equal(client.confirmedKeepAlives, []int64{1234}) {
			t.Errorf("phase %d: confirmed %v, want the id the packet carried", phase, client.confirmedKeepAlives)
		}

		// The server asks and the client answers, so an answer needs no answer.
		if len(client.written) != 0 {
			t.Errorf("phase %d: wrote %d packets, want none", phase, len(client.written))
		}
	}
}

func TestHandleKeepAliveServerboundPacketReportsAnAnswerToNothing(t *testing.T) {
	client := &fakeClient{phase: types.PhasePlay, keepAliveErr: errors.New("answers nothing that was sent")}

	if err := HandleKeepAliveServerboundPacket(client, &common.KeepAliveServerboundPacket{Id: 1}); err == nil {
		t.Error("error = nil, want the client's rejection passed back")
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

	if err := HandleKeepAliveServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}

	if err := HandleEncryptionResponseServerboundPacket(client, &handshake.HandshakeServerboundPacket{}); err == nil {
		t.Error("expected an error for a mismatched packet type")
	}
}
