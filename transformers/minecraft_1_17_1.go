package transformers

import (
	"fmt"

	"github.com/Shonz1/go-void-limbo/streams"
)

// 1.17.1 is 1.17 with one packet put back the way it was. Read off the two
// clients' own classes, compared whole: the protocol registration is the
// same list in the same order, so nothing is renumbered, and of the packets
// whose layout 1.17.1 changed -- the container content and slot, the
// container click, the book edit and the entity removal -- the removal is
// the only one this server speaks. The registries, the tags' names, the
// entity metadata and the blocks are 1.17.1's in 1.17 as they stand.

// DowngradeRemoveEntitiesTo1_17 rewrites the remove entities packet from what
// 1.17.1 sends into what 1.17 reads.
//
// 1.17 is the one version that takes entities out of the world one to a
// packet: its packet is the entity id alone, where every version before and
// after it reads a count and that many ids. A packet cannot become several
// on its way down, so a packet naming any number of entities but one is
// refused rather than cut short; this server removes a player at a time, so
// one is what it sends.
func DowngradeRemoveEntitiesTo1_17(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	count, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	if count != 1 {
		return fmt.Errorf("a 1.17 remove entity packet takes out one entity, and this one names %d", count)
	}

	entityId, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	return out.WriteVarInt(entityId)
}
