package streams

import (
	"bytes"
	"testing"
)

// benchBody is a packet-sized body with enough repetition to be worth
// deflating, the way the registry data a limbo actually compresses is.
var benchBody = bytes.Repeat([]byte("minecraft:the_void"), 128)

func BenchmarkCompress(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		if _, err := Compress(benchBody); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecompress(b *testing.B) {
	compressed, err := Compress(benchBody)
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		if _, err := Decompress(compressed, int32(len(benchBody))); err != nil {
			b.Fatal(err)
		}
	}
}
