---
id: api_websocket_user_notifications
human_name: "WebSocket User Notifications (Private)"
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [websocket, matchmaking, notifications]
parents:
  - [[api_matchmaking]]
  - [[api_websocket_game_events]]
dependents: []
---

# WebSocket User Notifications (Private)

## INTENT
To provide authenticated, user-specific tactical state updates and account-level notifications.

## THE RULE / LOGIC
1. **Channel Name**: `private-user.{ws_channel_key}`
   - `{ws_channel_key}` is the pseudonym provided in the `UserResource`.
2. **Authorization**: Only the owner of the user account can subscribe via Sanctum-authenticated `/broadcasting/auth`.
3. **Core Tactical Events**:
   - `match.found`: Triggered when a match is successfully paired.
   - `board.updated`: Triggered for every tactical state change (Movement, Combat, Pass).
     - **Masking**: Tactical board events on this channel are surgically masked for the specific recipient.
     - **Payload**: Includes unmarshalled `BoardState` DTOs.

## TECHNICAL INTERFACE (The Bridge)
- **Channel Pattern:** `private-user.*`
- **Code Tag:** `@spec-link [[api_websocket_user_notifications]]`
- **Laravel Events:** `App\Events\MatchFound`, `App\Events\BoardUpdated`

## EXPECTATION (For Testing)
- User logs in -> Subscribes to `private-user.{id}` -> Signature valid.
- Matchmaking pairs player -> Event `match.found` received by client.
