package handlers

import (
	"fmt"
	"go-void-limbo/auth"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
)

// unauthenticatedProfile is who a client logging in on an unencrypted
// connection is taken to be, under the name it logged in with.
//
// A proxy that forwarded the login answered this already: it holds the
// connection with the player, asked Mojang there, and wrote the account it got
// back into the handshake. That is the account, textures and all, since nothing
// this end could ask would add to it and there is no player on the other side of
// this connection to ask.
//
// With no proxy in front of it, the client's own word is all there is. The uuid
// is then derived from the name rather than taken from what the client sent, so
// a name is worth the same account on every connection, and the skin nobody
// signed for is the one the client does without.
func unauthenticatedProfile(client types.Client, username string) types.GameProfile {
	if forwarded, ok := client.ForwardedLogin(); ok {
		// The name is the one in this packet, which the proxy filled in from the
		// same profile and is the only place it sends it.
		return types.GameProfile{Uuid: forwarded.Uuid, Username: username, Properties: forwarded.Properties}
	}

	return offlineProfile(username)
}

// offlineProfile is who a client is when there is nobody who could be asked: the
// name it logged in under, and a uuid derived from that name rather than taken
// from what the client sent, so a name is worth the same account on every
// connection. Nothing signed for the skin, so there is none to carry.
func offlineProfile(username string) types.GameProfile {
	return types.GameProfile{Uuid: types.OfflineUuid(username), Username: username}
}

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

	// A server holding a forwarding secret asks the connection to produce the
	// login signed under it before it settles anything. Whoever is out there
	// answers next: a proxy with the account it authenticated, or a client saying
	// it has never heard of the channel, which is how the two are told apart
	// without anything being configured about this connection. The answer decides
	// which way the login goes, and a client that has none is one this server
	// settles as it would with nothing in front of it at all.
	if client.ModernForwardingEnabled() {
		messageId, err := client.BeginModernForwarding()
		if err != nil {
			return fmt.Errorf("failed to ask for the forwarded login: %w", err)
		}

		// The version asked for is the version answered at, and the payload is
		// the whole of the request.
		request := clientboundLogin.LoginPluginRequestClientboundPacket{
			MessageId: messageId,
			Channel:   auth.ModernForwardingChannel,
			Data:      []byte{auth.ModernForwardingVersion},
		}

		return client.WritePacket(&request)
	}

	// A connection that is not encrypted is a connection that cannot be
	// authenticated here, since the session server is asked about a login by a
	// hash over the secret encrypting it. So the login is finished on somebody's
	// word: no encryption request goes out, and the client keeps reading in the
	// clear because none did.
	if !client.EncryptionEnabled() {
		return completeLogin(client, unauthenticatedProfile(client, loginStart.Name))
	}

	return askForEncryption(client)
}
