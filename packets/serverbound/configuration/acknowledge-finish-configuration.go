package configuration

import (
	"go-void-limbo/streams"
	"go-void-limbo/types"
)

// AcknowledgeFinishConfigurationServerboundPacket confirms the client left the
// configuration phase. It carries no fields.
type AcknowledgeFinishConfigurationServerboundPacket struct{}

func (p *AcknowledgeFinishConfigurationServerboundPacket) String() string {
	return "AcknowledgeFinishConfigurationServerboundPacket{}"
}

func DecodeAcknowledgeFinishConfigurationServerboundPacket(_ *streams.MinecraftStream) (types.ServerboundPacket, error) {
	return &AcknowledgeFinishConfigurationServerboundPacket{}, nil
}
