package configuration

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// RegistryDataClientboundPacket tells the client about one registry and the
// entries in it.
//
// The body is held pre-encoded rather than as fields, because every client on a
// given protocol version is sent byte-identical registries: package gamedata
// encodes each one once at startup and hands the bytes here, so a connection
// costs a write and nothing else. It is also what lets one packet type cover
// both wire shapes, the per-registry packets used from 1.20.5 onwards and the
// single combined codec sent before that.
type RegistryDataClientboundPacket struct {
	registryName string
	body         []byte
}

// NewRegistryDataClientboundPacket wraps an already encoded body. registryName
// is carried for logging only; the name the client reads is inside body.
func NewRegistryDataClientboundPacket(registryName string, body []byte) *RegistryDataClientboundPacket {
	return &RegistryDataClientboundPacket{registryName: registryName, body: body}
}

func (p *RegistryDataClientboundPacket) RegistryName() string {
	return p.registryName
}

func (p *RegistryDataClientboundPacket) String() string {
	return fmt.Sprintf("RegistryDataClientboundPacket{RegistryName: %s, Body: %d bytes}", p.registryName, len(p.body))
}

func (p *RegistryDataClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	return ms.WriteBytes(p.body)
}
