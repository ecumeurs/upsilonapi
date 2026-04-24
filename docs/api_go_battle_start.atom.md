---
id: api_go_battle_start
human_name: UpsilonBattle Arena Start API
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [api, golang, battle, initialization]
parents:
  - [[api_go_battle_engine]]
  - [[api_standard_envelope]]
dependents:
  - [[battleui_api_dtos]]
---
# UpsilonBattle Arena Start API

## INTENT
To initialize a new battle arena instance with players, entities, and map data.

## THE RULE / LOGIC
**Endpoint:** `POST /internal/arena/start`

### Request (Wrapped in [[api_standard_envelope]])
- `match_id`: `string (UUID)` [MANDATORY] - Unique identifier for the match.
- `callback_url`: `string` [MANDATORY] - Internal URL for webhook events.
- `players`: `Array<Player>` [MANDATORY] - At least one player required.
  - `id`: `string (UUID)` [MANDATORY]
  - `nickname`: `string` - Player display name.
  - `team`: `int`
  - `ia`: `boolean`
  - `entities`: `Array<Entity>` (See [[entity_character]]) [MANDATORY]
    - `max_hp`: `int` [MANDATORY] - Must be > 0.

### Response (Wrapped in [[api_standard_envelope]])
- `arena_id`: `string (UUID)` - The internally generated Arena ID.
- `initial_state`: `BoardState` (See [[api_go_battle_engine]])

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /internal/arena/start`
- **Code Tag:** `@spec-link [[api_go_battle_start]]`
- **Go Handler:** `handler.HandleArenaStart`
- **Request Type:** `api.ArenaStartRequest`
- **Response Type:** `api.ArenaStartResponse`

## EXPECTATION (For Testing)
- Valid `ArenaStartRequest` -> Returns `200 OK` with `ArenaStartResponse`.
- Invalid JSON or missing required fields -> Returns `400 Bad Request` with `Success: false`.
