package transformers

import (
	"bytes"
	"testing"
)

func TestUpgradePlayerCommandFrom1_21_5RenumbersTheActions(t *testing.T) {
	tests := []struct {
		name string
		from byte
		want byte
	}{
		{name: "press shift key", from: 0, want: 7},
		{name: "release shift key", from: 1, want: 8},
		{name: "stop sleeping", from: 2, want: 0},
		{name: "start sprinting", from: 3, want: 1},
		{name: "stop sprinting", from: 4, want: 2},
		{name: "start fall flying", from: 8, want: 6},
	}

	for _, test := range tests {
		// Entity id 300 as a two byte var int, the action, and the data.
		got := runTransformer(t, UpgradePlayerCommandFrom1_21_5, []byte{0xAC, 0x02, test.from, 0x05})
		want := []byte{0xAC, 0x02, test.want, 0x05}

		if !bytes.Equal(got, want) {
			t.Errorf("%s = % x, want % x", test.name, got, want)
		}
	}
}
