package handshake

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

type HandshakeServerboundPacket struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      int16
	Intent          int32
}

func (p *HandshakeServerboundPacket) String() string {
	return fmt.Sprintf("HandshakeServerboundPacket{ProtocolVersion:%d ServerAddress:%s ServerPort:%d Intent:%d}", p.ProtocolVersion, p.ServerAddress, p.ServerPort, p.Intent)
}

func DecodeHandshakeServerboundPacket(minecraftStream *streams.MinecraftStream) (types.ServerboundPacket, error) {
	protocolVersion, err := minecraftStream.ReadVarInt()
	if err != nil {
		return nil, err
	}

	serverAddress, err := minecraftStream.ReadString()
	if err != nil {
		return nil, err
	}

	serverPort, err := minecraftStream.ReadShort()
	if err != nil {
		return nil, err
	}

	intent, err := minecraftStream.ReadVarInt()
	if err != nil {
		return nil, err
	}

	return &HandshakeServerboundPacket{ProtocolVersion: protocolVersion, ServerAddress: serverAddress, ServerPort: serverPort, Intent: intent}, nil
}
