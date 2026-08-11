package gamedata

import (
	"embed"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/Shonz1/go-void-limbo/nbt"
)

//go:embed data/*.json
var dataFiles embed.FS

// loadDataFile parses one version's generated content out of the embedded data
// directory.
func loadDataFile(name string) ([]Registry, []TagSet, error) {
	raw, err := dataFiles.ReadFile("data/" + name)
	if err != nil {
		return nil, nil, fmt.Errorf("gamedata: %w", err)
	}

	var file dataFile
	if err := json.Unmarshal(raw, &file); err != nil {
		return nil, nil, fmt.Errorf("gamedata: parsing %s: %w", name, err)
	}

	registries, tags := file.toSets()

	return registries, tags, nil
}

// The generated registry content lives in JSON files under data/ rather than in
// Go source: diffs between versions stay data diffs, and adding a version stops
// touching compiled code. JSON has fewer number types than NBT, so every tag is
// written as an object with exactly one key naming its NBT type, which is what
// keeps a Byte from coming back as an Int and the encoded bytes from changing.

// tagTypeNames spells each NBT type the way the data files do.
var tagTypeNames = map[nbt.TagType]string{
	nbt.TagEnd:       "end",
	nbt.TagByte:      "byte",
	nbt.TagShort:     "short",
	nbt.TagInt:       "int",
	nbt.TagLong:      "long",
	nbt.TagFloat:     "float",
	nbt.TagDouble:    "double",
	nbt.TagByteArray: "byteArray",
	nbt.TagString:    "string",
	nbt.TagList:      "list",
	nbt.TagCompound:  "compound",
	nbt.TagIntArray:  "intArray",
	nbt.TagLongArray: "longArray",
}

var tagTypesByName = func() map[string]nbt.TagType {
	byName := make(map[string]nbt.TagType, len(tagTypeNames))
	for tagType, name := range tagTypeNames {
		byName[name] = tagType
	}

	return byName
}()

// tagValue carries one nbt.Tag through JSON.
//
// Longs travel as strings because a JSON number is a float64 to most readers,
// and a long does not fit one past 2^53. Lists carry their element type
// explicitly, because an empty list with a declared element type and an empty
// list without one encode differently on the wire, and the data files must
// round trip to the exact bytes the client was checked against.
type tagValue struct {
	Tag nbt.Tag
}

// jsonList is the shape a list takes in the files: the element type, and the
// elements themselves.
type jsonList struct {
	Of    string     `json:"of"`
	Items []tagValue `json:"items"`
}

func (v tagValue) MarshalJSON() ([]byte, error) {
	one := func(name string, value any) ([]byte, error) {
		return json.Marshal(map[string]any{name: value})
	}

	switch tag := v.Tag.(type) {
	case nbt.Byte:
		return one("byte", int8(tag))
	case nbt.Short:
		return one("short", int16(tag))
	case nbt.Int:
		return one("int", int32(tag))
	case nbt.Long:
		return one("long", strconv.FormatInt(int64(tag), 10))
	case nbt.Float:
		return one("float", float32(tag))
	case nbt.Double:
		return one("double", float64(tag))
	case nbt.String:
		return one("string", string(tag))
	case nbt.ByteArray:
		values := make([]int8, len(tag))
		for i, b := range tag {
			values[i] = int8(b)
		}

		return one("byteArray", values)
	case nbt.IntArray:
		return one("intArray", []int32(tag))
	case nbt.LongArray:
		values := make([]string, len(tag))
		for i, l := range tag {
			values[i] = strconv.FormatInt(l, 10)
		}

		return one("longArray", values)
	case nbt.List:
		of, ok := tagTypeNames[tag.ElementType]
		if !ok {
			return nil, fmt.Errorf("gamedata: list of unknown element type %d", tag.ElementType)
		}

		items := make([]tagValue, len(tag.Elements))
		for i, element := range tag.Elements {
			items[i] = tagValue{Tag: element}
		}

		return one("list", jsonList{Of: of, Items: items})
	case nbt.Compound:
		entries := make(map[string]tagValue, len(tag))
		for name, value := range tag {
			entries[name] = tagValue{Tag: value}
		}

		return one("compound", entries)
	default:
		return nil, fmt.Errorf("gamedata: cannot marshal tag %T", v.Tag)
	}
}

func (v *tagValue) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if len(raw) != 1 {
		return fmt.Errorf("gamedata: a tag is one type and one value, got %d keys", len(raw))
	}

	for name, value := range raw {
		tag, err := unmarshalTag(name, value)
		if err != nil {
			return err
		}

		v.Tag = tag
	}

	return nil
}

func unmarshalTag(name string, value json.RawMessage) (nbt.Tag, error) {
	switch name {
	case "byte":
		var parsed int8
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		return nbt.Byte(parsed), nil
	case "short":
		var parsed int16
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		return nbt.Short(parsed), nil
	case "int":
		var parsed int32
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		return nbt.Int(parsed), nil
	case "long":
		var text string
		if err := json.Unmarshal(value, &text); err != nil {
			return nil, err
		}

		parsed, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return nil, err
		}

		return nbt.Long(parsed), nil
	case "float":
		var parsed float32
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		return nbt.Float(parsed), nil
	case "double":
		var parsed float64
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		return nbt.Double(parsed), nil
	case "string":
		var parsed string
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		return nbt.String(parsed), nil
	case "byteArray":
		var parsed []int8
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		bytes := make(nbt.ByteArray, len(parsed))
		for i, b := range parsed {
			bytes[i] = byte(b)
		}

		return bytes, nil
	case "intArray":
		var parsed []int32
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		return nbt.IntArray(parsed), nil
	case "longArray":
		var texts []string
		if err := json.Unmarshal(value, &texts); err != nil {
			return nil, err
		}

		longs := make(nbt.LongArray, len(texts))
		for i, text := range texts {
			parsed, err := strconv.ParseInt(text, 10, 64)
			if err != nil {
				return nil, err
			}

			longs[i] = parsed
		}

		return longs, nil
	case "list":
		var parsed jsonList
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		elementType, ok := tagTypesByName[parsed.Of]
		if !ok {
			return nil, fmt.Errorf("gamedata: list of unknown element type %q", parsed.Of)
		}

		// An empty list keeps nil elements, the way the literals spelled it, so
		// a loaded set is deeply equal to one written in source.
		var elements []nbt.Tag
		for _, item := range parsed.Items {
			elements = append(elements, item.Tag)
		}

		return nbt.List{ElementType: elementType, Elements: elements}, nil
	case "compound":
		var parsed map[string]tagValue
		if err := json.Unmarshal(value, &parsed); err != nil {
			return nil, err
		}

		compound := make(nbt.Compound, len(parsed))
		for entryName, entry := range parsed {
			compound[entryName] = entry.Tag
		}

		return compound, nil
	default:
		return nil, fmt.Errorf("gamedata: unknown tag type %q", name)
	}
}

// dataFile is one version's generated content: the registries in the order
// they are sent, and every tag the client's jar declares.
type dataFile struct {
	Registries []dataRegistry `json:"registries"`
	Tags       []dataTagSet   `json:"tags"`
}

type dataRegistry struct {
	Name    string      `json:"name"`
	Entries []dataEntry `json:"entries"`
}

type dataEntry struct {
	Name string    `json:"name"`
	Data *tagValue `json:"data,omitempty"`
}

type dataTagSet struct {
	Registry string         `json:"registry"`
	Tags     []dataNamedTag `json:"tags"`
}

type dataNamedTag struct {
	Name    string  `json:"name"`
	Entries []int32 `json:"entries,omitempty"`
}

// toDataFile is the inverse of toSets, for the generator that writes the files.
func toDataFile(registries []Registry, tags []TagSet) dataFile {
	file := dataFile{
		Registries: make([]dataRegistry, len(registries)),
		Tags:       make([]dataTagSet, len(tags)),
	}

	for i, registry := range registries {
		entries := make([]dataEntry, len(registry.Entries))
		for j, entry := range registry.Entries {
			entries[j] = dataEntry{Name: entry.Name}
			if entry.Data != nil {
				entries[j].Data = &tagValue{Tag: entry.Data}
			}
		}

		file.Registries[i] = dataRegistry{Name: registry.Name, Entries: entries}
	}

	for i, set := range tags {
		named := make([]dataNamedTag, len(set.Tags))
		for j, tag := range set.Tags {
			named[j] = dataNamedTag{Name: tag.Name, Entries: tag.Entries}
		}

		file.Tags[i] = dataTagSet{Registry: set.Registry, Tags: named}
	}

	return file
}

// toSets turns a parsed file back into the registries and tags the provider is
// built from.
func (f dataFile) toSets() ([]Registry, []TagSet) {
	registries := make([]Registry, len(f.Registries))
	for i, registry := range f.Registries {
		entries := make([]Entry, len(registry.Entries))
		for j, entry := range registry.Entries {
			entries[j] = Entry{Name: entry.Name}
			if entry.Data != nil {
				entries[j].Data = entry.Data.Tag
			}
		}

		registries[i] = Registry{Name: registry.Name, Entries: entries}
	}

	tags := make([]TagSet, len(f.Tags))
	for i, set := range f.Tags {
		named := make([]NamedTag, len(set.Tags))
		for j, tag := range set.Tags {
			named[j] = NamedTag{Name: tag.Name, Entries: tag.Entries}
		}

		tags[i] = TagSet{Registry: set.Registry, Tags: named}
	}

	return registries, tags
}
