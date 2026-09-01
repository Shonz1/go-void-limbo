package types

import (
	"fmt"
	"strings"
)

// GameMode is how a client is allowed to interact with the world. It lives
// here rather than with the packets that carry it because it is a fact about
// a player before it is a field of anything: the connection holds one, the
// operator configures one, and the packets only repeat it.
type GameMode int8

const (
	GameModeSurvival  GameMode = 0
	GameModeCreative  GameMode = 1
	GameModeAdventure GameMode = 2
	GameModeSpectator GameMode = 3

	// GameModeNone is the absent value, which the previous game mode takes when
	// the client has not been in one yet. It is not a mode the client can be
	// put into: the client reads it as null and leaves its previous mode unset.
	GameModeNone GameMode = -1
)

func (g GameMode) String() string {
	switch g {
	case GameModeSurvival:
		return "survival"
	case GameModeCreative:
		return "creative"
	case GameModeAdventure:
		return "adventure"
	case GameModeSpectator:
		return "spectator"
	case GameModeNone:
		return "none"
	default:
		return fmt.Sprintf("GameMode(%d)", int8(g))
	}
}

// ParseGameMode reads a game mode from the name a player would use for it,
// case aside. It reports false for anything else, including the absent value:
// none is a wire detail, not a mode anyone is put in.
func ParseGameMode(name string) (GameMode, bool) {
	switch strings.ToLower(name) {
	case "survival":
		return GameModeSurvival, true
	case "creative":
		return GameModeCreative, true
	case "adventure":
		return GameModeAdventure, true
	case "spectator":
		return GameModeSpectator, true
	}

	return GameModeNone, false
}
