package types

import "testing"

func TestGetProtocolVersionById(t *testing.T) {
	cases := []struct {
		name string
		id   ProtocolId
		want ProtocolVersion
	}{
		{"zero", ProtocolVersions.ZERO.ID, ProtocolVersions.ZERO},
		{"minecraft_1_20_5", ProtocolVersions.MINECRAFT_1_20_5.ID, ProtocolVersions.MINECRAFT_1_20_5},
		{"minecraft_1_21", ProtocolVersions.MINECRAFT_1_21.ID, ProtocolVersions.MINECRAFT_1_21},
		{"minecraft_1_21_2", ProtocolVersions.MINECRAFT_1_21_2.ID, ProtocolVersions.MINECRAFT_1_21_2},
		{"minecraft_1_21_4", ProtocolVersions.MINECRAFT_1_21_4.ID, ProtocolVersions.MINECRAFT_1_21_4},
		{"minecraft_1_21_5", ProtocolVersions.MINECRAFT_1_21_5.ID, ProtocolVersions.MINECRAFT_1_21_5},
		{"minecraft_1_21_6", ProtocolVersions.MINECRAFT_1_21_6.ID, ProtocolVersions.MINECRAFT_1_21_6},
		{"minecraft_1_21_7", ProtocolVersions.MINECRAFT_1_21_7.ID, ProtocolVersions.MINECRAFT_1_21_7},
		{"minecraft_1_21_9", ProtocolVersions.MINECRAFT_1_21_9.ID, ProtocolVersions.MINECRAFT_1_21_9},
		{"minecraft_1_21_11", ProtocolVersions.MINECRAFT_1_21_11.ID, ProtocolVersions.MINECRAFT_1_21_11},
		{"minecraft_26_1", ProtocolVersions.MINECRAFT_26_1.ID, ProtocolVersions.MINECRAFT_26_1},
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

// The chain has to be in order, since the transformers registered for a step
// are looked up by the version on one side of it.
func TestSupportedProtocolVersionsAreOldestFirst(t *testing.T) {
	for i := 1; i < len(SupportedProtocolVersions); i++ {
		if SupportedProtocolVersions[i].ID <= SupportedProtocolVersions[i-1].ID {
			t.Errorf("protocol %d does not come after %d", SupportedProtocolVersions[i].ID, SupportedProtocolVersions[i-1].ID)
		}
	}

	if LatestProtocolVersion.ID != SupportedProtocolVersions[len(SupportedProtocolVersions)-1].ID {
		t.Errorf("the latest version is %d, which is not the end of the chain", LatestProtocolVersion.ID)
	}
}

func TestNextProtocolVersion(t *testing.T) {
	next, ok := NextProtocolVersion(ProtocolVersions.MINECRAFT_26_1)
	if !ok {
		t.Fatal("expected a version above 26.1")
	}

	if next.ID != ProtocolVersions.MINECRAFT_26_2.ID {
		t.Errorf("expected 26.2 above 26.1, got %d", next.ID)
	}

	if _, ok := NextProtocolVersion(LatestProtocolVersion); ok {
		t.Error("expected nothing above the latest version")
	}

	if _, ok := NextProtocolVersion(ProtocolVersions.ZERO); ok {
		t.Error("expected nothing above a version that is not on the chain")
	}
}

func TestPreviousProtocolVersion(t *testing.T) {
	previous, ok := PreviousProtocolVersion(ProtocolVersions.MINECRAFT_26_2)
	if !ok {
		t.Fatal("expected a version below 26.2")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_26_1.ID {
		t.Errorf("expected 26.1 below 26.2, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_26_1)
	if !ok {
		t.Fatal("expected a version below 26.1")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_21_11.ID {
		t.Errorf("expected 1.21.11 below 26.1, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_1_21_11)
	if !ok {
		t.Fatal("expected a version below 1.21.11")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_21_9.ID {
		t.Errorf("expected 1.21.9 below 1.21.11, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_1_21_9)
	if !ok {
		t.Fatal("expected a version below 1.21.9")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_21_7.ID {
		t.Errorf("expected 1.21.7 below 1.21.9, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_1_21_7)
	if !ok {
		t.Fatal("expected a version below 1.21.7")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_21_6.ID {
		t.Errorf("expected 1.21.6 below 1.21.7, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_1_21_6)
	if !ok {
		t.Fatal("expected a version below 1.21.6")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_21_5.ID {
		t.Errorf("expected 1.21.5 below 1.21.6, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_1_21_5)
	if !ok {
		t.Fatal("expected a version below 1.21.5")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_21_4.ID {
		t.Errorf("expected 1.21.4 below 1.21.5, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_1_21_4)
	if !ok {
		t.Fatal("expected a version below 1.21.4")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_21_2.ID {
		t.Errorf("expected 1.21.2 below 1.21.4, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_1_21_2)
	if !ok {
		t.Fatal("expected a version below 1.21.2")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_21.ID {
		t.Errorf("expected 1.21 below 1.21.2, got %d", previous.ID)
	}

	previous, ok = PreviousProtocolVersion(ProtocolVersions.MINECRAFT_1_21)
	if !ok {
		t.Fatal("expected a version below 1.21")
	}

	if previous.ID != ProtocolVersions.MINECRAFT_1_20_5.ID {
		t.Errorf("expected 1.20.5 below 1.21, got %d", previous.ID)
	}

	if _, ok := PreviousProtocolVersion(SupportedProtocolVersions[0]); ok {
		t.Error("expected nothing below the oldest version")
	}

	if _, ok := PreviousProtocolVersion(ProtocolVersions.ZERO); ok {
		t.Error("expected nothing below a version that is not on the chain")
	}
}

func TestIsSupportedProtocolVersion(t *testing.T) {
	for _, version := range SupportedProtocolVersions {
		if !IsSupportedProtocolVersion(version) {
			t.Errorf("protocol %d is on the chain and should be supported", version.ID)
		}
	}

	// ZERO is what a connection speaks before its handshake, which is not a
	// version anything is carried to or from.
	if IsSupportedProtocolVersion(ProtocolVersions.ZERO) {
		t.Error("expected the zero version not to be supported")
	}
}
