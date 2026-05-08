// Package bridge provides unit tests for the entity skill registration and conflict resolution logic.
// It ensures that skills from various origins (inventory, items) are correctly prioritized and mapped.
// @spec-link [[api_go_battle_engine]]
// @spec-link [[api_character_skill_inventory]]
package bridge

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArenaInit_EquippedSkillRegistered verifies that skills from the inventory are correctly registered in the engine.
func TestArenaInit_EquippedSkillRegistered(t *testing.T) {
	// 1. Setup Phase: Initialize the bridge and define a match request with a character carrying an inventory skill.
	b := Get()
	matchID := uuid.New()
	req := createTestRequest(matchID)
	req.Players[0].Entities[0].EquippedSkills = []api.EquippedSkill{
		{SkillID: uuid.New().String(), Name: "PowerStrike", Behavior: "Direct", Origin: "inventory"},
	}
	// 2. Execution Phase: Start the arena and wait for the async initialization to settle.
	_, _, entities, _, _, _, err := b.StartArena(req)
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)
	// 3. Validation Phase: Check if the engine's skill registry for the entity contains the expected inventory skill.
	assert.NotEmpty(t, entities[0].Skills, "entity must have skills registered after initialization")
	found := false
	for _, s := range entities[0].Skills {
		if s.Name == "PowerStrike" { found = true; break }
	}
	// 4. Final Assertion: Verify the skill name presence in the engine.
	assert.True(t, found, "skill 'PowerStrike' must be found in the engine-side skill map")
	b.DestroyArena(matchID)
}

// TestArenaInit_ItemSkillAndInventorySkillCoexist ensures that skills from different origins do not overwrite each other.
func TestArenaInit_ItemSkillAndInventorySkillCoexist(t *testing.T) {
	// 1. Setup Phase: Define an entity carrying both an inventory skill and a skill granted by an equipped weapon.
	b := Get()
	matchID := uuid.New()
	req := createTestRequest(matchID)
	inventorySkill := api.EquippedSkill{SkillID: uuid.New().String(), Name: "Heal", Behavior: "Direct", Origin: "inventory"}
	itemSkill := api.EquippedSkill{SkillID: uuid.New().String(), Name: "Cleave", Behavior: "Zone", Origin: "item:axe_001"}
	req.Players[0].Entities[0].EquippedSkills = []api.EquippedSkill{inventorySkill, itemSkill}
	// 2. Execution Phase: Start the arena and allow initialization to complete.
	_, _, entities, _, _, _, err := b.StartArena(req)
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)
	// 3. Validation Phase: Ensure both skills are present in the final engine state.
	assert.Len(t, entities[0].Skills, 2, "engine must preserve both inventory and item-based skills")
	b.DestroyArena(matchID)
}

// TestArenaInit_InvalidSkillUUIDSkipped verifies the resilience of the engine when encountering malformed skill identifiers.
func TestArenaInit_InvalidSkillUUIDSkipped(t *testing.T) {
	// 1. Setup Phase: Define a request with an invalid (non-UUID) SkillID.
	b := Get()
	matchID := uuid.New()
	req := createTestRequest(matchID)
	req.Players[0].Entities[0].EquippedSkills = []api.EquippedSkill{{SkillID: "not-a-uuid", Name: "GhostSkill"}}
	// 2. Execution Phase: Start the arena. The bridge should ignore the invalid skill without crashing.
	_, _, entities, _, _, _, err := b.StartArena(req)
	require.NoError(t, err)
	// 3. Validation Phase: Confirm the entity has zero skills registered due to invalid identifier.
	assert.Empty(t, entities[0].Skills, "malformed skill IDs must be silently ignored to prevent initialization failure")
	b.DestroyArena(matchID)
}
