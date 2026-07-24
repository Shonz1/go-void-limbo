package streams

import (
	"bytes"
	"io"
	"testing"
)

// fragmentedReadWriter wraps a byte slice and returns at most maxChunk bytes
// per Read call, simulating a TCP connection that delivers data in small
// fragments rather than filling the caller's buffer in one call.
type fragmentedReadWriter struct {
	data     []byte
	pos      int
	maxChunk int
	written  bytes.Buffer
}

func (f *fragmentedReadWriter) Read(b []byte) (int, error) {
	if f.pos >= len(f.data) {
		return 0, io.EOF
	}

	n := f.maxChunk
	if n > len(b) {
		n = len(b)
	}
	if remaining := len(f.data) - f.pos; n > remaining {
		n = remaining
	}

	copy(b, f.data[f.pos:f.pos+n])
	f.pos += n
	return n, nil
}

func (f *fragmentedReadWriter) ReadByte() (byte, error) {
	var b [1]byte
	n, err := f.Read(b[:])
	if n == 0 && err != nil {
		return 0, err
	}
	return b[0], nil
}

func (f *fragmentedReadWriter) Write(b []byte) (int, error) {
	return f.written.Write(b)
}

func (f *fragmentedReadWriter) WriteByte(c byte) error {
	return f.written.WriteByte(c)
}

func (f *fragmentedReadWriter) Flush() error {
	return nil
}

func TestReadBytesReassemblesFragmentedReads(t *testing.T) {
	payload := []byte("this is a payload longer than four bytes")
	rw := &fragmentedReadWriter{data: payload, maxChunk: 4}
	s := NewMinecraftStream(rw)

	got, err := s.ReadBytes(int32(len(payload)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(got, payload) {
		t.Errorf("payload not reassembled correctly.\n got: %q\nwant: %q", got, payload)
	}
}

func TestVarIntRoundTrip(t *testing.T) {
	cases := []int32{0, 1, 127, 128, 255, 300, 2097151, 2097152, 2147483647, -1}

	for _, v := range cases {
		buf := new(bytes.Buffer)
		s := NewMinecraftStreamFromBuffer(buf)

		if err := s.WriteVarInt(v); err != nil {
			t.Fatalf("WriteVarInt(%d): %v", v, err)
		}
		if err := s.Flush(); err != nil {
			t.Fatalf("Flush: %v", err)
		}

		s2 := NewMinecraftStreamFromBuffer(buf)
		got, err := s2.ReadVarInt()
		if err != nil {
			t.Fatalf("ReadVarInt after WriteVarInt(%d): %v", v, err)
		}
		if got != v {
			t.Errorf("VarInt round-trip mismatch: wrote %d, got %d", v, got)
		}
	}
}

func TestStringRoundTrip(t *testing.T) {
	cases := []string{"", "hello", "a longer string with spaces and punctuation!"}

	for _, v := range cases {
		buf := new(bytes.Buffer)
		s := NewMinecraftStreamFromBuffer(buf)

		if err := s.WriteString(v); err != nil {
			t.Fatalf("WriteString(%q): %v", v, err)
		}
		if err := s.Flush(); err != nil {
			t.Fatalf("Flush: %v", err)
		}

		s2 := NewMinecraftStreamFromBuffer(buf)
		got, err := s2.ReadString()
		if err != nil {
			t.Fatalf("ReadString after WriteString(%q): %v", v, err)
		}
		if got != v {
			t.Errorf("String round-trip mismatch: wrote %q, got %q", v, got)
		}
	}
}

func TestReadUuid(t *testing.T) {
	raw := []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
	}
	buf := bytes.NewBuffer(raw)
	s := NewMinecraftStreamFromBuffer(buf)

	got, err := s.ReadUuid()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := "01020304-0506-0708-090a-0b0c0d0e0f10"
	if got != want {
		t.Errorf("ReadUuid mismatch: got %q, want %q", got, want)
	}
}
