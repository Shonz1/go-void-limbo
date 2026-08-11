package types

// ForwardedLogin is who a proxy says is behind a connection.
//
// A server behind a proxy never holds the player's connection: the proxy holds
// that one, checked the account on it with Mojang, and opened a second
// connection here to carry the login on. So everything this end would otherwise
// have asked the player is instead the proxy's word, and this is the word.
type ForwardedLogin struct {
	// Address is where the player connected to the proxy from, which is the one
	// thing this end cannot see for itself: every connection it accepts comes
	// from the proxy.
	Address string

	// Uuid is the account the proxy authenticated, and Properties the signed
	// textures the session server gave it, which are the only way anyone is
	// shown a skin.
	//
	// The username is not here. It arrives in login start, where the proxy fills
	// it in from the same profile.
	Uuid       string
	Properties []ProfileProperty
}
