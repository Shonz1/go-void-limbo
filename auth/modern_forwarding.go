package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// ModernForwardingChannel is the login plugin channel a proxy forwards a login
// on. The server asks on it during the login phase, and only a proxy doing
// modern forwarding has anything to answer with: a client reached directly
// answers that it has never heard of the channel.
const ModernForwardingChannel = "velocity:player_info"

// ModernForwardingVersion is the payload version this server asks for, and the
// only one it needs. A proxy answers at the version asked for or at the default,
// never above it, so asking for the first is asking for the one thing every
// version of the payload begins with: the address, the account, the name, and
// the properties.
//
// The versions above it add a client's signed chat key, which is of no use to a
// limbo that carries no chat, and which the newest clients do not send at all.
const ModernForwardingVersion = 1

// maxModernForwardingVersion is the highest version whose leading fields are
// laid out as the ones read here. Every version to date extends the payload
// rather than rearranging it, so a proxy that answered above what was asked for
// is still read, and what it appended is left where it lies.
const maxModernForwardingVersion = 4

// maxModernForwardingProperties bounds the profile properties a payload can
// carry. A signed payload comes from the proxy and is not an attacker's to
// shape, so this is a guard against a proxy that went wrong rather than one that
// went rogue: vanilla accounts carry one property, and a handful is already
// generous.
const maxModernForwardingProperties = 16

// modernForwardingSignatureSize is how many bytes of signature the payload opens
// with, which is the width of the digest it is taken with.
const modernForwardingSignatureSize = sha256.Size

// ParseModernForwarding reads the login a proxy vouched for in its answer to a
// modern forwarding request, having first checked that the proxy is the proxy.
//
// This is the whole of the difference from the login a BungeeCord proxy writes
// into a handshake. That one is plain text anyone who can reach the port can
// write, and is worth reading only because the port is one only the proxy can
// reach. This one opens with a digest taken over the rest of it under a secret
// only the proxy and this server hold, so a payload that verifies is a payload
// the proxy wrote, whoever carried it here.
//
// So the secret is the whole of the trust. Nothing else about a connection --
// not the address it came from, not the name it logged in under -- is allowed to
// stand in for a login the proxy vouched for. A connection that produces none is
// not being refused a login; it is a connection nobody vouched for, and is worth
// whatever the server would make of one without a proxy in front of it.
func ParseModernForwarding(secret, payload []byte) (types.ForwardedLogin, error) {
	if len(secret) == 0 {
		return types.ForwardedLogin{}, fmt.Errorf("no forwarding secret is configured")
	}

	if len(payload) < modernForwardingSignatureSize {
		return types.ForwardedLogin{}, fmt.Errorf("expected at least %d bytes of forwarded login, got %d", modernForwardingSignatureSize, len(payload))
	}

	signature := payload[:modernForwardingSignatureSize]
	forwarded := payload[modernForwardingSignatureSize:]

	digest := hmac.New(sha256.New, secret)
	digest.Write(forwarded)

	// A comparison that gives up early is one that says how far it got, and the
	// bytes it would say that about are the ones being guessed at.
	if !hmac.Equal(signature, digest.Sum(nil)) {
		return types.ForwardedLogin{}, fmt.Errorf("the forwarded login is not signed with the forwarding secret")
	}

	// Everything past here was written by the holder of the secret, so a field
	// that does not make sense is a proxy that went wrong rather than anything
	// worth defending against.
	return readModernForwarding(forwarded)
}

// readModernForwarding reads the fields of a payload whose signature has already
// been checked. Nothing calls it with anything else.
func readModernForwarding(forwarded []byte) (types.ForwardedLogin, error) {
	stream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(forwarded))

	version, err := stream.ReadVarInt()
	if err != nil {
		return types.ForwardedLogin{}, fmt.Errorf("failed to read the forwarded version: %w", err)
	}

	if version < ModernForwardingVersion || version > maxModernForwardingVersion {
		return types.ForwardedLogin{}, fmt.Errorf("unsupported forwarded login version: %d", version)
	}

	address, err := stream.ReadString()
	if err != nil {
		return types.ForwardedLogin{}, fmt.Errorf("failed to read the forwarded address: %w", err)
	}

	uuid, err := stream.ReadUuid()
	if err != nil {
		return types.ForwardedLogin{}, fmt.Errorf("failed to read the forwarded uuid: %w", err)
	}

	username, err := stream.ReadString()
	if err != nil {
		return types.ForwardedLogin{}, fmt.Errorf("failed to read the forwarded username: %w", err)
	}

	properties, err := readModernForwardingProperties(stream)
	if err != nil {
		return types.ForwardedLogin{}, err
	}

	// Whatever a later version appends comes after the properties, and this
	// server has nothing to do with any of it.
	return types.ForwardedLogin{Address: address, Uuid: uuid, Username: username, Properties: properties}, nil
}

// readModernForwardingProperties reads the profile properties off the end of a
// payload: the signed textures, and whatever else the proxy holds for the
// account.
func readModernForwardingProperties(stream *streams.MinecraftStream) ([]types.ProfileProperty, error) {
	count, err := stream.ReadVarInt()
	if err != nil {
		return nil, fmt.Errorf("failed to read the forwarded property count: %w", err)
	}

	if count < 0 || count > maxModernForwardingProperties {
		return nil, fmt.Errorf("invalid forwarded property count: %d", count)
	}

	properties := make([]types.ProfileProperty, 0, count)

	for index := int32(0); index < count; index++ {
		name, err := stream.ReadString()
		if err != nil {
			return nil, fmt.Errorf("failed to read the name of forwarded property %d: %w", index, err)
		}

		value, err := stream.ReadString()
		if err != nil {
			return nil, fmt.Errorf("failed to read the value of forwarded property %d: %w", index, err)
		}

		property := types.ProfileProperty{Name: name, Value: value}

		signed, err := stream.ReadBoolean()
		if err != nil {
			return nil, fmt.Errorf("failed to read whether forwarded property %d is signed: %w", index, err)
		}

		// Only Mojang's signature makes a texture worth anything to a client,
		// and a property that carries one carries it here.
		if signed {
			signature, err := stream.ReadString()
			if err != nil {
				return nil, fmt.Errorf("failed to read the signature of forwarded property %d: %w", index, err)
			}

			property.Signature = &signature
		}

		properties = append(properties, property)
	}

	return properties, nil
}
