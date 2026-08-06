package bridge

// @spec-link [[mechanic_mec_skill_payload_resolution]]

import (
	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/ecumeurs/upsilontypes/property/effect"
)

var propertyAliasMap = map[string]string{
	"ArmorRating": "Armor",
	"CritChance":  "CriticalChance",
	"CritDamage":  "CriticalMultiplier",
}

// parseBehaviorType converts the wire string to a def.BehaviorType.
// It defaults to Direct if the behavior is unknown or unspecified.
// This mapping ensures compatibility with the unified skill behavioral model.
// @spec-link [[mechanic_mec_skill_payload_resolution]]
func parseBehaviorType(s string) def.BehaviorType {
	// Match the incoming string to the engine's internal BehaviorType enum.
	// We handle Direct, Reaction, Passive, Counter, and Trap types.
	switch s {
	case "Reaction":
		return def.BehaviorTypeReaction
	case "Passive":
		return def.BehaviorTypePassive
	case "Counter":
		return def.BehaviorTypeCounter
	case "Trap":
		return def.BehaviorTypeTrap
	default:
		// Direct is the fallback for any unspecified or unknown behavior.
		return def.BehaviorTypeDirect
	}
}

// setSkillPropValue applies a PropertyDTO to a property.Property.
// It handles the polymorphic mapping from DTO fields (Value, FValue, BValue, SValue)
// to the appropriate engine property type. Returns true if a value was actually applied.
func setSkillPropValue(prop property.Property, dto api.PropertyDTO) bool {
	// Flag to track if any value was successfully applied from the DTO.
	hasValue := false
	// Process integer values, handling both simple IntProperty and IntCounterProperty.
	// We check for nil to differentiate between a zero-value and an omitted field.
	if dto.Value != nil {
		if cp, ok := prop.(property.IntCounterProperty); ok {
			// Counter properties maintain stateful values with upper bounds.
			cp.SetValue(*dto.Value)
			hasValue = true
		} else if ip, ok := prop.(property.IntProperty); ok {
			// Standard integer properties for simple numerical attributes.
			ip.SetI(*dto.Value)
			hasValue = true
		}
	}
	// Process float values for precision-based properties (e.g. Critical Multipliers).
	if dto.FValue != nil {
		if fp, ok := prop.(property.FloatProperty); ok {
			fp.SetF(*dto.FValue)
			hasValue = true
		}
	}
	// Handle the 'Max' threshold for counter-based properties (e.g., HP/MP pools).
	if dto.Max != nil {
		if cp, ok := prop.(property.IntCounterProperty); ok {
			cp.SetMaxValue(*dto.Max)
			hasValue = true
		}
	}
	// Process boolean flags for state-based properties (e.g. Invulnerable, Stunned).
	if dto.BValue != nil {
		if bp, ok := prop.(property.BoolProperty); ok {
			bp.SetB(*dto.BValue)
			hasValue = true
		}
	}
	// Fallback to string-based property setting for generic types or complex labels.
	// This covers properties that are passed as raw strings without specialized DTO fields.
	if dto.SValue != nil {
		prop.Set(*dto.SValue)
		hasValue = true
	}
	// Return the success status to the caller to confirm payload consumption.
	// This allows the builder to know if the property should be included in the engine state.
	return hasValue
}

// buildSkillPropertyMap reconstructs a Targeting or Costs property map from
// the DTO payload. Unknown keys are silently skipped to ensure robustness.
func buildSkillPropertyMap(raw api.PropertyMap) map[string]property.Property {
	// Initialize the result map for the engine's internal consumption.
	result := make(map[string]property.Property)
	for key, dto := range raw {
		// Look up the property definition by name from the SkillProperty registry.
		prop := def.SkillProperty(property.SkillProperties(key))
		if prop == nil {
			// Skip unknown properties to avoid engine initialization failures.
			continue
		}
		// Attempt to apply the DTO value to the property instance.
		if setSkillPropValue(prop, dto) {
			result[key] = prop
		}
	}
	return result
}

// buildSkillEffect reconstructs an effect.Effect from the DTO payload.
// It iterates through the property map and appends valid properties to the effect slice.
func buildSkillEffect(raw api.PropertyMap) effect.Effect {
	// Create a new effect instance and populate its property collection.
	eff := *effect.New()
	for key, dto := range raw {
		// Resolve the skill property from the global definition registry based on the key.
		prop := def.SkillProperty(property.SkillProperties(key))
		if prop == nil {
			// Warn or skip if the property key is not recognized by the engine.
			continue
		}
		// Map the payload value to the resolved property instance and append it.
		if setSkillPropValue(prop, dto) {
			eff.Properties = append(eff.Properties, prop)
		}
	}
	// Return the populated effect for skill registration or buff application.
	return eff
}
