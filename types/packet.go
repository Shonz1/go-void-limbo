package types

import "go-void-limbo/streams"

type PacketId = int32

type ServerboundPacket interface {
	String() string
}

type ClientboundPacket interface {
	String() string
	Encode(ms *streams.MinecraftStream) error
}
