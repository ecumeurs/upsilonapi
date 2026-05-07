package bridge

// @spec-link [[rule_team_mechanics]]
// @spec-link [[rule_forfeit_battle]]
// @spec-link [[module_upsilonapi]]

import (
	"log"
	"sync"

	"github.com/ecumeurs/upsilonbattle/battlearena"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler"
	"github.com/ecumeurs/upsilontools/tools/actor"
	"github.com/ecumeurs/upsilontools/tools/messagequeue/message"
	"github.com/google/uuid"
)

// ArenaBridge manages the set of active battle arenas.
// It is a thread-safe singleton that maps match IDs to BattleArena instances.
// @spec-link [[rule_team_mechanics]]
// @spec-link [[rule_forfeit_battle]]
// @spec-link [[module_upsilonapi]]
type ArenaBridge struct {
	mu     sync.RWMutex
	arenas map[uuid.UUID]*battlearena.BattleArena
	// @spec-link [[mech_game_state_versioning]]
	lastSentWebhookVersion map[uuid.UUID]int64
}

var bridge = &ArenaBridge{
	arenas:                 make(map[uuid.UUID]*battlearena.BattleArena),
	lastSentWebhookVersion: make(map[uuid.UUID]int64),
}

// Get returns the singleton ArenaBridge instance.
func Get() *ArenaBridge {
	return bridge
}

// GetArena retrieves the Ruler for the given arena ID.
// Returns (nil, false) if the arena does not exist.
func (b *ArenaBridge) GetArena(id uuid.UUID) (*ruler.Ruler, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	arena, ok := b.arenas[id]
	if !ok {
		return nil, false
	}
	return arena.Ruler, ok
}

// DestroyArena stops the Ruler and all associated controllers, then removes the arena from memory.
// @spec-link [[mechanic_mech_arena_lifecycle]]
func (b *ArenaBridge) DestroyArena(matchID uuid.UUID) {
	b.mu.Lock()
	arena, ok := b.arenas[matchID]
	if ok {
		delete(b.arenas, matchID)
		delete(b.lastSentWebhookVersion, matchID)
	}
	b.mu.Unlock()

	if ok && arena.Ruler != nil {
		log.Printf("[ArenaBridge] Destroying arena %s", matchID)
		// Sending ActorStop to Ruler triggers cascading shutdown of controllers.
		arena.Ruler.NotifyActor(message.Create(nil, actor.ActorStop{}, nil))
	}
}

// GetActiveMatchCount returns the number of battle arenas currently tracked in memory.
func (b *ArenaBridge) GetActiveMatchCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.arenas)
}
