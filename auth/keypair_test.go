package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"slices"
	"testing"
)

// The client encrypts against the encoded public half and this end decrypts with
// the private one, so what matters is that the encoding the client is handed is
// one it can encrypt to.
func TestKeyPairDecryptsWhatItsPublicKeyEncrypted(t *testing.T) {
	keyPair, err := NewKeyPair()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parsed, err := x509.ParsePKIXPublicKey(keyPair.PublicKey())
	if err != nil {
		t.Fatalf("the public key is not readable as the client reads it: %v", err)
	}

	publicKey, ok := parsed.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("the public key is a %T, want an RSA key", parsed)
	}

	secret := []byte("0123456789abcdef")

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, secret)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	decrypted, err := keyPair.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(decrypted, secret) {
		t.Errorf("decrypted % x, want % x", decrypted, secret)
	}

	// Anything else is a client that encrypted against another server's key, or
	// a field that was tampered with on the way here.
	if _, err := keyPair.Decrypt([]byte("not a ciphertext")); err == nil {
		t.Error("error = nil, want a ciphertext this key did not produce refused")
	}
}

// The token is only worth anything for as long as nobody can guess it, and it is
// what tells one connection's encryption response from another's.
func TestNewVerifyTokenIsNewEachTime(t *testing.T) {
	seen := make([][]byte, 0, 32)

	for range cap(seen) {
		token, err := NewVerifyToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(token) != verifyTokenSize {
			t.Fatalf("token is %d bytes, want %d", len(token), verifyTokenSize)
		}

		if slices.ContainsFunc(seen, func(other []byte) bool { return bytes.Equal(other, token) }) {
			t.Fatalf("token % x has been handed out before", token)
		}

		seen = append(seen, token)
	}
}
