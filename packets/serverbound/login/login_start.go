package login

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

type LoginStartServerboundPacket struct {
	Name string
	Uuid string
}

func (p *LoginStartServerboundPacket) String() string {
	return fmt.Sprintf("LoginStartServerboundPacket{Name:%s Uuid:%s}", p.Name, p.Uuid)
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
