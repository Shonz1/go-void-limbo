package config

import (
	"os"
	"testing"
)

func TestEncryptionEnabled(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  bool
	}{
		// Nothing said is a server that encrypts, which is the only default
		// worth having.
		{name: "unset", set: false, want: true},
		{name: "true", value: "true", set: true, want: true},
		{name: "false", value: "false", set: true, want: false},
		{name: "1", value: "1", set: true, want: true},
		{name: "0", value: "0", set: true, want: false},
		// A value nobody can read is not a reason to stop encrypting.
		{name: "nonsense", value: "yes please", set: true, want: true},
		{name: "empty", value: "", set: true, want: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// t.Setenv is what puts back whatever the environment had; the
			// unset the test is after comes on top of it.
			t.Setenv("ENCRYPTION", test.value)

			if !test.set {
				os.Unsetenv("ENCRYPTION")
			}

			if got := EncryptionEnabled(); got != test.want {
				t.Errorf("EncryptionEnabled() = %t, want %t", got, test.want)
			}
		})
	}
}

func TestDescription(t *testing.T) {
	tests := []struct {
		name  string
		value string
		set   bool
		want  string
	}{
		// Nothing said is a server that says what it is, since a player looking
		// at a list of them has nothing else to go on.
		{name: "unset", set: false, want: defaultDescription},
		{name: "set", value: "somewhere to wait", set: true, want: "somewhere to wait"},
		// A blank line in a server list reads as a server that failed to answer.
		{name: "empty", value: "", set: true, want: defaultDescription},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("MOTD", test.value)

			if !test.set {
				os.Unsetenv("MOTD")
			}

			if got := Description(); got != test.want {
				t.Errorf("Description() = %q, want %q", got, test.want)
			}
		})
	}
}

// The secret is the setting: a server that holds one asks every login for a
// payload signed with it, and a server that holds none never asks. So an empty
// value is no secret rather than an empty one, whichever of the two places it
// came from.
func TestForwardingSecret(t *testing.T) {
	tests := []struct {
		name        string
		argument    string
		environment string
		set         bool
		want        string
	}{
		{name: "neither", want: ""},
		{name: "the environment alone", environment: "from the environment", set: true, want: "from the environment"},
		{name: "the argument alone", argument: "from the argument", want: "from the argument"},
		// A flag is typed and an environment is inherited, so the typed one is
		// the one that was meant.
		{name: "both", argument: "from the argument", environment: "from the environment", set: true, want: "from the argument"},
		// A secret nobody set is a secret everybody has, and checking a
		// signature under one is worse than not asking for a signature at all.
		{name: "an empty environment", environment: "", set: true, want: ""},
		{name: "an empty argument", argument: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("FORWARDING_SECRET", test.environment)

			if !test.set {
				os.Unsetenv("FORWARDING_SECRET")
			}

			got := ForwardingSecret(test.argument)

			if string(got) != test.want {
				t.Errorf("ForwardingSecret(%q) = %q, want %q", test.argument, got, test.want)
			}

			// Nothing is what an unconfigured server checks against, and it is
			// what the connection reads as having no proxy in front of it.
			if test.want == "" && got != nil {
				t.Errorf("ForwardingSecret(%q) = %q, want no secret at all", test.argument, got)
			}
		})
	}
}
