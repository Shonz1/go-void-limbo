package gamedata

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/configuration"
	"github.com/Shonz1/go-void-limbo/types"
	"sort"
)

// Set is the registry content for a range of protocol versions. It applies from
// MinProtocol up to whichever set starts next, so a version that changes nothing
// about registries needs no set of its own.
type Set struct {
	MinProtocol types.ProtocolId
	Registries  []Registry
	Tags        []TagSet
}

// bucket is a Set with its packets already encoded.
type bucket struct {
	minProtocol types.ProtocolId
	packets     []types.ClientboundPacket

	// registryCodec is the registries as a version before 1.20.2 reads them,
	// out of its play login rather than out of a packet: see
	// encodeRegistryCodec. Nil for a version that is sent them as packets.
	registryCodec []byte

	// dimensionType is the dimension type the play login puts the player
	// into as a version before 1.19 reads it, spelled out in the login
	// itself: see encodeDimensionType. Nil for a version that reads a name
	// there.
	dimensionType []byte
}

// Provider resolves the registry packets for a client's protocol version, and
// for a version from before the configuration phase the compound its play
// login carries the registries in instead.
//
// Bodies are encoded once, when the provider is built, and the resulting
// packets are shared by every connection on that version. Registry data is the
// largest thing a limbo sends, and it is identical for all of them, so encoding
// it per connection would be the server's main cost for no gain.
type Provider struct {
	buckets []bucket
}

// NewProvider encodes sets into a provider. Sets may be given in any order.
func NewProvider(sets ...Set) (*Provider, error) {
	buckets := make([]bucket, 0, len(sets))

	for _, set := range sets {
		encoded, err := encodeSet(set)
		if err != nil {
			return nil, err
		}

		buckets = append(buckets, encoded)
	}

	return newProvider(buckets)
}

// encodeSet encodes one set's packets, after which the set itself is no
// longer needed.
func encodeSet(set Set) (bucket, error) {
	packets := make([]types.ClientboundPacket, 0, len(set.Registries)+1)

	var registryCodec, dimensionType []byte

	// The shape is the set's version's own. A set that starts below 1.20.5
	// is read by clients that take every registry in one packet, one that
	// starts below 1.20.2 by clients that take them inside the play login,
	// with no packet at all, and one that starts below 1.19 by clients that
	// read the dimension type they are put into out of that login as well,
	// spelled out rather than named. Those are the three differences in
	// this package's output between the versions: the content of a set is
	// what varies, and the shape only at those three steps.
	if set.MinProtocol < registryCodecProtocol {
		codec, err := encodeRegistryCodec(set.Registries)
		if err != nil {
			return bucket{}, fmt.Errorf("gamedata: protocol %d: %w", set.MinProtocol, err)
		}

		registryCodec = codec

		if set.MinProtocol < inlineDimensionTypeProtocol {
			encoded, err := encodeDimensionType(set.Registries)
			if err != nil {
				return bucket{}, fmt.Errorf("gamedata: protocol %d: %w", set.MinProtocol, err)
			}

			dimensionType = encoded
		}
	} else if set.MinProtocol < combinedRegistryDataProtocol {
		body, err := encodeCombined(set.Registries)
		if err != nil {
			return bucket{}, fmt.Errorf("gamedata: protocol %d: %w", set.MinProtocol, err)
		}

		packets = append(packets, configuration.NewRegistryDataClientboundPacket(combinedRegistryName(set.Registries), body))
	} else {
		for _, registry := range set.Registries {
			body, err := registry.encode()
			if err != nil {
				return bucket{}, fmt.Errorf("gamedata: protocol %d: %w", set.MinProtocol, err)
			}

			packets = append(packets, configuration.NewRegistryDataClientboundPacket(registry.Name, body))
		}
	}

	// Tags go out after the registries they point into, since a tag names
	// its entries by registry id.
	if len(set.Tags) > 0 {
		body, err := encodeTags(set.Tags)
		if err != nil {
			return bucket{}, fmt.Errorf("gamedata: protocol %d: %w", set.MinProtocol, err)
		}

		packets = append(packets, configuration.NewUpdateTagsClientboundPacket(len(set.Tags), countTags(set.Tags), body))
	}

	return bucket{minProtocol: set.MinProtocol, packets: packets, registryCodec: registryCodec, dimensionType: dimensionType}, nil
}

// newProvider orders encoded buckets by version and refuses two that start at
// the same one.
func newProvider(buckets []bucket) (*Provider, error) {
	sort.Slice(buckets, func(i, j int) bool { return buckets[i].minProtocol < buckets[j].minProtocol })

	for i := 1; i < len(buckets); i++ {
		if buckets[i].minProtocol == buckets[i-1].minProtocol {
			return nil, fmt.Errorf("gamedata: two sets both start at protocol %d", buckets[i].minProtocol)
		}
	}

	return &Provider{buckets: buckets}, nil
}

// PacketsFor returns the packets to send to a client on version, which is the
// content of the newest set that version reaches. They are written in the
// configuration phase from 1.20.2 on, and for a version from before it -- one
// with no such phase -- in the play phase right after the login, which is
// where such a version reads the tags, the only packet it is sent here. A
// version older than every set gets nil.
//
// The returned slice and the packets in it are shared across connections and
// must not be modified.
func (p *Provider) PacketsFor(version types.ProtocolVersion) []types.ClientboundPacket {
	if b := p.bucketFor(version); b != nil {
		return b.packets
	}

	return nil
}

// RegistryCodecFor returns the registries as a client on version reads them
// out of its play login: the one compound every version before 1.20.2
// carries there, named as a root. It is nil for a version from 1.20.2 on,
// which is sent its registries as packets and reads nothing of the kind in
// its login. The bytes are shared across connections and must not be
// modified.
func (p *Provider) RegistryCodecFor(version types.ProtocolVersion) []byte {
	if b := p.bucketFor(version); b != nil {
		return b.registryCodec
	}

	return nil
}

// DimensionTypeFor returns the dimension type a client on version is put
// into as it reads it out of its play login: the entry itself, named as a
// root, which is how every version before 1.19 reads the field. A version
// from 1.19 on reads a name there, one of the entries the registries it
// carries hold, and gets nil. The bytes are shared across connections and
// must not be modified.
func (p *Provider) DimensionTypeFor(version types.ProtocolVersion) []byte {
	if b := p.bucketFor(version); b != nil {
		return b.dimensionType
	}

	return nil
}

// bucketFor finds the newest set version reaches, or nil for a version older
// than every set.
func (p *Provider) bucketFor(version types.ProtocolVersion) *bucket {
	for i := len(p.buckets) - 1; i >= 0; i-- {
		if version.ID >= p.buckets[i].minProtocol {
			return &p.buckets[i]
		}
	}

	return nil
}
