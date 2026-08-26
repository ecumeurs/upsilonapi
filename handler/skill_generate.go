package handler

import (
	"fmt"
	"net/http"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/entity/skill/skillgenerator"
	"github.com/ecumeurs/upsilontypes/entity/skill/skillweight"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/gin-gonic/gin"
)

// HandleSkillGenerate generates a random balanced skill and returns its full JSON representation.
// @spec-link [[api_skill_generation]]
func HandleSkillGenerate(c *gin.Context) {
	// Initialize the generation request with default values.
	var req skillgenerator.GenerateRequest
	// BindJSON is optional; if body is empty, req remains at default (Grade I, all tags).
	// This follows CODING_RULE.md §4 by ensuring input is bound to a concrete struct.
	_ = c.ShouldBindJSON(&req)

	// Invoke the core generator logic to produce a balanced skill and its associated tags.
	sk, tags, err := skillgenerator.Generate(req)
	if err != nil {
		// Return 400 Bad Request on generation failures (e.g., impossible constraints).
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}

	// Calculate the power weights for both positive and negative components of the skill.
	// Weights are used to determine the relative power grade of the generated skill.
	positiveSW, negativeSW, _ := skillweight.Calculate(&sk)

	// Extract the human-readable behavior name from the underlying property.
	// Behavior types are mapped from internal engine constants to API-friendly strings.
	behaviorStr := behaviorName(def.BehaviorType(sk.Behavior.Get().(string)))

	// Map engine-specific property structures to API-standard DTOs.
	targeting := serializePropertyMap(sk.Targeting)
	costs := serializePropertyMap(sk.Costs)
	effectMap := serializePropertySlice(sk.Effect.Properties)

	// Assemble the final response DTO for the API.
	resp := api.SkillGenerateResponse{
		ID:             sk.ID.String(),
		Name:           sk.Name,
		Behavior:       behaviorStr,
		Targeting:      api.Flex[api.PropertyMap]{Data: targeting},
		Costs:          api.Flex[api.PropertyMap]{Data: costs},
		Effect:         api.Flex[api.PropertyMap]{Data: effectMap},
		Grade:          skillweight.GetGrade(positiveSW),
		Tags:           tags,
		WeightPositive: positiveSW,
		WeightNegative: negativeSW,
	}

	// NewSuccess follows the [[api_standard_envelope]] format.
	c.JSON(http.StatusOK, api.NewSuccess("", "Skill generated", resp))
}

// behaviorName returns the string representation of a BehaviorType.
func behaviorName(bt def.BehaviorType) string {
	// Simple cast from BehaviorType alias to string.
	return string(bt)
}

// serializePropertyMap transforms a map of engine properties into a serializable api.PropertyMap.
// It ensures that properties are mapped to the correct DTO format as per [[mechanic_skill_payload_resolution]].
func serializePropertyMap(props map[string]property.Property) api.PropertyMap {
	// Pre-allocate the output map with the required capacity.
	out := make(api.PropertyMap, len(props))
	// Iterate and serialize each property in the map.
	for k, v := range props {
		out[k] = serializeProperty(v)
	}
	return out
}

// serializePropertySlice transforms a slice of engine properties into a serializable api.PropertyMap.
// This is used for Effect properties which are stored as slices in the engine.
func serializePropertySlice(props []property.Property) api.PropertyMap {
	// Translate the slice into a map for easier API consumption.
	out := make(api.PropertyMap, len(props))
	for _, v := range props {
		// Use the GameMaster context for consistent property naming.
		out[v.Name(property.GameMaster)] = serializeProperty(v)
	}
	return out
}

// serializeProperty transforms a single engine property into an api.PropertyDTO.
// It is total across every property.Property implementation reachable from
// upsilontypes/property: IntCounterProperty maps to Value+Max, primitive Get()
// results (int/float64/bool/string) map onto their matching DTO field, and the
// Zone-typed property (whose Get() returns itself, not a primitive) maps onto
// SValue via its PatternType. Any property type with no defined mapping panics,
// naming the concrete Go type, per CODING_RULE.md §3 (crash early): a
// generation-time panic here is strictly better than letting an unmapped
// property degrade to an empty {} DTO, which PropertyDTO.UnmarshalJSON later
// rejects at battle start (ISS-131).
// @spec-link [[mechanic_skill_payload_resolution]]
func serializeProperty(p property.Property) api.PropertyDTO {
	// Initialize an empty DTO to hold the extracted property data.
	dto := api.PropertyDTO{}
	// Check for specialized counter properties (e.g., HP/MP).
	// These properties include both a current value and a maximum threshold.
	if cp, ok := p.(property.IntCounterProperty); ok {
		val := cp.GetValue()
		max := cp.GetMaxValue()
		dto.Value = &val
		dto.Max = &max
		return dto
	}
	// Zone-typed properties are not primitives: Get() returns the ZoneProperty
	// itself (def.ZoneProperty.Get). Serialize its PatternType instead — the
	// field commit cd75926 added expressly for this purpose.
	if zp, ok := p.(*def.ZoneProperty); ok {
		pt := zp.PatternType
		dto.SValue = &pt
		return dto
	}
	// EffectProperty.Get() returns the nested *effect.Effect struct (Properties,
	// Name, CasterID). Nothing on PropertyDTO honestly represents that shape:
	// stuffing e.g. the effect's Name into SValue would silently drop its
	// Properties and CasterID, trading one silent failure for another. Until the
	// wire format grows a real representation for nested effects, fail loudly
	// and specifically here instead of inventing a lossy encoding.
	if _, ok := p.(*def.EffectProperty); ok {
		panic(fmt.Sprintf("serializeProperty: EffectProperty has no defined wire mapping (property %q) — extend api.PropertyDTO before serializing nested effects", p.Name(property.GameMaster)))
	}
	// Fallback to generic property extraction for simple types.
	val := p.Get()
	// Polymorphic type mapping to DTO fields.
	// We handle integers, floats, booleans, and strings to cover the core engine types.
	switch v := val.(type) {
	case int:
		dto.Value = &v
	case float64:
		dto.FValue = &v
	case bool:
		dto.BValue = &v
	case string:
		dto.SValue = &v
	default:
		// Crash early (CODING_RULE.md §3): an unrecognized property type must
		// never silently serialize to an empty {} DTO — that shape is rejected
		// downstream by PropertyDTO.UnmarshalJSON, turning this into a much
		// later, harder-to-diagnose bind failure at battle start.
		panic(fmt.Sprintf("serializeProperty: unrecognized property type %T for property %q", val, p.Name(property.GameMaster)))
	}
	// Return the populated DTO for JSON serialization.
	return dto
}
