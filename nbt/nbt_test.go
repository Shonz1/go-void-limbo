package nbt

import (
	"bytes"
	"github.com/Shonz1/go-void-limbo/streams"
	"reflect"
	"strings"
	"testing"
)

func encodeTag(t *testing.T, write func(ms *streams.MinecraftStream) error) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	ms := streams.NewMinecraftStreamFromBuffer(buf)

	if err := write(ms); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if err := ms.Flush(); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	return buf.Bytes()
}

func streamOver(data []byte) *streams.MinecraftStream {
	return streams.NewMinecraftStreamFromBuffer(bytes.NewBuffer(data))
}

// helloWorld is the canonical example from the NBT specification: a root
// compound named "hello world" holding a single string.
var helloWorld = Compound{"name": String("Bananrama")}

var helloWorldNamed = []byte{
	0x0A,                                                              // TAG_Compound
	0x00, 0x0B, 'h', 'e', 'l', 'l', 'o', ' ', 'w', 'o', 'r', 'l', 'd', // root name
	0x08,                           // TAG_String
	0x00, 0x04, 'n', 'a', 'm', 'e', // entry name
	0x00, 0x09, 'B', 'a', 'n', 'a', 'n', 'r', 'a', 'm', 'a', // entry value
	0x00, // TAG_End
}

func TestWriteNamedMatchesSpecExample(t *testing.T) {
	got := encodeTag(t, func(ms *streams.MinecraftStream) error {
		return WriteNamed(ms, "hello world", helloWorld)
	})

	if !bytes.Equal(got, helloWorldNamed) {
		t.Errorf("named encoding mismatch.\n got: % X\nwant: % X", got, helloWorldNamed)
	}
}

// TestWriteOmitsRootName covers the network form used from 1.20.2 onwards,
// where the root tag has a type byte and a payload but no name.
func TestWriteOmitsRootName(t *testing.T) {
	got := encodeTag(t, func(ms *streams.MinecraftStream) error {
		return Write(ms, helloWorld)
	})

	// Everything after the type byte and the 13-byte root name field is shared
	// with the named form.
	want := append([]byte{0x0A}, helloWorldNamed[14:]...)

	if !bytes.Equal(got, want) {
		t.Errorf("unnamed encoding mismatch.\n got: % X\nwant: % X", got, want)
	}
}

func TestReadNamed(t *testing.T) {
	name, tag, err := ReadNamed(streamOver(helloWorldNamed))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "hello world" {
		t.Errorf("root name = %q, want %q", name, "hello world")
	}

	if !reflect.DeepEqual(tag, helloWorld) {
		t.Errorf("tag = %v, want %v", tag, helloWorld)
	}
}

func TestRoundTrip(t *testing.T) {
	tags := map[string]Tag{
		"byte":             Byte(-128),
		"short":            Short(-4096),
		"int":              Int(-2147483648),
		"long":             Long(-9223372036854775808),
		"float":            Float(3.5),
		"double":           Double(-1.0 / 3.0),
		"string":           String("minecraft:plains"),
		"empty string":     String(""),
		"byte array":       ByteArray{0x00, 0x7F, 0x80, 0xFF},
		"int array":        IntArray{-1, 0, 1},
		"long array":       LongArray{-1, 0, 1},
		"empty byte array": ByteArray{},
		"list":             List{ElementType: TagString, Elements: []Tag{String("a"), String("b")}},
		"empty list":       List{ElementType: TagEnd, Elements: []Tag{}},
		"list of lists": List{ElementType: TagList, Elements: []Tag{
			List{ElementType: TagInt, Elements: []Tag{Int(1)}},
			List{ElementType: TagEnd, Elements: []Tag{}},
		}},
		"empty compound": Compound{},
		"nested compound": Compound{
			"effects":           Compound{"sky_color": Int(0x78A7FF), "fog_color": Int(0xC0D8FF)},
			"has_precipitation": Byte(0),
			"temperature":       Float(0.8),
		},
	}

	for name, tag := range tags {
		t.Run(name, func(t *testing.T) {
			encoded := encodeTag(t, func(ms *streams.MinecraftStream) error {
				return Write(ms, tag)
			})

			got, err := Read(streamOver(encoded))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(got, tag) {
				t.Errorf("round trip changed the tag.\n got: %v\nwant: %v", got, tag)
			}
		})
	}
}

func TestRoundTripNamed(t *testing.T) {
	encoded := encodeTag(t, func(ms *streams.MinecraftStream) error {
		return WriteNamed(ms, "root", Compound{"list": List{ElementType: TagByte, Elements: []Tag{Byte(1)}}})
	})

	name, tag, err := ReadNamed(streamOver(encoded))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if name != "root" {
		t.Errorf("name = %q, want %q", name, "root")
	}

	want := Compound{"list": List{ElementType: TagByte, Elements: []Tag{Byte(1)}}}
	if !reflect.DeepEqual(tag, want) {
		t.Errorf("tag = %v, want %v", tag, want)
	}
}

// TestEncodingIsDeterministic matters because registry payloads are encoded
// once at startup and reused, so Go's randomised map iteration must not leak
// into the bytes.
func TestEncodingIsDeterministic(t *testing.T) {
	tag := Compound{}
	for _, name := range []string{"zulu", "alpha", "mike", "bravo", "yankee", "charlie"} {
		tag[name] = String(name)
	}

	first := encodeTag(t, func(ms *streams.MinecraftStream) error { return Write(ms, tag) })

	for i := 0; i < 20; i++ {
		next := encodeTag(t, func(ms *streams.MinecraftStream) error { return Write(ms, tag) })
		if !bytes.Equal(first, next) {
			t.Fatalf("encoding is not stable across runs.\nfirst: % X\n next: % X", first, next)
		}
	}
}

func TestNilTagWritesEnd(t *testing.T) {
	for _, test := range []struct {
		name  string
		write func(ms *streams.MinecraftStream) error
	}{
		{"unnamed nil", func(ms *streams.MinecraftStream) error { return Write(ms, nil) }},
		{"named nil", func(ms *streams.MinecraftStream) error { return WriteNamed(ms, "root", nil) }},
		{"unnamed end", func(ms *streams.MinecraftStream) error { return Write(ms, End{}) }},
		{"named end", func(ms *streams.MinecraftStream) error { return WriteNamed(ms, "root", End{}) }},
	} {
		t.Run(test.name, func(t *testing.T) {
			got := encodeTag(t, test.write)

			if !bytes.Equal(got, []byte{0x00}) {
				t.Fatalf("encoded = % X, want 00", got)
			}

			tag, err := Read(streamOver(got))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tag != (End{}) {
				t.Errorf("tag = %v, want END", tag)
			}
		})
	}
}

func TestModifiedUtf8(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		encoded []byte
	}{
		{"ascii", "name", []byte{'n', 'a', 'm', 'e'}},
		// U+0000 is the two-byte C0 80 so that no encoded string holds a zero byte.
		{"null", "\x00", []byte{0xC0, 0x80}},
		{"two byte", "é", []byte{0xC3, 0xA9}},
		{"three byte", "€", []byte{0xE2, 0x82, 0xAC}},
		// U+1F600 is written as the surrogate pair D83D DE00, three bytes each,
		// rather than as a single four-byte sequence.
		{"supplementary", "😀", []byte{0xED, 0xA0, 0xBD, 0xED, 0xB8, 0x80}},
		{"mixed", "a\x00😀€", []byte{'a', 0xC0, 0x80, 0xED, 0xA0, 0xBD, 0xED, 0xB8, 0x80, 0xE2, 0x82, 0xAC}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			encoded, err := encodeModifiedUtf8(test.value)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !bytes.Equal(encoded, test.encoded) {
				t.Errorf("encoded = % X, want % X", encoded, test.encoded)
			}

			decoded, err := decodeModifiedUtf8(test.encoded)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if decoded != test.value {
				t.Errorf("decoded = %q, want %q", decoded, test.value)
			}
		})
	}
}

func TestStringRoundTripThroughStream(t *testing.T) {
	value := "§6ключ 😀 \x00"

	encoded := encodeTag(t, func(ms *streams.MinecraftStream) error {
		return Write(ms, String(value))
	})

	got, err := Read(streamOver(encoded))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != String(value) {
		t.Errorf("value = %q, want %q", got, value)
	}
}

func TestWriteRejectsOversizedString(t *testing.T) {
	buf := new(bytes.Buffer)
	err := Write(streams.NewMinecraftStreamFromBuffer(buf), String(strings.Repeat("a", maxStringLength+1)))

	if err == nil {
		t.Fatal("expected an error for a string longer than the length prefix allows")
	}
}

func TestWriteRejectsHeterogeneousList(t *testing.T) {
	buf := new(bytes.Buffer)
	list := List{ElementType: TagInt, Elements: []Tag{Int(1), String("nope")}}

	if err := Write(streams.NewMinecraftStreamFromBuffer(buf), list); err == nil {
		t.Fatal("expected an error for a list holding more than one tag type")
	}
}

// TestWriteInfersListElementType lets callers leave ElementType zero when the
// list is non-empty.
func TestWriteInfersListElementType(t *testing.T) {
	encoded := encodeTag(t, func(ms *streams.MinecraftStream) error {
		return Write(ms, List{Elements: []Tag{Int(7)}})
	})

	got, err := Read(streamOver(encoded))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := List{ElementType: TagInt, Elements: []Tag{Int(7)}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tag = %v, want %v", got, want)
	}
}

func TestReadRejectsMalformedInput(t *testing.T) {
	// The root type byte, then one unnamed compound opened inside the previous
	// one for every level of nesting.
	deeplyNested := []byte{byte(TagCompound)}
	for i := 0; i < maxDepth+64; i++ {
		deeplyNested = append(deeplyNested, byte(TagCompound), 0x00, 0x00)
	}

	tests := []struct {
		name string
		data []byte
	}{
		{"unknown tag type", []byte{0x0D}},
		{"unknown nested tag type", []byte{0x0A, 0x63, 0x00, 0x00}},
		{"truncated payload", []byte{0x03, 0x00, 0x00}},
		{"negative array length", []byte{0x0B, 0xFF, 0xFF, 0xFF, 0xFF}},
		{"implausible array length", []byte{0x0C, 0x7F, 0xFF, 0xFF, 0xFF}},
		{"non empty list of end", []byte{0x09, 0x00, 0x00, 0x00, 0x00, 0x01}},
		{"nesting past the depth limit", deeplyNested},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if _, err := Read(streamOver(test.data)); err == nil {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestSnbt(t *testing.T) {
	tests := []struct {
		tag  Tag
		want string
	}{
		{Byte(-1), "-1b"},
		{Short(2), "2s"},
		{Int(3), "3"},
		{Long(4), "4L"},
		{Float(0.5), "0.5f"},
		{Double(0.5), "0.5d"},
		{String(`say "hi"`), `"say \"hi\""`},
		{ByteArray{0xFF, 0x01}, "[B;-1b,1b]"},
		{IntArray{1, 2}, "[I;1,2]"},
		{LongArray{1, 2}, "[L;1L,2L]"},
		{List{ElementType: TagInt, Elements: []Tag{Int(1), Int(2)}}, "[1,2]"},
		{Compound{}, "{}"},
		{Compound{"b": Int(2), "a": Int(1)}, "{a:1,b:2}"},
		{Compound{"minecraft:id": Int(1)}, `{"minecraft:id":1}`},
		{End{}, "END"},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			if got := test.tag.String(); got != test.want {
				t.Errorf("String() = %s, want %s", got, test.want)
			}
		})
	}
}
