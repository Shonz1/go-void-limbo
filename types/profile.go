package types

import (
	"fmt"
	"strings"
)

// ProfileProperty is a key/value pair attached to a game profile. Skins and
// capes arrive as the "textures" property, whose value is a base64 blob that
// Mojang signs; Signature is nil for properties that carry no signature.
type ProfileProperty struct {
	Name      string
	Value     string
	Signature *string
}

func (p ProfileProperty) String() string {
	if p.Signature == nil {
		return fmt.Sprintf("ProfileProperty{Name:%s Value:%s}", p.Name, p.Value)
	}

	return fmt.Sprintf("ProfileProperty{Name:%s Value:%s Signature:%s}", p.Name, p.Value, *p.Signature)
}

// GameProfile identifies a player: the account uuid, the username and the
// properties describing their appearance.
type GameProfile struct {
	Uuid       string
	Username   string
	Properties []ProfileProperty
}

func (p GameProfile) String() string {
	properties := make([]string, 0, len(p.Properties))
	for _, property := range p.Properties {
		properties = append(properties, property.String())
	}

	return fmt.Sprintf("GameProfile{Uuid:%s Username:%s Properties:[%s]}", p.Uuid, p.Username, strings.Join(properties, " "))
}
