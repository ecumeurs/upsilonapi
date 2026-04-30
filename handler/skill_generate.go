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
// @spec-link [[api_skill_generate_engine]]
func HandleSkillGenerate(c *gin.Context) {
	sk := skillgenerator.GenerateRandomSkill()

	positiveSW, negativeSW, _ := skillweight.Calculate(sk)

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
		WeightPositive: positiveSW,
		WeightNegative: negativeSW,
	}

	c.JSON(http.StatusOK, api.NewSuccess("", "Skill generated", resp))
}

func behaviorName(bt def.BehaviorType) string {
	return string(bt)
}

func serializePropertyMap(props map[string]property.Property) api.PropertyMap {
	out := make(api.PropertyMap, len(props))
	for k, v := range props {
		out[k] = serializeProperty(v)
	}
	return out
}

func serializePropertySlice(props []property.Property) api.PropertyMap {
	out := make(api.PropertyMap, len(props))
	for _, v := range props {
		out[v.Name(property.GameMaster)] = serializeProperty(v)
	}
	return out
}

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
