package configuration

import (
	"go-void-limbo/streams"
)

// FinishConfigurationClientboundPacket asks the client to leave the
// configuration phase and enter play. It carries no fields.
type FinishConfigurationClientboundPacket struct{}

func (p *FinishConfigurationClientboundPacket) String() string {
	return "FinishConfigurationClientboundPacket{}"
}

func (p *FinishConfigurationClientboundPacket) Encode(_ *streams.MinecraftStream) error {
	return nil
}
