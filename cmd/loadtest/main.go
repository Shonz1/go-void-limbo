// Command loadtest drives a crowd of simulated Minecraft clients at the limbo
// and watches what serving them costs. Each bot completes a real login and then
// stays connected the way a joined player does, answering keep alives, so the
// server holds and counts it as online. The bots are ramped up through a series
// of player counts, and at each count the server process's memory and cpu are
// sampled and reported, which is the picture of how the limbo scales.
//
// By default it builds and launches the limbo itself, with encryption off so a
// bot needs no Mojang account, and measures that child process in isolation.
// Point -server at an already-running instance to measure one you started
// yourself instead.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

func main() {
	var (
		serverAddr = flag.String("server", "", "address of an already-running limbo to measure; when empty the tool builds and launches one itself")
		host       = flag.String("host", "localhost", "the address bots claim to have reached the server at, sent in the handshake")
		levelsFlag = flag.String("levels", "100,250,500,1000,2000", "comma-separated online player counts to step through")
		hold       = flag.Duration("hold", 20*time.Second, "how long to hold each level while sampling")
		rate       = flag.Int("rate", 200, "how many new bots to connect per second while ramping to a level")
		settle     = flag.Duration("settle", 3*time.Second, "how long to let a level settle after the last bot joins before sampling")
		moves      = flag.Int("moves", 100, "position update packets each joined bot sends per second")
		csvPath    = flag.String("csv", "", "optional path to write the per-level results to as csv")
		serverEnv  = flag.String("serverenv", "", "comma-separated KEY=VALUE pairs added to the launched limbo's environment, for tuning experiments (e.g. GOGC=400,GOMAXPROCS=8); they reach the server only, never the bots")
	)
	flag.Parse()

	levels, err := parseLevels(*levelsFlag)
	if err != nil {
		log.Fatalf("levels: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	addr := *serverAddr
	var server *serverProcess

	if addr == "" {
		server, err = launchServer(ctx, parseEnvPairs(*serverEnv))
		if err != nil {
			log.Fatalf("launching the limbo: %v", err)
		}
		defer server.shutdown()
		addr = server.addr
	}

	if err := waitForPort(ctx, addr, 15*time.Second); err != nil {
		log.Fatalf("the limbo never answered on %s: %v", addr, err)
	}

	pid := 0
	if server != nil {
		pid = server.cmd.Process.Pid
	} else {
		log.Printf("measuring an external server on %s; cpu/ram sampling needs its pid", addr)
	}

	h := &harness{addr: addr, host: *host, moves: *moves, stop: make(chan struct{})}
	defer h.shutdown()

	results := make([]levelResult, 0, len(levels))

	for _, target := range levels {
		if ctx.Err() != nil {
			break
		}

		log.Printf("ramping to %d online players...", target)
		h.rampTo(ctx, target, *rate)

		// Let the join wave settle so the reading is of holding the crowd, not of
		// letting it in.
		if !sleep(ctx, *settle) {
			break
		}

		online := int(h.joined.Load())
		failed := int(h.failed.Load())

		res := levelResult{target: target, online: online, failed: failed}

		if pid > 0 {
			res.metrics, err = measure(ctx, pid, *hold)
			if err != nil {
				log.Printf("sampling at %d players: %v", target, err)
			}
		} else if !sleep(ctx, *hold) {
			break
		}

		results = append(results, res)
		log.Printf("  %s", res.line())
	}

	fmt.Println()
	report(results, pid > 0)

	if *csvPath != "" {
		if err := writeCSV(*csvPath, results); err != nil {
			log.Printf("writing csv: %v", err)
		} else {
			log.Printf("wrote %s", *csvPath)
		}
	}
}

// harness owns the live bots: the address they dial, the one channel that tells
// them all to leave, and the tallies of who joined and who could not.
type harness struct {
	addr  string
	host  string
	moves int

	stop chan struct{}
	wg   sync.WaitGroup

	spawned int
	joined  atomic.Int64
	failed  atomic.Int64
}

// rampTo brings the number of connected bots up to target, opening no more than
// rate new connections a second so the ramp is a arrival curve rather than a
// thundering herd. It returns once every new bot has either joined or failed.
func (h *harness) rampTo(ctx context.Context, target, rate int) {
	need := target - h.spawned
	if need <= 0 {
		return
	}

	ticker := time.NewTicker(time.Second / time.Duration(max(rate, 1)))
	defer ticker.Stop()

	var joining sync.WaitGroup

	for i := 0; i < need; i++ {
		select {
		case <-ctx.Done():
			joining.Wait()
			return
		case <-ticker.C:
		}

		n := h.spawned
		h.spawned++

		h.wg.Add(1)
		joining.Add(1)

		go func(n int) {
			defer h.wg.Done()

			b, err := dialBot(h.addr, h.host, 25565, fmt.Sprintf("Bot%06d", n), offlineUUID(n), h.moves)
			if err != nil {
				h.failed.Add(1)
				joining.Done()
				return
			}

			// The join wait is released once, whichever comes first: the bot
			// reaching play, or its connection ending before it got there. A bot
			// that never joins must not hang the ramp on it.
			var settled sync.Once

			joined := false
			onJoin := func() {
				joined = true
				h.joined.Add(1)
				settled.Do(joining.Done)
			}

			runErr := b.run(h.stop, onJoin)

			settled.Do(func() {
				h.failed.Add(1)
				joining.Done()
			})

			// A bot that had joined and then left lowers the online count. The
			// error, if any, is only interesting when nothing else explains the
			// drop, and at this scale logging every one would drown the report.
			if joined {
				h.joined.Add(-1)
				_ = runErr
			}
		}(n)
	}

	joining.Wait()
}

func (h *harness) shutdown() {
	close(h.stop)
	h.wg.Wait()
}

// serverProcess is the limbo the tool launched, kept so it can be measured and
// then stopped.
type serverProcess struct {
	cmd    *exec.Cmd
	addr   string
	binary string
}

// launchAddr is where the limbo the tool launches listens: loopback, on a port
// of the tool's own, so a run never collides with a dev server already sitting
// on the default port -- a collision the tool would otherwise misread, with the
// bots joining the dev server while the sampler watches the launched limbo die.
const launchAddr = "127.0.0.1:25599"

// launchServer builds the limbo and starts it with encryption off and its log
// quieted, so the only thing the sampler sees the process doing is serving bots.
// extraEnv goes to the limbo alone: tuning the runtime under test must not also
// tune the process generating its load.
func launchServer(ctx context.Context, extraEnv []string) (*serverProcess, error) {
	binary, err := os.CreateTemp("", "limbo-loadtest-*")
	if err != nil {
		return nil, err
	}
	binary.Close()

	build := exec.CommandContext(ctx, "go", "build", "-o", binary.Name(), ".")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		os.Remove(binary.Name())
		return nil, fmt.Errorf("go build: %w", err)
	}

	cmd := exec.Command(binary.Name())
	cmd.Env = append(os.Environ(), "ENCRYPTION=false", "LOG_LEVEL=error", "ADDRESS="+launchAddr)
	cmd.Env = append(cmd.Env, extraEnv...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		os.Remove(binary.Name())
		return nil, err
	}

	log.Printf("launched the limbo (pid %d, encryption off) on %s", cmd.Process.Pid, launchAddr)

	return &serverProcess{cmd: cmd, addr: launchAddr, binary: binary.Name()}, nil
}

func (s *serverProcess) shutdown() {
	if s.cmd.Process != nil {
		s.cmd.Process.Signal(syscall.SIGTERM)
		done := make(chan struct{})
		go func() { s.cmd.Wait(); close(done) }()
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			s.cmd.Process.Kill()
		}
	}
	os.Remove(s.binary)
}

// metrics is what holding a level cost the server: the memory it settled at and
// the cpu it drew, averaged and peaked over the hold.
type metrics struct {
	rssBytes int64
	peakRSS  int64
	avgCPU   float64
	peakCPU  float64
}

// measure samples the process across the hold and folds the readings into the
// average and peak of each. cpu is the busy share of one core between readings,
// so it is only known from the second sample on.
func measure(ctx context.Context, pid int, hold time.Duration) (metrics, error) {
	const interval = time.Second

	prev, err := readSample(pid)
	if err != nil {
		return metrics{}, err
	}

	var m metrics
	m.rssBytes = prev.rssBytes
	m.peakRSS = prev.rssBytes

	var cpuSum float64
	var cpuReadings int

	deadline := time.After(hold)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return finishMetrics(m, cpuSum, cpuReadings), ctx.Err()
		case <-deadline:
			return finishMetrics(m, cpuSum, cpuReadings), nil
		case <-ticker.C:
			cur, err := readSample(pid)
			if err != nil {
				return finishMetrics(m, cpuSum, cpuReadings), err
			}

			pct := cpuPercent(prev, cur)
			cpuSum += pct
			cpuReadings++
			if pct > m.peakCPU {
				m.peakCPU = pct
			}

			m.rssBytes = cur.rssBytes
			if cur.rssBytes > m.peakRSS {
				m.peakRSS = cur.rssBytes
			}

			prev = cur
		}
	}
}

func finishMetrics(m metrics, cpuSum float64, readings int) metrics {
	if readings > 0 {
		m.avgCPU = cpuSum / float64(readings)
	}
	return m
}

// levelResult is one row of the report: the players asked for, the players that
// actually joined, and what serving them cost.
type levelResult struct {
	target  int
	online  int
	failed  int
	metrics metrics
}

func (r levelResult) line() string {
	return fmt.Sprintf("target %d online %d failed %d rss %s cpu avg %.0f%% peak %.0f%%",
		r.target, r.online, r.failed, humanBytes(r.metrics.rssBytes), r.metrics.avgCPU, r.metrics.peakCPU)
}

func report(results []levelResult, haveMetrics bool) {
	fmt.Println("Load test results")
	fmt.Println(strings.Repeat("=", 78))
	if haveMetrics {
		fmt.Printf("%-8s %-8s %-8s %-12s %-12s %-10s %-10s\n",
			"target", "online", "failed", "RSS", "peak RSS", "CPU avg", "CPU peak")
	} else {
		fmt.Printf("%-8s %-8s %-8s\n", "target", "online", "failed")
	}
	fmt.Println(strings.Repeat("-", 78))

	for _, r := range results {
		if haveMetrics {
			fmt.Printf("%-8d %-8d %-8d %-12s %-12s %-10s %-10s\n",
				r.target, r.online, r.failed,
				humanBytes(r.metrics.rssBytes), humanBytes(r.metrics.peakRSS),
				fmt.Sprintf("%.0f%%", r.metrics.avgCPU), fmt.Sprintf("%.0f%%", r.metrics.peakCPU))
		} else {
			fmt.Printf("%-8d %-8d %-8d\n", r.target, r.online, r.failed)
		}
	}
	fmt.Println(strings.Repeat("=", 78))

	if haveMetrics && len(results) > 0 {
		first, last := results[0], results[len(results)-1]
		if first.online > 0 && last.online > first.online {
			perPlayer := float64(last.metrics.rssBytes-first.metrics.rssBytes) / float64(last.online-first.online)
			fmt.Printf("\nMarginal memory per online player (between %d and %d): %s\n",
				first.online, last.online, humanBytes(int64(perPlayer)))
		}
	}
}

func writeCSV(path string, results []levelResult) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	fmt.Fprintln(w, "target,online,failed,rss_bytes,peak_rss_bytes,cpu_avg_pct,cpu_peak_pct")
	for _, r := range results {
		fmt.Fprintf(w, "%d,%d,%d,%d,%d,%.2f,%.2f\n",
			r.target, r.online, r.failed,
			r.metrics.rssBytes, r.metrics.peakRSS, r.metrics.avgCPU, r.metrics.peakCPU)
	}

	return nil
}

// --- small helpers ---

func parseLevels(s string) ([]int, error) {
	parts := strings.Split(s, ",")
	levels := make([]int, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil, fmt.Errorf("%q is not a number", p)
		}
		if n <= 0 {
			return nil, fmt.Errorf("%d is not a positive player count", n)
		}
		levels = append(levels, n)
	}
	if len(levels) == 0 {
		return nil, fmt.Errorf("no levels given")
	}
	return levels, nil
}

// parseEnvPairs splits a comma-separated list of KEY=VALUE pairs, dropping
// empty entries so a flag left unset adds nothing.
func parseEnvPairs(s string) []string {
	var pairs []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			pairs = append(pairs, p)
		}
	}
	return pairs
}

// waitForPort blocks until something is listening on addr or the timeout runs
// out, so the ramp does not start dialling before the limbo is up.
func waitForPort(ctx context.Context, addr string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err == nil {
			conn.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
	return fmt.Errorf("timed out")
}

// sleep waits for d, or returns false if the context is cancelled first.
func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// offlineUUID mirrors what the server would give a nameless offline login: a
// deterministic uuid per bot so two bots never share one, which the player list
// would collapse.
func offlineUUID(n int) string {
	return fmt.Sprintf("00000000-0000-0000-0000-%012x", n)
}

func humanBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}
