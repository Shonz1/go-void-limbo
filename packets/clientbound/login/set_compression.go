package login

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// SetCompressionClientboundPacket tells the client the body size at or above
// which packets are deflated from here on.
//
// It changes how every later packet is framed, in both directions: each one
// gains a var int in front of its body carrying the size that body inflates to,
// or zero for a body left in full because it was under the threshold. This
// packet is itself the last one framed the plain way, which is why the client
// has to be told before anything else is sent.
type SetCompressionClientboundPacket struct {
	Threshold int32
}

func (p *SetCompressionClientboundPacket) String() string {
	return fmt.Sprintf("SetCompressionClientboundPacket{Threshold:%d}", p.Threshold)
}

func (p *SetCompressionClientboundPacket) Encode(minecraftStream *streams.MinecraftStream) error {
	return minecraftStream.WriteVarInt(p.Threshold)
}
