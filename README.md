# go-void-limbo

A tiny Minecraft: Java Edition server written in Go with no dependencies
outside the standard library. It is a *limbo*: a place a player can be sent
to and left waiting, holding the connection open without a game running
behind it. Put it behind a proxy as the fallback for when a real server is
down or restarting, or run it on its own as somewhere to idle.

What a joined player gets:

- **The void**, or a world. Without a world the player floats in an empty
  dimension. Point `WORLD` at a saved world and the chunks around its spawn
  are served instead, prebuilt once at startup for every protocol version.
- **A server list entry.** Pings are answered with the description from
  `MOTD`, the online count, and the client's own version whenever it is one
  the server speaks.
- **Other players.** Everyone in the limbo is visible to everyone else, with
  skins, in the player list and in the world.
- **A connection that stays up.** Keep alives go out every fifteen seconds and
  packets above 256 bytes are compressed, as a full server would.

The server speaks every protocol version from **1.19.3 through 26.2**
(protocols 761 to 776). Every packet is implemented once, at the latest
version, and carried to older clients through per-version transformers.

## Running

With Go 1.24 or newer:

```bash
go run .
```

Or with the published image:

```bash
docker run --rm -p 25565:25565 ghcr.io/shonz1/go-void-limbo:latest
```

Either way the server listens on `:25565`, encrypts connections, checks every
login with Mojang, and puts players in creative mode in the void. Everything
below is optional.

## Settings

Every setting is read from the environment. An unset or unreadable value falls
back to the default rather than stopping the server, and the fallback is
logged.

| Variable            | Default        | Meaning                                                                                                                                                                                                          |
| ------------------- | -------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ADDRESS`           | `:25565`       | Where to listen, as anything `net.Listen` takes: `:25565`, `127.0.0.1:25599`, `0.0.0.0:25565`.                                                                                                                 |
| `MOTD`              | `A void limbo` | What a server list ping describes the server as.                                                                                                                                                                 |
| `GAME_MODE`         | `creative`     | The mode joining players are put in: `survival`, `creative`, `adventure` or `spectator`, case aside. See [Game mode and the void](#game-mode-and-the-void).                                                       |
| `WORLD`             | *(unset)*      | Directory of a saved world to show, holding a `level.dat` and a `region` folder. Unset, `worlds/Lobby` under the working directory is used if it exists, and the void otherwise.                                 |
| `ENCRYPTION`        | `true`         | Whether to encrypt connections and check logins with Mojang. Anything `strconv.ParseBool` accepts. See [Encryption](#encryption).                                                                                |
| `FORWARDING_SECRET` | *(unset)*      | The secret a modern (Velocity-style) proxy signs forwarded logins with. Setting it is what turns modern forwarding on. See [Behind a proxy](#behind-a-proxy).                                                     |
| `LOG_LEVEL`         | `INFO`         | `DEBUG`, `INFO`, `WARN` or `ERROR`. Packet traffic is logged at `DEBUG`.                                                                                                                                         |

There is one command line flag:

| Flag                 | Meaning                                                                                                                                                         |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `-forwarding-secret` | The same secret as `FORWARDING_SECRET`, for an operator who would rather not leave it in the environment. The flag wins when both are set.                        |

### Encryption

By default every login is settled the way a vanilla online-mode server settles
it: the connection is encrypted with a key generated at startup, and the
account is checked with Mojang's session server. Only a player with a real
account can join, and only under their own name.

With `ENCRYPTION=false` nothing is encrypted and nothing is checked: a login is
taken on the word of whoever is on the connection. That is the right setting
behind a proxy that has already authenticated the player, and the wrong one
for a port anyone on the internet can reach. The server warns at startup when
it is off.

### Behind a proxy

A limbo is usually a backend of a proxy, which authenticates the player itself
and forwards who they are. Both forwarding schemes are supported:

- **Modern forwarding** (Velocity, and BungeeCord forks that implement it).
  Set `FORWARDING_SECRET` to the secret from the proxy's configuration. The
  server then asks every login for a payload signed with it, and a signed login
  is taken from the proxy without being checked with Mojang here. Encryption
  can stay on: it only governs logins nobody signed.
- **BungeeCord forwarding** (`ip_forward: true`). The proxy writes the account
  into the handshake in plain text, so there is nothing to sign and nothing to
  check. It is read only when `ENCRYPTION=false` and no forwarding secret is
  set, because on an encrypted server those fields would be the one way to
  skip the check.

In either case the port should be one only the proxy can reach. A limbo that
holds a forwarding secret but has encryption off will still let in a
connection that answers the forwarding request with nothing, under whatever
name it asked for.

### Worlds

A world is a normal Java Edition save: the server reads its `level.dat` for
the spawn and the region files around it, and sends the chunks within nine
chunks of spawn. The rest of the world is never loaded. Loading happens once,
at startup, for every protocol version, so the cost of a world is paid before
the first player connects: every chunk is translated, encoded and compressed
then, and held compressed, so a join sends bytes that are already on hand and
a world costs a few megabytes to keep.

If `WORLD` names a directory that cannot be loaded, the server stops rather
than starting empty: an operator who pointed at a world wants that world or
the reason there is none.

### Game mode and the void

Creative is the default because a limbo is a place to wait, and creative lets
a player fly around while doing it. One thing to know when choosing another:
a joining client waits up to thirty seconds for the chunk it stands in to
render before it gives up on the loading screen, and only a spectator skips
that wait. On a server with no world, any mode but spectator leaves a joining
player on the loading screen for those thirty seconds.

## Docker

The image is a static binary on a distroless root, running as a non-root
user, and is published for `linux/amd64` and `linux/arm64`:

```
ghcr.io/shonz1/go-void-limbo:latest
ghcr.io/shonz1/go-void-limbo:sha-<commit>
```

Every merge to `main` publishes a new `latest` alongside a tag for the commit
it was built from. Settings are passed as environment variables, and a world
is mounted in:

```bash
docker run --rm -p 25565:25565 \
  -e MOTD="Waiting for the lobby" \
  -e GAME_MODE=spectator \
  -e FORWARDING_SECRET="$(cat forwarding.secret)" \
  -v "$PWD/worlds/Lobby:/data/worlds/Lobby:ro" \
  ghcr.io/shonz1/go-void-limbo:latest
```

The working directory inside the image is `/data`, so a world mounted at
`/data/worlds/Lobby` is found without setting `WORLD`. A world mounted anywhere
else needs `WORLD` pointed at it. The flag form of the secret works too, as
the container's command:

```bash
docker run --rm -p 25565:25565 ghcr.io/shonz1/go-void-limbo:latest -forwarding-secret "$SECRET"
```

To build the image yourself:

```bash
docker build -t go-void-limbo .
```

## Development

The ordinary suite is unit tests only and runs in seconds:

```bash
go test ./...
```

The end-to-end suite launches a real Minecraft client of every supported
version in a container and has it join the limbo. It needs Docker and downloads
a large client image on first use, so it sits behind a build tag:

```bash
go test -tags e2e -timeout 60m ./e2e
```

Each version takes well under a minute, most of it the client's own wait for
a spawn chunk, and the versions are spread over a pool of client containers
that run at once. `E2E_CLIENTS` sets the pool's size and defaults to two; each
client is a full Minecraft client that wants a few gigabytes of memory, so
raise it as far as the machine running Docker allows.

A load generator under [`cmd/loadtest`](cmd/loadtest/README.md) connects a
crowd of simulated clients and reports what serving them costs.

### Continuous integration

- Every pull request is built, vetted, formatted-checked and tested under the
  race detector, the Docker image is built without being pushed, and the
  end-to-end suite has the real client of every version join the limbo on a
  hosted runner.
- Every push to `main` builds the multi-platform image and publishes it to the
  GitHub Container Registry.

### Layout

| Package        | Holds                                                                                          |
| -------------- | ---------------------------------------------------------------------------------------------- |
| `config`       | Reading the settings above from the environment and command line.                              |
| `server`       | Accepting connections, the shared registries, the status a ping is answered from.              |
| `client`       | One connection: its phases, encryption, compression, keep alives, and the packet loop.          |
| `handlers`     | What is done with each serverbound packet.                                                     |
| `packets`      | Every packet, clientbound and serverbound, implemented at the latest protocol version.          |
| `transformers` | How a packet at the latest version is carried down to each older one, and back up.              |
| `protocol`     | The packet id tables per version and phase.                                                    |
| `gamedata`     | The registries and tags the configuration phase sends, generated from the client's own data.    |
| `auth`         | Mojang session checks, the server key, and both forwarding schemes.                            |
| `world`, `anvil` | Reading a saved world and turning its chunks into packets.                                   |
| `nbt`, `streams`, `types` | The wire formats everything above is written in.                                     |
