package transformers

import (
	"bytes"
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.14 step is where the light left the chunk. 1.13.2 below it reads a
// section's light in the chunk packet itself, behind the section's blocks,
// and has no packet for the light alone, none for the chunk a client's view
// is centred on, and no heightmaps on the wire either: it works those out
// for itself. Mojang published no mappings for either version, so the two
// were read off their own obfuscated classes, packet by packet, by the calls
// each makes on the byte buffer: the handshake, the status and the login
// phases lay out alike -- 1.13.2 caps the login disconnect's text at a
// shorter length, longer than anything this server says -- and of the play
// phase so do the add player with its metadata, the animate, the keep alive,
// the player list, the player position, the entity removal, the head
// rotation, the teleport and everything this server reads. The two number the play phase differently on both sides, which the
// id tables say, and 1.13.2 reads the tags in three runs rather than four,
// which package gamedata writes them as.
//
// What differs is below: the chunk, the play login, the spawn position, whose
// block position packs its three coordinates in another order, and the
// entity metadata, which has no pose.

const (
	// sections1_13_2 is how many sections a 1.13.2 chunk holds, as 1.16.4's
	// does: sixteen from zero up.
	sections1_13_2 = 16

	// peaceful1_13_2 is the difficulty a 1.13.2 play login is given. 1.14
	// took the difficulty out of the play login for a packet of its own,
	// which this server does not send, and a 1.14 client that is never told
	// one holds its world at peaceful, which its level data starts at. So
	// 1.13.2 is told what the version above it assumes; nothing in a limbo
	// is hostile either way.
	peaceful1_13_2 = 0
)

// Paired carries a body made of two packets' bodies from one version to the
// next: the first behind a var int of how long it is, and then the second,
// each carried by its own packet's transformer, or as it stands where that
// packet has none. It is how the chunk of a version that reads its light
// inside the chunk reaches the 1.14 step as the chunk and the light 1.14
// would send, apart: see the chunk with section light packet in package
// play.
func Paired(first, second func(in *streams.MinecraftStream, out *streams.MinecraftStream) error) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		body, err := in.ReadByteArray(streams.MaxPacketSize)
		if err != nil {
			return err
		}

		if first != nil {
			if body, err = transformBody(first, body); err != nil {
				return fmt.Errorf("the first of the pair: %w", err)
			}
		}

		if err := out.WriteByteArray(body); err != nil {
			return err
		}

		if second == nil {
			return copyRest(in, out)
		}

		if err := second(in, out); err != nil {
			return fmt.Errorf("the second of the pair: %w", err)
		}

		return nil
	}
}

// transformBody runs one transformer over a body of its own.
func transformBody(transformer func(in *streams.MinecraftStream, out *streams.MinecraftStream) error, body []byte) ([]byte, error) {
	in := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(body))

	buf := bytes.NewBuffer(make([]byte, 0, len(body)))
	out := streams.NewMinecraftStreamFromBuffer(buf)

	if err := transformer(in, out); err != nil {
		return nil, err
	}

	if err := out.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// DowngradeLevelChunkWithSectionLightTo1_13_2 makes the chunk and its light,
// which arrive as 1.14 sends them -- the chunk packet, behind its length, and
// the light update -- into the one chunk packet 1.13.2 reads.
//
// 1.13.2 lays the chunk out as 1.14 does up to the sections, but for the
// heightmaps, which it does not read. Each section it reads is 1.14's less
// the block count in front -- the same container of blocks, with the same
// palettes at the same widths, packed the same way -- and followed by the
// section's block light and then its sky light, 2048 bytes each and nothing
// in front of them, since a section holds a whole array of each. The biomes
// of a whole chunk follow the last section, as on 1.14, and the block
// entities the sections.
//
// The light is 1.14's light update, which names the sections from one below
// the lowest to one above the highest: a section's arrays are the ones of
// the bit above its own. A section the light says nothing of, or says holds
// none, is given arrays that hold none, which for a saved world is what it
// holds: the game stores every array that holds any light.
func DowngradeLevelChunkWithSectionLightTo1_13_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	chunk, err := in.ReadByteArray(streams.MaxPacketSize)
	if err != nil {
		return err
	}

	skyLight, blockLight, lightX, lightZ, err := readLightUpdate1_14(in)
	if err != nil {
		return fmt.Errorf("the light: %w", err)
	}

	chunkIn := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(chunk))

	x, err := chunkIn.ReadInt()
	if err != nil {
		return err
	}

	z, err := chunkIn.ReadInt()
	if err != nil {
		return err
	}

	if x != lightX || z != lightZ {
		return fmt.Errorf("the chunk at %d,%d and the light of %d,%d", x, z, lightX, lightZ)
	}

	if err := out.WriteInt(x); err != nil {
		return err
	}

	if err := out.WriteInt(z); err != nil {
		return err
	}

	if _, err := copyBoolean(chunkIn, out); err != nil {
		return err
	}

	mask, err := copyVarInt(chunkIn, out)
	if err != nil {
		return err
	}

	// The heightmaps.
	if _, _, err := nbt.ReadNamed(chunkIn); err != nil {
		return err
	}

	sectionData, err := chunkIn.ReadByteArray(streams.MaxPacketSize)
	if err != nil {
		return err
	}

	sections, err := lightSectionsTo1_13_2(sectionData, mask, skyLight, blockLight)
	if err != nil {
		return err
	}

	if err := out.WriteByteArray(sections); err != nil {
		return err
	}

	// The block entities.
	return copyRest(chunkIn, out)
}

// readLightUpdate1_14 reads a light update as 1.14 lays it out -- the chunk's
// coordinates, the four masks, then the sky arrays and the block arrays in
// the order of their masks' bits, each counted -- into the arrays of every
// section the first two masks name, by bit.
func readLightUpdate1_14(in *streams.MinecraftStream) (skyLight, blockLight map[int][]byte, x, z int32, err error) {
	if x, err = in.ReadVarInt(); err != nil {
		return nil, nil, 0, 0, err
	}

	if z, err = in.ReadVarInt(); err != nil {
		return nil, nil, 0, 0, err
	}

	var masks [4]int32

	for i := range masks {
		if masks[i], err = in.ReadVarInt(); err != nil {
			return nil, nil, 0, 0, err
		}
	}

	arrays := [2]map[int][]byte{}

	for i, mask := range masks[:2] {
		arrays[i] = map[int][]byte{}

		for bit := 0; bit < lightSections1_16_4; bit++ {
			if mask&(1<<bit) == 0 {
				continue
			}

			array, err := in.ReadByteArray(lightArrayBytes)
			if err != nil {
				return nil, nil, 0, 0, err
			}

			if len(array) != lightArrayBytes {
				return nil, nil, 0, 0, fmt.Errorf("a light array of %d bytes, want %d", len(array), lightArrayBytes)
			}

			arrays[i][bit] = array
		}
	}

	if rest, err := in.ReadRest(); err != nil {
		return nil, nil, 0, 0, err
	} else if len(rest) != 0 {
		return nil, nil, 0, 0, fmt.Errorf("%d bytes past the arrays the masks name", len(rest))
	}

	return arrays[0], arrays[1], x, z, nil
}

// lightSectionsTo1_13_2 rewrites 1.14's sections, followed by the biomes of a
// whole chunk, into 1.13.2's, each with its light behind it.
func lightSectionsTo1_13_2(data []byte, mask int32, skyLight, blockLight map[int][]byte) ([]byte, error) {
	in := streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(data))

	buf := bytes.NewBuffer(make([]byte, 0, len(data)+2*lightArrayBytes*len(skyLight)))
	out := streams.NewMinecraftStreamFromBuffer(buf)

	noLight := make([]byte, lightArrayBytes)

	if mask>>sections1_13_2 != 0 {
		return nil, fmt.Errorf("the mask %b names sections past the %d a 1.13.2 chunk holds", mask, sections1_13_2)
	}

	for index := 0; index < sections1_13_2; index++ {
		if mask&(1<<index) == 0 {
			continue
		}

		if err := sectionTo1_13_2(in, out); err != nil {
			return nil, fmt.Errorf("section %d: %w", index, err)
		}

		// The light arrays are named from one below the lowest section.
		for _, arrays := range []map[int][]byte{blockLight, skyLight} {
			array := arrays[index+1]
			if array == nil {
				array = noLight
			}

			if err := out.WriteBytes(array); err != nil {
				return nil, err
			}
		}
	}

	// The biomes, when the chunk is whole, and nothing when it is not.
	if err := copyRest(in, out); err != nil {
		return nil, err
	}

	if err := out.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// sectionTo1_13_2 rewrites one of 1.14's sections into 1.13.2's blocks: the
// same container without the block count in front of it.
func sectionTo1_13_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The block count, a short, which 1.13.2 counts for itself.
	if _, err := in.ReadShort(); err != nil {
		return err
	}

	bits, err := in.ReadByte()
	if err != nil {
		return err
	}

	// A palette of one is a form neither side of this step reads: the 1.18
	// step spells it out before a section gets here.
	if bits == 0 || bits > 32 {
		return fmt.Errorf("a container of %d bits to an entry, which 1.13.2 has no form for", bits)
	}

	if err := out.WriteByte(bits); err != nil {
		return err
	}

	if bits <= blockIndirectBits {
		paletteSize, err := copyVarInt(in, out)
		if err != nil {
			return err
		}

		for range paletteSize {
			if _, err := copyVarInt(in, out); err != nil {
				return err
			}
		}
	}

	longCount, err := copyVarInt(in, out)
	if err != nil {
		return err
	}

	if longCount < 0 || longCount > streams.MaxPacketSize/8 {
		return fmt.Errorf("a container of %d longs", longCount)
	}

	return copyBytes(in, out, 8*longCount)
}

// DowngradePlayLoginTo1_13_2 rewrites the play phase login packet from what
// 1.14 sends into what 1.13.2 reads.
//
// 1.13.2 reads the difficulty behind the dimension, a byte, where 1.14 took
// it out for a packet of its own, and has no view distance, which 1.14 added
// behind the level type: a 1.13.2 client takes its view distance from its
// own options alone. The rest is laid out alike.
func DowngradePlayLoginTo1_13_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id, a plain int, the game mode's byte and the dimension,
	// another int.
	if err := copyBytes(in, out, 9); err != nil {
		return err
	}

	if err := out.WriteByte(peaceful1_13_2); err != nil {
		return err
	}

	// The most players, a byte.
	if err := copyBytes(in, out, 1); err != nil {
		return err
	}

	// The level type.
	if err := copyString(in, out); err != nil {
		return err
	}

	// The view distance.
	if _, err := in.ReadVarInt(); err != nil {
		return err
	}

	// Whether the debug screen is cut down.
	if _, err := copyBoolean(in, out); err != nil {
		return err
	}

	return nil
}

// The bits a block position packs each coordinate into, the same widths on
// both sides of the step: 26 for either horizontal one and 12 for the
// height. 1.14 packs the x, then the z, then the height from the top bits
// down; 1.13.2 packs the x, then the height, then the z.
const (
	blockPosHorizontalBits = 26
	blockPosHeightBits     = 12
)

// DowngradeSetDefaultSpawnPositionTo1_13_2 rewrites the spawn position --
// which the game event packet is below the 1.20.3 step -- from what 1.14
// sends into what 1.13.2 reads: the same one long, with the coordinates
// packed in 1.13.2's order.
func DowngradeSetDefaultSpawnPositionTo1_13_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	packed, err := in.ReadLong()
	if err != nil {
		return err
	}

	// Each coordinate is signed: shifted to the top of the long, and back
	// down arithmetically.
	x := packed >> (64 - blockPosHorizontalBits)
	z := packed << blockPosHorizontalBits >> (64 - blockPosHorizontalBits)
	y := packed << (64 - blockPosHeightBits) >> (64 - blockPosHeightBits)

	const (
		horizontalMask = 1<<blockPosHorizontalBits - 1
		heightMask     = 1<<blockPosHeightBits - 1
	)

	repacked := x<<(blockPosHeightBits+blockPosHorizontalBits) | (y&heightMask)<<blockPosHorizontalBits | z&horizontalMask

	if err := out.WriteLong(repacked); err != nil {
		return err
	}

	return copyRest(in, out)
}

// DowngradeSetEntityDataTo1_13_2 rewrites the entity metadata packet from
// what 1.14 sends into what 1.13.2 reads.
//
// 1.14 is where an entity took on a pose, the field behind the six every
// entity defined before it, and the serializer the pose is written with, the
// last 1.14 registers. 1.13.2 has neither: a player it is told is sneaking by the
// flag byte crouches by that alone. So the pose comes off, and the flag byte
// travels as it is. Every entry this server sends is one of the two, and
// anything else is refused rather than guessed at.
func DowngradeSetEntityDataTo1_13_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The entity id.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	for {
		index, err := in.ReadByte()
		if err != nil {
			return err
		}

		if index == entityDataTerminator {
			return out.WriteByte(index)
		}

		serializer, err := in.ReadVarInt()
		if err != nil {
			return err
		}

		switch serializer {
		case byteSerializer:
			if err := out.WriteByte(index); err != nil {
				return err
			}

			if err := out.WriteVarInt(serializer); err != nil {
				return err
			}

			if err := copyBytes(in, out, 1); err != nil {
				return err
			}
		case poseSerializer1_19_1:
			// The pose's ordinal.
			if _, err := in.ReadVarInt(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("set entity data carries serializer %d, which this rewrite does not know the shape of", serializer)
		}
	}
}
