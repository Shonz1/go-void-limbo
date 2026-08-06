package nbt

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

// NBT strings are encoded as Java's "modified UTF-8", which differs from the
// standard encoding in two places: U+0000 is written as the two-byte sequence
// C0 80 so that an encoded string never contains a zero byte, and code points
// above U+FFFF are written as a surrogate pair with each half taking its own
// three-byte sequence rather than as a single four-byte sequence.

// encodeModifiedUtf8 encodes value. Bytes that are not valid UTF-8 in the Go
// string are replaced with U+FFFD, matching how Go itself iterates strings.
func encodeModifiedUtf8(value string) ([]byte, error) {
	encoded := make([]byte, 0, len(value))

	for _, r := range value {
		switch {
		case r == 0:
			encoded = append(encoded, 0xC0, 0x80)
		case r < 0x80:
			encoded = append(encoded, byte(r))
		case r < 0x800:
			encoded = append(encoded, byte(0xC0|r>>6), byte(0x80|r&0x3F))
		case r <= 0xFFFF:
			encoded = appendCodeUnit(encoded, r)
		default:
			offset := r - 0x10000
			encoded = appendCodeUnit(encoded, 0xD800+(offset>>10))
			encoded = appendCodeUnit(encoded, 0xDC00+(offset&0x3FF))
		}
	}

	if len(encoded) > maxStringLength {
		return nil, fmt.Errorf("nbt: string encodes to %d bytes, limit is %d", len(encoded), maxStringLength)
	}

	return encoded, nil
}

// appendCodeUnit appends a code point in the range U+0800..U+FFFF, which is
// also the range every surrogate half falls in.
func appendCodeUnit(encoded []byte, r rune) []byte {
	return append(encoded, byte(0xE0|r>>12), byte(0x80|(r>>6)&0x3F), byte(0x80|r&0x3F))
}

// decodeModifiedUtf8 decodes encoded. A surrogate that is not part of a well
// formed pair is replaced with U+FFFD, since Go strings cannot represent one.
func decodeModifiedUtf8(encoded []byte) (string, error) {
	var sb strings.Builder
	sb.Grow(len(encoded))

	for i := 0; i < len(encoded); {
		r, size, err := decodeCodeUnit(encoded[i:])
		if err != nil {
			return "", err
		}

		i += size

		switch {
		case r >= 0xD800 && r <= 0xDBFF:
			low, lowSize, lowErr := decodeCodeUnit(encoded[i:])
			if lowErr != nil || low < 0xDC00 || low > 0xDFFF {
				sb.WriteRune(utf8.RuneError)
				continue
			}

			sb.WriteRune(0x10000 + (r-0xD800)<<10 + (low - 0xDC00))
			i += lowSize
		case r >= 0xDC00 && r <= 0xDFFF:
			sb.WriteRune(utf8.RuneError)
		default:
			sb.WriteRune(r)
		}
	}

	return sb.String(), nil
}

// decodeCodeUnit decodes the single UTF-16 code unit at the front of encoded
// and reports how many bytes it consumed.
func decodeCodeUnit(encoded []byte) (rune, int, error) {
	if len(encoded) == 0 {
		return 0, 0, errors.New("nbt: truncated modified utf-8 sequence")
	}

	switch {
	case encoded[0]&0x80 == 0:
		return rune(encoded[0]), 1, nil
	case encoded[0]&0xE0 == 0xC0:
		if len(encoded) < 2 || encoded[1]&0xC0 != 0x80 {
			return 0, 0, errors.New("nbt: malformed two-byte modified utf-8 sequence")
		}

		return rune(encoded[0]&0x1F)<<6 | rune(encoded[1]&0x3F), 2, nil
	case encoded[0]&0xF0 == 0xE0:
		if len(encoded) < 3 || encoded[1]&0xC0 != 0x80 || encoded[2]&0xC0 != 0x80 {
			return 0, 0, errors.New("nbt: malformed three-byte modified utf-8 sequence")
		}

		return rune(encoded[0]&0x0F)<<12 | rune(encoded[1]&0x3F)<<6 | rune(encoded[2]&0x3F), 3, nil
	}

	return 0, 0, fmt.Errorf("nbt: invalid modified utf-8 lead byte 0x%02X", encoded[0])
}
