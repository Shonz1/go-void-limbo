// Package anvil reads the world format Minecraft saves: region files of
// chunks, and the level.dat beside them. It reads names -- block names,
// property names, heightmap names -- and leaves what the network calls those
// things to whoever is sending them, which is what lets one saved world serve
// every protocol version.
package anvil

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ErrChunkNotFound reports a chunk the world never generated: a region file
// that does not exist, or an empty slot in one that does. It is the one read
// failure that is not damage, since a world is under no obligation to be any
// particular size.
var ErrChunkNotFound = errors.New("anvil: chunk not generated")

// A region file holds 32x32 chunks. The header is two 4KiB tables of one entry
// per chunk: where the chunk sits in the file, and when it was written. Only
// the first is read here.
const (
	regionChunkBits = 5
	regionChunkSpan = 1 << regionChunkBits

	sectorSize = 4096
)

// Compression schemes a chunk payload announces. LZ4 and custom schemes exist
// in newer saves as opt-in settings; a world using them was configured to, and
// gets an error naming the scheme rather than a guess.
const (
	compressionGzip = 1
	compressionZlib = 2
	compressionNone = 3
)

// ReadChunkPayload returns the decompressed NBT of one chunk from the region
// directory, or ErrChunkNotFound for a chunk the world never generated.
// Coordinates are in chunks, not regions; the region file and the slot in it
// are worked out here.
func ReadChunkPayload(regionDir string, x, z int32) ([]byte, error) {
	name := fmt.Sprintf("r.%d.%d.mca", x>>regionChunkBits, z>>regionChunkBits)

	file, err := os.Open(filepath.Join(regionDir, name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrChunkNotFound
		}

		return nil, fmt.Errorf("anvil: %w", err)
	}
	defer file.Close()

	// The location entry: three bytes of sector offset, one of sector count.
	// Both zero means the slot is empty.
	slot := (x & (regionChunkSpan - 1)) + (z&(regionChunkSpan-1))*regionChunkSpan

	var location [4]byte
	if _, err := file.ReadAt(location[:], int64(slot)*4); err != nil {
		return nil, fmt.Errorf("anvil: reading %s location table: %w", name, err)
	}

	offset := int64(location[0])<<16 | int64(location[1])<<8 | int64(location[2])
	sectors := int64(location[3])

	if offset == 0 && sectors == 0 {
		return nil, ErrChunkNotFound
	}

	// The chunk starts with its own length -- of the compression byte and the
	// compressed payload after it -- which is what to trust over the sector
	// count, since sectors only say where the next chunk may begin.
	header := make([]byte, 5)
	if _, err := file.ReadAt(header, offset*sectorSize); err != nil {
		return nil, fmt.Errorf("anvil: reading %s chunk header: %w", name, err)
	}

	length := int64(binary.BigEndian.Uint32(header))
	if length < 1 || length > sectors*sectorSize {
		return nil, fmt.Errorf("anvil: %s claims a %d byte chunk in %d sectors", name, length, sectors)
	}

	compressed := make([]byte, length-1)
	if _, err := file.ReadAt(compressed, offset*sectorSize+5); err != nil {
		return nil, fmt.Errorf("anvil: reading %s chunk: %w", name, err)
	}

	return decompress(header[4], compressed)
}

func decompress(scheme byte, compressed []byte) ([]byte, error) {
	switch scheme {
	case compressionGzip:
		reader, err := gzip.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, fmt.Errorf("anvil: %w", err)
		}
		defer reader.Close()

		payload, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("anvil: %w", err)
		}

		return payload, nil
	case compressionZlib:
		reader, err := zlib.NewReader(bytes.NewReader(compressed))
		if err != nil {
			return nil, fmt.Errorf("anvil: %w", err)
		}
		defer reader.Close()

		payload, err := io.ReadAll(reader)
		if err != nil {
			return nil, fmt.Errorf("anvil: %w", err)
		}

		return payload, nil
	case compressionNone:
		return compressed, nil
	}

	return nil, fmt.Errorf("anvil: unsupported compression scheme %d", scheme)
}
