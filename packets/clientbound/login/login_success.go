package login

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

type LoginSuccessClientboundPacket struct {
	Profile   types.GameProfile
	SessionId string
}

func (p *LoginSuccessClientboundPacket) String() string {
	return fmt.Sprintf("LoginSuccessClientboundPacket{Profile:%s SessionId:%s}", p.Profile, p.SessionId)
}

func (p *LoginSuccessClientboundPacket) Encode(minecraftStream *streams.MinecraftStream) error {
	err := minecraftStream.WriteUuid(p.Profile.Uuid)
	if err != nil {
		return err
	}

	err = minecraftStream.WriteString(p.Profile.Username)
	if err != nil {
		return err
	}

	err = minecraftStream.WriteVarInt(int32(len(p.Profile.Properties)))
	if err != nil {
		return err
	}

	for _, property := range p.Profile.Properties {
		err = minecraftStream.WriteString(property.Name)
		if err != nil {
			return err
		}

		err = minecraftStream.WriteString(property.Value)
		if err != nil {
			return err
		}

		err = minecraftStream.WriteBoolean(property.Signature != nil)
		if err != nil {
			return err
		}

		if property.Signature != nil {
			err = minecraftStream.WriteString(*property.Signature)
			if err != nil {
				return err
			}
		}
	}

	return minecraftStream.WriteUuid(p.SessionId)
}
