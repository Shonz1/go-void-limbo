package client

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"reflect"
	"time"

	clientboundCommon "github.com/Shonz1/go-void-limbo/packets/clientbound/common"
	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
	"github.com/Shonz1/go-void-limbo/types"
)

// keepAliveInterval is how often a keep alive goes out, and equally how long an
// unanswered one has before the connection is given up on.
//
// Both ends drop a connection they have read nothing from for thirty seconds,
// so the interval has to leave room for an answer to come back inside that
// window; fifteen seconds is what a vanilla server uses and is half of it.
const keepAliveInterval = 15 * time.Second

// readTimeout is how long a connection has to send its next packet in before it
// is given up on, and is the thirty seconds both ends already treat as a dead
// connection.
//
// Keep alives are what keep a joined client inside it without the player doing
// anything, which is why this has to be longer than the interval they go out on.
// The phases before them have no keep alive and nothing else that ends a quiet
// connection, so for a handshake, a status ping and a login this is the whole of
// it: a peer that opens a socket and sends nothing holds a goroutine, a ticker
// and the buffers behind them for exactly this long rather than for as long as
// the process runs.
const readTimeout = 2 * keepAliveInterval

// errKeepAliveTimeout is what a client that let a keep alive go unanswered for
// a whole interval leaves behind. Nothing on that connection is worth waiting
// for any longer, since the client either stopped reading or stopped existing.
var errKeepAliveTimeout = errors.New("keep alive went unanswered")

// packetError is a failure to make sense of a packet whose body was already
// read from the connection in full: an unknown id, or a decode that did not
// work out. The connection is still in sync and still usable, so the read loop
// reports these and carries on, unlike the read failures that end it.
type packetError struct {
	err error
}

func (e *packetError) Error() string { return e.err.Error() }

func (e *packetError) Unwrap() error { return e.err }

// packetLogBlacklist holds the packet types that are never logged, in either
// direction. A joined client sends some of these on every tick, and at that rate
// they bury everything else the log has to say.
var packetLogBlacklist = map[reflect.Type]bool{
	reflect.TypeOf(serverboundPlay.ClientTickEndServerboundPacket{}): true,
}

// logPacket records a packet crossing the connection, unless its type is
// blacklisted. Every connection carries the same traffic, so this is detail one
// asks for rather than detail one is told.
func logPacket(message string, packet any) {
	packetType := reflect.TypeOf(packet)
	if packetType != nil && packetType.Kind() == reflect.Pointer {
		packetType = packetType.Elem()
	}

	if packetType != nil && packetLogBlacklist[packetType] {
		return
	}

	slog.Debug(message, "packet", packet)
}

// ConfirmKeepAlive records the client's answer to the keep alive the server is
// waiting on.
func (c *Client) ConfirmKeepAlive(id int64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pendingKeepAlive == 0 {
		return fmt.Errorf("keep alive %d answers nothing that was sent", id)
	}

	if id != c.pendingKeepAlive {
		return fmt.Errorf("keep alive %d answers the wrong packet, expected %d", id, c.pendingKeepAlive)
	}

	c.pendingKeepAlive = 0

	return nil
}

// sendKeepAlive asks the client to prove it is still there, unless the last ask
// is still unanswered, which is errKeepAliveTimeout.
//
// Only configuration and play have a keep alive packet, so only they get one.
// The phases before them need none: a handshake, a status ping and a login are
// exchanges the client drives from one packet to the next, and a client that
// stops driving one has stopped connecting rather than gone quiet. What ends one
// that stopped is readTimeout, which every phase is under, rather than anything
// asked of the client here.
func (c *Client) sendKeepAlive() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.phase != types.PhaseConfiguration && c.phase != types.PhasePlay {
		return nil
	}

	if c.pendingKeepAlive != 0 {
		return errKeepAliveTimeout
	}

	// Any id the answer can be matched against works. The clock is what vanilla
	// uses, and it never repeats a value inside a connection.
	id := time.Now().UnixMilli()
	c.pendingKeepAlive = id

	return c.writePacket(&clientboundCommon.KeepAliveClientboundPacket{Id: id})
}

// keepAliveLoop sends a keep alive every interval until done is closed, and
// closes the connection when one is not answered or cannot be sent. Closing is
// what ends the read loop, which is what closes done.
func (c *Client) keepAliveLoop(done <-chan struct{}, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := c.sendKeepAlive(); err != nil {
				slog.Error("dropping connection", "addr", c.conn.RemoteAddr(), "err", err)
				c.conn.Close()

				return
			}
		}
	}
}

// Run drives the connection until it ends: keep alives go out on a clock of
// their own, and everything the client sends is read, decoded and handled here.
// It returns once the connection is done for any reason, with the player count
// no longer including this client.
func (c *Client) Run() {
	remoteAddr := c.conn.RemoteAddr().String()

	// A client that joined stops being counted when its connection ends, and one
	// that never joined was never counted.
	defer c.leavePlay()

	// A limbo has nothing to say to a client that has arrived, and thirty
	// seconds of having nothing to say is what both ends treat as a dead
	// connection. Keep alives are the something, and they go out on a clock of
	// their own rather than in reaction to what the client sends.
	done := make(chan struct{})
	defer close(done)

	go c.keepAliveLoop(done, keepAliveInterval)

	for {
		// Refreshed for every packet rather than set once, so the window is one
		// of silence rather than a cap on how long a connection may live. A
		// joined client refreshes it by answering keep alives; one that has not
		// got that far refreshes it by getting on with the exchange it opened.
		if err := c.conn.SetReadDeadline(time.Now().Add(readTimeout)); err != nil {
			slog.Error("failed to set the read deadline", "addr", remoteAddr, "err", err)
			return
		}

		packet, handler, err := c.ReadPacket()
		if err != nil {
			// A packet the server could not make sense of is one packet lost,
			// since its body was read in full and the next one starts where it
			// should. Anything else is the connection, including the close a
			// keep alive that went unanswered performs.
			var packetErr *packetError
			if errors.As(err, &packetErr) {
				slog.Error("failed to read packet", "err", err)
				continue
			}

			// A client that left, a connection this server closed on a keep
			// alive that went unanswered, and one that went quiet long enough to
			// run out its read window are all connections already accounted for.
			if !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) && !errors.Is(err, os.ErrDeadlineExceeded) {
				slog.Error("connection lost", "addr", remoteAddr, "err", err)
			}

			return
		}

		if handler == nil {
			continue
		}

		if err := handler(c, packet); err != nil {
			slog.Error("failed to handle packet", "packet", packet, "err", err)
		}
	}
}
