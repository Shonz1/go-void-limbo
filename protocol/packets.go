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

	protocol1_21    = types.ProtocolVersions.MINECRAFT_1_21.ID
	protocol1_21_2  = types.ProtocolVersions.MINECRAFT_1_21_2.ID
	protocol1_21_4  = types.ProtocolVersions.MINECRAFT_1_21_4.ID
	protocol1_21_5  = types.ProtocolVersions.MINECRAFT_1_21_5.ID
	protocol1_21_6  = types.ProtocolVersions.MINECRAFT_1_21_6.ID
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
// 26.1 and 26.2 number all of these the same, and the 1.21.x versions from
// 1.21.6 on -- which number every one of these identically to each other --
// differ from them only in the play phase, where the packets 26.1 added (the
// attack, game rule and spectate packets among others) shifted the ids of
// everything registered after them. 1.21.5 sits one lower again through most
// of that phase: 1.21.6 is where the change game mode packet landed, at 0x04,
// and every play packet this server reads but the teleport acknowledgement
// is registered after it. 1.21.4 numbers all but one of these as 1.21.5
// does: 1.21.5 inserted the set test block packet at 0x39, and of the
// packets this server reads only the swing is registered after it. 1.21.2
// sits one lower than 1.21.4 from 0x23 on: 1.21.4 split the pick item packet
// at 0x22 in two, and the player input and the swing are registered after
// it. 1.21.2 has no player loaded packet at all -- 1.21.4 is where it
// appeared -- and 1.21 has no client tick end either, which 1.21.2 added at
// 0x0B and shifted everything after it by: the two absences in these tables
// that mean what they say. Every other play packet 1.21 reads sits two lower
// than on 1.21.2, but the teleport acknowledgement at 0x00, since 1.21.2
// also added the bundle item selected packet at 0x02. Outside the play phase
// 1.21 numbers everything as 1.21.2 does. The ids are written out per
// version anyway rather than shared,
// because a table that says what each version does is one where the version
// that differs shows up as a different number rather than as an absence.
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
		ids:     packetIds{protocolZero: 0x00, protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhaseStatus,
		packet:  reflect.TypeOf(status.PingRequestServerboundPacket{}),
		decoder: status.DecodePingRequestServerboundPacket,
		handler: handlers.HandlePingRequestServerboundPacket,
		ids:     packetIds{protocolZero: 0x01, protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},

	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginStartServerboundPacket{}),
		decoder: login.DecodeLoginStartServerboundPacket,
		handler: handlers.HandleLoginStartServerboundPacket,
		ids:     packetIds{protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.EncryptionResponseServerboundPacket{}),
		decoder: login.DecodeEncryptionResponseServerboundPacket,
		handler: handlers.HandleEncryptionResponseServerboundPacket,
		ids:     packetIds{protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginPluginResponseServerboundPacket{}),
		decoder: login.DecodeLoginPluginResponseServerboundPacket,
		handler: handlers.HandleLoginPluginResponseServerboundPacket,
		ids:     packetIds{protocol1_21: 0x02, protocol1_21_2: 0x02, protocol1_21_4: 0x02, protocol1_21_5: 0x02, protocol1_21_6: 0x02, protocol1_21_7: 0x02, protocol1_21_9: 0x02, protocol1_21_11: 0x02, protocol26_1: 0x02, protocol26_2: 0x02},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginAcknowledgedServerboundPacket{}),
		decoder: login.DecodeLoginAcknowledgedServerboundPacket,
		handler: handlers.HandleLoginAcknowledgedServerboundPacket,
		ids:     packetIds{protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x03, protocol1_21_6: 0x03, protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},

	{
		phase:   types.PhaseConfiguration,
		packet:  reflect.TypeOf(configuration.AcknowledgeFinishConfigurationServerboundPacket{}),
		decoder: configuration.DecodeAcknowledgeFinishConfigurationServerboundPacket,
		handler: handlers.HandleAcknowledgeFinishConfigurationServerboundPacket,
		ids:     packetIds{protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x03, protocol1_21_6: 0x03, protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},

	// The same keep alive in both phases that have one, under the id each phase
	// gives it.
	{
		phase:   types.PhaseConfiguration,
		packet:  reflect.TypeOf(serverboundCommon.KeepAliveServerboundPacket{}),
		decoder: serverboundCommon.DecodeKeepAliveServerboundPacket,
		handler: handlers.HandleKeepAliveServerboundPacket,
		ids:     packetIds{protocol1_21: 0x04, protocol1_21_2: 0x04, protocol1_21_4: 0x04, protocol1_21_5: 0x04, protocol1_21_6: 0x04, protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundCommon.KeepAliveServerboundPacket{}),
		decoder: serverboundCommon.DecodeKeepAliveServerboundPacket,
		handler: handlers.HandleKeepAliveServerboundPacket,
		ids:     packetIds{protocol1_21: 0x18, protocol1_21_2: 0x1A, protocol1_21_4: 0x1A, protocol1_21_5: 0x1A, protocol1_21_6: 0x1B, protocol1_21_7: 0x1B, protocol1_21_9: 0x1B, protocol1_21_11: 0x1B, protocol26_1: 0x1C, protocol26_2: 0x1C},
	},

	// What a joined client sends on its own, none of which needs a reaction
	// from a limbo.
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.AcceptTeleportationServerboundPacket{}),
		decoder: serverboundPlay.DecodeAcceptTeleportationServerboundPacket,
		ids:     packetIds{protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.ClientTickEndServerboundPacket{}),
		decoder: serverboundPlay.DecodeClientTickEndServerboundPacket,
		ids:     packetIds{protocol1_21_2: 0x0B, protocol1_21_4: 0x0B, protocol1_21_5: 0x0B, protocol1_21_6: 0x0C, protocol1_21_7: 0x0C, protocol1_21_9: 0x0C, protocol1_21_11: 0x0C, protocol26_1: 0x0D, protocol26_2: 0x0D},
	},
	// The move player packets feed the player sync: what each one reports is
	// recorded and relayed to the other players.
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerPositionServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerPositionServerboundPacket,
		handler: handlers.HandleMovePlayerPositionServerboundPacket,
		ids:     packetIds{protocol1_21: 0x1A, protocol1_21_2: 0x1C, protocol1_21_4: 0x1C, protocol1_21_5: 0x1C, protocol1_21_6: 0x1D, protocol1_21_7: 0x1D, protocol1_21_9: 0x1D, protocol1_21_11: 0x1D, protocol26_1: 0x1E, protocol26_2: 0x1E},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerPositionRotationServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerPositionRotationServerboundPacket,
		handler: handlers.HandleMovePlayerPositionRotationServerboundPacket,
		ids:     packetIds{protocol1_21: 0x1B, protocol1_21_2: 0x1D, protocol1_21_4: 0x1D, protocol1_21_5: 0x1D, protocol1_21_6: 0x1E, protocol1_21_7: 0x1E, protocol1_21_9: 0x1E, protocol1_21_11: 0x1E, protocol26_1: 0x1F, protocol26_2: 0x1F},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerRotationServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerRotationServerboundPacket,
		handler: handlers.HandleMovePlayerRotationServerboundPacket,
		ids:     packetIds{protocol1_21: 0x1C, protocol1_21_2: 0x1E, protocol1_21_4: 0x1E, protocol1_21_5: 0x1E, protocol1_21_6: 0x1F, protocol1_21_7: 0x1F, protocol1_21_9: 0x1F, protocol1_21_11: 0x1F, protocol26_1: 0x20, protocol26_2: 0x20},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerStatusServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerStatusServerboundPacket,
		handler: handlers.HandleMovePlayerStatusServerboundPacket,
		ids:     packetIds{protocol1_21: 0x1D, protocol1_21_2: 0x1F, protocol1_21_4: 0x1F, protocol1_21_5: 0x1F, protocol1_21_6: 0x20, protocol1_21_7: 0x20, protocol1_21_9: 0x20, protocol1_21_11: 0x20, protocol26_1: 0x21, protocol26_2: 0x21},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.PlayerInputServerboundPacket{}),
		decoder: serverboundPlay.DecodePlayerInputServerboundPacket,
		handler: handlers.HandlePlayerInputServerboundPacket,
		ids:     packetIds{protocol1_21: 0x26, protocol1_21_2: 0x28, protocol1_21_4: 0x29, protocol1_21_5: 0x29, protocol1_21_6: 0x2A, protocol1_21_7: 0x2A, protocol1_21_9: 0x2A, protocol1_21_11: 0x2A, protocol26_1: 0x2B, protocol26_2: 0x2B},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.SwingServerboundPacket{}),
		decoder: serverboundPlay.DecodeSwingServerboundPacket,
		handler: handlers.HandleSwingServerboundPacket,
		ids:     packetIds{protocol1_21: 0x36, protocol1_21_2: 0x38, protocol1_21_4: 0x3A, protocol1_21_5: 0x3B, protocol1_21_6: 0x3C, protocol1_21_7: 0x3C, protocol1_21_9: 0x3C, protocol1_21_11: 0x3C, protocol26_1: 0x3F, protocol26_2: 0x3F},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.PlayerLoadedServerboundPacket{}),
		decoder: serverboundPlay.DecodePlayerLoadedServerboundPacket,
		ids:     packetIds{protocol1_21_4: 0x2A, protocol1_21_5: 0x2A, protocol1_21_6: 0x2B, protocol1_21_7: 0x2B, protocol1_21_9: 0x2B, protocol1_21_11: 0x2B, protocol26_1: 0x2C, protocol26_2: 0x2C},
	},
}

// clientboundPackets is every packet the server writes.
//
// 1.21.4 numbers every play packet after the add entity one higher than
// 1.21.5 does: 1.21.5 is where the add experience orb packet, 0x02, was
// retired, and everything registered after it moved down. Outside the play
// phase the two agree. 1.21.2 numbers everything as 1.21.4 does: 1.21.4
// added no clientbound packet in front of any this server sends. 1.21 numbers
// the play packets this server sends lower than 1.21.2 does: by one from the
// game event on, since 1.21.2 is where the entity position sync landed, at
// 0x20, and by more further along, where 1.21.2's minecart, player rotation
// and recipe book packets fall in between. The add entity and animate
// packets in front of all of them agree. Outside the play phase 1.21 agrees
// with 1.21.2 too, though the login success goes by another name there --
// the game profile packet -- under the same id.
var clientboundPackets = []clientboundPacket{
	// Answered on protocol zero as well, for the same reason the requests are
	// read there: a client on a version this server does not speak still gets an
	// answer, and works out from the version in it that it cannot join.
	{
		phase:  types.PhaseStatus,
		packet: reflect.TypeOf(clientboundStatus.StatusResponseClientboundPacket{}),
		ids:    packetIds{protocolZero: 0x00, protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:  types.PhaseStatus,
		packet: reflect.TypeOf(clientboundStatus.PongResponseClientboundPacket{}),
		ids:    packetIds{protocolZero: 0x01, protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},

	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.DisconnectClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.EncryptionRequestClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x02, protocol1_21_2: 0x02, protocol1_21_4: 0x02, protocol1_21_5: 0x02, protocol1_21_6: 0x02, protocol1_21_7: 0x02, protocol1_21_9: 0x02, protocol1_21_11: 0x02, protocol26_1: 0x02, protocol26_2: 0x02},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.SetCompressionClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x03, protocol1_21_6: 0x03, protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.LoginPluginRequestClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x04, protocol1_21_2: 0x04, protocol1_21_4: 0x04, protocol1_21_5: 0x04, protocol1_21_6: 0x04, protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},

	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.FinishConfigurationClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x03, protocol1_21_6: 0x03, protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x04, protocol1_21_2: 0x04, protocol1_21_4: 0x04, protocol1_21_5: 0x04, protocol1_21_6: 0x04, protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.RegistryDataClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x07, protocol1_21_2: 0x07, protocol1_21_4: 0x07, protocol1_21_5: 0x07, protocol1_21_6: 0x07, protocol1_21_7: 0x07, protocol1_21_9: 0x07, protocol1_21_11: 0x07, protocol26_1: 0x07, protocol26_2: 0x07},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.UpdateTagsClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x0D, protocol1_21_2: 0x0D, protocol1_21_4: 0x0D, protocol1_21_5: 0x0D, protocol1_21_6: 0x0D, protocol1_21_7: 0x0D, protocol1_21_9: 0x0D, protocol1_21_11: 0x0D, protocol26_1: 0x0D, protocol26_2: 0x0D},
	},

	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.AnimateClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x02, protocol1_21_6: 0x02, protocol1_21_7: 0x02, protocol1_21_9: 0x02, protocol1_21_11: 0x02, protocol26_1: 0x02, protocol26_2: 0x02},
	},
	// 1.21 has no entity position sync: 1.21.2 introduced it, and before it
	// the teleport entity packet was how an entity was put where the server
	// says. The id below for 1.21 is the teleport's, and the 1.21.2 step
	// rewrites the body into the teleport's shape to go under it.
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.EntityPositionSyncClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x70, protocol1_21_2: 0x20, protocol1_21_4: 0x20, protocol1_21_5: 0x1F, protocol1_21_6: 0x1F, protocol1_21_7: 0x1F, protocol1_21_9: 0x23, protocol1_21_11: 0x23, protocol26_1: 0x23, protocol26_2: 0x23},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.GameEventClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x22, protocol1_21_2: 0x23, protocol1_21_4: 0x23, protocol1_21_5: 0x22, protocol1_21_6: 0x22, protocol1_21_7: 0x22, protocol1_21_9: 0x26, protocol1_21_11: 0x26, protocol26_1: 0x26, protocol26_2: 0x26},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x26, protocol1_21_2: 0x27, protocol1_21_4: 0x27, protocol1_21_5: 0x26, protocol1_21_6: 0x26, protocol1_21_7: 0x26, protocol1_21_9: 0x2B, protocol1_21_11: 0x2B, protocol26_1: 0x2C, protocol26_2: 0x2C},
	},
	// The chunk packet's shape is identical from 1.21.5 on, so no transformer
	// carries it between those versions, even though a body for one is wrong
	// for another: sections name block states by each version's own
	// numbering, and package world resolves that before the packet exists by
	// building a packet per version. 1.21.4 is the one version that reads the
	// packet's own fields differently, and the one step that carries it;
	// 1.21.2 reads them as 1.21.4 does.
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.LevelChunkWithLightClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x27, protocol1_21_2: 0x28, protocol1_21_4: 0x28, protocol1_21_5: 0x27, protocol1_21_6: 0x27, protocol1_21_7: 0x27, protocol1_21_9: 0x2C, protocol1_21_11: 0x2C, protocol26_1: 0x2D, protocol26_2: 0x2D},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x2B, protocol1_21_2: 0x2C, protocol1_21_4: 0x2C, protocol1_21_5: 0x2B, protocol1_21_6: 0x2B, protocol1_21_7: 0x2B, protocol1_21_9: 0x30, protocol1_21_11: 0x30, protocol26_1: 0x31, protocol26_2: 0x31},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerInfoRemoveClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x3D, protocol1_21_2: 0x3F, protocol1_21_4: 0x3F, protocol1_21_5: 0x3E, protocol1_21_6: 0x3E, protocol1_21_7: 0x3E, protocol1_21_9: 0x43, protocol1_21_11: 0x43, protocol26_1: 0x45, protocol26_2: 0x45},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x3E, protocol1_21_2: 0x40, protocol1_21_4: 0x40, protocol1_21_5: 0x3F, protocol1_21_6: 0x3F, protocol1_21_7: 0x3F, protocol1_21_9: 0x44, protocol1_21_11: 0x44, protocol26_1: 0x46, protocol26_2: 0x46},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x40, protocol1_21_2: 0x42, protocol1_21_4: 0x42, protocol1_21_5: 0x41, protocol1_21_6: 0x41, protocol1_21_7: 0x41, protocol1_21_9: 0x46, protocol1_21_11: 0x46, protocol26_1: 0x48, protocol26_2: 0x48},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.RemoveEntitiesClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x42, protocol1_21_2: 0x47, protocol1_21_4: 0x47, protocol1_21_5: 0x46, protocol1_21_6: 0x46, protocol1_21_7: 0x46, protocol1_21_9: 0x4B, protocol1_21_11: 0x4B, protocol26_1: 0x4D, protocol26_2: 0x4D},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.RotateHeadClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x48, protocol1_21_2: 0x4D, protocol1_21_4: 0x4D, protocol1_21_5: 0x4C, protocol1_21_6: 0x4C, protocol1_21_7: 0x4C, protocol1_21_9: 0x51, protocol1_21_11: 0x51, protocol26_1: 0x53, protocol26_2: 0x53},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.SetChunkCacheCenterClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x54, protocol1_21_2: 0x58, protocol1_21_4: 0x58, protocol1_21_5: 0x57, protocol1_21_6: 0x57, protocol1_21_7: 0x57, protocol1_21_9: 0x5C, protocol1_21_11: 0x5C, protocol26_1: 0x5E, protocol26_2: 0x5E},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.SetEntityDataClientboundPacket{}),
		ids:    packetIds{protocol1_21: 0x58, protocol1_21_2: 0x5D, protocol1_21_4: 0x5D, protocol1_21_5: 0x5C, protocol1_21_6: 0x5C, protocol1_21_7: 0x5C, protocol1_21_9: 0x61, protocol1_21_11: 0x61, protocol26_1: 0x63, protocol26_2: 0x63},
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

// registerTransformers records every packet whose body differs between two
// neighbouring versions, and how to carry it across.
//
// A transformer is registered against the version its input is at: an upgrade
// from the version the client sent, and a downgrade from the version the server
// encoded at. A packet with nothing registered for a step crosses it untouched,
// which is what all but the few below do.
//
// The 26.2 step carries three packets. Two are 26.2 adding a field to the end
// of something -- the login success and the play login -- and the third is the
// add entity packet, whose entity type field 26.2's registry additions
// renumbered. Every other packet this server speaks is identical in the two,
// which is checked against the client's own classes rather than assumed.
//
// The 1.21.11 step carries only the add entity packet, for the registry
// renumbering again: 26.1 numbers entity types as 1.21.11 does, and 1.21.9
// does not. Of everything else this server speaks, the only shape 26.1
// changed is inside the chunk packet's section data -- 26.1 added a fluid
// count to each section -- and chunk packets never cross versions, because
// package world builds one per version before anything is sent. The rest of
// 774's bump is all data -- the reworked dimension type and biome schema, and
// two new synchronized registries -- which is package gamedata's job.
//
// The 1.21.9 step carries two. The add entity packet changed shape there --
// 1.21.9 turned the three velocity shorts at the end into a quantized vector
// in the middle -- on top of the registry renumbering. And the set entity
// data packet names the pose serializer by its spot in a registry 1.21.9
// removed an entry from, so the number moves back up on the way down.
//
// The 1.21.7 step registers nothing, checked the same way: every packet this
// server speaks is wire-identical in 772 and 773, entity packets included. It
// does renumber packet ids, though -- the clientbound play packets 773 added
// shifted seven ids this server sends -- and the rest of the bump is data
// again, all of it from 1.21.7's own jar.
//
// The 1.21.6 step carries the add entity packet once more, for the registry
// renumbering alone: 1.21.6 added the happy ghast, which sorts before the
// player. Every other packet this server speaks is wire-identical in 770 and
// 771 -- 1.21.6's dialogs and waypoints are new packets appended after
// everything this server sends, so not even the clientbound ids moved -- and
// the rest of the bump is data: one synchronized registry fewer, and the
// block state table from 1.21.5's own jar.
//
// The 1.21.5 step carries two. The add entity packet, for the registry
// renumbering yet again: 1.21.5 split the potion entity in two, and both
// halves sort before the player. And the chunk packet, for the first time:
// 1.21.5 turned its heightmaps from an NBT compound into a counted map,
// which is a change to the packet's own shape rather than to the sections
// inside it, so it is a transformer's job. The other thing 1.21.5 changed
// about a chunk -- the length prefix a 1.21.4 paletted container carries
// before its data -- is inside the sections, and package world settles it
// per version as it does every other difference in there. Everything else
// this server speaks is wire-identical in 769 and 770, checked the same way,
// and the rest of the bump is data: eight synchronized registries fewer, and
// the block state table from 1.21.4's own jar.
//
// The 1.21.4 step carries two. The add entity packet, for the registry
// renumbering once more, in the other direction for once: 1.21.4 retired the
// transient creaking, which sat before the player, so the number goes up on
// the way down. And the player info update packet, whose hat action 1.21.4
// added: 1.21.2 has no bit for it and no field, so both come off. Everything
// else this server speaks is wire-identical in 768 and 769, chunk packet
// included, and the rest of the bump is data: the same twelve registries
// with 1.21.2's own content, and the block state table from its jar.
//
// The 1.21.2 step carries seven, more than every other step together, because
// 1.21.2 is where movement was reworked. Downwards: the add entity packet, for
// the registry renumbering at its widest; the login success, which 1.21 ends
// with a flag 1.21.2 removed; the play login, whose spawn info gained the sea
// level; the player info update, whose list order action 1.21.2 added; the
// player position packet, which 1.21.2 rebuilt around a delta movement; and
// the entity position sync, which 1.21 has no packet for and reads as a
// teleport. Upwards: the player input, which 1.21 reports as two floats and a
// flag byte where 1.21.2 reads a byte of bits. The chunk packet, the registry
// and tag payloads and the entity metadata are wire-identical in 767 and 768,
// checked the same way, and the rest of the bump is data: one synchronized
// registry fewer, and the block state table from 1.21.1's jar.
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

	// The add entity packet names the player in each version's own entity type
	// registry, which 26.2's additions renumbered.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_26_2,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo26_1,
	)

	// 1.21.11's registry additions renumbered the player too; 26.1 numbers it
	// as 1.21.11 does.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_11,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_21_9,
	)

	// 1.21.9 reworked the add entity packet's velocity on top of renumbering
	// the player, and removed a metadata serializer that sat before the pose.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_9,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_21_7,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_9,
		reflect.TypeOf(clientboundPlay.SetEntityDataClientboundPacket{}),
		transformers.DowngradeSetEntityDataTo1_21_7,
	)

	// 1.21.6's happy ghast renumbered the player one last time on the way down.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_6,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_21_5,
	)

	// 1.21.5's potion split renumbered the player at the bottom of the chain,
	// and turned the chunk packet's heightmaps into a map.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_5,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_21_4,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_5,
		reflect.TypeOf(clientboundPlay.LevelChunkWithLightClientboundPacket{}),
		transformers.DowngradeLevelChunkWithLightTo1_21_4,
	)

	// 1.21.4 retired an entity that sat before the player, and gave the player
	// list entry a hat.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_4,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_21_2,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_4,
		reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}),
		transformers.DowngradePlayerInfoUpdateTo1_21_2,
	)

	// 1.21.2 reworked movement on both sides of the wire, gave each wood its
	// own boat in front of the player, and touched the login twice.
	packetRegistry.RegisterDowngrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_21_2,
		reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}),
		transformers.DowngradeLoginSuccessTo1_21,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_2,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_21,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_2,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_21,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_2,
		reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}),
		transformers.DowngradePlayerInfoUpdateTo1_21,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_2,
		reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}),
		transformers.DowngradePlayerPositionTo1_21,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21_2,
		reflect.TypeOf(clientboundPlay.EntityPositionSyncClientboundPacket{}),
		transformers.DowngradeEntityPositionSyncTo1_21,
	)

	// The one packet this server reads that 1.21 lays out differently. An
	// upgrade is registered against the version the client sent it at.
	packetRegistry.RegisterUpgrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21,
		reflect.TypeOf(serverboundPlay.PlayerInputServerboundPacket{}),
		transformers.UpgradePlayerInputFrom1_21,
	)
}
