---
id: api_laravel_gateway
human_name: Laravel API Gateway & WebSockets Hub
type: API
layer: ARCHITECTURE
version: 1.1
status: REVIEW
priority: 5
tags: [api, gateway, websockets, proxy, laravel-reverb]
parents: []
dependents:
  - [[api_auth_logout]]
  - [[api_auth_register]]
  - [[api_battle_proxy]]
  - [[api_laravel_health_check]]
  - [[api_profile_export]]
  - [[api_websocket]]
---
# Laravel API Gateway & WebSockets Hub

## INTENT
To define how the Vue.js frontend communicates with the overall ecosystem via Laravel, utilizing HTTP REST for actions/queries and WebSockets (Laravel Reverb) for real-time state streaming.

## THE RULE / LOGIC
**Authentication & Identity (HTTP AuthController):**
- `POST /api/v1/auth/login` -> Authenticates and returns token.
- `POST /api/v1/auth/register` -> Creates user and initial roster.
- `POST /api/v1/auth/update` -> Updates User identity data (address, birth date, nickname).
- `POST /api/v1/auth/password` -> Updates User credentials.
- `GET /api/v1/auth/export` -> GDPR Data portability dump.
- `DELETE /api/v1/auth/delete` -> GDPR Right to be forgotten (soft delete + anonymize).

**Meta-game & Roster (HTTP ProfileController):**
- `GET /api/v1/profile` -> Returns player record (wins, ratio).
- `GET /api/v1/profile/characters` -> List character roster.
- `GET /api/v1/profile/character/{id}` -> Specific character details.

**Battle State & Proxying (HTTP & Websocket):**
- `GET /api/v1/battle/{arena_id}/state` -> Cached board state.
- `POST /api/v1/battle/{arena_id}/action` -> Proxies commands to Go.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `/api/v1/*` and Laravel Reverb Channels.
- **Discovery Tool:** `GET /api/v1/help` (Introspected via `CodeDiscoveryService`)
- **Code Tag:** `@spec-link [[api_laravel_gateway]]`
- **Related Issue:** `ISS-005`, `ISS-007`
- **Test Names:** `TestLoginRoute`, `TestProxyAction`, `TestWebhookUpdatesDatabaseCacheAndBroadcasts`, `TestReverbBroadcasting`

## EXPECTATION (For Testing)
- Vue hits `/action` -> Laravel proxies to Go -> Go validates and pushes to `/webhook` -> Laravel updates `game_matches` JSON -> Laravel Broadcasts `board.updated` -> Vue receives event via WebSocket.
