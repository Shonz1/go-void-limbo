package main

import (
	"flag"
	"log/slog"
	"runtime/debug"

	"github.com/Shonz1/go-void-limbo/auth"
	"github.com/Shonz1/go-void-limbo/config"
	"github.com/Shonz1/go-void-limbo/gamedata"
	"github.com/Shonz1/go-void-limbo/protocol"
	"github.com/Shonz1/go-void-limbo/server"
	"github.com/Shonz1/go-void-limbo/world"
)

// startupGCPercent is how far the heap may grow past what is live while the
// server is being built, as a percentage. Lower is a lower peak and more
// collections; below this the peak stops falling.
const startupGCPercent = 25

// forwardingSecretFlag is the secret on the command line, which is the one place
// an operator can put it that does not outlive the process.
var forwardingSecretFlag = flag.String("forwarding-secret", "", "the secret a modern proxy signs the logins it forwards with, taken from FORWARDING_SECRET when this is empty")

func main() {
	config.ConfigureLogging()

	flag.Parse()

	// Starting up is the one time this process allocates in earnest: every
	// registry and every chunk is parsed, translated and encoded, and nearly
	// all of it is garbage a moment later. The collector's default lets that
	// garbage pile up to the size of everything live before it runs, which is
	// the right trade for a server on a tick and the wrong one here, where it
	// only sets how high the process peaks before the first connection. The
	// default is put back once the server is built, along with the peak.
	gcPercent := debug.SetGCPercent(startupGCPercent)

	gameData, err := gamedata.NewDefaultProvider()
	if err != nil {
		slog.Error("failed to encode game registries", "err", err)
		return
	}

	// The world is optional: without one the server is the empty limbo it
	// always was. With one named and unreadable, the server stops rather than
	// starts empty, because an operator who pointed at a world wants that
	// world or the reason there is none.
	packetRegistry := protocol.NewDefaultRegistry(gameData)

	var lobby server.WorldProvider
	if dir, ok := config.WorldDir(); ok {
		loaded, err := world.Load(dir, packetRegistry)
		if err != nil {
			slog.Error("failed to load the world", "dir", dir, "err", err)
			return
		}

		lobby = loaded
	}

	// One key for the process, generated before the first client can ask for it.
	// It is only ever used to get a login's secret across, so nothing is lost by
	// it going away with the process.
	keyPair, err := auth.NewKeyPair()
	if err != nil {
		slog.Error("failed to generate the server key", "err", err)
		return
	}

	encryptionEnabled := config.EncryptionEnabled()

	// The two settings answer different questions, so neither overrules the
	// other. A secret says what a forwarded login is worth: the proxy holds the
	// connection with the player and asked Mojang there, so nothing on this side
	// of it is asked to encrypt anything, and the signature is the whole of the
	// check. Encryption says what a login nobody forwarded is worth, which is
	// still a login this server has to settle, since holding a secret does not
	// stop anything else reaching the port.
	forwardingSecret := config.ForwardingSecret(*forwardingSecretFlag)
	if len(forwardingSecret) > 0 {
		slog.Info("a forwarding secret is configured, a login signed with it is taken from the proxy that signed it and is not checked with Mojang here")

		// A secret used to force encryption off and refuse every login that was
		// not signed, so an operator who set both got what the secret alone
		// meant. It no longer does: a login nobody signed for now falls through
		// to the setting below, and with that setting off there is nothing left
		// to check it against. Said on its own because the pair is the one
		// configuration where the secret stops being worth what it looks worth,
		// and nothing about a connection says which of the two let it in.
		if !encryptionEnabled {
			slog.Warn("a forwarding secret is configured with encryption disabled, so a connection that answers the forwarding request with nothing is logged in under whatever name it asked for; the port should be one only the proxy can reach")
		}
	}

	if !encryptionEnabled {
		// The one thing this costs is the only thing anyone would want back, so
		// it is said out loud rather than left to be discovered. A login here is
		// taken on the word of whoever is on the connection, which is the proxy's
		// when one forwarded it and the client's when none did, so the port
		// should be one only what the operator trusts can reach.
		slog.Warn("encryption is disabled, logins nobody forwarded are taken on the word of whoever connects and are not checked with Mojang")
	}

	srv := server.New(server.Config{
		PacketRegistry:    packetRegistry,
		GameData:          gameData,
		World:             lobby,
		KeyPair:           keyPair,
		SessionServer:     auth.NewSessionServer(),
		Description:       config.Description(),
		GameMode:          config.GameMode(),
		EncryptionEnabled: encryptionEnabled,
		ForwardingSecret:  forwardingSecret,
	})

	// What starting up left behind is garbage the runtime would otherwise
	// hand back to the system over the following minutes, if at all. Handed
	// back now, the process settles at what it holds before the first
	// connection arrives.
	debug.SetGCPercent(gcPercent)
	debug.FreeOSMemory()

	if err := srv.ListenAndServe(config.Address()); err != nil {
		slog.Error("failed to start server", "err", err)
	}
}
