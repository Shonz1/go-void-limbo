package transformers

import (
	"bytes"
	"testing"

	"github.com/Shonz1/go-void-limbo/packets/clientbound/play"
	"github.com/Shonz1/go-void-limbo/streams"
)

func encodeGameEvent(t *testing.T, packet *play.GameEventClientboundPacket) []byte {
	t.Helper()

	buf := new(bytes.Buffer)
	stream := streams.NewMinecraftStreamFromBuffer(buf)

	if err := packet.Encode(stream); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	if err := stream.Flush(); err != nil {
		t.Fatalf("Flush() error = %v", err)
	}

	return buf.Bytes()
}

// 124 and 122 are both one byte var ints, so the body keeps its length and
// the result is what encoding the packet with 1.20.2's own number gives.
func TestDowngradeAddEntityTo1_20_2RenumbersThePlayer(t *testing.T) {
	packet := play.AddEntityClientboundPacket{
		EntityId:     5,
		Uuid:         "01020304-0506-0708-090a-0b0c0d0e0f10",
		EntityTypeId: playerEntityType1_20_3,
		X:            1,
		Y:            2,
		Z:            3,
	}

	sent := encodeAddEntity(t, &packet)
	got := runTransformer(t, DowngradeAddEntityTo1_20_2, sent)

	packet.EntityTypeId = playerEntityType1_20_2
	want := encodeAddEntity(t, &packet)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.20.2 = % x\nwant = % x", got, want)
	}

	if len(got) != len(sent) {
		t.Errorf("the body is %d bytes, want the %d sent", len(got), len(sent))
	}
}

// What goes out under 1.20.2's default spawn position id is that packet's
// body: a packed block position of zero as a long, then a zero angle as a
// float, and nothing of the event's own byte and float.
func TestDowngradeGameEventTo1_20_2BecomesADefaultSpawnPosition(t *testing.T) {
	sent := encodeGameEvent(t, &play.GameEventClientboundPacket{Event: play.GameEventStartWaitingForChunks})
	got := runTransformer(t, DowngradeGameEventTo1_20_2, sent)

	want := make([]byte, 12)

	if !bytes.Equal(got, want) {
		t.Errorf("to 1.20.2 = % x\nwant = % x", got, want)
	}
}

// The rewrite only means one thing, so an event that is not the one it was
// written for is refused rather than sent as a spawn position.
func TestDowngradeGameEventTo1_20_2RefusesAnyOtherEvent(t *testing.T) {
	sent := encodeGameEvent(t, &play.GameEventClientboundPacket{Event: play.GameEvent(3), Value: 1})

	if err := failingTransformer(t, DowngradeGameEventTo1_20_2, sent); err == nil {
		t.Error("error = nil, want a refusal for an event 1.20.2 has no spawn position for")
	}
}
