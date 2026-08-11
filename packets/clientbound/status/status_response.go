// Package status holds the clientbound packets of the status phase: what a
// server list ping is answered with, which is a description of the server and
// the number the client asked to have sent back.
package status

import (
	"encoding/json"
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// StatusResponseClientboundPacket describes the server to a client that has not
// connected to it yet.
//
// The whole packet is one string, and the string is a JSON document, so this is
// the one packet whose shape is not the fields it is written from. A field the
// client cannot read is a field it leaves at whatever it defaults to, which is
// why the document is built from a type rather than assembled by hand.
type StatusResponseClientboundPacket struct {
	Status types.ServerStatus
}

func (p *StatusResponseClientboundPacket) String() string {
	return fmt.Sprintf("StatusResponseClientboundPacket{Version:%s Protocol:%d Players:%d/%d Description:%s}",
		p.Status.Version.Name, p.Status.Version.Protocol, p.Status.Players.Online, p.Status.Players.Max, p.Status.Description.Text)
}

func (p *StatusResponseClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	document, err := json.Marshal(p.Status)
	if err != nil {
		return fmt.Errorf("failed to encode the server status: %w", err)
	}

	return ms.WriteString(string(document))
}
