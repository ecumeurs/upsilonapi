---
id: api_go_battle_forfeit
human_name: "UpsilonBattle Arena Forfeit API"
type: API
layer: ARCHITECTURE
version: 1.0
status: DRAFT
priority: 5
tags: [api, golang, battle, forfeit]
parents:
  - [[api_go_battle_engine]]
  - [[api_standard_envelope]]
dependents:
  - [[battleui_upsilon_api_service]]
---

# UpsilonBattle Arena Forfeit API

## INTENT
To allow a player to concede a match through a dedicated endpoint that does not require an entity context.

## THE RULE / LOGIC
**Endpoint:** `POST /internal/arena/{id}/forfeit`

### Request (Wrapped in [[api_standard_envelope]])
- `player_id`: `string (UUID)` [MANDATORY]

### Response (Wrapped in [[api_standard_envelope]])
Standard success envelope. Triggers an immediate `game.ended` event with the remaining team marked as winner.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /internal/arena/:id/forfeit`
- **Code Tag:** `@spec-link [[api_go_battle_forfeit]]`
- **Go Handler:** `handler.HandleArenaForfeit`
- **Request Type:** `api.ArenaForfeitRequest`

## EXPECTATION (For Testing)
- Valid `player_id` for an active participant -> Match ends -> Returns `200 OK`.
- `player_id` does not belong to the match -> Returns `403 Forbidden`.
- Arena ID not found -> Returns `400 Bad Request`.
