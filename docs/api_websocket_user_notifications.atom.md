---
id: api_websocket_user_notifications
human_name: "Realtime User Notifications (Private)"
type: API
layer: ARCHITECTURE
version: 2.0
status: STABLE
priority: 3
tags: [sse, matchmaking, notifications]
parents:
  - [[api_matchmaking]]
  - [[api_websocket_game_events]]
dependents:
  - [[upsilonbattleui:ui_dashboard_matchmaking]]
has_tests: true
linked_codes:
  - upsilonbattleui/src/Components/Dashboard/EngagementHub.vue:94
  - upsilonbattleui/tests/playwright/user_flows.spec.ts
---

# Realtime User Notifications (Private)

## INTENT
To provide authenticated, user-specific tactical state updates and account-level notifications.

## THE RULE / LOGIC
1. **Delivery**: on the user's bearer-authenticated SSE stream (`GET /api/v1/events`, [[api_websocket]]). The connection itself is the private channel — no channel-auth handshake, no per-user channel key. Client-side "channel" names (`user.{account_name}`) are local subscription bookkeeping only.
2. **Core Tactical Events**:
   - `match.found`: Triggered when a match is successfully paired (transient — no replay id).
   - `board.updated`: Triggered for every tactical state change (Movement, Combat, Pass).
     - **Masking**: Tactical board events are surgically masked for the specific recipient before framing.
     - **Payload**: one Standard Envelope with the recipient's `BoardState` view.

## TECHNICAL INTERFACE (The Bridge)
- **Stream:** `GET /api/v1/events`
- **Code Tag:** `@spec-link [[api_websocket_user_notifications]]`
- **Server:** `upsilonhub/internal/gateway/sse` fan-out, `upsilonhub/internal/games/battle` MatchFound publisher

## EXPECTATION (For Testing)
- User logs in -> opens the events stream -> connection LED goes private-linked.
- Matchmaking pairs player -> Event `match.found` received by that user's stream only.
