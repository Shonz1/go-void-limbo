package main

import (
	"bytes"
	"errors"
	"fmt"
	clientboundLogin "go-void-limbo/packets/clientbound/login"
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
	ProtocolVersion types.ProtocolVersion
	Phase           types.Phase
	conn            net.Conn
	stream          *streams.MinecraftStream
	packetRegistry  *registries.PacketRegistry
}

func (c *MinecraftClient) ReadPacket() (types.ServerboundPacket, error) {
	_, err := c.stream.ReadVarInt()
	if err != nil {
		return nil, err
	}

	packetId, err := c.stream.ReadVarInt()
	if err != nil {
		return nil, err
	}

	packetDecoder := c.packetRegistry.GetServerbound(c.Phase, c.ProtocolVersion, packetId)
	if packetDecoder == nil {
		return nil, fmt.Errorf("unknown packet id: %d", packetId)
	}

	packet, err := packetDecoder(c.stream)
	if err != nil {
		return nil, fmt.Errorf("failed to decode packet: %w", err)
	}

	slog.Info("packet received", "packet", packet.ToString())

	return packet, nil
}

func (c *MinecraftClient) WritePacket(packet types.ClientboundPacket) error {
	if packet == nil {
		return errors.New("packet is nil")
	}

	packetId := c.packetRegistry.GetClientboundId(c.Phase, reflect.TypeOf(packet).Elem(), c.ProtocolVersion)
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

	return c.stream.Flush()
}

func main() {
	packetRegistry := registries.NewPacketRegistry()

	registerPackets(packetRegistry)

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

		go handleConnection(conn, packetRegistry)
	}
}

func registerPackets(packetRegistry *registries.PacketRegistry) {
	packetRegistry.RegisterServerbound(types.PhaseHandshake, types.ProtocolVersions.ZERO, 0x00, handshake.DecodeHandshakeServerboundPacket)
	packetRegistry.RegisterServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x00, login.DecodeLoginStartServerboundPacket)

	packetRegistry.RegisterClientbound(types.PhaseLogin, reflect.TypeOf(clientboundLogin.DisconnectClientboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2, 0x00)
}

func handleConnection(conn net.Conn, packetRegistry *registries.PacketRegistry) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	slog.Info("new client connected", "addr", remoteAddr)

	mc := &MinecraftClient{ProtocolVersion: types.ProtocolVersions.ZERO, Phase: types.PhaseHandshake, conn: conn, stream: streams.NewMinecraftStreamFromNetConn(conn), packetRegistry: packetRegistry}

	for {
		packet, err := mc.ReadPacket()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return
			}

			slog.Error("failed to read packet size", "err", err)
			continue
		}

		switch typedPacket := packet.(type) {
		case *handshake.HandshakeServerboundPacket:
			mc.ProtocolVersion = types.GetProtocolVersionById(types.ProtocolId(typedPacket.ProtocolVersion))
			mc.Phase = types.Phase(typedPacket.Intent)

		case *login.LoginStartServerboundPacket:
			p := clientboundLogin.DisconnectClientboundPacket{Reason: `{"text": "TODO"}`}
			err = mc.WritePacket(&p)
			if err != nil {
				slog.Error("failed to encode DisconnectClientboundPacket", "err", err)
				continue
			}
		}
	}
}
