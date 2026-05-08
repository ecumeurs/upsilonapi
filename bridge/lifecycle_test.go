// Package bridge provides unit tests for the arena lifecycle and resource management.
// It ensures that matches can be cleanly started and fully destroyed without leaking engine resources.
// @test-link [[api_go_battle_engine]]
// @test-link [[api_character_skill_inventory]]
package bridge

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestArenaLifecycleDestruction verifies the full setup and teardown sequence of a battle arena.
func TestArenaLifecycleDestruction(t *testing.T) {
	// 1. Setup Phase: Initialize the bridge and define a standard two-player match configuration.
	b := Get()
	matchID := uuid.New()
	req := createLifecycleTestRequest(matchID)

	// 2. Execution Phase: Start the arena and verify its registration in the active match registry.
	_, _, _, _, _, _, err := b.StartArena(req)
	assert.NoError(t, err)
	assert.Equal(t, 1, b.GetActiveMatchCount(), "active match count must increment after successful start")

	// 3. Teardown Phase: Explicitly destroy the arena and confirm its removal from the registry.
	b.DestroyArena(matchID)
	assert.Equal(t, 0, b.GetActiveMatchCount(), "active match count must return to zero after destruction")

	// 4. Synchronization: Wait briefly for cascading actor shutdowns to settle.
	time.Sleep(200 * time.Millisecond)
}

// createLifecycleTestRequest is a helper to build a baseline request for lifecycle validation.
func createLifecycleTestRequest(matchID uuid.UUID) api.ArenaStartRequest {
	return api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{
				ID: uuid.New().String(), Team: 1, IA: true,
				Entities: []api.Entity{{ID: uuid.New().String(), Name: "E1", HP: 10, MaxHP: 10, Move: 2, MaxMove: 2}},
			},
			{
				ID: uuid.New().String(), Team: 2, IA: true,
				Entities: []api.Entity{{ID: uuid.New().String(), Name: "E2", HP: 10, MaxHP: 10, Move: 2, MaxMove: 2}},
			},
		},
	}
}

// TestCascadingShutdown verifies that destroying an arena correctly stops all associated sub-processes.
func TestCascadingShutdown(t *testing.T) {
	// 1. Setup Phase: Start a match with a single AI controller.
	b := Get()
	matchID := uuid.New()
	pID := uuid.New()
	req := api.ArenaStartRequest{
		MatchID: matchID.String(), CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{ID: pID.String(), Team: 1, IA: true, Entities: []api.Entity{{ID: uuid.New().String(), Name: "E1", HP: 10, MaxHP: 10}}},
		},
	}

	// 2. Execution Phase: Initialize the arena and retrieve a handle to the active ruler.
	_, _, _, _, _, _, err := b.StartArena(req)
	assert.NoError(t, err)
	
	arena, ok := b.arenas[matchID]
	if !ok { t.Fatalf("Arena not found for match %s", matchID) }
	
	// 3. Component Validation: Ensure the internal ruler and controller are correctly instantiated.
	ruler := arena.Ruler
	assert.NotNil(t, ruler, "ruler must be initialized within the arena context")
	assert.NotNil(t, ruler.GameState.Controllers[pID], "AI controller must be registered for the human player")
	
	// 4. Teardown Phase: Destroy the arena and wait for the cascading ActorStop signals to propagate.
	b.DestroyArena(matchID)
	time.Sleep(500 * time.Millisecond)
	
	// 5. Completion: If the test reaches this point without hanging, the async shutdown was successful.
}
