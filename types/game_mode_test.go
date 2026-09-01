package types

import "testing"

func TestParseGameMode(t *testing.T) {
	tests := []struct {
		name string
		want GameMode
		ok   bool
	}{
		{name: "survival", want: GameModeSurvival, ok: true},
		{name: "creative", want: GameModeCreative, ok: true},
		{name: "adventure", want: GameModeAdventure, ok: true},
		{name: "spectator", want: GameModeSpectator, ok: true},
		{name: "Spectator", want: GameModeSpectator, ok: true},
		// The absent value is a wire detail, not a mode anyone is put in.
		{name: "none", want: GameModeNone},
		{name: "hardcore", want: GameModeNone},
		{name: "", want: GameModeNone},
	}

	for _, test := range tests {
		got, ok := ParseGameMode(test.name)
		if got != test.want || ok != test.ok {
			t.Errorf("ParseGameMode(%q) = %s, %t, want %s, %t", test.name, got, ok, test.want, test.ok)
		}
	}
}

// Every mode parses back from its own name, so the two stay one table.
func TestGameModeStringRoundTrips(t *testing.T) {
	for _, mode := range []GameMode{GameModeSurvival, GameModeCreative, GameModeAdventure, GameModeSpectator} {
		parsed, ok := ParseGameMode(mode.String())
		if !ok || parsed != mode {
			t.Errorf("ParseGameMode(%q) = %s, %t, want the mode itself back", mode.String(), parsed, ok)
		}
	}
}
