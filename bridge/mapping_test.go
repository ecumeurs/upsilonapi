// Package bridge provides unit tests for the bidirectional mapping between API DTOs and engine properties.
// It ensures that complex effects and zones are correctly preserved during match initialization.
// @spec-link [[api_go_battle_engine]]
// @spec-link [[mechanic_mec_skill_payload_resolution]]
package bridge

import (
	"testing"
	"time"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMapping_ZoneAndEffect verifies that skill area-of-effect and payload properties are correctly hydrated in the engine.
func TestMapping_ZoneAndEffect(t *testing.T) {
	// 1. Setup Phase: Initialize the bridge and define a skill with specific Zone and Effect properties.
	b := Get()
	matchID := uuid.New()
	
	zoneName := "cross"
	req := createTestRequest(matchID)
	req.Players[0].Entities[0].EquippedSkills = []api.EquippedSkill{
		{
			SkillID: uuid.New().String(), Name: "Fireball", Behavior: "Zone",
			Zone: &zoneName,
			Effect: api.Flex[api.PropertyMap]{Data: api.PropertyMap{"damage": {Value: intPtr(10)}}},
		},
	}

	// 2. Execution Phase: Start the arena and wait for the async initialization to complete.
	_, _, entities, _, _, _, err := b.StartArena(req)
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	// 3. Validation Phase: Retrieve the engine-side skill and verify its internal property mapping.
	require.NotEmpty(t, entities[0].Skills)
	var engineSkill *api.EquippedSkill
	for _, s := range NewEntity(entities[0]).EquippedSkills {
		if s.Name == "Fireball" { engineSkill = &s; break }
	}

	// 4. Verification: Confirm the 'cross' zone and 'damage' effect were correctly mapped.
	require.NotNil(t, engineSkill, "skill 'Fireball' must exist in the engine")
	assert.Equal(t, "cross", *engineSkill.Zone, "zone pattern must be preserved in the engine state")
	assert.Equal(t, 10, *engineSkill.Effect.Data["damage"].Value, "effect properties must be correctly serialized to the engine")
	
	b.DestroyArena(matchID)
}

// intPtr is a utility to create a pointer to an integer value.
func intPtr(i int) *int {
	// 1. Memory Management: Allocate integer on the heap.
	return &i
}
