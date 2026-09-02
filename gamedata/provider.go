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
}

// Provider resolves the configuration-phase registry packets for a client's
// protocol version.
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
		packets := make([]types.ClientboundPacket, 0, len(set.Registries)+1)

		// The shape is the set's version's own. A set that starts below 1.20.5
		// is read by clients that take every registry in one packet, which is
		// the one difference in this package's output between the versions:
		// the content of a set is what varies, and the shape only once.
		if set.MinProtocol < combinedRegistryDataProtocol {
			body, err := encodeCombined(set.Registries)
			if err != nil {
				return nil, fmt.Errorf("gamedata: protocol %d: %w", set.MinProtocol, err)
			}

			packets = append(packets, configuration.NewRegistryDataClientboundPacket(combinedRegistryName(set.Registries), body))
		} else {
			for _, registry := range set.Registries {
				body, err := registry.encode()
				if err != nil {
					return nil, fmt.Errorf("gamedata: protocol %d: %w", set.MinProtocol, err)
				}

				packets = append(packets, configuration.NewRegistryDataClientboundPacket(registry.Name, body))
			}
		}

		// Tags go out after the registries they point into, since a tag names
		// its entries by registry id.
		if len(set.Tags) > 0 {
			body, err := encodeTags(set.Tags)
			if err != nil {
				return nil, fmt.Errorf("gamedata: protocol %d: %w", set.MinProtocol, err)
			}

			packets = append(packets, configuration.NewUpdateTagsClientboundPacket(len(set.Tags), countTags(set.Tags), body))
		}

		buckets = append(buckets, bucket{minProtocol: set.MinProtocol, packets: packets})
	}

	sort.Slice(buckets, func(i, j int) bool { return buckets[i].minProtocol < buckets[j].minProtocol })

	for i := 1; i < len(buckets); i++ {
		if buckets[i].minProtocol == buckets[i-1].minProtocol {
			return nil, fmt.Errorf("gamedata: two sets both start at protocol %d", buckets[i].minProtocol)
		}
	}

	return &Provider{buckets: buckets}, nil
}

// PacketsFor returns the packets to send to a client on version, which is the
// content of the newest set that version reaches. A version older than every
// set gets nil: the configuration phase itself only exists from 1.20.2, so
// there is no sensible fallback to offer a client from before it.
//
// The returned slice and the packets in it are shared across connections and
// must not be modified.
func (p *Provider) PacketsFor(version types.ProtocolVersion) []types.ClientboundPacket {
	for i := len(p.buckets) - 1; i >= 0; i-- {
		if version.ID >= p.buckets[i].minProtocol {
			return p.buckets[i].packets
		}
	}

	return nil
}
