//go:build e2e

// Package e2e proves the limbo against the one thing every unit test stands in
// for: a real Minecraft client. The Void client image
// (https://void.caunt.world/docs/client/) runs an actual Mojang release in a
// container and takes orders over HTTP, so the test can launch the genuine
// client for every version this server speaks, point it at a limbo served
// in-process, and watch it join.
//
// The suite needs Docker, downloads the client image on first use, and
// launches a full Minecraft client once per version name, so it hides behind
// the e2e build tag rather than running with the ordinary suite:
//
//	go test -tags e2e -timeout 60m ./e2e
//
// A launch is seconds and a version is done in under a minute, most of which
// is the client's own thirty second wait for a spawn chunk. E2E_CLIENTS sets
// how many clients run at once -- each is a full Minecraft client, a few
// gigabytes of memory and every core it can find while it boots -- and the
// versions are spread over that many containers. It defaults to two.
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
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
const clientImage = "ghcr.io/void-community/portable-minecraft-client:offline"

// serverHostFromContainer is where the container finds the limbo, which is
// listening on the host the container runs on. Docker Desktop resolves the name
// on its own; the --add-host flag below gives plain Linux engines the same
// name, so the address works on both.
const serverHostFromContainer = "host.docker.internal"

// clientsEnv names the environment variable that sets how many client
// containers a test runs at once, and defaultClients is what it is without one.
// Two is what a laptop running Docker with its default memory comfortably
// holds; a machine with more to spare finishes proportionally sooner.
const (
	clientsEnv     = "E2E_CLIENTS"
	defaultClients = 2
)

// The client launches in demo mode, so the login is offline and the limbo has
// to take the username on the connection's word: encryption stays off, exactly
// as it would behind a proxy.
//
// How long each wait may take is split by what is being waited for. A launch
// unpacks a version's assets from the image and boots the client, minutes at
// the outside on a cold cache; the join is the game connecting to a server on
// the same machine once it has a window and has loaded; and the hold is two
// keep alive rounds, which is the proof the connection settled into the play
// phase rather than merely reaching it.
//
// The client is pointed at the limbo only after loadSettle has passed since
// its window came up, rather than told to join as it launches. A client told
// to join at launch connects before its first resource reload is through, and
// a login that completes inside that window -- a limbo's does, in
// milliseconds -- has the client drawing a world whose block atlas and
// shaders do not exist yet, which crashes it on the spot: the versions from
// 1.19 on happen to sit out the window on a profile key fetch that fails
// slowly offline, and 1.18.2, 1.18 and 1.17.1, with no key to fetch, do not. A player joins
// from a loaded client anyway, which is what this waits for; the reload takes
// a few seconds here, and the settle is what a slow machine may need.
const (
	launchTimeout = 20 * time.Minute
	loadSettle    = 15 * time.Second
	joinTimeout   = 3 * time.Minute
	stopTimeout   = 3 * time.Minute
	holdFor       = 35 * time.Second
	pollEvery     = time.Second
)

// What the position sync test waits for and settles for.
//
// A joined client waits thirty seconds for its spawn chunk before giving its
// player to gravity, which is what roamerFallTimeout has to sit through and
// what the anchor -- falling since the suite began -- never makes anyone
// wait for again. The fall covers about seventy-eight blocks a second at
// terminal velocity, which is what makes the bounds undemanding: a relayed
// fall sails past minObservedFall inside a second, and a frozen view misses
// positionTolerance by thousands after the minutes a session runs.
const (
	playersTimeout    = 2 * time.Minute
	anchorFallTimeout = time.Minute
	roamerFallTimeout = 2 * time.Minute
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
	pool := startClientPool(t)

	for _, version := range types.SupportedProtocolVersions {
		for _, name := range version.Names {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				client := pool.acquire(t)
				username := usernameFor(name)

				client.joinLimbo(t, version, name, username, port)

				// A client that failed the configuration phase, choked on a
				// mistransformed packet, or missed its keep alives would be
				// back on the menu by now, with no player to report. Still
				// having one after two keep alive intervals is the join
				// having actually held.
				time.Sleep(holdFor)

				local := client.localPlayer(t)
				if local.Name != username {
					t.Fatalf("the client is in the world as %q, want %q", local.Name, username)
				}
			})
		}
	}
}

// anchorUsername is the name the position sync test's fixed observer logs in
// under, distinct from anything usernameFor produces.
const anchorUsername = "e2eSyncAnchor"

// TestPositionSyncsWithEveryVersion proves the player sync against the real
// client of every release this server speaks. One client on the latest
// release -- the anchor -- joins once and stays; the genuine client of each
// version then joins beside it, and each side's own picture of the world is
// read back over the live players API. Both directions are checked for every
// version: the version under test must see the anchor by name and watch it
// move, the anchor must see the version's player and watch it move, each must
// be shown where the other actually is, and the version leaving must take its
// player back off the anchor's screen.
//
// The movement needs nobody at the controls: this limbo serves no world, so
// once a client's wait for its spawn chunk times out, its player falls into
// the void, streaming genuine move packets the whole way down. The anchor has
// been falling since the suite began, so its motion shows on a fresh client's
// screen at once; the fresh client's own fall starts once its chunk wait
// times out, which is what the longer of the two timeouts sits through. Were
// the relay broken, a view would sit frozen wherever the spawn packet put it.
//
// The versions share the one anchor and run as many at a time as there are
// containers in the pool, so at any moment the anchor may be watching several
// players and each of them may see the others: every lookup is by name, and a
// version leaving is checked as its own name going, not as the anchor being
// left alone.
func TestPositionSyncsWithEveryVersion(t *testing.T) {
	port := startLimbo(t)

	anchor := startClientContainer(t)
	pool := startClientPool(t)

	t.Cleanup(func() { anchor.ensureStopped(t) })

	latest := types.LatestProtocolVersion
	anchor.joinLimbo(t, latest, latest.Names[len(latest.Names)-1], anchorUsername, port)

	for _, version := range types.SupportedProtocolVersions {
		for _, name := range version.Names {
			t.Run(name, func(t *testing.T) {
				t.Parallel()

				roamer := pool.acquire(t)

				// Every subtest leans on the one anchor, so an anchor that
				// fell off the server fails loudly here rather than as ten
				// mysteries.
				if _, err := anchor.players(); err != nil {
					t.Fatalf("the anchor is no longer in the world: %v", err)
				}

				username := usernameFor(name)
				roamer.joinLimbo(t, version, name, username, port)

				// Each side is shown the other, under the name it logged in
				// as. The name proves the player info entry; the position
				// beside it is the spawn.
				roamer.awaitRemotePlayer(t, anchorUsername, playersTimeout)
				anchor.awaitRemotePlayer(t, username, playersTimeout)

				// The anchor's endless fall shows on this version's screen
				// almost at once, and this version's own fall shows on the
				// anchor's once its chunk wait times out.
				roamer.awaitRemoteFall(t, anchorUsername, anchorFallTimeout)
				anchor.awaitRemoteFall(t, username, roamerFallTimeout)

				// And what each client shows is where the other actually is,
				// read back-to-back from both APIs. The tolerance covers the
				// moments between the two reads -- a falling player covers
				// ground quickly -- while staying far under the thousands of
				// blocks a frozen view would be off by.
				for _, side := range []struct {
					name     string
					observer *voidClient
					subject  *voidClient
				}{
					{name: anchorUsername, observer: roamer, subject: anchor},
					{name: username, observer: anchor, subject: roamer},
				} {
					seen := side.observer.remotePlayer(t, side.name).Position
					actual := side.subject.localPlayer(t).Position

					t.Logf("%s is shown at %+v and stands at %+v", side.name, seen, actual)

					if diff := seen.Y - actual.Y; diff > positionTolerance || diff < -positionTolerance {
						t.Errorf("%s is shown at y=%g but stands at y=%g", side.name, seen.Y, actual.Y)
					}

					// Nothing moves a falling player sideways, so the
					// horizontal coordinates are the spawn's to within a step.
					if dx := seen.X - actual.X; dx > 1 || dx < -1 {
						t.Errorf("%s is shown at x=%g but stands at x=%g", side.name, seen.X, actual.X)
					}

					if dz := seen.Z - actual.Z; dz > 1 || dz < -1 {
						t.Errorf("%s is shown at z=%g but stands at z=%g", side.name, seen.Z, actual.Z)
					}
				}

				// Leaving is half the sync too: the player has to come back
				// off the anchor's screen, list and world both.
				roamer.ensureStopped(t)
				anchor.awaitRemoteGone(t, username, playersTimeout)
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
		PacketRegistry:    protocol.NewDefaultRegistry(gameData),
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

// clientPool is a fixed crowd of client containers that parallel subtests take
// turns with. Each container runs one game at a time, so a subtest holds one
// for as long as it runs and the pool's size is the number of versions being
// tested at once.
type clientPool struct {
	free chan *voidClient
}

// startClientPool runs as many client containers as E2E_CLIENTS asks for and
// waits for each to answer.
func startClientPool(t *testing.T) *clientPool {
	t.Helper()

	count := clientCount(t)
	pool := &clientPool{free: make(chan *voidClient, count)}

	for i := 0; i < count; i++ {
		pool.free <- startClientContainer(t)
	}

	return pool
}

// acquire hands the calling subtest a container of its own, waiting for one to
// come free if every container is busy. The container goes back to the pool
// when the subtest ends, with whatever game it left running stopped first, so
// one version's wreckage never reaches the next.
func (p *clientPool) acquire(t *testing.T) *voidClient {
	t.Helper()

	client := <-p.free

	t.Cleanup(func() {
		client.ensureStopped(t)
		p.free <- client
	})

	return client
}

// clientCount is how many client containers E2E_CLIENTS asks for, or the
// default when it is unset.
func clientCount(t *testing.T) int {
	t.Helper()

	raw, ok := os.LookupEnv(clientsEnv)
	if !ok || raw == "" {
		return defaultClients
	}

	count, err := strconv.Atoi(raw)
	if err != nil || count < 1 {
		t.Fatalf("%s must be a positive count of client containers, got %q", clientsEnv, raw)
	}

	return count
}

// voidClient is the containerized Minecraft client, spoken to over its HTTP
// API on the host port Docker picked for it. The container id is kept for the
// one thing the API does not offer, which is the game's own log after a
// failure: see dumpGameLog.
type voidClient struct {
	container string
	baseURL   string
	http      *http.Client
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

// gameJvmOptions is what every game's JVM is started with, handed to it as
// JAVA_TOOL_OPTIONS, which a JVM reads on its own: the container's launcher
// passes its environment through and sets no heap size of its own, and a JVM
// told nothing takes a quarter of the memory it can see. The memory it can
// see is the whole of Docker's, so five clients that each grow into their
// quarter of it outgrow it together, and the kernel's answer is to kill one
// of the games where it stands -- a client the suite last saw in the world,
// gone from it mid-hold with "Game exited" for a status, and a player list
// that was fine a moment ago. The two gigabytes are what Mojang's own launcher
// gives a release, and are enough for every one of them at the render
// distance the options file asks for.
const gameJvmOptions = "-Xmx2G"

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
	// Linux engines. The environment reaches the game's JVM through the
	// launcher: see gameJvmOptions.
	run := exec.Command("docker", "run", "--rm", "--detach",
		"--publish", "127.0.0.1:0:80",
		"--add-host", "host.docker.internal:host-gateway",
		"--env", "JAVA_TOOL_OPTIONS="+gameJvmOptions,
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

	client := &voidClient{
		container: id,
		baseURL:   "http://" + address,
		http:      &http.Client{Timeout: time.Minute},
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

		time.Sleep(pollEvery)
	}

	t.Fatalf("the client container did not become healthy within %v: %v", timeout, lastErr)
}

// post sends body to path and fails the test on anything but a 2xx, with
// whatever the API said about it.
func (c *voidClient) post(t *testing.T, path string, body any) {
	t.Helper()

	if err := c.tryPost(path, body); err != nil {
		t.Fatal(err)
	}
}

// tryPost is post reporting the refusal rather than failing the test on it,
// for the one request a refusal is not the end of: see joinLimbo. The error
// carries what the API said, which is what decides what the refusal is worth.
func (c *voidClient) tryPost(path string, body any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("encoding the request for %s: %w", path, err)
	}

	resp, err := c.http.Post(c.baseURL+path, "application/json", bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("POST %s: %w", path, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		answer, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("POST %s answered %s: %s", path, resp.Status, answer)
	}

	return nil
}

// put sends a plain text body to path and fails the test on anything but a
// 2xx, with whatever the API said about it. It is what the options file goes
// up as: the one endpoint that takes a file rather than JSON.
func (c *voidClient) put(t *testing.T, path, body string) {
	t.Helper()

	req, err := http.NewRequest(http.MethodPut, c.baseURL+path, strings.NewReader(body))
	if err != nil {
		t.Fatalf("building the request for %s: %v", path, err)
	}

	req.Header.Set("Content-Type", "text/plain")

	resp, err := c.http.Do(req)
	if err != nil {
		t.Fatalf("PUT %s: %v", path, err)
	}

	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		answer, _ := io.ReadAll(resp.Body)
		t.Fatalf("PUT %s answered %s: %s", path, resp.Status, answer)
	}
}

// clientOptions is the options.txt a release is launched with, in the file's
// own key:value lines. The client fills in every option the file leaves out
// with its default, so only what the tests need said is said here.
//
// The frame cap and the render distance keep a client that draws nothing worth
// seeing from taking every core it can find, which matters with several of
// them booting on one machine. A container has no focus to lose, and a client
// that pauses on losing it would sit on the pause menu instead of in the
// world.
//
// From 1.19.4 the first launch opens an accessibility onboarding screen
// ahead of the title, which no earlier release knows the option for; it is
// turned off where the release reads it and left unsaid where it would not.
func clientOptions(version types.ProtocolVersion) string {
	options := "maxFps:5\nrenderDistance:2\npauseOnLostFocus:false\n"

	if version.ID >= types.ProtocolVersions.MINECRAFT_1_19_4.ID {
		options += "onboardAccessibility:false\n"
	}

	return options
}

// joinAttempts is how many launches a join may take before it is a failure:
// the one every join needs, and one more for a client that broke the connect
// on its own side. Why one more is fair is with clientSideConnectFailures.
const joinAttempts = 2

// clientSideConnectFailures is what the client's API answers a connect with
// when the client, not the server, is what broke it. Each of these is a
// failure the suite has seen and traced to the client's own side, and the one
// thing a retry with a fresh launch is fair for: a limbo that broke the join
// breaks it again on the next launch and still fails the test, only one launch
// later.
//
// The disconnect timeout is the client's own thirty seconds of reading
// nothing, and the way this suite meets it is a race in the client's connect
// on the releases from before 1.20.2. Those send the handshake and the login
// start from the thread that opened the socket, and a send whose phase is not
// the channel's disables auto read there while the channel's own activation,
// on the event loop, is what enables it; Netty defers the read-interest clear
// the disabling asks for onto the loop, and with five clients booting on one
// machine the activation can run first, so the deferred clear lands last and
// the channel never reads again. The server's answer sits in the socket, the
// client counts thirty seconds of silence, and the limbo's log shows a joined
// client whose keep alive went unanswered. Nothing the limbo sends reaches a
// client that is not reading; a fresh launch does not sit on the same race.
//
// The empty agent response is the client's own driver -- the agent inside the
// game that the container's API drives the menus through -- not answering the
// container, with no connection to the limbo made yet.
var clientSideConnectFailures = []string{
	"disconnect.timeout",
	"The Minecraft agent returned an empty response",
}

// isClientSideConnectFailure reports whether a refused connect is one of the
// client's own, and so worth one more launch.
func isClientSideConnectFailure(err error) bool {
	for _, failure := range clientSideConnectFailures {
		if strings.Contains(err.Error(), failure) {
			return true
		}
	}

	return false
}

// joinLimbo drives one full join: launch name, a release of version, point it
// at the limbo on port once it has loaded, and wait until the game has a
// player in the world. A connect the client refused on its own side is given
// one more launch, since a client that broke its own connect is not evidence
// about the limbo: see clientSideConnectFailures. Any other refusal, and a
// second one of these, fails the test with what the client said.
func (c *voidClient) joinLimbo(t *testing.T, version types.ProtocolVersion, name, username string, port int) {
	t.Helper()

	// Whatever the game has to say about a failure is in its own log, which
	// the next launch on this container overwrites: read now, before the
	// container goes back to the pool, and only when there is a failure to
	// explain.
	t.Cleanup(func() {
		if t.Failed() {
			c.dumpGameLog(t)
		}
	})

	for attempt := 1; ; attempt++ {
		c.launch(t, version, name, username)

		err := c.connect(port)
		if err == nil {
			break
		}

		if attempt == joinAttempts || !isClientSideConnectFailure(err) {
			t.Fatalf("launch %d of %s: %v", attempt, name, err)
		}

		t.Logf("launch %d of %s broke the connect on the client's own side, launching again: %v", attempt, name, err)
	}

	c.awaitJoined(t, joinTimeout)
}

// launch starts name, a release of version, under username, and returns once
// the game has a window and has had loadSettle to load behind it. Whatever the
// container was running before is stopped first.
func (c *voidClient) launch(t *testing.T, version types.ProtocolVersion, name, username string) {
	t.Helper()

	c.ensureStopped(t)

	// The options are read as the game boots, so they go up before it does,
	// and again before every launch: the file is the container's, and the
	// release before may have rewritten it on its way out.
	c.put(t, "/api/game/options", clientOptions(version))

	c.post(t, "/api/game/start/vanilla", map[string]any{
		"version":   name,
		"arguments": []string{"--username", username},
	})

	c.awaitState(t, "ready", launchTimeout)

	// The window is up before the client has loaded: see loadSettle.
	time.Sleep(loadSettle)
}

// gameLogScript is what dumpGameLog runs inside the container: the end of the
// game's log, and the head of every crash report the game has written, which
// is where a game that exited says why.
const gameLogScript = `tail -n 60 /root/.minecraft/logs/latest.log 2>/dev/null;
for report in /root/.minecraft/crash-reports/*; do
	[ -f "$report" ] && { echo "== $report"; head -n 40 "$report"; }
done; exit 0`

// dumpGameLog logs what the game in this container wrote, for a failure the
// API's status cannot explain on its own: a game that exited, or a client
// that stopped answering. The image has no diagnostics endpoint, so the log
// is read straight off the container's filesystem.
func (c *voidClient) dumpGameLog(t *testing.T) {
	t.Helper()

	out, err := exec.Command("docker", "exec", c.container, "sh", "-c", gameLogScript).CombinedOutput()
	if err != nil {
		t.Logf("reading the game's log from the container: %v\n%s", err, out)
		return
	}

	t.Logf("the game's log ends:\n%s", out)
}

// connect points the running game at the limbo on port, the way the game's
// own direct connect screen would, and returns once the client says it is in
// the world or how it refused. The API drives the game's menus for this and
// waits on the game's own answer, so a refusal is what the client would have
// shown on its disconnect screen.
func (c *voidClient) connect(port int) error {
	return c.tryPost("/api/game/connect", map[string]any{
		"host": serverHostFromContainer,
		"port": port,
	})
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
// rather than failing the test, because the API refuses the question whenever
// the client has no player -- on the menu, on a disconnect screen, or in the
// moment around a join -- and the callers that poll want to ride that out or
// read it as the answer it is.
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

// localPlayer is the client's own player, right now. A client with no player
// to report is not in any world, which fails the test.
func (c *voidClient) localPlayer(t *testing.T) livePlayer {
	t.Helper()

	players, err := c.players()
	if err != nil {
		// A game that crashed and a client sitting on a disconnect screen
		// both have no world to ask about; the status is what tells the two
		// apart, and it is worth the extra request when the test is failing
		// anyway.
		t.Fatalf("asking the client about its world: %v (%s)", err, describe(c.status(t)))
	}

	return players.Local
}

// findRemote picks the player named name out of everyone the client sees, and
// reports whether exactly one such player is there. A crowd of others is
// fine; two of the same name is a broken player list.
func findRemote(players livePlayers, name string) (livePlayer, bool) {
	var found livePlayer
	matches := 0

	for _, player := range players.Remote {
		if player.Name == name {
			found = player
			matches++
		}
	}

	return found, matches == 1
}

// remotePlayer is the one player named name in the client's world, right now.
// Nobody of that name, or more than one, is a failure.
func (c *voidClient) remotePlayer(t *testing.T, name string) livePlayer {
	t.Helper()

	players, err := c.players()
	if err != nil {
		t.Fatalf("asking the client about its world: %v", err)
	}

	player, ok := findRemote(players, name)
	if !ok {
		t.Fatalf("the client sees %+v, want exactly one %q", players.Remote, name)
	}

	return player
}

// awaitJoined polls until the client has a player in a world, which is the
// join having gone through: the players API refuses the question until then.
func (c *voidClient) awaitJoined(t *testing.T, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	last := "no answer yet"

	for time.Now().Before(deadline) {
		_, err := c.players()
		if err == nil {
			return
		}

		last = err.Error()

		time.Sleep(pollEvery)
	}

	t.Fatalf("the client did not join within %v; last: %s; status: %s", timeout, last, describe(c.status(t)))
}

// awaitRemotePlayer polls until the client sees exactly one other player
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

		if err != nil {
			last = err.Error()
		} else if player, ok := findRemote(players, name); ok {
			return player
		} else {
			last = fmt.Sprintf("sees %+v", players.Remote)
		}

		time.Sleep(pollEvery)
	}

	t.Fatalf("the client never saw %q within %v; last: %s", name, timeout, last)

	return livePlayer{}
}

// awaitRemoteFall polls until the player named name has sunk minObservedFall
// blocks below where this client first showed it. That drop appearing on this
// client's screen is another client's move packets demonstrably crossing the
// relay; a view that never sinks is a relay that stopped at the spawn packet.
func (c *voidClient) awaitRemoteFall(t *testing.T, name string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	baseline, sighted := 0.0, false
	last := "no sighting yet"

	for time.Now().Before(deadline) {
		players, err := c.players()

		if err != nil {
			last = err.Error()
		} else if player, ok := findRemote(players, name); ok {
			y := player.Position.Y

			if !sighted {
				baseline, sighted = y, true
			}

			if baseline-y >= minObservedFall {
				return
			}

			last = fmt.Sprintf("fallen %g of %g blocks", baseline-y, minObservedFall)
		} else {
			last = fmt.Sprintf("sees %+v", players.Remote)
		}

		time.Sleep(pollEvery)
	}

	t.Fatalf("%q never fell %g blocks on this client's screen within %v; last: %s", name, minObservedFall, timeout, last)
}

// awaitRemoteGone polls until the client sees nobody named name any more,
// which is what a departed player has to leave behind.
func (c *voidClient) awaitRemoteGone(t *testing.T, name string, timeout time.Duration) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	last := "no answer yet"

	for time.Now().Before(deadline) {
		players, err := c.players()

		if err != nil {
			last = err.Error()
		} else if _, ok := findRemote(players, name); !ok {
			return
		} else {
			last = fmt.Sprintf("still sees %+v", players.Remote)
		}

		time.Sleep(pollEvery)
	}

	t.Fatalf("the client kept seeing %q within %v; last: %s", name, timeout, last)
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

		time.Sleep(pollEvery)
	}

	t.Fatalf("the client did not reach %q within %v; last: %s", want, timeout, describe(last))
}

// ensureStopped brings the container back to idle, whatever it was doing. Used
// on the way into a game as well as out of it, so one version's wreckage
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
