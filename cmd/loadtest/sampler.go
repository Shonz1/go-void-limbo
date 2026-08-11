package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// sample is one reading of what the server process is costing: how much resident
// memory it holds and how much cpu time it has burned so far. The cpu figure is
// cumulative, so a rate is the difference between two of these over the wall
// clock between them.
type sample struct {
	rssBytes  int64
	cpuTime   time.Duration
	wallClock time.Time
}

// readSample asks the OS what the process at pid is holding right now, through
// ps, which needs no privileges the load test does not already have and reports
// the same two numbers on Linux and macOS.
func readSample(pid int) (sample, error) {
	out, err := exec.Command("ps", "-o", "rss=,time=", "-p", strconv.Itoa(pid)).Output()
	if err != nil {
		return sample{}, fmt.Errorf("ps for pid %d: %w", pid, err)
	}

	fields := strings.Fields(strings.TrimSpace(string(out)))
	if len(fields) < 2 {
		return sample{}, fmt.Errorf("ps returned %q, want rss and time", string(out))
	}

	rssKB, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return sample{}, fmt.Errorf("parsing rss %q: %w", fields[0], err)
	}

	cpuTime, err := parseCPUTime(fields[1])
	if err != nil {
		return sample{}, err
	}

	return sample{rssBytes: rssKB * 1024, cpuTime: cpuTime, wallClock: time.Now()}, nil
}

// parseCPUTime reads the cumulative cpu time ps prints, which is
// [[hours:]minutes:]seconds with the seconds carrying a fraction. macOS prints
// minutes:seconds.hundredths, and a busy process eventually grows an hours
// field, so both are handled.
func parseCPUTime(value string) (time.Duration, error) {
	parts := strings.Split(value, ":")

	var seconds float64
	multiplier := 1.0

	for i := len(parts) - 1; i >= 0; i-- {
		n, err := strconv.ParseFloat(parts[i], 64)
		if err != nil {
			return 0, fmt.Errorf("parsing cpu time %q: %w", value, err)
		}

		seconds += n * multiplier
		multiplier *= 60
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

// cpuPercent turns two cumulative readings into the busy fraction of one core
// between them, so 100 means one core saturated and 400 means four. It is the
// share of wall time the process spent on cpu, which is the number that says
// what the players cost rather than how fast any one packet was served.
func cpuPercent(before, after sample) float64 {
	wall := after.wallClock.Sub(before.wallClock).Seconds()
	if wall <= 0 {
		return 0
	}

	cpu := (after.cpuTime - before.cpuTime).Seconds()

	return cpu / wall * 100
}
