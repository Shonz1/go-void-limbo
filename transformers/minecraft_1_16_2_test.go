package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// dimensionList and dimensionTypeName stand in for what package gamedata
// hands the transformer for 1.16.1: the list of the dimension types, and the
// name of the one the login puts the player into, as the string it is
// written as.
func dimensionList(t *testing.T) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		return nbt.WriteNamed(ms, "", nbt.Compound{"dimension": nbt.List{ElementType: nbt.TagCompound, Elements: []nbt.Tag{
			nbt.Compound{"name": nbt.String("minecraft:overworld"), "shrunk": nbt.Byte(0)},
		}}})
	})
}

func dimensionTypeName(t *testing.T) []byte {
	t.Helper()

	return encodeString(t, "minecraft:overworld")
}

// playLogin1_16_1 is the login as 1.16.1 lays it out, saying what
// playLogin1_17_1 says: the hardcore flag is the fourth bit of the game
// mode's byte, the dimension type is a name, and the most players is a byte.
func playLogin1_16_1(t *testing.T, gameMode byte, dimensionTypes []byte, dimensionTypeName []byte, maxPlayers byte) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteByte(gameMode),
			ms.WriteByte(0xFF),
			ms.WriteVarInt(2),
			ms.WriteString("minecraft:overworld"),
			ms.WriteString("minecraft:the_end"),
			ms.WriteBytes(dimensionTypes),
			ms.WriteBytes(dimensionTypeName),
			ms.WriteString("minecraft:the_end"),
			ms.WriteLong(0x1122334455667788),
			ms.WriteByte(maxPlayers),
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

// The login 1.16.1 reads is the login 1.16.2 reads with the hardcore flag
// folded into the game mode, the registries swapped for the list of the
// dimension types, the dimension type spelled out swapped for its name, and
// the most players in a byte. Everything else is copied.
func TestDowngradePlayLoginTo1_16_1FoldsTheFlagAndNamesTheDimensionType(t *testing.T) {
	// 1.16.2 lays its login out as 1.17.1 does: a hardcore player in
	// adventure mode, on a server of twenty.
	sent := playLogin1_17_1(t, flatRegistryCodec(t), flatDimensionType(t))
	got := runTransformer(t, DowngradePlayLoginTo1_16_1(dimensionList(t), dimensionTypeName(t)), sent)
	want := playLogin1_16_1(t, 2|8, dimensionList(t), dimensionTypeName(t), 20)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.16.1 = % x\nwant = % x", got, want)
	}
}

// login1_16_2 is a 1.16.2 login that says what the test asks of the three
// fields 1.16.1 lays out differently.
func login1_16_2(t *testing.T, hardcore bool, gameMode byte, maxPlayers int32) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{
			ms.WriteInt(7),
			ms.WriteBoolean(hardcore),
			ms.WriteByte(gameMode),
			ms.WriteByte(0xFF),
			ms.WriteVarInt(2),
			ms.WriteString("minecraft:overworld"),
			ms.WriteString("minecraft:the_end"),
			ms.WriteBytes(flatRegistryCodec(t)),
			ms.WriteBytes(flatDimensionType(t)),
			ms.WriteString("minecraft:the_end"),
			ms.WriteLong(0x1122334455667788),
			ms.WriteVarInt(maxPlayers),
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

// A game mode with no hardcore flag keeps its byte as it is, and a server
// that holds more players than a byte says is said to hold as many as one
// does.
func TestDowngradePlayLoginTo1_16_1HoldsMaxPlayersToAByte(t *testing.T) {
	got := runTransformer(t, DowngradePlayLoginTo1_16_1(dimensionList(t), dimensionTypeName(t)), login1_16_2(t, false, 1, 1000))
	want := playLogin1_16_1(t, 1, dimensionList(t), dimensionTypeName(t), 255)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.16.1 = % x\nwant = % x", got, want)
	}

	got = runTransformer(t, DowngradePlayLoginTo1_16_1(dimensionList(t), dimensionTypeName(t)), login1_16_2(t, false, 1, -1))
	want = playLogin1_16_1(t, 1, dimensionList(t), dimensionTypeName(t), 0)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.16.1 = % x\nwant = % x", got, want)
	}
}

func TestDowngradePlayLoginTo1_16_1RefusesAGameModeUnderTheHardcoreBit(t *testing.T) {
	if err := failingTransformer(t, DowngradePlayLoginTo1_16_1(dimensionList(t), dimensionTypeName(t)), login1_16_2(t, false, 8, 20)); err == nil {
		t.Error("expected a game mode that reads as the hardcore bit to be refused")
	}
}

func TestDowngradePlayLoginTo1_16_1RefusesWithoutItsData(t *testing.T) {
	sent := playLogin1_17_1(t, flatRegistryCodec(t), flatDimensionType(t))

	if err := failingTransformer(t, DowngradePlayLoginTo1_16_1(nil, dimensionTypeName(t)), sent); err == nil {
		t.Error("expected a login with no dimension types to write to be refused")
	}

	if err := failingTransformer(t, DowngradePlayLoginTo1_16_1(dimensionList(t), nil), sent); err == nil {
		t.Error("expected a login with no dimension type name to write to be refused")
	}
}

// chunk1_16_2 is a chunk as 1.16.2 lays it out, and chunk1_16_1 the same
// chunk as 1.16.1 does.
func chunk1_16_2(t *testing.T, wholeChunk bool, biomes []int32, sections []byte) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{ms.WriteInt(3), ms.WriteInt(-2), ms.WriteBoolean(wholeChunk), ms.WriteVarInt(0x8001), nbt.WriteNamed(ms, "", nbt.Compound{"MOTION_BLOCKING": heightmap(65)})}

		if wholeChunk {
			steps = append(steps, ms.WriteVarInt(int32(len(biomes))))
			for _, biome := range biomes {
				steps = append(steps, ms.WriteVarInt(biome))
			}
		}

		steps = append(steps, ms.WriteByteArray(sections), ms.WriteVarInt(0))

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

func chunk1_16_1(t *testing.T, wholeChunk bool, sections []byte) []byte {
	t.Helper()

	return encodeBody(t, func(ms *streams.MinecraftStream) error {
		steps := []error{ms.WriteInt(3), ms.WriteInt(-2), ms.WriteBoolean(wholeChunk), ms.WriteBoolean(wholeChunk), ms.WriteVarInt(0x8001), nbt.WriteNamed(ms, "", nbt.Compound{"MOTION_BLOCKING": heightmap(65)})}

		if wholeChunk {
			for range 1024 {
				steps = append(steps, ms.WriteInt(1))
			}
		}

		steps = append(steps, ms.WriteByteArray(sections), ms.WriteVarInt(0))

		for _, err := range steps {
			if err != nil {
				return err
			}
		}

		return nil
	})
}

// A 1.16.1 chunk is 1.16.2's with the forget old data flag behind the whole
// chunk flag, and the biomes as 1024 plain ints out of the client's own
// numbering, the plains at 1, with no count in front.
func TestDowngradeLevelChunkTo1_16_1AddsTheFlagAndSpellsTheBiomesOut(t *testing.T) {
	sections := append(section1_17(4096, 33), section1_17(12, 44)...)

	got := runTransformer(t, DowngradeLevelChunkTo1_16_1, chunk1_16_2(t, true, make([]int32, 1024), sections))
	want := chunk1_16_1(t, true, sections)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.16.1 = %d bytes, want %d", len(got), len(want))
	}
}

// A chunk that is a change to one the client holds carries no biomes on
// either side of the step, and forgets nothing.
func TestDowngradeLevelChunkTo1_16_1LeavesAPartialChunkWithoutBiomes(t *testing.T) {
	sections := section1_17(4096, 33)

	got := runTransformer(t, DowngradeLevelChunkTo1_16_1, chunk1_16_2(t, false, nil, sections))
	want := chunk1_16_1(t, false, sections)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.16.1 = % x\nwant = % x", got, want)
	}
}

func TestDowngradeLevelChunkTo1_16_1RefusesBiomesItCannotNumber(t *testing.T) {
	biomes := make([]int32, 1024)
	biomes[500] = 3

	if err := failingTransformer(t, DowngradeLevelChunkTo1_16_1, chunk1_16_2(t, true, biomes, nil)); err == nil {
		t.Error("expected a biome other than the one this server registers to be refused")
	}

	if err := failingTransformer(t, DowngradeLevelChunkTo1_16_1, chunk1_16_2(t, true, make([]int32, 64), nil)); err == nil {
		t.Error("expected a chunk of other than 1024 biomes to be refused")
	}
}
