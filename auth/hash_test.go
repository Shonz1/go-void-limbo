package auth

import "testing"

// The published examples, which are the digest of a string on its own. They are
// worth checking against because the hex is Java's rather than the usual kind:
// two of these are read back as negative numbers, and one has a leading zero
// that Java does not print.
func TestServerHashMatchesTheKnownExamples(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{name: "Notch", want: "4ed1f46bbe04bc756bcb17c0c7ce3e4632f06a48"},
		{name: "jeb_", want: "-7c9d5b0044c130109a5d7b5fb5c317c02b4e28c1"},
		{name: "simon", want: "88e16a1019277b15d58faf0541e11910eb756f6"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := ServerHash(test.name, nil, nil); got != test.want {
				t.Errorf("ServerHash(%q) = %q, want %q", test.name, got, test.want)
			}
		})
	}
}

// The hash is what ties a login to this server, so it has to move with every
// part of what only these two ends know.
func TestServerHashCoversTheSecretAndTheKey(t *testing.T) {
	secret := []byte("0123456789abcdef")
	publicKey := []byte("a public key")

	hash := ServerHash("", secret, publicKey)

	if hash == ServerHash("", []byte("fedcba9876543210"), publicKey) {
		t.Error("a different shared secret hashes the same, want a login that cannot be replayed onto another connection")
	}

	if hash == ServerHash("", secret, []byte("another public key")) {
		t.Error("a different public key hashes the same, want a login that cannot be replayed onto another server")
	}

	if hash != ServerHash("", secret, publicKey) {
		t.Error("the same login hashes differently each time, want the hash the client derived")
	}
}
