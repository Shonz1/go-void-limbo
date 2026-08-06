package types

// Client is the connection state a packet handler is allowed to observe and mutate.
type Client interface {
	ProtocolVersion() ProtocolVersion
	SetProtocolVersion(protocolVersion ProtocolVersion)
	Phase() Phase
	SetPhase(phase Phase)
	WritePacket(packet ClientboundPacket) error

	// RegistryPackets returns the configuration-phase registry packets for this
	// client's protocol version. The slice is shared across connections and must
	// not be modified.
	RegistryPackets() []ClientboundPacket
}

// PacketHandler reacts to a decoded serverbound packet.
type PacketHandler = func(client Client, packet ServerboundPacket) error
