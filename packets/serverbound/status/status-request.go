// Package status holds the serverbound packets of the status phase, which is
// the whole of a server list ping: a client asks what the server says about
// itself, times the answer, and closes the connection. Nothing here logs a
// player in, and nothing said here is remembered afterwards.
package status

import (
	"go-void-limbo/streams"
	"go-void-limbo/types"
)

// StatusRequestServerboundPacket asks the server to describe itself. It carries
// no fields: the handshake in front of it already said which version is asking,
// and there is nothing else a ping needs to say.
type StatusRequestServerboundPacket struct{}

func (p *StatusRequestServerboundPacket) String() string {
	return "StatusRequestServerboundPacket{}"
}

func DecodeStatusRequestServerboundPacket(_ *streams.MinecraftStream) (types.ServerboundPacket, error) {
	return &StatusRequestServerboundPacket{}, nil
}
