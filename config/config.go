// Package config reads what an operator said about the server, from the
// environment and the command line, and turns each setting into the value the
// rest of the process runs on.
package config

import (
	"log/slog"
	"os"
	"strconv"
)

// ConfigureLogging sets the level the default logger keeps, read from LOG_LEVEL
// as one of DEBUG, INFO, WARN or ERROR. Packet traffic is logged at DEBUG, so it
// is silent until asked for.
func ConfigureLogging() {
	level := slog.LevelInfo
	unrecognized := ""

	if raw, ok := os.LookupEnv("LOG_LEVEL"); ok {
		if err := level.UnmarshalText([]byte(raw)); err != nil {
			level = slog.LevelInfo
			unrecognized = raw
		}
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})))

	if unrecognized != "" {
		slog.Warn("unrecognized LOG_LEVEL, falling back to INFO", "value", unrecognized)
	}
}

// EncryptionEnabled reports whether connections are to be encrypted, read from
// ENCRYPTION as anything strconv.ParseBool accepts.
//
// It defaults to on, and an unrecognized value is treated as on rather than
// refused, because every way of misreading this setting has to land on the safe
// side: an unencrypted server is one anyone can log in to under anyone's name,
// and nothing about a connection says which of the two it got.
func EncryptionEnabled() bool {
	raw, ok := os.LookupEnv("ENCRYPTION")
	if !ok {
		return true
	}

	enabled, err := strconv.ParseBool(raw)
	if err != nil {
		slog.Warn("unrecognized ENCRYPTION, falling back to enabled", "value", raw)
		return true
	}

	return enabled
}

// defaultAddress is where the server listens when the operator has said
// nothing about it, which is the port every Minecraft client fills in on its
// own.
const defaultAddress = ":25565"

// Address reports where the server is to listen, read from ADDRESS as anything
// net.Listen takes, like ":25565" or "127.0.0.1:25599".
//
// An empty value is treated as nothing said rather than passed along, because
// net.Listen reads an empty address as a port picked by the OS, which is a
// server no client ever finds.
func Address() string {
	if raw, ok := os.LookupEnv("ADDRESS"); ok && raw != "" {
		return raw
	}

	return defaultAddress
}

// defaultDescription is what the server list draws under this server's address
// when the operator has said nothing about it. It says what the server is, which
// is the one thing a player looking at a list of them needs to be told.
const defaultDescription = "A void limbo"

// Description reports what a ping describes this server as, read from MOTD.
//
// An empty value is treated as nothing said rather than as an empty description,
// because a blank line in a server list is indistinguishable from a server that
// failed to answer.
func Description() string {
	if raw, ok := os.LookupEnv("MOTD"); ok && raw != "" {
		return raw
	}

	return defaultDescription
}

// defaultWorldDir is where a world is looked for when the operator has said
// nothing about it, relative to wherever the server was started.
const defaultWorldDir = "worlds/Lobby"

// WorldDir reports where the world to show joined players is saved, read from
// WORLD as a directory holding a level.dat and a region folder.
//
// It reports false for no world at all, which is what this server shows
// without one: the void it is named for. An operator who set WORLD gets that
// directory and whatever loading it says; one who set nothing gets the default
// only if it exists, because a server that has always started against an empty
// directory should go on starting against one.
func WorldDir() (string, bool) {
	if raw, ok := os.LookupEnv("WORLD"); ok && raw != "" {
		return raw, true
	}

	if _, err := os.Stat(defaultWorldDir); err == nil {
		return defaultWorldDir, true
	}

	return "", false
}

// ForwardingSecret reports the secret a modern proxy shares with this server,
// taken from the -forwarding-secret argument when it has one and from
// FORWARDING_SECRET otherwise. The flag wins over the environment when both are
// set, because it is the more deliberate of the two: the environment is
// inherited and a flag is typed.
//
// There is no setting that turns forwarding on. The secret is the setting: a
// server that holds one is a server behind a proxy, and asks every login for a
// payload signed with it. A server that holds none never asks, and logins there
// are settled as they were before any of this existed.
//
// So an empty value is no secret rather than an empty one. A secret nobody set
// is a secret everybody has, and a server that asked for a signature under it
// would be checking a signature anyone can produce, which is worse than not
// asking at all.
func ForwardingSecret(argument string) []byte {
	if argument != "" {
		return []byte(argument)
	}

	if raw, ok := os.LookupEnv("FORWARDING_SECRET"); ok && raw != "" {
		return []byte(raw)
	}

	return nil
}
