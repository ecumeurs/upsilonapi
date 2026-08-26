// Package bridge provides unit tests for the bidirectional mapping between API DTOs and engine properties.
// It ensures that complex effects and zones are correctly preserved during match initialization.
// @test-link [[api_go_battle_engine]]
// @test-link [[mechanic_skill_payload_resolution]]
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

// TestMapping_ZoneAndEffect verifies that skill area-of-effect and payload properties are correctly hydrated in the engine.
// The zone pattern "Circle:3" is a valid AoE pattern and exercises the parameterised path in ZoneProperty.Set.
func TestMapping_ZoneAndEffect(t *testing.T) {
	// 1. Setup Phase: Initialize the bridge and define a skill with a Circle AoE zone and an Effect property.
	b := Get()
	matchID := uuid.New()

	zoneName := "Circle:3"
	req := createTestRequest(matchID)
	req.Players[0].Entities[0].EquippedSkills = []api.EquippedSkill{
		{
			SkillID: uuid.New().String(), Name: "Fireball", Behavior: "Zone",
			Zone: &zoneName,
			Effect: api.Flex[api.PropertyMap]{Data: api.PropertyMap{property.Damage.String(): {Value: intPtr(10)}}},
		},
	}

	// 2. Execution Phase: Start the arena and wait for the async initialization to complete.
	_, _, entities, _, _, _, err := b.StartArena(req)
	require.NoError(t, err)
	time.Sleep(150 * time.Millisecond)

	// 3. Validation Phase: Retrieve the engine-side skill and verify its internal property mapping.
	require.NotEmpty(t, entities[0].Skills)
	var engineSkill *api.EquippedSkill
	for _, s := range api.NewEntity(entities[0]).EquippedSkills {
		if s.Name == "Fireball" {
			engineSkill = &s
			break
		}
	}

	// 4. Verification: Confirm the "Circle:3" zone and 'damage' effect were correctly mapped.
	require.NotNil(t, engineSkill, "skill 'Fireball' must exist in the engine")
	assert.Equal(t, "Circle:3", *engineSkill.Zone, "zone pattern must be preserved in the engine state")
	assert.Equal(t, 10, *engineSkill.Effect.Data[property.Damage.String()].Value, "effect properties must be correctly serialized to the engine")

	b.DestroyArena(matchID)
}
