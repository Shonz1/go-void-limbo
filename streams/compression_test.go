package streams

import (
	"bytes"
	"strings"
	"testing"
)

func TestCompressRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{name: "empty", data: []byte{}},
		{name: "small", data: []byte("minecraft:overworld")},
		{name: "repetitive", data: bytes.Repeat([]byte("minecraft:dimension_type"), 512)},
		{name: "incompressible", data: []byte(strings.Repeat("\x00\xff\x7f\x80", 1))},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			compressed, err := Compress(test.data)
			if err != nil {
				t.Fatalf("compress error = %v", err)
			}

			got, err := Decompress(compressed, int32(len(test.data)))
			if err != nil {
				t.Fatalf("decompress error = %v", err)
			}

			if !bytes.Equal(got, test.data) {
				t.Errorf("round trip gave % x, want % x", got, test.data)
			}
		})
	}
}

// A registry packet is the reason compression is worth having, so the bytes
// actually have to come down.
func TestCompressShrinksRepetitiveData(t *testing.T) {
	data := bytes.Repeat([]byte("minecraft:dimension_type"), 512)

	compressed, err := Compress(data)
	if err != nil {
		t.Fatalf("compress error = %v", err)
	}

	if len(compressed) >= len(data) {
		t.Errorf("compressed %d bytes into %d, want fewer", len(data), len(compressed))
	}
}

func TestDecompressRejectsABodyThatIsNotTheSizeItClaims(t *testing.T) {
	data := []byte("minecraft:overworld")

	compressed, err := Compress(data)
	if err != nil {
		t.Fatalf("compress error = %v", err)
	}

	// The claimed size is the only bound on what a compressed body can expand
	// into, so a body that does not inflate to exactly it is refused rather
	// than kept in part. Shorter is the shape a decompression bomb arrives in.
	for _, size := range []int32{int32(len(data)) - 1, int32(len(data)) + 1, -1} {
		if _, err := Decompress(compressed, size); err == nil {
			t.Errorf("size %d: error = nil, want a body that inflates to %d to be refused", size, len(data))
		}
	}
}

func TestDecompressRejectsDataThatIsNotCompressed(t *testing.T) {
	if _, err := Decompress([]byte("not deflated at all"), 19); err == nil {
		t.Error("error = nil, want an error for a body that is not zlib data")
	}

	if _, err := Decompress(nil, 4); err == nil {
		t.Error("error = nil, want an error for an empty body")
	}
}
