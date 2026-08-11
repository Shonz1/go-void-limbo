# loadtest

A load generator that connects a crowd of simulated Minecraft clients to the
limbo and reports what serving them costs in memory and CPU at each player
count.

Each simulated client ("bot") completes a real login over TCP and then behaves
like a joined player: it answers keep alives and teleports, and streams position
update packets (100 per second by default) — the busiest serverbound traffic a
real client produces. So the server holds the connection open, counts it as
online, and decodes a steady stream of packets for it. Bots are ramped up
through a series of player counts; at each count the server process is sampled
and a row is printed.

## Running

By default the tool builds and launches the limbo itself, with encryption
disabled (so a bot needs no Mojang account) and its log quieted, then measures
that child process in isolation. The launched limbo listens on a loopback port
of the tool's own (`127.0.0.1:25599`), so a run never collides with a dev
server already sitting on the default port:

```bash
go run ./cmd/loadtest -levels 100,250,500,1000,2000 -hold 20s
```

To measure a server you started yourself instead, point `-server` at it. Note
that CPU/RAM sampling only runs for a server the tool launched (it needs the
pid); an external run reports connection counts only:

```bash
ENCRYPTION=false go run .            # in one terminal, from the repo root
go run ./cmd/loadtest -server 127.0.0.1:25565 -levels 100,500,1000
```

Raise the open-file limit before a large run, since every bot holds a socket:

```bash
ulimit -n 65536
```

## Flags

| flag       | default                 | meaning                                                        |
| ---------- | ----------------------- | -------------------------------------------------------------- |
| `-levels`  | `100,250,500,1000,2000` | comma-separated online player counts to step through          |
| `-hold`    | `20s`                   | how long to hold each level while sampling                     |
| `-moves`   | `100`                   | position update packets each joined bot sends per second       |
| `-rate`    | `200`                   | new bots connected per second while ramping                    |
| `-settle`  | `3s`                    | pause after the last join before sampling begins               |
| `-server`  | *(empty)*               | address of a running limbo to measure; empty means launch one |
| `-host`    | `localhost`             | address the bots claim in the handshake                        |
| `-csv`     | *(empty)*               | optional path to also write the results as csv                 |
| `-serverenv` | *(empty)*             | comma-separated KEY=VALUE pairs for the launched limbo's environment, e.g. `GOMAXPROCS=4`; they reach the server only, never the bots |

## Reading the results

CPU is the busy share of one core (100% is one core saturated, 400% is four).
With bots streaming position updates, CPU is driven by the total packet rate —
online players times `-moves` — so it climbs steeply with the crowd; set
`-moves 0` to measure idle connections alone, where CPU sits near zero between
keep alives. The memory column grows with the number of connections rather than
the packet rate. The tool prints the marginal memory per online player across
the first and last level, which is roughly what one connection's goroutine and
buffers cost.
