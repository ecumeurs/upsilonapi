// Package bridge provides unit tests pinning the fixed contract for ISS-157: a bare-int
// `Range` in an authored skill payload previously produced an inverted, unreachable
// [value,max] window (value = the authored int as the MINIMUM, max left at the registry
// default of 1). setSkillPropValue now special-cases the Range key so a bare int resolves
// to value 0 / max N instead, while the structured form ({"value":X,"max":Y}) is untouched.
package bridge

import (
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/entity"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// rangeCounterFrom pulls the Range key out of a resolved targeting map as an
// IntCounterProperty, failing the test immediately if it is missing or the wrong shape.
func rangeCounterFrom(t *testing.T, targeting map[string]property.Property) property.IntCounterProperty {
	t.Helper()
	prop, ok := targeting[property.Range.String()]
	require.True(t, ok, "Range must resolve into the targeting map")
	counter, ok := prop.(property.IntCounterProperty)
	require.True(t, ok, "Range must resolve to an IntCounterProperty")
	return counter
}

// TestRangeIngestion_BareIntMeansZeroToN pins ISS-157's fixed contract for the bare-int
// authoring shortcut: `"Range": N` must resolve to value 0 / max N (a usable, reachable
// window), not value N / max 1 (the original inverted, unreachable defect).
// @test-link [[mechanic_skill_payload_resolution]]
func TestRangeIngestion_BareIntMeansZeroToN(t *testing.T) {
	raw := api.PropertyMap{property.Range.String(): {Value: intPtr(3)}}

	result, err := buildSkillPropertyMap(raw)

	require.NoError(t, err, "a bare-int Range must resolve without error")
	rng := rangeCounterFrom(t, result)
	assert.Equal(t, 0, rng.GetValue(), "bare-int Range must set the minimum to 0")
	assert.Equal(t, 3, rng.GetMaxValue(), "bare-int Range must set the maximum to the authored int")
}

// TestRangeIngestion_StructuredFormIsUnchanged is the regression guard: the structured
// {"value":X,"max":Y} form must keep its literal, unchanged [X,Y] meaning after the ISS-157
// fix. This is load-bearing for Sprint (upsilonhub/internal/seed/seed.go), which relies on a
// nonzero minimum of 1 so a reposition always has a direction to compute.
// @test-link [[mechanic_skill_payload_resolution]]
func TestRangeIngestion_StructuredFormIsUnchanged(t *testing.T) {
	raw := api.PropertyMap{
		property.Range.String(): {Value: intPtr(1), Max: intPtr(3)},
	}

	result, err := buildSkillPropertyMap(raw)

	require.NoError(t, err, "a valid structured Range must resolve without error")
	rng := rangeCounterFrom(t, result)
	assert.Equal(t, 1, rng.GetValue(), "structured Range must preserve the authored minimum")
	assert.Equal(t, 3, rng.GetMaxValue(), "structured Range must preserve the authored maximum")
}

// TestRangeIngestion_InvertedStructuredRangeIsRejected pins the crash-early half of the
// ISS-157 fix: a structured Range payload that is itself inverted (value > max) can never
// produce a reachable target, so it must be rejected loudly at ingestion rather than silently
// producing another dead skill.
// @test-link [[mechanic_skill_payload_resolution]]
func TestRangeIngestion_InvertedStructuredRangeIsRejected(t *testing.T) {
	raw := api.PropertyMap{
		property.Range.String(): {Value: intPtr(5), Max: intPtr(2)},
	}

	_, err := buildSkillPropertyMap(raw)

	require.Error(t, err, "an inverted structured Range (value > max) must be rejected, not silently constructed")
}

// TestRangeIngestion_SeededSkillsResolveToUsableWindows feeds each currently-seeded bare-int
// Range payload (upsilonhub/internal/seed/seed.go) through the real ingestion path and asserts
// every one resolves to a usable, non-inverted [0,N] window per the ISS-157 resolution table.
// @test-link [[mechanic_skill_payload_resolution]]
func TestRangeIngestion_SeededSkillsResolveToUsableWindows(t *testing.T) {
	cases := []struct {
		skill      string
		rangeValue int
		wantMin    int
		wantMax    int
	}{
		{"Fireball", 3, 0, 3},
		{"Heal", 2, 0, 2},
		{"Lightning Strike", 2, 0, 2},
		{"Shield Bash", 1, 0, 1},
		{"Regen Aura", 0, 0, 0},
	}

	for _, tc := range cases {
		t.Run(tc.skill, func(t *testing.T) {
			raw := api.PropertyMap{property.Range.String(): {Value: intPtr(tc.rangeValue)}}

			result, err := buildSkillPropertyMap(raw)

			require.NoError(t, err, "%s's seeded Range payload must resolve without error", tc.skill)
			rng := rangeCounterFrom(t, result)
			assert.Equal(t, tc.wantMin, rng.GetValue(), "%s: minimum range", tc.skill)
			assert.Equal(t, tc.wantMax, rng.GetMaxValue(), "%s: maximum range", tc.skill)
			assert.LessOrEqual(t, rng.GetValue(), rng.GetMaxValue(), "%s: range must not be inverted/unreachable", tc.skill)
		})
	}
}

// TestRangeIngestion_ItemPathBareIntIsUnaffected guards the most damaging possible regression:
// the ISS-157 fix is gated on the Range property key specifically (Range is ScopeSkill-only),
// so a bare-int ITEM property such as {"HP":5} must still mean value=5, exactly as before. A
// blanket "IntCounter with no Max" rule would have rewritten this to value=0/max=5, silently
// zeroing out every item's stat bonus.
// @test-link [[mechanic_skill_payload_resolution]]
func TestRangeIngestion_ItemPathBareIntIsUnaffected(t *testing.T) {
	b := Get()
	e := entity.New()
	item := api.EquippedItem{
		ItemID: uuid.New().String(),
		Name:   "Vitality Ring",
		Slot:   "finger",
		Properties: api.Flex[api.PropertyMap]{
			Data: api.PropertyMap{property.HP.String(): {Value: intPtr(5)}},
		},
	}

	err := b.applyItemAsBuff(&e, item)

	require.NoError(t, err, "a bare-int item HP property must resolve without error")
	buffed := e.GetBuffsFor(property.HP)
	require.Len(t, buffed, 1, "the item's HP buff must be registered")
	counter, ok := buffed[0].(property.IntCounterProperty)
	require.True(t, ok, "HP must resolve to an IntCounterProperty")
	assert.Equal(t, 5, counter.GetValue(), "a bare-int item property must still mean value=N, unaffected by the Range-specific ISS-157 fix")
}
