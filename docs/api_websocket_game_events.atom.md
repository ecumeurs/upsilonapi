---
id: api_websocket_game_events
status: STABLE
layer: ARCHITECTURE
version: 1.0
parents:
  - [[api_websocket]]
dependents:
  - [[api_websocket_arena_updates]]
  - [[api_websocket_user_notifications]]
type: API
priority: 3
tags: websocket,game,events,real-time
human_name: WebSocket Game Events Registry
---

# New Atom

## INTENT
To define the real-time event registry and payload contracts for game synchronization and player notifications, prioritizing private user streams for tactical integrity.

## THE RULE / LOGIC
1. **Event Dispatching**:
   - Authentication-related events (`pusher:*`) are handled at the master protocol level [[api_websocket]].
   - Game-logic events are dispatched on private channels.
2. **Channel Mapping**:
   - `private-user.*` channels carry [[api_websocket_user_notifications]] AND tactical [[api_websocket_arena_updates]].
   - `private-arena.*` channels carry common, shared events (e.g. Chat, Emojis).
3. **Common Event Lifecycle**:
   - `match.found` (User Channel) -> Client identifies match and initiates tactical streams.
   - `board.updated` (User Channel) -> Client receives customized, surgically masked tactical state.
   - `common.event` (Arena Channel) -> Client receives shared non-sensitive events.

## TECHNICAL INTERFACE
- **Event Registry:** `match.found`, `board.updated`, `game.started`, `turn.started`.
- **Protocol:** Pusher v7 compatible (See [[api_websocket]])
- **Code Tag:** `@spec-link [[api_websocket_game_events]]`

## EXPECTATION
- Event `match.found` received by client -\u003e Bot initializes match state.
- Event `board.updated` received by client -\u003e Bot/UI updates local board display.
- Server `pusher:ping` received -\u003e Client `pusher:pong` sent.
