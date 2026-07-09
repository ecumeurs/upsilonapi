---
id: api_websocket
human_name: "Realtime Stream Protocol (Master)"
type: API
layer: ARCHITECTURE
version: 2.0
status: STABLE
priority: 3
tags: [sse, real-time, api, events]
parents:
  - [[api_laravel_gateway]]
dependents:
  - [[api_websocket_game_events]]
---

# Realtime Stream Protocol (Master)

## INTENT
To define the low-level communication contract for all real-time server→client traffic via Server-Sent Events (SSE), served by the hub.

## THE RULE / LOGIC
1. **Transport**: HTTP `text/event-stream` (`GET /api/v1/events`). One long-lived response per client; the server never reads from it (server→client only).
2. **Authorization**: standard `Authorization: Bearer {token}` on the stream request. The connection itself is the user's private channel — there is no channel-auth handshake and no per-user channel key.
3. **Framing**: each frame is `id:` (optional) + `event:` (engine event type passthrough: `board.updated`, `turn.started`, `game.ended`, …, plus `match.found`) + `data:` (one Standard Envelope JSON).
4. **Replay (Stability)**: frame ids are `{match_id}:{version}`. On reconnect the client sends `Last-Event-ID`; the server replays from the persisted board snapshot when the cursor is stale. Transient events (`match.found`) carry no id so the cursor always points at replayable board state.
5. **Backpressure**: a consumer that falls behind the per-connection buffer is disconnected and expected to reconnect with `Last-Event-ID` — webhook ingestion is never blocked on a slow socket.

## TECHNICAL INTERFACE (The Bridge)
- **Stream URL:** `GET /api/v1/events` (front door `:8085`, hub `:8090`)
- **Server:** `upsilonhub/internal/gateway/sse` (fan-out + masking), mounted in `upsilonhub/internal/gateway/router.go`
- **Code Tag:** `@spec-link [[api_websocket]]`

## EXPECTATION (For Testing)
- Valid bearer token → 200 `text/event-stream`, frames arrive unbuffered as events apply.
- Missing/invalid token → enveloped 401; no stream.
- Reconnect with `Last-Event-ID` older than current board version → snapshot replay before live frames.
