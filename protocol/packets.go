package protocol

import (
	"github.com/Shonz1/go-void-limbo/handlers"
	clientboundCommon "github.com/Shonz1/go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "github.com/Shonz1/go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "github.com/Shonz1/go-void-limbo/packets/clientbound/login"
	clientboundPlay "github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	clientboundStatus "github.com/Shonz1/go-void-limbo/packets/clientbound/status"
	serverboundCommon "github.com/Shonz1/go-void-limbo/packets/serverbound/common"
	"github.com/Shonz1/go-void-limbo/packets/serverbound/configuration"
	"github.com/Shonz1/go-void-limbo/packets/serverbound/handshake"
	"github.com/Shonz1/go-void-limbo/packets/serverbound/login"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
	"github.com/Shonz1/go-void-limbo/packets/serverbound/status"
	"github.com/Shonz1/go-void-limbo/transformers"
	"github.com/Shonz1/go-void-limbo/types"
	"reflect"
)

// The versions the tables below give ids for. A packet is spoken by a version
// when the version appears in that packet's ids, and by no version that does
// not.
var (
	// protocolZero is the version a connection speaks before its handshake
	// says otherwise, which only the handshake itself is read at.
	protocolZero = types.ProtocolVersions.ZERO.ID

	protocol1_21_7  = types.ProtocolVersions.MINECRAFT_1_21_7.ID
	protocol1_21_9  = types.ProtocolVersions.MINECRAFT_1_21_9.ID
	protocol1_21_11 = types.ProtocolVersions.MINECRAFT_1_21_11.ID
	protocol26_1    = types.ProtocolVersions.MINECRAFT_26_1.ID
	protocol26_2    = types.ProtocolVersions.MINECRAFT_26_2.ID
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
	decoder PacketDecoder

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
// 26.1 and 26.2 number all of these the same, and the 1.21.x versions -- which
// number every one of these identically to each other -- differ only in the
// play phase, where the packets 26.1 added (the attack, game rule and spectate
// packets among others) shifted the ids of everything registered after them.
// The ids are written out per version anyway rather than shared, because a
// table that says what each version does is one where the version that differs
// shows up as a different number rather than as an absence.
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

	// The status phase is answered on every version, including the ones this
	// server cannot be joined on. A client whose protocol is not one of the two
	// below is a client whose handshake left the connection on protocol zero, and
	// it is asking precisely so that its own server list can say the versions do
	// not match. Refusing to answer would leave it with nothing to say that of.
	{
		phase:   types.PhaseStatus,
		packet:  reflect.TypeOf(status.StatusRequestServerboundPacket{}),
		decoder: status.DecodeStatusRequestServerboundPacket,
		handler: handlers.HandleStatusRequestServerboundPacket,
		ids:     packetIds{protocolZero: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhaseStatus,
		packet:  reflect.TypeOf(status.PingRequestServerboundPacket{}),
		decoder: status.DecodePingRequestServerboundPacket,
		handler: handlers.HandlePingRequestServerboundPacket,
		ids:     packetIds{protocolZero: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},

	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginStartServerboundPacket{}),
		decoder: login.DecodeLoginStartServerboundPacket,
		handler: handlers.HandleLoginStartServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.EncryptionResponseServerboundPacket{}),
		decoder: login.DecodeEncryptionResponseServerboundPacket,
		handler: handlers.HandleEncryptionResponseServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginPluginResponseServerboundPacket{}),
		decoder: login.DecodeLoginPluginResponseServerboundPacket,
		handler: handlers.HandleLoginPluginResponseServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x02, protocol1_21_9: 0x02, protocol1_21_11: 0x02, protocol26_1: 0x02, protocol26_2: 0x02},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginAcknowledgedServerboundPacket{}),
		decoder: login.DecodeLoginAcknowledgedServerboundPacket,
		handler: handlers.HandleLoginAcknowledgedServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},

	{
		phase:   types.PhaseConfiguration,
		packet:  reflect.TypeOf(configuration.AcknowledgeFinishConfigurationServerboundPacket{}),
		decoder: configuration.DecodeAcknowledgeFinishConfigurationServerboundPacket,
		handler: handlers.HandleAcknowledgeFinishConfigurationServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},

	// The same keep alive in both phases that have one, under the id each phase
	// gives it.
	{
		phase:   types.PhaseConfiguration,
		packet:  reflect.TypeOf(serverboundCommon.KeepAliveServerboundPacket{}),
		decoder: serverboundCommon.DecodeKeepAliveServerboundPacket,
		handler: handlers.HandleKeepAliveServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundCommon.KeepAliveServerboundPacket{}),
		decoder: serverboundCommon.DecodeKeepAliveServerboundPacket,
		handler: handlers.HandleKeepAliveServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x1B, protocol1_21_9: 0x1B, protocol1_21_11: 0x1B, protocol26_1: 0x1C, protocol26_2: 0x1C},
	},

	// What a joined client sends on its own, none of which needs a reaction
	// from a limbo.
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.AcceptTeleportationServerboundPacket{}),
		decoder: serverboundPlay.DecodeAcceptTeleportationServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.ClientTickEndServerboundPacket{}),
		decoder: serverboundPlay.DecodeClientTickEndServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x0C, protocol1_21_9: 0x0C, protocol1_21_11: 0x0C, protocol26_1: 0x0D, protocol26_2: 0x0D},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerPositionServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerPositionServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x1D, protocol1_21_9: 0x1D, protocol1_21_11: 0x1D, protocol26_1: 0x1E, protocol26_2: 0x1E},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerPositionRotationServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerPositionRotationServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x1E, protocol1_21_9: 0x1E, protocol1_21_11: 0x1E, protocol26_1: 0x1F, protocol26_2: 0x1F},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerRotationServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerRotationServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x1F, protocol1_21_9: 0x1F, protocol1_21_11: 0x1F, protocol26_1: 0x20, protocol26_2: 0x20},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerStatusServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerStatusServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x20, protocol1_21_9: 0x20, protocol1_21_11: 0x20, protocol26_1: 0x21, protocol26_2: 0x21},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.PlayerLoadedServerboundPacket{}),
		decoder: serverboundPlay.DecodePlayerLoadedServerboundPacket,
		ids:     packetIds{protocol1_21_7: 0x2B, protocol1_21_9: 0x2B, protocol1_21_11: 0x2B, protocol26_1: 0x2C, protocol26_2: 0x2C},
	},
}

// clientboundPackets is every packet the server writes.
var clientboundPackets = []clientboundPacket{
	// Answered on protocol zero as well, for the same reason the requests are
	// read there: a client on a version this server does not speak still gets an
	// answer, and works out from the version in it that it cannot join.
	{
		phase:  types.PhaseStatus,
		packet: reflect.TypeOf(clientboundStatus.StatusResponseClientboundPacket{}),
		ids:    packetIds{protocolZero: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:  types.PhaseStatus,
		packet: reflect.TypeOf(clientboundStatus.PongResponseClientboundPacket{}),
		ids:    packetIds{protocolZero: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},

	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.DisconnectClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.EncryptionRequestClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x02, protocol1_21_9: 0x02, protocol1_21_11: 0x02, protocol26_1: 0x02, protocol26_2: 0x02},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.SetCompressionClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.LoginPluginRequestClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},

	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.FinishConfigurationClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.RegistryDataClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x07, protocol1_21_9: 0x07, protocol1_21_11: 0x07, protocol26_1: 0x07, protocol26_2: 0x07},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.UpdateTagsClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x0D, protocol1_21_9: 0x0D, protocol1_21_11: 0x0D, protocol26_1: 0x0D, protocol26_2: 0x0D},
	},

	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.GameEventClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x22, protocol1_21_9: 0x26, protocol1_21_11: 0x26, protocol26_1: 0x26, protocol26_2: 0x26},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x26, protocol1_21_9: 0x2B, protocol1_21_11: 0x2B, protocol26_1: 0x2C, protocol26_2: 0x2C},
	},
	// The chunk packet's shape is identical in the two versions, so no
	// transformer carries it between them, even though a body for one is wrong
	// for the other: sections name block states by each version's own
	// numbering, and package world resolves that before the packet exists by
	// building a packet per version.
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.LevelChunkWithLightClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x27, protocol1_21_9: 0x2C, protocol1_21_11: 0x2C, protocol26_1: 0x2D, protocol26_2: 0x2D},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x2B, protocol1_21_9: 0x30, protocol1_21_11: 0x30, protocol26_1: 0x31, protocol26_2: 0x31},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x3F, protocol1_21_9: 0x44, protocol1_21_11: 0x44, protocol26_1: 0x46, protocol26_2: 0x46},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x41, protocol1_21_9: 0x46, protocol1_21_11: 0x46, protocol26_1: 0x48, protocol26_2: 0x48},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.SetChunkCacheCenterClientboundPacket{}),
		ids:    packetIds{protocol1_21_7: 0x57, protocol1_21_9: 0x5C, protocol1_21_11: 0x5C, protocol26_1: 0x5E, protocol26_2: 0x5E},
	},
}

// NewDefaultRegistry builds the registry every connection resolves its packets
// through: every packet this server speaks, the id each version gives it, and
// the transformers that carry bodies between neighbouring versions.
func NewDefaultRegistry() *Registry {
	registry := NewRegistry()
	registerPackets(registry)

	return registry
}

func registerPackets(packetRegistry *Registry) {
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
//
// The 1.21.11 step registers nothing at all: of everything this server speaks,
// checked the same way, the only shape 26.1 changed is inside the chunk
// packet's section data -- 26.1 added a fluid count to each section -- and
// chunk packets never cross versions, because package world builds one per
// version before anything is sent.
//
// The 1.21.9 step registers nothing either, and does not even renumber: 774
// changed no shape and no id this server touches. What earned 774 its bump is
// all data -- the reworked dimension type and biome schema, and two new
// synchronized registries -- which is package gamedata's job, not a
// transformer's.
//
// The 1.21.7 step registers nothing as well, checked the same way: every
// packet this server speaks is wire-identical in 772 and 773. It does
// renumber, though -- the clientbound play packets 773 added shifted seven
// ids this server sends -- and the rest of the bump is data again, all of it
// from 1.21.7's own jar.
func registerTransformers(packetRegistry *Registry) {
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
