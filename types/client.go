package types

// Client is the connection state a packet handler is allowed to observe and mutate.
type Client interface {
	ProtocolVersion() ProtocolVersion
	SetProtocolVersion(protocolVersion ProtocolVersion)
	Phase() Phase
	SetPhase(phase Phase)

	// Profile is who the client is. It starts out as what the client said about
	// itself in login start, which is worth nothing on its own, and becomes what
	// Mojang confirmed once the login has been authenticated. Either way it
	// outlives the packets that carried it, because the play phase has to tell
	// the client about itself and nothing later asks.
	Profile() GameProfile
	SetProfile(profile GameProfile)

	WritePacket(packet ClientboundPacket) error

	// ServerStatus is what a server list ping arriving on this connection is
	// answered with: what the operator set the server to say about itself, how
	// many players are on it, and a version.
	//
	// The version is the only part of it the connection has a say in, and is why
	// this is asked of the client rather than of the server: the answer reports
	// the version the client speaks whenever this server speaks it too, so that
	// every version it supports sees a server that can be joined.
	ServerStatus() ServerStatus

	// EncryptionEnabled reports whether this connection is to be encrypted, and
	// with it whether the login is checked with Mojang at all. The two are one
	// setting because they are one exchange: the client only encrypts once it
	// has been sent an encryption request, and the session server is asked about
	// a login by a hash over the secret that request produced.
	//
	// A connection that is not encrypted is a connection anyone can log in on
	// under anyone's name, which is what turning it off is for: a test client
	// with no account behind it.
	EncryptionEnabled() bool

	// SetForwardedLogin records the account a proxy vouched for, whether it was
	// written into the handshake by a BungeeCord proxy or signed into a
	// forwarding payload by a modern one.
	SetForwardedLogin(forwarded ForwardedLogin)

	// ForwardedLogin returns what a proxy vouched for, and whether anything did.
	// Nothing did on a connection that no proxy forwarded, which is every
	// connection on a server nothing is pointed at, and it is what tells the two
	// apart: on a server with no forwarding secret there is no setting saying
	// which kind a connection is, only a handshake that either carried a login
	// or did not.
	ForwardedLogin() (ForwardedLogin, bool)

	// ModernForwardingEnabled reports whether this server holds a forwarding
	// secret, which is the whole of what says a modern proxy is in front of it.
	// A server that holds one asks every login for a payload signed with it, so
	// this is not a hint about a connection but a statement about the server.
	//
	// It says nothing about how a login ends. A connection that produces a signed
	// payload is settled by it; one that has never heard of the channel is settled
	// as it would be on a server no proxy was pointed at.
	ModernForwardingEnabled() bool

	// BeginModernForwarding returns the message id the forwarding request goes
	// out under, and is what makes the answer to it something this connection
	// asked for. A connection can only begin it once.
	BeginModernForwarding() (messageId int32, err error)

	// CompleteModernForwarding checks a forwarding payload against the message
	// id that was sent and the secret the proxy shares, and returns the login it
	// vouches for. A payload that fails either is no login at all, which is a
	// connection that does not come from the proxy.
	CompleteModernForwarding(messageId int32, payload []byte) (ForwardedLogin, error)

	// DeclineModernForwarding records that the answer to the forwarding request
	// carried no login, which is what a client that has never heard of the
	// channel answers. The request stops being outstanding either way, so a
	// payload arriving after it cannot settle the login a second time.
	//
	// It reports an error when nothing was waiting on an answer or when the id is
	// not the one that was sent, since a connection can only give up on the one
	// question this server asked it.
	DeclineModernForwarding(messageId int32) error

	// BeginEncryption starts the encryption handshake and returns what an
	// encryption request has to carry: the server's public key, and a verify
	// token the client encrypts under it and sends back. A connection can only
	// begin it once.
	BeginEncryption() (publicKey []byte, verifyToken []byte, err error)

	// CompleteEncryption takes the two fields of an encryption response,
	// decrypts them with the server's private key, refuses a verify token that
	// is not the one that was sent, and puts the connection under the shared
	// secret from there on.
	//
	// The client encrypts everything it sends after its response, so this has to
	// happen before anything is written back, and a failure here leaves a
	// connection neither end can read.
	CompleteEncryption(sharedSecret, verifyToken []byte) error

	// Authenticate asks the session server whether the client really is the
	// account it logged in as, and returns the profile Mojang holds for it.
	// Only an encrypted connection can be authenticated, since what the session
	// server is asked about is a hash over the secret encrypting it.
	Authenticate() (GameProfile, error)

	// EnableCompression tells the client the body size at or above which
	// packets are deflated, and frames everything sent afterwards that way.
	// Announcing the threshold and starting to use it are one step because the
	// packet that announces it is the last one framed uncompressed, and a
	// packet written in between would be framed as neither end expects.
	//
	// Only the login phase has a packet to announce it with, so this can only
	// be called there, and only once.
	EnableCompression(threshold int32) error

	// ConfirmKeepAlive records the client's answer to the keep alive the server
	// is waiting on. It reports an error when nothing was waiting on an answer
	// or when the id is not the one that was sent, neither of which a client
	// that is keeping itself alive does.
	ConfirmKeepAlive(id int64) error

	// RegistryPackets returns the configuration-phase registry packets for this
	// client's protocol version. The slice is shared across connections and must
	// not be modified.
	RegistryPackets() []ClientboundPacket
}

// PacketHandler reacts to a decoded serverbound packet.
type PacketHandler = func(client Client, packet ServerboundPacket) error
