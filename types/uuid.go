package types

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
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

// offlineUuidPrefix is what a username is hashed under to get the uuid of an
// account nobody vouched for. Vanilla hashes the same string, so a player keeps
// the uuid they have on any other server that never asked Mojang about them.
const offlineUuidPrefix = "OfflinePlayer:"

// OfflineUuid returns the uuid a login that was never authenticated goes by: a
// version 3 (name-based) uuid over the username.
//
// It is derived rather than taken from what the client claimed because the same
// name has to mean the same account every time it connects, and a client that
// picks its own uuid is a client that can be anyone twice over. MD5 is what a
// version 3 uuid is built from and what vanilla derives this one with; nothing
// here rests on it being hard to reverse, since an unauthenticated name is worth
// nothing to begin with.
func OfflineUuid(username string) string {
	sum := md5.Sum([]byte(offlineUuidPrefix + username))

	sum[6] = (sum[6] & 0x0F) | 0x30
	sum[8] = (sum[8] & 0x3F) | 0x80

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		sum[0:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16])
}

// UuidFromHex turns the thirty-two characters the session server writes a uuid
// as into the hyphenated form the rest of the codebase passes around.
func UuidFromHex(value string) (string, error) {
	buf, err := hex.DecodeString(value)
	if err != nil {
		return "", fmt.Errorf("invalid uuid %q: %w", value, err)
	}

	if len(buf) != 16 {
		return "", fmt.Errorf("invalid uuid %q: expected 16 bytes, got %d", value, len(buf))
	}

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16]), nil
}
