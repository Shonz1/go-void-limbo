package protocol

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"reflect"
)

type PacketDecoder = func(ms *streams.MinecraftStream) (types.ServerboundPacket, error)

// Transformer rewrites one packet's body from the version it currently speaks
// into the one version along, reading the body it is given and writing the body
// that replaces it.
//
// A transformer sees the body alone, without the packet id in front of it: the
// id is a property of the version the packet is framed at rather than of the
// body, and it is resolved from the packet's type at whichever end the body
// ends up. So a version that only renumbered a packet needs no transformer, and
// one that changed its fields writes out the fields it changed and copies the
// rest.
type Transformer = func(in *streams.MinecraftStream, out *streams.MinecraftStream) error

// ServerboundEntry ties a packet's wire decoder to the handler that reacts to
// it. Both belong to the latest protocol version, which is the only version any
// packet is implemented at.
type ServerboundEntry struct {
	Decoder PacketDecoder

	// Handler is nil for a packet the server decodes but has nothing to do
	// about.
	Handler types.PacketHandler
}

// serverboundIdKey resolves what arrived on the wire to the packet it is. The
// protocol version is part of it because that is the whole point of the table:
// the same packet is a different id on different versions.
type serverboundIdKey struct {
	Phase      types.Phase
	ProtocolID types.ProtocolId
	PacketID   types.PacketId
}

// serverboundImplKey finds the one implementation a packet has, which no
// protocol version is part of.
type serverboundImplKey struct {
	Phase      types.Phase
	PacketType reflect.Type
}

type clientboundIdKey struct {
	Phase      types.Phase
	PacketType reflect.Type
	ProtocolID types.ProtocolId
}

// transformKey names one step of the chain: the packet being carried, and the
// version it is being carried away from.
type transformKey struct {
	Phase      types.Phase
	PacketType reflect.Type
	ProtocolID types.ProtocolId
}

type Registry struct {
	serverboundIds     map[serverboundIdKey]reflect.Type
	serverboundPackets map[serverboundImplKey]ServerboundEntry
	clientboundIds     map[clientboundIdKey]types.PacketId

	// upgrades carry a serverbound body from the version keyed up to the next
	// one, and downgrades carry a clientbound body from the version keyed down
	// to the previous one. They are separate tables because the two directions
	// of one version's change are not each other's inverse in general, and a
	// packet only ever travels one of them.
	upgrades   map[transformKey]Transformer
	downgrades map[transformKey]Transformer
}

func NewRegistry() *Registry {
	return &Registry{
		serverboundIds:     make(map[serverboundIdKey]reflect.Type),
		serverboundPackets: make(map[serverboundImplKey]ServerboundEntry),
		clientboundIds:     make(map[clientboundIdKey]types.PacketId),
		upgrades:           make(map[transformKey]Transformer),
		downgrades:         make(map[transformKey]Transformer),
	}
}

// RegisterServerbound records how a packet is decoded and what reacts to it.
// This is the implementation, and there is one of it however many versions
// speak the packet, because everything is decoded at the latest version.
func (r *Registry) RegisterServerbound(phase types.Phase, packet reflect.Type, decoder PacketDecoder, handler types.PacketHandler) {
	r.serverboundPackets[serverboundImplKey{Phase: phase, PacketType: packet}] = ServerboundEntry{Decoder: decoder, Handler: handler}
}

// RegisterServerboundId records the id one version gives a packet. A version
// needs nothing else to speak a packet it did not change.
func (r *Registry) RegisterServerboundId(phase types.Phase, protocolVersion types.ProtocolVersion, packet reflect.Type, packetId types.PacketId) {
	r.serverboundIds[serverboundIdKey{Phase: phase, ProtocolID: protocolVersion.ID, PacketID: packetId}] = packet
}

// RegisterClientboundId records the id one version gives a packet.
func (r *Registry) RegisterClientboundId(phase types.Phase, protocolVersion types.ProtocolVersion, packet reflect.Type, packetId types.PacketId) {
	r.clientboundIds[clientboundIdKey{Phase: phase, PacketType: packet, ProtocolID: protocolVersion.ID}] = packetId
}

// RegisterUpgrade records how a serverbound packet's body is carried from
// protocolVersion up to the version above it.
func (r *Registry) RegisterUpgrade(phase types.Phase, protocolVersion types.ProtocolVersion, packet reflect.Type, transformer Transformer) {
	r.upgrades[transformKey{Phase: phase, PacketType: packet, ProtocolID: protocolVersion.ID}] = transformer
}

// RegisterDowngrade records how a clientbound packet's body is carried from
// protocolVersion down to the version below it.
func (r *Registry) RegisterDowngrade(phase types.Phase, protocolVersion types.ProtocolVersion, packet reflect.Type, transformer Transformer) {
	r.downgrades[transformKey{Phase: phase, PacketType: packet, ProtocolID: protocolVersion.ID}] = transformer
}

// GetServerboundType says which packet an id means on the version it arrived
// on.
func (r *Registry) GetServerboundType(phase types.Phase, protocolVersion types.ProtocolVersion, packetId types.PacketId) (reflect.Type, bool) {
	packet, ok := r.serverboundIds[serverboundIdKey{Phase: phase, ProtocolID: protocolVersion.ID, PacketID: packetId}]
	return packet, ok
}

// GetServerbound returns how a packet is decoded and what reacts to it.
func (r *Registry) GetServerbound(phase types.Phase, packet reflect.Type) (ServerboundEntry, bool) {
	entry, ok := r.serverboundPackets[serverboundImplKey{Phase: phase, PacketType: packet}]
	return entry, ok
}

// GetClientboundId returns the id a packet goes out under on a version, or -1
// when that version does not carry it.
func (r *Registry) GetClientboundId(phase types.Phase, packet reflect.Type, protocolVersion types.ProtocolVersion) types.PacketId {
	id, ok := r.clientboundIds[clientboundIdKey{Phase: phase, PacketType: packet, ProtocolID: protocolVersion.ID}]
	if !ok {
		return -1
	}

	return id
}

// EncodeClientbound puts a packet in the form it goes out in on a version: its
// id at that version in front of its body, encoded at the latest version --
// the only one a packet knows how to be -- and carried down from there. It
// reports an error for a packet the version does not carry.
func (r *Registry) EncodeClientbound(phase types.Phase, protocolVersion types.ProtocolVersion, packet types.ClientboundPacket) ([]byte, error) {
	if packet == nil {
		return nil, errors.New("packet is nil")
	}

	packetType := reflect.TypeOf(packet).Elem()

	packetId := r.GetClientboundId(phase, packetType, protocolVersion)
	if packetId == -1 {
		return nil, errors.New("unknown packet id")
	}

	payloadBuf := new(bytes.Buffer)
	payloadStream := streams.NewMinecraftStreamFromBuffer(payloadBuf)

	if err := packet.Encode(payloadStream); err != nil {
		return nil, err
	}

	if err := payloadStream.Flush(); err != nil {
		return nil, err
	}

	payload, err := r.DowngradeBody(phase, packetType, protocolVersion, payloadBuf.Bytes())
	if err != nil {
		return nil, err
	}

	body := streams.AppendVarInt(make([]byte, 0, 5+len(payload)), packetId)

	return append(body, payload...), nil
}

// UpgradeBody carries a body that arrived on from up to the latest version, one
// version at a time, so that what comes back is what the latest version's
// decoder expects.
//
// A version that changed nothing about the packet has no transformer registered
// and passes the body through untouched, which is the usual case. A version
// that is not on the chain at all is left alone entirely: the handshake is read
// before a connection has said what it speaks, and there is nothing to carry it
// up from.
func (r *Registry) UpgradeBody(phase types.Phase, packet reflect.Type, from types.ProtocolVersion, body []byte) ([]byte, error) {
	if !types.IsSupportedProtocolVersion(from) {
		return body, nil
	}

	var err error

	for version := from; version.ID != types.LatestProtocolVersion.ID; {
		next, ok := types.NextProtocolVersion(version)
		if !ok {
			return nil, fmt.Errorf("no version above protocol %d to upgrade to", version.ID)
		}

		if transformer, ok := r.upgrades[transformKey{Phase: phase, PacketType: packet, ProtocolID: version.ID}]; ok {
			body, err = transform(transformer, body)
			if err != nil {
				return nil, fmt.Errorf("failed to upgrade %s from protocol %d: %w", packet, version.ID, err)
			}
		}

		version = next
	}

	return body, nil
}

// DowngradeBody carries a body encoded at the latest version down to the one
// the client speaks, one version at a time.
func (r *Registry) DowngradeBody(phase types.Phase, packet reflect.Type, to types.ProtocolVersion, body []byte) ([]byte, error) {
	if !types.IsSupportedProtocolVersion(to) {
		return body, nil
	}

	var err error

	for version := types.LatestProtocolVersion; version.ID != to.ID; {
		previous, ok := types.PreviousProtocolVersion(version)
		if !ok {
			return nil, fmt.Errorf("no version below protocol %d to downgrade to", version.ID)
		}

		if transformer, ok := r.downgrades[transformKey{Phase: phase, PacketType: packet, ProtocolID: version.ID}]; ok {
			body, err = transform(transformer, body)
			if err != nil {
				return nil, fmt.Errorf("failed to downgrade %s from protocol %d: %w", packet, version.ID, err)
			}
		}

		version = previous
	}

	return body, nil
}

// transform runs one transformer over a body, giving it the body to read and a
// buffer of its own to write the replacement into.
func transform(transformer Transformer, body []byte) ([]byte, error) {
	in := streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(body))

	buf := new(bytes.Buffer)
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := transformer(in, out); err != nil {
		return nil, err
	}

	if err := out.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
