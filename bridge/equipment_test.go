// Package bridge provides unit tests for the item-to-buff conversion logic.
// It ensures that equipped items are correctly translated into active engine buffs during arena initialization.
// @test-link [[api_go_battle_engine]]
// @test-link [[mechanic_mec_skill_payload_resolution]]
package bridge

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArenaInit_EquippedItemsBecomeBuffs verifies that items in the start request are correctly injected as entity buffs.
func TestArenaInit_EquippedItemsBecomeBuffs(t *testing.T) {
	// 1. Setup Phase: Define a match request with an entity carrying a specific 'PowerRing' item.
	b := Get()
	matchID := uuid.New()
	req := createTestRequest(matchID)
	req.Players[0].Entities[0].EquippedItems = []api.EquippedItem{
		{ItemID: uuid.New().String(), Name: "PowerRing", Slot: "finger", Properties: api.Flex[api.PropertyMap]{Data: api.PropertyMap{property.Attack.String(): {Value: intPtr(5)}}}},
	}

	// 2. Execution Phase: Initialize the arena and allow the async start sequence to settle.
	_, _, entities, _, _, _, err := b.StartArena(req)
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	// 3. Validation Phase: Check if the engine entity has an active buff corresponding to the Item ID.
	assert.NotEmpty(t, entities[0].Buffs, "entity must have active buffs after equipping an item")
	found := false
	for _, b := range entities[0].Buffs {
		if b.OriginEntityID.String() == req.Players[0].Entities[0].EquippedItems[0].ItemID { found = true; break }
	}
	assert.True(t, found, "engine must contain a buff originating from the equipped ItemID")
	
	b.DestroyArena(matchID)
}

// TestArenaInit_StatMapping ensures that core entity stats (Attack/Defense) are correctly mapped from the DTO to the engine.
func TestArenaInit_StatMapping(t *testing.T) {
	// 1. Setup Phase: Define an entity with specific high combat stats.
	b := Get()
	matchID := uuid.New()
	req := createTestRequest(matchID)
	req.Players[0].Entities[0].Attack = 99
	req.Players[0].Entities[0].Defense = 42

	// 2. Execution Phase: Start the arena.
	_, _, entities, _, _, _, err := b.StartArena(req)
	require.NoError(t, err)

	// 3. Validation Phase: Ensure the engine properties for Attack and Defense match the input exactly.
	ent := entities[0]
	assert.Equal(t, 99, ent.GetPropertyI(property.Attack).I(), "attack stat must be mapped correctly to the engine")
	assert.Equal(t, 42, ent.GetPropertyI(property.Defense).I(), "defense stat must be mapped correctly to the engine")
	
	b.DestroyArena(matchID)
}
