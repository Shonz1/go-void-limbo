package gamedata

import (
	"bytes"
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
)

// NamedTag is one tag and the entries in it.
//
// Entries are registry ids, not names: indices into the registry the tag
// belongs to, which means a tag can only name entries the server actually sent.
// A limbo sends almost nothing, so its tags are almost all empty, which is
// fine. Being empty and not existing at all are different things to the client,
// and only the second one is a problem.
type NamedTag struct {
	Name    string
	Entries []int32
}

// TagSet is every tag for one registry.
type TagSet struct {
	Registry string
	Tags     []NamedTag
}

// encodeTags writes the Update Tags body: how many registries follow, then for
// each one its name, how many tags it has, and each tag's name and entry ids.
func encodeTags(sets []TagSet) ([]byte, error) {
	buf := new(bytes.Buffer)
	ms := streams.NewMinecraftStreamFromBuffer(buf)

	if err := ms.WriteVarInt(int32(len(sets))); err != nil {
		return nil, err
	}

	for _, set := range sets {
		if err := ms.WriteString(set.Registry); err != nil {
			return nil, fmt.Errorf("tags for %s: %w", set.Registry, err)
		}

		if err := writeTagRun(ms, set); err != nil {
			return nil, err
		}
	}

	if err := ms.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// unnamedTagRegistries is the registries a 1.16.4 client reads tags for, in the
// order it reads them in: see encodeTags1_16_4.
var unnamedTagRegistries = []string{"minecraft:block", "minecraft:item", "minecraft:fluid", "minecraft:entity_type"}

// encodeTags1_16_4 writes the Update Tags body as 1.16.4 reads it. 1.17 is
// where the packet came to name the registry in front of each run of tags,
// behind a count of the registries; before it the client reads four runs it
// knows by their position -- the blocks, the items, the fluids and the entity
// types -- each the count of its tags and then each tag's name and entry
// ids, as from 1.17 on. There is no saying which registry a run is for, so a
// set that is not those four in that order is refused rather than read as
// something else.
func encodeTags1_16_4(sets []TagSet) ([]byte, error) {
	if len(sets) != len(unnamedTagRegistries) {
		return nil, fmt.Errorf("a 1.16.4 client reads tags for %d registries, and the set holds %d", len(unnamedTagRegistries), len(sets))
	}

	for i, set := range sets {
		if set.Registry != unnamedTagRegistries[i] {
			return nil, fmt.Errorf("a 1.16.4 client reads the tags for %s at position %d, and the set holds %s there", unnamedTagRegistries[i], i, set.Registry)
		}
	}

	buf := new(bytes.Buffer)
	ms := streams.NewMinecraftStreamFromBuffer(buf)

	for _, set := range sets {
		if err := writeTagRun(ms, set); err != nil {
			return nil, err
		}
	}

	if err := ms.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// writeTagRun writes one registry's tags the way every version reads them:
// how many there are, and each tag's name and entry ids.
func writeTagRun(ms *streams.MinecraftStream, set TagSet) error {
	if err := ms.WriteVarInt(int32(len(set.Tags))); err != nil {
		return fmt.Errorf("tags for %s: %w", set.Registry, err)
	}

	for _, tag := range set.Tags {
		if err := ms.WriteString(tag.Name); err != nil {
			return fmt.Errorf("tag %s: %w", tag.Name, err)
		}

		if err := ms.WriteVarInt(int32(len(tag.Entries))); err != nil {
			return fmt.Errorf("tag %s: %w", tag.Name, err)
		}

		for _, id := range tag.Entries {
			if err := ms.WriteVarInt(id); err != nil {
				return fmt.Errorf("tag %s: %w", tag.Name, err)
			}
		}
	}

	return nil
}

func countTags(sets []TagSet) int {
	total := 0
	for _, set := range sets {
		total += len(set.Tags)
	}

	return total
}
