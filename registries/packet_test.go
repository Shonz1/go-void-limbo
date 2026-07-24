package registries

import (
	"go-void-limbo/streams"
	"go-void-limbo/types"
	"reflect"
	"testing"
)

type fakeServerboundPacket struct{}

func (p *fakeServerboundPacket) ToString() string { return "fakeServerboundPacket" }

type fakeClientboundPacket struct{}

func decodeFake(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	return &fakeServerboundPacket{}, nil
}

func TestRegisterAndGetServerbound(t *testing.T) {
	r := NewPacketRegistry()
	r.RegisterServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x00, decodeFake)

	decoder := r.GetServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x00)
	if decoder == nil {
		t.Fatal("expected registered decoder, got nil")
	}

	if r.GetServerbound(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x01) != nil {
		t.Error("expected nil for unregistered packet id")
	}

	if r.GetServerbound(types.PhaseLogin, types.ProtocolVersions.ZERO, 0x00) != nil {
		t.Error("expected nil for unregistered protocol version")
	}

	if r.GetServerbound(types.PhaseStatus, types.ProtocolVersions.MINECRAFT_26_2, 0x00) != nil {
		t.Error("expected nil for unregistered phase")
	}
}

func TestGetClientboundIdReturnsMinusOneForUnregisteredProtocolVersion(t *testing.T) {
	r := NewPacketRegistry()
	packetType := reflect.TypeOf(fakeClientboundPacket{})
	r.RegisterClientbound(types.PhaseLogin, packetType, types.ProtocolVersions.MINECRAFT_26_2, 0x00)

	id := r.GetClientboundId(types.PhaseLogin, packetType, types.ProtocolVersions.ZERO)
	if id != -1 {
		t.Errorf("expected -1 for a registered phase+packet-type on an unregistered protocol version, got %d", id)
	}
}

func TestGetClientboundIdRoundTrip(t *testing.T) {
	r := NewPacketRegistry()
	packetType := reflect.TypeOf(fakeClientboundPacket{})
	r.RegisterClientbound(types.PhaseLogin, packetType, types.ProtocolVersions.MINECRAFT_26_2, 0x05)

	id := r.GetClientboundId(types.PhaseLogin, packetType, types.ProtocolVersions.MINECRAFT_26_2)
	if id != 0x05 {
		t.Errorf("expected 0x05, got %d", id)
	}

	if r.GetClientboundId(types.PhaseStatus, packetType, types.ProtocolVersions.MINECRAFT_26_2) != -1 {
		t.Error("expected -1 for unregistered phase")
	}

	if r.GetClientboundId(types.PhaseLogin, reflect.TypeOf(fakeServerboundPacket{}), types.ProtocolVersions.MINECRAFT_26_2) != -1 {
		t.Error("expected -1 for unregistered packet type")
	}
}
