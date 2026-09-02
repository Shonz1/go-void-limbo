package streams

import (
	"fmt"
	"io"
)

// MaxPacketSize is the largest packet body the protocol allows (2^21 - 1
// bytes).
const MaxPacketSize = 2097151

// ReadFrame reads one length-prefixed frame off the stream and returns the
// body inside it. The length arrives from the other end, and the body is
// allocated before a byte of it has been read, so max is what the caller knows
// a frame can hold, up to the MaxPacketSize the protocol itself allows:
// anything longer is refused rather than reserved for. A body small enough to
// be read is read in full even when it is about to be refused further up, so
// the frame after it starts where it should.
func (s *MinecraftStream) ReadFrame(max int32) ([]byte, error) {
	length, err := s.ReadVarInt()
	if err != nil {
		return nil, err
	}

	if length < 1 || length > max {
		return nil, fmt.Errorf("invalid packet length: %d", length)
	}

	return s.ReadBytes(length)
}

// frameScratchLimit is the biggest frame body the scratch a connection reuses
// may grow to hold. The traffic a limbo carries is tens of bytes a packet, so
// anything bigger is rare and gets a buffer of its own rather than pinning its
// size to the connection for good.
const frameScratchLimit = 4096

// ReadFrameInto is ReadFrame for a caller with one frame in flight at a time:
// the body lands in scratch, which grows to the biggest ordinary frame the
// connection has seen and is then reused for every one after it, so a steady
// stream of packets is not a steady stream of allocations. The returned body
// aliases scratch and is only good until the next call.
func (s *MinecraftStream) ReadFrameInto(scratch *[]byte, max int32) ([]byte, error) {
	length, err := s.ReadVarInt()
	if err != nil {
		return nil, err
	}

	if length < 1 || length > max {
		return nil, fmt.Errorf("invalid packet length: %d", length)
	}

	if length > frameScratchLimit {
		return s.ReadBytes(length)
	}

	if cap(*scratch) < int(length) {
		*scratch = make([]byte, length)
	}

	buf := (*scratch)[:length]
	_, err = io.ReadFull(s.stream, buf)

	return buf, err
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

// WriteCompressedFrame writes one frame for a connection that has been told a
// compression threshold, from the two parts CompressBody would have joined:
// the size the payload inflates to, or zero for one that travels in full, and
// the payload itself. The parts go to the connection as they are, which is
// what lets a body deflated ahead of time be sent without being copied.
func (s *MinecraftStream) WriteCompressedFrame(size int32, payload []byte) error {
	var sizeScratch [5]byte
	sizeBytes := AppendVarInt(sizeScratch[:0], size)

	if err := s.WriteVarInt(int32(len(sizeBytes) + len(payload))); err != nil {
		return err
	}

	if err := s.WriteBytes(sizeBytes); err != nil {
		return err
	}

	if err := s.WriteBytes(payload); err != nil {
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

	framed := AppendVarInt(make([]byte, 0, 5+len(payload)), size)

	return append(framed, payload...), nil
}

// DecompressBody undoes CompressBody on a body that arrived from a client that
// was told the same threshold. A size the client had no business compressing at
// is refused: a body it should have sent in full is one this end cannot tell
// from a frame that lost its place. The size is also the client's word on how
// much memory inflating is worth, so max bounds it the same way it bounds the
// frame the body arrived in.
func DecompressBody(body []byte, threshold int32, max int32) ([]byte, error) {
	size, read, err := ReadVarIntFrom(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read data length: %w", err)
	}

	payload := body[read:]

	if size == 0 {
		return payload, nil
	}

	if size < threshold || size > max {
		return nil, fmt.Errorf("invalid data length: %d", size)
	}

	return Decompress(payload, size)
}
