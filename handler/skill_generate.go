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
// @spec-link [[rule_dto_strict_typing]]
func HandleSkillGenerate(c *gin.Context) {
	var req skillgenerator.GenerateRequest
	// BindJSON is optional; if body is empty, req remains at default (Grade I, all tags).
	// This follows the [[rule_dto_strict_typing]] by ensuring input is bound to a concrete struct.
	_ = c.ShouldBindJSON(&req)

	sk, tags, err := skillgenerator.Generate(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, api.NewError("", err.Error()))
		return
	}

	positiveSW, negativeSW, _ := skillweight.Calculate(&sk)

	behaviorStr := behaviorName(def.BehaviorType(sk.Behavior.Get().(string)))

	targeting := serializePropertyMap(sk.Targeting)
	costs := serializePropertyMap(sk.Costs)
	effectMap := serializePropertySlice(sk.Effect.Properties)

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
	return string(bt)
}

// serializePropertyMap transforms a map of engine properties into a serializable api.PropertyMap.
// It ensures that properties are mapped to the correct DTO format as per [[mechanic_mec_skill_payload_resolution]].
func serializePropertyMap(props map[string]property.Property) api.PropertyMap {
	out := make(api.PropertyMap, len(props))
	for k, v := range props {
		out[k] = serializeProperty(v)
	}
	return out
}

// serializePropertySlice transforms a slice of engine properties into a serializable api.PropertyMap.
// This is used for Effect properties which are stored as slices in the engine.
func serializePropertySlice(props []property.Property) api.PropertyMap {
	out := make(api.PropertyMap, len(props))
	for _, v := range props {
		out[v.Name(property.GameMaster)] = serializeProperty(v)
	}
	return out
}

// serializeProperty transforms a single engine property into an api.PropertyDTO.
// It handles counters and primitives, maintaining type safety for the [[rule_dto_strict_typing]].
func serializeProperty(p property.Property) api.PropertyDTO {
	dto := api.PropertyDTO{}
	if cp, ok := p.(property.IntCounterProperty); ok {
		val := cp.GetValue()
		max := cp.GetMaxValue()
		dto.Value = &val
		dto.Max = &max
		return dto
	}
	val := p.Get()
	if i, ok := val.(int); ok {
		dto.Value = &i
	} else if b, ok := val.(bool); ok {
		dto.BValue = &b
	} else if s, ok := val.(string); ok {
		dto.SValue = &s
	}
	return dto
}
