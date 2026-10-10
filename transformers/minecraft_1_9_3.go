package transformers

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.9.3 step is where the chunk packet grew its list of block entities.
// 1.9.2 below it reads a chunk that ends at its sections, and was told of a
// chest or a sign by a packet of its own. The two jars were compared class by
// class with the names taken out: every phase registers the same packets in
// the same order in both, but that 1.9.2's play phase registers one
// clientbound packet more, the update sign at 0x46, which 1.9.3 dropped, so
// every clientbound play id from 0x47 up sits one higher on 1.9.2, which the
// id tables say. Of the packets both register, the chunk below is the one
// that differs on the wire: the keep alive, the login, the player position,
// the spawn player, the entity metadata and everything this server reads are
// the same classes with the names taken out, and so are the packet buffer,
// the section and its container, the palettes, the bit array and the nibble
// arrays the chunk is made of. The blocks are 1.9.3's to the last state: see
// package gamedata.

// DowngradeLevelChunkWithSectionLightTo1_9_2 rewrites the chunk packet from
// what 1.9.3 reads into what 1.9.2 reads: the same chunk, cut off behind its
// sections. The coordinates, the whole flag, the mask and the section bytes
// are laid out alike; behind them 1.9.3 reads a count of block entities and
// that many compounds, and 1.9.2 reads nothing.
//
// This server sends no block entity, so the count it cuts off is none in
// practice. One it is handed all the same is read through, compound by
// compound, so that a chunk that is not what the count says it is -- cut
// short, or with bytes past its last compound -- is refused rather than sent
// on as a chunk 1.9.2 would read something else out of.
func DowngradeLevelChunkWithSectionLightTo1_9_2(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	// The chunk's coordinates, two plain ints.
	if err := copyBytes(in, out, 8); err != nil {
		return err
	}

	// Whether the chunk is whole.
	if _, err := copyBoolean(in, out); err != nil {
		return err
	}

	// The mask of the sections sent.
	if _, err := copyVarInt(in, out); err != nil {
		return err
	}

	if err := copyByteArray(in, out); err != nil {
		return err
	}

	// The block entities, which 1.9.2 has no room for.
	count, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if count < 0 {
		return fmt.Errorf("%d block entities, which is no count", count)
	}

	for index := range count {
		if _, _, err := nbt.ReadNamed(in); err != nil {
			return fmt.Errorf("block entity %d: %w", index, err)
		}
	}

	if rest, err := in.ReadRest(); err != nil {
		return err
	} else if len(rest) != 0 {
		return fmt.Errorf("%d bytes past the block entities the count names", len(rest))
	}

	return nil
}
