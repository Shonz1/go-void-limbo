package play

import (
	"fmt"
	"github.com/Shonz1/go-void-limbo/streams"
	"github.com/Shonz1/go-void-limbo/types"
	"strings"
)

// PlayerInfoAction is one field of a player list entry.
//
// A player info update packet announces which fields it carries as a bitmask
// and then every entry in it carries exactly those, in the order below. The
// order is the client's own and the bit is a position in it, so neither can be
// renumbered: a wrong bit makes the client read one field as another.
type PlayerInfoAction uint8

const (
	// PlayerInfoAddPlayer carries the profile, and is what creates the entry.
	// The client ignores an update for a player it was never told about, so
	// the first packet about a player has to include this.
	PlayerInfoAddPlayer PlayerInfoAction = 1 << iota

	// PlayerInfoInitializeChat carries the session that signs a player's chat.
	// This packet always writes it absent, which is what an unauthenticated
	// player has.
	PlayerInfoInitializeChat

	PlayerInfoUpdateGameMode
	PlayerInfoUpdateListed
	PlayerInfoUpdateLatency

	// PlayerInfoUpdateDisplayName carries the name the player list shows in
	// place of the profile's. This packet always writes it absent, which
	// leaves the client showing the profile name.
	PlayerInfoUpdateDisplayName

	PlayerInfoUpdateListOrder
	PlayerInfoUpdateHat
)

var playerInfoActionNames = []struct {
	action PlayerInfoAction
	name   string
}{
	{PlayerInfoAddPlayer, "add_player"},
	{PlayerInfoInitializeChat, "initialize_chat"},
	{PlayerInfoUpdateGameMode, "update_game_mode"},
	{PlayerInfoUpdateListed, "update_listed"},
	{PlayerInfoUpdateLatency, "update_latency"},
	{PlayerInfoUpdateDisplayName, "update_display_name"},
	{PlayerInfoUpdateListOrder, "update_list_order"},
	{PlayerInfoUpdateHat, "update_hat"},
}

func (a PlayerInfoAction) String() string {
	if a == 0 {
		return "none"
	}

	names := make([]string, 0, len(playerInfoActionNames))
	for _, known := range playerInfoActionNames {
		if a&known.action != 0 {
			names = append(names, known.name)
		}
	}

	return strings.Join(names, "|")
}

// PlayerInfoEntry is what one player's list entry is set to. Only the fields
// the packet's actions name are sent, and the rest are ignored.
type PlayerInfoEntry struct {
	Profile types.GameProfile

	GameMode GameMode
	Listed   bool

	// Latency is the round trip in milliseconds, which decides how many bars
	// the player list draws. A limbo that sends no keep alives has not
	// measured one.
	Latency int32

	// ListOrder sorts the player list, ahead of the name for players that
	// share a value.
	ListOrder int32

	// ShowHat draws the second skin layer on the player's head.
	ShowHat bool
}

func (e PlayerInfoEntry) String() string {
	return fmt.Sprintf("PlayerInfoEntry{Profile:%s GameMode:%s Listed:%t Latency:%d ListOrder:%d ShowHat:%t}",
		e.Profile, e.GameMode, e.Listed, e.Latency, e.ListOrder, e.ShowHat)
}

func (e PlayerInfoEntry) encode(ms *streams.MinecraftStream, actions PlayerInfoAction) error {
	if err := ms.WriteUuid(e.Profile.Uuid); err != nil {
		return err
	}

	// Every action the packet declares is written here, in the order the bits
	// are declared in. The client reads them the same way, with nothing to say
	// where one ends and the next begins.
	if actions&PlayerInfoAddPlayer != 0 {
		// The client refuses a name longer than 16 characters, which is the
		// limit Mojang accounts are held to anyway.
		if err := ms.WriteString(e.Profile.Username); err != nil {
			return err
		}

		if err := encodeProfileProperties(ms, e.Profile.Properties); err != nil {
			return err
		}
	}

	if actions&PlayerInfoInitializeChat != 0 {
		if err := ms.WriteBoolean(false); err != nil {
			return err
		}
	}

	if actions&PlayerInfoUpdateGameMode != 0 {
		if err := ms.WriteVarInt(int32(e.GameMode)); err != nil {
			return err
		}
	}

	if actions&PlayerInfoUpdateListed != 0 {
		if err := ms.WriteBoolean(e.Listed); err != nil {
			return err
		}
	}

	if actions&PlayerInfoUpdateLatency != 0 {
		if err := ms.WriteVarInt(e.Latency); err != nil {
			return err
		}
	}

	if actions&PlayerInfoUpdateDisplayName != 0 {
		if err := ms.WriteBoolean(false); err != nil {
			return err
		}
	}

	if actions&PlayerInfoUpdateListOrder != 0 {
		if err := ms.WriteVarInt(e.ListOrder); err != nil {
			return err
		}
	}

	if actions&PlayerInfoUpdateHat != 0 {
		if err := ms.WriteBoolean(e.ShowHat); err != nil {
			return err
		}
	}

	return nil
}

// encodeProfileProperties writes the block of profile properties the client
// reads with the same codec the login success packet's profile goes through.
func encodeProfileProperties(ms *streams.MinecraftStream, properties []types.ProfileProperty) error {
	if err := ms.WriteVarInt(int32(len(properties))); err != nil {
		return err
	}

	for _, property := range properties {
		if err := ms.WriteString(property.Name); err != nil {
			return err
		}

		if err := ms.WriteString(property.Value); err != nil {
			return err
		}

		if err := ms.WriteBoolean(property.Signature != nil); err != nil {
			return err
		}

		if property.Signature != nil {
			if err := ms.WriteString(*property.Signature); err != nil {
				return err
			}
		}
	}

	return nil
}

// PlayerInfoUpdateClientboundPacket tells the client about players. It is what
// fills the player list, and what a client reads a player's skin and game mode
// from -- including its own, which is why a joining client is sent an entry
// for itself.
type PlayerInfoUpdateClientboundPacket struct {
	Actions PlayerInfoAction
	Entries []PlayerInfoEntry
}

func (p *PlayerInfoUpdateClientboundPacket) String() string {
	entries := make([]string, 0, len(p.Entries))
	for _, entry := range p.Entries {
		entries = append(entries, entry.String())
	}

	return fmt.Sprintf("PlayerInfoUpdateClientboundPacket{Actions:%s Entries:[%s]}", p.Actions, strings.Join(entries, " "))
}

func (p *PlayerInfoUpdateClientboundPacket) Encode(ms *streams.MinecraftStream) error {
	// The mask is a fixed bit set, one byte per eight actions rather than a
	// VarInt, so the eight that exist take exactly one.
	if err := ms.WriteByte(byte(p.Actions)); err != nil {
		return err
	}

	if err := ms.WriteVarInt(int32(len(p.Entries))); err != nil {
		return err
	}

	for _, entry := range p.Entries {
		if err := entry.encode(ms, p.Actions); err != nil {
			return err
		}
	}

	return nil
}
