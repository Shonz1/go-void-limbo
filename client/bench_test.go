package client

import (
	"testing"

	clientboundCommon "github.com/Shonz1/go-void-limbo/packets/clientbound/common"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
)

// The benchmarks drive a keep alive both ways a connection does, since that is
// the packet a joined limbo client exchanges most. The deflated variants put
// the connection on a threshold of zero, which deflates every body, so the
// compression sits on the same packet rather than needing a bigger one.

func BenchmarkWritePacket(b *testing.B) {
	client, buf := newTestClient(types.PhasePlay)

	b.ReportAllocs()
	for b.Loop() {
		buf.Reset()

		if err := client.WritePacket(&clientboundCommon.KeepAliveClientboundPacket{Id: 1234}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWritePacketDeflated(b *testing.B) {
	client, buf := newTestClient(types.PhasePlay)
	enableCompressionOn(client, 0)

	b.ReportAllocs()
	for b.Loop() {
		buf.Reset()

		if err := client.WritePacket(&clientboundCommon.KeepAliveClientboundPacket{Id: 1234}); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadPacket(b *testing.B) {
	client, buf := newTestClient(types.PhasePlay)
	frame := append([]byte{byte(len(keepAliveBody))}, keepAliveBody...)

	b.ReportAllocs()
	for b.Loop() {
		buf.Write(frame)

		if _, _, err := client.ReadPacket(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadPacketDeflated(b *testing.B) {
	client, buf := newTestClient(types.PhasePlay)
	enableCompressionOn(client, 0)

	body, err := streams.CompressBody(keepAliveBody, 0)
	if err != nil {
		b.Fatal(err)
	}

	frame := append([]byte{byte(len(body))}, body...)

	b.ReportAllocs()
	for b.Loop() {
		buf.Write(frame)

		if _, _, err := client.ReadPacket(); err != nil {
			b.Fatal(err)
		}
	}
}
