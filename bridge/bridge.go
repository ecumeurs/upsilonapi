// Package bridge provides the central orchestration layer for the UpsilonAPI.
// It manages the lifecycle of in-memory battle arenas and routes tactical requests to the underlying engine.
package bridge

import (
	"log"
	"sync"

	"github.com/ecumeurs/upsilonbattle/battlearena"
	"github.com/ecumeurs/upsilonbattle/battlearena/ruler"
	"github.com/ecumeurs/upsilontools/tools/actor"
	"github.com/ecumeurs/upsilontools/tools/messagequeue/message"
	"github.com/google/uuid"
)

// @spec-link [[api_go_battle_engine]]
// @spec-link [[mechanic_mech_arena_lifecycle]]

// ArenaBridge manages the set of active battle arenas.
// It is a thread-safe singleton that maps match IDs to BattleArena instances.
type ArenaBridge struct {
	mu     sync.RWMutex
	arenas map[uuid.UUID]*battlearena.BattleArena
	// lastSentWebhookVersion prevents redundant UI updates by tracking the latest delivered event version.
	lastSentWebhookVersion map[uuid.UUID]int64
}

var b = &ArenaBridge{
	arenas:                 make(map[uuid.UUID]*battlearena.BattleArena),
	lastSentWebhookVersion: make(map[uuid.UUID]int64),
}

// Get returns the singleton ArenaBridge instance.
func Get() *ArenaBridge {
	// 1. Singleton Pattern: Return the globally shared bridge instance.
	return b
}

// GetArena retrieves the Ruler for the given arena ID.
func (b *ArenaBridge) GetArena(id uuid.UUID) (*ruler.Ruler, bool) {
	// 1. Thread Safety: Acquire a read lock before accessing the arena map.
	b.mu.RLock()
	defer b.mu.RUnlock()
	
	// 2. Lookup: Verify presence and return the active ruler actor handle.
	arena, ok := b.arenas[id]
	if !ok { return nil, false }
	return arena.Ruler, ok
}

// DestroyArena stops the Ruler and all associated controllers, then removes the arena from memory.
func (b *ArenaBridge) DestroyArena(matchID uuid.UUID) {
	// 1. Registry Cleanup: Securely remove the arena from the active tracking maps.
	b.mu.Lock()
	arena, ok := b.arenas[matchID]
	if ok {
		delete(b.arenas, matchID)
		delete(b.lastSentWebhookVersion, matchID)
	}
	b.mu.Unlock()

	// 2. Resource Release: Signal the Ruler actor to stop, triggering a cascading shutdown.
	if ok && arena.Ruler != nil {
		log.Printf("[ArenaBridge] Destroying arena %s", matchID)
		arena.Ruler.NotifyActor(message.Create(nil, actor.ActorStop{}, nil))
	}
}

// GetActiveMatchCount returns the number of battle arenas currently tracked in memory.
func (b *ArenaBridge) GetActiveMatchCount() int {
	// 1. Telemetry: Access the map size under a shared read lock for monitoring.
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.arenas)
}
