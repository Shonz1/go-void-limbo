package login

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// maxEncryptedLength bounds each of the two ciphertexts. An RSA ciphertext is as
// wide as the key that produced it, which is 128 bytes for the key this server
// offers; the bound is loose enough to leave room for a wider one and tight
// enough that a length arriving from the connection cannot ask for much.
const maxEncryptedLength = 512

// EncryptionResponseServerboundPacket carries the secret the client picked for
// the connection, and the verify token it was asked to send back, both under the
// server's public key.
type EncryptionResponseServerboundPacket struct {
	SharedSecret []byte
	VerifyToken  []byte
}

// String reports the two fields by length only. What is inside them is the key
// to the connection, and the ciphertext around it says nothing worth reading.
func (p *EncryptionResponseServerboundPacket) String() string {
	return fmt.Sprintf("EncryptionResponseServerboundPacket{SharedSecret:%d bytes VerifyToken:%d bytes}", len(p.SharedSecret), len(p.VerifyToken))
}

func DecodeEncryptionResponseServerboundPacket(minecraftStream *streams.MinecraftStream) (types.ServerboundPacket, error) {
	sharedSecret, err := minecraftStream.ReadByteArray(maxEncryptedLength)
	if err != nil {
		return nil, err
	}

	verifyToken, err := minecraftStream.ReadByteArray(maxEncryptedLength)
	if err != nil {
		return nil, err
	}

	return &EncryptionResponseServerboundPacket{SharedSecret: sharedSecret, VerifyToken: verifyToken}, nil
}
