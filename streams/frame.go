package streams

import (
	"bytes"
	"fmt"
)

// MaxPacketSize is the largest packet body the protocol allows (2^21 - 1
// bytes).
const MaxPacketSize = 2097151

// ReadFrame reads one length-prefixed frame off the stream and returns the
// body inside it. The body is read in full even when it is about to be refused
// further up, so the frame after it starts where it should.
func (s *MinecraftStream) ReadFrame() ([]byte, error) {
	length, err := s.ReadVarInt()
	if err != nil {
		return nil, err
	}

	if length < 1 || length > MaxPacketSize {
		return nil, fmt.Errorf("invalid packet length: %d", length)
	}

	return s.ReadBytes(length)
}

// WriteFrame writes body as one length-prefixed frame and flushes it, which is
// the whole of the framing on a connection that has not been told a
// compression threshold.
func (s *MinecraftStream) WriteFrame(body []byte) error {
	if err := s.WriteVarInt(int32(len(body))); err != nil {
		return err
	}

	if err := s.WriteBytes(body); err != nil {
		return err
	}

	return s.Flush()
}

// CompressBody frames a packet body for a connection that has been told a
// compression threshold, deflating it when it is big enough to be worth it. The
// var int in front carries the size the body inflates to, or zero for a body
// small enough to be left in full.
func CompressBody(body []byte, threshold int32) ([]byte, error) {
	size := int32(0)
	payload := body

	if int32(len(body)) >= threshold {
		compressed, err := Compress(body)
		if err != nil {
			return nil, fmt.Errorf("failed to compress packet: %w", err)
		}

		size = int32(len(body))
		payload = compressed
	}

	buf := new(bytes.Buffer)
	stream := NewMinecraftStreamFromBuffer(buf)

	if err := stream.WriteVarInt(size); err != nil {
		return nil, err
	}

	if err := stream.Flush(); err != nil {
		return nil, err
	}

	return append(buf.Bytes(), payload...), nil
}

// DecompressBody undoes CompressBody on a body that arrived from a client that
// was told the same threshold. A size the client had no business compressing at
// is refused: a body it should have sent in full is one this end cannot tell
// from a frame that lost its place.
func DecompressBody(body []byte, threshold int32) ([]byte, error) {
	size, read, err := ReadVarIntFrom(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}

	payload := body[read:]

	if size == 0 {
		return payload, nil
	}

	if size < threshold || size > MaxPacketSize {
		return nil, fmt.Errorf("invalid data length: %d", size)
	}

	return Decompress(payload, size)
}
