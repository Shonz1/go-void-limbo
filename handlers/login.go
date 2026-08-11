package handlers

import (
	"fmt"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/types"
)

// compressionThreshold is the body size at or above which packets are deflated.
// It is what a vanilla server uses, and it sits above everything a limbo sends
// on a tick and below the registries and tags it sends once, which are the only
// packets here big enough for deflating to save more than it costs.
const compressionThreshold = 256

// completeLogin welcomes a client whose profile is settled, whether it was
// settled by Mojang or, on an unencrypted connection, by the name the client
// logged in under. It is the last of the login phase: the client acknowledges
// the success packet and moves on to configuration.
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

	return client.WritePacket(&loginSuccess)
}
