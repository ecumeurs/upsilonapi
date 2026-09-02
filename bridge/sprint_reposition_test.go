// Package bridge provides a regression test pinning the Sprint skill-template seed row
// (upsilonhub/internal/seed/seed.go) as a genuine self-reposition: the original defect
// (TargetType:Self, Range:0, empty effect) made a direction impossible to compute and
// carried no reposition effect at all, and no test ever exercised a seeded skill's payload
// through the real ingestion path to catch it.
package bridge

import (
	"encoding/json"
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sprintTargetingJSON and sprintEffectJSON are the exact seeded payload strings from
// upsilonhub/internal/seed/seed.go's Sprint row (kept in sync by hand — a drift here means
// the seed row changed without this regression test being updated alongside it).
const (
	sprintTargetingJSON = `{"TargetType":"Tile","Range":{"value":1,"max":3}}`
	sprintEffectJSON    = `{"RepositionSubject":"Self","RepositionDistance":3}`
)

// TestSprintSeedPayload_IsAGenuineSelfReposition feeds the Sprint skill template's exact
// seeded JSON strings through the real buildSkillPropertyMap/buildSkillEffect ingestion path
// (the same path StartArena uses for every equipped skill) and asserts the result is
// genuinely a self-dash: a resolvable, non-degenerate aiming range and a reposition effect
// carrying a non-zero distance for the Self subject.
// @test-link [[mech_movement_reposition]]
func TestSprintSeedPayload_IsAGenuineSelfReposition(t *testing.T) {
	// 1. Setup Phase: unmarshal the seeded targeting/effect strings into the wire DTO shape,
	// exactly as they arrive over the API boundary.
	var targeting api.PropertyMap
	require.NoError(t, json.Unmarshal([]byte(sprintTargetingJSON), &targeting))
	var effectRaw api.PropertyMap
	require.NoError(t, json.Unmarshal([]byte(sprintEffectJSON), &effectRaw))

	// 2. Execution Phase: resolve both blocks through the real ingestion functions.
	targetingProps, targetingErr := buildSkillPropertyMap(targeting)
	eff, effectErr := buildSkillEffect(effectRaw)

	// 3. Validation Phase: the targeting block resolves without error and yields a usable
	// aiming range — the crux of the original defect. A bare-int Range (e.g. `"Range":3`)
	// would only set the counter's Value and leave MaxValue at the registry default (1),
	// producing an inverted, unreachable [3,1] range; the seeded row must instead resolve to
	// a genuine [1,3] window so a target tile can actually be selected.
	require.NoError(t, targetingErr, "Sprint's seeded targeting payload must resolve every key")
	targetTypeProp, ok := targetingProps[property.TargetType.String()]
	require.True(t, ok, "TargetType must resolve into the targeting map")
	assert.Equal(t, "Tile", targetTypeProp.Get().(string), "Sprint must target a tile, not itself, so a direction can exist")

	rangeProp, ok := targetingProps[property.Range.String()]
	require.True(t, ok, "Range must resolve into the targeting map")
	rangeCounter, ok := rangeProp.(property.IntCounterProperty)
	require.True(t, ok, "Range must resolve to an IntCounterProperty")
	assert.Equal(t, 1, rangeCounter.GetValue(), "min range must allow a nonzero caster-to-target distance")
	assert.Equal(t, 3, rangeCounter.GetMaxValue(), "max range must exceed min range for a usable aiming window")
	assert.GreaterOrEqual(t, rangeCounter.GetMaxValue(), rangeCounter.GetValue(), "range must not be inverted/unreachable")

	// 4. Validation Phase: the effect block resolves to a genuine self-reposition.
	require.NoError(t, effectErr, "Sprint's seeded effect payload must resolve every key")

	subjectProp := eff.GetProperty(property.RepositionSubject)
	require.NotNil(t, subjectProp, "RepositionSubject must resolve into the effect")
	assert.Equal(t, "Self", subjectProp.Get().(string), "Sprint must reposition the caster, not a target")

	distProp := eff.GetProperty(property.RepositionDistance)
	require.NotNil(t, distProp, "RepositionDistance must resolve into the effect")
	dist, ok := distProp.(property.IntProperty)
	require.True(t, ok, "RepositionDistance must resolve to an IntProperty")
	assert.NotZero(t, dist.I(), "a zero RepositionDistance means the skill does not reposition at all (the original defect)")
	assert.Equal(t, 3, dist.I())
}
