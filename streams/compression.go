package streams

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
)

// Compress deflates a packet body. Once a connection has been told a
// compression threshold, every body at or above it travels this way.
func Compress(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)
	writer := zlib.NewWriter(buf)

	if _, err := writer.Write(data); err != nil {
		writer.Close()

		return nil, err
	}

	// Close is what writes the trailing checksum, so a body is only whole once
	// it has been called.
	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Decompress inflates a packet body, where size is the length the sender said
// it would inflate to.
//
// That size is the only bound there is on what a handful of bytes can expand
// into, so it is read as a limit rather than as a hint: a body that inflates to
// anything else is refused rather than kept, which also refuses a body built to
// expand until there is no memory left.
func Decompress(data []byte, size int32) ([]byte, error) {
	if size < 0 {
		return nil, fmt.Errorf("invalid uncompressed size: %d", size)
	}

	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	defer reader.Close()

	buf := make([]byte, size)
	if _, err := io.ReadFull(reader, buf); err != nil {
		return nil, fmt.Errorf("compressed body does not inflate to the %d bytes it claims: %w", size, err)
	}

	// Reaching the end is also what verifies the checksum, so the byte that is
	// expected to be missing has to be asked for.
	var extra [1]byte
	if _, err := io.ReadFull(reader, extra[:]); !errors.Is(err, io.EOF) {
		if err != nil {
			return nil, fmt.Errorf("compressed body is not readable to its end: %w", err)
		}

		return nil, fmt.Errorf("compressed body inflates to more than the %d bytes it claims", size)
	}

	return buf, nil
}
