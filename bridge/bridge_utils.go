package bridge

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ecumeurs/upsilonapi/api"
	"github.com/ecumeurs/upsilontypes/property"
	"github.com/ecumeurs/upsilontypes/property/def"
	"github.com/ecumeurs/upsilontypes/property/effect"
)

// ErrUnknownPropertyKey and ErrPropertyKeyWrongScope are the two sentinel errors the bridge's
// property-key ingestion boundary returns for untrusted-JSON input (ISS-140, ISS-147). A bad key
// here is bad INPUT, not a programming error, so it is always reported via a returned error —
// never a panic — and the two cases are kept distinguishable via errors.Is.
var (
	// ErrUnknownPropertyKey means the key is not registered in the property registry at all.
	ErrUnknownPropertyKey = errors.New("unknown property key")
	// ErrPropertyKeyWrongScope means the key is registered, but not for the scope being resolved
	// (e.g. an Entity-only key supplied as an item property).
	ErrPropertyKeyWrongScope = errors.New("property key not valid for this scope")
)

// scopeLabel renders a def.Scope bitmask as a readable "Entity|Skill|Item" label for error
// messages, without requiring any change to upsilontypes/property/def.
func scopeLabel(s def.Scope) string {
	var parts []string
	if s&def.ScopeEntity != 0 {
		parts = append(parts, "Entity")
	}
	if s&def.ScopeSkill != 0 {
		parts = append(parts, "Skill")
	}
	if s&def.ScopeItem != 0 {
		parts = append(parts, "Item")
	}
	if len(parts) == 0 {
		return "None"
	}
	return strings.Join(parts, "|")
}

// resolveScopedProperty looks up key in the registry and returns a fresh Property instance if,
// and only if, the key is registered AND scoped to want. It distinguishes an unknown key from a
// known key of the wrong scope via the two package sentinel errors, so callers (and tests) can
// tell the two failure modes apart with errors.Is instead of both collapsing into a bare nil.
func resolveScopedProperty(key string, want def.Scope) (def.Entry, property.Property, error) {
	entry, ok := def.Lookup(property.Key(key))
	if !ok {
		return def.Entry{}, nil, fmt.Errorf("%w: %q", ErrUnknownPropertyKey, key)
	}
	if entry.Scopes&want == 0 {
		return def.Entry{}, nil, fmt.Errorf("%w: key %q has scopes %s, expected %s", ErrPropertyKeyWrongScope, key, scopeLabel(entry.Scopes), scopeLabel(want))
	}
	return entry, entry.New(), nil
}

// parseBehaviorType converts the wire string to a def.BehaviorType.
// It defaults to Direct if the behavior is unknown or unspecified.
// This mapping ensures compatibility with the unified skill behavioral model.
// @spec-link [[mechanic_skill_payload_resolution]]
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
// to the appropriate engine property type. Returns true if a value was actually applied,
// and an error if the DTO encodes a value that must be rejected rather than silently
// coerced (ISS-157: an inverted structured Range).
//
// key identifies the resolved property (e.g. property.Range) so this function can special-case
// Range's bare-int authoring shortcut (see below) without affecting any other IntCounter key.
func setSkillPropValue(prop property.Property, dto api.PropertyDTO, key property.Key) (bool, error) {
	// Flag to track if any value was successfully applied from the DTO.
	hasValue := false

	// ISS-157: Range is KindIntCounter with an inverted meaning versus every other IntCounter
	// key — Value is the MINIMUM reachable distance and Max is the MAXIMUM (skill_validation.go
	// rejects a target when dist > GetMaxValue() || dist < GetValue()). Every other IntCounter
	// key (HP/SP/MP/Movement/Shield/Delay/Channeling/Cooldown/Duration) treats a bare int as a
	// plain "the value is N" and leaves Max alone, which is correct for those. For Range, "the
	// value is N" as an author intends "reachable up to N" — value 0, max N — not "min N, max
	// left at the registry default of 1", which produced the inverted, unreachable [N,1] window
	// this issue fixes. This is deliberately gated on the Range key (not "any IntCounter with no
	// Max"): Range is registered ScopeSkill-only (registry_skill_targeting.go), so this branch
	// can never fire for an Entity/Item IntCounter such as HP.
	if key == property.Range && dto.Value != nil && dto.Max == nil {
		if cp, ok := prop.(property.IntCounterProperty); ok {
			cp.SetValue(0)
			cp.SetMaxValue(*dto.Value)
			return true, nil
		}
	}
	// ISS-157: a structured Range payload ({"value":X,"max":Y}) keeps its literal, unchanged
	// [X,Y] meaning. An inverted structured payload (value > max) can never produce a reachable
	// target and is authored nonsense, not a value to silently coerce — crash-early (CODING_RULE
	// §3) by rejecting it instead of letting it fall through as a dead skill.
	if key == property.Range && dto.Value != nil && dto.Max != nil && *dto.Value > *dto.Max {
		return false, fmt.Errorf("inverted skill range: value %d exceeds max %d", *dto.Value, *dto.Max)
	}

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
	return hasValue, nil
}

// buildSkillPropertyMap reconstructs a Targeting or Costs property map from the DTO payload.
// Every key is resolved against the registry scoped to Skill (ISS-140): an unknown key or a
// known key of the wrong scope is collected as an error rather than dropped, so a typo cannot
// silently substitute engine defaults for authored targeting/cost semantics. Per the collect-all
// rule, resolution does not stop at the first bad key — every key is processed in one pass, all
// valid keys land in the returned map, and every rejection is joined into the returned error.
// @spec-link [[mechanic_skill_payload_resolution]]
func buildSkillPropertyMap(raw api.PropertyMap) (map[string]property.Property, error) {
	// Initialize the result map for the engine's internal consumption.
	result := make(map[string]property.Property)
	var errs []error
	for key, dto := range raw {
		entry, prop, err := resolveScopedProperty(key, def.ScopeSkill)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// Attempt to apply the DTO value to the property instance.
		applied, err := setSkillPropValue(prop, dto, entry.Key)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
			continue
		}
		if applied {
			result[entry.Key.String()] = prop
		}
	}
	return result, errors.Join(errs...)
}

// buildSkillEffect reconstructs an effect.Effect from the DTO payload. Every key is resolved
// against the registry scoped to Skill (ISS-140, same resolver pattern as
// buildSkillPropertyMap): an unknown key or a known key of the wrong scope is collected as an
// error rather than dropped. Per the collect-all rule, all keys are processed in one pass.
// @spec-link [[mechanic_skill_payload_resolution]]
func buildSkillEffect(raw api.PropertyMap) (effect.Effect, error) {
	// Create a new effect instance and populate its property collection.
	eff := *effect.New()
	var errs []error
	for key, dto := range raw {
		entry, prop, err := resolveScopedProperty(key, def.ScopeSkill)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		// Map the payload value to the resolved property instance and append it.
		applied, err := setSkillPropValue(prop, dto, entry.Key)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", key, err))
			continue
		}
		if applied {
			eff.Properties = append(eff.Properties, prop)
		}
	}
	// Return the populated effect for skill registration or buff application.
	return eff, errors.Join(errs...)
}
