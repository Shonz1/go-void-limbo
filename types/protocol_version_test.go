package types

import "testing"

func TestGetProtocolVersionById(t *testing.T) {
	cases := []struct {
		name string
		id   ProtocolId
		want ProtocolVersion
	}{
		{"zero", ProtocolVersions.ZERO.ID, ProtocolVersions.ZERO},
		{"minecraft_26_2", ProtocolVersions.MINECRAFT_26_2.ID, ProtocolVersions.MINECRAFT_26_2},
		{"unknown falls back to zero", 9999, ProtocolVersions.ZERO},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := GetProtocolVersionById(c.id)
			if got.ID != c.want.ID {
				t.Errorf("GetProtocolVersionById(%d) = %+v, want %+v", c.id, got, c.want)
			}
		})
	}
}
