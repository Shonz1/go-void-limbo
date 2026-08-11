//go:build !darwin && !linux

package server

// readRawFd never runs here: without a poller no connection is ever taken
// over, so this exists only so the reactor compiles everywhere.
func readRawFd(uintptr, []byte) (int, error) { return 0, errNoPoller }
