// Package bridge provides unit tests pinning the fixed contract for untrusted-JSON
// property-key ingestion (ISS-147, ISS-140), per CODING_RULE §5 (test-first on bugs).
package bridge

import (
	"errors"
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPropertyIngestion_ItemPoisonMustNotBecomeEntityBuff reproduces ISS-147: an equipped
// item carrying a property key that collides with an ENTITY status-effect name (but is not
// a valid ITEM property) falls through applyItemAsBuff's scope-resolution chain and is silently
// reinterpreted as the real entity property, injecting a genuine combat status effect
// (Poison/Shield/Stun) from mere item data. It also pins the distinguishability half of the
// error contract (item d): the rejection must surface as ErrPropertyKeyWrongScope, not the
// ErrUnknownPropertyKey case covered by TestPropertyIngestion_UnknownSkillKeyMustNotBeSilentlyDropped.
// @test-link [[mechanic_item_buff_application]]
func TestPropertyIngestion_ItemPoisonMustNotBecomeEntityBuff(t *testing.T) {
	cases := []struct {
		name       string
		key        string
		entityProp property.Key
	}{
		{"Poison", "Poison", property.Poison},
		{"Shield", "Shield", property.Shield},
		{"Stun", "Stun", property.Stun},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// 1. Setup Phase: A bridge, a fresh entity, and an item whose Properties.Data
			// carries a key that is not a real item property but does collide with a
			// genuine entity status-effect property name.
			b := Get()
			e := entity.New()
			item := api.EquippedItem{
				ItemID: uuid.New().String(),
				Name:   "Cursed " + tc.name + " Trinket",
				Slot:   "finger",
				Properties: api.Flex[api.PropertyMap]{
					Data: api.PropertyMap{tc.key: {Value: intPtr(5)}},
				},
			}

			// 2. Execution Phase: Apply the item as a buff directly, bypassing StartArena.
			err := b.applyItemAsBuff(&e, item)

			// 3. Validation Phase: Item data must never inject a real entity status-effect
			// buff. The key must be rejected as "not an item property", not silently
			// reinterpreted via the entity-property fallback.
			buffed := e.GetBuffsFor(tc.entityProp)
			assert.Empty(t, buffed, "item property key %q is not a valid ITEM property, but item data injected a real entity %q status-effect buff via the entity-property fallthrough in applyItemAsBuff", tc.key, tc.key)

			// 4. Distinguishability: the rejection is a known key of the wrong scope
			// (Entity-only), which must be reported via ErrPropertyKeyWrongScope — a
			// different sentinel than an unrecognized key entirely (ErrUnknownPropertyKey).
			require.Error(t, err, "applyItemAsBuff must report the rejected key, not silently drop it")
			assert.True(t, errors.Is(err, ErrPropertyKeyWrongScope), "error must be (or wrap) ErrPropertyKeyWrongScope; got: %v", err)
			assert.False(t, errors.Is(err, ErrUnknownPropertyKey), "a registered, wrong-scope key must not also match ErrUnknownPropertyKey")
		})
	}
}

// TestPropertyIngestion_UnknownSkillKeyMustNotBeSilentlyDropped pins ISS-140's fixed contract:
// buildSkillPropertyMap now resolves every key against the registry (scoped to Skill) and
// returns (map[string]property.Property, error). An unrecognized key is never swallowed — it is
// reported via the ErrUnknownPropertyKey sentinel, named in the error message, and per the
// collect-all rule does not prevent a valid key in the same payload from resolving.
// @test-link [[mechanic_skill_payload_resolution]]
func TestPropertyIngestion_UnknownSkillKeyMustNotBeSilentlyDropped(t *testing.T) {
	t.Run("unknown key is rejected without discarding the valid key", func(t *testing.T) {
		// 1. Setup Phase: One legitimate skill key alongside one clearly-bogus key.
		raw := api.PropertyMap{
			property.Accuracy.String(): {Value: intPtr(50)},
			"NotARealProperty":         {Value: intPtr(999)},
		}

		// 2. Execution Phase: Build the skill property map from the raw DTO payload.
		result, err := buildSkillPropertyMap(raw)

		// 3. Validation Phase (replaces a pre-fix length-proxy assertion — see rationale
		// below): the bogus key must produce a distinguishable ErrUnknownPropertyKey naming
		// the offending key, while the valid key in the same payload still resolves into the
		// result map. This proves per-key errors are accumulated (collect-all), not that the
		// whole payload is discarded on the first bad key.
		//
		// The proxy this replaced (`assert.Len(t, result, len(raw))`) was written before the
		// fix shape was decided; making it literally pass would require inserting an
		// UNREGISTERED key into the engine's property map, which is the wrong contract.
		require.Error(t, err, "buildSkillPropertyMap must reject an unrecognized property key instead of silently dropping it")
		assert.True(t, errors.Is(err, ErrUnknownPropertyKey), "error must be (or wrap) ErrUnknownPropertyKey; got: %v", err)
		assert.ErrorContains(t, err, "NotARealProperty", "error message must name the offending key")
		assert.Contains(t, result, property.Accuracy.String(), "the valid key in the same payload must still resolve into the result map")
	})

	t.Run("TargetNumber is a documented, registered skill property", func(t *testing.T) {
		// 1. Setup Phase: TargetNumber is a real, documented skill property, registered
		// ScopeSkill in the registry (registry_skill_targeting.go), with a working
		// constructor (def.TargetNumber).
		raw := api.PropertyMap{
			property.TargetNumber.String(): {Value: intPtr(3)},
		}

		// 2. Execution Phase: Build the skill property map from the raw DTO payload.
		result, err := buildSkillPropertyMap(raw)

		// 3. Validation Phase: A legitimate, documented skill property key must survive
		// ingestion with no error.
		require.NoError(t, err)
		assert.Contains(t, result, property.TargetNumber.String(), "TargetNumber is a documented, registered skill property and must resolve via buildSkillPropertyMap")
	})
}
