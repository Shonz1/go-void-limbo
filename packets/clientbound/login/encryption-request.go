package login

import (
	"fmt"
	"go-void-limbo/streams"
)

// EncryptionRequestClientboundPacket asks the client to pick the secret the rest
// of the connection is encrypted with, and to send it back under the server's
// public key.
//
// It is the last packet either end sends in the clear: the client turns its
// cipher on the moment it has answered, so everything after the answer travels
// encrypted whether or not this end has caught up.
type EncryptionRequestClientboundPacket struct {
	// ServerId is what the login is named at the session server. It has been the
	// empty string since the protocol stopped having anything to put in it, and
	// the hash both ends derive is over it all the same.
	ServerId string

	PublicKey   []byte
	VerifyToken []byte

	// ShouldAuthenticate says whether the client should tell Mojang about this
	// join. A client that is not asked to skips it, which is a server that
	// wanted the encryption without the account behind it.
	ShouldAuthenticate bool
}

func (p *EncryptionRequestClientboundPacket) String() string {
	return fmt.Sprintf("EncryptionRequestClientboundPacket{ServerId:%s PublicKey:%d bytes VerifyToken:%d bytes ShouldAuthenticate:%t}", p.ServerId, len(p.PublicKey), len(p.VerifyToken), p.ShouldAuthenticate)
}

func (p *EncryptionRequestClientboundPacket) Encode(minecraftStream *streams.MinecraftStream) error {
	err := minecraftStream.WriteString(p.ServerId)
	if err != nil {
		return err
	}

	err = minecraftStream.WriteByteArray(p.PublicKey)
	if err != nil {
		return err
	}

	err = minecraftStream.WriteByteArray(p.VerifyToken)
	if err != nil {
		return err
	}

	return minecraftStream.WriteBoolean(p.ShouldAuthenticate)
}
