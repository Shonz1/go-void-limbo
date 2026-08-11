package handlers

import (
	"fmt"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
)

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

	loginSuccess := clientboundLogin.LoginSuccessClientboundPacket{
		Profile:   profile,
		SessionId: sessionId,
	}

	return client.WritePacket(&loginSuccess)
}
