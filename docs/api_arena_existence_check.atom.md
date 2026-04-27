---
id: api_arena_existence_check
human_name: "Arena Existence Check API"
type: API
layer: ARCHITECTURE
version: 1.0
status: DRAFT
priority: 3
tags: [api, golang, battle, arena, existence]
parents:
  - [[api_go_battle_engine]]
dependents: []
---

# Arena Existence Check API

## INTENT
To verify if a specific battle arena instance (identified by its UUID) currently exists in the engine's memory. This is useful for external services to synchronize state or check match validity.

## THE RULE / LOGIC
1. **Endpoint:** `GET /v1/arena/:id/exists`
2. **Authorization:** Internal only (as per `api_go_battle_engine`).
3. **Response Structure:**
   - Always returns HTTP 200 on success.
   - Follows [[api_standard_envelope]].
   - `data.exists` is `true` if the arena is active in the engine, `false` otherwise.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `GET /v1/arena/:id/exists`
- **Code Tag:** `@spec-link [[api_arena_existence_check]]`
- **Go Handler:** `handler.HandleArenaExists`
- **Response Type:** `api.ArenaExistsResponse`

## EXPECTATION (For Testing)
- Valid UUID of an active match -> Returns `{"exists": true}`.
- Valid UUID of a finished/non-existent match -> Returns `{"exists": false}`.
- Invalid UUID format -> Returns HTTP 400 with error message.
