package handlers

import (
	"fmt"
	clientboundPlay "go-void-limbo/packets/clientbound/play"
	"go-void-limbo/packets/serverbound/configuration"
	"go-void-limbo/types"
)

// What the limbo puts a joining client into. The world itself is nothing: one
// dimension, no terrain, and a player floating in it.
const (
	// overworldDimension is the only world this server has. The name is the
	// client's own, because a client refuses a dimension it has no id for and
	// the ids it has are the ones package gamedata sent during configuration.
	overworldDimension = "minecraft:overworld"

	// overworldDimensionTypeId is the index of minecraft:overworld in the
	// minecraft:dimension_type registry, where package gamedata puts it first
	// and alone. A wrong index here reads as a different dimension type, or as
	// none at all, and the client leaves rather than guess.
	overworldDimensionTypeId = 0

	// playerEntityId is the id of the joining player's own entity. Anything the
	// play phase later says about that player names this number, and with one
	// player per connection there is nothing for it to collide with.
	playerEntityId = 1

	// viewDistance and simulationDistance are the smallest the client will
	// hold, since a limbo that sends no chunks has nothing to fill a larger
	// cache with. The client raises anything below two to two anyway.
	viewDistance       = 2
	simulationDistance = 2

	// spawnTeleportId identifies the teleport that places a joining player, so
	// its acknowledgement can be told from later ones.
	spawnTeleportId = 1
)

// spawnGameMode is sent twice, in the login packet and again in the player
// list entry the client keeps about itself, and the client reads its own mode
// from both. They have to agree.
const spawnGameMode = clientboundPlay.GameModeSpectator

// Where a joining player is put. Any position inside the dimension's vertical
// bounds works, since there is nothing to stand on or fall through.
const (
	spawnX = 0.5
	spawnY = 64.0
	spawnZ = 0.5
)

// HandleAcknowledgeFinishConfigurationServerboundPacket moves a client that is
// done configuring into play and sends it the join.
//
// Four packets are what a join needs, and this sends exactly those. What a
// vanilla server also sends -- difficulty, abilities, held slot, recipes,
// world border, time, spawn position -- is either a value the client already
// defaults to or a fact about a world this one does not have. Keep alives are
// what a joined client needs after this, and they are not sent from here: they
// belong to a clock rather than to a packet that arrived.
func HandleAcknowledgeFinishConfigurationServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	_, ok := packet.(*configuration.AcknowledgeFinishConfigurationServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *configuration.AcknowledgeFinishConfigurationServerboundPacket, got %T", packet)
	}

	// The phase has to move first: WritePacket resolves a packet id from the
	// phase the client is currently in, and none of what follows is registered
	// anywhere but play.
	client.SetPhase(types.PhasePlay)

	login := clientboundPlay.LoginClientboundPacket{
		EntityId:   playerEntityId,
		Dimensions: []string{overworldDimension},
		// The client only shows this on the player list, and shows nothing when
		// it is zero.
		MaxPlayers:         0,
		ViewDistance:       viewDistance,
		SimulationDistance: simulationDistance,
		ShowDeathScreen:    true,
		SpawnInfo: clientboundPlay.SpawnInfo{
			DimensionTypeId: overworldDimensionTypeId,
			Dimension:       overworldDimension,
			// The spectator mode is what lets a player into a world with no
			// chunks in it: a client that cannot be standing in a chunk does
			// not wait for one to render, so the loading screen closes as soon
			// as the join is through instead of after its thirty second
			// timeout. It also settles what a player does in a world with
			// nothing to stand on, which is float.
			GameMode:         spawnGameMode,
			PreviousGameMode: clientboundPlay.GameModeNone,
		},
	}

	if err := client.WritePacket(&login); err != nil {
		return err
	}

	// The login packet says nothing about where the player is, so a teleport is
	// what puts it at the spawn. The client replies acknowledging this id.
	position := clientboundPlay.PlayerPositionClientboundPacket{
		TeleportId: spawnTeleportId,
		X:          spawnX,
		Y:          spawnY,
		Z:          spawnZ,
	}

	if err := client.WritePacket(&position); err != nil {
		return err
	}

	// The client holds a player list entry about itself, which is where it
	// reads its own name, skin and game mode from, and which it has none of
	// until it is told. Only one player exists here, so the entry is its own.
	//
	// The five actions left out are the five whose values would be the ones a
	// new entry already holds: no chat session, no display name of its own, no
	// measured latency, no place of its own in the list order, and a hat.
	playerInfo := clientboundPlay.PlayerInfoUpdateClientboundPacket{
		Actions: clientboundPlay.PlayerInfoAddPlayer | clientboundPlay.PlayerInfoUpdateGameMode | clientboundPlay.PlayerInfoUpdateListed,
		Entries: []clientboundPlay.PlayerInfoEntry{
			{
				Profile:  client.Profile(),
				GameMode: spawnGameMode,
				// An entry the client is not told to list is one it keeps but
				// does not draw, and a player that cannot see itself on the
				// list is the one thing about a limbo a player checks.
				Listed: true,
			},
		},
	}

	if err := client.WritePacket(&playerInfo); err != nil {
		return err
	}

	// Last, because it means the join is over and chunks are next. The client
	// waits on its loading screen with no timeout until it arrives.
	chunksNext := clientboundPlay.GameEventClientboundPacket{Event: clientboundPlay.GameEventStartWaitingForChunks}

	return client.WritePacket(&chunksNext)
}
