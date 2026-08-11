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

	// What the client says about itself, which is worth nothing until Mojang
	// confirms it and is replaced wholesale when Mojang does. It is kept because
	// the session server is asked about the client by the name it logged in
	// under, and no later packet carries one.
	client.SetProfile(types.GameProfile{Uuid: loginStart.Uuid, Username: loginStart.Name})

	// A connection that is not encrypted is a connection that cannot be
	// authenticated, since the session server is asked about a login by a hash
	// over the secret encrypting it. So the login is finished here and on the
	// client's word alone: no encryption request goes out, and the client keeps
	// reading in the clear because none did.
	//
	// The uuid is derived from the name rather than taken from what the client
	// sent, so a name is worth the same account on every connection, and the
	// skin nobody signed for is the one the client does without.
	if !client.EncryptionEnabled() {
		profile := types.GameProfile{Uuid: types.OfflineUuid(loginStart.Name), Username: loginStart.Name}

		return completeLogin(client, profile)
	}

	publicKey, verifyToken, err := client.BeginEncryption()
	if err != nil {
		return fmt.Errorf("failed to begin encryption: %w", err)
	}

	// Nothing else goes out with this. The client answers an encryption request
	// with the secret and then encrypts everything after it, so a packet sent
	// alongside is one that arrives in a framing the client has already stopped
	// reading for. Compression is announced later, once the cipher is on.
	encryptionRequest := clientboundLogin.EncryptionRequestClientboundPacket{
		PublicKey:   publicKey,
		VerifyToken: verifyToken,

		// A limbo that took the client's word for who it is would be a limbo
		// anyone can enter under anyone's name.
		ShouldAuthenticate: true,
	}

	return client.WritePacket(&encryptionRequest)
}
