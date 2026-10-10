package transformers

import (
	"fmt"
	"math"

	"github.com/Shonz1/go-void-limbo/streams"
)

// The 1.12.2 step is where the keep alive's id became a long. 1.12.1 below
// it reads a var int there and sends one back, and 1.12, 1.11.1, 1.11, 1.10,
// 1.9.3, 1.9.2, 1.9.1 and 1.9 below that read and send the same packet, so what goes to
// 1.12.1 goes on to them as it is. The keep alive is the whole of what
// 1.12.1 and 1.12.2 differ in on the wire: of the fifty classes 1.12.2's jar
// changed, three are packets -- the two keep alives, and the handshake for
// the number in it -- and every phase registers the same packets in the same
// order. The chunk, the blocks, the tags that are not there and the login
// that cannot be asked are 1.12.2's: see the 1.13 step.

// DowngradeKeepAliveTo1_12_1 rewrites the keep alive from what 1.12.2 reads
// into what 1.12.1 reads: the id, a long, as a var int.
//
// A var int holds thirty-two bits, and the id is whatever this server chose,
// so an id a var int cannot hold is refused rather than cut down: the answer
// is matched against the id as sent, and a cut-down id would be answered
// with a number that matches nothing. The server picks one that fits for
// every version, so this never refuses in practice: see client.SendKeepAlive.
func DowngradeKeepAliveTo1_12_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	id, err := in.ReadLong()
	if err != nil {
		return err
	}

	if id < math.MinInt32 || id > math.MaxInt32 {
		return fmt.Errorf("keep alive id %d, which 1.12.1 has no var int for", id)
	}

	return out.WriteVarInt(int32(id))
}

// UpgradeKeepAliveFrom1_12_1 rewrites the keep alive from what 1.12.1 sends
// into what 1.12.2 sends: the id, a var int, as a long. A 1.12.1 client
// sends back the id it was given, which went out as the low thirty-two bits
// of a long those bits fit, so widening it with its sign is the id as it was
// sent, and the answer matches what is waited on.
func UpgradeKeepAliveFrom1_12_1(in *streams.MinecraftStream, out *streams.MinecraftStream) error {
	id, err := in.ReadVarInt()
	if err != nil {
		return err
	}

	return out.WriteLong(int64(id))
}
