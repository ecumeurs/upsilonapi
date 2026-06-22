---
id: mechanic_skill_payload_resolution
status: DRAFT
human_name: "Skill Payload Resolution & Normalization"
layer: IMPLEMENTATION
tags: ["api","serialization","resilience"]
parents:
  - [[shared:rule_dto_strict_typing]]
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
2. **Polymorphic Unmarshaling:**
   - Attempt to unmarshal as `PropertyDTO` struct (matching fields `value`, `fvalue`, `max`, `bvalue`, `svalue`).
   - If that fails or yields an empty DTO, fallback to unmarshaling as primitive types in order: `int`, `float64`, `bool`, `string`.
3. **Normalization:** The `PropertyDTO` struct preserves the original value and optional metadata (like `max` for counters), providing a unified interface for the engine and bridge.

## TECHNICAL INTERFACE
- **Type:** `api.Flex[T]`, `api.PropertyDTO`
- **Location:** `upsilonapi/api/input.go`
- **Code Tag:** `@spec-link [[mechanic_skill_payload_resolution]]`

## EXPECTATION
- Empty JSON arrays `[]` sent by Laravel are correctly unmarshaled as empty Go structs or maps, not errors.
- Structured property DTOs (with `value`, `max`, etc.) are prioritized over primitive values during unmarshaling.
- Unmarshaling an invalid type returns a clear "invalid property format" error.
