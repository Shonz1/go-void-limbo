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
