---
id: api_websocket_arena_updates
human_name: "Realtime Arena Updates (Customized)"
type: API
layer: ARCHITECTURE
version: 2.0
status: STABLE
priority: 2
tags: [sse, battle, tactical, updates]
parents:
  - [[api_battle_proxy]]
  - [[api_websocket_game_events]]
dependents: []
has_tests: true
linked_codes:
  - upsilonbattleui/src/Pages/BattleArena.vue:111
  - upsilonbattleui/src/services/game.js:72
  - upsilonbattleui/tests/playwright/battle_arena.spec.ts
---

# Realtime Arena Updates (Private)

## INTENT
To synchronize tactical game state and turn changes to specific participants of a match in real-time on their private streams.

## THE RULE / LOGIC
1. **Delivery**: per-participant frames on each user's bearer-authenticated SSE stream ([[api_websocket]]); the stream is the private channel.
2. **Surgical Privacy Masking**:
   - Frames are rendered per-recipient from the applied board state.
   - **Own Characters**: full `EntityDTO` details.
   - **Opponent/AI Characters**: mask sensitive fields (attributes, logic) while leaving public identifiers.
   - **Identity Identification**: Pre-populates `is_self` and `current_player_is_self` based on the targeted user.
3. **Core Events**:
   - `board.updated` (also `game.started`, `turn.started`, `game.ended`, `game.forfeited`): engine event types pass through as SSE event names.
     - **Payload**: `{"match_id": "uuid", ...BoardState...}` (Flattened), one Standard Envelope.
   - **Replay**: frame id `{match_id}:{version}`; reconnect with `Last-Event-ID` resumes from the persisted snapshot.

## TECHNICAL INTERFACE (The Bridge)
- **Stream:** `GET /api/v1/events`
- **Code Tag:** `@spec-link [[api_websocket_arena_updates]]`
- **Server:** `upsilonhub/internal/gateway/sse` (masking + fan-out), fed by the engine webhook (`/api/webhook/upsilon`)

## EXPECTATION (For Testing)
- Game Action processed -> Engine Webhook hits the hub -> `board.updated` framed per participant.
- Client on match page -> listening on its stream -> Board state updates without refresh.
