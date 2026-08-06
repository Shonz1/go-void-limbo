// Package gamedata holds the Minecraft registries a server sends to a client
// during the configuration phase, and resolves which content a given protocol
// version should be sent.
//
// It is deliberately separate from package registries, which maps packet ids to
// decoders and handlers. The two are unrelated despite both being called
// registries: this one is game content, that one is protocol wiring.
package gamedata

import (
	"bytes"
	"fmt"
	"go-void-limbo/nbt"
	"go-void-limbo/streams"
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
