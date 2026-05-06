---
id: api_websocket_arena_updates
human_name: "WebSocket Arena Updates (Customized)"
type: API
layer: ARCHITECTURE
version: 1.1
status: STABLE
priority: 2
tags: [websocket, battle, tactical, updates]
parents:
  - [[api_battle_proxy]]
  - [[api_websocket_game_events]]
dependents: []
has_tests: true
linked_codes:
  - battleui/resources/js/Pages/BattleArena.vue:42
  - battleui/resources/js/services/game.js:50
  - battleui/tests/playwright/battle_arena.spec.ts
---

# WebSocket Arena Updates (Private)

## INTENT
To synchronize tactical game state and turn changes to specific participants of a match in real-time on their private notification channels.

## THE RULE / LOGIC
1. **Channel Name**: `private-user.{ws_channel_key}`
   - Tactical updates are sent to the private notification channel of each user.
2. **Authorization**: Managed via `user.{key}` private channel rules (Sanctum/ws_channel_key).
3. **Surgical Privacy Masking**:
   - Updates are triggered per-user using `BoardStateResource`.
   - **Own Characters**: Broadcast full `EntityDTO` details.
   - **Opponent/AI Characters**: Mask sensitive fields (attributes, logic) while leaving public identifiers.
   - **Identity Identification**: Pre-populates `is_self` and `current_player_is_self` based on the targeted user.
4. **Core Events**:
   - `board.updated`: Triggered by engine change.
     - **Payload**: `{"match_id": "uuid", ...BoardState...}` (Flattened)

## TECHNICAL INTERFACE (The Bridge)
- **Channel Pattern:** `private-user.*`
- **Code Tag:** `@spec-link [[api_websocket_arena_updates]]`
- **Laravel Event:** `App\Events\BoardUpdated`
- **Pseudonym:** Uses the `ws_channel_key` mapped to the User ID.

## EXPECTATION (For Testing)
- Game Action processed -> Engine Webhook hits BattleUI -> `board.updated` broadcasted.
- Client on match page -> Subscribed to `private-arena.{id}` -> Board state updates without refresh.
