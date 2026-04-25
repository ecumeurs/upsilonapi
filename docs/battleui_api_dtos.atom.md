---
id: battleui_api_dtos
human_name: BattleUI API Data Transfer Objects
type: DATA
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 5
tags: [battleui, dto, api, types]
parents:
  - [[api_go_battle_action]]
  - [[api_go_battle_start]]
  - [[battleui_upsilon_api_service]]
dependents: []
---
# BattleUI API Data Transfer Objects

## INTENT
To provide strongly-typed representations of the JSON payloads exchanged with the Go Battle Engine, ensuring that Laravel's implementation matches the Go `api` package exactly.

## THE RULE / LOGIC
Defines the Data Transfer Objects for the battle system. 

### BoardState
Contains `players`, `grid`, `turn`, and the new `action` field.

### Grid
`Grid` is a width-major 2D projection of the engine's 3D grid, exposing the **topmost cell at every `(x, y)` column** (the walkable surface). Caves/underground are not exposed in this iteration.
- `width`: `int` — X columns.
- `height`: `int` — Y rows (grid depth; not elevation).
- `max_height`: `int` — engine Z ceiling. Clients scale vertical rendering against this value.
- `cells`: `Cell[x][y]`.

### Cell
Represents the topmost cell at its `(x, y)` column.
- `entity_id`, `obstacle` — existing semantics.
- `height`: `int` — Z index of this topmost cell (surface elevation). 3D clients use it for terrain; 2D clients may ignore or shade.

### ActionFeedback
Captures standard tactical outcomes:
- `move`: includes `actor_id` and `path`.
- `attack`: includes `actor_id`, `target_id`, `damage`, `prev_hp`, `new_hp`.
- `pass`: includes `actor_id`.

Each Player nests an 'entities' array where tactical stats (HP, position) and identity metadata (team, is_self) are unified. This removes the need for flat entity mapping.

## TECHNICAL INTERFACE (The Bridge)
- **Namespace:** `App\DTOs` or `App\Http\Resources`
- **Code Tag:** `@spec-link [[battleui_api_dtos]]`

## EXPECTATION (For Testing)
- Mapping a Go `BoardState` JSON to `BoardStateDTO` must not lose data.
- All DTOs must be serializable to JSON in a format accepted by the Go `gin` handlers.
