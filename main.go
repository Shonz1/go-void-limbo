package main

import (
	"bytes"
	"errors"
	"fmt"
	"go-void-limbo/gamedata"
	"go-void-limbo/handlers"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	clientboundPlay "go-void-limbo/packets/clientbound/play"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/packets/serverbound/login"
	serverboundPlay "go-void-limbo/packets/serverbound/play"
	"go-void-limbo/registries"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"io"
	"log/slog"
	"net"
	"os"
	"reflect"
)

const address = ":25565"

// maxPacketSize is the largest packet body the protocol allows (2^21 - 1 bytes).
const maxPacketSize = 2097151

// packetLogBlacklist holds the packet types that are never logged, in either
// direction. A joined client sends some of these on every tick, and at that rate
// they bury everything else the log has to say.
var packetLogBlacklist = map[reflect.Type]bool{
	reflect.TypeOf(serverboundPlay.ClientTickEndServerboundPacket{}): true,
}

// logPacket records a packet crossing the connection, unless its type is
// blacklisted. Every connection carries the same traffic, so this is detail one
// asks for rather than detail one is told.
func logPacket(message string, packet any) {
	packetType := reflect.TypeOf(packet)
	if packetType != nil && packetType.Kind() == reflect.Pointer {
		packetType = packetType.Elem()
	}

	if packetType != nil && packetLogBlacklist[packetType] {
		return
	}

	slog.Debug(message, "packet", packet)
}

func VarIntSize(value int32) int {
	uvalue := uint32(value)
	size := 1
	for (uvalue & ^uint32(0x7F)) != 0 {
		size++
		uvalue >>= 7
	}
	return size
}

type MinecraftClient struct {
	protocolVersion types.ProtocolVersion
	phase           types.Phase
	profile         types.GameProfile
	conn            net.Conn
	stream          *streams.MinecraftStream
	packetRegistry  *registries.PacketRegistry
	gameRegistries  *gamedata.Provider
}

func (c *MinecraftClient) RegistryPackets() []types.ClientboundPacket {
	return c.gameRegistries.PacketsFor(c.protocolVersion)
}

func (c *MinecraftClient) ProtocolVersion() types.ProtocolVersion {
	return c.protocolVersion
}

func (c *MinecraftClient) SetProtocolVersion(protocolVersion types.ProtocolVersion) {
	c.protocolVersion = protocolVersion
}

func (c *MinecraftClient) Phase() types.Phase {
	return c.phase
}

func (c *MinecraftClient) SetPhase(phase types.Phase) {
	c.phase = phase
}

func (c *MinecraftClient) Profile() types.GameProfile {
	return c.profile
}

func (c *MinecraftClient) SetProfile(profile types.GameProfile) {
	c.profile = profile
}

// ReadPacket decodes the next packet and returns the handler registered for it,
// which may be nil when the packet needs no reaction. The packet body is consumed
// from the connection in full before decoding, so an unknown packet id or a failed
// decode cannot desynchronize subsequent reads.
func (c *MinecraftClient) ReadPacket() (types.ServerboundPacket, types.PacketHandler, error) {
	length, err := c.stream.ReadVarInt()
	if err != nil {
		return nil, nil, err
	}

	if length < 1 || length > maxPacketSize {
		return nil, nil, fmt.Errorf("invalid packet length: %d", length)
	}

	body, err := c.stream.ReadBytes(length)
	if err != nil {
		return nil, nil, err
	}

	bodyStream := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	packetId, err := bodyStream.ReadVarInt()
	if err != nil {
		return nil, nil, err
	}

	entry, ok := c.packetRegistry.GetServerbound(c.phase, c.protocolVersion, packetId)
	if !ok || entry.Decoder == nil {
		return nil, nil, fmt.Errorf("unknown packet id: %d", packetId)
	}

	packet, err := entry.Decoder(bodyStream)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode packet: %w", err)
	}

	logPacket("packet received", packet)

	return packet, entry.Handler, nil
}

func (c *MinecraftClient) WritePacket(packet types.ClientboundPacket) error {
	if packet == nil {
		return errors.New("packet is nil")
	}

	packetId := c.packetRegistry.GetClientboundId(c.phase, reflect.TypeOf(packet).Elem(), c.protocolVersion)
	if packetId == -1 {
		return errors.New("unknown packet id")
	}

	buf := new(bytes.Buffer)
	tempStream := streams.NewMinecraftStreamFromBuffer(buf)

	err := packet.Encode(tempStream)
	if err != nil {
		return err
	}

	err = tempStream.Flush()
	if err != nil {
		return err
	}

	err = c.stream.WriteVarInt(int32(buf.Len() + VarIntSize(packetId)))
	if err != nil {
		return err
	}

	err = c.stream.WriteVarInt(packetId)
	if err != nil {
		return err
	}

	err = c.stream.WriteBytes(buf.Bytes())
	if err != nil {
		return err
	}

	err = c.stream.Flush()
	if err != nil {
		return err
	}

	logPacket("packet sent", packet)

	return nil
}

// configureLogging sets the level the default logger keeps, read from LOG_LEVEL
// as one of DEBUG, INFO, WARN or ERROR. Packet traffic is logged at DEBUG, so it
// is silent until asked for.
func configureLogging() {
	level := slog.LevelInfo
	unrecognized := ""

	if raw, ok := os.LookupEnv("LOG_LEVEL"); ok {
		if err := level.UnmarshalText([]byte(raw)); err != nil {
			level = slog.LevelInfo
			unrecognized = raw
		}
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	if unrecognized != "" {
		slog.Warn("unrecognized LOG_LEVEL, falling back to INFO", "value", unrecognized)
	}
}

func main() {
	configureLogging()

	packetRegistry := registries.NewPacketRegistry()

	registerPackets(packetRegistry)

	gameRegistries, err := gamedata.NewDefaultProvider()
	if err != nil {
		slog.Error("failed to encode game registries", "err", err)
		return
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		slog.Error("failed to start server", "err", err)
		return
	}

	defer listener.Close()

	slog.Info("TCP server is running", "address", address)

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "err", err)
			continue
		}

		go handleConnection(conn, packetRegistry, gameRegistries)
	}
}

func registerPackets(packetRegistry *registries.PacketRegistry) {
	packetRegistry.RegisterServerbound(types.PhaseHandshake, types.ProtocolVersions.ZERO, 0x00, handshake.DecodeHandshakeServerboundPacket, handlers.HandleHandshakeServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x00, login.DecodeLoginStartServerboundPacket, handlers.HandleLoginStartServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x03, login.DecodeLoginAcknowledgedServerboundPacket, handlers.HandleLoginAcknowledgedServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhaseConfiguration, types.ProtocolVersions.MINECRAFT_26_2, 0x03, configuration.DecodeAcknowledgeFinishConfigurationServerboundPacket, handlers.HandleAcknowledgeFinishConfigurationServerboundPacket)

	// What a joined client sends on its own. None of it needs a reaction from a
	// limbo, but a packet with no decoder is one the read loop can only report
	// as an unknown id, and the client sends these every tick.
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x00, serverboundPlay.DecodeAcceptTeleportationServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x0D, serverboundPlay.DecodeClientTickEndServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x1E, serverboundPlay.DecodeMovePlayerPositionServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x1F, serverboundPlay.DecodeMovePlayerPositionRotationServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x20, serverboundPlay.DecodeMovePlayerRotationServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x21, serverboundPlay.DecodeMovePlayerStatusServerboundPacket, nil)
	packetRegistry.RegisterServerbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, 0x2C, serverboundPlay.DecodePlayerLoadedServerboundPacket, nil)

	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.DisconnectClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x00)
	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x02)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.RegistryDataClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x07)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.UpdateTagsClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x0D)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.FinishConfigurationClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x03)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundPlay.GameEventClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x26)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x31)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundPlay.PlayerInfoUpdateClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x46)
	packetRegistry.RegisterClientbound(types.PhasePlay, reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x48)
}

func handleConnection(conn net.Conn, packetRegistry *registries.PacketRegistry, gameRegistries *gamedata.Provider) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	slog.Info("new client connected", "addr", remoteAddr)

	mc := &MinecraftClient{protocolVersion: types.ProtocolVersions.ZERO, phase: types.PhaseHandshake, conn: conn, stream: streams.NewMinecraftStreamFromNetConn(conn), packetRegistry: packetRegistry, gameRegistries: gameRegistries}

	for {
		packet, handler, err := mc.ReadPacket()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			slog.Error("failed to read packet", "err", err)
			continue
		}

		if handler == nil {
			continue
		}

		if err := handler(mc, packet); err != nil {
			slog.Error("failed to handle packet", "packet", packet, "err", err)
		}
	}
}
