package streams

import (
	"bytes"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"sync"
)

// compressors and decompressors are reused across packets and connections. A
// fresh deflater carries its hash tables and window -- most of a megabyte --
// and a fresh inflater tens of kilobytes, which would otherwise be allocated
// again for every body that travels compressed.
var compressors = sync.Pool{
	New: func() any { return zlib.NewWriter(nil) },
}

var decompressors sync.Pool

// Compress deflates a packet body. Once a connection has been told a
// compression threshold, every body at or above it travels this way.
func Compress(data []byte) ([]byte, error) {
	buf := new(bytes.Buffer)

	writer := compressors.Get().(*zlib.Writer)
	defer compressors.Put(writer)

	writer.Reset(buf)

	if _, err := writer.Write(data); err != nil {
		writer.Close()

		return nil, err
	}

	// Close is what writes the trailing checksum, so a body is only whole once
	// it has been called. The writer stays reusable: Reset is what readies it
	// for the next body, closed or not.
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

	reader, err := newDecompressor(data)
	if err != nil {
		return nil, err
	}

	defer func() {
		reader.Close()
		decompressors.Put(reader)
	}()

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

// newDecompressor puts a pooled inflater on data, or builds the pool's first
// one. Both paths read the zlib header, so both can refuse data that does not
// start with one. A reader whose header was refused still goes back in the
// pool: the next Reset starts it over.
func newDecompressor(data []byte) (io.ReadCloser, error) {
	pooled := decompressors.Get()
	if pooled == nil {
		return zlib.NewReader(bytes.NewReader(data))
	}

	reader := pooled.(io.ReadCloser)
	if err := reader.(zlib.Resetter).Reset(bytes.NewReader(data), nil); err != nil {
		decompressors.Put(reader)

		return nil, err
	}

	return reader, nil
}
