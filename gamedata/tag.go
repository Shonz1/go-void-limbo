package gamedata

import (
	"bytes"
	"fmt"
	"go-void-limbo/streams"
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

		if err := ms.WriteVarInt(int32(len(set.Tags))); err != nil {
			return nil, fmt.Errorf("tags for %s: %w", set.Registry, err)
		}

		for _, tag := range set.Tags {
			if err := ms.WriteString(tag.Name); err != nil {
				return nil, fmt.Errorf("tag %s: %w", tag.Name, err)
			}

			if err := ms.WriteVarInt(int32(len(tag.Entries))); err != nil {
				return nil, fmt.Errorf("tag %s: %w", tag.Name, err)
			}

			for _, id := range tag.Entries {
				if err := ms.WriteVarInt(id); err != nil {
					return nil, fmt.Errorf("tag %s: %w", tag.Name, err)
				}
			}
		}
	}

	if err := ms.Flush(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func countTags(sets []TagSet) int {
	total := 0
	for _, set := range sets {
		total += len(set.Tags)
	}

	return total
}
