package transformers

import (
	"errors"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.19.1 step is the smallest below the configuration phase, read off
// the 1.19 client's own classes through Mojang's mappings the way the 1.19.3
// step was. 1.19.1 and 1.19 lay all but two of the packets this server
// speaks out alike -- the play login, laid out alike but carrying registries
// that are 1.19's own, since 1.19.1 is where the chat types were reworked;
// and the hello, onto whose end 1.19.1 put the optional uuid -- and each
// rewrite below is one of those two. Everything else is wire-identical in
// 759 and 760: the encryption request and the response with its signed or
// encrypted challenge, the login on both sides of the success packet, the
// add player packet, the player list with its five actions, the player
// position with its dismount flag, the movement on both sides, the entity
// metadata with the pose at 18, the tags, and the chunk packet with its
// named heightmaps and its trust edges flag. 1.19.1 did renumber the play
// phase, in both directions, which the id tables say; this file is only
// about the bodies.

// DowngradePlayLoginTo1_19 rewrites the play phase login packet from what
// 1.19.1 sends into what 1.19 reads, given the registries that version
// reads out of it.
//
// The two versions lay the packet out alike from front to back, and read the
// same three registries out of it -- the dimension types, the biomes and the
// chat types -- so the one thing to change is the compound in the middle,
// which is 1.19.1's, put there by the 1.19.3 step, and which 1.19 has its
// own of: the same three registries with the chat types laid out as 1.19
// reads them, encoded once by package gamedata. As with the steps above, a
// transformer built with no registries to write refuses every login rather
// than send a client into a world it cannot make sense of.
func DowngradePlayLoginTo1_19(registryCodec []byte) func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		if len(registryCodec) == 0 {
			return errors.New("a 1.19 play login carries the registries, and this transformer was given none")
		}

		return swapPlayLoginRegistries(in, out, registryCodec)
	}
}

// UpgradeLoginStartFrom1_19 rewrites the login start packet from what 1.19
// sends into what 1.19.1 reads.
//
// 1.19's hello is the name and the optional profile key, and 1.19.1 put an
// optional uuid behind the two, which 1.19 has no field for. So the body is
// copied whole -- the key with it, for the 1.19.3 step to take off -- and
// the uuid goes on the end absent, which is what the 1.20 step turns into a
// hello with no uuid in it, as it would for a 1.19.1 client that sent none.
func UpgradeLoginStartFrom1_19(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	if err := copyRest(in, out); err != nil {
		return err
	}

	return out.WriteBoolean(false)
}
