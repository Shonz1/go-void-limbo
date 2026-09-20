package protocol

import (
	"bytes"
	clientboundCommon "github.com/Shonz1/go-void-limbo/packets/clientbound/common"
	clientboundConfiguration "github.com/Shonz1/go-void-limbo/packets/clientbound/configuration"
	clientboundPlay "github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
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

// dimensionTypes stands in for the dimension type a play login before 1.19
// spells out, which comes from the same source as the registries and only
// for those versions, each's own.
var dimensionTypes = map[types.ProtocolId][]byte{
	// 1.16.1's is a name rather than an entry, as the string it is written as.
	types.ProtocolVersions.MINECRAFT_1_16_1.ID: {0x07, 'o', 'l', 'd', ':', 'd', 'i', 'm'},
	types.ProtocolVersions.MINECRAFT_1_16_4.ID: {0x0A, 0x00, 0x00, 0x03, 0x00, 0x05, 'm', 'i', 'n', '_', 'y', 0xFF, 0xFF, 0xFF, 0xF8, 0x00},
	types.ProtocolVersions.MINECRAFT_1_17_1.ID: {0x0A, 0x00, 0x00, 0x03, 0x00, 0x05, 'm', 'i', 'n', '_', 'y', 0xFF, 0xFF, 0xFF, 0xF0, 0x00},
	types.ProtocolVersions.MINECRAFT_1_18.ID:   {0x0A, 0x00, 0x00, 0x03, 0x00, 0x05, 'm', 'i', 'n', '_', 'y', 0xFF, 0xFF, 0xFF, 0xE0, 0x00},
	types.ProtocolVersions.MINECRAFT_1_18_2.ID: {0x0A, 0x00, 0x00, 0x03, 0x00, 0x05, 'm', 'i', 'n', '_', 'y', 0xFF, 0xFF, 0xFF, 0xC0, 0x00},
}

func (r registryCodecs) DimensionTypeFor(version types.ProtocolVersion) []byte {
	if r[version.ID] != nil {
		return dimensionTypes[version.ID]
	}

	return nil
}

// A play login before 1.20.2 carries the registries, which the packet
// encoded at the latest version has nothing of: they come from the source
// the registry was built with, each version's own -- a 1.19.4 login carries
// 1.19.4's and not 1.20's, which the chain writes in on the way down, a
// 1.19.3 login 1.19.3's and neither of the others', a 1.19.1 login
// 1.19.1's, a 1.19 login 1.19's, a 1.18.2 login 1.18.2's, with the
// dimension type it spells out from the same source, a 1.18 login 1.18's,
// registries and dimension type both, a 1.17.1 login 1.17.1's, the
// same two, a 1.16.4 login 1.16.4's, and a 1.16.1 login 1.16.1's
// dimension types and the name of the one it is put into -- and a registry built
// without one refuses the login rather than send it without them. Every other version's login is untouched by the source, since none
// of them reads registries there.
func TestEncodeClientboundWritesTheRegistriesIntoALoginBefore1_20_2(t *testing.T) {
	codecs := registryCodecs{
		types.ProtocolVersions.MINECRAFT_1_16_1.ID: {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x59, 0x09, 0x00},
		types.ProtocolVersions.MINECRAFT_1_16_4.ID: {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x5A, 0x08, 0x00},
		types.ProtocolVersions.MINECRAFT_1_17_1.ID: {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x5B, 0x07, 0x00},
		types.ProtocolVersions.MINECRAFT_1_18.ID:   {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x5C, 0x06, 0x00},
		types.ProtocolVersions.MINECRAFT_1_18_2.ID: {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x5D, 0x05, 0x00},
		types.ProtocolVersions.MINECRAFT_1_19.ID:   {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x5E, 0x04, 0x00},
		types.ProtocolVersions.MINECRAFT_1_19_1.ID: {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x5F, 0x00, 0x00},
		types.ProtocolVersions.MINECRAFT_1_19_3.ID: {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x60, 0x01, 0x00},
		types.ProtocolVersions.MINECRAFT_1_19_4.ID: {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x61, 0x02, 0x00},
		types.ProtocolVersions.MINECRAFT_1_20.ID:   {0x0A, 0x00, 0x00, 0x01, 0x00, 0x01, 0x62, 0x03, 0x00},
	}
	login := &clientboundPlay.LoginClientboundPacket{EntityId: 1, Dimensions: []string{"minecraft:overworld"}, SpawnInfo: clientboundPlay.SpawnInfo{Dimension: "minecraft:overworld"}}

	for _, version := range types.SupportedProtocolVersions[6:20] {
		body, err := NewDefaultRegistry(codecs).EncodeClientbound(types.PhasePlay, version, login)
		if err != nil {
			t.Fatalf("protocol %d: EncodeClientbound() error: %v", version.ID, err)
		}

		// 1.17 reads 1.17.1's login as it stands, registries and dimension
		// type included: nothing rewrites the login on the step between them.
		// Nor does anything on the steps from 1.16.4 to 1.16.3 and on to
		// 1.16.2, or on the one from 1.16.1 to 1.16.
		source := version.ID
		switch source {
		case types.ProtocolVersions.MINECRAFT_1_17.ID:
			source = types.ProtocolVersions.MINECRAFT_1_17_1.ID
		case types.ProtocolVersions.MINECRAFT_1_16_2.ID, types.ProtocolVersions.MINECRAFT_1_16_3.ID:
			source = types.ProtocolVersions.MINECRAFT_1_16_4.ID
		case types.ProtocolVersions.MINECRAFT_1_16.ID:
			source = types.ProtocolVersions.MINECRAFT_1_16_1.ID
		}

		if !bytes.Contains(body, codecs[source]) {
			t.Errorf("protocol %d: the login % x does not carry its registries % x", version.ID, body, codecs[source])
		}

		// 1.18.2, 1.18, 1.17.1 and 1.16.4 spell the dimension type out behind the
		// registries, each its own, 1.16.1 names its own there, and no other
		// version does either.
		for other, dimensionType := range dimensionTypes {
			if spelled, want := bytes.Contains(body, dimensionType), source == other; spelled != want {
				t.Errorf("protocol %d: the login spells protocol %d's dimension type out: %t, want %t", version.ID, other, spelled, want)
			}
		}

		for other, codec := range codecs {
			if other != source && bytes.Contains(body, codec) {
				t.Errorf("protocol %d: the login carries protocol %d's registries", version.ID, other)
			}
		}

		if _, err := NewDefaultRegistry(nil).EncodeClientbound(types.PhasePlay, version, login); err == nil {
			t.Errorf("protocol %d: EncodeClientbound() of a login with no registries succeeded, want a refusal", version.ID)
		}
	}

	// A registry with 1.20's registries and not 1.19.4's refuses the 1.19.4
	// login, one with 1.19.4's and not 1.19.3's the 1.19.3 login, one with
	// 1.19.3's and not 1.19.1's the 1.19.1 login, one with 1.19.1's and not
	// 1.19's the 1.19 login, one with 1.19's and not 1.18.2's the 1.18.2
	// login, one with 1.18.2's and not 1.18's the 1.18 login, and one with
	// 1.18's and not 1.17.1's the 1.17.1 login, one with 1.17.1's and
	// not 1.16.4's the 1.16.4 login, and one with 1.16.4's and not 1.16.1's
	// the 1.16.1 login: the chain does not hand a version another's.
	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_20.ID: codecs[types.ProtocolVersions.MINECRAFT_1_20.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_19_4, login); err == nil {
		t.Error("EncodeClientbound() of a 1.19.4 login with only 1.20's registries succeeded, want a refusal")
	}

	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_19_4.ID: codecs[types.ProtocolVersions.MINECRAFT_1_19_4.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_19_3, login); err == nil {
		t.Error("EncodeClientbound() of a 1.19.3 login with only 1.19.4's registries succeeded, want a refusal")
	}

	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_19_3.ID: codecs[types.ProtocolVersions.MINECRAFT_1_19_3.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_19_1, login); err == nil {
		t.Error("EncodeClientbound() of a 1.19.1 login with only 1.19.3's registries succeeded, want a refusal")
	}

	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_19_1.ID: codecs[types.ProtocolVersions.MINECRAFT_1_19_1.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_19, login); err == nil {
		t.Error("EncodeClientbound() of a 1.19 login with only 1.19.1's registries succeeded, want a refusal")
	}

	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_19.ID: codecs[types.ProtocolVersions.MINECRAFT_1_19.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_18_2, login); err == nil {
		t.Error("EncodeClientbound() of a 1.18.2 login with only 1.19's registries succeeded, want a refusal")
	}

	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_18_2.ID: codecs[types.ProtocolVersions.MINECRAFT_1_18_2.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_18, login); err == nil {
		t.Error("EncodeClientbound() of a 1.18 login with only 1.18.2's registries succeeded, want a refusal")
	}

	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_18.ID: codecs[types.ProtocolVersions.MINECRAFT_1_18.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_17_1, login); err == nil {
		t.Error("EncodeClientbound() of a 1.17.1 login with only 1.18's registries succeeded, want a refusal")
	}

	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_17_1.ID: codecs[types.ProtocolVersions.MINECRAFT_1_17_1.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_16_4, login); err == nil {
		t.Error("EncodeClientbound() of a 1.16.4 login with only 1.17.1's registries succeeded, want a refusal")
	}

	if _, err := NewDefaultRegistry(registryCodecs{types.ProtocolVersions.MINECRAFT_1_16_4.ID: codecs[types.ProtocolVersions.MINECRAFT_1_16_4.ID]}).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_16_1, login); err == nil {
		t.Error("EncodeClientbound() of a 1.16.1 login with only 1.16.4's registries succeeded, want a refusal")
	}

	// 1.15.2 reads nothing of a registry out of its login, 1.15.1 and 1.15
	// below it read the same login, and 1.14.4, 1.14.3 and 1.14.2 that login less its
	// seed and its last flag, so theirs
	// carries no version's: not the registries, not a dimension type, not a
	// name. It comes down the same chain all the same, which refuses it
	// above without what 1.16.1's login is made of.
	for _, oldest := range types.SupportedProtocolVersions[:6] {
		body, err := NewDefaultRegistry(codecs).EncodeClientbound(types.PhasePlay, oldest, login)
		if err != nil {
			t.Fatalf("protocol %d: EncodeClientbound() error: %v", oldest.ID, err)
		}

		for other, codec := range codecs {
			if bytes.Contains(body, codec) {
				t.Errorf("protocol %d: the login carries protocol %d's registries", oldest.ID, other)
			}
		}

		for other, dimensionType := range dimensionTypes {
			if bytes.Contains(body, dimensionType) {
				t.Errorf("protocol %d: the login spells protocol %d's dimension type out", oldest.ID, other)
			}
		}

		if _, err := NewDefaultRegistry(nil).EncodeClientbound(types.PhasePlay, oldest, login); err == nil {
			t.Errorf("protocol %d: EncodeClientbound() of a login with no registries above it succeeded, want a refusal", oldest.ID)
		}
	}

	codec := codecs

	for _, version := range types.SupportedProtocolVersions[20:] {
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

// 1.17 alone takes an entity out of the world with the id and nothing else,
// where 1.17.1 and everything above it read a list: the packet keeps its id
// on the step between the two and loses its count.
func TestEncodeClientboundRemovesOneEntityToAPacketOn1_17(t *testing.T) {
	removal := &clientboundPlay.RemoveEntitiesClientboundPacket{EntityIds: []int32{128}}

	cases := []struct {
		version types.ProtocolVersion
		id      byte
		want    []byte
	}{
		// 1.16.4 reads the list as well, so the count the 1.17.1 step took
		// off goes back in front on the step below it.
		{types.ProtocolVersions.MINECRAFT_1_16_4, 0x36, []byte{0x01, 0x80, 0x01}},
		{types.ProtocolVersions.MINECRAFT_1_16_3, 0x36, []byte{0x01, 0x80, 0x01}},
		{types.ProtocolVersions.MINECRAFT_1_16_2, 0x36, []byte{0x01, 0x80, 0x01}},
		{types.ProtocolVersions.MINECRAFT_1_16_1, 0x37, []byte{0x01, 0x80, 0x01}},
		{types.ProtocolVersions.MINECRAFT_1_16, 0x37, []byte{0x01, 0x80, 0x01}},
		{types.ProtocolVersions.MINECRAFT_1_17, 0x3A, []byte{0x80, 0x01}},
		{types.ProtocolVersions.MINECRAFT_1_17_1, 0x3A, []byte{0x01, 0x80, 0x01}},
	}

	for _, c := range cases {
		body, err := NewDefaultRegistry(nil).EncodeClientbound(types.PhasePlay, c.version, removal)
		if err != nil {
			t.Fatalf("protocol %d: EncodeClientbound() error: %v", c.version.ID, err)
		}

		if !bytes.HasSuffix(body, c.want) || len(body) != len(c.want)+1 || body[0] != c.id {
			t.Errorf("protocol %d: the removal is % x, want the id %#x and % x", c.version.ID, body, c.id, c.want)
		}
	}

	several := &clientboundPlay.RemoveEntitiesClientboundPacket{EntityIds: []int32{1, 2}}
	if _, err := NewDefaultRegistry(nil).EncodeClientbound(types.PhasePlay, types.ProtocolVersions.MINECRAFT_1_17, several); err == nil {
		t.Error("EncodeClientbound() of two entities in one 1.17 removal succeeded, want a refusal")
	}
}

// 1.16.4 moved to 754 over nothing this server speaks: 1.16.3's jar numbers
// every packet as 1.16.4's does and lays every one of them out alike, so the
// tables give 753 the ids they give 754, packet for packet, and no
// transformer sits on the step between the two. 1.16.3 moved to 753 over
// less still -- its jar is 1.16.2's but for the version it names and five
// classes of the mob and its pathfinding -- so 751 stands to 753 as 753 does to 754.
// And 1.16.1 moved to 736 over two Realms screens, so 735 stands to 736 the
// same way, as 575 does to 578 and 573 to 575: neither 1.15.2's fixes nor
// 1.15.1's reach a packet or a class that reads or writes one. And 490 to
// 498: what 1.14.4 added is a packet this server never sends, registered
// behind every one it does. And 485 to 490: 1.14.3 added no packet and
// moved none, and changed the body of none this server sends or reads.
func TestProtocolsOnAnEmptyStepAreNumberedAndLaidOutAsTheOneAbove(t *testing.T) {
	steps := []struct{ older, newer types.ProtocolVersion }{
		{types.ProtocolVersions.MINECRAFT_1_14_2, types.ProtocolVersions.MINECRAFT_1_14_3},
		{types.ProtocolVersions.MINECRAFT_1_14_3, types.ProtocolVersions.MINECRAFT_1_14_4},
		{types.ProtocolVersions.MINECRAFT_1_15, types.ProtocolVersions.MINECRAFT_1_15_1},
		{types.ProtocolVersions.MINECRAFT_1_15_1, types.ProtocolVersions.MINECRAFT_1_15_2},
		{types.ProtocolVersions.MINECRAFT_1_16, types.ProtocolVersions.MINECRAFT_1_16_1},
		{types.ProtocolVersions.MINECRAFT_1_16_2, types.ProtocolVersions.MINECRAFT_1_16_3},
		{types.ProtocolVersions.MINECRAFT_1_16_3, types.ProtocolVersions.MINECRAFT_1_16_4},
	}

	for _, step := range steps {
		older, newer := step.older.ID, step.newer.ID
		olderName, newerName := step.older.Names[0], step.newer.Names[0]

		same := func(name string, ids packetIds) {
			olderId, olderOk := ids[older]
			newerId, newerOk := ids[newer]

			if olderOk != newerOk || olderId != newerId {
				t.Errorf("%s: %s has id %#x (%t) and %s %#x (%t), want the two alike", name, olderName, olderId, olderOk, newerName, newerId, newerOk)
			}
		}

		for _, packet := range serverboundPackets {
			same(packet.packet.Name(), packet.ids)
		}

		for _, packet := range clientboundPackets {
			same(packet.packet.Name(), packet.ids)
		}

		registry := NewDefaultRegistry(nil)
		for key := range registry.upgrades {
			if key.ProtocolID == older {
				t.Errorf("an upgrade of %s is registered from %s, want none: %s reads it as sent", key.PacketType.Name(), olderName, newerName)
			}
		}

		for key := range registry.downgrades {
			if key.ProtocolID == newer {
				t.Errorf("a downgrade of %s is registered at %s, want none: %s reads it as sent", key.PacketType.Name(), newerName, olderName)
			}
		}

		// What the chain makes of a packet is then the same bytes for both.
		removal := &clientboundPlay.RemoveEntitiesClientboundPacket{EntityIds: []int32{128}}

		olderBody, err := registry.EncodeClientbound(types.PhasePlay, step.older, removal)
		if err != nil {
			t.Fatalf("%s: EncodeClientbound() error: %v", olderName, err)
		}

		newerBody, err := registry.EncodeClientbound(types.PhasePlay, step.newer, removal)
		if err != nil {
			t.Fatalf("%s: EncodeClientbound() error: %v", newerName, err)
		}

		if !bytes.Equal(olderBody, newerBody) {
			t.Errorf("%s is sent % x and %s % x, want the same bytes", olderName, olderBody, newerName, newerBody)
		}
	}
}

// 1.16.1 numbers the play phase its own way, read off its jar's
// registrations: what this server sends between the chunk blocks update
// 1.16.2 retired, at 0x0F, and the section blocks update it added, at 0x3B,
// sits one higher, and the swing, behind the recipe book update 1.16.2 split
// in two, one lower. Everything else is numbered as on 1.16.2, and two
// packets are laid out differently, both on the way down.
func TestProtocol736IsNumberedAs751ButForTwoStretches(t *testing.T) {
	older, newer := types.ProtocolVersions.MINECRAFT_1_16_1.ID, types.ProtocolVersions.MINECRAFT_1_16_2.ID

	for _, packet := range serverboundPackets {
		olderId, olderOk := packet.ids[older]
		newerId, newerOk := packet.ids[newer]

		want := newerId
		if packet.phase == types.PhasePlay && newerId >= 0x1F {
			want--
		}

		if olderOk != newerOk || olderId != want {
			t.Errorf("%s: 1.16.1 has id %#x (%t), want %#x (%t)", packet.packet.Name(), olderId, olderOk, want, newerOk)
		}
	}

	for _, packet := range clientboundPackets {
		olderId, olderOk := packet.ids[older]
		newerId, newerOk := packet.ids[newer]

		want := newerId
		if packet.phase == types.PhasePlay && newerId >= 0x0F && newerId <= 0x3A {
			want++
		}

		if olderOk != newerOk || olderId != want {
			t.Errorf("%s: 1.16.1 has id %#x (%t), want %#x (%t)", packet.packet.Name(), olderId, olderOk, want, newerOk)
		}
	}

	// The ids the jar gives the packets that moved.
	moved := map[reflect.Type]types.PacketId{
		// The swing, under the name the latest version knows it by.
		reflect.TypeOf(serverboundPlay.PunchServerboundPacket{}):               0x2B,
		reflect.TypeOf(clientboundCommon.KeepAliveClientboundPacket{}):         0x20,
		reflect.TypeOf(clientboundPlay.LevelChunkWithLightClientboundPacket{}): 0x21,
		reflect.TypeOf(clientboundPlay.LightUpdateClientboundPacket{}):         0x24,
		reflect.TypeOf(clientboundPlay.LoginClientboundPacket{}):               0x25,
		reflect.TypeOf(clientboundPlay.PlayerPositionClientboundPacket{}):      0x35,
		reflect.TypeOf(clientboundPlay.RemoveEntitiesClientboundPacket{}):      0x37,
	}

	for _, packet := range serverboundPackets {
		if want, ok := moved[packet.packet]; ok && packet.phase == types.PhasePlay && packet.ids[older] != want {
			t.Errorf("%s: 1.16.1 has id %#x, want %#x", packet.packet.Name(), packet.ids[older], want)
		}
	}

	for _, packet := range clientboundPackets {
		if want, ok := moved[packet.packet]; ok && packet.phase == types.PhasePlay && packet.ids[older] != want {
			t.Errorf("%s: 1.16.1 has id %#x, want %#x", packet.packet.Name(), packet.ids[older], want)
		}
	}

	registry := NewDefaultRegistry(nil)
	for key := range registry.upgrades {
		if key.ProtocolID == older {
			t.Errorf("an upgrade of %s is registered from 1.16.1, want none: 1.16.2 reads it as sent", key.PacketType.Name())
		}
	}

	downgrades := 0
	for key := range registry.downgrades {
		if key.ProtocolID == newer {
			downgrades++
		}
	}

	if downgrades != 2 {
		t.Errorf("%d downgrades are registered at 1.16.2, want the login and the chunk", downgrades)
	}
}

// 1.15.2 numbers the play phase its own way, read off its jar's
// registrations. Of what this server reads, everything behind the jigsaw
// generate 1.16 added at 0x0F sits one lower. Of what it sends, everything
// from the add global entity 1.16 retired, at 0x02, sits one higher, up to
// the spawn position: 1.16 moved that from 0x4E to 0x42, so what lies between
// the two is numbered alike, and what lies behind them one higher again.
// Four packets are laid out differently, all on the way down.
func TestProtocol578IsNumberedAs735ButForTheStretchesAroundFourPackets(t *testing.T) {
	older, newer := types.ProtocolVersions.MINECRAFT_1_15_2.ID, types.ProtocolVersions.MINECRAFT_1_16.ID

	for _, packet := range serverboundPackets {
		olderId, olderOk := packet.ids[older]
		newerId, newerOk := packet.ids[newer]

		want := newerId
		if packet.phase == types.PhasePlay && newerId >= 0x10 {
			want--
		}

		if olderOk != newerOk || olderId != want {
			t.Errorf("%s: 1.15.2 has id %#x (%t), want %#x (%t)", packet.packet.Name(), olderId, olderOk, want, newerOk)
		}
	}

	for _, packet := range clientboundPackets {
		olderId, olderOk := packet.ids[older]
		newerId, newerOk := packet.ids[newer]

		want := newerId

		switch {
		case packet.phase != types.PhasePlay:
		case newerId == 0x42:
			want = 0x4E
		case newerId >= 0x02 && newerId < 0x42, newerId >= 0x4E:
			want++
		}

		if olderOk != newerOk || olderId != want {
			t.Errorf("%s: 1.15.2 has id %#x (%t), want %#x (%t)", packet.packet.Name(), olderId, olderOk, want, newerOk)
		}
	}

	registry := NewDefaultRegistry(nil)
	for key := range registry.upgrades {
		if key.ProtocolID == older {
			t.Errorf("an upgrade of %s is registered from 1.15.2, want none: 1.16 reads it as sent", key.PacketType.Name())
		}
	}

	downgrades := 0
	for key := range registry.downgrades {
		if key.ProtocolID == newer {
			downgrades++
		}
	}

	if downgrades != 4 {
		t.Errorf("%d downgrades are registered at 1.16, want the login success, the login, the chunk and the light", downgrades)
	}
}

// 1.14.4 numbers the play phase its own way as well, read off its jar's
// registrations. What this server reads is numbered alike: 1.15 added no
// serverbound packet and retired none. Of what it sends, everything behind
// 0x08 sits one lower: 1.14.4's block break acknowledgement is the last
// packet of the phase, which 1.15 moved up to 0x08. Three packets are laid
// out differently, all on the way down.
func TestProtocol498IsNumberedAs573ButBehindTheBlockBreakAcknowledgement(t *testing.T) {
	older, newer := types.ProtocolVersions.MINECRAFT_1_14_4.ID, types.ProtocolVersions.MINECRAFT_1_15.ID

	for _, packet := range serverboundPackets {
		olderId, olderOk := packet.ids[older]
		newerId, newerOk := packet.ids[newer]

		if olderOk != newerOk || olderId != newerId {
			t.Errorf("%s: 1.14.4 has id %#x (%t), want %#x (%t)", packet.packet.Name(), olderId, olderOk, newerId, newerOk)
		}
	}

	for _, packet := range clientboundPackets {
		olderId, olderOk := packet.ids[older]
		newerId, newerOk := packet.ids[newer]

		want := newerId
		if packet.phase == types.PhasePlay && newerId > 0x08 {
			want--
		}

		if olderOk != newerOk || olderId != want {
			t.Errorf("%s: 1.14.4 has id %#x (%t), want %#x (%t)", packet.packet.Name(), olderId, olderOk, want, newerOk)
		}
	}

	registry := NewDefaultRegistry(nil)
	for key := range registry.upgrades {
		if key.ProtocolID == older {
			t.Errorf("an upgrade of %s is registered from 1.14.4, want none: 1.15 reads it as sent", key.PacketType.Name())
		}
	}

	downgrades := 0
	for key := range registry.downgrades {
		if key.ProtocolID == newer {
			downgrades++
		}
	}

	if downgrades != 3 {
		t.Errorf("%d downgrades are registered at 1.15, want the login, the chunk and the add player", downgrades)
	}
}
