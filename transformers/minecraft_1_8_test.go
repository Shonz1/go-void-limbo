package transformers

import (
	"bytes"
	"compress/zlib"
	"math"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
)

// 1.7.6 reads the keep alive's id as an int where 1.8 reads a var int, and
// sends it back the same way: an id goes down as an int and comes back up
// as the var int it was, every bit of it.
func TestKeepAliveCrossesThe1_8StepBothWays(t *testing.T) {
	for _, id := range []int32{1, -1, math.MaxInt32, math.MinInt32, 1234567890} {
		sent := encodeBody(t, func(ms *streams.MinecraftStream) error { return ms.WriteVarInt(id) })
		want := encodeBody(t, func(ms *streams.MinecraftStream) error { return ms.WriteInt(id) })

		down := runTransformer(t, DowngradeKeepAliveTo1_7_6, sent)
		if !bytes.Equal(down, want) {
			t.Errorf("id %d to 1.7.6 = % x, want % x", id, down, want)
		}

		if up := runTransformer(t, UpgradeKeepAliveFrom1_7_6, down); !bytes.Equal(up, sent) {
			t.Errorf("id %d back from 1.7.6 = % x, want % x", id, up, sent)
		}
	}

	if err := failingTransformer(t, DowngradeKeepAliveTo1_7_6, []byte{0x01, 0x00}); err == nil {
		t.Error("expected a keep alive with bytes past its id to be refused")
	}

	if err := failingTransformer(t, UpgradeKeepAliveFrom1_7_6, []byte{0, 0, 0, 1, 0}); err == nil {
		t.Error("expected a keep alive with bytes past its id to be refused")
	}
}

// 1.7.6 counts the encryption request's two arrays with a short where 1.8
// counts with a var int, and the response's the same way: the request goes
// down with short counts and the response comes up with var int counts.
func TestEncryptionCrossesThe1_8StepBothWays(t *testing.T) {
	key := bytes.Repeat([]byte{0xAB}, 162)
	token := []byte{1, 2, 3, 4}

	request := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteString(""), ms.WriteByteArray(key), ms.WriteByteArray(token))
	})

	wantRequest := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteString(""), ms.WriteShort(162), ms.WriteBytes(key), ms.WriteShort(4), ms.WriteBytes(token))
	})

	if got := runTransformer(t, DowngradeEncryptionRequestTo1_7_6, request); !bytes.Equal(got, wantRequest) {
		t.Errorf("request to 1.7.6 = % x, want % x", got, wantRequest)
	}

	if err := failingTransformer(t, DowngradeEncryptionRequestTo1_7_6, append(request, 0x00)); err == nil {
		t.Error("expected a request with bytes past its token to be refused")
	}

	wide := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteString(""), ms.WriteByteArray(make([]byte, math.MaxInt16+1)), ms.WriteByteArray(token))
	})

	if err := failingTransformer(t, DowngradeEncryptionRequestTo1_7_6, wide); err == nil {
		t.Error("expected a key past what a short counts to be refused")
	}

	secret := bytes.Repeat([]byte{0xCD}, 128)

	response := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteShort(128), ms.WriteBytes(secret), ms.WriteShort(128), ms.WriteBytes(secret))
	})

	wantResponse := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteByteArray(secret), ms.WriteByteArray(secret))
	})

	if got := runTransformer(t, UpgradeEncryptionResponseFrom1_7_6, response); !bytes.Equal(got, wantResponse) {
		t.Errorf("response from 1.7.6 = % x, want % x", got, wantResponse)
	}

	if err := failingTransformer(t, UpgradeEncryptionResponseFrom1_7_6, []byte{0xFF, 0xFF}); err == nil {
		t.Error("expected a negative count to be refused")
	}

	if err := failingTransformer(t, UpgradeEncryptionResponseFrom1_7_6, append(response, 0x00)); err == nil {
		t.Error("expected a response with bytes past its token to be refused")
	}
}

// 1.7.6 reads the play login 1.8 does, cut off before the reduced debug
// flag.
func TestDowngradePlayLoginTo1_7_6(t *testing.T) {
	head := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteInt(7), ms.WriteByte(1), ms.WriteByte(0), ms.WriteByte(2), ms.WriteByte(20), ms.WriteString("flat"))
	})

	sent := append(append([]byte{}, head...), 0x01)

	if got := runTransformer(t, DowngradePlayLoginTo1_7_6, sent); !bytes.Equal(got, head) {
		t.Errorf("to 1.7.6 = % x, want % x", got, head)
	}

	if err := failingTransformer(t, DowngradePlayLoginTo1_7_6, head); err == nil {
		t.Error("expected a login with no reduced debug flag to be refused")
	}

	if err := failingTransformer(t, DowngradePlayLoginTo1_7_6, append(sent, 0x00)); err == nil {
		t.Error("expected a login with bytes past the flag to be refused")
	}
}

// 1.7.6 reads the spawn position as three ints where 1.8 packs them in a
// long in 1.13.2's order, the negative ones included.
func TestDowngradeSetDefaultSpawnPositionTo1_7_6(t *testing.T) {
	for _, position := range [][3]int64{{8, 65, 8}, {-1, -64, -33554432}, {33554431, 2047, 0}} {
		x, y, z := position[0], position[1], position[2]
		packed := x<<38 | (y&0xFFF)<<26 | z&0x3FFFFFF

		sent := encodeBody(t, func(ms *streams.MinecraftStream) error { return ms.WriteLong(packed) })
		want := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(ms.WriteInt(int32(x)), ms.WriteInt(int32(y)), ms.WriteInt(int32(z)))
		})

		if got := runTransformer(t, DowngradeSetDefaultSpawnPositionTo1_7_6, sent); !bytes.Equal(got, want) {
			t.Errorf("%v to 1.7.6 = % x, want % x", position, got, want)
		}
	}

	if err := failingTransformer(t, DowngradeSetDefaultSpawnPositionTo1_7_6, make([]byte, 9)); err == nil {
		t.Error("expected a position with bytes past its long to be refused")
	}
}

// 1.7.6 reads the player position 1.8 does with the height raised to the
// eyes and an on-ground flag of false where 1.8 reads the relative flags;
// any relative flag is refused.
func TestDowngradePlayerPositionTo1_7_6(t *testing.T) {
	position := func(flags byte) []byte {
		return encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(ms.WriteDouble(8.5), ms.WriteDouble(65), ms.WriteDouble(-8.5), ms.WriteFloat(90), ms.WriteFloat(-10), ms.WriteByte(flags))
		})
	}

	want := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteDouble(8.5), ms.WriteDouble(65+eyeHeight1_7_6), ms.WriteDouble(-8.5), ms.WriteFloat(90), ms.WriteFloat(-10), ms.WriteBoolean(false))
	})

	if got := runTransformer(t, DowngradePlayerPositionTo1_7_6, position(0)); !bytes.Equal(got, want) {
		t.Errorf("to 1.7.6 = % x, want % x", got, want)
	}

	if err := failingTransformer(t, DowngradePlayerPositionTo1_7_6, position(0x02)); err == nil {
		t.Error("expected a relative height to be refused")
	}

	if err := failingTransformer(t, DowngradePlayerPositionTo1_7_6, append(position(0), 0x00)); err == nil {
		t.Error("expected a position with bytes past the flags to be refused")
	}
}

// 1.8 reads the moves a 1.7.6 client sends with the eye height taken out
// from between the feet and the z.
func TestUpgradeMovesFrom1_7_6(t *testing.T) {
	sentPosition := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteDouble(8.5), ms.WriteDouble(65), ms.WriteDouble(66.62), ms.WriteDouble(-8.5), ms.WriteBoolean(true))
	})

	wantPosition := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteDouble(8.5), ms.WriteDouble(65), ms.WriteDouble(-8.5), ms.WriteBoolean(true))
	})

	if got := runTransformer(t, UpgradeMovePlayerPositionFrom1_7_6, sentPosition); !bytes.Equal(got, wantPosition) {
		t.Errorf("position from 1.7.6 = % x, want % x", got, wantPosition)
	}

	if err := failingTransformer(t, UpgradeMovePlayerPositionFrom1_7_6, append(sentPosition, 0x00)); err == nil {
		t.Error("expected a position with bytes past the flag to be refused")
	}

	sentRotation := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteDouble(8.5), ms.WriteDouble(65), ms.WriteDouble(66.62), ms.WriteDouble(-8.5), ms.WriteFloat(90), ms.WriteFloat(-10), ms.WriteBoolean(false))
	})

	wantRotation := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteDouble(8.5), ms.WriteDouble(65), ms.WriteDouble(-8.5), ms.WriteFloat(90), ms.WriteFloat(-10), ms.WriteBoolean(false))
	})

	if got := runTransformer(t, UpgradeMovePlayerPositionRotationFrom1_7_6, sentRotation); !bytes.Equal(got, wantRotation) {
		t.Errorf("position and rotation from 1.7.6 = % x, want % x", got, wantRotation)
	}

	if err := failingTransformer(t, UpgradeMovePlayerPositionRotationFrom1_7_6, sentRotation[:len(sentRotation)-1]); err == nil {
		t.Error("expected a move cut short to be refused")
	}
}

// uuid1_8 is a uuid as 1.8 carries one, and uuidText1_7_6 the same uuid as
// 1.7.6 carries it, spelled out.
var (
	uuid1_8       = []byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0, 0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC, 0xDE, 0xF0}
	uuidText1_7_6 = "12345678-9abc-def0-1234-56789abcdef0"
)

// playerListAdd1_8 lays a player list add out as 1.8 reads it: one entry
// with the uuid, the name, one signed property and one unsigned, the game
// mode, the ping and no display name.
func playerListAdd1_8(t *testing.T, ping int32) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(0),
			ms.WriteVarInt(1),
			ms.WriteBytes(uuid1_8),
			ms.WriteString("Notch"),
			ms.WriteVarInt(2),
			ms.WriteString("textures"),
			ms.WriteString("value"),
			ms.WriteBoolean(true),
			ms.WriteString("signature"),
			ms.WriteString("other"),
			ms.WriteString("plain"),
			ms.WriteBoolean(false),
			ms.WriteVarInt(1),
			ms.WriteVarInt(ping),
			ms.WriteBoolean(false),
		)
	})
}

// playerListItem1_7_6 lays a player list item out as 1.7.6 reads it.
func playerListItem1_7_6(t *testing.T, name string, online bool, ping int16) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteString(name), ms.WriteBoolean(online), ms.WriteShort(ping))
	})
}

// 1.7.6 reads a player list add as the name, on and the ping, and the
// removal of the same player as the name, off and no ping, with the name
// remembered from the add; the spawn player between the two carries the
// name and the properties the add carried, each property with a signature
// whether or not it had one.
func TestPlayerListAndSpawnPlayerTo1_7_6(t *testing.T) {
	cache := NewProfileCache1_7_6()

	add := DowngradePlayerInfoUpdateTo1_7_6(cache)
	remove := DowngradePlayerInfoRemoveTo1_7_6(cache)
	spawn := DowngradeAddEntityTo1_7_6(cache)

	spawnBody := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(300),
			ms.WriteBytes(uuid1_8),
			ms.WriteInt(16),
			ms.WriteInt(64*32),
			ms.WriteInt(16),
			ms.WriteByte(10),
			ms.WriteByte(20),
			ms.WriteShort(0),
			ms.WriteBytes(entityData1_8(0x02, true)),
		)
	})

	removeBody := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteVarInt(4), ms.WriteVarInt(1), ms.WriteBytes(uuid1_8))
	})

	// Before the add, neither the spawn nor the removal has a name to go
	// out under.
	if err := failingTransformer(t, spawn, spawnBody); err == nil {
		t.Error("expected a spawn before the player list entry to be refused")
	}

	if err := failingTransformer(t, remove, removeBody); err == nil {
		t.Error("expected a removal before the player list entry to be refused")
	}

	if got, want := runTransformer(t, add, playerListAdd1_8(t, 42)), playerListItem1_7_6(t, "Notch", true, 42); !bytes.Equal(got, want) {
		t.Errorf("add to 1.7.6 = % x, want % x", got, want)
	}

	wantSpawn := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(
			ms.WriteVarInt(300),
			ms.WriteString(uuidText1_7_6),
			ms.WriteString("Notch"),
			ms.WriteVarInt(2),
			ms.WriteString("textures"),
			ms.WriteString("value"),
			ms.WriteString("signature"),
			ms.WriteString("other"),
			ms.WriteString("plain"),
			ms.WriteString(""),
			ms.WriteInt(16),
			ms.WriteInt(64*32),
			ms.WriteInt(16),
			ms.WriteByte(10),
			ms.WriteByte(20),
			ms.WriteShort(0),
			ms.WriteBytes(entityData1_8(0x02, true)),
		)
	})

	if got := runTransformer(t, spawn, spawnBody); !bytes.Equal(got, wantSpawn) {
		t.Errorf("spawn to 1.7.6 = % x, want % x", got, wantSpawn)
	}

	// A spawn with no metadata in it takes on the flag byte, unset: a 1.7.6
	// client crashes on a run with nothing in it.
	bare := append(append([]byte{}, spawnBody[:len(spawnBody)-len(entityData1_8(0x02, true))]...), entityData1_8(0, false)...)
	wantBare := append(append([]byte{}, wantSpawn[:len(wantSpawn)-len(entityData1_8(0x02, true))]...), entityData1_8(0x00, true)...)

	if got := runTransformer(t, spawn, bare); !bytes.Equal(got, wantBare) {
		t.Errorf("a spawn with no metadata to 1.7.6 = % x, want % x", got, wantBare)
	}

	if got, want := runTransformer(t, remove, removeBody), playerListItem1_7_6(t, "Notch", false, 0); !bytes.Equal(got, want) {
		t.Errorf("removal to 1.7.6 = % x, want % x", got, want)
	}

	// A removal does not forget the player: another 1.7.6 client may have
	// the same removal to come.
	if got, want := runTransformer(t, remove, removeBody), playerListItem1_7_6(t, "Notch", false, 0); !bytes.Equal(got, want) {
		t.Errorf("a second removal to 1.7.6 = % x, want % x", got, want)
	}

	// Another action, two entries, a display name and a ping past a short
	// are refused.
	refused := map[string][]byte{
		"game mode action": append([]byte{0x01}, playerListAdd1_8(t, 0)[1:]...),
		"two entries":      append([]byte{0x00, 0x02}, playerListAdd1_8(t, 0)[2:]...),
		"display name": func() []byte {
			body := playerListAdd1_8(t, 0)
			body[len(body)-1] = 0x01

			return body
		}(),
		"ping past a short":    playerListAdd1_8(t, math.MaxInt16+1),
		"bytes past the entry": append(playerListAdd1_8(t, 0), 0x00),
	}

	for name, body := range refused {
		if err := failingTransformer(t, add, body); err == nil {
			t.Errorf("%s: expected the player list add to be refused", name)
		}
	}

	if err := failingTransformer(t, remove, append([]byte{0x00}, removeBody[1:]...)); err == nil {
		t.Error("expected a removal under another action to be refused")
	}
}

// A cache past its capacity forgets the profiles put longest ago, and a
// profile put again is as new as its latest put.
func TestProfileCache1_7_6ForgetsTheOldest(t *testing.T) {
	cache := NewProfileCache1_7_6()

	cache.put("first", profile1_7_6{name: "first"})

	for i := range profileCacheCapacity1_7_6 - 1 {
		cache.put(string(rune('a'+i%26))+string(rune(i)), profile1_7_6{})
	}

	// Full to the brim, and the first still there; put again, it survives
	// the next put, which forgets its older spot and nothing else.
	if _, ok := cache.get("first"); !ok {
		t.Fatal("the first profile is gone from a cache that is merely full")
	}

	cache.put("first", profile1_7_6{name: "again"})
	cache.put("one more", profile1_7_6{})

	if profile, ok := cache.get("first"); !ok || profile.name != "again" {
		t.Errorf("the first profile is %v, %t, want the one put again", profile, ok)
	}

	for i := range profileCacheCapacity1_7_6 {
		cache.put(string(rune(i))+"later", profile1_7_6{})
	}

	if _, ok := cache.get("first"); ok {
		t.Error("the first profile outlived a capacity of later ones")
	}

	if len(cache.profiles) > profileCacheCapacity1_7_6 || len(cache.order) > profileCacheCapacity1_7_6 {
		t.Errorf("the cache holds %d profiles over %d spots, want at most %d", len(cache.profiles), len(cache.order), profileCacheCapacity1_7_6)
	}
}

// 1.7.6 reads the entity removal with a byte count and int ids, the
// teleport, the head rotation and the entity metadata with an int entity
// id, and the teleport without the on-ground flag.
func TestEntityPacketsTo1_7_6(t *testing.T) {
	remove := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteVarInt(2), ms.WriteVarInt(300), ms.WriteVarInt(-1))
	})

	wantRemove := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteByte(2), ms.WriteInt(300), ms.WriteInt(-1))
	})

	if got := runTransformer(t, DowngradeRemoveEntitiesTo1_7_6, remove); !bytes.Equal(got, wantRemove) {
		t.Errorf("removal to 1.7.6 = % x, want % x", got, wantRemove)
	}

	tooMany := encodeBody(t, func(ms *streams.MinecraftStream) error { return ms.WriteVarInt(128) })
	if err := failingTransformer(t, DowngradeRemoveEntitiesTo1_7_6, tooMany); err == nil {
		t.Error("expected more entities than a signed byte counts to be refused")
	}

	teleport := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteVarInt(300), ms.WriteInt(16), ms.WriteInt(2048), ms.WriteInt(-16), ms.WriteByte(10), ms.WriteByte(20), ms.WriteBoolean(true))
	})

	wantTeleport := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteInt(300), ms.WriteInt(16), ms.WriteInt(2048), ms.WriteInt(-16), ms.WriteByte(10), ms.WriteByte(20))
	})

	if got := runTransformer(t, DowngradeEntityPositionSyncTo1_7_6, teleport); !bytes.Equal(got, wantTeleport) {
		t.Errorf("teleport to 1.7.6 = % x, want % x", got, wantTeleport)
	}

	if err := failingTransformer(t, DowngradeEntityPositionSyncTo1_7_6, append(teleport, 0x00)); err == nil {
		t.Error("expected a teleport with bytes past the flag to be refused")
	}

	head := encodeBody(t, func(ms *streams.MinecraftStream) error { return writeAll(ms.WriteVarInt(300), ms.WriteByte(64)) })
	wantHead := encodeBody(t, func(ms *streams.MinecraftStream) error { return writeAll(ms.WriteInt(300), ms.WriteByte(64)) })

	if got := runTransformer(t, DowngradeRotateHeadTo1_7_6, head); !bytes.Equal(got, wantHead) {
		t.Errorf("head rotation to 1.7.6 = % x, want % x", got, wantHead)
	}

	if err := failingTransformer(t, DowngradeRotateHeadTo1_7_6, append(head, 0x00)); err == nil {
		t.Error("expected a head rotation with bytes past the yaw to be refused")
	}

	metadata := append([]byte{0x07}, entityData1_8(0x0A, true)...)
	wantMetadata := append([]byte{0, 0, 0, 7}, entityData1_8(0x0A, true)...)

	if got := runTransformer(t, DowngradeSetEntityDataTo1_7_6, metadata); !bytes.Equal(got, wantMetadata) {
		t.Errorf("metadata to 1.7.6 = % x, want % x", got, wantMetadata)
	}
}

// inflate1_7_6 reads a 1.7.6 bulk back out: the deflated bytes inflated,
// and the fields around them.
func inflate1_7_6(t *testing.T, body []byte) (skyLight bool, chunk []byte, tail []byte) {
	t.Helper()

	ms := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(body))

	count, err := ms.ReadShort()
	if err != nil || count != 1 {
		t.Fatalf("count = %d, %v, want one chunk", count, err)
	}

	length, err := ms.ReadInt()
	if err != nil {
		t.Fatalf("reading the length: %v", err)
	}

	if skyLight, err = ms.ReadBoolean(); err != nil {
		t.Fatalf("reading the sky light flag: %v", err)
	}

	deflated, err := ms.ReadBytes(length)
	if err != nil {
		t.Fatalf("reading %d deflated bytes: %v", length, err)
	}

	reader, err := zlib.NewReader(bytes.NewReader(deflated))
	if err != nil {
		t.Fatalf("opening the deflated bytes: %v", err)
	}

	var inflated bytes.Buffer
	if _, err := inflated.ReadFrom(reader); err != nil {
		t.Fatalf("inflating: %v", err)
	}

	if tail, err = ms.ReadRest(); err != nil {
		t.Fatalf("reading the tail: %v", err)
	}

	return skyLight, inflated.Bytes(), tail
}

// 1.7.6 reads the chunk 1.8 does as a map chunk bulk whose bytes are
// deflated, with every section's block ids a byte each in front of every
// section's variants a nibble each, the light and the biomes behind, and
// the chunk's coordinates and two masks behind the bytes.
func TestDowngradeLevelChunkWithSectionLightTo1_7_6(t *testing.T) {
	// Stone at block 0, coarse dirt at block 1 and the last block, a
	// block of id 255 at variant 15 at block 2.
	ids := make([]int32, 4096)
	ids[0], ids[1], ids[2], ids[4095] = 16, 49, 0xFFF, 49

	var data []byte
	data = append(data, blocks1_8(ids)...)
	data = append(data, bytes.Repeat([]byte{0x0F}, 2048)...)
	data = append(data, bytes.Repeat([]byte{0xF0}, 2048)...)
	biomes := bytes.Repeat([]byte{0x01}, 256)
	data = append(data, biomes...)

	got := runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_7_6, bulk1_8(t, 1<<3, data))

	skyLight, chunk, tail := inflate1_7_6(t, got)
	if !skyLight {
		t.Error("the sky light flag is off, want it on as it was sent")
	}

	wantBlocks := make([]byte, 4096)
	wantBlocks[0], wantBlocks[1], wantBlocks[2], wantBlocks[4095] = 1, 3, 255, 3

	wantVariants := make([]byte, 2048)
	wantVariants[0], wantVariants[1], wantVariants[2047] = 1<<4, 0xF, 1<<4

	var want []byte
	want = append(want, wantBlocks...)
	want = append(want, wantVariants...)
	want = append(want, bytes.Repeat([]byte{0x0F}, 2048)...)
	want = append(want, bytes.Repeat([]byte{0xF0}, 2048)...)
	want = append(want, biomes...)

	if !bytes.Equal(chunk, want) {
		t.Errorf("the chunk inflates to %d bytes, want %d; first difference at %d", len(chunk), len(want), firstDifference(chunk, want))
	}

	wantTail := encodeBody(t, func(ms *streams.MinecraftStream) error {
		return writeAll(ms.WriteInt(3), ms.WriteInt(-4), ms.WriteShort(1<<3), ms.WriteShort(0))
	})

	if !bytes.Equal(tail, wantTail) {
		t.Errorf("the tail is % x, want % x", tail, wantTail)
	}

	// A chunk with no section in it is a bulk of one chunk with nothing
	// but its biomes.
	_, chunk, _ = inflate1_7_6(t, runTransformer(t, DowngradeLevelChunkWithSectionLightTo1_7_6, bulk1_8(t, 0, biomes)))
	if !bytes.Equal(chunk, biomes) {
		t.Errorf("an empty chunk inflates to % x, want the biomes alone", chunk)
	}

	// A block past a byte, a bulk of two chunks and bytes past the biomes
	// are refused.
	wide := make([]int32, 4096)
	wide[0] = 0x1000

	var wideData []byte
	wideData = append(wideData, blocks1_8(wide)...)
	wideData = append(wideData, make([]byte, 2*2048)...)
	wideData = append(wideData, biomes...)

	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_7_6, bulk1_8(t, 1, wideData)); err == nil {
		t.Error("expected a block past a byte to be refused")
	}

	two := bulk1_8(t, 0, biomes)
	two[1] = 0x02

	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_7_6, two); err == nil {
		t.Error("expected a bulk of two chunks to be refused")
	}

	if err := failingTransformer(t, DowngradeLevelChunkWithSectionLightTo1_7_6, bulk1_8(t, 1<<3, append(data, 0x00))); err == nil {
		t.Error("expected bytes past the biomes to be refused")
	}
}

// 1.8 reads the swing with nothing in it, which 1.7.6 sends with the
// client's entity and the arm swing; another animation is refused.
func TestUpgradePunchFrom1_7_6(t *testing.T) {
	swing := func(animation byte) []byte {
		return encodeBody(t, func(ms *streams.MinecraftStream) error { return writeAll(ms.WriteInt(7), ms.WriteByte(animation)) })
	}

	if got := runTransformer(t, UpgradePunchFrom1_7_6, swing(1)); len(got) != 0 {
		t.Errorf("from 1.7.6 = % x, want nothing", got)
	}

	if err := failingTransformer(t, UpgradePunchFrom1_7_6, swing(2)); err == nil {
		t.Error("expected an animation other than the arm swing to be refused")
	}

	if err := failingTransformer(t, UpgradePunchFrom1_7_6, append(swing(1), 0x00)); err == nil {
		t.Error("expected a swing with bytes past the animation to be refused")
	}
}

// 1.8 reads the player command 1.7.6 sends with its actions counted from
// zero rather than one and its three fields as var ints; an action 1.7.6
// has no number for is refused.
func TestUpgradePlayerCommandFrom1_7_6(t *testing.T) {
	command := func(action byte) []byte {
		return encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(ms.WriteInt(300), ms.WriteByte(action), ms.WriteInt(50))
		})
	}

	for action := byte(1); action <= 7; action++ {
		want := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(ms.WriteVarInt(300), ms.WriteVarInt(int32(action-1)), ms.WriteVarInt(50))
		})

		if got := runTransformer(t, UpgradePlayerCommandFrom1_7_6, command(action)); !bytes.Equal(got, want) {
			t.Errorf("action %d: from 1.7.6 = % x, want % x", action, got, want)
		}
	}

	for _, action := range []byte{0, 8, 0xFF} {
		if err := failingTransformer(t, UpgradePlayerCommandFrom1_7_6, command(action)); err == nil {
			t.Errorf("action %d: expected an action 1.7.6 has no number for to be refused", action)
		}
	}

	if err := failingTransformer(t, UpgradePlayerCommandFrom1_7_6, append(command(1), 0x00)); err == nil {
		t.Error("expected a command with bytes past its jump strength to be refused")
	}
}

// 1.8 reads the player input 1.7.6 sends with its two booleans packed into
// one byte of flags.
func TestUpgradePlayerInputFrom1_7_6(t *testing.T) {
	for _, test := range []struct {
		jump, unmount bool
		flags         byte
	}{{false, false, 0}, {true, false, 1}, {false, true, 2}, {true, true, 3}} {
		sent := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(ms.WriteFloat(0.5), ms.WriteFloat(-1), ms.WriteBoolean(test.jump), ms.WriteBoolean(test.unmount))
		})

		want := encodeBody(t, func(ms *streams.MinecraftStream) error {
			return writeAll(ms.WriteFloat(0.5), ms.WriteFloat(-1), ms.WriteByte(test.flags))
		})

		if got := runTransformer(t, UpgradePlayerInputFrom1_7_6, sent); !bytes.Equal(got, want) {
			t.Errorf("jump %t, unmount %t: from 1.7.6 = % x, want % x", test.jump, test.unmount, got, want)
		}

		if err := failingTransformer(t, UpgradePlayerInputFrom1_7_6, append(sent, 0x00)); err == nil {
			t.Error("expected an input with bytes past its flags to be refused")
		}
	}
}
