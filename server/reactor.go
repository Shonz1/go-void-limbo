package server

import (
	"crypto/cipher"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/Shonz1/go-void-limbo/client"
	"github.com/Shonz1/go-void-limbo/streams"
)

// reactor serves every joined connection from one goroutine and one poller,
// instead of a goroutine parked in a read each. A joined limbo client is
// almost all silence: a goroutine per connection pays a scheduler sleep and
// wake for every packet of it, and across thousands of connections that
// waking is most of what serving them costs. The connections that have not
// joined yet stay on goroutines of their own, where the login's back and
// forth is worth the straight-line read loop.
//
// The loop is one goroutine on purpose. Everything a joined client sends is
// decoded and either dropped or answered by mutating its own state, so there
// is no work worth spreading, and one loop means the descriptor maps need no
// lock. The one rule that keeps them honest: only this loop closes a
// connection it holds, always unmapping first, so a descriptor number the
// kernel reuses can never wear an old connection's state.
type reactor struct {
	server *Server
	poller *poller

	// queue, under queueMu, is how every other goroutine reaches the loop:
	// connections to take over, and connections to close. The poller's wake is
	// what makes the loop look.
	queueMu sync.Mutex
	queue   []reactorCommand

	// byFd and byClient belong to the loop goroutine alone.
	byFd     map[int]*reactorConn
	byClient map[*client.Client]*reactorConn

	// readBuf is the one buffer every ready connection is read into before its
	// bytes are carried off, shared because the loop serves one connection at
	// a time.
	readBuf []byte
}

// reactorCommand is one thing asked of the loop: a takeover when conn is set,
// and a close when only the client is.
type reactorCommand struct {
	c *client.Client

	conn      net.Conn
	raw       syscall.RawConn
	residue   []byte
	decrypter cipher.Stream
}

// reactorConn is one connection the loop holds: who it belongs to, the raw
// descriptor it is read through, the cipher its bytes are under when it has
// one, and however much of the next frame has arrived.
type reactorConn struct {
	c         *client.Client
	conn      net.Conn
	raw       syscall.RawConn
	fd        int
	decrypter cipher.Stream

	pending []byte
}

// pendingKeepLimit is the biggest buffer a connection keeps once it has
// nothing waiting in it. What usually remains between reads is nothing or the
// first bytes of the next frame, so a buffer grown for one large frame is
// returned rather than held for the life of the connection.
const pendingKeepLimit = 16384

func newReactor(s *Server) (*reactor, error) {
	p, err := newPoller()
	if err != nil {
		return nil, err
	}

	return &reactor{
		server:   s,
		poller:   p,
		byFd:     make(map[int]*reactorConn),
		byClient: make(map[*client.Client]*reactorConn),
		readBuf:  make([]byte, 65536),
	}, nil
}

// take offers the reactor a connection whose client has reached play, and
// reports whether it was taken. A nil reactor and a connection without a real
// descriptor -- the in-memory pipes the tests serve -- both decline, and the
// read loop that offered carries on serving the connection itself.
func (r *reactor) take(c *client.Client, conn net.Conn) bool {
	if r == nil {
		return false
	}

	sc, ok := conn.(syscall.Conn)
	if !ok {
		return false
	}

	raw, err := sc.SyscallConn()
	if err != nil {
		return false
	}

	residue, decrypter, err := c.TakeoverRead()
	if err != nil {
		return false
	}

	// The read deadline was the blocking loop's watchdog, and reads bypass it
	// from here. What ends a quiet connection now is the keep alive sweep: one
	// it cannot reach, or that stops answering, is closed within two
	// intervals, the same thirty seconds the deadline enforced.
	conn.SetReadDeadline(time.Time{})

	r.enqueue(reactorCommand{c: c, conn: conn, raw: raw, residue: residue, decrypter: decrypter})

	return true
}

// dropClient asks the loop to close c's connection. It is how everything
// outside the loop -- the keep alive sweep -- ends a connection that may be
// polled: the loop unmaps before it closes, and a client it never held is
// simply closed there instead, so asking is always safe.
func (r *reactor) dropClient(c *client.Client) {
	r.enqueue(reactorCommand{c: c})
}

func (r *reactor) enqueue(cmd reactorCommand) {
	r.queueMu.Lock()
	r.queue = append(r.queue, cmd)
	r.queueMu.Unlock()

	r.poller.wake()
}

// run is the loop: wait for something to be ready, serve it, wait again. It
// runs for the life of the server.
func (r *reactor) run() {
	events := make([]pollEvent, 128)

	for {
		n, err := r.poller.wait(events)
		if err != nil {
			slog.Error("the reactor's poller failed", "err", err)
			return
		}

		for _, event := range events[:n] {
			if event.wake {
				r.drainQueue()
				continue
			}

			// A descriptor with no connection behind it was dropped after this
			// event was already reported; there is nothing left to serve.
			if rc, ok := r.byFd[event.fd]; ok {
				r.readReady(rc)
			}
		}
	}
}

func (r *reactor) drainQueue() {
	r.queueMu.Lock()
	commands := r.queue
	r.queue = nil
	r.queueMu.Unlock()

	for _, cmd := range commands {
		if cmd.conn != nil {
			r.register(cmd)
			continue
		}

		if rc, ok := r.byClient[cmd.c]; ok {
			r.drop(rc, nil)
			continue
		}

		// A client this loop never held, or let go of already: its connection
		// is closed directly, and whoever is serving it sees the close as
		// ever.
		cmd.c.Close()
	}
}

// register starts polling a connection the takeover handed over, and serves
// whatever arrived around the handover before waiting on anything new.
func (r *reactor) register(cmd reactorCommand) {
	rc := &reactorConn{
		c:         cmd.c,
		conn:      cmd.conn,
		raw:       cmd.raw,
		decrypter: cmd.decrypter,
		pending:   cmd.residue,
	}

	if err := cmd.raw.Control(func(fd uintptr) { rc.fd = int(fd) }); err != nil {
		r.drop(rc, err)
		return
	}

	if err := r.poller.arm(rc.fd); err != nil {
		r.drop(rc, err)
		return
	}

	r.byFd[rc.fd] = rc
	r.byClient[rc.c] = rc

	// The residue may already hold whole frames, and bytes may have arrived
	// between the takeover and the arming. Both are served now, rather than
	// waiting on bytes that may never follow them.
	if r.processPending(rc) {
		return
	}

	r.readReady(rc)
}

// drop ends a connection: the maps let go of it first, the connection is
// closed, and the server stops counting it. err is what ended it, when
// anything worth a word did.
func (r *reactor) drop(rc *reactorConn, err error) {
	delete(r.byFd, rc.fd)
	delete(r.byClient, rc.c)

	rc.conn.Close()
	r.server.removeClient(rc.c)
	rc.c.LeavePlay()

	// A client that left and a connection something else already closed are
	// both connections accounted for, same as the read loop treats them.
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, net.ErrClosed) && !errors.Is(err, syscall.ECONNRESET) {
		slog.Error("connection lost", "addr", rc.conn.RemoteAddr(), "err", err)
	}
}

// readReady drains a connection until the kernel has nothing more for it,
// which is what edge-triggered polling requires, and what batches a burst of
// packets into one wakeup.
func (r *reactor) readReady(rc *reactorConn) {
	for {
		n, err := rc.read(r.readBuf)

		if n > 0 {
			chunk := r.readBuf[:n]

			// The cipher deals in one running stream, and this is the one
			// place the stream arrives.
			if rc.decrypter != nil {
				rc.decrypter.XORKeyStream(chunk, chunk)
			}

			rc.pending = append(rc.pending, chunk...)

			if r.processPending(rc) {
				return
			}

			continue
		}

		switch {
		case errors.Is(err, syscall.EAGAIN):
			return
		case errors.Is(err, syscall.EINTR):
			continue
		case err == nil:
			// Zero bytes and no error is the other end done writing.
			r.drop(rc, io.EOF)
			return
		default:
			r.drop(rc, err)
			return
		}
	}
}

// read pulls what the kernel holds for this connection, through the raw
// descriptor so no goroutine ever parks on it. Every socket the runtime hands
// out is already non-blocking, so an empty kernel buffer is EAGAIN rather
// than a wait.
func (rc *reactorConn) read(buf []byte) (int, error) {
	var n int
	var err error

	rawErr := rc.raw.Read(func(fd uintptr) bool {
		n, err = readRawFd(fd, buf)

		// True either way: returning false would hand the wait back to the
		// runtime's poller, which is the wait this loop exists to replace.
		return true
	})

	if n < 0 {
		n = 0
	}

	if err == nil && rawErr != nil {
		err = rawErr
	}

	return n, err
}

// processPending serves every whole frame sitting in a connection's buffer,
// and reports whether doing so ended the connection. What remains afterwards
// is moved to the front, so the buffer only ever holds the head of a frame
// still arriving.
func (r *reactor) processPending(rc *reactorConn) bool {
	consumed := 0

	for {
		length, size, err := streams.ReadVarIntFrom(rc.pending[consumed:])
		if err != nil {
			// A length cut off mid-byte is a frame still arriving; anything
			// else is a stream this end can no longer find frames in.
			if errors.Is(err, io.EOF) {
				break
			}

			r.drop(rc, err)

			return true
		}

		if length < 1 || length > streams.MaxPacketSize {
			r.drop(rc, fmt.Errorf("invalid packet length: %d", length))

			return true
		}

		frame := rc.pending[consumed+size:]
		if len(frame) < int(length) {
			break
		}

		if err := rc.c.HandlePlayBody(frame[:length]); err != nil {
			r.drop(rc, err)

			return true
		}

		consumed += size + int(length)
	}

	if consumed > 0 {
		// Overlapping append is a memmove, so this is the leftover walked to
		// the front rather than copied aside and back.
		rc.pending = append(rc.pending[:0], rc.pending[consumed:]...)
	}

	if len(rc.pending) == 0 && cap(rc.pending) > pendingKeepLimit {
		rc.pending = nil
	}

	return false
}
