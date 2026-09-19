package transformers

import (
	"bytes"
	"testing"
)

// A 1.17.1 removal is a count and the ids; a 1.17 one is the id alone.
func TestDowngradeRemoveEntitiesTo1_17IsTheIdAlone(t *testing.T) {
	got := runTransformer(t, DowngradeRemoveEntitiesTo1_17, []byte{0x01, 0x80, 0x01})

	if want := []byte{0x80, 0x01}; !bytes.Equal(got, want) {
		t.Errorf("to 1.17 = % x, want % x", got, want)
	}
}

func TestDowngradeRemoveEntitiesTo1_17RefusesAnythingButOneEntity(t *testing.T) {
	cases := map[string][]byte{
		"none":       {0x00},
		"two":        {0x02, 0x05, 0x06},
		"no id":      {0x01},
		"empty body": nil,
	}

	for name, body := range cases {
		if err := failingTransformer(t, DowngradeRemoveEntitiesTo1_17, body); err == nil {
			t.Errorf("%s: expected the packet to be refused", name)
		}
	}
}
