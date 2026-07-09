---
id: api_websocket_game_events
status: STABLE
layer: ARCHITECTURE
version: 2.0
parents:
  - [[api_websocket]]
dependents:
  - [[api_websocket_arena_updates]]
  - [[api_websocket_user_notifications]]
type: API
priority: 3
tags: sse,game,events,real-time
human_name: Realtime Game Events Registry
---

# Realtime Game Events Registry

## INTENT
To define the real-time event registry and payload contracts for game synchronization and player notifications, prioritizing private user streams for tactical integrity.

## THE RULE / LOGIC
1. **Event Dispatching**:
   - Transport concerns (auth, framing, replay, backpressure) are handled at the master protocol level [[api_websocket]].
   - Game-logic events are dispatched per recipient on each participant's bearer-authenticated stream.
2. **Stream Mapping**:
   - The user's single `/api/v1/events` stream carries [[api_websocket_user_notifications]] AND tactical [[api_websocket_arena_updates]] — there is one stream per session, not one per channel.
3. **Common Event Lifecycle**:
   - `match.found` (transient, no replay id) -> Client identifies match and navigates to the arena.
   - `board.updated` (id `{match_id}:{version}`) -> Client receives customized, surgically masked tactical state.
   - `game.started` / `turn.started` / `game.ended` / `game.forfeited` -> engine lifecycle passthrough.

## TECHNICAL INTERFACE
- **Event Registry:** `match.found`, `board.updated`, `game.started`, `turn.started`, `game.ended`, `game.forfeited`.
- **Protocol:** SSE `text/event-stream` (See [[api_websocket]])
- **Code Tag:** `@spec-link [[api_websocket_game_events]]`

## EXPECTATION
- Event `match.found` received by client -> Bot initializes match state.
- Event `board.updated` received by client -> Bot/UI updates local board display.
- Dropped stream -> client reconnects with `Last-Event-ID` and resumes from snapshot replay.
