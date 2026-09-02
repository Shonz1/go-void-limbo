package protocol

import (
	"bytes"
	clientboundCommon "github.com/Shonz1/go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "github.com/Shonz1/go-void-limbo/packets/clientbound/configuration"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"reflect"
	"testing"
)

type fakeServerboundPacket struct{}

func (p *fakeServerboundPacket) String() string { return "fakeServerboundPacket" }

type fakeClientboundPacket struct{}

func decodeFake(ms *streams.MinecraftStream) (types.ServerboundPacket, error) {
	return &fakeServerboundPacket{}, nil
}

func handleFake(client types.Client, packet types.ServerboundPacket) error {
	return nil
}

var (
	fakeServerboundType = reflect.TypeOf(fakeServerboundPacket{})
	fakeClientboundType = reflect.TypeOf(fakeClientboundPacket{})
)

func TestRegisterAndGetServerbound(t *testing.T) {
	r := NewRegistry()
	r.RegisterServerbound(types.PhaseLogin, fakeServerboundType, decodeFake, handleFake)
	r.RegisterServerboundId(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, fakeServerboundType, 0x00)

	packetType, ok := r.GetServerboundType(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x00)
	if !ok {
		t.Fatal("expected registered packet type, got none")
	}

	if packetType != fakeServerboundType {
		t.Errorf("expected %s, got %s", fakeServerboundType, packetType)
	}

	entry, ok := r.GetServerbound(types.PhaseLogin, packetType)
	if !ok {
		t.Fatal("expected registered entry, got none")
	}

	if entry.Decoder == nil {
		t.Error("expected registered decoder, got nil")
	}

	if entry.Handler == nil {
		t.Error("expected registered handler, got nil")
	}

	if _, ok := r.GetServerboundType(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, 0x01); ok {
		t.Error("expected no entry for unregistered packet id")
	}

	if _, ok := r.GetServerboundType(types.PhaseLogin, types.ProtocolVersions.ZERO, 0x00); ok {
		t.Error("expected no entry for unregistered protocol version")
	}

	if _, ok := r.GetServerboundType(types.PhaseStatus, types.ProtocolVersions.MINECRAFT_26_2, 0x00); ok {
		t.Error("expected no entry for unregistered phase")
	}
}

// A packet is implemented once and given an id per version, so the two versions
// may number it differently and still reach the same decoder.
func TestServerboundIdsAreVersionedAndImplementationIsNot(t *testing.T) {
	r := NewRegistry()
	r.RegisterServerbound(types.PhasePlay, fakeServerboundType, decodeFake, handleFake)
	r.RegisterServerboundId(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_1, fakeServerboundType, 0x11)
	r.RegisterServerboundId(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, fakeServerboundType, 0x22)

	for _, test := range []struct {
		version  types.ProtocolVersion
		packetId types.PacketId
	}{
		{types.ProtocolVersions.MINECRAFT_26_1, 0x11},
		{types.ProtocolVersions.MINECRAFT_26_2, 0x22},
	} {
		packetType, ok := r.GetServerboundType(types.PhasePlay, test.version, test.packetId)
		if !ok {
			t.Fatalf("protocol %d: expected id %#x to resolve", test.version.ID, test.packetId)
		}

		if packetType != fakeServerboundType {
			t.Errorf("protocol %d: expected %s, got %s", test.version.ID, fakeServerboundType, packetType)
		}
	}

	// The id one version gives a packet means nothing on the other.
	if _, ok := r.GetServerboundType(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_1, 0x22); ok {
		t.Error("expected 26.2's id to mean nothing on 26.1")
	}
}

func TestGetClientboundIdReturnsMinusOneForUnregisteredProtocolVersion(t *testing.T) {
	r := NewRegistry()
	r.RegisterClientboundId(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, fakeClientboundType, 0x00)

	id := r.GetClientboundId(types.PhaseLogin, fakeClientboundType, types.ProtocolVersions.ZERO)
	if id != -1 {
		t.Errorf("expected -1 for a registered phase+packet-type on an unregistered protocol version, got %d", id)
	}
}

func TestGetClientboundIdRoundTrip(t *testing.T) {
	r := NewRegistry()
	r.RegisterClientboundId(types.PhaseLogin, types.ProtocolVersions.MINECRAFT_26_2, fakeClientboundType, 0x05)

	id := r.GetClientboundId(types.PhaseLogin, fakeClientboundType, types.ProtocolVersions.MINECRAFT_26_2)
	if id != 0x05 {
		t.Errorf("expected 0x05, got %d", id)
	}

	if r.GetClientboundId(types.PhaseStatus, fakeClientboundType, types.ProtocolVersions.MINECRAFT_26_2) != -1 {
		t.Error("expected -1 for unregistered phase")
	}

	if r.GetClientboundId(types.PhaseLogin, fakeServerboundType, types.ProtocolVersions.MINECRAFT_26_2) != -1 {
		t.Error("expected -1 for unregistered packet type")
	}
}

// appendByte is a transformer that marks the body it was given, so that a test
// can tell how many steps a body was carried across and in what order.
func appendByte(marker byte) Transformer {
	return func(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
		body, err := in.ReadRest()
		if err != nil {
			return err
		}

		return out.WriteBytes(append(body, marker))
	}
}

func TestUpgradeBodyRunsTheStepsFromTheClientVersionUp(t *testing.T) {
	r := NewRegistry()
	r.RegisterUpgrade(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_1, fakeServerboundType, appendByte('a'))

	body, err := r.UpgradeBody(types.PhasePlay, fakeServerboundType, types.ProtocolVersions.MINECRAFT_26_1, []byte{0x01})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(body, []byte{0x01, 'a'}) {
		t.Errorf("expected the 26.1 step to have run, got %v", body)
	}

	// A client already on the latest version has nothing to be carried across.
	body, err = r.UpgradeBody(types.PhasePlay, fakeServerboundType, types.ProtocolVersions.MINECRAFT_26_2, []byte{0x01})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(body, []byte{0x01}) {
		t.Errorf("expected the body to be left alone, got %v", body)
	}
}

func TestDowngradeBodyRunsTheStepsFromTheLatestDown(t *testing.T) {
	r := NewRegistry()
	r.RegisterDowngrade(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, fakeClientboundType, appendByte('z'))

	body, err := r.DowngradeBody(types.PhasePlay, fakeClientboundType, types.ProtocolVersions.MINECRAFT_26_1, []byte{0x01})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(body, []byte{0x01, 'z'}) {
		t.Errorf("expected the 26.2 step to have run, got %v", body)
	}

	body, err = r.DowngradeBody(types.PhasePlay, fakeClientboundType, types.ProtocolVersions.MINECRAFT_26_2, []byte{0x01})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(body, []byte{0x01}) {
		t.Errorf("expected the body to be left alone, got %v", body)
	}
}

// A packet with nothing registered for a step crosses it untouched, which is
// what all but a handful do.
func TestBodyWithNoTransformerCrossesUntouched(t *testing.T) {
	r := NewRegistry()

	body, err := r.DowngradeBody(types.PhasePlay, fakeClientboundType, types.ProtocolVersions.MINECRAFT_26_1, []byte{0x01, 0x02})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(body, []byte{0x01, 0x02}) {
		t.Errorf("expected the body to be left alone, got %v", body)
	}
}

// The handshake is read before the connection has said what it speaks, so the
// version it is read at is not on the chain and nothing is carried anywhere.
func TestUnsupportedVersionIsLeftAlone(t *testing.T) {
	r := NewRegistry()
	r.RegisterUpgrade(types.PhaseHandshake, types.ProtocolVersions.MINECRAFT_26_1, fakeServerboundType, appendByte('a'))

	body, err := r.UpgradeBody(types.PhaseHandshake, fakeServerboundType, types.ProtocolVersions.ZERO, []byte{0x01})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !bytes.Equal(body, []byte{0x01}) {
		t.Errorf("expected the body to be left alone, got %v", body)
	}
}

func TestEncodeClientboundPutsTheVersionsIdInFrontOfTheBody(t *testing.T) {
	registry := NewDefaultRegistry()
	keepAlive := &clientboundCommon.KeepAliveClientboundPacket{Id: 1}

	latest, err := registry.EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, keepAlive)
	if err != nil {
		t.Fatalf("EncodeClientbound() error: %v", err)
	}

	older, err := registry.EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_20_2, keepAlive)
	if err != nil {
		t.Fatalf("EncodeClientbound() error: %v", err)
	}

	body := []byte{0, 0, 0, 0, 0, 0, 0, 1}

	for _, version := range []struct {
		version types.ProtocolVersion
		encoded []byte
	}{
		{types.ProtocolVersions.MINECRAFT_26_2, latest},
		{types.ProtocolVersions.MINECRAFT_1_20_2, older},
	} {
		wantId := registry.GetClientboundId(types.PhasePlay, reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}), version.version)
		want := append([]byte{byte(wantId)}, body...)

		if !bytes.Equal(version.encoded, want) {
			t.Errorf("protocol %d: encoded % x, want % x", version.version.ID, version.encoded, want)
		}
	}

	// The two versions number the keep alive differently, which is the whole
	// reason the id is resolved per version.
	if latest[0] == older[0] {
		t.Errorf("both versions encode the keep alive under id %#x, want the ids to differ", latest[0])
	}
}

func TestEncodeClientboundRefusesWhatTheVersionDoesNotCarry(t *testing.T) {
	registry := NewDefaultRegistry()

	if _, err := registry.EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, nil); err == nil {
		t.Error("EncodeClientbound(nil) succeeded, want an error")
	}

	// A configuration packet has no id in the play phase.
	packet := clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:dimension_type", []byte{1})
	if _, err := registry.EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, packet); err == nil {
		t.Error("EncodeClientbound() of a configuration packet in play succeeded, want an error")
	}
}
