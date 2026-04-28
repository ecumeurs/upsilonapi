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
To allow players to perform tactical actions (Move, Attack, Skill) within an active battle arena.

## THE RULE / LOGIC
**Endpoint:** `POST /internal/arena/{id}/action`
**Payload:** `api.ArenaActionRequest` (contains `type`, `entity_id`, `player_id`, `target_coords`, and optional `skill_id`).

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /internal/arena/:id/action`
- **Code Tag:** `@spec-link [[api_go_battle_action]]`
- **Go Handler:** `handler.HandleArenaAction`
- **Request Type:** `api.ArenaActionRequest`
- **Response Map:**
  - `Attack/Skill`: Returns `gin.H` with `results` (ActionResults) and `attacker`/`entity`.
  - `Move`: Returns `gin.H` with `results` (path/HP deltas) and `entity`.
  - `Credits`: Mapped into `results` or top-level depending on action type.
- **Versioning:** No `Version` field in synchronous reply (broadcast only).

## EXPECTATION (For Testing)
- Valid `ArenaActionRequest` -> Ruler processes action -> Returns `200 OK`.
- Action target out of range -> Returns `400 Bad Request`.
- Arena ID not found -> Returns `400 Bad Request`.
