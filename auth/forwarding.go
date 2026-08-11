package auth

import (
	"encoding/json"
	"fmt"
	"github.com/Shonz1/go-void-limbo/types"
	"strings"
)

// forwardedSeparator is what a proxy joins the fields it forwards with. The
// field it writes them into is the handshake's server address, which can hold
// no null byte of its own, so nothing a client could put there is mistaken for
// a field boundary.
const forwardedSeparator = "\x00"

// The fields BungeeCord forwards, in the order it writes them: the address the
// proxy was reached at, the address the player reached it from, the uuid of the
// account it authenticated, and the profile properties the session server gave
// it. The last is left out for a proxy that has none, which is one running in
// offline mode.
const (
	forwardedHostField       = 0
	forwardedAddressField    = 1
	forwardedUuidField       = 2
	forwardedPropertiesField = 3

	minForwardedFields = forwardedPropertiesField
	maxForwardedFields = forwardedPropertiesField + 1
)

// IsForwardedAddress reports whether a handshake's server address carries a
// proxy's fields rather than the address a client was told to connect to.
//
// A null byte is the whole of the tell, and it is enough of one: it cannot
// appear in an address, so nothing a client connecting here on its own could
// have put in that field is mistaken for a proxy's work. What it does not tell
// anyone is who wrote it, which is why only a server that authenticates nobody
// itself reads any further.
func IsForwardedAddress(serverAddress string) bool {
	return strings.Contains(serverAddress, forwardedSeparator)
}

// forwardedProperty is a profile property as a proxy writes it, which is the
// json the session server answered it with, passed along untouched.
type forwardedProperty struct {
	Name      string  `json:"name"`
	Value     string  `json:"value"`
	Signature *string `json:"signature"`
}

// ParseForwardedLogin reads the identity a BungeeCord proxy put in a handshake's
// server address field in place of the address it was reached at.
//
// Nothing here proves anything, and it is not meant to: the fields are plain
// text that anyone who can open a connection to this server can write, and what
// makes them worth reading is a port only the proxy can reach. What this
// refuses is a handshake that is not a proxy's rather than a proxy that is not
// the proxy.
//
// So this is only ever asked on a connection whose login was going to be taken
// on somebody's word regardless. A proxy's word is the better of the two on
// offer there: it is at least an account somebody looked up, where the
// alternative is whatever name the client typed.
//
// A Forge client's handshake carries fields of its own, which a proxy forwards
// alongside these; that is a client this limbo has nothing to offer, and its
// handshake is refused here rather than half read.
func ParseForwardedLogin(serverAddress string) (types.ForwardedLogin, error) {
	fields := strings.Split(serverAddress, forwardedSeparator)

	if len(fields) < minForwardedFields || len(fields) > maxForwardedFields {
		return types.ForwardedLogin{}, fmt.Errorf("expected %d or %d forwarded fields, got %d", minForwardedFields, maxForwardedFields, len(fields))
	}

	// A proxy writes the uuid as the thirty-two characters the session server
	// answered with, without the hyphens the rest of the codebase carries.
	uuid, err := types.UuidFromHex(fields[forwardedUuidField])
	if err != nil {
		return types.ForwardedLogin{}, fmt.Errorf("failed to read the forwarded uuid: %w", err)
	}

	forwarded := types.ForwardedLogin{Address: fields[forwardedAddressField], Uuid: uuid}

	if len(fields) == minForwardedFields {
		return forwarded, nil
	}

	var properties []forwardedProperty
	if err := json.Unmarshal([]byte(fields[forwardedPropertiesField]), &properties); err != nil {
		return types.ForwardedLogin{}, fmt.Errorf("failed to read the forwarded properties: %w", err)
	}

	forwarded.Properties = make([]types.ProfileProperty, 0, len(properties))
	for _, property := range properties {
		forwarded.Properties = append(forwarded.Properties, types.ProfileProperty{Name: property.Name, Value: property.Value, Signature: property.Signature})
	}

	return forwarded, nil
}
