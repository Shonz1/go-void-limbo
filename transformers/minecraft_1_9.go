package transformers

import (
	"bytes"
	"fmt"
	"math"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.9 step is where the play phase was rebuilt. 1.8 below it registers
// its packets in another order in both directions, so no play id is 1.9's,
// which the id tables say, and the handshake, the status and the login
// phases are numbered and laid out alike. Mojang published no mappings for
// either version, so the two were read off their own obfuscated classes,
// packet by packet, by the calls each makes on the byte buffer. Of the play
// packets this server sends, the keep alive, the login -- the dimension a
// byte in both -- the animate, the player list in each of its five actions,
// the entity removal, the head rotation and the spawn position are laid out
// alike, and so is everything this server reads but the swing, which 1.8
// sends with nothing in it, and the player command, whose actions stop one
// short of 1.9's where they overlap. A block position packs its three
// coordinates the same way in both.
//
// What differs is below: the player position, the spawn player, the
// teleport, the entity metadata and the chunk.

const (
	// playerPositionHead1_9 is the part of the player position packet 1.8
	// and 1.9 lay out alike: three doubles, two floats and the flags byte.
	playerPositionHead1_9 = 3*8 + 2*4 + 1

	// fixedPointScale1_8 is what 1.8 multiplies a coordinate by to carry it
	// as an int: a block is thirty-two steps.
	fixedPointScale1_8 = 32

	// heldItemNone1_8 is what the spawn player packet says of a player
	// holding nothing: the item's number, a short, with zero for none.
	heldItemNone1_8 = 0

	// A 1.8 entity metadata entry starts with one byte that packs the kind
	// of the value in its top three bits and the index in the low five, and
	// the run ends with a byte no entry starts with. The byte is the first
	// of 1.8's eight kinds, as the byte serializer is the first of 1.9's.
	entityDataKindShift1_8 = 5
	entityDataByte1_8      = 0
	entityDataEnd1_8       = 0x7F

	// entityFlagsIndex is the index of the flag byte every entity defines
	// first, in 1.8 as in 1.9: the one field this server sets.
	entityFlagsIndex = 0

	// A 1.8 chunk holds sixteen sections from zero up, as a 1.13.2 chunk
	// does, each of 4096 blocks, and one biome for each of its 256 columns.
	sections1_8         = 16
	blocksPerSection1_8 = 4096
	biomes1_8           = 16 * 16

	// blockIdMax1_8 is the largest number a block goes on the wire as in a
	// 1.8 chunk: a short, unsigned.
	blockIdMax1_8 = 0xFFFF

	// A 1.8 player command's actions, as it numbers them: the first six are
	// 1.9's, and 1.9 put stopping a riding jump in front of the inventory.
	playerCommandOpenInventory1_8 = 6
	playerCommandOpenInventory1_9 = 7

	// mainHand1_9 is the hand a 1.9 swing names, the only one 1.8 has.
	mainHand1_9 = 0
)

// DowngradePlayerPositionTo1_8 rewrites the player position packet from what
// 1.9 reads into what 1.8 reads: the same position, flags included, without
// the teleport id 1.9 put on its end. A 1.8 client has no packet to answer a
// teleport with, and the server waits on none from it.
func DowngradePlayerPositionTo1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyBytes(in, out, playerPositionHead1_9); err != nil {
		return err
	}

	if _, err := in.ReadVarInt(); err != nil {
		return err
	}

	return refuseRest(in, "the teleport id")
}

// DowngradeEntityPositionSyncTo1_8 rewrites the teleport -- which is what the
// entity position sync has been since the 1.21.2 step -- from what 1.9 reads
// into what 1.8 reads: the position as three ints of thirty-seconds of a
// block, rounded down, where 1.9 reads three doubles. The angles and the
// flag behind them are laid out alike.
func DowngradeEntityPositionSyncTo1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	if err := positionTo1_8(in, out); err != nil {
		return err
	}

	// The yaw, the pitch and whether the entity is on the ground.
	if err := copyBytes(in, out, 3); err != nil {
		return err
	}

	return refuseRest(in, "the teleport")
}

// DowngradeAddEntityTo1_8 rewrites the spawn player -- which is what the add
// entity packet has been since the 1.20.2 step -- from what 1.9 reads into
// what 1.8 reads: the position as three ints of thirty-seconds of a block
// where 1.9 reads three doubles, the item the player holds, a short 1.9 took
// off and this server has none for, in front of the metadata, and the
// metadata in 1.8's form.
func DowngradeAddEntityTo1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id and the uuid.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	if err := copyBytes(in, out, uuidSize); err != nil {
		return err
	}

	if err := positionTo1_8(in, out); err != nil {
		return err
	}

	// The yaw and the pitch.
	if err := copyBytes(in, out, 2); err != nil {
		return err
	}

	if err := out.WriteShort(heldItemNone1_8); err != nil {
		return err
	}

	return entityDataTo1_8(in, out)
}

// DowngradeSetEntityDataTo1_8 rewrites the entity metadata packet from what
// 1.9 reads into what 1.8 reads: the same entity, and its metadata in 1.8's
// form.
func DowngradeSetEntityDataTo1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	return entityDataTo1_8(in, out)
}

// positionTo1_8 rewrites a position from the three doubles 1.9 reads into
// the three ints 1.8 reads, each the coordinate in thirty-seconds of a block
// rounded down, as 1.8's own server writes one. A coordinate too far out for
// an int to hold its thirty-seconds is refused rather than wrapped into a
// place the server did not name.
func positionTo1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	for range 3 {
		coordinate, err := in.ReadDouble()
		if err != nil {
			return err
		}

		scaled := math.Floor(coordinate * fixedPointScale1_8)
		if math.IsNaN(scaled) || scaled < math.MinInt32 || scaled > math.MaxInt32 {
			return fmt.Errorf("coordinate %g, which 1.8 has no int for", coordinate)
		}

		if err := out.WriteInt(int32(scaled)); err != nil {
			return err
		}
	}

	return nil
}

// entityDataTo1_8 rewrites a run of entity metadata from 1.9's form into
// 1.8's. 1.9 starts an entry with its index and a var int naming its
// serializer, and ends the run with 0xFF; 1.8 starts one with a byte that
// packs the kind of the value above the index, and ends the run with 0x7F.
//
// The two versions lay their entities' fields out differently from the
// second field on -- the air 1.8 keeps as a short 1.9 keeps as a var int --
// so an index means the same field in both only for the flag byte at index
// zero, which is the one field this server sets: it travels as the one kind
// 1.8 reads a byte as, and any other entry is refused rather than filed
// under a field it is not.
func entityDataTo1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	for {
		index, err := in.ReadByte()
		if err != nil {
			return err
		}

		if index == entityDataTerminator {
			if err := out.WriteByte(entityDataEnd1_8); err != nil {
				return err
			}

			return refuseRest(in, "the entity metadata")
		}

		serializer, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		if index != entityFlagsIndex || serializer != byteSerializer {
			return fmt.Errorf("entity metadata carries index %d as serializer %d, and the flag byte at index %d is the one field 1.8 keeps where 1.9 does", index, serializer, entityFlagsIndex)
		}

		if err := out.WriteByte(entityDataByte1_8<<entityDataKindShift1_8 | index); err != nil {
			return err
		}

		if err := copyBytes(in, out, 1); err != nil {
			return err
		}
	}
}

// DowngradeLevelChunkWithSectionLightTo1_8 rewrites the chunk packet -- the
// one packet that carries a chunk and its light on 1.9 -- from what 1.9
// sends into the map chunk bulk packet 1.8 reads, holding that one chunk.
//
// 1.8 has two packets for a chunk. The chunk data packet is laid out like
// 1.9's -- the coordinates, whether the chunk is whole, the mask and the
// bytes -- but a 1.8 client told of a whole chunk with no section in it by
// that packet unloads the chunk, which is how 1.8's server says a chunk has
// gone out of view, and a chunk with nothing in it is what this server
// sends over the void. The bulk packet is how 1.8's server sends a chunk
// that has come into view, whatever is in it, and a client loads every
// chunk it names: so every chunk goes out as a bulk of one. It says whether
// the sky light is there, which the overworld's is, then how many chunks
// follow, then for each its coordinates as two ints and its mask as a
// short, and then the bytes of each, with no count in front, since the mask
// says how many there are. A bulk carries whole chunks alone, so a chunk
// that is not whole is refused.
//
// The bytes are laid out another way than 1.9's. 1.9 writes each section as
// a paletted container and its two light arrays, then the biomes; 1.8 knows
// no palette, and writes every section's blocks first, then every section's
// block light, then every section's sky light, then the biomes, a byte each
// in both. A block is a little-endian short of the number 1.9 keeps in its
// palette, which is 1.8's own number for it -- the block's number shifted up
// four bits with its variant in the low four -- since the chunk was built in
// 1.8's numbering: see package world.
func DowngradeLevelChunkWithSectionLightTo1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	x, err := in.ReadInt()
	if err != nil {
		return err
	}

	z, err := in.ReadInt()
	if err != nil {
		return err
	}

	whole, err := in.ReadBoolean()
	if err != nil {
		return err
	}

	if !whole {
		return fmt.Errorf("a chunk that is not whole, which 1.8's map chunk bulk has no form for")
	}

	mask, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if mask < 0 || mask>>sections1_8 != 0 {
		return fmt.Errorf("the mask %b names sections past the %d a 1.8 chunk holds", mask, sections1_8)
	}

	sectionData, err := in.ReadByteArray(streams.MaxPacketSize)
	if err != nil {
		return err
	}

	if err := refuseRest(in, "the sections"); err != nil {
		return err
	}

	chunk, err := chunkTo1_8(sectionData, mask)
	if err != nil {
		return err
	}

	// The sky light is there: this server's chunks are the overworld's.
	if err := out.WriteBoolean(true); err != nil {
		return err
	}

	// One chunk.
	if err := out.WriteVarInt(1); err != nil {
		return err
	}

	if err := out.WriteInt(x); err != nil {
		return err
	}

	if err := out.WriteInt(z); err != nil {
		return err
	}

	if err := out.WriteShort(int16(mask)); err != nil {
		return err
	}

	return out.WriteBytes(chunk)
}

// chunkTo1_8 rewrites 1.9's sections, and the biomes behind them, into the
// bytes of one chunk of a 1.8 map chunk bulk.
func chunkTo1_8(data []byte, mask int32) ([]byte, error) {
	in := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(data))

	var blocks, blockLight, skyLight bytes.Buffer

	for index := 0; index < sections1_8; index++ {
		if mask&(1<<index) == 0 {
			continue
		}

		if err := sectionTo1_8(in, &blocks, &blockLight, &skyLight); err != nil {
			return nil, fmt.Errorf("section %d: %w", index, err)
		}
	}

	biomes, err := in.ReadBytes(biomes1_8)
	if err != nil {
		return nil, fmt.Errorf("the biomes: %w", err)
	}

	if err := refuseRest(in, "the sections and biomes the mask names"); err != nil {
		return nil, err
	}

	out := make([]byte, 0, blocks.Len()+blockLight.Len()+skyLight.Len()+len(biomes))
	out = append(out, blocks.Bytes()...)
	out = append(out, blockLight.Bytes()...)
	out = append(out, skyLight.Bytes()...)

	return append(out, biomes...), nil
}

// sectionTo1_8 reads one of 1.9's sections -- the bits, the palette, the
// counted longs packed across one another, and the block light and the sky
// light behind them -- and writes its blocks as 1.8's little-endian shorts
// and its two light arrays each where the chunk collects them.
//
// The container is read as a 1.9 client reads it: up to four bits as a
// four-bit palette, up to eight as a palette of as many bits as it says,
// and past eight as the ids themselves at as many bits as it says, with the
// palette of none 1.12.2 and 1.9 read there all the same.
func sectionTo1_8(in *streams.MinecraftStream, blocks, blockLight, skyLight *bytes.Buffer) error {
	bits, err := in.ReadByte()
	if err != nil {
		return err
	}

	if bits == 0 || bits > 32 {
		return fmt.Errorf("a container of %d bits to an entry, which 1.9 has no form for", bits)
	}

	width := max(int(bits), 4)

	var palette []int32

	if bits <= blockIndirectBits {
		paletteSize, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		if paletteSize < 0 || paletteSize > 1<<width {
			return fmt.Errorf("a palette of %d entries at %d bits", paletteSize, width)
		}

		palette = make([]int32, paletteSize)
		for i := range palette {
			if palette[i], err = in.ReadVarInt(); err != nil {
				return err
			}
		}
	} else if dummy, err := in.ReadVarInt(); err != nil {
		return err
	} else if dummy != 0 {
		return fmt.Errorf("a container packing ids directly has a palette of %d entries, want none", dummy)
	}

	longCount, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if want := int32((blocksPerSection1_8*width + 63) / 64); longCount != want {
		return fmt.Errorf("a container of %d longs, want %d for %d bits", longCount, want, width)
	}

	packed, err := in.ReadBytes(8 * longCount)
	if err != nil {
		return err
	}

	longs := make([]uint64, longCount)
	for i := range longs {
		for j := range 8 {
			longs[i] = longs[i]<<8 | uint64(packed[8*i+j])
		}
	}

	valueMask := uint64(1)<<width - 1

	for i := range blocksPerSection1_8 {
		at := i * width
		long, shift := at/64, at%64

		value := longs[long] >> shift
		if shift+width > 64 {
			value |= longs[long+1] << (64 - shift)
		}

		id := int64(value & valueMask)
		if palette != nil {
			if id >= int64(len(palette)) {
				return fmt.Errorf("block %d names palette entry %d of %d", i, id, len(palette))
			}

			id = int64(palette[id])
		}

		if id < 0 || id > blockIdMax1_8 {
			return fmt.Errorf("block %d is state %d, which 1.8 has no short for", i, id)
		}

		blocks.WriteByte(byte(id))
		blocks.WriteByte(byte(id >> 8))
	}

	for _, light := range []*bytes.Buffer{blockLight, skyLight} {
		array, err := in.ReadBytes(lightArrayBytes)
		if err != nil {
			return err
		}

		light.Write(array)
	}

	return nil
}

// UpgradePunchFrom1_8 rewrites the swing a 1.8 client sends -- which is what
// the punch is below 26.3 -- into what 1.9 reads. 1.8 sends the swing with
// nothing in it, and 1.9 reads the hand that swung: 1.8 has the main hand
// alone, so that is the hand written.
func UpgradePunchFrom1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := refuseRest(in, "a swing 1.8 sends with nothing in it"); err != nil {
		return err
	}

	return out.WriteVarInt(mainHand1_9)
}

// UpgradePlayerCommandFrom1_8 rewrites the player command a 1.8 client sends
// into what 1.9 reads. The two number the first six actions alike --
// sneaking and sprinting starting and stopping, leaving a bed, a riding jump
// -- and 1.9 put stopping a riding jump behind them, so opening the
// inventory, the last action 1.8 has, sits one higher on 1.9. An action 1.8
// has no number for is refused.
func UpgradePlayerCommandFrom1_8(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	action, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	switch {
	case action < 0 || action > playerCommandOpenInventory1_8:
		return fmt.Errorf("player command action %d, which 1.8 has no action for", action)
	case action == playerCommandOpenInventory1_8:
		action = playerCommandOpenInventory1_9
	}

	if err := out.WriteVarInt(action); err != nil {
		return err
	}

	// The riding jump's strength.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	return refuseRest(in, "the player command")
}

// refuseRest refuses a body that goes on past the last field it should hold,
// named by what that field was, rather than sending the rest on as a body
// the version below would read something else out of.
func refuseRest(in *streams.MinecraftStream, past string) error {
	rest, err := in.ReadRest()
	if err != nil {
		return err
	}

	if len(rest) != 0 {
		return fmt.Errorf("%d bytes past %s", len(rest), past)
	}

	return nil
}
