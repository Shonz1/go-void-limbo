// Package transformers carries packet bodies between neighbouring protocol
// versions.
//
// A packet is implemented once, at the latest version. What an older client
// sends is carried up to that version before it is decoded, and what is sent
// back to it is carried down from that version before it goes out. Each
// transformer here is one step of that, and knows only the one difference
// between the two versions it sits between.
//
// The bulk of any transformer is copying the fields it does not touch, since a
// body has to be rewritten in full to change anything in the middle of it. The
// helpers below are that copying, and read from the body given to the
// transformer straight into the body replacing it.
package transformers

import "github.com/Shonz1/go-void-limbo/streams"

// copyBytes moves a fixed width field across.
func copyBytes(in *streams.MinecraftStream, out *streams.MinecraftStream, count int32) error {
	value, err := in.ReadBytes(count)
	if err != nil {
		return err
	}

	return out.WriteBytes(value)
}

// copyVarInt moves a var int across and returns it, since a var int in front of
// an array is what says how much of the array follows.
func copyVarInt(in *streams.MinecraftStream, out *streams.MinecraftStream) (int32, error) {
	value, err := in.ReadVarInt()
	if err != nil {
		return 0, err
	}

	return value, out.WriteVarInt(value)
}

func copyString(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	value, err := in.ReadString()
	if err != nil {
		return err
	}

	return out.WriteString(value)
}

// copyBoolean moves a boolean across and returns it, since a boolean in front
// of an optional field is what says whether the field follows.
func copyBoolean(in *streams.MinecraftStream, out *streams.MinecraftStream) (bool, error) {
	value, err := in.ReadBoolean()
	if err != nil {
		return false, err
	}

	return value, out.WriteBoolean(value)
}

// copyRest moves everything left across, for the tail of a body that has
// nothing further to change in it.
func copyRest(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	value, err := in.ReadRest()
	if err != nil {
		return err
	}

	return out.WriteBytes(value)
}
