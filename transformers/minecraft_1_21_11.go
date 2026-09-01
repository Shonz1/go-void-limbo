package transformers

import "github.com/Shonz1/go-void-limbo/streams"

// DowngradeAddEntityTo1_21_9 rewrites the add entity packet from what 1.21.11
// sends into what 1.21.9 reads.
//
// The packet's shape is identical in the two; what moved is the entity type
// registry behind one of its fields, where 1.21.11's additions renumbered the
// player. 26.1 numbers the registry as 1.21.11 does, so this is the first step
// on the way down where the id moves.
func DowngradeAddEntityTo1_21_9(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType1_21_11, playerEntityType1_21_9)
}
