package handler

import (
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
// It ensures that properties are mapped to the correct DTO format as per [[mechanic_mec_skill_payload_resolution]].
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
// It handles counters and primitives, maintaining type safety per CODING_RULE.md §4.
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
	// Fallback to generic property extraction for simple types.
	val := p.Get()
	// Polymorphic type mapping to DTO fields.
	// We handle integers, booleans, and strings to cover the core engine types.
	if i, ok := val.(int); ok {
		dto.Value = &i
	} else if b, ok := val.(bool); ok {
		dto.BValue = &b
	} else if s, ok := val.(string); ok {
		dto.SValue = &s
	}
	// Return the populated DTO for JSON serialization.
	return dto
}
