---
id: mechanic_skill_payload_resolution
status: DRAFT
human_name: "Skill Payload Resolution & Normalization"
layer: IMPLEMENTATION
tags: ["api","serialization","resilience"]
parents:
  - [[shared:requirement_customer_api_first]]
type: MECHANIC
priority: 2
version: 1.0
dependents: []
---

# Skill Payload Resolution & Normalization

## INTENT
To provide a resilient, polymorphic unmarshaling mechanism for skill properties that can handle both structured DTOs and primitive values, while normalizing platform-specific JSON inconsistencies (like empty arrays representing empty objects).

## THE RULE / LOGIC
1. **Flex Wrapper:** Uses a generic `Flex[T]` wrapper to intercept `[]` and treat it as a zero-value for the underlying type `T`.
2. **Polymorphic Unmarshaling (wire -> engine):**
   - Attempt to unmarshal as `PropertyDTO` struct (matching fields `value`, `fvalue`, `max`, `bvalue`, `svalue`).
   - If that fails or yields an empty DTO, fallback to unmarshaling as primitive types in order: `int`, `float64`, `bool`, `string`.
3. **Normalization:** The `PropertyDTO` struct preserves the original value and optional metadata (like `max` for counters), providing a unified interface for the engine and bridge.
4. **Serialization (engine -> wire):** The reverse direction — mapping an engine property onto a `PropertyDTO` for the wire — is equally in scope of this mechanic, not a separate concern:
   - `IntCounterProperty`, `int`, `bool`, and `string` map onto `PropertyDTO`'s matching fields.
   - A `float64`-valued property populates `PropertyDTO.FValue`.
   - A Zone-typed property is not a primitive: its accessor returns the property object itself, not a scalar. Serializing it means emitting its `PatternType` field onto the DTO, not attempting to coerce it into one of the primitive fallbacks above.
   - Any property type with no defined mapping is a normalization failure, not a value to silently drop: it must not be allowed to serialize to a zero/empty `PropertyDTO` (`{}`), because an empty DTO is rejected by the unmarshal side (point 2 above) as "invalid property format" — turning a generation-time problem into a much later, harder-to-diagnose bind failure at battle start.

## TECHNICAL INTERFACE
- **Type:** `api.Flex[T]`, `api.PropertyDTO`
- **Location:** `upsilonapi/api/input.go`
- **Code Tag:** `@spec-link [[mechanic_skill_payload_resolution]]`

## EXPECTATION
- Empty JSON arrays `[]` sent by Laravel are correctly unmarshaled as empty Go structs or maps, not errors.
- Structured property DTOs (with `value`, `max`, etc.) are prioritized over primitive values during unmarshaling.
- Unmarshaling an invalid type returns a clear "invalid property format" error.
- Serializing a Zone-typed property onto the wire produces a `PropertyDTO` carrying its `PatternType`, not an empty `{}`.
- Serializing a `float64`-valued property onto the wire populates `PropertyDTO.FValue`, not an empty `{}`.
- A round trip (engine generates a property, serializes it to the wire, a client persists and later replays it, the engine unmarshals it back) never produces an empty `{}` DTO for a recognized property type.

**Prior wording (superseded 2026-08-24, ISS-131):** the LOGIC section previously covered only the unmarshal (wire -> engine) direction (points 1-3 above, unchanged) and said nothing about the serialize (engine -> wire) direction, despite a doc comment on the serialization entry point claiming this mechanic for the marshal side too. That asymmetry let the engine's property serializer fall through silently on Zone-typed and float64 inputs, producing the empty `{}` payload at the root of ISS-131. The EXPECTATION section previously listed only the first three bullets above, with no serialize-direction expectations at all.
