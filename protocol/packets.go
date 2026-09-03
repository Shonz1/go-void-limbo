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

	protocol1_18    = types.ProtocolVersions.MINECRAFT_1_18.ID
	protocol1_18_2  = types.ProtocolVersions.MINECRAFT_1_18_2.ID
	protocol1_19    = types.ProtocolVersions.MINECRAFT_1_19.ID
	protocol1_19_1  = types.ProtocolVersions.MINECRAFT_1_19_1.ID
	protocol1_19_3  = types.ProtocolVersions.MINECRAFT_1_19_3.ID
	protocol1_19_4  = types.ProtocolVersions.MINECRAFT_1_19_4.ID
	protocol1_20    = types.ProtocolVersions.MINECRAFT_1_20.ID
	protocol1_20_2  = types.ProtocolVersions.MINECRAFT_1_20_2.ID
	protocol1_20_3  = types.ProtocolVersions.MINECRAFT_1_20_3.ID
	protocol1_20_5  = types.ProtocolVersions.MINECRAFT_1_20_5.ID
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
// 1.21 numbers everything as 1.21.2 does. 1.20.5 numbers everything as 1.21
// does, in every phase: 1.21 added no serverbound packet at all, which the
// two jars' protocol registrations agree on. 1.20.3 sits three lower than
// 1.20.5 through the play phase from the keep alive on, and one lower in the
// configuration phase from the finish acknowledgement on: 1.20.5 added the
// cookie response to both, and two more play packets in front of the keep
// alive. The login and status phases are numbered alike. 1.20.2 sits one
// lower than 1.20.3 through the play phase from the keep alive on: 1.20.3
// added the container slot state change at 0x0F, for the crafter, in front
// of everything this server reads but the teleport acknowledgement. Every
// other phase is numbered alike. 1.20 sits two lower than 1.20.2 through the
// play phase from the keep alive on: 1.20.2 added the chunk batch received
// and the configuration acknowledgement in front of it. 1.20 has no
// configuration phase at all, and so no login acknowledgement and no packet
// of that phase: the three absences in these tables that mean what they say
// at the bottom of the chain, alongside the two at the top. 1.19.4 is
// numbered as 1.20 is in every phase, with the same three absences: 1.20
// added no packet and retired none, in either direction. 1.19.3 sits one
// lower than 1.19.4 through the play phase from the keep alive to the player
// input, with the same three absences: 1.19.4 moved the chat session update
// from the end of the phase to 0x06, in front of everything this server
// reads but the teleport acknowledgement, and the swing sits behind where it
// came from, so it is numbered alike. The login and status phases are
// numbered alike. 1.19.1 sits one higher than 1.19.3 through the play phase
// from the keep alive to the player input, with the same three absences:
// 1.19.3 retired the chat preview at 0x06, which sat in front of everything
// this server reads but the teleport acknowledgement, and put the chat
// session update at the end of the phase in its place, behind the swing, so
// the swing is numbered alike. The login and status phases are numbered
// alike. 1.19 sits one lower than 1.19.1 through the play phase from the
// keep alive to the swing, with the same three absences: 1.19.1 added the
// chat acknowledgement at 0x03, in front of everything this server reads
// but the teleport acknowledgement. The login and status phases are
// numbered alike. 1.18.2 sits two lower than 1.19 through the play phase
// from the keep alive to the swing, with the same three absences: 1.19
// added the chat command at 0x03 and the chat preview at 0x05, in front of
// everything this server reads but the teleport acknowledgement. The login
// and status phases are numbered alike. 1.18 numbers every phase as 1.18.2
// does: 1.18.2 added no packet and retired none, in either direction, which
// the two jars' protocol registrations agree on. The ids are written out
// per version anyway rather than shared, because a table that says what
// each version does is one where the version that differs shows up as a
// different number rather than as an absence.
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
		ids:     packetIds{protocolZero: 0x00, protocol1_18: 0x00, protocol1_18_2: 0x00, protocol1_19: 0x00, protocol1_19_1: 0x00, protocol1_19_3: 0x00, protocol1_19_4: 0x00, protocol1_20: 0x00, protocol1_20_2: 0x00, protocol1_20_3: 0x00, protocol1_20_5: 0x00, protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhaseStatus,
		packet:  reflect.TypeOf(status.PingRequestServerboundPacket{}),
		decoder: status.DecodePingRequestServerboundPacket,
		handler: handlers.HandlePingRequestServerboundPacket,
		ids:     packetIds{protocolZero: 0x01, protocol1_18: 0x01, protocol1_18_2: 0x01, protocol1_19: 0x01, protocol1_19_1: 0x01, protocol1_19_3: 0x01, protocol1_19_4: 0x01, protocol1_20: 0x01, protocol1_20_2: 0x01, protocol1_20_3: 0x01, protocol1_20_5: 0x01, protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},

	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginStartServerboundPacket{}),
		decoder: login.DecodeLoginStartServerboundPacket,
		handler: handlers.HandleLoginStartServerboundPacket,
		ids:     packetIds{protocol1_18: 0x00, protocol1_18_2: 0x00, protocol1_19: 0x00, protocol1_19_1: 0x00, protocol1_19_3: 0x00, protocol1_19_4: 0x00, protocol1_20: 0x00, protocol1_20_2: 0x00, protocol1_20_3: 0x00, protocol1_20_5: 0x00, protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.EncryptionResponseServerboundPacket{}),
		decoder: login.DecodeEncryptionResponseServerboundPacket,
		handler: handlers.HandleEncryptionResponseServerboundPacket,
		ids:     packetIds{protocol1_18: 0x01, protocol1_18_2: 0x01, protocol1_19: 0x01, protocol1_19_1: 0x01, protocol1_19_3: 0x01, protocol1_19_4: 0x01, protocol1_20: 0x01, protocol1_20_2: 0x01, protocol1_20_3: 0x01, protocol1_20_5: 0x01, protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginPluginResponseServerboundPacket{}),
		decoder: login.DecodeLoginPluginResponseServerboundPacket,
		handler: handlers.HandleLoginPluginResponseServerboundPacket,
		ids:     packetIds{protocol1_18: 0x02, protocol1_18_2: 0x02, protocol1_19: 0x02, protocol1_19_1: 0x02, protocol1_19_3: 0x02, protocol1_19_4: 0x02, protocol1_20: 0x02, protocol1_20_2: 0x02, protocol1_20_3: 0x02, protocol1_20_5: 0x02, protocol1_21: 0x02, protocol1_21_2: 0x02, protocol1_21_4: 0x02, protocol1_21_5: 0x02, protocol1_21_6: 0x02, protocol1_21_7: 0x02, protocol1_21_9: 0x02, protocol1_21_11: 0x02, protocol26_1: 0x02, protocol26_2: 0x02},
	},
	{
		phase:   types.PhaseLogin,
		packet:  reflect.TypeOf(login.LoginAcknowledgedServerboundPacket{}),
		decoder: login.DecodeLoginAcknowledgedServerboundPacket,
		handler: handlers.HandleLoginAcknowledgedServerboundPacket,
		ids:     packetIds{protocol1_20_2: 0x03, protocol1_20_3: 0x03, protocol1_20_5: 0x03, protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x03, protocol1_21_6: 0x03, protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},

	{
		phase:   types.PhaseConfiguration,
		packet:  reflect.TypeOf(configuration.AcknowledgeFinishConfigurationServerboundPacket{}),
		decoder: configuration.DecodeAcknowledgeFinishConfigurationServerboundPacket,
		handler: handlers.HandleAcknowledgeFinishConfigurationServerboundPacket,
		ids:     packetIds{protocol1_20_2: 0x02, protocol1_20_3: 0x02, protocol1_20_5: 0x03, protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x03, protocol1_21_6: 0x03, protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},

	// The same keep alive in both phases that have one, under the id each phase
	// gives it.
	{
		phase:   types.PhaseConfiguration,
		packet:  reflect.TypeOf(serverboundCommon.KeepAliveServerboundPacket{}),
		decoder: serverboundCommon.DecodeKeepAliveServerboundPacket,
		handler: handlers.HandleKeepAliveServerboundPacket,
		ids:     packetIds{protocol1_20_2: 0x03, protocol1_20_3: 0x03, protocol1_20_5: 0x04, protocol1_21: 0x04, protocol1_21_2: 0x04, protocol1_21_4: 0x04, protocol1_21_5: 0x04, protocol1_21_6: 0x04, protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundCommon.KeepAliveServerboundPacket{}),
		decoder: serverboundCommon.DecodeKeepAliveServerboundPacket,
		handler: handlers.HandleKeepAliveServerboundPacket,
		ids:     packetIds{protocol1_18: 0x0F, protocol1_18_2: 0x0F, protocol1_19: 0x11, protocol1_19_1: 0x12, protocol1_19_3: 0x11, protocol1_19_4: 0x12, protocol1_20: 0x12, protocol1_20_2: 0x14, protocol1_20_3: 0x15, protocol1_20_5: 0x18, protocol1_21: 0x18, protocol1_21_2: 0x1A, protocol1_21_4: 0x1A, protocol1_21_5: 0x1A, protocol1_21_6: 0x1B, protocol1_21_7: 0x1B, protocol1_21_9: 0x1B, protocol1_21_11: 0x1B, protocol26_1: 0x1C, protocol26_2: 0x1C},
	},

	// What a joined client sends on its own, none of which needs a reaction
	// from a limbo.
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.AcceptTeleportationServerboundPacket{}),
		decoder: serverboundPlay.DecodeAcceptTeleportationServerboundPacket,
		ids:     packetIds{protocol1_18: 0x00, protocol1_18_2: 0x00, protocol1_19: 0x00, protocol1_19_1: 0x00, protocol1_19_3: 0x00, protocol1_19_4: 0x00, protocol1_20: 0x00, protocol1_20_2: 0x00, protocol1_20_3: 0x00, protocol1_20_5: 0x00, protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
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
		ids:     packetIds{protocol1_18: 0x11, protocol1_18_2: 0x11, protocol1_19: 0x13, protocol1_19_1: 0x14, protocol1_19_3: 0x13, protocol1_19_4: 0x14, protocol1_20: 0x14, protocol1_20_2: 0x16, protocol1_20_3: 0x17, protocol1_20_5: 0x1A, protocol1_21: 0x1A, protocol1_21_2: 0x1C, protocol1_21_4: 0x1C, protocol1_21_5: 0x1C, protocol1_21_6: 0x1D, protocol1_21_7: 0x1D, protocol1_21_9: 0x1D, protocol1_21_11: 0x1D, protocol26_1: 0x1E, protocol26_2: 0x1E},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerPositionRotationServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerPositionRotationServerboundPacket,
		handler: handlers.HandleMovePlayerPositionRotationServerboundPacket,
		ids:     packetIds{protocol1_18: 0x12, protocol1_18_2: 0x12, protocol1_19: 0x14, protocol1_19_1: 0x15, protocol1_19_3: 0x14, protocol1_19_4: 0x15, protocol1_20: 0x15, protocol1_20_2: 0x17, protocol1_20_3: 0x18, protocol1_20_5: 0x1B, protocol1_21: 0x1B, protocol1_21_2: 0x1D, protocol1_21_4: 0x1D, protocol1_21_5: 0x1D, protocol1_21_6: 0x1E, protocol1_21_7: 0x1E, protocol1_21_9: 0x1E, protocol1_21_11: 0x1E, protocol26_1: 0x1F, protocol26_2: 0x1F},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerRotationServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerRotationServerboundPacket,
		handler: handlers.HandleMovePlayerRotationServerboundPacket,
		ids:     packetIds{protocol1_18: 0x13, protocol1_18_2: 0x13, protocol1_19: 0x15, protocol1_19_1: 0x16, protocol1_19_3: 0x15, protocol1_19_4: 0x16, protocol1_20: 0x16, protocol1_20_2: 0x18, protocol1_20_3: 0x19, protocol1_20_5: 0x1C, protocol1_21: 0x1C, protocol1_21_2: 0x1E, protocol1_21_4: 0x1E, protocol1_21_5: 0x1E, protocol1_21_6: 0x1F, protocol1_21_7: 0x1F, protocol1_21_9: 0x1F, protocol1_21_11: 0x1F, protocol26_1: 0x20, protocol26_2: 0x20},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.MovePlayerStatusServerboundPacket{}),
		decoder: serverboundPlay.DecodeMovePlayerStatusServerboundPacket,
		handler: handlers.HandleMovePlayerStatusServerboundPacket,
		ids:     packetIds{protocol1_18: 0x14, protocol1_18_2: 0x14, protocol1_19: 0x16, protocol1_19_1: 0x17, protocol1_19_3: 0x16, protocol1_19_4: 0x17, protocol1_20: 0x17, protocol1_20_2: 0x19, protocol1_20_3: 0x1A, protocol1_20_5: 0x1D, protocol1_21: 0x1D, protocol1_21_2: 0x1F, protocol1_21_4: 0x1F, protocol1_21_5: 0x1F, protocol1_21_6: 0x20, protocol1_21_7: 0x20, protocol1_21_9: 0x20, protocol1_21_11: 0x20, protocol26_1: 0x21, protocol26_2: 0x21},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.PlayerInputServerboundPacket{}),
		decoder: serverboundPlay.DecodePlayerInputServerboundPacket,
		handler: handlers.HandlePlayerInputServerboundPacket,
		ids:     packetIds{protocol1_18: 0x1C, protocol1_18_2: 0x1C, protocol1_19: 0x1E, protocol1_19_1: 0x1F, protocol1_19_3: 0x1E, protocol1_19_4: 0x1F, protocol1_20: 0x1F, protocol1_20_2: 0x22, protocol1_20_3: 0x23, protocol1_20_5: 0x26, protocol1_21: 0x26, protocol1_21_2: 0x28, protocol1_21_4: 0x29, protocol1_21_5: 0x29, protocol1_21_6: 0x2A, protocol1_21_7: 0x2A, protocol1_21_9: 0x2A, protocol1_21_11: 0x2A, protocol26_1: 0x2B, protocol26_2: 0x2B},
	},
	{
		phase:   types.PhasePlay,
		packet:  reflect.TypeOf(serverboundPlay.SwingServerboundPacket{}),
		decoder: serverboundPlay.DecodeSwingServerboundPacket,
		handler: handlers.HandleSwingServerboundPacket,
		ids:     packetIds{protocol1_18: 0x2C, protocol1_18_2: 0x2C, protocol1_19: 0x2E, protocol1_19_1: 0x2F, protocol1_19_3: 0x2F, protocol1_19_4: 0x2F, protocol1_20: 0x2F, protocol1_20_2: 0x32, protocol1_20_3: 0x33, protocol1_20_5: 0x36, protocol1_21: 0x36, protocol1_21_2: 0x38, protocol1_21_4: 0x3A, protocol1_21_5: 0x3B, protocol1_21_6: 0x3C, protocol1_21_7: 0x3C, protocol1_21_9: 0x3C, protocol1_21_11: 0x3C, protocol26_1: 0x3F, protocol26_2: 0x3F},
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
// the game profile packet -- under the same id. 1.20.5 numbers everything as
// 1.21 does: the two clientbound packets 1.21 added, the custom report
// details and the server links, went in at the end of the play and
// configuration phases, behind everything this server sends. 1.20.3 has no
// cookies and no transfer, which 1.20.5 put in front of most of what this
// server sends: the play phase sits two lower from the game event on and
// three from the teleport, and the configuration phase one lower from the
// finish on, with the registry data two lower and the tags four. The login
// phase is numbered alike, since the cookie request went in at its end.
// 1.20.2 sits two lower than 1.20.3 through the play phase from the rotate
// head on -- 1.20.3 added the reset score packet at 0x42 and split the
// resource pack packet in two -- and one lower for the tags alone in the
// configuration phase, for the same split. The login phase is numbered alike.
// 1.20 numbers the play phase its own way: it still has the add player packet
// at 0x03, which 1.20.2 retired and which the add entity packet goes out as
// on 1.20, so the animate sits one higher; 1.20.2 added the two chunk batch
// packets in front of the game event, the pong response in front of the
// player list, and the start configuration in front of the teleport, so
// everything from the game event on sits one lower, from the player list on
// two, and the teleport three. 1.20 has no configuration phase, so the tags
// are a play packet there, at 0x6E behind everything else this server sends,
// and the registry data and the finish configuration have no id at all: the
// registries travel inside the play login. 1.19.4 is numbered as 1.20 is in
// every phase, absences included: 1.20 added no packet and retired none, in
// either direction. The login phase is numbered alike. 1.19.3 numbers the
// play phase its own way, absences included: 1.19.4 put the bundle delimiter
// at 0x00 in front of everything, and added the chunk biomes, the damage
// event and the hurt animation, so the add player and the animate sit one
// lower, everything from the game event to the teleport four, and the tags
// four as well, at 0x6A. The login phase is numbered alike. 1.19.1 numbers
// the play phase its own way as well: 1.19.3 retired the chat preview at
// 0x0C, the chat header and the chat preview display, and added the
// disguised chat and the enabled features, so the add player and the
// animate sit alike, the game event, the keep alive, the chunk and the login
// one higher, and everything from the player list to the tags, at 0x6B, two
// higher. The player list is one packet on 1.19.1, at 0x37, which both the
// remove and the update go out as, under actions of their own, and the
// 1.19.3 step rewrites each into its shape. The login phase is numbered
// alike. 1.19 numbers the play phase its own way below that: 1.19.1 added
// the custom chat completions at 0x15, the delete chat at 0x18 and the chat
// header at 0x32, so the add player and the animate sit alike, the game
// event, the keep alive, the chunk and the login two lower, and everything
// from the player list -- one packet on 1.19 as on 1.19.1, at 0x34 -- to the
// tags, at 0x68, three lower. The login phase is numbered alike. 1.18.2
// numbers the play phase its own way below that: it still has the add mob,
// the add painting and the add vibration signal packets, at 0x02, 0x03 and
// 0x05, which 1.19 retired -- the first two folded into the add entity --
// and the chat at 0x0F, which 1.19 split into the player chat at 0x30 and
// the system chat at 0x5F, and 1.19 added the chat preview at 0x0C, the
// server data at 0x3F and the chat preview display at 0x4B. So the add
// player sits two higher, at 0x04, and the animate three; the keep alive,
// the chunk and the login three higher; the player list -- one packet on
// 1.18.2 as on 1.19, at 0x36 -- the player position, the remove entities
// and the rotate head two higher; the chunk cache centre and the spawn
// position one higher; the entity metadata alike at 0x4D; and the teleport
// and the tags, at 0x67, one lower. The login phase is numbered alike. 1.18
// numbers every phase as 1.18.2 does, the two jars' registrations being the
// same list in the same order.
var clientboundPackets = []clientboundPacket{
	// Answered on protocol zero as well, for the same reason the requests are
	// read there: a client on a version this server does not speak still gets an
	// answer, and works out from the version in it that it cannot join.
	{
		phase:  types.PhaseStatus,
		packet: reflect.TypeOf(clientboundStatus.StatusResponseClientboundPacket{}),
		ids:    packetIds{protocolZero: 0x00, protocol1_18: 0x00, protocol1_18_2: 0x00, protocol1_19: 0x00, protocol1_19_1: 0x00, protocol1_19_3: 0x00, protocol1_19_4: 0x00, protocol1_20: 0x00, protocol1_20_2: 0x00, protocol1_20_3: 0x00, protocol1_20_5: 0x00, protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:  types.PhaseStatus,
		packet: reflect.TypeOf(clientboundStatus.PongResponseClientboundPacket{}),
		ids:    packetIds{protocolZero: 0x01, protocol1_18: 0x01, protocol1_18_2: 0x01, protocol1_19: 0x01, protocol1_19_1: 0x01, protocol1_19_3: 0x01, protocol1_19_4: 0x01, protocol1_20: 0x01, protocol1_20_2: 0x01, protocol1_20_3: 0x01, protocol1_20_5: 0x01, protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},

	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.DisconnectClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x00, protocol1_18_2: 0x00, protocol1_19: 0x00, protocol1_19_1: 0x00, protocol1_19_3: 0x00, protocol1_19_4: 0x00, protocol1_20: 0x00, protocol1_20_2: 0x00, protocol1_20_3: 0x00, protocol1_20_5: 0x00, protocol1_21: 0x00, protocol1_21_2: 0x00, protocol1_21_4: 0x00, protocol1_21_5: 0x00, protocol1_21_6: 0x00, protocol1_21_7: 0x00, protocol1_21_9: 0x00, protocol1_21_11: 0x00, protocol26_1: 0x00, protocol26_2: 0x00},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.EncryptionRequestClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x01, protocol1_18_2: 0x01, protocol1_19: 0x01, protocol1_19_1: 0x01, protocol1_19_3: 0x01, protocol1_19_4: 0x01, protocol1_20: 0x01, protocol1_20_2: 0x01, protocol1_20_3: 0x01, protocol1_20_5: 0x01, protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x02, protocol1_18_2: 0x02, protocol1_19: 0x02, protocol1_19_1: 0x02, protocol1_19_3: 0x02, protocol1_19_4: 0x02, protocol1_20: 0x02, protocol1_20_2: 0x02, protocol1_20_3: 0x02, protocol1_20_5: 0x02, protocol1_21: 0x02, protocol1_21_2: 0x02, protocol1_21_4: 0x02, protocol1_21_5: 0x02, protocol1_21_6: 0x02, protocol1_21_7: 0x02, protocol1_21_9: 0x02, protocol1_21_11: 0x02, protocol26_1: 0x02, protocol26_2: 0x02},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.SetCompressionClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x03, protocol1_18_2: 0x03, protocol1_19: 0x03, protocol1_19_1: 0x03, protocol1_19_3: 0x03, protocol1_19_4: 0x03, protocol1_20: 0x03, protocol1_20_2: 0x03, protocol1_20_3: 0x03, protocol1_20_5: 0x03, protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x03, protocol1_21_6: 0x03, protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},
	{
		phase:  types.PhaseLogin,
		packet: reflect.TypeOf(clientboundLogin.LoginPluginRequestClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x04, protocol1_18_2: 0x04, protocol1_19: 0x04, protocol1_19_1: 0x04, protocol1_19_3: 0x04, protocol1_19_4: 0x04, protocol1_20: 0x04, protocol1_20_2: 0x04, protocol1_20_3: 0x04, protocol1_20_5: 0x04, protocol1_21: 0x04, protocol1_21_2: 0x04, protocol1_21_4: 0x04, protocol1_21_5: 0x04, protocol1_21_6: 0x04, protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},

	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.FinishConfigurationClientboundPacket{}),
		ids:    packetIds{protocol1_20_2: 0x02, protocol1_20_3: 0x02, protocol1_20_5: 0x03, protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x03, protocol1_21_6: 0x03, protocol1_21_7: 0x03, protocol1_21_9: 0x03, protocol1_21_11: 0x03, protocol26_1: 0x03, protocol26_2: 0x03},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}),
		ids:    packetIds{protocol1_20_2: 0x03, protocol1_20_3: 0x03, protocol1_20_5: 0x04, protocol1_21: 0x04, protocol1_21_2: 0x04, protocol1_21_4: 0x04, protocol1_21_5: 0x04, protocol1_21_6: 0x04, protocol1_21_7: 0x04, protocol1_21_9: 0x04, protocol1_21_11: 0x04, protocol26_1: 0x04, protocol26_2: 0x04},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.RegistryDataClientboundPacket{}),
		ids:    packetIds{protocol1_20_2: 0x05, protocol1_20_3: 0x05, protocol1_20_5: 0x07, protocol1_21: 0x07, protocol1_21_2: 0x07, protocol1_21_4: 0x07, protocol1_21_5: 0x07, protocol1_21_6: 0x07, protocol1_21_7: 0x07, protocol1_21_9: 0x07, protocol1_21_11: 0x07, protocol26_1: 0x07, protocol26_2: 0x07},
	},
	// The same tags in the one phase 1.20, 1.19.4, 1.19.3 and 1.19.1 have to read them in: a
	// client before 1.20.2 has no configuration phase, and is sent them in
	// play right after the login, the way a vanilla server of that version
	// does.
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundConfiguration.UpdateTagsClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x67, protocol1_18_2: 0x67, protocol1_19: 0x68, protocol1_19_1: 0x6B, protocol1_19_3: 0x6A, protocol1_19_4: 0x6E, protocol1_20: 0x6E},
	},
	{
		phase:  types.PhaseConfiguration,
		packet: reflect.TypeOf(clientboundConfiguration.UpdateTagsClientboundPacket{}),
		ids:    packetIds{protocol1_20_2: 0x08, protocol1_20_3: 0x09, protocol1_20_5: 0x0D, protocol1_21: 0x0D, protocol1_21_2: 0x0D, protocol1_21_4: 0x0D, protocol1_21_5: 0x0D, protocol1_21_6: 0x0D, protocol1_21_7: 0x0D, protocol1_21_9: 0x0D, protocol1_21_11: 0x0D, protocol26_1: 0x0D, protocol26_2: 0x0D},
	},

	// 1.20 does not spawn a player from the add entity packet: 1.20.2 is where
	// that came to be, and before it the client only spawned one from the add
	// player packet. The ids below for 1.20, 1.19.4, 1.19.3 and 1.19.1 are
	// that packet's, and the 1.20.2 step rewrites the body into its shape to
	// go under it.
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x04, protocol1_18_2: 0x04, protocol1_19: 0x02, protocol1_19_1: 0x02, protocol1_19_3: 0x02, protocol1_19_4: 0x03, protocol1_20: 0x03, protocol1_20_2: 0x01, protocol1_20_3: 0x01, protocol1_20_5: 0x01, protocol1_21: 0x01, protocol1_21_2: 0x01, protocol1_21_4: 0x01, protocol1_21_5: 0x01, protocol1_21_6: 0x01, protocol1_21_7: 0x01, protocol1_21_9: 0x01, protocol1_21_11: 0x01, protocol26_1: 0x01, protocol26_2: 0x01},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.AnimateClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x06, protocol1_18_2: 0x06, protocol1_19: 0x03, protocol1_19_1: 0x03, protocol1_19_3: 0x03, protocol1_19_4: 0x04, protocol1_20: 0x04, protocol1_20_2: 0x03, protocol1_20_3: 0x03, protocol1_20_5: 0x03, protocol1_21: 0x03, protocol1_21_2: 0x03, protocol1_21_4: 0x03, protocol1_21_5: 0x02, protocol1_21_6: 0x02, protocol1_21_7: 0x02, protocol1_21_9: 0x02, protocol1_21_11: 0x02, protocol26_1: 0x02, protocol26_2: 0x02},
	},
	// 1.21 has no entity position sync: 1.21.2 introduced it, and before it
	// the teleport entity packet was how an entity was put where the server
	// says. The ids below for 1.21 and every version before it are the
	// teleport's, and the 1.21.2 step rewrites the body into the teleport's
	// shape to go under it.
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.EntityPositionSyncClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x62, protocol1_18_2: 0x62, protocol1_19: 0x63, protocol1_19_1: 0x66, protocol1_19_3: 0x64, protocol1_19_4: 0x68, protocol1_20: 0x68, protocol1_20_2: 0x6B, protocol1_20_3: 0x6D, protocol1_20_5: 0x70, protocol1_21: 0x70, protocol1_21_2: 0x20, protocol1_21_4: 0x20, protocol1_21_5: 0x1F, protocol1_21_6: 0x1F, protocol1_21_7: 0x1F, protocol1_21_9: 0x23, protocol1_21_11: 0x23, protocol26_1: 0x23, protocol26_2: 0x23},
	},
	// 1.20.2 has no event for what this server's one game event says: 1.20.3
	// is where the event that lets a joining client off its loading screen
	// appeared, and what 1.20.2 waits for instead is the default spawn
	// position packet. The ids below for 1.20.2 and every version before it
	// are that packet's, and the 1.20.3 step rewrites the body into its shape
	// to go under it.
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.GameEventClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x4B, protocol1_18_2: 0x4B, protocol1_19: 0x4A, protocol1_19_1: 0x4D, protocol1_19_3: 0x4C, protocol1_19_4: 0x50, protocol1_20: 0x50, protocol1_20_2: 0x52, protocol1_20_3: 0x20, protocol1_20_5: 0x22, protocol1_21: 0x22, protocol1_21_2: 0x23, protocol1_21_4: 0x23, protocol1_21_5: 0x22, protocol1_21_6: 0x22, protocol1_21_7: 0x22, protocol1_21_9: 0x26, protocol1_21_11: 0x26, protocol26_1: 0x26, protocol26_2: 0x26},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x21, protocol1_18_2: 0x21, protocol1_19: 0x1E, protocol1_19_1: 0x20, protocol1_19_3: 0x1F, protocol1_19_4: 0x23, protocol1_20: 0x23, protocol1_20_2: 0x24, protocol1_20_3: 0x24, protocol1_20_5: 0x26, protocol1_21: 0x26, protocol1_21_2: 0x27, protocol1_21_4: 0x27, protocol1_21_5: 0x26, protocol1_21_6: 0x26, protocol1_21_7: 0x26, protocol1_21_9: 0x2B, protocol1_21_11: 0x2B, protocol26_1: 0x2C, protocol26_2: 0x2C},
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
		ids:    packetIds{protocol1_18: 0x22, protocol1_18_2: 0x22, protocol1_19: 0x1F, protocol1_19_1: 0x21, protocol1_19_3: 0x20, protocol1_19_4: 0x24, protocol1_20: 0x24, protocol1_20_2: 0x25, protocol1_20_3: 0x25, protocol1_20_5: 0x27, protocol1_21: 0x27, protocol1_21_2: 0x28, protocol1_21_4: 0x28, protocol1_21_5: 0x27, protocol1_21_6: 0x27, protocol1_21_7: 0x27, protocol1_21_9: 0x2C, protocol1_21_11: 0x2C, protocol26_1: 0x2D, protocol26_2: 0x2D},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x26, protocol1_18_2: 0x26, protocol1_19: 0x23, protocol1_19_1: 0x25, protocol1_19_3: 0x24, protocol1_19_4: 0x28, protocol1_20: 0x28, protocol1_20_2: 0x29, protocol1_20_3: 0x29, protocol1_20_5: 0x2B, protocol1_21: 0x2B, protocol1_21_2: 0x2C, protocol1_21_4: 0x2C, protocol1_21_5: 0x2B, protocol1_21_6: 0x2B, protocol1_21_7: 0x2B, protocol1_21_9: 0x30, protocol1_21_11: 0x30, protocol26_1: 0x31, protocol26_2: 0x31},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerInfoRemoveClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x36, protocol1_18_2: 0x36, protocol1_19: 0x34, protocol1_19_1: 0x37, protocol1_19_3: 0x35, protocol1_19_4: 0x39, protocol1_20: 0x39, protocol1_20_2: 0x3B, protocol1_20_3: 0x3B, protocol1_20_5: 0x3D, protocol1_21: 0x3D, protocol1_21_2: 0x3F, protocol1_21_4: 0x3F, protocol1_21_5: 0x3E, protocol1_21_6: 0x3E, protocol1_21_7: 0x3E, protocol1_21_9: 0x43, protocol1_21_11: 0x43, protocol26_1: 0x45, protocol26_2: 0x45},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x36, protocol1_18_2: 0x36, protocol1_19: 0x34, protocol1_19_1: 0x37, protocol1_19_3: 0x36, protocol1_19_4: 0x3A, protocol1_20: 0x3A, protocol1_20_2: 0x3C, protocol1_20_3: 0x3C, protocol1_20_5: 0x3E, protocol1_21: 0x3E, protocol1_21_2: 0x40, protocol1_21_4: 0x40, protocol1_21_5: 0x3F, protocol1_21_6: 0x3F, protocol1_21_7: 0x3F, protocol1_21_9: 0x44, protocol1_21_11: 0x44, protocol26_1: 0x46, protocol26_2: 0x46},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x38, protocol1_18_2: 0x38, protocol1_19: 0x36, protocol1_19_1: 0x39, protocol1_19_3: 0x38, protocol1_19_4: 0x3C, protocol1_20: 0x3C, protocol1_20_2: 0x3E, protocol1_20_3: 0x3E, protocol1_20_5: 0x40, protocol1_21: 0x40, protocol1_21_2: 0x42, protocol1_21_4: 0x42, protocol1_21_5: 0x41, protocol1_21_6: 0x41, protocol1_21_7: 0x41, protocol1_21_9: 0x46, protocol1_21_11: 0x46, protocol26_1: 0x48, protocol26_2: 0x48},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.RemoveEntitiesClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x3A, protocol1_18_2: 0x3A, protocol1_19: 0x38, protocol1_19_1: 0x3B, protocol1_19_3: 0x3A, protocol1_19_4: 0x3E, protocol1_20: 0x3E, protocol1_20_2: 0x40, protocol1_20_3: 0x40, protocol1_20_5: 0x42, protocol1_21: 0x42, protocol1_21_2: 0x47, protocol1_21_4: 0x47, protocol1_21_5: 0x46, protocol1_21_6: 0x46, protocol1_21_7: 0x46, protocol1_21_9: 0x4B, protocol1_21_11: 0x4B, protocol26_1: 0x4D, protocol26_2: 0x4D},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.RotateHeadClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x3E, protocol1_18_2: 0x3E, protocol1_19: 0x3C, protocol1_19_1: 0x3F, protocol1_19_3: 0x3E, protocol1_19_4: 0x42, protocol1_20: 0x42, protocol1_20_2: 0x44, protocol1_20_3: 0x46, protocol1_20_5: 0x48, protocol1_21: 0x48, protocol1_21_2: 0x4D, protocol1_21_4: 0x4D, protocol1_21_5: 0x4C, protocol1_21_6: 0x4C, protocol1_21_7: 0x4C, protocol1_21_9: 0x51, protocol1_21_11: 0x51, protocol26_1: 0x53, protocol26_2: 0x53},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.SetChunkCacheCenterClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x49, protocol1_18_2: 0x49, protocol1_19: 0x48, protocol1_19_1: 0x4B, protocol1_19_3: 0x4A, protocol1_19_4: 0x4E, protocol1_20: 0x4E, protocol1_20_2: 0x50, protocol1_20_3: 0x52, protocol1_20_5: 0x54, protocol1_21: 0x54, protocol1_21_2: 0x58, protocol1_21_4: 0x58, protocol1_21_5: 0x57, protocol1_21_6: 0x57, protocol1_21_7: 0x57, protocol1_21_9: 0x5C, protocol1_21_11: 0x5C, protocol26_1: 0x5E, protocol26_2: 0x5E},
	},
	{
		phase:  types.PhasePlay,
		packet: reflect.TypeOf(clientboundPlay.SetEntityDataClientboundPacket{}),
		ids:    packetIds{protocol1_18: 0x4D, protocol1_18_2: 0x4D, protocol1_19: 0x4D, protocol1_19_1: 0x50, protocol1_19_3: 0x4E, protocol1_19_4: 0x52, protocol1_20: 0x52, protocol1_20_2: 0x54, protocol1_20_3: 0x56, protocol1_20_5: 0x58, protocol1_21: 0x58, protocol1_21_2: 0x5D, protocol1_21_4: 0x5D, protocol1_21_5: 0x5C, protocol1_21_6: 0x5C, protocol1_21_7: 0x5C, protocol1_21_9: 0x61, protocol1_21_11: 0x61, protocol26_1: 0x63, protocol26_2: 0x63},
	},
}

// A RegistryCodecSource hands out the registries a client before 1.20.2 reads
// out of its play login, encoded as the one compound that login carries.
//
// It is asked for here, of all places, because that compound is the one thing
// a transformer has to write that the body it is given does not carry: the
// login packet is encoded at the latest version, which sends the registries
// in the configuration phase and puts nothing of them in the login, so the
// 1.20.2 step's login transformer is built around the compound rather than
// reading it from anywhere. Package gamedata is the source there is; nil
// builds a registry that refuses every 1.20 login rather than sending one
// without its registries.
type RegistryCodecSource interface {
	RegistryCodecFor(version types.ProtocolVersion) []byte

	// DimensionTypeFor is the dimension type a version's play login spells
	// out, for the one version below 1.19 that reads it there rather than
	// as a name: see the 1.19 step's login transformer.
	DimensionTypeFor(version types.ProtocolVersion) []byte
}

// NewDefaultRegistry builds the registry every connection resolves its packets
// through: every packet this server speaks, the id each version gives it, and
// the transformers that carry bodies between neighbouring versions, one of
// which carries the registries registryCodecs hands out.
func NewDefaultRegistry(registryCodecs RegistryCodecSource) *Registry {
	registry := NewRegistry()
	registerPackets(registry, registryCodecs)

	return registry
}

func registerPackets(packetRegistry *Registry, registryCodecs RegistryCodecSource) {
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

	registerTransformers(packetRegistry, registryCodecs)
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
//
// The 1.21 step registers nothing, checked the same way as the 1.21.7 step:
// every packet this server speaks is wire-identical in 766 and 767, the
// chunk packet and the entity metadata included, and the entity type registry
// numbers the player alike in both, as does the block state table. 1.21's
// additions all sit behind what this server sends. The rest of the bump is
// data: three synchronized registries fewer, and 1.20.6's own content in the
// ones it keeps.
//
// The 1.20.5 step carries five, all downwards, because 1.20.5 is where the
// login was reworked on both sides of the configuration phase. The
// encryption request, whose should authenticate flag 1.20.5 appended; the
// login success, whose strict error handling flag 1.20.5 is where it first
// appeared; the play login, whose dimension type 1.20.5 turned from a name
// into a registry index and whose enforces secure chat flag it moved in from
// the server data packet; the add entity packet, for the registry renumbering
// at the bottom of the chain; and the entity metadata, whose pose serializer
// 1.20.5's particles serializer pushed up one. The one change this step has
// that no transformer can make is the registry data itself: 1.20.3 reads
// every registry in one packet, as one compound, where 1.20.5 reads one
// packet per registry, so package gamedata encodes that version's set in
// the older shape rather than rewriting a packet into several. Everything
// else this server speaks is wire-identical in 765 and 766, the chunk packet
// and the tags included, and the rest of the bump is data: two synchronized
// registries fewer, 1.20.4's own content in the ones it keeps, and the block
// state table from its jar.
//
// The 1.20.3 step carries two, both downwards. The add entity packet, for
// the registry renumbering once more: 1.20.3 added the breeze and its wind
// charge, both before the player. And the game event, which 1.20.2 has no
// event for: 1.20.3 is where the event that lets a joining client off its
// loading screen appeared, and 1.20.2 reads the default spawn position
// packet as that, so the id table names that packet for it and the step
// rewrites the body into its shape. Everything else this server speaks is
// wire-identical in 764 and 765, checked the same way as the 1.20.5 step --
// 1.20.3 turned text components from JSON into NBT, but the only component
// this server sends is the login disconnect's, which stayed JSON -- and the
// rest of the bump is data: the same six registries with 1.20.2's own tags,
// and the block state table from its jar.
//
// The 1.20.2 step carries four, because 1.20.2 is where the configuration
// phase appeared and with it everything a 1.20 client does without.
// Downwards: the play login, which 1.20 lays out in its own order and reads
// the registries out of, as the one compound the configuration phase went on
// to carry in a packet -- which is why the transformer is built around what
// registryCodecs hands out, the one body rewrite that needs something the
// body does not carry; the add entity packet, which 1.20 has no way to spawn
// a player from and reads as the add player packet instead; and the chunk
// packet, whose heightmap compound 1.20 reads with a root name, since 1.20.2
// is where network NBT lost its name. Upwards: the login start, whose uuid
// 1.20 sends as an optional. Everything else this server speaks is
// wire-identical in 763 and 764, checked the same way -- the entity
// metadata's serializers are registered in the same order, and the tags,
// which 1.20 reads as a play packet, are the same body -- and the rest of
// the bump is data: the same six registries with 1.20's own content, and
// the block state table from its jar.
//
// The 1.20 step carries two, both downwards, for the two packets 1.20 laid
// out differently from 1.19.4: the play login, to whose end 1.20 appended
// the portal cooldown and whose registries are 1.19.4's own -- so this
// rewrite is built around what registryCodecs hands out for 1.19.4, the
// way the 1.20.2 step's is around 1.20's -- and the chunk packet, whose
// light data 1.19.4 opens with the trust edges flag 1.20 took off. Every
// packet this server speaks is numbered alike in 762 and 763, in every
// phase, and everything else is wire-identical, checked the same way; the
// rest of the bump is data: the same six registries, two of them empty
// because 1.19.4 keeps the armor trims behind a feature flag, with 1.19.4's
// own content, and the block state table from its jar.
//
// The 1.19.4 step carries three, all downwards, for the three packets 1.19.4
// laid out differently from 1.19.3: the play login, whose registries are
// 1.19.3's own -- three to 1.19.4's six, since 1.19.4 is where the damage
// types and the armor trims appeared -- so this rewrite is built around what
// registryCodecs hands out for 1.19.3, the way the two steps above it are;
// the player position, to whose end 1.19.3 puts the dismount vehicle flag
// 1.19.4 took off; and the entity metadata, whose pose serializer 1.19.4's
// optional block state serializer pushed up one. 1.19.4 renumbered the play
// phase in both directions, which the id tables say, and everything else is
// wire-identical in 761 and 762, checked the same way; the rest of the bump
// is data: the three registries with 1.19.3's own content, a biome whose
// climate 1.19.3 spells its own way, and the block state table from its jar.
//
// The 1.19.3 step carries six, four downwards and two upwards, because
// 1.19.3 is where the player list and the login were reworked. Downwards:
// the two player list packets, which 1.19.1 reads as one packet under an
// action of its own -- the update goes out as an add of everything the entry
// carries, or as the one change it makes, and the remove as the remove --
// the entity metadata, whose pose serializer 1.19.3's long serializer
// pushed up one, and the play login, laid out alike but carrying 1.19.1's
// own registries, so this rewrite is built around what registryCodecs hands
// out for 1.19.1, the way the three steps above it are. Upwards: the login start, which 1.19.1 sends with the
// profile key 1.19.3 took off, and the encryption response, which a 1.19.1
// client holding that key answers with a signature rather than an encrypted
// challenge -- a signature this server has no key left to check, so the
// response goes on without a challenge and the connection lets it through on
// that version alone, on the session server's word. 1.19.3 renumbered the
// play phase in both directions, which the id tables say, and everything else
// is wire-identical in 760 and 761, checked the same way; the rest of the
// bump is data: the same three registries with 1.19.1's own content, and the
// block state table from its jar.
//
// The 1.19.1 step carries two, one each way, for the two packets 1.19.1
// laid out differently from 1.19: the play login, laid out alike but
// carrying 1.19's own registries, so this rewrite is built around what
// registryCodecs hands out for 1.19, the way the four steps above it are;
// and the login start, onto whose end 1.19.1 put the optional uuid. 1.19.1
// renumbered the play phase in both directions, which the id tables say,
// and everything else is wire-identical in 759 and 760, checked the same
// way; the rest of the bump is data: the same three registries with the
// chat types laid out 1.19's own way.
//
// The 1.19 step carries five, three downwards and two upwards, because 1.19
// is where the profile key arrived. Downwards: the play login, whose
// registries are 1.18.2's own -- two to 1.19's three, since 1.19 is where
// the chat types joined them -- and which 1.18.2 reads with the dimension
// type spelled out in it rather than named and with no death location on
// the end, so this rewrite is built around what registryCodecs hands out
// for 1.18.2, the registries and the dimension type both; the login
// success, onto which 1.19 put the profile's properties; and the player
// list, whose added player 1.19 gave the profile key. Upwards: the login
// start, onto which 1.19 put the optional profile key, and the encryption
// response, whose challenge 1.19 lets a client sign, with a flag in front
// saying whether it did. 1.19 renumbered the play phase in both directions,
// which the id tables say, and everything else is wire-identical in 758 and
// 759, checked the same way; the rest of the bump is data: the two
// registries with 1.18.2's own dimension type and biome, six tag sets to
// 1.19's ten, and the block state table from its jar.
//
// The 1.18.2 step carries one, downwards: the play login, whose registries
// and whose spelled-out dimension type are 1.18's own -- the same two
// registries, and the dimension type differing by one field, since 1.18.2
// is where tag keys appeared and the infiniburn field took its hash -- so
// this rewrite is built around what registryCodecs hands out for 1.18, the
// registries and the dimension type both. Nothing was renumbered and
// nothing else is laid out differently: every packet this server speaks is
// wire-identical in 757 and 758, checked the same way, and the block state
// table is byte-identical too; the rest of the bump is data: five tag sets
// to 1.18.2's six, since 1.18.2 is where the biomes came to have tags, and
// one block tag fewer.
func registerTransformers(packetRegistry *Registry, registryCodecs RegistryCodecSource) {
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

	// 1.20.5 reworked the login on both sides of the configuration phase,
	// added four entity types in front of the player and a metadata
	// serializer in front of the pose.
	packetRegistry.RegisterDowngrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_20_5,
		reflect.TypeOf(clientboundLogin.EncryptionRequestClientboundPacket{}),
		transformers.DowngradeEncryptionRequestTo1_20_3,
	)

	packetRegistry.RegisterDowngrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_20_5,
		reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}),
		transformers.DowngradeLoginSuccessTo1_20_3,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20_5,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_20_3,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20_5,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_20_3,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20_5,
		reflect.TypeOf(clientboundPlay.SetEntityDataClientboundPacket{}),
		transformers.DowngradeSetEntityDataTo1_20_3,
	)

	// 1.20.3 added two entity types in front of the player, and the game
	// event 1.20.2 reads as a default spawn position.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20_3,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_20_2,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20_3,
		reflect.TypeOf(clientboundPlay.GameEventClientboundPacket{}),
		transformers.DowngradeGameEventTo1_20_2,
	)

	// 1.20.2 put the registries into the configuration phase and out of the
	// play login, folded the player into the add entity packet, and took the
	// root name off every NBT on the wire.
	var registryCodec1_20, registryCodec1_19_4, registryCodec1_19_3, registryCodec1_19_1, registryCodec1_19, registryCodec1_18_2, dimensionType1_18_2, registryCodec1_18, dimensionType1_18 []byte
	if registryCodecs != nil {
		registryCodec1_20 = registryCodecs.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_20)
		registryCodec1_19_4 = registryCodecs.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_19_4)
		registryCodec1_19_3 = registryCodecs.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_19_3)
		registryCodec1_19_1 = registryCodecs.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_19_1)
		registryCodec1_19 = registryCodecs.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_19)
		registryCodec1_18_2 = registryCodecs.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_18_2)
		dimensionType1_18_2 = registryCodecs.DimensionTypeFor(types.ProtocolVersions.MINECRAFT_1_18_2)
		registryCodec1_18 = registryCodecs.RegistryCodecFor(types.ProtocolVersions.MINECRAFT_1_18)
		dimensionType1_18 = registryCodecs.DimensionTypeFor(types.ProtocolVersions.MINECRAFT_1_18)
	}

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20_2,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_20(registryCodec1_20),
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20_2,
		reflect.TypeOf(clientboundPlay.AddEntityClientboundPacket{}),
		transformers.DowngradeAddEntityTo1_20,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20_2,
		reflect.TypeOf(clientboundPlay.LevelChunkWithLightClientboundPacket{}),
		transformers.DowngradeLevelChunkWithLightTo1_20,
	)

	// 1.20 appended the portal cooldown to the play login, and took the trust
	// edges flag off the chunk packet's light data.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_19_4(registryCodec1_19_4),
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_20,
		reflect.TypeOf(clientboundPlay.LevelChunkWithLightClientboundPacket{}),
		transformers.DowngradeLevelChunkWithLightTo1_19_4,
	)

	// 1.19.4 is where the damage types and the armor trims joined the
	// registries a play login carries, where the player position lost its
	// dismount vehicle flag, and where the optional block state serializer
	// pushed the pose serializer up one.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19_4,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_19_3(registryCodec1_19_3),
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19_4,
		reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}),
		transformers.DowngradePlayerPositionTo1_19_3,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19_4,
		reflect.TypeOf(clientboundPlay.SetEntityDataClientboundPacket{}),
		transformers.DowngradeSetEntityDataTo1_19_3,
	)

	// 1.19.3 is where the player list packet was split in two and given its
	// mask of actions, and where the long serializer pushed the pose
	// serializer up one; the play login is laid out alike, with 1.19.1's own
	// registries in it.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19_3,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_19_1(registryCodec1_19_1),
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19_3,
		reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}),
		transformers.DowngradePlayerInfoUpdateTo1_19_1,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19_3,
		reflect.TypeOf(clientboundPlay.PlayerInfoRemoveClientboundPacket{}),
		transformers.DowngradePlayerInfoRemoveTo1_19_1,
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19_3,
		reflect.TypeOf(clientboundPlay.SetEntityDataClientboundPacket{}),
		transformers.DowngradeSetEntityDataTo1_19_1,
	)

	// 1.19.1 is where the chat types were reworked, and with them the one
	// registry a play login carries that 1.19 reads differently; the login
	// is laid out alike, with 1.19's own registries in it.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19_1,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_19(registryCodec1_19),
	)

	// 1.19 is where the profile key arrived: the login success took on the
	// profile's properties and the player list's added player the key, and
	// the play login is 1.18.2's own below it, with its registries, with the
	// dimension type spelled out and with no death location.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_18_2(registryCodec1_18_2, dimensionType1_18_2),
	)

	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_19,
		reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}),
		transformers.DowngradePlayerInfoUpdateTo1_18_2,
	)

	packetRegistry.RegisterDowngrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_19,
		reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}),
		transformers.DowngradeLoginSuccessTo1_18_2,
	)

	// 1.18.2 is where tag keys arrived, and with them the hash on the
	// dimension type's infiniburn field: the play login is 1.18's own below
	// it, with its registries and with its dimension type spelled out.
	packetRegistry.RegisterDowngrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_18_2,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}),
		transformers.DowngradePlayLoginTo1_18(registryCodec1_18, dimensionType1_18),
	)

	// The seven packets this server reads that an older version lays out
	// differently. An upgrade is registered against the version the client
	// sent it at.
	packetRegistry.RegisterUpgrade(
		types.PhasePlay,
		types.ProtocolVersions.MINECRAFT_1_21,
		reflect.TypeOf(serverboundPlay.PlayerInputServerboundPacket{}),
		transformers.UpgradePlayerInputFrom1_21,
	)

	packetRegistry.RegisterUpgrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_20,
		reflect.TypeOf(login.LoginStartServerboundPacket{}),
		transformers.UpgradeLoginStartFrom1_20,
	)

	// 1.19.3 took the profile key off the login start, and with it the
	// client's way of signing the encryption challenge.
	packetRegistry.RegisterUpgrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_19_1,
		reflect.TypeOf(login.LoginStartServerboundPacket{}),
		transformers.UpgradeLoginStartFrom1_19_1,
	)

	packetRegistry.RegisterUpgrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_19_1,
		reflect.TypeOf(login.EncryptionResponseServerboundPacket{}),
		transformers.UpgradeEncryptionResponseFrom1_19_1,
	)

	// 1.19.1 put the optional uuid onto the end of the login start, behind
	// the profile key 1.19 already sends.
	packetRegistry.RegisterUpgrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_19,
		reflect.TypeOf(login.LoginStartServerboundPacket{}),
		transformers.UpgradeLoginStartFrom1_19,
	)

	// 1.19 put the optional profile key onto the end of the login start, and
	// a flag in front of the encryption response's challenge saying whether
	// the client signed it instead; 1.18.2 has neither.
	packetRegistry.RegisterUpgrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_18_2,
		reflect.TypeOf(login.LoginStartServerboundPacket{}),
		transformers.UpgradeLoginStartFrom1_18_2,
	)

	packetRegistry.RegisterUpgrade(
		types.PhaseLogin,
		types.ProtocolVersions.MINECRAFT_1_18_2,
		reflect.TypeOf(login.EncryptionResponseServerboundPacket{}),
		transformers.UpgradeEncryptionResponseFrom1_18_2,
	)
}
