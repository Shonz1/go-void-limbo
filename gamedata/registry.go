// Package gamedata holds the Minecraft registries a server sends to a client
// during the configuration phase -- or, for a client from before there was
// one, inside its play login -- and resolves which content a given protocol
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
// every registry at once, in the shape encodeCombined writes -- or, below
// registryCodecProtocol, no packet at all.
var combinedRegistryDataProtocol = types.ProtocolVersions.MINECRAFT_1_20_5.ID

// registryCodecProtocol is the first version to read the registries out of
// packets of their own, in the configuration phase. Every version below it
// has no such phase and reads them out of its play login instead, as the
// same one compound named as a root: see encodeRegistryCodec.
var registryCodecProtocol = types.ProtocolVersions.MINECRAFT_1_20_2.ID

// inlineDimensionTypeProtocol is the first version to read the dimension
// type its play login puts it into as a name, one of the entries the
// registries the login carries hold. Every version below it reads the entry
// itself there, spelled out in full a second time: see encodeDimensionType.
var inlineDimensionTypeProtocol = types.ProtocolVersions.MINECRAFT_1_19.ID

// namedTagRegistriesProtocol is the first version to read the tags with the
// registry each run belongs to named in front of it. A version before it
// reads four runs it knows by their position: see encodeTags1_16_4.
var namedTagRegistriesProtocol = types.ProtocolVersions.MINECRAFT_1_17.ID

// dimensionListProtocol is the first version to read the combined compound
// out of its play login, the biomes and every registry after them included,
// and the dimension type it is put into spelled out behind it. A version
// before it -- 1.16.1 or 1.16 -- knows its biomes for itself, reads a list of the
// dimension types alone and is put into one of them by name: see
// encodeDimensionList and encodeDimensionTypeName.
var dimensionListProtocol = types.ProtocolVersions.MINECRAFT_1_16_2.ID

// loginRegistriesProtocol is the first version to read anything of a
// registry out of its play login. A version before it -- 1.15.2, 1.15.1, 1.15, 1.14.4, 1.14.3, 1.14.2, 1.14.1 or 1.14 -- knows its
// dimensions for itself as it knows its biomes, is put into one by number,
// and is sent the tags alone: a set for it holds no registry.
var loginRegistriesProtocol = types.ProtocolVersions.MINECRAFT_1_16.ID

// combinedCompound is the one compound a client before 1.20.5 reads every
// registry from: keyed by registry name, holding for each registry its name
// again under "type" and under "value" a list of its entries, each an entry's
// name, its id and its definition under "element".
//
// The id is explicit here where the per-registry packets leave it to the
// entry's position, and it is written as that position, so the two shapes
// number every entry alike and the play phase can name a dimension or a biome
// by the same index on either side of 1.20.5. A definition is not optional in
// this shape: nothing before 1.20.5 knows a pack to fall back on, so an entry
// with no data is refused rather than sent as a name the client cannot fill.
func combinedCompound(registries []Registry) (nbt.Compound, error) {
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

	return compound, nil
}

// encodeCombined writes the one packet body a client from 1.20.2 up to 1.20.5
// reads every registry from: the combined compound as a nameless root, the
// network NBT 1.20.2 introduced.
func encodeCombined(registries []Registry) ([]byte, error) {
	compound, err := combinedCompound(registries)
	if err != nil {
		return nil, err
	}

	return encodeNbt(func(ms *streams.MinecraftStream) error { return nbt.Write(ms, compound) })
}

// encodeRegistryCodec writes the registries as a client before 1.20.2 reads
// them out of its play login: the same combined compound, as a root named
// with the empty string, the way every NBT on the wire was named before
// 1.20.2. It is not a packet body but a field of one, which the 1.20.2
// step's login transformer writes into the packet where that version reads
// it.
func encodeRegistryCodec(registries []Registry) ([]byte, error) {
	compound, err := combinedCompound(registries)
	if err != nil {
		return nil, err
	}

	return encodeNbt(func(ms *streams.MinecraftStream) error { return nbt.WriteNamed(ms, "", compound) })
}

// encodeDimensionType writes the dimension type a client before 1.19 reads
// out of its play login: the definition of the first dimension type among
// registries, which is the one the login puts the player into, as a root
// named with the empty string the way every NBT on the wire was named before
// 1.20.2. It is not a packet body but a field of one, alongside the
// registries the same login carries, and a set with no dimension type to
// spell out is refused, since such a login has nowhere to send the client.
func encodeDimensionType(registries []Registry) ([]byte, error) {
	for _, registry := range registries {
		if registry.Name != dimensionTypeRegistry {
			continue
		}

		if len(registry.Entries) == 0 {
			return nil, fmt.Errorf("registry %s holds no entry to put the player into", registry.Name)
		}

		entry := registry.Entries[0]
		if entry.Data == nil || entry.Data.Type() == nbt.TagEnd {
			return nil, fmt.Errorf("registry %s: entry %s has no definition, which a play login before 1.19 spells out", registry.Name, entry.Name)
		}

		return encodeNbt(func(ms *streams.MinecraftStream) error { return nbt.WriteNamed(ms, "", entry.Data) })
	}

	return nil, fmt.Errorf("no %s registry to take the play login's dimension type from", dimensionTypeRegistry)
}

// encodeDimensionList writes the registries as a 1.16.1 client reads them
// out of its play login: one compound, named as a root, holding under
// "dimension" a list of the dimension types, each an entry's name beside the
// fields of its definition rather than above them -- the client's codec
// pairs the name with a map codec, which lays the two out flat -- with no id,
// an entry's place in the list being its id. No other registry crosses the
// wire on that version, so a set holding one is refused rather than sent
// short of it.
func encodeDimensionList(registries []Registry) ([]byte, error) {
	var entries []nbt.Tag

	for _, registry := range registries {
		if registry.Name != dimensionTypeRegistry {
			return nil, fmt.Errorf("registry %s: a play login before 1.16.2 carries the dimension types alone", registry.Name)
		}

		for _, entry := range registry.Entries {
			definition, ok := entry.Data.(nbt.Compound)
			if !ok {
				return nil, fmt.Errorf("registry %s: entry %s has no definition to lay out beside its name", registry.Name, entry.Name)
			}

			if _, taken := definition["name"]; taken {
				return nil, fmt.Errorf("registry %s: entry %s has a field called name, which is where its own name goes", registry.Name, entry.Name)
			}

			flat := make(nbt.Compound, len(definition)+1)
			for field, value := range definition {
				flat[field] = value
			}

			flat["name"] = nbt.String(entry.Name)

			entries = append(entries, flat)
		}
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("no %s entry to put the player into", dimensionTypeRegistry)
	}

	compound := nbt.Compound{"dimension": nbt.List{ElementType: nbt.TagCompound, Elements: entries}}

	return encodeNbt(func(ms *streams.MinecraftStream) error { return nbt.WriteNamed(ms, "", compound) })
}

// encodeDimensionTypeName writes the dimension type a 1.16.1 client reads
// out of its play login: the name of the first dimension type among
// registries, as the string the login holds where every version from 1.16.2
// up to 1.19 holds the definition. Like that definition it is a field of the
// packet, written by the login transformer of the step it belongs to.
func encodeDimensionTypeName(registries []Registry) ([]byte, error) {
	for _, registry := range registries {
		if registry.Name == dimensionTypeRegistry && len(registry.Entries) > 0 {
			return encodeNbt(func(ms *streams.MinecraftStream) error { return ms.WriteString(registry.Entries[0].Name) })
		}
	}

	return nil, fmt.Errorf("no %s entry to take the play login's dimension type from", dimensionTypeRegistry)
}

// dimensionTypeRegistry is the registry the play login's dimension type is
// an entry of.
const dimensionTypeRegistry = "minecraft:dimension_type"

// encodeNbt runs one NBT write into a buffer and returns what it wrote.
func encodeNbt(write func(ms *streams.MinecraftStream) error) ([]byte, error) {
	buf := new(bytes.Buffer)
	ms := streams.NewMinecraftStreamFromBuffer(buf)

	if err := write(ms); err != nil {
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
