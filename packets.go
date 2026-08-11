package main

import (
	"go-void-limbo/handlers"
	clientboundCommon "go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	clientboundPlay "go-void-limbo/packets/clientbound/play"
	serverboundCommon "go-void-limbo/packets/serverbound/common"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/packets/serverbound/login"
	serverboundPlay "go-void-limbo/packets/serverbound/play"
	"go-void-limbo/registries"
	"go-void-limbo/transformers"
	"go-void-limbo/types"
	"reflect"
)

// The versions the tables below give ids for. A packet is spoken by a version
// when the version appears in that packet's ids, and by no version that does
// not.
var (
	// protocolZero is the version a connection speaks before its handshake
	// says otherwise, which only the handshake itself is read at.
	protocolZero = types.ProtocolVersions.ZERO.ID

	protocol26_1 = types.ProtocolVersions.MINECRAFT_26_1.ID
	protocol26_2 = types.ProtocolVersions.MINECRAFT_26_2.ID
)

// packetIds is the id a packet has in each version that carries it.
type packetIds = map[types.ProtocolId]types.PacketId

// serverboundPacket is one packet the server reads, with the id every version
// gives it.
//
// The decoder and the handler belong to the latest version, because that is the
// only version anything is implemented at. What an older client sends is
// carried up to it by the transformers first, so a version that changed a
// packet's shape needs a transformer registered for it and a version that only
// renumbered the packet needs nothing beyond its id here.
type serverboundPacket struct {
	phase   types.Phase
	packet  reflect.Type
	decoder registries.PacketDecoder

	// handler is nil for the packets a limbo has nothing to do about. They are
	// registered all the same: a packet with no decoder is one the read loop can
	// only report as an unknown id, and a joined client sends some of these
	// every tick.
	handler types.PacketHandler

	ids packetIds
}

// clientboundPacket is one packet the server writes, with the id every version
// gives it. How it is encoded lives with the packet itself, at the latest
// version, and the transformers carry what that produces back down.
type clientboundPacket struct {
	phase  types.Phase
	packet reflect.Type
	ids    packetIds
}

// serverboundPackets is every packet the server reads.
//
// 26.1 and 26.2 number all of these the same. The ids are written out per
// version anyway rather than shared, because a table that says what each
// version does is one where the version that differs shows up as a different
// number rather than as an absence.
var serverboundPackets = []serverboundPacket{
	{
		phase:   types.PhaseHandshake,
		packet:  reflect.TypeOf(handshake.HandshakeServerboundPacket{}),
		decoder: handshake.DecodeHandshakeServerboundPacket,
		handler: handlers.HandleHandshakeServerboundPacket,
		// The handshake is what says which version the client speaks, so it is
		// read before there is a version to read it at.
		ids: packetIds{protocolZero: 0x00},
	},

	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginStartServerboundPacket{}),
		decoder: login.DecodeLoginStartServerboundPacket,
		handler: handlers.HandleLoginStartServerboundPacket,
		ids:     packetIds{protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.EncryptionResponseServerboundPacket{}),
		decoder: login.DecodeEncryptionResponseServerboundPacket,
		handler: handlers.HandleEncryptionResponseServerboundPacket,
		ids:     packetIds{protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginPluginResponseServerboundPacket{}),
		decoder: login.DecodeLoginPluginResponseServerboundPacket,
		handler: handlers.HandleLoginPluginResponseServerboundPacket,
		ids:     packetIds{protocol26_1: 0x02, protocol26_2: 0x02},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginAcknowledgedServerboundPacket{}),
		decoder: login.DecodeLoginAcknowledgedServerboundPacket,
		handler: handlers.HandleLoginAcknowledgedServerboundPacket,
		ids:     packetIds{protocol26_1: 0x03, protocol26_2: 0x03},
	},

	{
		phase:   types.PhaseConfiguration,
		packet:  reflect.TypeOf(configuration.AcknowledgeFinishConfigurationServerboundPacket{}),
		decoder: configuration.DecodeAcknowledgeFinishConfigurationServerboundPacket,
		handler: handlers.HandleAcknowledgeFinishConfigurationServerboundPacket,
		ids:     packetIds{protocol26_1: 0x03, protocol26_2: 0x03},
	},

	// The same keep alive in both phases that have one, under the id each phase
	// gives it.
	{
		phase:   types.PhaseConfiguration,
		packet:  reflect.TypeOf(serverboundCommon.KeepAliveServerboundPacket{}),
		decoder: serverboundCommon.DecodeKeepAliveServerboundPacket,
		handler: handlers.HandleKeepAliveServerboundPacket,
		ids:     packetIds{protocol26_1: 0x04, protocol26_2: 0x04},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundCommon.KeepAliveServerboundPacket{}),
		decoder: serverboundCommon.DecodeKeepAliveServerboundPacket,
		handler: handlers.HandleKeepAliveServerboundPacket,
		ids:     packetIds{protocol26_1: 0x1C, protocol26_2: 0x1C},
	},

	// What a joined client sends on its own, none of which needs a reaction
	// from a limbo.
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.AcceptTeleportationServerboundPacket{}),
		decoder: serverboundPlay.DecodeAcceptTeleportationServerboundPacket,
		ids:     packetIds{protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.ClientTickEndServerboundPacket{}),
		decoder: serverboundPlay.DecodeClientTickEndServerboundPacket,
		ids:     packetIds{protocol26_1: 0x0D, protocol26_2: 0x0D},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerPositionServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerPositionServerboundPacket,
		ids:     packetIds{protocol26_1: 0x1E, protocol26_2: 0x1E},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerPositionRotationServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerPositionRotationServerboundPacket,
		ids:     packetIds{protocol26_1: 0x1F, protocol26_2: 0x1F},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerRotationServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerRotationServerboundPacket,
		ids:     packetIds{protocol26_1: 0x20, protocol26_2: 0x20},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerStatusServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerStatusServerboundPacket,
		ids:     packetIds{protocol26_1: 0x21, protocol26_2: 0x21},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.PlayerLoadedServerboundPacket{}),
		decoder: serverboundPlay.DecodePlayerLoadedServerboundPacket,
		ids:     packetIds{protocol26_1: 0x2C, protocol26_2: 0x2C},
	},
}

// clientboundPackets is every packet the server writes.
var clientboundPackets = []clientboundPacket{
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.DisconnectClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.EncryptionRequestClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x02, protocol26_2: 0x02},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.SetCompressionClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x03, protocol26_2: 0x03},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.LoginPluginRequestClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x04, protocol26_2: 0x04},
	},

	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.FinishConfigurationClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x03, protocol26_2: 0x03},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x04, protocol26_2: 0x04},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.RegistryDataClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x07, protocol26_2: 0x07},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.UpdateTagsClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x0D, protocol26_2: 0x0D},
	},

	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.GameEventClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x26, protocol26_2: 0x26},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x2C, protocol26_2: 0x2C},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x31, protocol26_2: 0x31},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x46, protocol26_2: 0x46},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}),
		ids:    packetIds{protocol26_1: 0x48, protocol26_2: 0x48},
	},
}

func registerPackets(packetRegistry *registries.PacketRegistry) {
	for _, entry := range serverboundPackets {
		packetRegistry.RegisterServerbound(entry.phase, entry.packet, entry.decoder, entry.handler)

		for protocolId, packetId := range entry.ids {
			packetRegistry.RegisterServerboundId(entry.phase, types.GetProtocolVersionById(protocolId), entry.packet, packetId)
		}
	}

	for _, entry := range clientboundPackets {
		for protocolId, packetId := range entry.ids {
			packetRegistry.RegisterClientboundId(entry.phase, types.GetProtocolVersionById(protocolId), entry.packet, packetId)
		}
	}

	registerTransformers(packetRegistry)
}

// registerTransformers records every packet whose shape differs between two
// neighbouring versions, and how to carry it across.
//
// A transformer is registered against the version its input is at: an upgrade
// from the version the client sent, and a downgrade from the version the server
// encoded at. A packet with nothing registered for a step crosses it untouched,
// which is what all but the one below do.
// Two packets differ in shape between 26.1 and 26.2, and both are 26.2 adding a
// field to the end of something. Every other packet this server speaks is
// identical in the two, which is checked against the client's own classes rather
// than assumed.
func registerTransformers(packetRegistry *registries.PacketRegistry) {
	// 26.2 appended a session id to the login success packet.
	packetRegistry.RegisterDowngrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_26_2,
		reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}),
		transformers.DowngradeLoginSuccessTo26_1,
	)

	// 26.2 added the online mode flag to the play login packet.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_26_2,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo26_1,
	)
}
