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

	behaviorStr := behaviorName(sk.Behavior.BehaviorType)

	targeting := serializePropertyMap(sk.Targeting)
	costs := serializePropertyMap(sk.Costs)
	effectMap := serializePropertySlice(sk.Effect.Properties)

	resp := api.SkillGenerateResponse{
		ID:             sk.ID.String(),
		Name:           sk.Name,
		Behavior:       behaviorStr,
		Targeting:      targeting,
		Costs:          costs,
		Effect:         effectMap,
		Grade:          skillweight.GetGrade(positiveSW),
		WeightPositive: positiveSW,
		WeightNegative: negativeSW,
	}

	c.JSON(http.StatusOK, api.NewSuccess("", "Skill generated", resp))
}

func behaviorName(bt def.BehaviorType) string {
	switch bt {
	case def.BehaviorTypeReaction:
		return "Reaction"
	case def.BehaviorTypePassive:
		return "Passive"
	case def.BehaviorTypeCounter:
		return "Counter"
	case def.BehaviorTypeTrap:
		return "Trap"
	default:
		return "Direct"
	}
}

func serializePropertyMap(props map[string]property.Property) map[string]any {
	out := make(map[string]any, len(props))
	for k, v := range props {
		out[k] = serializeProperty(v)
	}
	return out
}

func serializePropertySlice(props []property.Property) map[string]any {
	out := make(map[string]any, len(props))
	for _, v := range props {
		out[v.Name(property.GameMaster)] = serializeProperty(v)
	}
	return out
}

func serializeProperty(p property.Property) any {
	if cp, ok := p.(property.IntCounterProperty); ok {
		return map[string]int{"value": cp.GetValue(), "max": cp.GetMaxValue()}
	}
	if ip, ok := p.(property.IntProperty); ok {
		return ip.I()
	}
	return p.Get()
}
