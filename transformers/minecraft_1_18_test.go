package transformers

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

// floorRegistryCodec and floorDimensionType stand in for what package
// gamedata hands the transformer for 1.17.1: the registries and the
// dimension type, told apart from 1.18's by their content.
func floorRegistryCodec(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"minecraft:worldgen/biome": nbt.Compound{"type": nbt.String("minecraft:worldgen/biome"), "value": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{nbt.Compound{"depth": nbt.Float(0.125)}}}}})
	})
}

func floorDimensionType(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"infiniburn": nbt.String("minecraft:infiniburn_overworld"), "height": nbt.Int(384)})
	})
}

// playLogin1_17_1 is the login as 1.17.1 lays it out: 1.18's without the
// simulation distance behind the view distance.
func playLogin1_17_1(t *testing.T, registryCodec []byte, dimensionType []byte) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteBoolean(true),
			ms.WriteByte(2),
			ms.WriteByte(0xFF),
			ms.WriteVarInt(2),
			ms.WriteString("minecraft:overworld"),
			ms.WriteString("minecraft:the_end"),
			ms.WriteBytes(registryCodec),
			ms.WriteBytes(dimensionType),
			ms.WriteString("minecraft:the_end"),
			ms.WriteLong(0x1122334455667788),
			ms.WriteVarInt(20),
			ms.WriteVarInt(8),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
			ms.WriteBoolean(true),
			ms.WriteBoolean(false),
		}

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// The login 1.17.1 reads is the login 1.18 reads with the registries and
// the spelled-out dimension type in the middle swapped for 1.17.1's own and
// the simulation distance gone. Everything else is copied.
func TestDowngradePlayLoginTo1_17_1DropsTheSimulationDistance(t *testing.T) {
	newerCodec, newerDimensionType := bottomRegistryCodec(t), bottomDimensionType(t)
	olderCodec, olderDimensionType := floorRegistryCodec(t), floorDimensionType(t)

	sent := playLogin1_18_2(t, newerCodec, newerDimensionType)
	got := runTransformer(t, DowngradePlayLoginTo1_17_1(olderCodec, olderDimensionType), sent)
	want := playLogin1_17_1(t, olderCodec, olderDimensionType)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.17.1 = % x\nwant = % x", got, want)
	}

	if bytes.Contains(got, newerCodec) {
		t.Error("the 1.17.1 login still carries 1.18's registries")
	}

	if bytes.Contains(got, newerDimensionType) {
		t.Error("the 1.17.1 login still spells out 1.18's dimension type")
	}
}

// The whole chain from 1.19 down: a 1.19 login walked to 1.18.2, to 1.18
// and then to 1.17.1 carries 1.17.1's registries and dimension type, and
// none of the other three versions'.
func TestDowngradePlayLoginTo1_17_1FollowsTheStepsAbove(t *testing.T) {
	codec1_19, codec1_18_2, codec1_18, codec1_17_1 := earliestRegistryCodec(t), lowestRegistryCodec(t), bottomRegistryCodec(t), floorRegistryCodec(t)
	dimensionType1_18_2, dimensionType1_18, dimensionType1_17_1 := lowestDimensionType(t), bottomDimensionType(t), floorDimensionType(t)

	sent := playLogin1_19_4(t, codec1_19, true)
	at1_18_2 := runTransformer(t, DowngradePlayLoginTo1_18_2(codec1_18_2, dimensionType1_18_2), sent)
	at1_18 := runTransformer(t, DowngradePlayLoginTo1_18(codec1_18, dimensionType1_18), at1_18_2)
	got := runTransformer(t, DowngradePlayLoginTo1_17_1(codec1_17_1, dimensionType1_17_1), at1_18)
	want := playLogin1_17_1(t, codec1_17_1, dimensionType1_17_1)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.17.1 = % x\nwant = % x", got, want)
	}

	for name, other := range map[string][]byte{"1.19's registries": codec1_19, "1.18.2's registries": codec1_18_2, "1.18's registries": codec1_18, "1.18.2's dimension type": dimensionType1_18_2, "1.18's dimension type": dimensionType1_18} {
		if bytes.Contains(got, other) {
			t.Errorf("the 1.17.1 login still carries %s", name)
		}
	}
}

// A 1.17.1 client reads the registries and the dimension type out of this
// packet and nothing else, so a transformer with either missing refuses the
// login rather than send a client into a world it cannot make sense of.
func TestDowngradePlayLoginTo1_17_1RefusesToSendNoRegistries(t *testing.T) {
	sent := playLogin1_18_2(t, bottomRegistryCodec(t), bottomDimensionType(t))

	if err := failingTransformer(t, DowngradePlayLoginTo1_17_1(nil, floorDimensionType(t)), sent); err == nil {
		t.Error("error = nil, want a refusal for a login with no registries to carry")
	}

	if err := failingTransformer(t, DowngradePlayLoginTo1_17_1(floorRegistryCodec(t), nil), sent); err == nil {
		t.Error("error = nil, want a refusal for a login with no dimension type to spell out")
	}
}

// A login cut short of its simulation distance is refused rather than
// written as if the field had been there to take off.
func TestDowngradePlayLoginTo1_17_1RefusesWhatItCannotWalk(t *testing.T) {
	sent := playLogin1_18_2(t, bottomRegistryCodec(t), bottomDimensionType(t))
	truncated := sent[:len(sent)-4-1-1]

	if err := failingTransformer(t, DowngradePlayLoginTo1_17_1(floorRegistryCodec(t), floorDimensionType(t)), truncated); err == nil {
		t.Error("error = nil, want a refusal for a login cut short of its simulation distance")
	}
}

// chunk1_18 is a chunk packet as the 1.18 step is handed it: encoded at the
// latest version and walked down through the 1.21.5, 1.20.2 and 1.20 steps,
// which put the heightmaps into a named compound and the trust edges flag
// in front of the light.
func chunk1_18(t *testing.T, packet *play.LevelChunkWithLightClientboundPacket) []byte {
	t.Helper()

	body := runTransformer(t, DowngradeLevelChunkWithLightTo1_21_4, encodeChunk(t, packet))
	body = runTransformer(t, DowngradeLevelChunkWithLightTo1_20, body)

	return runTransformer(t, DowngradeLevelChunkWithLightTo1_19_4, body)
}

// section1_18 is one section as 1.18 lays it out: the block count, a block
// container and a biome container, the containers' longs counted.
func section1_18(blockCount int, blocks []byte) []byte {
	section := []byte{byte(blockCount >> 8), byte(blockCount)}
	section = append(section, blocks...)

	// A biome container of one value over no longs.
	return append(section, 0x00, 0x00, 0x00)
}

// singleValueContainer is a block container of one state over no longs.
func singleValueContainer(state int32) []byte {
	return append(streams.AppendVarInt([]byte{0x00}, state), 0x00)
}

// The chunk 1.17.1 reads is the coordinates, a mask of the sections that
// hold a block, the heightmaps, one biome for every four blocks, the
// sections the mask names -- a palette of one spelled out at four bits over
// an array of zeros, every other section as it was -- and no block
// entities, with the light left off.
func TestDowngradeLevelChunkWithLightTo1_17_1MasksTheSectionsAndDropsTheLight(t *testing.T) {
	// A four bit palette of two entries over its 256 longs, as 1.18 and
	// 1.17.1 both lay it out.
	twoStates := []byte{0x04, 0x02, 0x00, 0x09}
	twoStates = streams.AppendVarInt(twoStates, 256)
	longs := make([]byte, 256*8)
	longs[7] = 0x01
	twoStates = append(twoStates, longs...)

	// Four sections: an empty one of a single value, one of a single state
	// throughout, one of two states, and an empty one with a palette that
	// names a block it never uses.
	var sections []byte
	sections = append(sections, section1_18(0, singleValueContainer(0))...)
	sections = append(sections, section1_18(4096, singleValueContainer(9))...)
	sections = append(sections, section1_18(1, twoStates)...)
	sections = append(sections, section1_18(0, twoStates)...)

	packet := &play.LevelChunkWithLightClientboundPacket{
		X: 1,
		Z: -2,
		Heightmaps: []play.Heightmap{
			{Type: play.HeightmapMotionBlocking, Data: []int64{65}},
		},
		SectionData: sections,
		LightData: play.LightData{
			SkyLightMask:        []int64{0b100000},
			EmptySkyLightMask:   []int64{0b011111},
			EmptyBlockLightMask: []int64{0b111111},
			SkyLight:            [][]byte{{0xFF, 0xEE}},
		},
	}

	got := runTransformer(t, DowngradeLevelChunkWithLightTo1_17_1, chunk1_18(t, packet))

	want := []byte{
		0x00, 0x00, 0x00, 0x01, // x
		0xff, 0xff, 0xff, 0xfe, // z
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0b0110, // the mask: sections 1 and 2
		0x0a, 0x00, 0x00, // the heightmaps, named as a root
		0x0c, 0x00, 0x0f,
	}
	want = append(want, "MOTION_BLOCKING"...)
	want = append(want,
		0x00, 0x00, 0x00, 0x01,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x41, // 65
		0x00, // the end of the compound
	)

	// The biomes: sixty-four per section, four sections.
	want = streams.AppendVarInt(want, 256)
	want = append(want, make([]byte, 256)...)

	// The sections: the single state at four bits over 256 zero longs, then
	// the two states as they were.
	wantSections := []byte{0x10, 0x00, 0x04, 0x01, 0x09}
	wantSections = streams.AppendVarInt(wantSections, 256)
	wantSections = append(wantSections, make([]byte, 256*8)...)
	wantSections = append(wantSections, 0x00, 0x01)
	wantSections = append(wantSections, twoStates...)

	want = streams.AppendVarInt(want, int32(len(wantSections)))
	want = append(want, wantSections...)
	want = append(want, 0x00) // no block entities

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.17.1 = % x\nwant      = % x", got, want)
	}
}

// A chunk whose sections are all empty is a mask of nothing and no section
// bytes, which is what a vanilla server of that version sends for such a
// chunk.
func TestDowngradeLevelChunkWithLightTo1_17_1SendsAnEmptyChunkAsNoSections(t *testing.T) {
	var sections []byte
	for range 24 {
		sections = append(sections, section1_18(0, singleValueContainer(0))...)
	}

	packet := &play.LevelChunkWithLightClientboundPacket{SectionData: sections}
	got := runTransformer(t, DowngradeLevelChunkWithLightTo1_17_1, chunk1_18(t, packet))

	want := []byte{
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // x and z
		0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, // the mask
		0x0a, 0x00, 0x00, 0x00, // the heightmaps, none
	}
	want = streams.AppendVarInt(want, 24*64)
	want = append(want, make([]byte, 24*64)...)
	want = append(want, 0x00, 0x00) // no section bytes, no block entities

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.17.1 = % x\nwant      = % x", got, want)
	}
}

// A section of ids packed directly, past the palettes, travels as it was:
// 1.17.1 packs them at the bits the registry's size needs as 1.18 does.
func TestDowngradeLevelChunkWithLightTo1_17_1KeepsAGlobalPalette(t *testing.T) {
	global := []byte{0x0f}
	global = streams.AppendVarInt(global, 1024)
	global = append(global, make([]byte, 1024*8)...)

	packet := &play.LevelChunkWithLightClientboundPacket{SectionData: section1_18(4096, global)}
	got := runTransformer(t, DowngradeLevelChunkWithLightTo1_17_1, chunk1_18(t, packet))

	wantSection := append([]byte{0x10, 0x00}, global...)
	wantTail := streams.AppendVarInt(nil, int32(len(wantSection)))
	wantTail = append(wantTail, wantSection...)
	wantTail = append(wantTail, 0x00)

	if !bytes.HasSuffix(got, wantTail) {
		t.Errorf("to 1.17.1 ends with % x, want the section as it was", got[len(got)-min(len(got), 16):])
	}

	if got[16] != 0x01 {
		t.Errorf("mask = %#x, want the one section", got[16])
	}
}

// This server sends no block entities, and 1.17.1 lays them out in a way a
// chunk carrying one cannot be rewritten into: such a chunk is refused.
func TestDowngradeLevelChunkWithLightTo1_17_1RefusesBlockEntities(t *testing.T) {
	body := chunk1_18(t, &play.LevelChunkWithLightClientboundPacket{SectionData: section1_18(0, singleValueContainer(0))})

	// The count sits right after the section buffer; the buffer is a
	// counted byte array of the one section.
	count := bytes.Index(body, append([]byte{byte(len(section1_18(0, singleValueContainer(0))))}, section1_18(0, singleValueContainer(0))...))
	if count < 0 {
		t.Fatal("the encoded body does not hold the section buffer where expected")
	}

	count += 1 + len(section1_18(0, singleValueContainer(0)))
	body[count] = 0x01

	err := failingTransformer(t, DowngradeLevelChunkWithLightTo1_17_1, body)
	if err == nil || !strings.Contains(err.Error(), "block entity") {
		t.Errorf("error = %v, want a refusal naming the block entities", err)
	}
}

// A section buffer that ends inside a section is refused rather than
// written as a chunk cut short.
func TestDowngradeLevelChunkWithLightTo1_17_1RefusesACutSection(t *testing.T) {
	packet := &play.LevelChunkWithLightClientboundPacket{SectionData: section1_18(4096, singleValueContainer(9))[:4]}

	if err := failingTransformer(t, DowngradeLevelChunkWithLightTo1_17_1, chunk1_18(t, packet)); err == nil {
		t.Error("error = nil, want a refusal for a section cut short")
	}
}
