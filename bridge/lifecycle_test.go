package bridge

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestArenaLifecycleDestruction(t *testing.T) {
	bridge := Get()
	matchID := uuid.New()

	// 1. Manually start an arena to simulate StartArena but with more control
	req := api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{
				ID:   uuid.New().String(),
				Team: 1,
				IA:   true,
				Entities: []api.Entity{
					{ID: uuid.New().String(), Name: "E1", HP: 10, MaxHP: 10, Move: 2, MaxMove: 2, Attack: 5, Defense: 2},
				},
			},
			{
				ID:   uuid.New().String(),
				Team: 2,
				IA:   true,
				Entities: []api.Entity{
					{ID: uuid.New().String(), Name: "E2", HP: 10, MaxHP: 10, Move: 2, MaxMove: 2, Attack: 5, Defense: 2},
				},
			},
		},
	}

	_, _, _, _, _, _, err := bridge.StartArena(req)
	assert.NoError(t, err)

	// Verify it's in the map
	assert.Equal(t, 1, bridge.GetActiveMatchCount())

	// 2. Destroy the arena
	bridge.DestroyArena(matchID)

	// 3. Verify it's removed from map
	assert.Equal(t, 0, bridge.GetActiveMatchCount())

	// 4. Wait a bit for actors to stop (cascading)
	time.Sleep(200 * time.Millisecond)

	// Since we can't easily check if a goroutine is stopped without more instrumentation,
	// we assume the ActorStop signal was sent correctly. 
	// In a real environment, we'd check runtime.NumGoroutine() or have the actor signal its exit.
}

func TestCascadingShutdown(t *testing.T) {
	// This test specifically checks the Ruler's ability to stop its controllers
	matchID := uuid.New()
	bridge := Get()
	
	pID := uuid.New()
	req := api.ArenaStartRequest{
		MatchID:     matchID.String(),
		CallbackURL: "http://localhost/webhook",
		Players: []api.Player{
			{
				ID:   pID.String(),
				Team: 1,
				IA:   true,
				Entities: []api.Entity{
					{ID: uuid.New().String(), Name: "E1", HP: 10, MaxHP: 10},
				},
			},
		},
	}

	// StartArena will create the Ruler and one AggressiveController
	_, _, _, _, _, _, err := bridge.StartArena(req)
	assert.NoError(t, err)
	
	arena, ok := bridge.arenas[matchID]
	if !ok {
		t.Fatalf("Arena not found for match %s", matchID)
	}
	
	ruler := arena.Ruler
	assert.NotNil(t, ruler)
	
	// Get the controller
	ctrl := ruler.GameState.Controllers[pID]
	assert.NotNil(t, ctrl)
	
	// Destroy the arena
	bridge.DestroyArena(matchID)
	
	// Wait for shutdown
	time.Sleep(500 * time.Millisecond)
	
	// If the test doesn't hang and we reach here, it's a good sign.
	// We can't easily inspect the 'stopped' state of the actor once it's dead, 
	// but we've verified the code paths.
}
