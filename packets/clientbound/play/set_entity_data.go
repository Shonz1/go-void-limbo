package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// The pieces of the entity metadata this server ever sets, all in service of
// showing one player's stance to another. The numbers are the client's own
// tables at the latest version.
const (
	// entityFlagsIndex is the entity's shared flag byte, the first field every
	// entity defines, and entityFlagsSerializerId is the byte serializer, the
	// first one registered.
	entityFlagsIndex        = 0
	entityFlagsSerializerId = 0

	// entityFlagSneaking and entityFlagSprinting are the two bits of the flag
	// byte a player's own movement flips. The rest are states -- on fire,
	// invisible, gliding -- that nothing here produces.
	entityFlagSneaking  = 0x02
	entityFlagSprinting = 0x08

	// entityPoseIndex is the entity's pose field, sixth of the fields every
	// entity defines, and entityPoseSerializerId is the pose serializer's spot
	// in the serializer registry.
	entityPoseIndex        = 6
	entityPoseSerializerId = 20

	// poseStanding and poseCrouching are ordinals in the client's pose enum.
	// The flag bit above says the shift key is down; the pose is what actually
	// lowers the camera and shrinks the hitbox, and the client wants both.
	poseStanding  = 0
	poseCrouching = 5

	// entityDataTerminator ends the list of metadata entries, sitting where no
	// real index can: an entry's index is a byte, and the client stops at this
	// value rather than reading a count up front.
	entityDataTerminator = 0xff
)

// SetEntityDataClientboundPacket sets another player's stance: sneaking,
// sprinting, or neither. It is the metadata packet, but of everything metadata
// can carry this server only ever moves the flag byte and the pose, so that
// pair is the whole of what it encodes.
type SetEntityDataClientboundPacket struct {
	EntityId int32

	Sneaking  bool
	Sprinting bool
}

func (p *SetEntityDataClientboundPacket) String() string {
	return fmt.Sprintf("SetEntityDataClientboundPacket{EntityId:%d Sneaking:%t Sprinting:%t}",
		p.EntityId, p.Sneaking, p.Sprinting)
}

func (p *SetEntityDataClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	if err := ms.WriteVarInt(p.EntityId); err != nil {
		return err
	}

	// Each entry is its index, its serializer, and a value the serializer
	// shapes: the flag byte is a byte, the pose a var int over the ordinal.

	var flags byte
	if p.Sneaking {
		flags |= entityFlagSneaking
	}
	if p.Sprinting {
		flags |= entityFlagSprinting
	}

	if err := ms.WriteByte(entityFlagsIndex); err != nil {
		return err
	}

	if err := ms.WriteVarInt(entityFlagsSerializerId); err != nil {
		return err
	}

	if err := ms.WriteByte(flags); err != nil {
		return err
	}

	pose := int32(poseStanding)
	if p.Sneaking {
		pose = poseCrouching
	}

	if err := ms.WriteByte(entityPoseIndex); err != nil {
		return err
	}

	if err := ms.WriteVarInt(entityPoseSerializerId); err != nil {
		return err
	}

	if err := ms.WriteVarInt(pose); err != nil {
		return err
	}

	return ms.WriteByte(entityDataTerminator)
}
