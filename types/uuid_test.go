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

// TestOfflineUuidIsTheOneVanillaDerives pins the uuids to the values a vanilla
// server gives the same names, since a player who has one elsewhere expects to
// keep it here.
func TestOfflineUuidIsTheOneVanillaDerives(t *testing.T) {
	tests := []struct {
		username string
		want     string
	}{
		{username: "Notch", want: "b50ad385-829d-3141-a216-7e7d7539ba7f"},
		// The case of a name is part of what is hashed, so two spellings of the
		// same name are two accounts.
		{username: "notch", want: "42653081-a90e-3475-b3d6-3550cdb43f8e"},
	}

	for _, test := range tests {
		if got := OfflineUuid(test.username); got != test.want {
			t.Errorf("OfflineUuid(%q) = %q, want %q", test.username, got, test.want)
		}
	}
}

func TestOfflineUuidIsTheSameForTheSameName(t *testing.T) {
	// A name that came back as a different account on every connection would be
	// a player who is nobody, which is the one thing this uuid has to avoid.
	if first, second := OfflineUuid("Notch"), OfflineUuid("Notch"); first != second {
		t.Errorf("OfflineUuid(%q) = %q then %q, want the same account every time", "Notch", first, second)
	}

	if OfflineUuid("Notch") == OfflineUuid("Herobrine") {
		t.Error("two names share a uuid, want an account each")
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
