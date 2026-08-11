package nbt

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

const (
	// maxDepth bounds nesting so that a hostile payload of nothing but opening
	// compounds cannot exhaust the stack. Vanilla applies the same limit.
	maxDepth = 512

	// maxElements bounds how much a single length prefix can make the reader
	// allocate. A packet body cannot exceed 2^21-1 bytes, so no list or array
	// arriving over the network can legitimately reach this many elements.
	maxElements = 1 << 21
)

// Read reads an unnamed root tag, the form the network protocol uses from
// 1.20.2 onwards. A TagEnd on the wire means the field is absent and is
// returned as End{}.
func Read(ms *streams.MinecraftStream) (Tag, error) {
	tagType, err := readTagType(ms)
	if err != nil {
		return nil, err
	}

	if tagType == TagEnd {
		return End{}, nil
	}

	return readPayload(ms, tagType, 0)
}

// ReadNamed reads a named root tag, the form used by files and by the network
// protocol before 1.20.2. A TagEnd root carries no name and yields End{} with
// an empty name.
func ReadNamed(ms *streams.MinecraftStream) (string, Tag, error) {
	tagType, err := readTagType(ms)
	if err != nil {
		return "", nil, err
	}

	if tagType == TagEnd {
		return "", End{}, nil
	}

	name, err := readString(ms)
	if err != nil {
		return "", nil, err
	}

	tag, err := readPayload(ms, tagType, 0)
	if err != nil {
		return "", nil, err
	}

	return name, tag, nil
}

func readTagType(ms *streams.MinecraftStream) (TagType, error) {
	value, err := ms.ReadByte()
	if err != nil {
		return TagEnd, err
	}

	tagType := TagType(value)
	if tagType > TagLongArray {
		return TagEnd, fmt.Errorf("nbt: unknown tag type %d", value)
	}

	return tagType, nil
}

func readPayload(ms *streams.MinecraftStream, tagType TagType, depth int) (Tag, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("nbt: nesting deeper than %d levels", maxDepth)
	}

	switch tagType {
	case TagEnd:
		return End{}, nil
	case TagByte:
		value, err := ms.ReadByte()
		return Byte(value), err
	case TagShort:
		value, err := ms.ReadShort()
		return Short(value), err
	case TagInt:
		value, err := ms.ReadInt()
		return Int(value), err
	case TagLong:
		value, err := ms.ReadLong()
		return Long(value), err
	case TagFloat:
		value, err := ms.ReadFloat()
		return Float(value), err
	case TagDouble:
		value, err := ms.ReadDouble()
		return Double(value), err
	case TagString:
		value, err := readString(ms)
		return String(value), err
	case TagByteArray:
		return readByteArray(ms)
	case TagIntArray:
		return readIntArray(ms)
	case TagLongArray:
		return readLongArray(ms)
	case TagList:
		return readList(ms, depth)
	case TagCompound:
		return readCompound(ms, depth)
	}

	return nil, fmt.Errorf("nbt: unknown tag type %d", tagType)
}

func readByteArray(ms *streams.MinecraftStream) (ByteArray, error) {
	length, err := readLength(ms, "byte array")
	if err != nil {
		return nil, err
	}

	value, err := ms.ReadBytes(length)
	if err != nil {
		return nil, err
	}

	return ByteArray(value), nil
}

func readIntArray(ms *streams.MinecraftStream) (IntArray, error) {
	length, err := readLength(ms, "int array")
	if err != nil {
		return nil, err
	}

	values := make(IntArray, length)
	for i := range values {
		if values[i], err = ms.ReadInt(); err != nil {
			return nil, err
		}
	}

	return values, nil
}

func readLongArray(ms *streams.MinecraftStream) (LongArray, error) {
	length, err := readLength(ms, "long array")
	if err != nil {
		return nil, err
	}

	values := make(LongArray, length)
	for i := range values {
		if values[i], err = ms.ReadLong(); err != nil {
			return nil, err
		}
	}

	return values, nil
}

func readList(ms *streams.MinecraftStream, depth int) (List, error) {
	elementType, err := readTagType(ms)
	if err != nil {
		return List{}, err
	}

	length, err := readLength(ms, "list")
	if err != nil {
		return List{}, err
	}

	// A list of TagEnd only exists as the empty list. Anything else would ask
	// the reader to produce a payload the format does not define.
	if elementType == TagEnd && length > 0 {
		return List{}, fmt.Errorf("nbt: list of %s has %d elements, want 0", TagEnd, length)
	}

	list := List{ElementType: elementType, Elements: make([]Tag, length)}
	for i := range list.Elements {
		if list.Elements[i], err = readPayload(ms, elementType, depth+1); err != nil {
			return List{}, err
		}
	}

	return list, nil
}

func readCompound(ms *streams.MinecraftStream, depth int) (Compound, error) {
	compound := Compound{}

	for {
		tagType, err := readTagType(ms)
		if err != nil {
			return nil, err
		}

		if tagType == TagEnd {
			return compound, nil
		}

		name, err := readString(ms)
		if err != nil {
			return nil, err
		}

		value, err := readPayload(ms, tagType, depth+1)
		if err != nil {
			return nil, err
		}

		compound[name] = value
	}
}

// readLength reads an element count and rejects the values that would have the
// caller allocate an implausible amount of memory.
func readLength(ms *streams.MinecraftStream, what string) (int32, error) {
	length, err := ms.ReadInt()
	if err != nil {
		return 0, err
	}

	if length < 0 {
		return 0, fmt.Errorf("nbt: %s has negative length %d", what, length)
	}

	if length > maxElements {
		return 0, fmt.Errorf("nbt: %s has %d elements, limit is %d", what, length, maxElements)
	}

	return length, nil
}

func readString(ms *streams.MinecraftStream) (string, error) {
	length, err := ms.ReadShort()
	if err != nil {
		return "", err
	}

	encoded, err := ms.ReadBytes(int32(uint16(length)))
	if err != nil {
		return "", err
	}

	return decodeModifiedUtf8(encoded)
}
