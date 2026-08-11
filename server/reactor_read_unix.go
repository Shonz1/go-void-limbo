//go:build darwin || linux

package server

import "syscall"

// readRawFd is one non-blocking read from a raw descriptor, split out by
// platform because the syscall package spells a descriptor differently
// elsewhere.
func readRawFd(fd uintptr, buf []byte) (int, error) {
	return syscall.Read(int(fd), buf)
}
