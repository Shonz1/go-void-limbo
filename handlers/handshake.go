package handlers

import (
	"fmt"
	"go-void-limbo/auth"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/types"
	"log/slog"
)

func HandleHandshakeServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	p, ok := packet.(*handshake.HandshakeServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *handshake.HandshakeServerboundPacket, got %T", packet)
	}

	intent := types.Phase(p.Intent)

	client.SetProtocolVersion(types.GetProtocolVersionById(types.ProtocolId(p.ProtocolVersion)))
	client.SetPhase(intent)

	// An encrypted server settles who a client is by asking Mojang about the
	// secret on this connection, and nothing written into a handshake can add to
	// that or stand in for it. Reading a proxy's fields here would be reading
	// them from whoever wrote them, which on a server that checks logins is the
	// one way left to skip the check, so it does not look.
	if client.EncryptionEnabled() {
		return nil
	}

	// A server holding a forwarding secret has a better question to ask, and
	// asks it in the login phase. Reading these fields alongside it would put
	// back exactly what the secret was configured to remove: an account anyone
	// who can open a connection can name, arriving one packet ahead of the one
	// that has to be signed.
	if client.ModernForwardingEnabled() {
		return nil
	}

	// The address a client sends is the one it was told to connect to and is of
	// no interest here, unless a proxy has written the login it is forwarding
	// into it. A proxy only does that to a login: anything else it opens, a ping
	// on the player's behalf being the one that exists, has no account behind it
	// to forward and carries the plain address.
	if intent != types.PhaseLogin || !auth.IsForwardedAddress(p.ServerAddress) {
		return nil
	}

	// Nothing is kept from a handshake that could not be read, so the login start
	// behind it is finished on the client's own word, as it would be had no proxy
	// written anything. That is the word this server was already going to take;
	// what is lost is the account the proxy meant to name, and losing it quietly
	// is how a player ends up as somebody else.
	forwarded, err := auth.ParseForwardedLogin(p.ServerAddress)
	if err != nil {
		return fmt.Errorf("failed to read the forwarded login: %w", err)
	}

	// A forwarded connection comes from the proxy rather than from the player, so
	// this is the only place the player's own address is ever said.
	slog.Info("connection forwarded by a proxy", "addr", forwarded.Address)

	client.SetForwardedLogin(forwarded)

	return nil
}
