package registries

import (
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"reflect"
)

type PacketDecoder = func(ms *streams.MinecraftStream) (types.ServerboundPacket, error)

// ServerboundEntry ties a packet's wire decoder to the handler that reacts to it.
type ServerboundEntry struct {
	Decoder PacketDecoder
	Handler types.PacketHandler
}

type serverboundKey struct {
	Phase      types.Phase
	ProtocolID types.ProtocolId
	PacketID   types.PacketId
}

type clientboundKey struct {
	Phase      types.Phase
	PacketType reflect.Type
	ProtocolID types.ProtocolId
}

type PacketRegistry struct {
	serverboundRegistry map[serverboundKey]ServerboundEntry
	encoderRegistry     map[clientboundKey]types.PacketId
}

func NewPacketRegistry() *PacketRegistry {
	return &PacketRegistry{
		serverboundRegistry: make(map[serverboundKey]ServerboundEntry),
		encoderRegistry:     make(map[clientboundKey]types.PacketId),
	}
}

func (r *PacketRegistry) RegisterServerbound(phase types.Phase, protocolVersion types.ProtocolVersion, packetId types.PacketId, decoder PacketDecoder, handler types.PacketHandler) {
	r.serverboundRegistry[serverboundKey{Phase: phase, ProtocolID: protocolVersion.ID, PacketID: packetId}] = ServerboundEntry{Decoder: decoder, Handler: handler}
}

func (r *PacketRegistry) RegisterClientbound(phase types.Phase, packet reflect.Type, protocolVersion types.ProtocolVersion, packetId types.PacketId) {
	r.encoderRegistry[clientboundKey{Phase: phase, PacketType: packet, ProtocolID: protocolVersion.ID}] = packetId
}

func (r *PacketRegistry) GetServerbound(phase types.Phase, protocolVersion types.ProtocolVersion, packetId types.PacketId) (ServerboundEntry, bool) {
	entry, ok := r.serverboundRegistry[serverboundKey{Phase: phase, ProtocolID: protocolVersion.ID, PacketID: packetId}]
	return entry, ok
}

func (r *PacketRegistry) GetClientboundId(phase types.Phase, packet reflect.Type, protocolVersion types.ProtocolVersion) types.PacketId {
	id, ok := r.encoderRegistry[clientboundKey{Phase: phase, PacketType: packet, ProtocolID: protocolVersion.ID}]
	if !ok {
		return -1
	}

	return id
}
