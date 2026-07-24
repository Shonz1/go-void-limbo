package login

import (
	"fmt"
	"go-void-limbo/streams"
	"go-void-limbo/types"
)

type LoginStartServerboundPacket struct {
	Name string
	Uuid string
}

func (p *LoginStartServerboundPacket) ToString() string {
	return fmt.Sprintf("LoginStartServerboundPacket %v", p)
}

func DecodeLoginStartServerboundPacket(minecraftStream *streams.MinecraftStream) (types.ServerboundPacket, error) {
	name, err := minecraftStream.ReadString()
	if err != nil {
		return nil, err
	}

	uuid, err := minecraftStream.ReadUuid()
	if err != nil {
		return nil, err
	}

	return &LoginStartServerboundPacket{Name: name, Uuid: uuid}, nil
}
