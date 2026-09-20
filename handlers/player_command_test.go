package handlers

import (
	"testing"

	serverboundPlay "github.com/Shonz1/go-void-limbo/packets/serverbound/play"
)

func TestHandlePlayerCommandSyncsTheStances(t *testing.T) {
	client := &fakeClient{}

	command := func(action serverboundPlay.PlayerCommandAction) {
		t.Helper()

		if err := HandlePlayerCommandServerboundPacket(client, &serverboundPlay.PlayerCommandServerboundPacket{Action: action}); err != nil {
			t.Fatalf("handler error = %v", err)
		}
	}

	command(serverboundPlay.PlayerCommandStartSprinting)
	command(serverboundPlay.PlayerCommandPressShiftKey)

	if !client.syncedSprinting || !client.syncedSneaking {
		t.Fatalf("sprinting, sneaking = %t, %t, want both", client.syncedSprinting, client.syncedSneaking)
	}

	// Something a limbo has no use for leaves both alone.
	command(serverboundPlay.PlayerCommandStopSleeping)

	if !client.syncedSprinting || !client.syncedSneaking {
		t.Fatalf("sprinting, sneaking = %t, %t after stop sleeping, want both", client.syncedSprinting, client.syncedSneaking)
	}

	command(serverboundPlay.PlayerCommandStopSprinting)
	command(serverboundPlay.PlayerCommandReleaseShiftKey)

	if client.syncedSprinting || client.syncedSneaking {
		t.Fatalf("sprinting, sneaking = %t, %t, want neither", client.syncedSprinting, client.syncedSneaking)
	}
}
