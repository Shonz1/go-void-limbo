package types

// ServerStatus is what a server list ping is answered with: everything a client
// draws beside an address it has not connected to, and the version it compares
// its own against to decide whether it can.
//
// It travels as JSON rather than as fields of a packet, so the tags are not a
// convention this end chose: they are the names the client reads, and renaming
// one is the same as leaving it out.
type ServerStatus struct {
	Version     ServerVersion `json:"version"`
	Players     ServerPlayers `json:"players"`
	Description TextComponent `json:"description"`
}

// ServerVersion is what a client decides compatibility on. Protocol is the
// whole of that decision -- a client refuses a number that is not its own -- and
// Name is only what it draws beside the refusal, which is why a server whose
// version a client cannot speak still says something a player can read.
type ServerVersion struct {
	Name     string     `json:"name"`
	Protocol ProtocolId `json:"protocol"`
}

// ServerPlayers is the count drawn to the right of the address. Both numbers are
// the server's own word: nothing checks them, and a client draws a server as
// full when they are equal.
//
// The sample of names a client shows on hovering the count is left out. It is a
// list a server chooses to advertise about the players on it, and a limbo has
// nothing to advertise about players who are only waiting.
type ServerPlayers struct {
	Max    int32 `json:"max"`
	Online int32 `json:"online"`
}

// TextComponent is a piece of text the client renders, in the one form this
// server ever needs: the text itself and nothing about how it is drawn.
type TextComponent struct {
	Text string `json:"text"`
}
