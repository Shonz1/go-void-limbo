package types

// Client is the connection state a packet handler is allowed to observe and mutate.
type Client interface {
	ProtocolVersion() ProtocolVersion
	SetProtocolVersion(protocolVersion ProtocolVersion)
	Phase() Phase
	SetPhase(phase Phase)

	// Profile is who the client says it is, which it only says once, in the
	// login phase. The play phase has to tell the client about itself, so the
	// profile outlives the packet that carried it. It is the zero profile
	// before login start has been handled.
	Profile() GameProfile
	SetProfile(profile GameProfile)

	WritePacket(packet ClientboundPacket) error

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
