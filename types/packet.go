package types

import "go-void-limbo/streams"

type PacketId = int32

type ServerboundPacket interface {
	ToString() string
}

type ClientboundPacket interface {
	Encode(ms *streams.MinecraftStream) error
}
