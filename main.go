package main

import (
	"bytes"
	"errors"
	"fmt"
	"go-void-limbo/gamedata"
	"go-void-limbo/handlers"
	clientboundConfiguration "go-void-limbo/packets/clientbound/configuration"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/packets/serverbound/handshake"
	"go-void-limbo/packets/serverbound/login"
	"go-void-limbo/registries"
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"io"
	"log/slog"
	"net"
	"reflect"
)

const address = ":25565"

// maxPacketSize is the largest packet body the protocol allows (2^21 - 1 bytes).
const maxPacketSize = 2097151

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

	slog.Info("packet received", "packet", packet)

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

	slog.Info("packet sent", "packet", packet)

	return nil
}

func main() {
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

	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.DisconnectClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x00)
	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.LoginSuccessClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x02)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.RegistryDataClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x07)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.UpdateTagsClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x0D)
	packetRegistry.RegisterClientbound(types.PhaseConfiguration, reflect.TypeOf(clientboundConfiguration.FinishConfigurationClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x03)
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
