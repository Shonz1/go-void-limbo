// Player synchronization: how one connection's player reaches the others.
//
// The roster below is the one piece of it the connections share -- who is in
// the world right now. Everything else lives on the connections themselves:
// each one tracks where its own player is from the move packets it reads, and
// each one remembers which other players it has been shown, so that a spawn, a
// removal and the relays in between are decided against what actually went out
// on that wire rather than against a global picture that may have moved on.
package client

import (
	"log/slog"
	"sync"
	"time"

	clientboundPlay "github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/types"
)

// syncWriteTimeout bounds every write one player's sync performs on another
// player's connection, for the same reason the keep alive sweep bounds its
// own: the writer here is somebody else's goroutine -- often the reactor
// serving every joined connection -- and a recipient that stopped draining its
// side must cost it a bounded wait, not hold everyone else's world behind it.
const syncWriteTimeout = 5 * time.Second

// PlayerSync is the roster of joined players, shared by every connection of
// one server the way the status and the world are. It holds membership and
// nothing else: what gets sent to whom is decided per recipient, by the
// recipient's own record of who it has been shown.
type PlayerSync struct {
	mu      sync.Mutex
	players map[*Client]struct{}
}

func NewPlayerSync() *PlayerSync {
	return &PlayerSync{players: make(map[*Client]struct{})}
}

func (ps *PlayerSync) add(c *Client) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.players[c] = struct{}{}
}

// remove takes a connection off the roster and reports whether it was on it,
// which is what makes leaving idempotent: only the removal that found it there
// broadcasts the departure.
func (ps *PlayerSync) remove(c *Client) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, ok := ps.players[c]
	delete(ps.players, c)

	return ok
}

func (ps *PlayerSync) contains(c *Client) bool {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	_, ok := ps.players[c]

	return ok
}

// others snapshots everyone on the roster but c, so the sends that follow
// happen off the lock: a recipient that has stopped draining its connection
// stalls the one write aimed at it, not the roster.
func (ps *PlayerSync) others(c *Client) []*Client {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	others := make([]*Client, 0, len(ps.players))
	for player := range ps.players {
		if player != c {
			others = append(others, player)
		}
	}

	return others
}

// playerSnapshot is one player's visible state, read in one go under its
// connection's lock and carried to the other connections as plain values, so
// showing a player never holds two connections' locks at once.
type playerSnapshot struct {
	entityId int32
	profile  types.GameProfile
	gameMode types.GameMode

	x, y, z    float64
	yaw, pitch float32
	onGround   bool

	sneaking  bool
	sprinting bool
}

// EntityId is the id everything in the play phase names this player's entity
// by. It is assigned when the connection is accepted and never changes, so it
// is read without the lock.
func (c *Client) EntityId() int32 {
	return c.entityId
}

// GameMode is the mode this player is in, fixed like the entity id: the
// operator chose it for the server, and nothing here changes a player's mode
// after it joined.
func (c *Client) GameMode() types.GameMode {
	return c.gameMode
}

// JoinPlayerSync puts this player among the others: everyone already on the
// roster appears on this connection, and this player appears on theirs.
//
// The order here closes the join/leave races. This connection goes onto the
// roster before anything is sent, so a player leaving from now on includes it
// in the departure broadcast. Each existing player is then shown, and checked
// against the roster afterwards: one that left mid-join had its departure
// broadcast to a connection that may not have been shown it yet, and the
// recheck is what takes that ghost back down. Both sides of that race go
// through showPlayer and hidePlayer, which decide under the recipient's lock
// against what was actually sent, so however the sends interleave, the
// recipient's world ends up matching its record of it.
func (c *Client) JoinPlayerSync() {
	ps := c.playerSync
	if ps == nil {
		return
	}

	ps.add(c)

	self := c.syncSnapshot()

	for _, other := range ps.others(c) {
		otherPlayer := other.syncSnapshot()

		c.showPlayer(otherPlayer)
		other.showPlayer(self)

		if !ps.contains(other) {
			c.hidePlayer(otherPlayer.entityId, otherPlayer.profile.Uuid)
		}
	}
}

// leavePlayerSync takes this player off the roster and out of everyone else's
// world. It is called by LeavePlay, whichever loop saw the connection end, and
// only the call that found the player on the roster does the broadcasting.
func (c *Client) leavePlayerSync() {
	ps := c.playerSync
	if ps == nil || !ps.remove(c) {
		return
	}

	entityId := c.EntityId()
	uuid := c.Profile().Uuid

	for _, other := range ps.others(c) {
		other.hidePlayer(entityId, uuid)
	}
}

// SyncPosition records where this player's move packet put it and relays the
// move. The client sends this variant when it moved without turning, so the
// rotation relayed is the one it last reported.
func (c *Client) SyncPosition(x, y, z float64, onGround bool) {
	c.mu.Lock()
	c.x, c.y, c.z, c.onGround = x, y, z, onGround
	self := c.snapshotLocked()
	c.mu.Unlock()

	c.relayMove(self, false)
}

// SyncPositionRotation records a move packet that carried both, and relays
// them both.
func (c *Client) SyncPositionRotation(x, y, z float64, yaw, pitch float32, onGround bool) {
	c.mu.Lock()
	c.x, c.y, c.z, c.onGround = x, y, z, onGround
	c.yaw, c.pitch = yaw, pitch
	self := c.snapshotLocked()
	c.mu.Unlock()

	c.relayMove(self, true)
}

// SyncRotation records a player that turned on the spot, and relays the turn.
func (c *Client) SyncRotation(yaw, pitch float32, onGround bool) {
	c.mu.Lock()
	c.yaw, c.pitch, c.onGround = yaw, pitch, onGround
	self := c.snapshotLocked()
	c.mu.Unlock()

	c.relayMove(self, true)
}

// SyncGround records the one flag the standing-still move packet carries.
// Nothing is relayed: the flag only qualifies positions, and the next position
// will carry it.
func (c *Client) SyncGround(onGround bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.onGround = onGround
}

// SyncSwing plays this player's arm swing on everyone else's view of it.
func (c *Client) SyncSwing(offHand bool) {
	ps := c.playerSync
	if ps == nil {
		return
	}

	animation := clientboundPlay.AnimationSwingMainArm
	if offHand {
		animation = clientboundPlay.AnimationSwingOffhand
	}

	swing := &clientboundPlay.AnimateClientboundPacket{
		EntityId:  c.EntityId(),
		Animation: animation,
	}

	for _, other := range ps.others(c) {
		other.relayToKnown(c.EntityId(), swing)
	}
}

// SyncInput records the movement keys this player is holding and, when the
// two stances other players can see changed, shows them the change.
func (c *Client) SyncInput(sneaking, sprinting bool) {
	ps := c.playerSync
	if ps == nil {
		return
	}

	c.mu.Lock()
	changed := c.sneaking != sneaking || c.sprinting != sprinting
	c.sneaking, c.sprinting = sneaking, sprinting
	c.mu.Unlock()

	// The client resends its input byte whenever any key changes, and most of
	// the bits are movement keys nobody else can see.
	if !changed {
		return
	}

	stance := &clientboundPlay.SetEntityDataClientboundPacket{
		EntityId:  c.EntityId(),
		Sneaking:  sneaking,
		Sprinting: sprinting,
	}

	for _, other := range ps.others(c) {
		other.relayToKnown(c.EntityId(), stance)
	}
}

// relayMove sends this player's current position to everyone who has it in
// their world. The head rotation goes along whenever the move carried a
// rotation, because the position packet's angles steer the body while the
// head -- the part of a player everyone actually watches -- only turns for
// the head packet.
func (c *Client) relayMove(self playerSnapshot, rotated bool) {
	ps := c.playerSync
	if ps == nil {
		return
	}

	packets := make([]types.ClientboundPacket, 0, 2)
	packets = append(packets, &clientboundPlay.EntityPositionSyncClientboundPacket{
		EntityId: self.entityId,
		X:        self.x,
		Y:        self.y,
		Z:        self.z,
		Yaw:      self.yaw,
		Pitch:    self.pitch,
		OnGround: self.onGround,
	})

	if rotated {
		packets = append(packets, &clientboundPlay.RotateHeadClientboundPacket{
			EntityId: self.entityId,
			HeadYaw:  self.yaw,
		})
	}

	for _, other := range ps.others(c) {
		other.relayToKnown(self.entityId, packets...)
	}
}

func (c *Client) syncSnapshot() playerSnapshot {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.snapshotLocked()
}

func (c *Client) snapshotLocked() playerSnapshot {
	return playerSnapshot{
		entityId:  c.entityId,
		profile:   c.profile,
		gameMode:  c.gameMode,
		x:         c.x,
		y:         c.y,
		z:         c.z,
		yaw:       c.yaw,
		pitch:     c.pitch,
		onGround:  c.onGround,
		sneaking:  c.sneaking,
		sprinting: c.sprinting,
	}
}

// showPlayer puts one player into this connection's world, unless it is
// already there. The check and the sends happen under the one lock, so what
// this connection was shown and what it remembers being shown cannot come
// apart, however many joins and leaves are in flight.
//
// The spawn is three packets, in the one order the client accepts them: the
// list entry first, because a player entity whose uuid has no entry is refused
// rather than spawned; the entity itself; and the stance, only when there is
// one to show. A write that fails here is this connection dying, which its own
// loops notice, so it is logged and otherwise left alone.
func (c *Client) showPlayer(player playerSnapshot) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.shownPlayers == nil {
		c.shownPlayers = make(map[int32]struct{})
	}

	if _, ok := c.shownPlayers[player.entityId]; ok {
		return
	}

	c.shownPlayers[player.entityId] = struct{}{}

	// The entry mirrors the one a client is sent about itself when it joins,
	// hat and all, so a player looks the same on its own screen and on
	// everyone else's.
	packets := []types.ClientboundPacket{
		&clientboundPlay.PlayerInfoUpdateClientboundPacket{
			Actions: clientboundPlay.PlayerInfoAddPlayer | clientboundPlay.PlayerInfoUpdateGameMode |
				clientboundPlay.PlayerInfoUpdateListed | clientboundPlay.PlayerInfoUpdateHat,
			Entries: []clientboundPlay.PlayerInfoEntry{
				{
					Profile:  player.profile,
					GameMode: player.gameMode,
					Listed:   true,
					ShowHat:  true,
				},
			},
		},
		&clientboundPlay.AddEntityClientboundPacket{
			EntityId:     player.entityId,
			Uuid:         player.profile.Uuid,
			EntityTypeId: clientboundPlay.PlayerEntityTypeId,
			X:            player.x,
			Y:            player.y,
			Z:            player.z,
			Yaw:          player.yaw,
			Pitch:        player.pitch,
			HeadYaw:      player.yaw,
		},
	}

	if player.sneaking || player.sprinting {
		packets = append(packets, &clientboundPlay.SetEntityDataClientboundPacket{
			EntityId:  player.entityId,
			Sneaking:  player.sneaking,
			Sprinting: player.sprinting,
		})
	}

	c.writeSyncPackets("show a player", packets)
}

// hidePlayer takes one player back out of this connection's world, if it was
// ever in it.
func (c *Client) hidePlayer(entityId int32, uuid string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.shownPlayers[entityId]; !ok {
		return
	}

	delete(c.shownPlayers, entityId)

	c.writeSyncPackets("hide a player", []types.ClientboundPacket{
		&clientboundPlay.RemoveEntitiesClientboundPacket{EntityIds: []int32{entityId}},
		&clientboundPlay.PlayerInfoRemoveClientboundPacket{Uuids: []string{uuid}},
	})
}

// relayToKnown writes packets about another player's entity, provided that
// player has been shown here. The check rides the same lock the spawns and
// removals decide under, so a relay cannot land before its player exists or
// after it is gone -- it is simply dropped, and the next move after the spawn
// puts the entity right.
func (c *Client) relayToKnown(entityId int32, packets ...types.ClientboundPacket) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.shownPlayers[entityId]; !ok {
		return
	}

	c.writeSyncPackets("relay to a player", packets)
}

// writeSyncPackets writes what the sync decided to send, with the lock
// already held, under the deadline above. A write that fails is this
// connection dying -- or too far behind to be waited on, which its next keep
// alive turns into the same thing -- so it is logged and otherwise left to
// the connection's own loops.
func (c *Client) writeSyncPackets(what string, packets []types.ClientboundPacket) {
	if err := c.conn.SetWriteDeadline(time.Now().Add(syncWriteTimeout)); err != nil {
		slog.Error("failed to "+what, "addr", c.conn.RemoteAddr(), "err", err)
		return
	}

	// Cleared afterwards because the deadline belongs to the connection: left
	// set, it would time out some unrelated write minutes later.
	defer c.conn.SetWriteDeadline(time.Time{})

	for _, packet := range packets {
		if err := c.writePacket(packet); err != nil {
			slog.Error("failed to "+what, "addr", c.conn.RemoteAddr(), "err", err)
			return
		}
	}
}
