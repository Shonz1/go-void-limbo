package types

// Client is the connection state a packet handler is allowed to observe and mutate.
type Client interface {
	ProtocolVersion() ProtocolVersion
	SetProtocolVersion(protocolVersion ProtocolVersion)
	Phase() Phase
	SetPhase(phase Phase)
	WritePacket(packet ClientboundPacket) error
}

// PacketHandler reacts to a decoded serverbound packet.
type PacketHandler = func(client Client, packet ServerboundPacket) error
