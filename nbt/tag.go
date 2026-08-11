// Package nbt implements Minecraft's Named Binary Tag format.
//
// A tag is a type byte followed by a payload. Only the root tag carries a name,
// and whether it does depends on where the tag is used: files and pre-1.20.2
// network packets write a named root, while the modern network protocol writes
// the root unnamed. Write and WriteNamed cover the two forms; nested tags inside
// a Compound always carry names regardless.
package nbt

import (
	"github.com/Shonz1/go-void-limbo/streams"
	"sort"
	"strconv"
	"strings"
)

type TagType byte

const (
	TagEnd TagType = iota
	TagByte
	TagShort
	TagInt
	TagLong
	TagFloat
	TagDouble
	TagByteArray
	TagString
	TagList
	TagCompound
	TagIntArray
	TagLongArray
)

func (t TagType) String() string {
	names := [...]string{
		"TAG_End", "TAG_Byte", "TAG_Short", "TAG_Int", "TAG_Long",
		"TAG_Float", "TAG_Double", "TAG_Byte_Array", "TAG_String",
		"TAG_List", "TAG_Compound", "TAG_Int_Array", "TAG_Long_Array",
	}

	if int(t) >= len(names) {
		return "TAG_Unknown(" + strconv.Itoa(int(t)) + ")"
	}

	return names[t]
}

// Tag is a single NBT value. The set of implementations is closed: writePayload
// is unexported so that every Tag a caller can build is also one Read can
// produce.
//
// String renders the tag as SNBT, the textual form Minecraft itself uses, which
// makes tags readable when logged through slog.
type Tag interface {
	Type() TagType
	String() string

	writePayload(ms *streams.MinecraftStream) error
}

// End is the tag that terminates a Compound. It appears as a value only as the
// root of an empty network NBT field, where Read returns it in place of a
// missing tag.
type End struct{}

type Byte int8

type Short int16

type Int int32

type Long int64

type Float float32

type Double float64

// ByteArray holds signed bytes on the wire. Go's unsigned byte is used for the
// element type because the payload is almost always opaque data, but String
// renders the elements signed, the way Minecraft does.
type ByteArray []byte

type String string

// List holds homogeneous, unnamed tags. ElementType may be left zero when
// Elements is non-empty: the write path then infers it from the first element.
// An empty list writes TagEnd as its element type, which is the canonical
// encoding.
type List struct {
	ElementType TagType
	Elements    []Tag
}

// Compound maps names to tags. NBT compounds are unordered, so entries are
// written in sorted key order: that keeps encoding deterministic, which matters
// for payloads that are encoded once and reused across connections.
type Compound map[string]Tag

type IntArray []int32

type LongArray []int64

func (End) Type() TagType       { return TagEnd }
func (Byte) Type() TagType      { return TagByte }
func (Short) Type() TagType     { return TagShort }
func (Int) Type() TagType       { return TagInt }
func (Long) Type() TagType      { return TagLong }
func (Float) Type() TagType     { return TagFloat }
func (Double) Type() TagType    { return TagDouble }
func (ByteArray) Type() TagType { return TagByteArray }
func (String) Type() TagType    { return TagString }
func (List) Type() TagType      { return TagList }
func (Compound) Type() TagType  { return TagCompound }
func (IntArray) Type() TagType  { return TagIntArray }
func (LongArray) Type() TagType { return TagLongArray }

func (End) String() string { return "END" }

func (t Byte) String() string { return strconv.FormatInt(int64(t), 10) + "b" }

func (t Short) String() string { return strconv.FormatInt(int64(t), 10) + "s" }

func (t Int) String() string { return strconv.FormatInt(int64(t), 10) }

func (t Long) String() string { return strconv.FormatInt(int64(t), 10) + "L" }

func (t Float) String() string { return strconv.FormatFloat(float64(t), 'g', -1, 32) + "f" }

func (t Double) String() string { return strconv.FormatFloat(float64(t), 'g', -1, 64) + "d" }

func (t String) String() string { return quoteSnbt(string(t)) }

func (t ByteArray) String() string {
	var sb strings.Builder
	sb.WriteString("[B;")

	for i, value := range t {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatInt(int64(int8(value)), 10))
		sb.WriteByte('b')
	}

	sb.WriteByte(']')
	return sb.String()
}

func (t IntArray) String() string {
	var sb strings.Builder
	sb.WriteString("[I;")

	for i, value := range t {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatInt(int64(value), 10))
	}

	sb.WriteByte(']')
	return sb.String()
}

func (t LongArray) String() string {
	var sb strings.Builder
	sb.WriteString("[L;")

	for i, value := range t {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatInt(value, 10))
		sb.WriteByte('L')
	}

	sb.WriteByte(']')
	return sb.String()
}

func (t List) String() string {
	var sb strings.Builder
	sb.WriteByte('[')

	for i, element := range t.Elements {
		if i > 0 {
			sb.WriteByte(',')
		}
		if element == nil {
			sb.WriteString("END")
			continue
		}
		sb.WriteString(element.String())
	}

	sb.WriteByte(']')
	return sb.String()
}

func (t Compound) String() string {
	var sb strings.Builder
	sb.WriteByte('{')

	for i, name := range t.sortedNames() {
		if i > 0 {
			sb.WriteByte(',')
		}

		sb.WriteString(quoteSnbtName(name))
		sb.WriteByte(':')

		if value := t[name]; value != nil {
			sb.WriteString(value.String())
		} else {
			sb.WriteString("END")
		}
	}

	sb.WriteByte('}')
	return sb.String()
}

func (t Compound) sortedNames() []string {
	names := make([]string, 0, len(t))
	for name := range t {
		names = append(names, name)
	}

	sort.Strings(names)
	return names
}

// quoteSnbtName leaves names bare when SNBT allows it and quotes them otherwise.
func quoteSnbtName(name string) string {
	if name == "" {
		return `""`
	}

	for _, r := range name {
		bare := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' ||
			r == '_' || r == '-' || r == '.' || r == '+'
		if !bare {
			return quoteSnbt(name)
		}
	}

	return name
}

func quoteSnbt(value string) string {
	var sb strings.Builder
	sb.WriteByte('"')

	for _, r := range value {
		if r == '"' || r == '\\' {
			sb.WriteByte('\\')
		}
		sb.WriteRune(r)
	}

	sb.WriteByte('"')
	return sb.String()
}
