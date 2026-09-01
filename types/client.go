package types

// The interfaces below are the roles a connection plays for its packet
// handlers, split so that a handler can say which slice of the connection it
// actually touches. Client composes all of them, and is what the handler
// signature carries, since handlers are looked up dynamically and need one
// common shape.

// ConnectionState is the state every phase of a connection reads and moves:
// which version it speaks, which phase it is in, and who is on it.
type ConnectionState interface {
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
}

// PacketWriter sends a packet to the client, framed however far the connection
// has come: compressed once a threshold has been announced, encrypted once a
// cipher is on.
type PacketWriter interface {
	WritePacket(packet ClientboundPacket) error
}

// StatusReporter answers the one question the status phase asks.
type StatusReporter interface {
	// ServerStatus is what a server list ping arriving on this connection is
	// answered with: what the operator set the server to say about itself, how
	// many players are on it, and a version.
	//
	// The version is the only part of it the connection has a say in, and is why
	// this is asked of the client rather than of the server: the answer reports
	// the version the client speaks whenever this server speaks it too, so that
	// every version it supports sees a server that can be joined.
	ServerStatus() ServerStatus
}

// Encrypting is the encryption handshake and the authentication that rides on
// it, which is how a login nobody vouched for is settled on a server that
// encrypts.
type Encrypting interface {
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
}

// Forwarding is the account a proxy vouched for, however it arrived: written
// into the handshake by a BungeeCord proxy, or signed into a payload of its
// own by a modern one.
type Forwarding interface {
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
}

// Compressing announces the compression threshold, which only the login phase
// has a packet for.
type Compressing interface {
	// EnableCompression tells the client the body size at or above which
	// packets are deflated, and frames everything sent afterwards that way.
	// Announcing the threshold and starting to use it are one step because the
	// packet that announces it is the last one framed uncompressed, and a
	// packet written in between would be framed as neither end expects.
	//
	// Only the login phase has a packet to announce it with, so this can only
	// be called there, and only once.
	EnableCompression(threshold int32) error
}

// KeepAliveConfirmer records the client's side of the keep alive exchange.
type KeepAliveConfirmer interface {
	// ConfirmKeepAlive records the client's answer to the keep alive the server
	// is waiting on. It reports an error when nothing was waiting on an answer
	// or when the id is not the one that was sent, neither of which a client
	// that is keeping itself alive does.
	ConfirmKeepAlive(id int64) error
}

// GameDataSource hands out the game content the configuration phase sends.
type GameDataSource interface {
	// RegistryPackets returns the configuration-phase registry packets for this
	// client's protocol version. The slice is shared across connections and must
	// not be modified.
	RegistryPackets() []ClientboundPacket
}

// WorldSource hands out the world a joined client is shown, when the server
// has one.
type WorldSource interface {
	// WorldPackets returns the packets that put the world on this client's
	// wire, in sending order, and nothing on a server with no world. The slice
	// is shared across connections and must not be modified.
	WorldPackets() []ClientboundPacket

	// WorldSpawn is where the world puts a joining player. It reports false on
	// a server with no world, whose spawn is nowhere in particular.
	WorldSpawn() (x, y, z float64, ok bool)
}

// PlayerSync is a joined player's presence among the other players: the
// entity the rest of them see, and the movements and stances it mirrors.
// Everything here is told to the connection rather than asked of it, and
// reaches the other players from there, so none of it returns an error: a
// relay that fails is some other connection dying, which that connection's own
// loops notice, and never a fact about the packet that was being relayed.
type PlayerSync interface {
	// EntityId is the id everything in the play phase names this player's
	// entity by, on this connection and every other one. It is this
	// connection's for as long as the server runs, so a player that left never
	// has its entity confused with a player that joined after it.
	EntityId() int32

	// GameMode is the mode this player is in, which the operator chose for
	// the server. It is a fact about the player rather than about the join:
	// the join handler sends it to the client itself, and the sync repeats it
	// in the player list entry everyone else keeps about this player.
	GameMode() GameMode

	// JoinPlayerSync puts this player among the others, once it is in the play
	// phase and has been sent its world: everyone already there appears on
	// this connection, and this player appears on theirs.
	JoinPlayerSync()

	// SyncPosition through SyncGround record what this player's move packets
	// report, and pass the parts other players can see along to them. Each
	// carries the fields the client actually sent, so the ones it left out
	// keep their last reported value.
	SyncPosition(x, y, z float64, onGround bool)
	SyncPositionRotation(x, y, z float64, yaw, pitch float32, onGround bool)
	SyncRotation(yaw, pitch float32, onGround bool)
	SyncGround(onGround bool)

	// SyncSwing plays this player's arm swing on everyone else's view of it.
	SyncSwing(offHand bool)

	// SyncInput records the movement keys this player is holding and shows the
	// two stances other players can see -- sneaking and sprinting -- when they
	// change.
	SyncInput(sneaking, sprinting bool)
}

// Client is the connection state a packet handler is allowed to observe and
// mutate: every role above, on one connection.
type Client interface {
	ConnectionState
	PacketWriter
	StatusReporter
	Encrypting
	Forwarding
	Compressing
	KeepAliveConfirmer
	GameDataSource
	WorldSource
	PlayerSync
}

// PacketHandler reacts to a decoded serverbound packet.
type PacketHandler = func(client Client, packet ServerboundPacket) error
