package protocol

import (
	"bytes"
	clientboundCommon "github.com/Shonz1/go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "github.com/Shonz1/go-void-limbo/packets/clientbound/configuration"
	clientboundPlay "github.com/Shonz1/go-void-limbo/packets/clientbound/play"
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
	registry := NewDefaultRegistry(nil)
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
	registry := NewDefaultRegistry(nil)

	if _, err := registry.EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, nil); err == nil {
		t.Error("EncodeClientbound(nil) succeeded, want an error")
	}

	// A configuration packet has no id in the play phase.
	packet := clientboundConfiguration.NewRegistryDataClientboundPacket("minecraft:dimension_type", []byte{1})
	if _, err := registry.EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_26_2, packet); err == nil {
		t.Error("EncodeClientbound() of a configuration packet in play succeeded, want an error")
	}
}

// registryCodecs is the one thing a default registry is built from beyond
// its tables: what a play login before 1.20.2 carries the registries as,
// each version's own.
type registryCodecs map[types.ProtocolId][]byte

func (r registryCodecs) RegistryCodecFor(version types.ProtocolVersion) []byte {
	return r[version.ID]
}

// A play login before 1.20.2 carries the registries, which the packet
// encoded at the latest version has nothing of: they come from the source
// the registry was built with, each version's own -- a 1.19.4 login carries
// 1.19.4's and not 1.20's, which the chain writes in on the way down -- and
// a registry built without one refuses the login rather than send it without
// them. Every other version's login is untouched by the source, since none
// of them reads registries there.
func TestEncodeClientboundWritesTheRegistriesIntoALoginBefore1_20_2(t *testing.T) {
	codecs := registryCodecs{
		types.ProtocolVersions.MINECRAFT_1_19_4.ID: {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x61, 0x02, 0x00},
		types.ProtocolVersions.MINECRAFT_1_20.ID:   {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x62, 0x03, 0x00},
	}
	login := &clientboundPlay.LoginClientboundPacket{EntityId: 1, Dimensions: []string{"minecraft:overworld"}, SpawnInfo: clientboundPlay.SpawnInfo{Dimension: "minecraft:overworld"}}

	for _, version := range types.SupportedProtocolVersions[:2] {
		body, err := NewDefaultRegistry(codecs).EncodeClientbound(types.PhasePlay, version, login)
		if err != nil {
			t.Fatalf("protocol %d: EncodeClientbound() error: %v", version.ID, err)
		}

		if !bytes.Contains(body, codecs[version.ID]) {
			t.Errorf("protocol %d: the login % x does not carry its registries % x", version.ID, body, codecs[version.ID])
		}

		for other, codec := range codecs {
			if other != version.ID && bytes.Contains(body, codec) {
				t.Errorf("protocol %d: the login carries protocol %d's registries", version.ID, other)
			}
		}

		if _, err := NewDefaultRegistry(nil).EncodeClientbound(types.PhasePlay, version, login); err == nil {
			t.Errorf("protocol %d: EncodeClientbound() of a login with no registries succeeded, want a refusal", version.ID)
		}
	}

	// A registry with 1.20's registries and not 1.19.4's refuses the 1.19.4
	// login: the chain does not hand a version another's.
	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_20.ID: codecs[types.ProtocolVersions.MINECRAFT_1_20.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_19_4, login); err == nil {
		t.Error("EncodeClientbound() of a 1.19.4 login with only 1.20's registries succeeded, want a refusal")
	}

	codec := codecs

	for _, version := range types.SupportedProtocolVersions[2:] {
		with, err := NewDefaultRegistry(codec).EncodeClientbound(types.PhasePlay, version, login)
		if err != nil {
			t.Fatalf("protocol %d: EncodeClientbound() error: %v", version.ID, err)
		}

		without, err := NewDefaultRegistry(nil).EncodeClientbound(types.PhasePlay, version, login)
		if err != nil {
			t.Fatalf("protocol %d: EncodeClientbound() error: %v", version.ID, err)
		}

		if !bytes.Equal(with, without) {
			t.Errorf("protocol %d: the login differs with and without a registry codec, want it untouched", version.ID)
		}
	}
}
