package types

import (
	"errors"
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

type PacketId = int32

type ServerboundPacket interface {
	String() string
}

type ClientboundPacket interface {
	String() string
	Encode(ms *streams.MinecraftStream) error
}

// PreparedPacket is a clientbound packet already in wire form for one phase
// and one protocol version: its id and its body, carried down to that version
// and deflated, the way it would sit inside a frame on a connection that has
// been told a compression threshold. It is for the packets a server sends to
// every connection on a version unchanged -- the chunks of a world -- which
// are encoded, downgraded and deflated once rather than on every join, and
// held deflated, which is a fraction of the size.
//
// A connection writes it as it stands when the threshold it was told is one
// the body would be deflated at, and inflates it otherwise, so a prepared
// packet is correct on any connection of its version; only the cost moves.
type PreparedPacket struct {
	Phase   Phase
	Version ProtocolId

	// Name says what the packet was, for the log.
	Name string

	// Size is how long the body is inflated, which is what the frame carries
	// in front of the deflated bytes and what inflating it is bounded by.
	Size int32

	// Deflated is the id and the body, zlib deflated.
	Deflated []byte
}

// PrepareClientbound deflates a body -- a packet id in front of a payload
// already at version -- into a prepared packet for phase and version.
func PrepareClientbound(phase Phase, version ProtocolVersion, name string, body []byte) (*PreparedPacket, error) {
	deflated, err := streams.Compress(body)
	if err != nil {
		return nil, err
	}

	// Deflating grows a buffer as it goes; what is held is a copy cut to what
	// came out of it, since the packet is held for the life of the process.
	held := make([]byte, len(deflated))
	copy(held, deflated)

	return &PreparedPacket{
		Phase:    phase,
		Version:  version.ID,
		Name:     name,
		Size:     int32(len(body)),
		Deflated: held,
	}, nil
}

// Body inflates the packet back to the id and body it was prepared from.
func (p *PreparedPacket) Body() ([]byte, error) {
	return streams.Decompress(p.Deflated, p.Size)
}

func (p *PreparedPacket) String() string {
	return fmt.Sprintf("%s{prepared for protocol %d, %d bytes}", p.Name, p.Version, p.Size)
}

// Encode is what makes a prepared packet a ClientboundPacket, so it can sit
// in a list with the packets that are not, and it always fails: the id is
// part of what was prepared, and a connection writes the packet whole rather
// than encoding it.
func (p *PreparedPacket) Encode(*streams.MinecraftStream) error {
	return errors.New("a prepared packet is written whole, not encoded")
}
