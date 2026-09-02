package streams

import (
	"bytes"
	"testing"
)

func TestWriteCompressedFrameWritesWhatCompressBodyWouldHaveFramed(t *testing.T) {
	payload := bytes.Repeat([]byte{0xAB}, 300)

	for _, size := range []int32{0, 300, 100000} {
		buf := new(bytes.Buffer)
		stream := NewMinecraftStreamFromBuffer(buf)

		if err := stream.WriteCompressedFrame(size, payload); err != nil {
			t.Fatalf("WriteCompressedFrame(%d) error: %v", size, err)
		}

		// The frame CompressBody's framing would have produced for the same
		// size and payload, written whole.
		body := AppendVarInt(nil, size)
		body = append(body, payload...)

		wantBuf := new(bytes.Buffer)
		if err := NewMinecraftStreamFromBuffer(wantBuf).WriteFrame(body); err != nil {
			t.Fatalf("WriteFrame() error: %v", err)
		}

		if !bytes.Equal(buf.Bytes(), wantBuf.Bytes()) {
			t.Errorf("size %d: wrote % x, want % x", size, buf.Bytes(), wantBuf.Bytes())
		}
	}
}
