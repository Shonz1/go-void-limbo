package handlers

import (
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

func HandleLoginStartServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	loginStart, ok := packet.(*login.LoginStartServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *login.LoginStartServerboundPacket, got %T", packet)
	}

	sessionId, err := types.NewRandomUuid()
	if err != nil {
		return fmt.Errorf("failed to generate session id: %w", err)
	}

	// The limbo does not authenticate, so the profile is taken at face value from
	// the client and carries no skin properties.
	profile := types.GameProfile{Uuid: loginStart.Uuid, Username: loginStart.Name}

	// The play phase has to tell the client who it is, and this is the only
	// packet that says so.
	client.SetProfile(profile)

	// Compression is announced here because login start is the first packet a
	// client sends that expects a reply, and the threshold has to reach it
	// before anything worth compressing does. The registries that follow in the
	// configuration phase are the bulk of what this connection will ever send.
	if err := client.EnableCompression(compressionThreshold); err != nil {
		return fmt.Errorf("failed to enable compression: %w", err)
	}

	loginSuccess := clientboundLogin.LoginSuccessClientboundPacket{
		Profile:   profile,
		SessionId: sessionId,
	}

	return client.WritePacket(&loginSuccess)
}
