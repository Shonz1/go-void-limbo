package login

import (
	"fmt"
	"go-void-limbo/streams"
)

// LoginPluginRequestClientboundPacket asks whoever is on the connection a
// question the protocol has no packet for, named by a channel both ends have to
// know already. A limbo asks exactly one: the channel a proxy forwards a login
// on.
//
// A client that has never heard of the channel says so, which is the answer that
// tells a proxy from a player.
type LoginPluginRequestClientboundPacket struct {
	// MessageId is what the answer comes back under, since nothing else about a
	// response says which request it answers.
	MessageId int32

	Channel string

	// Data is whatever the channel calls for, and is read to the end of the
	// packet rather than behind a length, so it can only be the last field.
	Data []byte
}

func (p *LoginPluginRequestClientboundPacket) String() string {
	return fmt.Sprintf("LoginPluginRequestClientboundPacket{MessageId:%d Channel:%s Data:%d bytes}", p.MessageId, p.Channel, len(p.Data))
}

func (p *LoginPluginRequestClientboundPacket) Encode(minecraftStream *streams.MinecraftStream) error {
	err := minecraftStream.WriteVarInt(p.MessageId)
	if err != nil {
		return err
	}

	err = minecraftStream.WriteString(p.Channel)
	if err != nil {
		return err
	}

	return minecraftStream.WriteBytes(p.Data)
}
