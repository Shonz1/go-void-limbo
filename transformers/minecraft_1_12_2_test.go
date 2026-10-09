package transformers

import (
	"bytes"
	"math"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
)

// 1.12.1 reads the keep alive's id as a var int where 1.12.2 reads a long,
// and sends it back the same way: an id that fits goes down as a var int and
// comes back up as the long it was.
func TestKeepAliveCrossesThe1_12_2StepBothWays(t *testing.T) {
	for _, id := range []int64{1, -1, 0x7FFFFFFF, math.MinInt32, 1234567890} {
		sent := encodeBody(t, func(ms *streams.MinecraftStream) error { return ms.WriteLong(id) })
		want := encodeBody(t, func(ms *streams.MinecraftStream) error { return ms.WriteVarInt(int32(id)) })

		down := runTransformer(t, DowngradeKeepAliveTo1_12_1, sent)
		if !bytes.Equal(down, want) {
			t.Errorf("id %d to 1.12.1 = % x, want % x", id, down, want)
		}

		if up := runTransformer(t, UpgradeKeepAliveFrom1_12_1, down); !bytes.Equal(up, sent) {
			t.Errorf("id %d back from 1.12.1 = % x, want % x", id, up, sent)
		}
	}
}

// An id a var int cannot hold is refused rather than cut down to one the
// client would answer with a number matching nothing.
func TestDowngradeKeepAliveTo1_12_1RefusesAnIdTooWide(t *testing.T) {
	for _, id := range []int64{math.MaxInt32 + 1, math.MinInt32 - 1, 1_700_000_000_000} {
		sent := encodeBody(t, func(ms *streams.MinecraftStream) error { return ms.WriteLong(id) })

		if err := failingTransformer(t, DowngradeKeepAliveTo1_12_1, sent); err == nil {
			t.Errorf("id %d went to 1.12.1, want it refused", id)
		}
	}

	if err := failingTransformer(t, DowngradeKeepAliveTo1_12_1, []byte{0x00, 0x01}); err == nil {
		t.Error("a short body went to 1.12.1, want it refused")
	}
}
