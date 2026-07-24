package registries

import (
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"reflect"
)

type PacketDecoder = func(ms *streams.MinecraftStream) (types.ServerboundPacket, error)

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
	decoderRegistry map[serverboundKey]PacketDecoder
	encoderRegistry map[clientboundKey]types.PacketId
}

func NewPacketRegistry() *PacketRegistry {
	return &PacketRegistry{
		decoderRegistry: make(map[serverboundKey]PacketDecoder),
		encoderRegistry: make(map[clientboundKey]types.PacketId),
	}
}

func (r *PacketRegistry) RegisterServerbound(phase types.Phase, protocolVersion types.ProtocolVersion, packetId types.PacketId, decoder PacketDecoder) {
	r.decoderRegistry[serverboundKey{Phase: phase, ProtocolID: protocolVersion.ID, PacketID: packetId}] = decoder
}

func (r *PacketRegistry) RegisterClientbound(phase types.Phase, packet reflect.Type, protocolVersion types.ProtocolVersion, packetId types.PacketId) {
	r.encoderRegistry[clientboundKey{Phase: phase, PacketType: packet, ProtocolID: protocolVersion.ID}] = packetId
}

func (r *PacketRegistry) GetServerbound(phase types.Phase, protocolVersion types.ProtocolVersion, packetId types.PacketId) PacketDecoder {
	return r.decoderRegistry[serverboundKey{Phase: phase, ProtocolID: protocolVersion.ID, PacketID: packetId}]
}

func (r *PacketRegistry) GetClientboundId(phase types.Phase, packet reflect.Type, protocolVersion types.ProtocolVersion) types.PacketId {
	id, ok := r.encoderRegistry[clientboundKey{Phase: phase, PacketType: packet, ProtocolID: protocolVersion.ID}]
	if !ok {
		return -1
	}

	return id
}
