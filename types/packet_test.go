package types

import (
	"bytes"
	"strings"
	"testing"

	"github.com/Shonz1/go-void-limbo/streams"
)

func TestPrepareClientboundHoldsTheBodyDeflatedAndGivesItBackWhole(t *testing.T) {
	body := append([]byte{0x2C}, bytes.Repeat([]byte("section"), 200)...)

	prepared, err := PrepareClientbound(PhasePlay, ProtocolVersions.MINECRAFT_26_2, "LevelChunkWithLightClientboundPacket", body)
	if err != nil {
		t.Fatalf("PrepareClientbound() error: %v", err)
	}

	if prepared.Phase != PhasePlay || prepared.Version != ProtocolVersions.MINECRAFT_26_2.ID {
		t.Errorf("prepared for phase %d on protocol %d, want play on %d", prepared.Phase, prepared.Version, ProtocolVersions.MINECRAFT_26_2.ID)
	}

	if prepared.Size != int32(len(body)) {
		t.Errorf("Size = %d, want the %d bytes of the body", prepared.Size, len(body))
	}

	if len(prepared.Deflated) >= len(body) {
		t.Errorf("deflated %d bytes into %d, want fewer", len(body), len(prepared.Deflated))
	}

	// What is held is exactly what is held: a deflated body of a few hundred
	// bytes must not pin a buffer grown past it.
	if cap(prepared.Deflated) != len(prepared.Deflated) {
		t.Errorf("Deflated has capacity %d for %d bytes, want no spare", cap(prepared.Deflated), len(prepared.Deflated))
	}

	inflated, err := prepared.Body()
	if err != nil {
		t.Fatalf("Body() error: %v", err)
	}

	if !bytes.Equal(inflated, body) {
		t.Errorf("Body() = % x, want the body as prepared", inflated)
	}

	if got := prepared.String(); !strings.Contains(got, "LevelChunkWithLightClientboundPacket") || !strings.Contains(got, "776") {
		t.Errorf("String() = %q, want the name and the protocol", got)
	}

	if err := prepared.Encode(streams.NewMinecraftStreamFromBuffer(new(bytes.Buffer))); err == nil {
		t.Error("Encode() succeeded, want a refusal: a prepared packet is written whole")
	}
}
