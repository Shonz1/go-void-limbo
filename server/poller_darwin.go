//go:build darwin

package server

import "syscall"

// poller is a kqueue and the pipe that wakes it: the readiness questions the
// reactor's loop sleeps on, asked of the kernel directly rather than through
// the runtime's own poller, which wakes a goroutine per connection where this
// wakes one loop for all of them.
type poller struct {
	kq int

	// wakeR and wakeW are the self-pipe: anything with something for the loop
	// writes a byte, and the loop finds the pipe readable among its events.
	wakeR int
	wakeW int

	// kevents backs every wait, sized once rather than per call.
	kevents []syscall.Kevent_t
}

// pollEvent is one thing the poller has to say: a connection's descriptor has
// bytes, or the wake pipe was written, which is the queue asking to be
// drained.
type pollEvent struct {
	fd   int
	wake bool
}

func newPoller() (*poller, error) {
	kq, err := syscall.Kqueue()
	if err != nil {
		return nil, err
	}

	var pipe [2]int
	if err := syscall.Pipe(pipe[:]); err != nil {
		syscall.Close(kq)
		return nil, err
	}

	// Non-blocking on both ends: the loop drains the read end without ever
	// waiting on it, and a wake finding the pipe full has nothing to add to
	// the wake already pending in it.
	syscall.SetNonblock(pipe[0], true)
	syscall.SetNonblock(pipe[1], true)

	// The wake pipe is level triggered on purpose: it stays readable until the
	// loop drains it, so a wake can never be lost between two waits.
	var kev syscall.Kevent_t
	syscall.SetKevent(&kev, pipe[0], syscall.EVFILT_READ, syscall.EV_ADD)

	if _, err := syscall.Kevent(kq, []syscall.Kevent_t{kev}, nil, nil); err != nil {
		syscall.Close(kq)
		syscall.Close(pipe[0])
		syscall.Close(pipe[1])

		return nil, err
	}

	return &poller{kq: kq, wakeR: pipe[0], wakeW: pipe[1], kevents: make([]syscall.Kevent_t, 128)}, nil
}

// arm asks to hear about bytes arriving on fd, edge triggered: the reactor
// drains a ready connection until the kernel has nothing more, so it only
// needs to hear about new bytes, never reminders about old ones.
func (p *poller) arm(fd int) error {
	var kev syscall.Kevent_t
	syscall.SetKevent(&kev, fd, syscall.EVFILT_READ, syscall.EV_ADD|syscall.EV_CLEAR)

	_, err := syscall.Kevent(p.kq, []syscall.Kevent_t{kev}, nil, nil)

	return err
}

// wake makes the next wait return with a wake event among its answers. Safe
// from any goroutine, which is the point of it.
func (p *poller) wake() {
	one := [1]byte{1}
	syscall.Write(p.wakeW, one[:])
}

// wait blocks until something is ready and fills events with what. A wait cut
// short by a signal is an empty answer rather than an error, so the loop asks
// again instead of deciding anything from it.
func (p *poller) wait(events []pollEvent) (int, error) {
	limit := min(len(events), len(p.kevents))

	n, err := syscall.Kevent(p.kq, nil, p.kevents[:limit], nil)
	if err != nil {
		if err == syscall.EINTR {
			return 0, nil
		}

		return 0, err
	}

	for i := 0; i < n; i++ {
		fd := int(p.kevents[i].Ident)

		if fd == p.wakeR {
			p.drainWake()
			events[i] = pollEvent{wake: true}

			continue
		}

		events[i] = pollEvent{fd: fd}
	}

	return n, nil
}

// drainWake empties the wake pipe, so the one byte a wake costs is also all a
// storm of them costs.
func (p *poller) drainWake() {
	var buf [64]byte

	for {
		n, _ := syscall.Read(p.wakeR, buf[:])
		if n < len(buf) {
			return
		}
	}
}
