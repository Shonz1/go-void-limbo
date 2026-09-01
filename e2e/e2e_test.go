//go:build e2e

// Package e2e proves the limbo against the one thing every unit test stands in
// for: a real Minecraft client. The Void client image
// (https://void.caunt.world/docs/client/) runs an actual Mojang release in a
// container and takes orders over HTTP, so the test can launch the genuine
// client for every version this server speaks, point it at a limbo served
// in-process, and watch it join.
//
// The suite needs Docker, downloads the client image on first use, and
// launches a full Minecraft client once per version name, so it is minutes of
// work per version and hides behind the e2e build tag rather than running with
// the ordinary suite:
//
//	go test -tags e2e -timeout 180m ./e2e
//
// One container serves every subtest: the image runs one game at a time, and
// reusing the container is what its own documentation says to do between
// sessions.
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Shonz1/go-void-limbo/auth"
	"github.com/Shonz1/go-void-limbo/gamedata"
	"github.com/Shonz1/go-void-limbo/protocol"
	"github.com/Shonz1/go-void-limbo/server"
	"github.com/Shonz1/go-void-limbo/types"
)

// clientImage is the containerized Minecraft client the test drives. The
// offline tag ships every release's assets prebuilt, so a launch asks nothing
// of Mojang or the network: the one big download is the image itself.
// VOID_CLIENT_IMAGE overrides it, for pinning a digest instead.
const clientImage = "ghcr.io/caunt/portable-minecraft-client:offline"

// serverHostFromContainer is where the container finds the limbo, which is
// listening on the host the container runs on. Docker Desktop resolves the name
// on its own; the --add-host flag below gives plain Linux engines the same
// name, so the address works on both.
const serverHostFromContainer = "host.docker.internal"

// The client launches in demo mode, so the login is offline and the limbo has
// to take the username on the connection's word: encryption stays off, exactly
// as it would behind a proxy.
//
// How long each wait may take is split by what is being waited for. A launch
// unpacks a version's assets from the image and boots the client, minutes at
// the outside; a join is a handshake against a server on the same machine,
// though the API only answers once the join has settled; and the hold is two
// keep alive rounds, which is the proof the connection settled into the play
// phase rather than merely reaching it.
const (
	launchTimeout  = 20 * time.Minute
	connectTimeout = 3 * time.Minute
	stopTimeout    = 3 * time.Minute
	holdFor        = 35 * time.Second
)

// connectAttempts is how many game sessions a version gets to produce one
// join. The container joins a server by driving the client's own menus, and
// once in a while its press of the join button never lands, leaving the game
// parked on the direct connection screen; a connect re-issued at a parked game
// just parks with it, so the recovery is a fresh game, not a second ask.
const connectAttempts = 3

// TestEveryVersionJoinsWithARealClient walks every name a supported version
// answers to -- each is a Mojang release identifier -- and has the real client
// of that release join the limbo and stay. A version that shares its protocol
// with another is still launched under each of its names, because the point of
// this suite is the clients themselves, not the protocol table.
func TestEveryVersionJoinsWithARealClient(t *testing.T) {
	port := startLimbo(t)
	client := startClientContainer(t)

	for _, version := range types.SupportedProtocolVersions {
		for _, name := range version.Names {
			t.Run(name, func(t *testing.T) {
				// Whatever the last subtest left running is torn down before
				// this one launches, and this one's game is torn down after
				// it, pass or fail: the container runs one game at a time.
				t.Cleanup(func() { client.ensureStopped(t) })

				joined := false
				for attempt := 1; attempt <= connectAttempts && !joined; attempt++ {
					client.ensureStopped(t)

					client.post(t, "/api/game/start/vanilla", map[string]any{
						"version":   name,
						"arguments": []string{"--username", usernameFor(name)},
					})
					client.awaitState(t, "ready", launchTimeout)

					// Dropping the request is what cancels a connect the
					// container has wedged on, which is why this one call
					// carries its own deadline rather than the client's.
					err := client.tryConnect(serverHostFromContainer, port, connectTimeout)
					if err == nil {
						joined = true
						break
					}

					t.Logf("connect attempt %d of %d: %v", attempt, connectAttempts, err)
				}

				if !joined {
					t.Fatalf("no game session joined in %d attempts; last status: %s", connectAttempts, describe(client.status(t)))
				}

				client.awaitState(t, "connected", connectTimeout)

				// A client that failed the configuration phase, choked on a
				// mistransformed packet, or missed its keep alives would be
				// back on the menu by now. Still being connected after two
				// keep alive intervals is the join having actually held.
				time.Sleep(holdFor)

				status := client.status(t)
				if status.State != "connected" {
					t.Fatalf("the client did not stay connected: %s", describe(status))
				}
			})
		}
	}
}

// startLimbo serves the limbo in-process on a port of the system's choosing and
// reports the port. The listener closes with the test, which is all the
// shutdown the server needs.
func startLimbo(t *testing.T) int {
	t.Helper()

	gameData, err := gamedata.NewDefaultProvider()
	if err != nil {
		t.Fatalf("encoding the game registries: %v", err)
	}

	keyPair, err := auth.NewKeyPair()
	if err != nil {
		t.Fatalf("generating the server key: %v", err)
	}

	srv := server.New(server.Config{
		PacketRegistry:    protocol.NewDefaultRegistry(),
		GameData:          gameData,
		KeyPair:           keyPair,
		SessionServer:     auth.NewSessionServer(),
		Description:       "go-void-limbo e2e",
		EncryptionEnabled: false,
	})

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("listening: %v", err)
	}

	t.Cleanup(func() { listener.Close() })

	go srv.Serve(listener)

	return listener.Addr().(*net.TCPAddr).Port
}

// voidClient is the containerized Minecraft client, spoken to over its HTTP
// API on the host port Docker picked for it.
type voidClient struct {
	baseURL string
	http    *http.Client
}

// gameStatus is the shape /api/game/status answers with, down to the fields
// the test reads.
type gameStatus struct {
	State          string `json:"state"`
	Operation      string `json:"operation"`
	OperationState string `json:"operationState"`
	Message        string `json:"message"`
	Error          string `json:"error"`
}

func describe(s gameStatus) string {
	return fmt.Sprintf("state=%q operation=%q operationState=%q message=%q error=%q",
		s.State, s.Operation, s.OperationState, s.Message, s.Error)
}

// startClientContainer runs the client image and waits for its API to answer.
// The test is skipped where there is no Docker to run it with; anything else
// that goes wrong is a failure, because the e2e tag is itself the request to
// run this for real.
func startClientContainer(t *testing.T) *voidClient {
	t.Helper()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("skipping: docker is not on PATH, and the client only ships as a container")
	}

	image := clientImage
	if override := os.Getenv("VOID_CLIENT_IMAGE"); override != "" {
		image = override
	}

	// Port zero on the host side keeps parallel checkouts from fighting over a
	// port; docker port says what was picked. The add-host flag is a no-op on
	// Docker Desktop and is what makes host.docker.internal exist on plain
	// Linux engines.
	run := exec.Command("docker", "run", "--rm", "--detach",
		"--publish", "127.0.0.1:0:80",
		"--add-host", "host.docker.internal:host-gateway",
		image)

	out, err := run.CombinedOutput()
	if err != nil {
		t.Fatalf("starting the client container: %v\n%s", err, out)
	}

	id := strings.TrimSpace(string(out))
	if lines := strings.Split(id, "\n"); len(lines) > 1 {
		// A first run interleaves pull progress with the id, which is the
		// last line.
		id = strings.TrimSpace(lines[len(lines)-1])
	}

	t.Cleanup(func() {
		stop := exec.Command("docker", "stop", "--time", "10", id)
		if out, err := stop.CombinedOutput(); err != nil {
			t.Logf("stopping the client container: %v\n%s", err, out)
		}
	})

	portOut, err := exec.Command("docker", "port", id, "80").Output()
	if err != nil {
		t.Fatalf("asking docker which port it published: %v", err)
	}

	address := strings.TrimSpace(strings.Split(strings.TrimSpace(string(portOut)), "\n")[0])

	// The API answers a connect only once the join has settled -- measured at
	// better than a minute against a server on the same machine -- so the
	// timeout here is sized for the slowest call rather than a typical one.
	client := &voidClient{
		baseURL: "http://" + address,
		http:    &http.Client{Timeout: 5 * time.Minute},
	}

	client.awaitHealthy(t, 5*time.Minute)

	return client
}

func (c *voidClient) awaitHealthy(t *testing.T, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var lastErr error

	for time.Now().Before(deadline) {
		resp, err := c.http.Get(c.baseURL + "/api/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}

			lastErr = fmt.Errorf("health answered %s", resp.Status)
		} else {
			lastErr = err
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatalf("the client container did not become healthy within %v: %v", timeout, lastErr)
}

// post sends body to path and fails the test on anything but a 2xx, with
// whatever the API said about it.
func (c *voidClient) post(t *testing.T, path string, body any) {
	t.Helper()

	encoded, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("encoding the request for %s: %v", path, err)
	}

	resp, err := c.http.Post(c.baseURL+path, "application/json", bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		answer, _ := io.ReadAll(resp.Body)
		t.Fatalf("POST %s answered %s: %s", path, resp.Status, answer)
	}
}

// tryConnect asks the game to join host:port and reports how it went rather
// than failing the test, because a connect that never answers is the one call
// here a retry genuinely recovers. The API answers it only once the join has
// settled, and abandoning the request cancels the operation on the container's
// side, which is what frees a wedged game to be stopped and relaunched.
func (c *voidClient) tryConnect(host string, port int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	encoded, err := json.Marshal(map[string]any{"host": host, "port": port})
	if err != nil {
		return fmt.Errorf("encoding the request: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/game/connect", bytes.NewReader(encoded))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		answer, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("connect answered %s: %s", resp.Status, answer)
	}

	return nil
}

func (c *voidClient) status(t *testing.T) gameStatus {
	t.Helper()

	resp, err := c.http.Get(c.baseURL + "/api/game/status")
	if err != nil {
		t.Fatalf("GET /api/game/status: %v", err)
	}

	defer resp.Body.Close()

	var status gameStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("decoding the game status: %v", err)
	}

	return status
}

// awaitState polls until the game reports want. A failed state or operation is
// reported at once rather than waited out, because nothing recovers from it
// but the deadline.
func (c *voidClient) awaitState(t *testing.T, want string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	var last gameStatus

	for time.Now().Before(deadline) {
		last = c.status(t)

		if last.State == want {
			return
		}

		if last.State == "failed" || last.OperationState == "failed" {
			t.Fatalf("the client failed on the way to %q: %s", want, describe(last))
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatalf("the client did not reach %q within %v; last: %s", want, timeout, describe(last))
}

// ensureStopped brings the container back to idle, whatever it was doing. Used
// on the way into a subtest as well as out of it, so one version's wreckage
// never reaches the next.
func (c *voidClient) ensureStopped(t *testing.T) {
	t.Helper()

	if c.status(t).State == "idle" {
		return
	}

	c.post(t, "/api/game/stop", map[string]any{})
	c.awaitState(t, "idle", stopTimeout)
}

// usernameFor turns a release name into a legal offline username: the dots go,
// and what remains fits the sixteen characters a name may have.
func usernameFor(release string) string {
	return "e2e" + strings.ReplaceAll(release, ".", "")
}
