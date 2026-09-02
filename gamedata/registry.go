// Package gamedata holds the Minecraft registries a server sends to a client
// during the configuration phase, and resolves which content a given protocol
// version should be sent.
//
// It is deliberately separate from package protocol, which maps packet ids to
// decoders and handlers. The two are unrelated despite both involving things
// Minecraft calls registries: this one is game content, that one is protocol
// wiring.
package gamedata

import (
	"bytes"
	"fmt"
	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"strings"
)

// Entry is one element of a registry.
//
// Data may be nil, which encodes the entry as a name with no definition and
// leaves the client to use the copy it already holds. That is only safe for
// entries covered by a pack the client reported in the known packs exchange, so
// a server that never sends Clientbound Known Packs must give every entry data.
type Entry struct {
	Name string
	Data nbt.Tag
}

// Registry is a named set of entries.
//
// Order is significant. An entry's index is the id everything later refers to
// it by, so the dimension field of the play login packet and the biome palette
// of a chunk are both numbers pointing into this slice. Reordering entries
// between versions silently changes what those numbers mean.
type Registry struct {
	Name    string
	Entries []Entry
}

// combinedRegistryDataProtocol is the first version to read one registry data
// packet per registry. Every version below it reads a single packet holding
// every registry at once, in the shape encodeCombined writes.
var combinedRegistryDataProtocol = types.ProtocolVersions.MINECRAFT_1_20_5.ID

// encodeCombined writes the one packet body a client before 1.20.5 reads every
// registry from: a compound keyed by registry name, holding for each registry
// its name again under "type" and under "value" a list of its entries, each
// an entry's name, its id and its definition under "element".
//
// The id is explicit here where the per-registry packets leave it to the
// entry's position, and it is written as that position, so the two shapes
// number every entry alike and the play phase can name a dimension or a biome
// by the same index on either side of 1.20.5. A definition is not optional in
// this shape: nothing before 1.20.5 knows a pack to fall back on, so an entry
// with no data is refused rather than sent as a name the client cannot fill.
func encodeCombined(registries []Registry) ([]byte, error) {
	compound := nbt.Compound{}

	for _, registry := range registries {
		entries := make([]nbt.Tag, 0, len(registry.Entries))

		for id, entry := range registry.Entries {
			if entry.Data == nil || entry.Data.Type() == nbt.TagEnd {
				return nil, fmt.Errorf("registry %s: entry %s has no definition, which the combined shape cannot leave out", registry.Name, entry.Name)
			}

			entries = append(entries, nbt.Compound{
				"name":    nbt.String(entry.Name),
				"id":      nbt.Int(int32(id)),
				"element": entry.Data,
			})
		}

		compound[registry.Name] = nbt.Compound{
			"type":  nbt.String(registry.Name),
			"value": nbt.List{ElementType: nbt.TagCompound, Elements: entries},
		}
	}

	buf := new(bytes.Buffer)
	ms := streams.NewMinecraftStreamFromBuffer(buf)

	if err := nbt.Write(ms, compound); err != nil {
		return nil, err
	}

	if err := ms.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// combinedRegistryName is what the one packet holding every registry is
// called where a packet is named for logging: every registry in it, in the
// order the set lists them.
func combinedRegistryName(registries []Registry) string {
	names := make([]string, 0, len(registries))
	for _, registry := range registries {
		names = append(names, registry.Name)
	}

	return strings.Join(names, ",")
}

// encode writes the packet body: the registry name, the entry count, then each
// entry's name, a flag for whether a definition follows, and the definition.
func (r Registry) encode() ([]byte, error) {
	buf := new(bytes.Buffer)
	ms := streams.NewMinecraftStreamFromBuffer(buf)

	if err := ms.WriteString(r.Name); err != nil {
		return nil, err
	}

	if err := ms.WriteVarInt(int32(len(r.Entries))); err != nil {
		return nil, err
	}

	for _, entry := range r.Entries {
		if err := ms.WriteString(entry.Name); err != nil {
			return nil, fmt.Errorf("registry %s: entry %s: %w", r.Name, entry.Name, err)
		}

		// An End tag is how NBT spells an absent value, so it means the same
		// thing here as a nil tag: send the name and nothing else.
		hasData := entry.Data != nil && entry.Data.Type() != nbt.TagEnd

		if err := ms.WriteBoolean(hasData); err != nil {
			return nil, fmt.Errorf("registry %s: entry %s: %w", r.Name, entry.Name, err)
		}

		if !hasData {
			continue
		}

		if err := nbt.Write(ms, entry.Data); err != nil {
			return nil, fmt.Errorf("registry %s: entry %s: %w", r.Name, entry.Name, err)
		}
	}

	if err := ms.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
