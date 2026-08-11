package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
)

// keySize is the width of the server's key. It is what a vanilla server
// generates, and it is enough for what the key is asked to do: hide a sixteen
// byte secret for the one round trip a login takes, from a key that only lives
// as long as the process.
const keySize = 1024

// verifyTokenSize is how many random bytes the client is asked to send back.
// Four is what vanilla uses.
const verifyTokenSize = 4

// KeyPair is the key a server logs clients in with. The public half travels in
// the encryption request, and the private half is what turns the client's answer
// back into the secret the connection is encrypted with.
type KeyPair struct {
	private   *rsa.PrivateKey
	publicKey []byte
}

// NewKeyPair generates a key pair. One serves every connection the process ever
// takes, the same as a vanilla server's: the key is not an identity anyone
// checks, since what ties a login to this server is the hash derived from it
// rather than the key itself.
func NewKeyPair() (*KeyPair, error) {
	private, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return nil, fmt.Errorf("failed to generate the server key: %w", err)
	}

	// The client reads the key as an X.509 subject public key info, which is
	// what Java hands it and the only encoding it accepts.
	publicKey, err := x509.MarshalPKIXPublicKey(&private.PublicKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encode the server key: %w", err)
	}

	return &KeyPair{private: private, publicKey: publicKey}, nil
}

// PublicKey is the encoded public half, as the encryption request carries it and
// as the login hash is derived over it. The slice is shared by every connection
// and must not be modified.
func (k *KeyPair) PublicKey() []byte {
	return k.publicKey
}

// Decrypt undoes what the client did to a field of its encryption response. The
// padding is PKCS #1 v1.5, which is what the client's cipher does and is not
// this end's to choose.
func (k *KeyPair) Decrypt(ciphertext []byte) ([]byte, error) {
	return rsa.DecryptPKCS1v15(rand.Reader, k.private, ciphertext)
}

// NewVerifyToken returns the few random bytes an encryption request asks the
// client to encrypt and send back. Getting them back is what ties an encryption
// response to the request it answers, so they have to be unguessable and they
// have to be new for every connection.
func NewVerifyToken() ([]byte, error) {
	token := make([]byte, verifyTokenSize)
	if _, err := rand.Read(token); err != nil {
		return nil, fmt.Errorf("failed to generate a verify token: %w", err)
	}

	return token, nil
}
