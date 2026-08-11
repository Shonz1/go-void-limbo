package handlers

import (
	"errors"
	"fmt"
	"go-void-limbo/auth"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/types"
	"log/slog"
)

// forwardingFailureReason is what a connection that could not produce the login
// this server asked the proxy for is told before it is let go. It is the same
// line whether the answer was missing, unsigned or signed with the wrong secret,
// because the three are one thing from here: a connection that did not come
// through the proxy.
const forwardingFailureReason = `{"text":"This server can only be reached through its proxy"}`

// HandleLoginPluginResponseServerboundPacket takes the answer to the one
// question this server asks during a login: the account the proxy in front of it
// vouches for, signed with the secret the two share.
//
// A server that asked has already decided that this is the only way in. So a
// client that has never heard of the channel is let go here rather than falling
// back to its own word for who it is, which is the word a forwarding secret
// exists to stop anyone having to take.
func HandleLoginPluginResponseServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	response, ok := packet.(*login.LoginPluginResponseServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *login.LoginPluginResponseServerboundPacket, got %T", packet)
	}

	if !client.ModernForwardingEnabled() {
		// Nothing was asked, so this answers nothing. A limbo has no other use
		// for the channel and no way to know what a client meant by it.
		return fmt.Errorf("a login plugin response arrived for message %d, which nothing asked for", response.MessageId)
	}

	if !response.Successful {
		return refuseForwarding(client, fmt.Errorf("the connection does not speak %s, so no proxy forwarded this login", auth.ModernForwardingChannel))
	}

	forwarded, err := client.CompleteModernForwarding(response.MessageId, response.Data)
	if err != nil {
		return refuseForwarding(client, err)
	}

	// A forwarded connection comes from the proxy rather than from the player,
	// so this is the only place the player's own address is ever said.
	slog.Info("connection forwarded by a proxy", "addr", forwarded.Address)

	// The payload settles the login on its own: the account, the name on it and
	// the signed textures all come out of it under one signature, and what the
	// client said about itself in login start is not part of it.
	profile := types.GameProfile{Uuid: forwarded.Uuid, Username: forwarded.Username, Properties: forwarded.Properties}

	return completeLogin(client, profile)
}

// refuseForwarding tells a connection why it is being let go, and reports what
// was wrong with it. A proxy misconfigured on one side of the secret and a
// client that arrived at the wrong port both end here, and the log is the only
// place the two are told apart.
func refuseForwarding(client types.Client, reason error) error {
	if err := client.WritePacket(&clientboundLogin.DisconnectClientboundPacket{Reason: forwardingFailureReason}); err != nil {
		return errors.Join(reason, err)
	}

	return fmt.Errorf("refused a login that carried no forwarded account: %w", reason)
}
