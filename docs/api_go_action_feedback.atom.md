---
id: api_go_action_feedback
human_name: "Action Feedback DTO"
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [api, golang, battle, feedback]
parents:
  - [[api_go_battle_action]]
dependents: []
---

# Action Feedback DTO

## INTENT
To provide a standardized structure for technical reporting of tactical results from the battle engine to all interested clients (UI, CLI, Observers).

## THE RULE / LOGIC
The `ActionFeedback` object represents the outcome of exactly ONE tactical action.

### Fields
- `type`: `string` - The logical action category (`move`, `attack`, `skill`, `pass`).
- `actor_id`: `string (UUID)` - The ID of the entity that performed the action.
- `target_id`: `string (UUID)` - (Optional) The ID of the entity targeted by the action.
- `damage`: `int` - (Optional) The raw damage value dealt during an `attack` or damaging `skill`.
- `prev_hp`: `int` - (Optional) The target's HP before the action.
- `new_hp`: `int` - (Optional) The target's HP after the action.
- `path`: `Array<Position>` - (Optional) The exact sequence of coordinates traversed during a `move`.
- `credits`: `Array<CreditAward>` - (Optional) Credits earned from this action. Each award contains `player_id`, `amount`, and `source`.
  - `x`: `int`, `y`: `int`

## TECHNICAL INTERFACE (The Bridge)
- **Go Type:** `api.ActionFeedback`
- **Location:** `upsilonapi/api/output.go`
- **JSON Key:** `action` (within `BoardState`)

## EXPECTATION (For Testing)
- **Completeness:** If `type` is `attack`, `damage` and HP deltas MUST be present.
- **Accuracy:** The HP transition `prev_hp` -> `new_hp` MUST reflect the damage calculation exactly.
- **Consistency:** The `actor_id` MUST correspond to an entity in the `players` roster of the same state update.
