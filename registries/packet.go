package registries

import (
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"reflect"
)

type PacketDecoder = func(ms *streams.MinecraftStream) (types.ServerboundPacket, error)

type ServerboundPacketRegistry = map[types.Phase]map[types.ProtocolId]map[types.PacketId]PacketDecoder
type ClientboundPacketRegistry = map[types.Phase]map[reflect.Type]map[types.ProtocolId]types.PacketId

type PacketRegistry struct {
	decoderRegistry ServerboundPacketRegistry
	encoderRegistry ClientboundPacketRegistry
}

func NewPacketRegistry() *PacketRegistry {
	return &PacketRegistry{decoderRegistry: make(ServerboundPacketRegistry), encoderRegistry: make(ClientboundPacketRegistry)}
}

func (r *PacketRegistry) RegisterServerbound(phase types.Phase, protocolVersion types.ProtocolVersion, packetId types.PacketId, decoder PacketDecoder) {
	phaseRegistry := r.decoderRegistry[phase]
	if phaseRegistry == nil {
		phaseRegistry = make(map[types.ProtocolId]map[types.PacketId]PacketDecoder)
		r.decoderRegistry[phase] = phaseRegistry
	}

	protocolRegistry := phaseRegistry[protocolVersion.ID]
	if protocolRegistry == nil {
		protocolRegistry = make(map[types.PacketId]PacketDecoder)
		phaseRegistry[protocolVersion.ID] = protocolRegistry
	}

	protocolRegistry[packetId] = decoder
}

func (r *PacketRegistry) RegisterClientbound(phase types.Phase, packet reflect.Type, protocolVersion types.ProtocolVersion, packetId types.PacketId) {
	phaseRegistry := r.encoderRegistry[phase]
	if phaseRegistry == nil {
		phaseRegistry = make(map[reflect.Type]map[types.ProtocolId]types.PacketId)
		r.encoderRegistry[phase] = phaseRegistry
	}

	packetRegistry := phaseRegistry[packet]
	if packetRegistry == nil {
		packetRegistry = make(map[types.ProtocolId]types.PacketId)
		phaseRegistry[packet] = packetRegistry
	}

	packetRegistry[protocolVersion.ID] = packetId
}

func (r *PacketRegistry) GetServerbound(phase types.Phase, protocolVersion types.ProtocolVersion, packetId types.PacketId) PacketDecoder {
	phaseRegistry := r.decoderRegistry[phase]
	if phaseRegistry == nil {
		return nil
	}

	protocolRegistry := phaseRegistry[protocolVersion.ID]
	if protocolRegistry == nil {
		return nil
	}

	return protocolRegistry[packetId]
}

func (r *PacketRegistry) GetClientboundId(phase types.Phase, packet reflect.Type, protocolVersion types.ProtocolVersion) types.PacketId {
	phaseRegistry := r.encoderRegistry[phase]
	if phaseRegistry == nil {
		return -1
	}

	packetRegistry := phaseRegistry[packet]
	if packetRegistry == nil {
		return -1
	}

	id, ok := packetRegistry[protocolVersion.ID]
	if !ok {
		return -1
	}

	return id
}
