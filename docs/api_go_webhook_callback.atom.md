---
id: api_go_webhook_callback
human_name: UpsilonBattle Webhook Callback
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [api, golang, callback, webhooks]
parents:
  - [[api_go_battle_engine]]
  - [[api_standard_envelope]]
dependents:
  - [[mech_game_state_versioning]]
---
# UpsilonBattle Webhook Callback

## INTENT
To asynchronously notify the Laravel Gateway of state changes, turn start/end, and battle results.

## THE RULE / LOGIC
**Destination:** The `callback_url` provided during [[api_go_battle_start]].

### ArenaEvent Payload (Wrapped in [[api_standard_envelope]])
- `match_id`: `string (UUID)` - Targeted match in Laravel.
- `event_type`: `string` ("game.started", "turn.started", "board.updated", "game.ended")
- `player_id`: `string (UUID)` (if applicable)
- `entity_id`: `string (UUID)` (if applicable)
- `data`: `BoardState` (See [[api_go_battle_engine]]) - **Note:** Now includes the full `players` roster for identity synchronization.
- `action`: `ActionFeedback` (See [[api_go_action_feedback]]) - **Optional:** The specific tactical result that triggered this update.
- `version`: `int64` - Monotonic sequence number synced with `data.sequence`.
- `timeout`: `string (ISO8601)` - End of the current turn clock.

### Event Types:
- `game.started`: Arena initialization complete.
- `turn.started`: New entity initiative active (starts 30s clock).
- `board.updated`: Position or stat change.
- `game.ended`: Win condition met.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[api_go_webhook_callback]]`
- **Go Dispatcher:** `bridge.HTTPController.forwardToWebhook`
- **Payload Type:** `api.ArenaEvent` (in some paths) or `map[string]interface{}` (in `forwardToWebhook`).

## EXPECTATION (For Testing)
- Ruler broadcasts `BattleStart` -> Dispatcher sends `POST` to `callback_url` with `event_type: "game.started"`.
- Dispatcher should handle non-200 responses from the callback URL (though current implementation just logs it).
