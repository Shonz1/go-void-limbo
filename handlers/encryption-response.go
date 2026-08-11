package handlers

import (
	"errors"
	"fmt"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
)

// compressionThreshold is the body size at or above which packets are deflated.
// It is what a vanilla server uses, and it sits above everything a limbo sends
// on a tick and below the registries and tags it sends once, which are the only
// packets here big enough for deflating to save more than it costs.
const compressionThreshold = 256

// authenticationFailureReason is what a client that could not be vouched for is
// told before it is let go. The login phase's disconnect carries a chat
// component rather than a plain line of text, so it is written as one.
const authenticationFailureReason = `{"text":"Failed to verify your login with Mojang"}`

func HandleEncryptionResponseServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	encryptionResponse, ok := packet.(*login.EncryptionResponseServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *login.EncryptionResponseServerboundPacket, got %T", packet)
	}

	// The client encrypted everything it sent after this packet, whatever this
	// end has done, so the cipher goes on before a byte is written back.
	if err := client.CompleteEncryption(encryptionResponse.SharedSecret, encryptionResponse.VerifyToken); err != nil {
		return fmt.Errorf("failed to complete encryption: %w", err)
	}

	profile, err := client.Authenticate()
	if err != nil {
		// The client is sitting on a connection it has no reason to think went
		// wrong, so it is told why it is being let go rather than dropped
		// without a word.
		if writeErr := client.WritePacket(&clientboundLogin.DisconnectClientboundPacket{Reason: authenticationFailureReason}); writeErr != nil {
			return errors.Join(err, writeErr)
		}

		return fmt.Errorf("failed to authenticate %s: %w", client.Profile().Username, err)
	}

	// From here the profile is Mojang's rather than the client's: the account's
	// own uuid and name, and the signed textures that are the only way anyone is
	// shown a skin.
	client.SetProfile(profile)

	sessionId, err := types.NewRandomUuid()
	if err != nil {
		return fmt.Errorf("failed to generate session id: %w", err)
	}

	// Compression is announced here because this is the first packet the server
	// sends on an encrypted connection, and the threshold has to reach the
	// client before anything worth compressing does. The registries that follow
	// in the configuration phase are the bulk of what this connection will ever
	// send.
	if err := client.EnableCompression(compressionThreshold); err != nil {
		return fmt.Errorf("failed to enable compression: %w", err)
	}

	loginSuccess := clientboundLogin.LoginSuccessClientboundPacket{
		Profile:   profile,
		SessionId: sessionId,
	}

	return client.WritePacket(&loginSuccess)
}
