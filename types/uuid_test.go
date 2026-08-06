package types

import (
	"regexp"
	"testing"
)

var uuidV4Pattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

func TestNewRandomUuidIsVersion4(t *testing.T) {
	uuid, err := NewRandomUuid()
	if err != nil {
		t.Fatalf("NewRandomUuid() error = %v", err)
	}

	if !uuidV4Pattern.MatchString(uuid) {
		t.Errorf("NewRandomUuid() = %q, want a hyphenated version 4 uuid", uuid)
	}
}

func TestNewRandomUuidIsRandom(t *testing.T) {
	first, err := NewRandomUuid()
	if err != nil {
		t.Fatalf("NewRandomUuid() error = %v", err)
	}

	second, err := NewRandomUuid()
	if err != nil {
		t.Fatalf("NewRandomUuid() error = %v", err)
	}

	if first == second {
		t.Errorf("NewRandomUuid() returned %q twice, want distinct uuids", first)
	}
}
