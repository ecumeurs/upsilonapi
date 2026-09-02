// Package handler provides unit tests for the skill-generation HTTP surface,
// including the engine-to-wire property serialization at the root of ISS-131.
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

// propertyRoundTripCase describes one property.Property implementation to be
// pushed through serializeProperty and then a full JSON round trip. verify is
// applied twice: once to the DTO returned directly by serializeProperty (to
// pin down field-level precision) and once to the DTO recovered from
// json.Marshal -> json.Unmarshal (to reproduce the ISS-131 failure mode,
// where a DTO that looked fine in memory still marshaled to "{}" and was
// later rejected by PropertyDTO.UnmarshalJSON).
type propertyRoundTripCase struct {
	name   string
	prop   property.Property
	verify func(t *testing.T, dto api.PropertyDTO)
}

// TestSerializeProperty_RoundTrip exercises every non-panicking
// property.Property implementation through serializeProperty and a full
// json.Marshal/json.Unmarshal round trip, reproducing ISS-131's actual
// failure mode (a DTO silently marshaling to "{}", later rejected by
// PropertyDTO.UnmarshalJSON at battle start) rather than merely inspecting
// the in-memory DTO.
//
// Implementation-set closure: property.Property is implemented by exactly 7
// concrete types today. This was established two independent ways: (1) the
// set of types implementing Get() is identical to the set implementing
// GetType() across upsilontypes/property (defaultproperty's five Default*
// types, plus the two def.*Property structs); (2) property.Property is only
// ever struct-embedded at two sites, upsilontypes/property/def/item.go:69
// and upsilontypes/property/def/skill.go:135, which are EffectProperty and
// ZoneProperty themselves — i.e. no other type wraps or extends Property. Of
// the 7, EffectProperty has no honest scalar wire mapping and panics by
// design (see TestSerializeProperty_PanicsOnEffectProperty); the other 6 are
// covered below. Anyone adding an 8th implementation must extend this table
// (Go cannot enforce the closure via reflection; re-run the two checks above
// to re-verify it).
// @test-link [[mechanic_skill_payload_resolution]]
func TestSerializeProperty_RoundTrip(t *testing.T) {
	cases := []propertyRoundTripCase{
		{
			name: "DefaultIntProperty",
			prop: defaultproperty.MakeIntProperty(property.Range, 3, property.Public, property.Skill),
			verify: func(t *testing.T, dto api.PropertyDTO) {
				require.NotNil(t, dto.Value)
				require.Equal(t, 3, *dto.Value)
				require.Nil(t, dto.Max)
				require.Nil(t, dto.FValue)
				require.Nil(t, dto.BValue)
				require.Nil(t, dto.SValue)
			},
		},
		{
			name: "DefaultIntCounterProperty",
			prop: defaultproperty.MakeIntCounterProperty(property.HP, 7, 10, property.Public, property.Character),
			verify: func(t *testing.T, dto api.PropertyDTO) {
				require.NotNil(t, dto.Value)
				require.Equal(t, 7, *dto.Value)
				require.NotNil(t, dto.Max)
				require.Equal(t, 10, *dto.Max)
				require.Nil(t, dto.FValue)
				require.Nil(t, dto.BValue)
				require.Nil(t, dto.SValue)
			},
		},
		{
			name: "DefaultFloatProperty",
			prop: defaultproperty.MakeFloatProperty(property.DamageScale, 12.5, property.Public, property.Skill),
			verify: func(t *testing.T, dto api.PropertyDTO) {
				require.NotNil(t, dto.FValue)
				require.Equal(t, 12.5, *dto.FValue)
				require.Nil(t, dto.Value)
				require.Nil(t, dto.Max)
				require.Nil(t, dto.BValue)
				require.Nil(t, dto.SValue)
			},
		},
		{
			name: "DefaultBoolProperty",
			prop: defaultproperty.MakeBoolProperty(property.HasMoved, true, property.Public, property.Character),
			verify: func(t *testing.T, dto api.PropertyDTO) {
				require.NotNil(t, dto.BValue)
				require.True(t, *dto.BValue)
				require.Nil(t, dto.Value)
				require.Nil(t, dto.Max)
				require.Nil(t, dto.FValue)
				require.Nil(t, dto.SValue)
			},
		},
		{
			name: "DefaultStringProperty",
			prop: defaultproperty.MakeStringProperty(property.AIArchetype, "fighter", property.Public, property.Character),
			verify: func(t *testing.T, dto api.PropertyDTO) {
				require.NotNil(t, dto.SValue)
				require.Equal(t, "fighter", *dto.SValue)
				require.Nil(t, dto.Value)
				require.Nil(t, dto.Max)
				require.Nil(t, dto.FValue)
				require.Nil(t, dto.BValue)
			},
		},
		{
			// Reproduces ISS-131 at its root: ZoneProperty's Get() returns the
			// property itself, not a primitive, so the generic fallback in
			// serializeProperty used to return a zero-value DTO that marshals
			// to "{}" — a shape PropertyDTO.UnmarshalJSON explicitly rejects.
			name: "ZoneProperty",
			prop: def.MakeZoneProperty(pattern.Circle(3), "Circle:3"),
			verify: func(t *testing.T, dto api.PropertyDTO) {
				require.NotNil(t, dto.SValue)
				require.Equal(t, "Circle:3", *dto.SValue)
				require.Nil(t, dto.Value)
				require.Nil(t, dto.Max)
				require.Nil(t, dto.FValue)
				require.Nil(t, dto.BValue)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runPropertyRoundTrip(t, tc)
		})
	}
}

// runPropertyRoundTrip serializes tc.prop, checks the DTO's fields, then
// pushes it through json.Marshal -> json.Unmarshal and re-checks the
// recovered DTO. It explicitly asserts the marshaled form is not "{}", the
// exact shape PropertyDTO.UnmarshalJSON rejects (see api/input.go), so that
// a regression collapsing any of these 6 types back to an empty DTO fails
// deterministically instead of depending on a randomized end-to-end run.
func runPropertyRoundTrip(t *testing.T, tc propertyRoundTripCase) {
	dto := serializeProperty(tc.prop)
	tc.verify(t, dto)

	data, err := json.Marshal(dto)
	require.NoError(t, err)
	require.NotEqual(t, "{}", string(data), "property must not serialize to an empty DTO")

	var roundTripped api.PropertyDTO
	err = json.Unmarshal(data, &roundTripped)
	require.NoError(t, err, "round-tripped property JSON must be decodable")
	tc.verify(t, roundTripped)
}

// TestSerializeProperty_PanicsOnUnrecognizedType verifies that serializeProperty
// crashes early (CODING_RULE.md §3) rather than silently returning an empty
// DTO for a property type it has no defined mapping for.
// @test-link [[mechanic_skill_payload_resolution]]
func TestSerializeProperty_PanicsOnUnrecognizedType(t *testing.T) {
	require.Panics(t, func() {
		serializeProperty(unmappedProperty{})
	})
}

// TestSerializeProperty_PanicsOnEffectProperty verifies EffectProperty is
// explicitly rejected rather than silently degraded: its Get() returns a
// nested *effect.Effect struct with no honest scalar DTO mapping.
// @test-link [[mechanic_skill_payload_resolution]]
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
