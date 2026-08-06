package configuration

import (
	"fmt"
	"go-void-limbo/streams"
)

// UpdateTagsClientboundPacket tells the client which tags exist and what is in
// them, for every registry at once.
//
// Like registry data the body is pre-encoded, for the same reason: it is the
// same bytes for every client on a protocol version. See package gamedata.
type UpdateTagsClientboundPacket struct {
	registries int
	tags       int
	body       []byte
}

// NewUpdateTagsClientboundPacket wraps an already encoded body. The counts are
// carried for logging only; what the client reads is inside body.
func NewUpdateTagsClientboundPacket(registries int, tags int, body []byte) *UpdateTagsClientboundPacket {
	return &UpdateTagsClientboundPacket{registries: registries, tags: tags, body: body}
}

func (p *UpdateTagsClientboundPacket) String() string {
	return fmt.Sprintf("UpdateTagsClientboundPacket{Registries: %d, Tags: %d, Body: %d bytes}",
		p.registries, p.tags, len(p.body))
}

func (p *UpdateTagsClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	return ms.WriteBytes(p.body)
}
