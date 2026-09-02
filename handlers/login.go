package handlers

import (
	"fmt"
	clientboundLogin "github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	"github.com/Shonz1/go-void-limbo/types"
)

// compressionThreshold is the body size at or above which packets are deflated.
// It is what a vanilla server uses, and it sits above everything a limbo sends
// on a tick and below the registries and tags it sends once, which are the only
// packets here big enough for deflating to save more than it costs.
const compressionThreshold = 256

// askForEncryption puts the connection's login to Mojang, by asking the client
// for the secret the session server is then asked about.
//
// It is where a login with nobody vouching for it goes on an encrypted server,
// whether nothing was ever pointed at this one or the proxy in front of it had
// nothing to say about this connection. Both have the same question left, and
// this is the only way this end can ask it.
//
// Nothing else goes out with the request. The client answers it with the secret
// and then encrypts everything after that, so a packet sent alongside is one that
// arrives in a framing the client has already stopped reading for. Compression is
// announced later, once the cipher is on.
func askForEncryption(client types.Client) error {
	publicKey, verifyToken, err := client.BeginEncryption()
	if err != nil {
		return fmt.Errorf("failed to begin encryption: %w", err)
	}

	encryptionRequest := clientboundLogin.EncryptionRequestClientboundPacket{
		PublicKey:   publicKey,
		VerifyToken: verifyToken,

		// A limbo that took the client's word for who it is would be a limbo
		// anyone can enter under anyone's name.
		ShouldAuthenticate: true,
	}

	return client.WritePacket(&encryptionRequest)
}

// completeLogin welcomes a client whose profile is settled, whether it was
// settled by Mojang or, on an unencrypted connection, by the name the client
// logged in under. It is the last of the login phase: the client acknowledges
// the success packet and moves on to configuration -- or, on a version from
// before there was a configuration phase, is in play the moment it reads the
// success packet, with nothing to acknowledge, and is put into the world from
// here.
//
// From here the profile is the one that was decided rather than the one the
// client claimed, since the play phase has to tell the client about itself and
// nothing later carries it.
func completeLogin(client types.Client, profile types.GameProfile) error {
	client.SetProfile(profile)

	sessionId, err := types.NewRandomUuid()
	if err != nil {
		return fmt.Errorf("failed to generate session id: %w", err)
	}

	// Compression is announced before the success packet because the threshold
	// has to reach the client before anything framed for it does. The registries
	// that follow in the configuration phase are the bulk of what this
	// connection will ever send.
	if err := client.EnableCompression(compressionThreshold); err != nil {
		return fmt.Errorf("failed to enable compression: %w", err)
	}

	loginSuccess := clientboundLogin.LoginSuccessClientboundPacket{
		Profile:   profile,
		SessionId: sessionId,
	}

	if err := client.WritePacket(&loginSuccess); err != nil {
		return err
	}

	// A client with a configuration phase ahead of it acknowledges the
	// success packet first, and the join waits on that. One without is in
	// play already, and nothing it sends says so: the join follows the
	// success packet straight away, the way it does on a vanilla server of
	// that version.
	if client.ProtocolVersion().HasConfigurationPhase() {
		return nil
	}

	return enterPlay(client)
}
