package transformers

import (
	"bytes"
	"fmt"
	"math"
	"math/bits"
	"sync"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.8 step is where the protocol's numbers stood still and its bodies
// moved. 1.7.6 below it registers every packet this server speaks under the
// id 1.8 does, in every phase and both directions, which the id tables say,
// and all but one of 1.8's login packets: the set compression, which 1.8
// added and 1.7.6 has no id for, so a 1.7.6 connection is never compressed
// at all. Mojang published no mappings for either version, so the two were
// read off their own obfuscated classes, packet by packet, by the calls
// each makes on the byte buffer. Of what this server sends, the animate,
// the login success, the two disconnects and the status are laid out
// alike; and so is everything this server reads but the player position
// with and without the rotation, which 1.7.6 sends with a second height,
// the swing, which 1.7.6 sends with an entity and an animation in it, the
// player command, whose actions 1.7.6 numbers from one and counts in a
// byte, and the player input, whose two flags 1.7.6 sends as two booleans.
//
// What differs is below: the keep alive, the encryption request, the play
// login, the spawn position, the player position, the spawn player, the
// player list, the entity removal, the teleport, the head rotation, the
// entity metadata and the chunk down; and the keep alive, the encryption
// response, the two moves, the swing, the player command and the player
// input up.

const (
	// eyeHeight1_7_6 is what 1.7.6's server adds to a player's height
	// before sending its position: 1.7.6 keeps a player's position at its
	// eyes, 1.8 at its feet, and the client sets its own feet that far
	// below what it is told. It is the float 1.62 widened to a double, as
	// the server's own constant is.
	eyeHeight1_7_6 = 1.6200000047683716

	// byteArrayMax1_7_6 is the longest byte array 1.7.6 reads out of the
	// login phase: a count in a short, signed.
	byteArrayMax1_7_6 = math.MaxInt16

	// playLoginHead1_8 is the part of the play login 1.7.6 and 1.8 lay out
	// alike in front of the level type: the entity id, an int, and the game
	// mode, the dimension, the difficulty and the most players, a byte each.
	playLoginHead1_8 = 4 + 4

	// entityPositionTail1_8 is what the teleport carries behind its entity
	// id that 1.7.6 reads as 1.8 does: three ints of thirty-seconds of a
	// block, and the yaw and the pitch, a byte each.
	entityPositionTail1_8 = 3*4 + 2

	// removeEntitiesMax1_7_6 is the most entities 1.7.6 reads out of one
	// entity removal: its count is a byte, signed.
	removeEntitiesMax1_7_6 = math.MaxInt8

	// A 1.7.6 player list packet carries no player's game mode, and 1.8's
	// carries each player's as a var int; the add action is the one 1.8
	// action this server sends through the update packet, and the remove
	// action the one it sends through the remove packet.
	playerInfoActionAdd1_8    = 0
	playerInfoActionRemove1_8 = 4

	// The 1.7.6 player list's ping is a short.
	pingMax1_7_6 = math.MaxInt16

	// A 1.7.6 chunk's sections are laid out as 1.8's are read: 4096 block
	// ids, a byte each, and 2048 bytes of variants, of block light and of
	// sky light, a nibble each. 1.8 carries a block as a short, the id above
	// its variant's four bits; a 1.7.6 section holds an id a byte wide, and
	// a second array for the ids past a byte, which no block registered on
	// 1.7.6 reaches.
	blockBytes1_7_6   = 4096
	nibbleBytes1_7_6  = 2048
	blockIdMax1_7_6   = 0xFF
	blockVariantShift = 4
	blockVariantMask  = 0xF

	// A 1.7.6 player command's actions are 1.8's numbered from one: sneaking
	// and sprinting starting and stopping, leaving a bed, a riding jump and
	// opening the inventory, seven in all, where 1.8 counts the same seven
	// from zero.
	playerCommandFirst1_7_6 = 1
	playerCommandLast1_7_6  = 7

	// swingArm1_7_6 is the one animation a 1.7.6 client swings with: the
	// swing packet names an entity and an animation, and the client sends
	// its own entity and this.
	swingArm1_7_6 = 1

	// The 1.8 player input packs its two flags in one byte, jumping in the
	// low bit and unmounting above it, which 1.7.6 sends as two booleans in
	// that order.
	playerInputFlagJump1_8    = 0x01
	playerInputFlagUnmount1_8 = 0x02
)

// DowngradeKeepAliveTo1_7_6 rewrites the keep alive from what 1.8 reads into
// what 1.7.6 reads: the id, a var int, as an int. The two hold the same
// thirty-two bits, so nothing is refused.
func DowngradeKeepAliveTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	id, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if err := out.WriteInt(id); err != nil {
		return err
	}

	return refuseRest(in, "the keep alive id")
}

// UpgradeKeepAliveFrom1_7_6 rewrites the keep alive from what 1.7.6 sends
// into what 1.8 sends: the id, an int, as a var int, which holds the same
// bits.
func UpgradeKeepAliveFrom1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	id, err := in.ReadInt()
	if err != nil {
		return err
	}

	if err := out.WriteVarInt(id); err != nil {
		return err
	}

	return refuseRest(in, "the keep alive id")
}

// DowngradeEncryptionRequestTo1_7_6 rewrites the encryption request from
// what 1.8 reads into what 1.7.6 reads: the server id as it is, and then
// the public key and the verify token each counted with a short where 1.8
// counts with a var int. An array past what a short counts is refused,
// since 1.7.6 would read a negative count out of it.
func DowngradeEncryptionRequestTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyString(in, out); err != nil {
		return err
	}

	for _, field := range []string{"public key", "verify token"} {
		array, err := in.ReadByteArray(streams.MaxPacketSize)
		if err != nil {
			return err
		}

		if len(array) > byteArrayMax1_7_6 {
			return fmt.Errorf("a %s of %d bytes, which 1.7.6 has no short for", field, len(array))
		}

		if err := out.WriteShort(int16(len(array))); err != nil {
			return err
		}

		if err := out.WriteBytes(array); err != nil {
			return err
		}
	}

	return refuseRest(in, "the verify token")
}

// UpgradeEncryptionResponseFrom1_7_6 rewrites the encryption response from
// what 1.7.6 sends into what 1.8 sends: the shared secret and the encrypted
// challenge, each counted with a short, counted with a var int instead. A
// negative count is refused, as 1.8's own server refuses one.
func UpgradeEncryptionResponseFrom1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	for _, field := range []string{"shared secret", "verify token"} {
		count, err := in.ReadShort()
		if err != nil {
			return err
		}

		if count < 0 {
			return fmt.Errorf("a %s of %d bytes", field, count)
		}

		array, err := in.ReadBytes(int32(count))
		if err != nil {
			return err
		}

		if err := out.WriteByteArray(array); err != nil {
			return err
		}
	}

	return refuseRest(in, "the verify token")
}

// DowngradePlayLoginTo1_7_6 rewrites the play phase login from what 1.8
// reads into what 1.7.6 reads: the same entity id, game mode, dimension,
// difficulty, most players and level type, and nothing after the level
// type, where 1.8 reads whether its debug screen is to be reduced. The flag
// comes off the end, read so that it is consumed and never written.
func DowngradePlayLoginTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyBytes(in, out, playLoginHead1_8); err != nil {
		return err
	}

	if err := copyString(in, out); err != nil {
		return err
	}

	if _, err := in.ReadBoolean(); err != nil {
		return err
	}

	return refuseRest(in, "the reduced debug flag")
}

// DowngradeSetDefaultSpawnPositionTo1_7_6 rewrites the spawn position --
// which the game event packet is below the 1.20.3 step -- from what 1.8
// reads into what 1.7.6 reads: the three coordinates as three ints, where
// 1.8 packs them in one long in 1.13.2's order, the x in the top twenty-six
// bits, the y in the twelve below and the z in the twenty-six at the
// bottom, each signed.
func DowngradeSetDefaultSpawnPositionTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	packed, err := in.ReadLong()
	if err != nil {
		return err
	}

	x := packed >> (blockPosHeightBits + blockPosHorizontalBits)
	y := packed << blockPosHorizontalBits >> (64 - blockPosHeightBits)
	z := packed << (blockPosHeightBits + blockPosHorizontalBits) >> (blockPosHeightBits + blockPosHorizontalBits)

	for _, coordinate := range []int64{x, y, z} {
		if err := out.WriteInt(int32(coordinate)); err != nil {
			return err
		}
	}

	return refuseRest(in, "the spawn position")
}

// DowngradePlayerPositionTo1_7_6 rewrites the player position packet from
// what 1.8 reads into what 1.7.6 reads: the same position and rotation,
// with the height raised by the eye height, and whether the player is on
// the ground where 1.8 reads which coordinates are relative.
//
// 1.7.6 keeps a player's position at its eyes, and its client puts the feet
// 1.62 below whatever height this packet carries, which is why 1.7.6's own
// server adds the same before sending one. 1.8 keeps the feet, and this
// server with it, so the height goes up by that much on the way down. The
// 1.8 flags say which coordinates the client adds to its own rather than
// takes as they are, which 1.7.6 has no way to be told: a position with any
// of them set is refused rather than taken as absolute. The client is told
// it is off the ground, as 1.7.6's own server tells one it has put
// somewhere, and the client says where it lands.
func DowngradePlayerPositionTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	x, err := in.ReadDouble()
	if err != nil {
		return err
	}

	y, err := in.ReadDouble()
	if err != nil {
		return err
	}

	z, err := in.ReadDouble()
	if err != nil {
		return err
	}

	// The yaw and the pitch.
	rotation, err := in.ReadBytes(2 * 4)
	if err != nil {
		return err
	}

	flags, err := in.ReadByte()
	if err != nil {
		return err
	}

	if flags != 0 {
		return fmt.Errorf("player position carries relative flags %#x, which 1.7.6 has no way to be told", flags)
	}

	if err := refuseRest(in, "the relative flags"); err != nil {
		return err
	}

	for _, coordinate := range []float64{x, y + eyeHeight1_7_6, z} {
		if err := out.WriteDouble(coordinate); err != nil {
			return err
		}
	}

	if err := out.WriteBytes(rotation); err != nil {
		return err
	}

	return out.WriteBoolean(false)
}

// UpgradeMovePlayerPositionFrom1_7_6 rewrites the position a 1.7.6 client
// sends into what 1.8 reads. 1.7.6 sends two heights, the feet and then the
// eyes -- the stance, as its server calls it, checked there to sit between
// a tenth and a block and two thirds above the feet -- where 1.8 sends the
// feet alone, so the eyes come out from between the feet and the z. The
// rest is laid out alike: three doubles in front of the one that goes, and
// whether the player is on the ground behind.
func UpgradeMovePlayerPositionFrom1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := dropStance1_7_6(in, out); err != nil {
		return err
	}

	// The z, and whether the player is on the ground.
	if err := copyBytes(in, out, 8+1); err != nil {
		return err
	}

	return refuseRest(in, "the on-ground flag")
}

// UpgradeMovePlayerPositionRotationFrom1_7_6 rewrites the position and
// rotation a 1.7.6 client sends into what 1.8 reads: as the position
// above, with the yaw and the pitch between the z and the flag.
func UpgradeMovePlayerPositionRotationFrom1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := dropStance1_7_6(in, out); err != nil {
		return err
	}

	// The z, the yaw, the pitch, and whether the player is on the ground.
	if err := copyBytes(in, out, 8+2*4+1); err != nil {
		return err
	}

	return refuseRest(in, "the on-ground flag")
}

// dropStance1_7_6 carries the x and the feet across and reads the eyes
// behind them off.
func dropStance1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyBytes(in, out, 2*8); err != nil {
		return err
	}

	_, err := in.ReadDouble()

	return err
}

// A ProfileCache1_7_6 remembers the name and the skin properties of every
// player the 1.8 step has carried a player list entry for, by uuid, for the
// packets 1.7.6 reads them in that 1.8 carries them nowhere near.
//
// 1.8 is where a player's profile went into the player list alone: its
// spawn player packet names the player by uuid, and the client takes the
// name and the skin from the list entry it was sent first. 1.7.6's spawn
// player carries the name and the skin properties itself, and its player
// list, the name and nothing else, keeps no uuid at all, so a removal names
// the player by name. Neither can be written from the 1.8 packet alone, and
// a transformer sees one body at a time: so this is what the 1.8 step's
// player list transformer puts the profile into, and what its spawn player
// and removal transformers read it back out of. The entry is sent before
// the spawn on every version, and before the removal by the time there is
// anything to remove, so a lookup that misses is a packet sent out of that
// order, which is refused.
//
// The cache is one per registry, and so shared by every connection: a
// profile is the same wherever it is sent, and whichever connection put it
// here. It is never emptied by a removal, since another 1.7.6 client may
// have the same player to remove yet, and is bounded instead: past its
// capacity the profiles remembered longest ago are forgotten, which is far
// more players than a limbo shows at once, and a player forgotten while
// still shown is one whose removal is refused and whose name stays in that
// list.
type ProfileCache1_7_6 struct {
	mu sync.Mutex

	profiles map[string]profile1_7_6

	// order is every uuid in the order it was last put, oldest first, as a
	// ring that forgets from its front once full: a uuid put again sits in
	// it twice, and the older spot is passed over when its turn comes.
	order []string
}

// profileCacheCapacity1_7_6 is how many profiles a cache keeps before
// forgetting the oldest. A profile is a name and a skin, a kilobyte or so
// with Mojang's signature, so the whole cache is a few tens of megabytes at
// its fullest.
const profileCacheCapacity1_7_6 = 1 << 14

// profile1_7_6 is what a 1.7.6 spawn player carries of a player beside its
// uuid and position: the name and the properties, each a name, a value and
// a signature.
type profile1_7_6 struct {
	name       string
	properties []property1_7_6
}

type property1_7_6 struct {
	name, value, signature string
}

func NewProfileCache1_7_6() *ProfileCache1_7_6 {
	return &ProfileCache1_7_6{profiles: make(map[string]profile1_7_6)}
}

func (c *ProfileCache1_7_6) put(uuid string, profile profile1_7_6) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.profiles[uuid] = profile
	c.order = append(c.order, uuid)

	for len(c.order) > profileCacheCapacity1_7_6 {
		oldest := c.order[0]
		c.order = c.order[1:]

		// A uuid put again since is still in the ring further on, and
		// keeps its profile until that spot comes round.
		stillInRing := false
		for _, later := range c.order {
			if later == oldest {
				stillInRing = true
				break
			}
		}

		if !stillInRing {
			delete(c.profiles, oldest)
		}
	}
}

func (c *ProfileCache1_7_6) get(uuid string) (profile1_7_6, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	profile, ok := c.profiles[uuid]

	return profile, ok
}

// DowngradePlayerInfoUpdateTo1_7_6 rewrites the player list update from what
// 1.8 reads into what 1.7.6 reads, remembering the profile it carries in
// cache for the spawn player and the removal that follow.
//
// 1.8 lays the packet out as an action, a count and that many entries, each
// keyed by uuid; 1.7.6 reads one player's name, whether the player is on
// and the player's ping, a short. The one action this server sends through
// this packet is the add, one player at a time, whose entry is the uuid, the
// name, the properties, the game mode, the ping and a display name this
// server leaves absent: the name and the ping go out, the player on. Any
// other action, more than one entry, a ping past a short and a display name
// are refused: 1.7.6 has no packet for a game mode or a display name, and
// no room in one for two players.
func DowngradePlayerInfoUpdateTo1_7_6(cache *ProfileCache1_7_6) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		action, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		if action != playerInfoActionAdd1_8 {
			return fmt.Errorf("player list action %d, which 1.7.6 has no packet for", action)
		}

		if err := onePlayer1_7_6(in); err != nil {
			return err
		}

		uuid, err := in.ReadUuid()
		if err != nil {
			return err
		}

		name, err := in.ReadString()
		if err != nil {
			return err
		}

		properties, err := readProfileProperties1_7_6(in)
		if err != nil {
			return err
		}

		// The game mode.
		if _, err := in.ReadVarInt(); err != nil {
			return err
		}

		ping, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		if ping < 0 || ping > pingMax1_7_6 {
			return fmt.Errorf("a ping of %d, which 1.7.6 has no short for", ping)
		}

		hasDisplayName, err := in.ReadBoolean()
		if err != nil {
			return err
		}

		if hasDisplayName {
			return fmt.Errorf("player list carries a display name, which 1.7.6 has no field for")
		}

		if err := refuseRest(in, "the player list entry"); err != nil {
			return err
		}

		cache.put(uuid, profile1_7_6{name: name, properties: properties})

		return writePlayerListItem1_7_6(out, name, true, int16(ping))
	}
}

// DowngradePlayerInfoRemoveTo1_7_6 rewrites the player list removal from
// what 1.8 reads -- the update packet under its remove action, a count and
// that many uuids -- into what 1.7.6 reads: the player's name, off, and a
// ping of nothing. The name is the one cache remembers for the uuid from
// the entry that added the player; a uuid it does not remember is refused.
func DowngradePlayerInfoRemoveTo1_7_6(cache *ProfileCache1_7_6) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		action, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		if action != playerInfoActionRemove1_8 {
			return fmt.Errorf("player list action %d, want the remove", action)
		}

		if err := onePlayer1_7_6(in); err != nil {
			return err
		}

		uuid, err := in.ReadUuid()
		if err != nil {
			return err
		}

		if err := refuseRest(in, "the uuid"); err != nil {
			return err
		}

		profile, ok := cache.get(uuid)
		if !ok {
			return fmt.Errorf("player list removes %s, whose name no entry has carried", uuid)
		}

		return writePlayerListItem1_7_6(out, profile.name, false, 0)
	}
}

// onePlayer1_7_6 reads a 1.8 player list packet's count of entries, which
// is one or refused: a 1.7.6 packet holds one player.
func onePlayer1_7_6(in *streams.MinecraftStream) error {
	entries, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if entries != 1 {
		return fmt.Errorf("player list carries %d entries, and a 1.7.6 packet holds one", entries)
	}

	return nil
}

// readProfileProperties1_7_6 reads a profile's properties as 1.8 lays them
// out: a count, and each a name, a value, whether it is signed and the
// signature if so.
func readProfileProperties1_7_6(in *streams.MinecraftStream) ([]property1_7_6, error) {
	count, err := in.ReadVarInt()
	if err != nil {
		return nil, err
	}

	if count < 0 {
		return nil, fmt.Errorf("%d properties", count)
	}

	properties := make([]property1_7_6, 0, count)

	for range count {
		var property property1_7_6

		if property.name, err = in.ReadString(); err != nil {
			return nil, err
		}

		if property.value, err = in.ReadString(); err != nil {
			return nil, err
		}

		signed, err := in.ReadBoolean()
		if err != nil {
			return nil, err
		}

		if signed {
			if property.signature, err = in.ReadString(); err != nil {
				return nil, err
			}
		}

		properties = append(properties, property)
	}

	return properties, nil
}

func writePlayerListItem1_7_6(out *streams.MinecraftStream, name string, online bool, ping int16) error {
	if err := out.WriteString(name); err != nil {
		return err
	}

	if err := out.WriteBoolean(online); err != nil {
		return err
	}

	return out.WriteShort(ping)
}

// DowngradeAddEntityTo1_7_6 rewrites the spawn player -- which is what the
// add entity packet has been since the 1.20.2 step -- from what 1.8 reads
// into what 1.7.6 reads: the same entity id, then the uuid as text, the way
// java.util.UUID spells one out, then the name and the properties cache
// remembers for it from the player list entry that added the player, and
// then the position, the rotation, the held item and the metadata as they
// came. Each property goes out as its name, its value and its signature,
// which 1.7.6 reads as three strings whether or not the property was
// signed, so an unsigned one is sent with an empty signature. A uuid cache
// does not remember is refused: the client would show a player with no
// name.
//
// The metadata is laid out alike, but a 1.7.6 client reads a run with
// nothing in it as no run at all, and then reaches for a data watcher the
// packet never has, which crashes it. So a run that is only its end takes
// on the flag byte at index zero, unset, which is what the player's flags
// are when this server sends none.
func DowngradeAddEntityTo1_7_6(cache *ProfileCache1_7_6) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		// The entity id.
		if _, err := copyVarInt(in, out); err != nil {
			return err
		}

		uuid, err := in.ReadUuid()
		if err != nil {
			return err
		}

		profile, ok := cache.get(uuid)
		if !ok {
			return fmt.Errorf("spawn player names %s, whose profile no player list entry has carried", uuid)
		}

		if err := out.WriteString(uuid); err != nil {
			return err
		}

		if err := out.WriteString(profile.name); err != nil {
			return err
		}

		if err := out.WriteVarInt(int32(len(profile.properties))); err != nil {
			return err
		}

		for _, property := range profile.properties {
			for _, field := range []string{property.name, property.value, property.signature} {
				if err := out.WriteString(field); err != nil {
					return err
				}
			}
		}

		// The position as three ints, the yaw, the pitch and the held item.
		if err := copyBytes(in, out, entityPositionTail1_8+2); err != nil {
			return err
		}

		metadata, err := in.ReadRest()
		if err != nil {
			return err
		}

		if bytes.Equal(metadata, []byte{entityDataEnd1_8}) {
			metadata = []byte{entityDataByte1_8<<entityDataKindShift1_8 | entityFlagsIndex, 0x00, entityDataEnd1_8}
		}

		return out.WriteBytes(metadata)
	}
}

// DowngradeRemoveEntitiesTo1_7_6 rewrites the entity removal from what 1.8
// reads into what 1.7.6 reads: the count as a byte and each id as an int,
// where 1.8 reads var ints throughout. 1.7.6 reads its count signed, so more
// entities than a byte holds that way are refused.
func DowngradeRemoveEntitiesTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	count, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if count < 0 || count > removeEntitiesMax1_7_6 {
		return fmt.Errorf("%d entities to remove, which 1.7.6 has no byte for", count)
	}

	if err := out.WriteByte(byte(count)); err != nil {
		return err
	}

	for range count {
		if err := varIntToInt(in, out); err != nil {
			return err
		}
	}

	return refuseRest(in, "the entity ids")
}

// DowngradeEntityPositionSyncTo1_7_6 rewrites the teleport -- which is what
// the entity position sync has been since the 1.21.2 step -- from what 1.8
// reads into what 1.7.6 reads: the entity id as an int where 1.8 reads a
// var int, the same position and rotation, and nothing behind them, where
// 1.8 reads whether the entity is on the ground. The flag comes off the
// end, read so that it is consumed and never written.
func DowngradeEntityPositionSyncTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := varIntToInt(in, out); err != nil {
		return err
	}

	if err := copyBytes(in, out, entityPositionTail1_8); err != nil {
		return err
	}

	if _, err := in.ReadBoolean(); err != nil {
		return err
	}

	return refuseRest(in, "the on-ground flag")
}

// DowngradeRotateHeadTo1_7_6 rewrites the head rotation from what 1.8 reads
// into what 1.7.6 reads: the entity id as an int where 1.8 reads a var int,
// and the same yaw behind it.
func DowngradeRotateHeadTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := varIntToInt(in, out); err != nil {
		return err
	}

	if err := copyBytes(in, out, 1); err != nil {
		return err
	}

	return refuseRest(in, "the head yaw")
}

// DowngradeSetEntityDataTo1_7_6 rewrites the entity metadata packet from
// what 1.8 reads into what 1.7.6 reads: the entity id as an int where 1.8
// reads a var int, and the metadata as it is, which the two lay out alike
// -- a byte packing the kind above the index, the value, and 0x7F at the
// end -- with the flag byte at index zero in both.
func DowngradeSetEntityDataTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := varIntToInt(in, out); err != nil {
		return err
	}

	return copyRest(in, out)
}

// varIntToInt carries an entity id across from the var int 1.8 reads to
// the int 1.7.6 reads, which hold the same bits.
func varIntToInt(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	id, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	return out.WriteInt(id)
}

// DowngradeLevelChunkWithSectionLightTo1_7_6 rewrites the chunk -- which
// goes to 1.8 as a map chunk bulk of one chunk, by the 1.9 step -- from
// what 1.8 reads into what 1.7.6 reads: the same bulk, laid out the older
// way. 1.7.6's bulk is the count of chunks, a short; the length of the
// chunks' bytes, deflated together; whether the sky light is there; the
// deflated bytes; and then for each chunk its coordinates as two ints and
// two masks as shorts, the sections with blocks in them and the sections
// with blocks past a byte in them. 1.8's is whether the sky light is there,
// a var int count, each chunk's coordinates and one mask, and the bytes of
// each chunk as they are, with no count in front of them.
//
// A 1.7.6 chunk's bytes are every section's block ids, a byte each, then
// every section's variants, then their block light, then their sky light,
// a nibble each, then the biomes, a byte each; the sections past a byte,
// which 1.7.6 reads between the sky light and the biomes, are none, since
// no block registered on 1.7.6 has an id past a byte. 1.8's are every
// section's blocks as little-endian shorts, the id above the variant's four
// bits, then the light arrays and the biomes laid out alike. So the shorts
// come apart into the two arrays, and a block whose id is past a byte is
// refused rather than sent under the eight bits 1.7.6 would read. The 1.8
// bulk's one chunk is the one here; a bulk of any other count is refused.
func DowngradeLevelChunkWithSectionLightTo1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	skyLight, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	count, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("a bulk of %d chunks, where the 1.9 step sends one", count)
	}

	// The chunk's coordinates.
	coordinates, err := in.ReadBytes(2 * 4)
	if err != nil {
		return err
	}

	mask, err := in.ReadShort()
	if err != nil {
		return err
	}

	data, err := in.ReadRest()
	if err != nil {
		return err
	}

	chunk, err := chunkTo1_7_6(data, uint16(mask), skyLight)
	if err != nil {
		return err
	}

	deflated, err := streams.Compress(chunk)
	if err != nil {
		return err
	}

	// One chunk.
	if err := out.WriteShort(1); err != nil {
		return err
	}

	if err := out.WriteInt(int32(len(deflated))); err != nil {
		return err
	}

	if err := out.WriteBoolean(skyLight); err != nil {
		return err
	}

	if err := out.WriteBytes(deflated); err != nil {
		return err
	}

	if err := out.WriteBytes(coordinates); err != nil {
		return err
	}

	if err := out.WriteShort(mask); err != nil {
		return err
	}

	// The mask of sections with blocks past a byte: none.
	return out.WriteShort(0)
}

// chunkTo1_7_6 rewrites the bytes of one chunk of a 1.8 map chunk bulk
// into the bytes of one chunk of a 1.7.6 bulk, before they are deflated.
func chunkTo1_7_6(data []byte, mask uint16, skyLight bool) ([]byte, error) {
	sections := bits.OnesCount16(mask)

	lightArrays := 1
	if skyLight {
		lightArrays = 2
	}

	want := sections*(2*blockBytes1_7_6+lightArrays*nibbleBytes1_7_6) + biomes1_8
	if len(data) != want {
		return nil, fmt.Errorf("a chunk of %d bytes for %d sections, want %d", len(data), sections, want)
	}

	out := bytes.NewBuffer(make([]byte, 0, sections*(blockBytes1_7_6+(1+lightArrays)*nibbleBytes1_7_6)+biomes1_8))

	variants := make([]byte, sections*nibbleBytes1_7_6)

	for section := range sections {
		for i := range blockBytes1_7_6 {
			at := 2 * (section*blockBytes1_7_6 + i)
			block := uint16(data[at]) | uint16(data[at+1])<<8

			id := block >> blockVariantShift
			if id > blockIdMax1_7_6 {
				return nil, fmt.Errorf("section %d block %d is id %d, which 1.7.6 has no byte for", section, i, id)
			}

			out.WriteByte(byte(id))

			variants[section*nibbleBytes1_7_6+i/2] |= byte(block&blockVariantMask) << (4 * (i % 2))
		}
	}

	out.Write(variants)

	// The light arrays and the biomes, laid out alike.
	out.Write(data[2*sections*blockBytes1_7_6:])

	return out.Bytes(), nil
}

// UpgradePunchFrom1_7_6 rewrites the swing a 1.7.6 client sends -- which is
// what the punch is below 26.3 -- into what 1.8 reads, which is nothing:
// 1.7.6 sends the entity that swung and the animation it swung with, and
// 1.8 sends the swing with nothing in it. The entity is the client's own,
// and the animation the one arm swing a client has; another animation is
// refused rather than sent on as a swing.
func UpgradePunchFrom1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := in.ReadInt(); err != nil {
		return err
	}

	animation, err := in.ReadByte()
	if err != nil {
		return err
	}

	if animation != swingArm1_7_6 {
		return fmt.Errorf("animation %d, which is not the arm swing", animation)
	}

	return refuseRest(in, "the animation")
}

// UpgradePlayerCommandFrom1_7_6 rewrites the player command a 1.7.6 client
// sends into what 1.8 reads: the entity id as a var int where 1.7.6 sends an
// int, the action as a var int counted from zero where 1.7.6 sends a byte
// counted from one -- the two number the same seven actions in the same
// order -- and the riding jump's strength as a var int where 1.7.6 sends an
// int. An action 1.7.6 has no number for is refused.
func UpgradePlayerCommandFrom1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	entityId, err := in.ReadInt()
	if err != nil {
		return err
	}

	if err := out.WriteVarInt(entityId); err != nil {
		return err
	}

	action, err := in.ReadByte()
	if err != nil {
		return err
	}

	if action < playerCommandFirst1_7_6 || action > playerCommandLast1_7_6 {
		return fmt.Errorf("player command action %d, which 1.7.6 has no action for", action)
	}

	if err := out.WriteVarInt(int32(action - playerCommandFirst1_7_6)); err != nil {
		return err
	}

	strength, err := in.ReadInt()
	if err != nil {
		return err
	}

	if err := out.WriteVarInt(strength); err != nil {
		return err
	}

	return refuseRest(in, "the player command")
}

// UpgradePlayerInputFrom1_7_6 rewrites the player input a 1.7.6 client sends
// into what 1.8 reads: the same two floats, sideways and forward, and then
// whether to jump and whether to unmount as the two bits of one byte where
// 1.7.6 sends them as two booleans.
func UpgradePlayerInputFrom1_7_6(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyBytes(in, out, 2*4); err != nil {
		return err
	}

	var flags byte

	for _, bit := range []byte{playerInputFlagJump1_8, playerInputFlagUnmount1_8} {
		set, err := in.ReadBoolean()
		if err != nil {
			return err
		}

		if set {
			flags |= bit
		}
	}

	if err := out.WriteByte(flags); err != nil {
		return err
	}

	return refuseRest(in, "the player input")
}
