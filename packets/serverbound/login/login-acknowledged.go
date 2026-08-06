package login

import (
	"go-void-limbo/streams"
	"go-void-limbo/types"
)

// LoginAcknowledgedServerboundPacket confirms the client accepted the login
// success packet. It carries no fields.
type LoginAcknowledgedServerboundPacket struct{}

func (p *LoginAcknowledgedServerboundPacket) String() string {
	return "LoginAcknowledgedServerboundPacket{}"
}

func DecodeLoginAcknowledgedServerboundPacket(_ *streams.MinecraftStream) (types.ServerboundPacket, error) {
	return &LoginAcknowledgedServerboundPacket{}, nil
}
