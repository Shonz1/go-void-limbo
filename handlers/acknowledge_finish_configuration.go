package handlers

import (
	"fmt"

	clientboundPlay "github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/packets/serverbound/configuration"
	"github.com/Shonz1/go-void-limbo/types"
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

	// viewDistance is how far the client is told it will be shown, in chunks.
	// Package world sends one ring more than this around the spawn, because
	// the client only renders a chunk whose neighbours it also holds; the two
	// numbers move together. A server with no world fills none of the cache
	// this asks the client to hold, which costs nothing.
	//
	// simulationDistance stays the smallest the client accepts, since nothing
	// here ticks either way. The client raises anything below two to two.
	viewDistance       = 8
	simulationDistance = 2

	// spawnTeleportId identifies the teleport that places a joining player, so
	// its acknowledgement can be told from later ones.
	spawnTeleportId = 1
)

// Where a joining player is put when no world says otherwise. Any position
// inside the dimension's vertical bounds works, since with no world there is
// nothing to stand on or fall through. A server with a world uses the spawn
// the world's own level.dat names instead.
const (
	spawnX = 0.5
	spawnY = 64.0
	spawnZ = 0.5
)

// HandleAcknowledgeFinishConfigurationServerboundPacket moves a client that is
// done configuring into play and sends it the join, and after the join the
// world, when the server holds one.
func HandleAcknowledgeFinishConfigurationServerboundPacket(client types.Client, packet types.ServerboundPacket) error {
	_, ok := packet.(*configuration.AcknowledgeFinishConfigurationServerboundPacket)
	if !ok {
		return fmt.Errorf("expected *configuration.AcknowledgeFinishConfigurationServerboundPacket, got %T", packet)
	}

	return enterPlay(client)
}

// enterPlay moves a client into play and sends it the join, and after the
// join the world, when the server holds one. It is the one way into the world
// on every version: a client with a configuration phase arrives here from its
// acknowledgement that the phase is over, and a client without one from its
// login success, since nothing separates the two on such a version.
//
// Four packets are what the join itself needs, and this sends exactly those
// before the world -- five on a version with no configuration phase, which
// reads the tags in play. What a vanilla server also sends -- difficulty,
// abilities, held slot, recipes, world border, time, spawn position -- is
// either a value the client already defaults to or a fact this server does
// not track. Keep alives are what a joined client needs after this, and they
// are not sent from here: they belong to a clock rather than to a packet that
// arrived.
func enterPlay(client types.Client) error {
	// The phase has to move first: WritePacket resolves a packet id from the
	// phase the client is currently in, and none of what follows is registered
	// anywhere but play.
	client.SetPhase(types.PhasePlay)

	_, forwarded := client.ForwardedLogin()

	// The entity id is the connection's own, unique among every connection the
	// server accepts: anything the play phase says about this player names it,
	// including what the other players are told.
	login := clientboundPlay.LoginClientboundPacket{
		EntityId:   client.EntityId(),
		Dimensions: []string{overworldDimension},
		// The client only shows this on the player list, and shows nothing when
		// it is zero.
		MaxPlayers:         0,
		ViewDistance:       viewDistance,
		SimulationDistance: simulationDistance,
		ShowDeathScreen:    true,
		// Whether the client was let in on Mojang's word rather than on its own.
		// Either the connection was encrypted and this end asked, or a proxy
		// asked before forwarding the answer; a login that is neither is one
		// nobody was ever asked about.
		//
		// The player list draws a head beside a name only for a server that says
		// yes, because a server that takes a client's word for who it is is a
		// server where the name and the skin beside it prove nothing. Saying so
		// is the only way the heads appear: the signed textures in the player
		// list entry are not enough on their own. A server that is neither
		// asking Mojang nor being told has no textures to put there anyway, and
		// says no.
		OnlineMode: client.EncryptionEnabled() || forwarded,
		SpawnInfo: clientboundPlay.SpawnInfo{
			DimensionTypeId: overworldDimensionTypeId,
			Dimension:       overworldDimension,
			// The mode is the connection's, which the operator chose for the
			// server; the same value goes into the player list entry below,
			// because the client reads its own mode from both and they have to
			// agree. One thing to know when choosing it: only a spectator
			// skips the wait for the chunk it stands in to render, so on a
			// server with no world any other mode leaves a joining client on
			// its loading screen for that wait's thirty second timeout.
			GameMode:         client.GameMode(),
			PreviousGameMode: types.GameModeNone,
		},
	}

	if err := client.WritePacket(&login); err != nil {
		return err
	}

	// A version with no configuration phase reads in play what the phase
	// carries from 1.20.2 on. The registries went inside the login packet
	// above, which is package gamedata's and the transformers' doing; what is
	// left is the tags, which such a version reads as a play packet right
	// after the login, where a vanilla server of that version sends them.
	if !client.ProtocolVersion().HasConfigurationPhase() {
		for _, packet := range client.RegistryPackets() {
			if err := client.WritePacket(packet); err != nil {
				return err
			}
		}
	}

	// The login packet says nothing about where the player is, so a teleport is
	// what puts it at the spawn: the world's own when there is one, and a spot
	// as good as any other when there is not. The client replies acknowledging
	// this id.
	x, y, z, hasWorld := client.WorldSpawn()
	if !hasWorld {
		x, y, z = spawnX, spawnY, spawnZ
	}

	position := clientboundPlay.PlayerPositionClientboundPacket{
		TeleportId: spawnTeleportId,
		X:          x,
		Y:          y,
		Z:          z,
	}

	if err := client.WritePacket(&position); err != nil {
		return err
	}

	// The teleport is also where this player is until its own move packets say
	// otherwise, and the sync state has to agree before the join below shows
	// it to anyone. Nothing is relayed by this: the player is not on the
	// roster yet, and nobody has been shown an entity to apply it to.
	client.SyncPositionRotation(x, y, z, 0, 0, false)

	// The client holds a player list entry about itself, which is where it
	// reads its own name, skin and game mode from, and which it has none of
	// until it is told. Only one player exists here, so the entry is its own.
	//
	// The four actions left out are the four whose values would be the ones a
	// new entry already holds: no chat session, no display name of its own, no
	// measured latency, and no place of its own in the list order.
	playerInfo := clientboundPlay.PlayerInfoUpdateClientboundPacket{
		Actions: clientboundPlay.PlayerInfoAddPlayer | clientboundPlay.PlayerInfoUpdateGameMode |
			clientboundPlay.PlayerInfoUpdateListed | clientboundPlay.PlayerInfoUpdateHat,
		Entries: []clientboundPlay.PlayerInfoEntry{
			{
				Profile:  client.Profile(),
				GameMode: client.GameMode(),
				// An entry the client is not told to list is one it keeps but
				// does not draw, and a player that cannot see itself on the
				// list is the one thing about a limbo a player checks.
				Listed: true,
				// A new entry holds no hat, and the head the player list draws
				// is the skin's two head layers over each other. Leaving this
				// out draws the base layer alone, which for a skin that carries
				// its hair on the second one is a head that looks missing.
				//
				// This is true rather than what the player asked for because
				// what they asked for arrives in a client information packet
				// this server does not read, and a hat is what all but a few
				// players have on.
				ShowHat: true,
			},
		},
	}

	if err := client.WritePacket(&playerInfo); err != nil {
		return err
	}

	// Last of the join, because it means the join is over and chunks are next.
	// The client waits on its loading screen with no timeout until it arrives.
	chunksNext := clientboundPlay.GameEventClientboundPacket{Event: clientboundPlay.GameEventStartWaitingForChunks}

	if err := client.WritePacket(&chunksNext); err != nil {
		return err
	}

	// And then the chunks themselves, prebuilt for this client's version:
	// nothing on a server with no world, and the chunk cache centre followed
	// by every chunk around the spawn on one that has it.
	for _, packet := range client.WorldPackets() {
		if err := client.WritePacket(packet); err != nil {
			return err
		}
	}

	// Last of all, the other players -- and this player to them. The client
	// accepts entities as soon as the join is through, so nothing has to wait
	// for it to acknowledge the teleport above.
	client.JoinPlayerSync()

	return nil
}
