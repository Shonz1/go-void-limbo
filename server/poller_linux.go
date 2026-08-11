//go:build linux

package server

import "syscall"

// epollET asks for edge-triggered delivery. The syscall package declares
// EPOLLET as a negative int, which does not fit the uint32 the event mask is,
// so the bit is named here instead.
const epollET = uint32(1) << 31

// poller is an epoll instance and the pipe that wakes it: the readiness
// questions the reactor's loop sleeps on, asked of the kernel directly rather
// than through the runtime's own poller, which wakes a goroutine per
// connection where this wakes one loop for all of them.
type poller struct {
	epfd int

	// wakeR and wakeW are the self-pipe: anything with something for the loop
	// writes a byte, and the loop finds the pipe readable among its events.
	wakeR int
	wakeW int

	// epevents backs every wait, sized once rather than per call.
	epevents []syscall.EpollEvent
}

// pollEvent is one thing the poller has to say: a connection's descriptor has
// bytes, or the wake pipe was written, which is the queue asking to be
// drained.
type pollEvent struct {
	fd   int
	wake bool
}

func newPoller() (*poller, error) {
	epfd, err := syscall.EpollCreate1(syscall.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}

	var pipe [2]int
	if err := syscall.Pipe(pipe[:]); err != nil {
		syscall.Close(epfd)
		return nil, err
	}

	// Non-blocking on both ends: the loop drains the read end without ever
	// waiting on it, and a wake finding the pipe full has nothing to add to
	// the wake already pending in it.
	syscall.SetNonblock(pipe[0], true)
	syscall.SetNonblock(pipe[1], true)

	// The wake pipe is level triggered on purpose: it stays readable until the
	// loop drains it, so a wake can never be lost between two waits.
	wake := syscall.EpollEvent{Events: syscall.EPOLLIN, Fd: int32(pipe[0])}

	if err := syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, pipe[0], &wake); err != nil {
		syscall.Close(epfd)
		syscall.Close(pipe[0])
		syscall.Close(pipe[1])

		return nil, err
	}

	return &poller{epfd: epfd, wakeR: pipe[0], wakeW: pipe[1], epevents: make([]syscall.EpollEvent, 128)}, nil
}

// arm asks to hear about bytes arriving on fd, edge triggered: the reactor
// drains a ready connection until the kernel has nothing more, so it only
// needs to hear about new bytes, never reminders about old ones.
func (p *poller) arm(fd int) error {
	event := syscall.EpollEvent{Events: syscall.EPOLLIN | epollET, Fd: int32(fd)}

	return syscall.EpollCtl(p.epfd, syscall.EPOLL_CTL_ADD, fd, &event)
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
	limit := min(len(events), len(p.epevents))

	n, err := syscall.EpollWait(p.epfd, p.epevents[:limit], -1)
	if err != nil {
		if err == syscall.EINTR {
			return 0, nil
		}

		return 0, err
	}

	for i := 0; i < n; i++ {
		fd := int(p.epevents[i].Fd)

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
