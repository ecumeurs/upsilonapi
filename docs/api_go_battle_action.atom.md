---
id: api_go_battle_action
human_name: UpsilonBattle Arena Action API
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [api, golang, battle, action]
parents:
  - [[api_go_battle_engine]]
  - [[api_standard_envelope]]
dependents:
  - [[battleui:battleui_api_dtos]]
  - [[api_go_action_feedback]]
---
# UpsilonBattle Arena Action API

## INTENT
To allow players to perform actions (Move, Attack, Skill) within an active battle arena.

## THE RULE / LOGIC
**Endpoint:** `POST /internal/arena/{id}/action`

### Request (Wrapped in [[api_standard_envelope]])
- `player_id`: `string (UUID)` [MANDATORY]
- `entity_id`: `string (UUID)` [MANDATORY]
- `type`: `string` [MANDATORY] - 'move', 'attack', or 'pass'.
- `target_coords`: `Array<Position>` [MANDATORY for 'move' and 'attack']

### Response (Wrapped in [[api_standard_envelope]])
Standard response with updated entity state or result.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /internal/arena/:id/action`
- **Code Tag:** `@spec-link [[api_go_battle_action]]`
- **Go Handler:** `handler.HandleArenaAction`
- **Request Type:** `api.ArenaActionRequest`
- **Response Map:**
  - `rulermethods.ControllerAttackReply` -> `api.NewEntity(d.Entity)`
  - `rulermethods.ControllerMoveReply` -> `api.NewEntity(d.Entity)`
  - Default -> `stdmessage.DataNil{}`

## EXPECTATION (For Testing)
- Valid `ArenaActionRequest` -> Ruler processes action -> Returns `200 OK`.
- Action target out of range -> Returns `400 Bad Request`.
- Arena ID not found -> Returns `400 Bad Request`.
