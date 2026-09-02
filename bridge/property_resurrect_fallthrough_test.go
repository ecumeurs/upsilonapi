// Package bridge provides a unit test pinning the fixed contract for ISS-147's item->entity
// fallthrough on the arena crash-recovery path (hydrateSingleBuffProperty / restoreEntityBuffs),
// which had zero coverage before this file. Written red-first, per CODING_RULE §5 (test-first
// on bugs): it was run against the unfixed hydrateSingleBuffProperty/restoreEntityBuffs (void
// return, alias+entity-fallthrough resolution) and failed on all three subtests before the fix
// landed.
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

// makeResurrectFallthroughBuff builds the single-buff persisted payload used by
// TestPropertyIngestion_ResurrectBuffPoisonMustNotBecomeEntityBuff: a buff whose
// Properties.Data carries key as its only entry. Extracted so the nested api.Flex/PropertyMap
// struct literal does not push the calling subtest past the nesting limit.
func makeResurrectFallthroughBuff(key string, value int) []api.Buff {
	return []api.Buff{
		{
			OriginID: uuid.New().String(),
			Forever:  true,
			Properties: api.Flex[api.PropertyMap]{
				Data: api.PropertyMap{key: {Value: intPtr(value)}},
			},
		},
	}
}

// TestPropertyIngestion_ResurrectBuffPoisonMustNotBecomeEntityBuff reproduces the second,
// byte-for-byte identical copy of ISS-147's item->entity fallthrough: hydrateSingleBuffProperty
// (bridge_resurrect.go), reached via restoreEntityBuffs on the ResurrectArena crash-recovery
// path. A persisted buff payload carrying a key that collides with an ENTITY status-effect name
// (Poison/Shield/Stun) but is not a valid ITEM property must be rejected as wrong-scope, not
// silently reinterpreted as the real entity property.
// @test-link [[mechanic_item_buff_application]]
func TestPropertyIngestion_ResurrectBuffPoisonMustNotBecomeEntityBuff(t *testing.T) {
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
			// 1. Setup Phase: A bridge, a fresh entity, and a persisted buff payload whose
			// Properties.Data carries a key that is not a real item property but does collide
			// with a genuine entity status-effect property name.
			b := Get()
			e := entity.New()
			buffs := makeResurrectFallthroughBuff(tc.key, 5)

			// 2. Execution Phase: Rehydrate the buff directly, bypassing ResurrectArena.
			err := b.restoreEntityBuffs(&e, buffs)

			// 3. Validation Phase: Resurrection data must never inject a real entity
			// status-effect buff. The key must be rejected as "not an item property", not
			// silently reinterpreted via the entity-property fallback in
			// hydrateSingleBuffProperty.
			buffed := e.GetBuffsFor(tc.entityProp)
			assert.Empty(t, buffed, "buff property key %q is not a valid ITEM property, but resurrection data injected a real entity %q status-effect buff via the entity-property fallthrough in hydrateSingleBuffProperty", tc.key, tc.key)

			// 4. Distinguishability: the rejection is a known key of the wrong scope
			// (Entity-only), reported via ErrPropertyKeyWrongScope, not silently dropped.
			require.Error(t, err, "restoreEntityBuffs must report the rejected key, not silently drop it")
			assert.True(t, errors.Is(err, ErrPropertyKeyWrongScope), "error must be (or wrap) ErrPropertyKeyWrongScope; got: %v", err)
		})
	}
}
