package login

import (
	"fmt"
	"go-void-limbo/streams"
	"go-void-limbo/types"
)

// LoginPluginResponseServerboundPacket answers a login plugin request, under the
// message id that request went out with.
//
// Successful is the answer to whether the channel meant anything on the other
// end. A vanilla client knows no channel at all and says so, with nothing behind
// it; a proxy forwarding a login says it does and puts the login in Data.
type LoginPluginResponseServerboundPacket struct {
	MessageId  int32
	Successful bool

	// Data runs to the end of the packet, and is empty for an answer that
	// carried nothing.
	Data []byte
}

// String reports the payload by length. What is inside it is a login and a
// signature over it, and neither is worth a line in the log.
func (p *LoginPluginResponseServerboundPacket) String() string {
	return fmt.Sprintf("LoginPluginResponseServerboundPacket{MessageId:%d Successful:%t Data:%d bytes}", p.MessageId, p.Successful, len(p.Data))
}

func DecodeLoginPluginResponseServerboundPacket(minecraftStream *streams.MinecraftStream) (types.ServerboundPacket, error) {
	messageId, err := minecraftStream.ReadVarInt()
	if err != nil {
		return nil, err
	}

	successful, err := minecraftStream.ReadBoolean()
	if err != nil {
		return nil, err
	}

	if !successful {
		return &LoginPluginResponseServerboundPacket{MessageId: messageId, Successful: false}, nil
	}

	// The body was read off the connection in full before it got here, so
	// reading to the end of it reads this packet and nothing after it.
	data, err := minecraftStream.ReadRest()
	if err != nil {
		return nil, err
	}

	return &LoginPluginResponseServerboundPacket{MessageId: messageId, Successful: true, Data: data}, nil
}
