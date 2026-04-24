---
id: api_websocket
human_name: "WebSocket Protocol (Master)"
type: API
layer: ARCHITECTURE
version: 1.1
status: STABLE
priority: 3
tags: [websocket, real-time, api, pusher, reverb]
parents:
  - [[api_laravel_gateway]]
dependents:
  - [[api_websocket_game_events]]
---

# WebSocket Protocol (Master)

## INTENT
To define the low-level communication contract and authorization handshake for all real-time bidirectional traffic via Laravel Reverb.

## THE RULE / LOGIC
1. **Transport**: Pure WebSocket (WSS in production, WS in dev).
2. **Protocol Wrapper**: Pusher v7 compatible JSON messages.
3. **Authorization Handshake**:
   - Private channels REQUIRE a signature via `POST /broadcasting/auth`.
   - Payload: `socket_id` (from handshake) and `channel_name`.
   - Header: `Authorization: Bearer {JWT}`.
4. **Heartbeats (Stability)**:
   - Server may send `pusher:ping`. Client MUST reply with `pusher:pong`.
   - Client SHOULD send `pusher:ping` during long periods of inactivity to prevent connection eviction (Error 4201).

## TECHNICAL INTERFACE (The Bridge)
- **Base URL (Dev):** `ws://127.0.0.1:8080`
- **Auth URL:** `/broadcasting/auth`
- **Registry:** `GET /api/v1/help` -> `websocket` section.
- **Code Tag:** `@spec-link [[api_websocket]]`

## EXPECTATION (For Testing)
- `REVERB_APP_KEY` provided -> Handshake successful -> `socket_id` received.
- Valid Signature obtained -> `pusher:subscribe` accepted by server.
- Server `pusher:ping` received -> Client `pusher:pong` sent -> Connection remains ACTIVE.
