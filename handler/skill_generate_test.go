// Package handler provides unit tests for the skill-generation HTTP surface,
// including the engine-to-wire property serialization at the root of ISS-131.
// @test-link [[mechanic_skill_payload_resolution]]
package handler

import (
	"encoding/json"
	"testing"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilonmapdata/grid/position/pattern"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/ecumeurs/upsilontypes/property/defaultproperty"
	"github.com/stretchr/testify/require"
)

// TestSerializeProperty_ZoneRoundTrip reproduces ISS-131 at its root: a
// ZoneProperty's Get() returns the property itself, not a primitive, so the
// generic fallback in serializeProperty used to return a zero-value DTO that
// marshals to "{}" — a shape PropertyDTO.UnmarshalJSON explicitly rejects.
// Before the fix, the round-tripped JSON is "{}" and re-decoding it fails with
// "invalid property format: {}"; after the fix, PatternType survives via
// SValue.
func TestSerializeProperty_ZoneRoundTrip(t *testing.T) {
	// 1. Setup: an AoE Zone property as the engine would build it for a rolled
	// skill (mirrors the CI-observed "dot"/"aoe" tagged skill from ISS-131).
	zp := def.MakeZoneProperty(pattern.Circle(3), "Circle:3")

	// 2. Execution: serialize through the function under test, then marshal.
	dto := serializeProperty(zp)
	data, err := json.Marshal(dto)
	require.NoError(t, err)
	require.NotEqual(t, "{}", string(data), "a Zone property must not serialize to an empty DTO")

	// 3. Validation: the JSON must decode back into a PropertyDTO carrying the
	// original PatternType — this is the actual round trip the engine performs
	// when a client persists and later replays a skill's targeting payload.
	var roundTripped api.PropertyDTO
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err, "round-tripped Zone property JSON must be decodable")
	require.NotNil(t, roundTripped.SValue)
	require.Equal(t, "Circle:3", *roundTripped.SValue)
}

// TestSerializeProperty_FloatValue verifies that a float64-valued property
// (e.g. DefaultFloatProperty) populates PropertyDTO.FValue instead of falling
// through to an empty DTO, matching the decode side which already handles
// FValue (input.go).
func TestSerializeProperty_FloatValue(t *testing.T) {
	fp := defaultproperty.MakeFloatProperty(property.Damage, 12.5, property.Public, property.Skill)

	dto := serializeProperty(fp)

	require.NotNil(t, dto.FValue)
	require.Equal(t, 12.5, *dto.FValue)
}

// TestSerializeProperty_PanicsOnUnrecognizedType verifies that serializeProperty
// crashes early (CODING_RULE.md §3) rather than silently returning an empty
// DTO for a property type it has no defined mapping for.
func TestSerializeProperty_PanicsOnUnrecognizedType(t *testing.T) {
	require.Panics(t, func() {
		serializeProperty(unmappedProperty{})
	})
}

// TestSerializeProperty_PanicsOnEffectProperty verifies EffectProperty is
// explicitly rejected rather than silently degraded: its Get() returns a
// nested *effect.Effect struct with no honest scalar DTO mapping.
func TestSerializeProperty_PanicsOnEffectProperty(t *testing.T) {
	ep := def.MakeEffectProperty(nil, property.Analyser)

	require.Panics(t, func() {
		serializeProperty(ep)
	})
}

// unmappedProperty is a minimal property.Property implementation whose Get()
// returns a type serializeProperty has no defined mapping for, used to assert
// the panic-on-unrecognized-type behavior.
type unmappedProperty struct{}

// Name returns the stub's fixed identifier at any information level.
func (unmappedProperty) Name(property.InformationLevel) string { return "Unmapped" }

// UserFriendlyGet is unused by serializeProperty; the stub returns nil.
func (unmappedProperty) UserFriendlyGet(property.InformationLevel) interface{} { return nil }

// Get returns an anonymous struct — a type serializeProperty has no mapping
// for — which is the whole point of this stub.
func (unmappedProperty) Get() interface{} { return struct{}{} }

// Set is a no-op; the stub is read-only.
func (unmappedProperty) Set(interface{}) {}

// Increase is a no-op; the stub carries no numeric value.
func (unmappedProperty) Increase() {}

// GetType reports None, the stub having no real engine property type.
func (unmappedProperty) GetType() property.PropertyType { return property.None }

// Duplicate returns the stub itself; it is immutable and stateless.
func (u unmappedProperty) Duplicate() property.Property { return u }

// ApplyBuff is a no-op; the stub is unaffected by buffs.
func (u unmappedProperty) ApplyBuff(property.Property) property.Property { return u }

// UnapplyBuff is a no-op; the stub is unaffected by buffs.
func (u unmappedProperty) UnapplyBuff(property.Property) property.Property { return u }
