package handlers

import (
	"errors"
	"fmt"
	"github.com/Shonz1/go-void-limbo/auth"
	clientboundLogin "github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	"github.com/Shonz1/go-void-limbo/packets/serverbound/login"
	"github.com/Shonz1/go-void-limbo/types"
	"log/slog"
)

// forwardingFailureReason is what a connection that answered with a login this
// server could not verify is told before it is let go. It is the same line
// whether the payload was unsigned, signed with the wrong secret or answering a
// request that was never sent, because the three are one thing from here: an
// answer claiming the proxy's authority without holding what the proxy holds.
const forwardingFailureReason = `{"text":"This login was not signed by this server's proxy"}`

// HandleLoginPluginResponseServerboundPacket takes the answer to the one
// question this server asks during a login: the account the proxy in front of it
// vouches for, signed with the secret the two share.
//
// The question is not a demand. A connection that has never heard of the channel
// is not a connection doing anything wrong -- it is a player who came to the port
// directly, on a server a proxy also happens to point at -- so its login goes on
// to be settled the way this server settles a login nobody forwarded. A payload
// that arrives and does not hold up is the other thing entirely, and is let go.
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
		return loginWithoutTheProxy(client, response.MessageId)
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

// loginWithoutTheProxy carries on a login the proxy had nothing to say about,
// which is what a client that has never heard of the channel leaves behind.
//
// It is settled from here exactly as it would be on a server nothing was pointed
// at: Mojang is asked when the connection is to be encrypted, and the client's
// own word is taken when it is not. What it is never settled from is anything a
// proxy wrote in plain text -- the fields a BungeeCord proxy puts in a handshake
// are not read on a server that holds a secret, and are not consulted here
// either. On a server that asks for signatures, the only account a proxy can name
// is a signed one.
func loginWithoutTheProxy(client types.Client, messageId int32) error {
	// The request has been answered, even though the answer carried nothing, and
	// a connection with nothing outstanding is one a payload arriving later
	// cannot settle a second time.
	//
	// An answer to a request nothing is waiting on is let go rather than carried
	// on with. It is a second answer to the one question this server asked, and
	// the login it would settle has been settled already.
	if err := client.DeclineModernForwarding(messageId); err != nil {
		return refuseForwarding(client, fmt.Errorf("failed to give up on the forwarded login: %w", err))
	}

	// Read once: it is the name the whole of the rest of this is settled under,
	// and the log and the profile saying different names is the one way this
	// could be wrong about who was let in.
	username := client.Profile().Username

	// Said once per connection that came to the port itself, which on a server
	// behind a proxy is the thing an operator wants to know about.
	slog.Info("no proxy forwarded this login", "channel", auth.ModernForwardingChannel, "username", username)

	if client.EncryptionEnabled() {
		return askForEncryption(client)
	}

	// The name the client logged in under, which is all there is: no proxy
	// answered for it, and an unencrypted connection is one nobody else can be
	// asked about either.
	return completeLogin(client, offlineProfile(username))
}

// refuseForwarding tells a connection why it is being let go, and reports what
// was wrong with it. A proxy misconfigured on one side of the secret and
// something claiming to be one both end here, and the log is the only place the
// two are told apart.
func refuseForwarding(client types.Client, reason error) error {
	if err := client.WritePacket(&clientboundLogin.DisconnectClientboundPacket{Reason: forwardingFailureReason}); err != nil {
		return errors.Join(reason, err)
	}

	return fmt.Errorf("refused a login the forwarding secret does not vouch for: %w", reason)
}
