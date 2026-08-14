package anvil

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/Shonz1/go-void-limbo/nbt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// Spawn is where the world puts a joining player, at the precision the world
// stores: a block.
type Spawn struct {
	X, Y, Z int32
}

// ReadSpawn reads the world spawn out of the level.dat in dir.
//
// Two shapes are read, because the format moved: modern saves keep a spawn
// compound holding a position array, and older ones kept three loose SpawnX,
// SpawnY, SpawnZ ints. A level.dat with neither is an error rather than an
// origin default, since a world exists precisely to say where things are.
func ReadSpawn(dir string) (Spawn, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "level.dat"))
	if err != nil {
		return Spawn{}, fmt.Errorf("anvil: %w", err)
	}

	reader, err := gzip.NewReader(bytes.NewReader(raw))
	if err != nil {
		return Spawn{}, fmt.Errorf("anvil: level.dat: %w", err)
	}
	defer reader.Close()

	payload, err := io.ReadAll(reader)
	if err != nil {
		return Spawn{}, fmt.Errorf("anvil: level.dat: %w", err)
	}

	_, root, err := nbt.ReadNamed(streams.NewMinecraftStreamFromBytesReader(bytes.NewReader(payload)))
	if err != nil {
		return Spawn{}, fmt.Errorf("anvil: level.dat: %w", err)
	}

	compound, ok := root.(nbt.Compound)
	if !ok {
		return Spawn{}, fmt.Errorf("anvil: level.dat root is %s, want compound", root.Type())
	}

	data, ok := compound["Data"].(nbt.Compound)
	if !ok {
		return Spawn{}, fmt.Errorf("anvil: level.dat has no Data compound")
	}

	if spawn, ok := data["spawn"].(nbt.Compound); ok {
		if pos, ok := spawn["pos"].(nbt.IntArray); ok && len(pos) == 3 {
			return Spawn{X: pos[0], Y: pos[1], Z: pos[2]}, nil
		}
	}

	if _, ok := data["SpawnX"]; ok {
		return Spawn{
			X: compoundInt(data, "SpawnX"),
			Y: compoundInt(data, "SpawnY"),
			Z: compoundInt(data, "SpawnZ"),
		}, nil
	}

	return Spawn{}, fmt.Errorf("anvil: level.dat holds no spawn")
}
