package handshake

import (
	"fmt"
	"go-void-limbo/streams"
	"go-void-limbo/types"
)

type HandshakeServerboundPacket struct {
	ProtocolVersion int32
	ServerAddress   string
	ServerPort      int16
	Intent          int32
}

func (p *HandshakeServerboundPacket) ToString() string {
	return fmt.Sprintf("HandshakeServerboundPacket %v", p)
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
