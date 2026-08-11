package auth

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"go-void-limbo/streams"
	"testing"
)

// testForwardingSecret is what the proxy and the server share in these tests.
// Nothing about it matters except that both sides of a check use the same one.
var testForwardingSecret = []byte("a shared secret")

// forwardedProperty is a profile property as it goes into a payload: the name,
// the value, and a signature that is there or is not.
type testForwardedProperty struct {
	name      string
	value     string
	signature string
	signed    bool
}

// buildForwardedPayload writes the fields of a payload the way a proxy does,
// without the signature in front of them.
func buildForwardedPayload(t *testing.T, version int32, address, uuid, username string, properties []testForwardedProperty) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	writes := []func() error{
		func() error { return stream.WriteVarInt(version) },
		func() error { return stream.WriteString(address) },
		func() error { return stream.WriteUuid(uuid) },
		func() error { return stream.WriteString(username) },
		func() error { return stream.WriteVarInt(int32(len(properties))) },
	}

	for _, property := range properties {
		writes = append(writes,
			func() error { return stream.WriteString(property.name) },
			func() error { return stream.WriteString(property.value) },
			func() error { return stream.WriteBoolean(property.signed) },
		)

		if property.signed {
			writes = append(writes, func() error { return stream.WriteString(property.signature) })
		}
	}

	writes = append(writes, stream.Flush)

	for _, write := range writes {
		if err := write(); err != nil {
			t.Fatalf("building the payload: %v", err)
		}
	}

	return buf.Bytes()
}

// signForwardedPayload puts the digest a proxy takes over a payload in front of
// it, which is what the whole thing is worth anything for.
func signForwardedPayload(secret, payload []byte) []byte {
	digest := hmac.New(sha256.New, secret)
	digest.Write(payload)

	return append(digest.Sum(nil), payload...)
}

func TestParseModernForwardingReadsWhatTheProxySigned(t *testing.T) {
	payload := buildForwardedPayload(t, ModernForwardingVersion, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", []testForwardedProperty{
		{name: "textures", value: "a base64 blob", signature: "a signature", signed: true},
		{name: "unsigned", value: "no signature here"},
	})

	forwarded, err := ParseModernForwarding(testForwardingSecret, signForwardedPayload(testForwardingSecret, payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The player reached the proxy from here, and the proxy reached this server
	// from somewhere else entirely, so this is the only place it is said.
	if forwarded.Address != "203.0.113.7" {
		t.Errorf("address = %q, want the one the player connected from", forwarded.Address)
	}

	if forwarded.Uuid != "069a79f4-44e9-4726-a5be-fca90e38aaf5" {
		t.Errorf("uuid = %q, want the account the proxy authenticated", forwarded.Uuid)
	}

	// Unlike the login a BungeeCord proxy writes into a handshake, this one says
	// the name, and says it under the same signature as the account.
	if forwarded.Username != "Notch" {
		t.Errorf("username = %q, want the name on the account", forwarded.Username)
	}

	if len(forwarded.Properties) != 2 {
		t.Fatalf("kept %d properties, want the 2 that were forwarded", len(forwarded.Properties))
	}

	textures := forwarded.Properties[0]
	if textures.Name != "textures" || textures.Value != "a base64 blob" || textures.Signature == nil || *textures.Signature != "a signature" {
		t.Errorf("textures = %s, want the property as it was forwarded", textures)
	}

	// A property that carries no signature has to stay that way rather than
	// gain an empty one.
	if forwarded.Properties[1].Signature != nil {
		t.Errorf("unsigned property = %s, want it left unsigned", forwarded.Properties[1])
	}
}

func TestParseModernForwardingReadsALoginWithNoProperties(t *testing.T) {
	// A proxy running in offline mode has no signed textures to pass on.
	payload := buildForwardedPayload(t, ModernForwardingVersion, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

	forwarded, err := ParseModernForwarding(testForwardingSecret, signForwardedPayload(testForwardingSecret, payload))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(forwarded.Properties) != 0 {
		t.Errorf("kept %d properties, want none forwarded", len(forwarded.Properties))
	}
}

// The one thing the whole exchange rests on: a payload from anyone who does not
// hold the secret is no login at all, however well formed the fields inside it
// are.
func TestParseModernForwardingRefusesAPayloadSignedWithAnotherSecret(t *testing.T) {
	payload := buildForwardedPayload(t, ModernForwardingVersion, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

	if _, err := ParseModernForwarding(testForwardingSecret, signForwardedPayload([]byte("a guess"), payload)); err == nil {
		t.Error("error = nil, want a payload signed under another secret refused")
	}
}

func TestParseModernForwardingRefusesAPayloadThatWasChangedAfterSigning(t *testing.T) {
	payload := buildForwardedPayload(t, ModernForwardingVersion, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)
	signed := signForwardedPayload(testForwardingSecret, payload)

	// The last byte of the account, which is the field anyone rewriting this
	// would be after.
	signed[len(signed)-1] ^= 0xFF

	if _, err := ParseModernForwarding(testForwardingSecret, signed); err == nil {
		t.Error("error = nil, want a payload that no longer matches its signature refused")
	}
}

func TestParseModernForwardingRefusesAnUnsignedPayload(t *testing.T) {
	payload := buildForwardedPayload(t, ModernForwardingVersion, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

	if _, err := ParseModernForwarding(testForwardingSecret, payload); err == nil {
		t.Error("error = nil, want a payload with no signature in front of it refused")
	}
}

func TestParseModernForwardingRefusesAPayloadWithNoRoomForASignature(t *testing.T) {
	if _, err := ParseModernForwarding(testForwardingSecret, make([]byte, modernForwardingSignatureSize-1)); err == nil {
		t.Error("error = nil, want a payload too short to hold a signature refused")
	}
}

// A server with no secret has nothing to check a payload against, so there is
// no answer it could accept. Reaching this at all would be a server that asked
// a question it cannot judge the answer to.
func TestParseModernForwardingRefusesEverythingWithoutASecret(t *testing.T) {
	payload := buildForwardedPayload(t, ModernForwardingVersion, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

	if _, err := ParseModernForwarding(nil, signForwardedPayload(nil, payload)); err == nil {
		t.Error("error = nil, want a check against no secret refused")
	}
}

// A proxy answers at the version it was asked for, and every version to date
// has begun with these same fields. One that answered above the ask is read as
// far as they go, and what it appended is left where it lies.
func TestParseModernForwardingReadsTheVersionsItKnows(t *testing.T) {
	tests := []struct {
		name    string
		version int32
		want    bool
	}{
		{name: "the version asked for", version: ModernForwardingVersion, want: true},
		{name: "a version above the ask", version: maxModernForwardingVersion, want: true},
		{name: "a version below the first", version: 0, want: false},
		{name: "a version nobody has written yet", version: maxModernForwardingVersion + 1, want: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload := buildForwardedPayload(t, test.version, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", nil)

			// A version this server does not know is still a version the proxy
			// signed, so what refuses it is the version check rather than the
			// digest.
			_, err := ParseModernForwarding(testForwardingSecret, signForwardedPayload(testForwardingSecret, payload))
			if (err == nil) != test.want {
				t.Errorf("err = %v, want a version %d that is read = %t", err, test.version, test.want)
			}
		})
	}
}

// A signed payload comes from the proxy, so a field that runs out part way
// through is a proxy that went wrong. It is reported rather than half read: the
// login is the whole of what this is for, and half of one names somebody else.
func TestParseModernForwardingReportsAPayloadThatEndsEarly(t *testing.T) {
	payload := buildForwardedPayload(t, ModernForwardingVersion, "203.0.113.7", "069a79f4-44e9-4726-a5be-fca90e38aaf5", "Notch", []testForwardedProperty{
		{name: "textures", value: "a base64 blob", signature: "a signature", signed: true},
	})

	for _, length := range []int{1, 4, 12, 30, len(payload) - 1} {
		truncated := payload[:length]

		if _, err := ParseModernForwarding(testForwardingSecret, signForwardedPayload(testForwardingSecret, truncated)); err == nil {
			t.Errorf("error = nil for %d bytes of payload, want a login that could not be read reported", length)
		}
	}
}
