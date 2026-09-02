package transformers

import "github.com/Shonz1/go-void-limbo/streams"

// DowngradeAddEntityTo1_21_5 rewrites the add entity packet from what 1.21.6
// sends into what 1.21.5 reads.
//
// The packet's shape is identical in the two -- 1.21.5 already reads the
// velocity as the three shorts at the end that 1.21.7 does -- and what moved
// is the entity type registry behind one of its fields: 1.21.6 added the happy
// ghast, which sorts before the player and so pushed its number up by one.
func DowngradeAddEntityTo1_21_5(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return downgradeAddEntityType(in, out, playerEntityType1_21_6, playerEntityType1_21_5)
}
