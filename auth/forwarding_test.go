package auth

import (
	"strings"
	"testing"
)

// forwardedAddress builds a handshake address the way a proxy does: the address
// it was reached at, and then the fields it appends, joined by null bytes.
func forwardedAddress(fields ...string) string {
	return strings.Join(append([]string{"limbo.example"}, fields...), forwardedSeparator)
}

func TestParseForwardedLoginReadsWhatTheProxyAppended(t *testing.T) {
	address := forwardedAddress(
		"203.0.113.7",
		"069a79f444e94726a5befca90e38aaf5",
		`[{"name":"textures","value":"a base64 blob","signature":"a signature"},{"name":"unsigned","value":"no signature here"}]`,
	)

	forwarded, err := ParseForwardedLogin(address)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The player reached the proxy from here, and the proxy reached this server
	// from somewhere else entirely, so this is the only place it is said.
	if forwarded.Address != "203.0.113.7" {
		t.Errorf("address = %q, want the one the player connected from", forwarded.Address)
	}

	// The proxy writes the uuid as the session server answered it, and the rest
	// of the codebase carries it hyphenated.
	if forwarded.Uuid != "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Errorf("uuid = %q, want it hyphenated", forwarded.Uuid)
	}

	if len(forwarded.Properties) != 2 {
		t.Fatalf("kept %d properties, want the 2 that were forwarded", len(forwarded.Properties))
	}

	// The signature is what every other client trusts the skin on, so it has to
	// survive the trip, and a property that has none has to stay that way rather
	// than gain an empty one.
	textures := forwarded.Properties[0]
	if textures.Name != "textures" || textures.Value != "a base64 blob" || textures.Signature == nil || *textures.Signature != "a signature" {
		t.Errorf("textures = %s, want the property as it was forwarded", textures)
	}

	if forwarded.Properties[1].Signature != nil {
		t.Errorf("unsigned property = %s, want it left unsigned", forwarded.Properties[1])
	}
}

func TestParseForwardedLoginReadsALoginWithNoProperties(t *testing.T) {
	// A proxy running in offline mode has no signed textures to pass on, and
	// leaves the field out rather than sending an empty one.
	forwarded, err := ParseForwardedLogin(forwardedAddress("203.0.113.7", "069a79f444e94726a5befca90e38aaf5"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if forwarded.Uuid != "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Errorf("uuid = %q, want it hyphenated", forwarded.Uuid)
	}

	if len(forwarded.Properties) != 0 {
		t.Errorf("kept %d properties, want none for a login that carried none", len(forwarded.Properties))
	}
}

func TestParseForwardedLoginRefusesAHandshakeItCannotRead(t *testing.T) {
	tests := []struct {
		name    string
		address string
	}{
		// A client that connected to this port itself sends the address it was
		// told to connect to and nothing else, which is the case worth refusing:
		// on a forwarding server it is a login with nobody behind it.
		{name: "no forwarded fields", address: "limbo.example"},
		{name: "only an address", address: forwardedAddress("203.0.113.7")},
		// A Forge client's handshake carries fields of its own, which a proxy
		// forwards alongside these. This limbo has nothing to offer one, and a
		// handshake it can only half read is not one to guess at.
		{name: "more fields than a proxy sends", address: forwardedAddress("FML", "", "203.0.113.7", "069a79f444e94726a5befca90e38aaf5", "[]")},
		{name: "a uuid that is not one", address: forwardedAddress("203.0.113.7", "not a uuid")},
		{name: "a hyphenated uuid", address: forwardedAddress("203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5")},
		{name: "properties that are not json", address: forwardedAddress("203.0.113.7", "069a79f444e94726a5befca90e38aaf5", "{")},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := ParseForwardedLogin(test.address); err == nil {
				t.Error("error = nil, want a handshake this end cannot make a login out of refused")
			}
		})
	}
}
