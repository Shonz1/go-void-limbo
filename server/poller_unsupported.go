//go:build !darwin && !linux

package server

import "errors"

// poller has no implementation on this platform, so newPoller declines and the
// server serves every connection on a goroutine of its own, as it would with a
// poller for the connections that never reach play.
type poller struct{}

// pollEvent is one thing a poller would have to say; it exists so the reactor
// compiles everywhere even though no loop ever runs here.
type pollEvent struct {
	fd   int
	wake bool
}

var errNoPoller = errors.New("no poller on this platform")

func newPoller() (*poller, error) { return nil, errNoPoller }

func (p *poller) arm(int) error { return errNoPoller }

func (p *poller) wake() {}

func (p *poller) wait([]pollEvent) (int, error) { return 0, errNoPoller }
