package types

import (
	"crypto/rand"
	"fmt"
)

// NewRandomUuid returns a random (version 4) UUID in the hyphenated form the
// rest of the codebase passes around.
func NewRandomUuid() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	buf[6] = (buf[6] & 0x0F) | 0x40
	buf[8] = (buf[8] & 0x3F) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16]), nil
}
