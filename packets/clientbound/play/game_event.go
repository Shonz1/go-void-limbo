package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// GameEvent is what a game event packet announces. Only the one a limbo needs
// is named here; the rest are weather, demo prompts and combat feedback.
type GameEvent byte

// GameEventStartWaitingForChunks tells the client the server is done sending
// what a join needs and chunks are next.
//
// A joining client sits on its loading screen until this arrives -- that wait
// has no timeout, so leaving the event out strands the client there for as
// long as it stays connected. What it starts waiting for afterwards does have
// one: the client watches for the chunk it is standing in to render, gives up
// after thirty seconds, and lets the player in regardless. A client that
// cannot be standing in a chunk skips the wait entirely, which is why the
// spectator game mode the join sends puts a player into an empty world at once.
const GameEventStartWaitingForChunks GameEvent = 13

func (e GameEvent) String() string {
	if e == GameEventStartWaitingForChunks {
		return "start_waiting_for_chunks"
	}

	return fmt.Sprintf("GameEvent(%d)", byte(e))
}

// GameEventClientboundPacket carries one game event and the single number that
// parameterises it.
type GameEventClientboundPacket struct {
	Event GameEvent

	// Value means something different for every event, and nothing at all for
	// most of them.
	Value float32
}

func (p *GameEventClientboundPacket) String() string {
	return fmt.Sprintf("GameEventClientboundPacket{Event:%s Value:%g}", p.Event, p.Value)
}

func (p *GameEventClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteByte(byte(p.Event)); err != nil {
		return err
	}

	return ms.WriteFloat(p.Value)
}
