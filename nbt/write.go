package nbt

import (
	"fmt"
	"go-void-limbo/streams"
	"math"
)

// maxStringLength is the ceiling the format imposes on names and string
// payloads, which are prefixed with an unsigned short byte count.
const maxStringLength = math.MaxUint16

// Write writes tag as an unnamed root: a type byte followed by the payload,
// with no name field. This is the form the network protocol uses from 1.20.2
// onwards. A nil tag writes a bare TagEnd, which is how an absent optional NBT
// field is encoded.
func Write(ms *streams.MinecraftStream, tag Tag) error {
	if tag == nil || tag.Type() == TagEnd {
		return ms.WriteByte(byte(TagEnd))
	}

	if err := ms.WriteByte(byte(tag.Type())); err != nil {
		return err
	}

	return tag.writePayload(ms)
}

// WriteNamed writes tag as a named root: a type byte, the name, then the
// payload. This is the form used by files on disk and by the network protocol
// before 1.20.2. A TagEnd root carries no name, so name is ignored in that case.
func WriteNamed(ms *streams.MinecraftStream, name string, tag Tag) error {
	if tag == nil || tag.Type() == TagEnd {
		return ms.WriteByte(byte(TagEnd))
	}

	if err := ms.WriteByte(byte(tag.Type())); err != nil {
		return err
	}

	if err := writeString(ms, name); err != nil {
		return err
	}

	return tag.writePayload(ms)
}

func (End) writePayload(*streams.MinecraftStream) error {
	return nil
}

func (t Byte) writePayload(ms *streams.MinecraftStream) error {
	return ms.WriteByte(byte(t))
}

func (t Short) writePayload(ms *streams.MinecraftStream) error {
	return ms.WriteShort(int16(t))
}

func (t Int) writePayload(ms *streams.MinecraftStream) error {
	return ms.WriteInt(int32(t))
}

func (t Long) writePayload(ms *streams.MinecraftStream) error {
	return ms.WriteLong(int64(t))
}

func (t Float) writePayload(ms *streams.MinecraftStream) error {
	return ms.WriteFloat(float32(t))
}

func (t Double) writePayload(ms *streams.MinecraftStream) error {
	return ms.WriteDouble(float64(t))
}

func (t ByteArray) writePayload(ms *streams.MinecraftStream) error {
	if err := ms.WriteInt(int32(len(t))); err != nil {
		return err
	}

	return ms.WriteBytes(t)
}

func (t String) writePayload(ms *streams.MinecraftStream) error {
	return writeString(ms, string(t))
}

func (t IntArray) writePayload(ms *streams.MinecraftStream) error {
	if err := ms.WriteInt(int32(len(t))); err != nil {
		return err
	}

	for _, value := range t {
		if err := ms.WriteInt(value); err != nil {
			return err
		}
	}

	return nil
}

func (t LongArray) writePayload(ms *streams.MinecraftStream) error {
	if err := ms.WriteInt(int32(len(t))); err != nil {
		return err
	}

	for _, value := range t {
		if err := ms.WriteLong(value); err != nil {
			return err
		}
	}

	return nil
}

func (t List) writePayload(ms *streams.MinecraftStream) error {
	elementType := t.ElementType
	if elementType == TagEnd && len(t.Elements) > 0 {
		elementType = t.Elements[0].Type()
	}

	for i, element := range t.Elements {
		if element == nil {
			return fmt.Errorf("nbt: list element %d is nil", i)
		}

		if element.Type() != elementType {
			return fmt.Errorf("nbt: list element %d is %s, want %s", i, element.Type(), elementType)
		}
	}

	if err := ms.WriteByte(byte(elementType)); err != nil {
		return err
	}

	if err := ms.WriteInt(int32(len(t.Elements))); err != nil {
		return err
	}

	for _, element := range t.Elements {
		if err := element.writePayload(ms); err != nil {
			return err
		}
	}

	return nil
}

func (t Compound) writePayload(ms *streams.MinecraftStream) error {
	for _, name := range t.sortedNames() {
		value := t[name]
		if value == nil || value.Type() == TagEnd {
			return fmt.Errorf("nbt: compound entry %q has no value", name)
		}

		if err := ms.WriteByte(byte(value.Type())); err != nil {
			return err
		}

		if err := writeString(ms, name); err != nil {
			return err
		}

		if err := value.writePayload(ms); err != nil {
			return err
		}
	}

	return ms.WriteByte(byte(TagEnd))
}

func writeString(ms *streams.MinecraftStream, value string) error {
	encoded, err := encodeModifiedUtf8(value)
	if err != nil {
		return err
	}

	if err := ms.WriteShort(int16(uint16(len(encoded)))); err != nil {
		return err
	}

	return ms.WriteBytes(encoded)
}
