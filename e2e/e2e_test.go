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

// What the position sync test waits for and settles for.
//
// A joined client waits thirty seconds for its spawn chunk before giving its
// player to gravity, so playersTimeout has to reach past that wait as well as
// the polling in front of it. The fall itself covers about seventy-eight
// blocks a second at terminal velocity, which is what makes the two bounds
// below undemanding: a relayed fall sails past minObservedFall inside a
// second, and a frozen view misses positionTolerance by thousands after the
// minutes a session runs.
const (
	playersTimeout    = 2 * time.Minute
	fallWindow        = 15 * time.Second
	minObservedFall   = 10.0
	positionTolerance = 256.0
)

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

				client.joinLimbo(t, name, usernameFor(name), port)

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

// TestPositionSyncsBetweenTwoRealClients has two genuine clients join the same
// limbo -- the oldest and the newest release it speaks, so every transformer
// step is crossed in both directions -- and reads each client's own picture of
// the world back over the live players API. Each must see the other, by name
// and at a position.
//
// The movement half needs nobody at the controls: this limbo serves no world,
// so once a client's wait for its spawn chunk times out, its player falls into
// the void, streaming genuine move packets the whole way down. The other
// client's view of that player falling is the position sync demonstrably
// relaying; were it broken, the view would sit frozen wherever the spawn
// packet put it.
//
// Two containers, because each runs one game at a time.
func TestPositionSyncsBetweenTwoRealClients(t *testing.T) {
	port := startLimbo(t)

	elder := startClientContainer(t)
	younger := startClientContainer(t)

	t.Cleanup(func() { elder.ensureStopped(t) })
	t.Cleanup(func() { younger.ensureStopped(t) })

	const (
		elderName   = "e2eSyncElder"
		youngerName = "e2eSyncYoung"
	)

	elder.joinLimbo(t, "1.21.6", elderName, port)
	younger.joinLimbo(t, "26.2", youngerName, port)

	// Each side is shown exactly the other: the join that worked puts one
	// remote player on each client's list, under the name the other logged in
	// as. What the name proves is the player info entry; the position beside
	// it is the spawn.
	elderView := elder.awaitRemotePlayer(t, youngerName, playersTimeout)
	youngerView := younger.awaitRemotePlayer(t, elderName, playersTimeout)

	t.Logf("the 1.21.6 client sees %q at %+v; the 26.2 client sees %q at %+v",
		youngerName, elderView.Position, elderName, youngerView.Position)

	// The elder joined first and has been falling since its chunk wait timed
	// out, so its position sinks on the younger's screen from one look to the
	// next. Ten blocks over this window is a fraction of a real fall and far
	// beyond any jitter, and a relay that stopped would show exactly zero.
	before := younger.remotePlayer(t, elderName)
	time.Sleep(fallWindow)
	after := younger.remotePlayer(t, elderName)

	if drop := before.Position.Y - after.Position.Y; drop < minObservedFall {
		t.Errorf("the elder player fell %g blocks on the younger's screen over %v, want at least %g: moves are not being relayed",
			drop, fallWindow, minObservedFall)
	}

	// And what each client shows is where the other actually is, read
	// back-to-back from both APIs. The tolerance covers the moments between
	// the two reads -- a falling player covers ground quickly -- while staying
	// far under the thousands of blocks a frozen view would be off by.
	for _, side := range []struct {
		name     string
		observer *voidClient
		subject  *voidClient
	}{
		{name: elderName, observer: younger, subject: elder},
		{name: youngerName, observer: elder, subject: younger},
	} {
		seen := side.observer.remotePlayer(t, side.name).Position
		actual := side.subject.localPlayer(t).Position

		if diff := seen.Y - actual.Y; diff > positionTolerance || diff < -positionTolerance {
			t.Errorf("%s is shown at y=%g but stands at y=%g", side.name, seen.Y, actual.Y)
		}

		// Nothing moves a falling player sideways, so the horizontal
		// coordinates are the spawn's to within a step.
		if dx := seen.X - actual.X; dx > 1 || dx < -1 {
			t.Errorf("%s is shown at x=%g but stands at x=%g", side.name, seen.X, actual.X)
		}

		if dz := seen.Z - actual.Z; dz > 1 || dz < -1 {
			t.Errorf("%s is shown at z=%g but stands at z=%g", side.name, seen.Z, actual.Z)
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

// joinLimbo drives one full join: launch the named release, connect it to the
// limbo on port, and wait until the session reports connected. The menu-driving
// flake and its recovery -- a fresh game, not a second ask -- live here so
// every test joins the same way.
func (c *voidClient) joinLimbo(t *testing.T, version, username string, port int) {
	t.Helper()

	joined := false
	for attempt := 1; attempt <= connectAttempts && !joined; attempt++ {
		c.ensureStopped(t)

		c.post(t, "/api/game/start/vanilla", map[string]any{
			"version":   version,
			"arguments": []string{"--username", username},
		})
		c.awaitState(t, "ready", launchTimeout)

		// Dropping the request is what cancels a connect the container has
		// wedged on, which is why this one call carries its own deadline
		// rather than the client's.
		err := c.tryConnect(serverHostFromContainer, port, connectTimeout)
		if err == nil {
			joined = true
			break
		}

		t.Logf("connect attempt %d of %d: %v", attempt, connectAttempts, err)
	}

	if !joined {
		t.Fatalf("no game session joined in %d attempts; last status: %s", connectAttempts, describe(c.status(t)))
	}

	c.awaitState(t, "connected", connectTimeout)
}

// The shapes /api/game/players answers with, down to the fields the tests
// read: the client's own player, and everyone else it can see in its world.
type playerPosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type livePlayer struct {
	Uuid     string         `json:"uuid"`
	Name     string         `json:"name"`
	Position playerPosition `json:"position"`
}

type livePlayers struct {
	Local  livePlayer   `json:"local"`
	Remote []livePlayer `json:"remote"`
}

// players asks the client who is in its world right now. It reports an error
// rather than failing the test, because the API refuses the question for a
// moment around a join, and the callers that poll want to ride that out.
func (c *voidClient) players() (livePlayers, error) {
	resp, err := c.http.Get(c.baseURL + "/api/game/players")
	if err != nil {
		return livePlayers{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		answer, _ := io.ReadAll(resp.Body)
		return livePlayers{}, fmt.Errorf("players answered %s: %s", resp.Status, answer)
	}

	var players livePlayers
	if err := json.NewDecoder(resp.Body).Decode(&players); err != nil {
		return livePlayers{}, fmt.Errorf("decoding the players: %w", err)
	}

	return players, nil
}

// localPlayer is the client's own player, right now.
func (c *voidClient) localPlayer(t *testing.T) livePlayer {
	t.Helper()

	players, err := c.players()
	if err != nil {
		t.Fatalf("asking the client about its world: %v", err)
	}

	return players.Local
}

// remotePlayer is the one player named name in the client's world, right now.
// Any other crowd -- nobody, somebody else, or more than one -- is a failure,
// because the tests using this put exactly two players on the server.
func (c *voidClient) remotePlayer(t *testing.T, name string) livePlayer {
	t.Helper()

	players, err := c.players()
	if err != nil {
		t.Fatalf("asking the client about its world: %v", err)
	}

	if len(players.Remote) != 1 || players.Remote[0].Name != name {
		t.Fatalf("the client sees %+v, want exactly %q", players.Remote, name)
	}

	return players.Remote[0]
}

// awaitRemotePlayer polls until the client sees exactly one other player,
// named name, and returns that sighting. On the way to it the API may refuse
// the question, know of nobody, or -- briefly, around the spawn packets --
// hold an entry whose entity has not appeared yet; all of that is waited out,
// and only the deadline turns it into an answer about the sync being broken.
func (c *voidClient) awaitRemotePlayer(t *testing.T, name string, timeout time.Duration) livePlayer {
	t.Helper()

	deadline := time.Now().Add(timeout)
	last := "no answer yet"

	for time.Now().Before(deadline) {
		players, err := c.players()

		switch {
		case err != nil:
			last = err.Error()
		case len(players.Remote) == 1 && players.Remote[0].Name == name:
			return players.Remote[0]
		default:
			last = fmt.Sprintf("sees %+v", players.Remote)
		}

		time.Sleep(2 * time.Second)
	}

	t.Fatalf("the client never saw %q within %v; last: %s", name, timeout, last)

	return livePlayer{}
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
