---
id: api_go_battle_start
human_name: UpsilonBattle Arena Start API
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [api, golang, battle, initialization]
parents:
  - [[api_go_battle_engine]]
  - [[api_standard_envelope]]
dependents:
  - [[upsilonbattleui:entity_api_dtos]]
---
# UpsilonBattle Arena Start API

## INTENT
To initialize a new battle arena instance with players, entities, and map data.

## THE RULE / LOGIC
**Endpoint:** `POST /v1/arena/start`

### Request (Wrapped in [[api_standard_envelope]])
- `match_id`: `string (UUID)` [MANDATORY]
- `callback_url`: `string` [MANDATORY]
- `players`: `Array<Player>` [MANDATORY]
  - `id`: `string (UUID)`
  - `nickname`: `string`
  - `team`: `int`
  - `ia`: `boolean`
  - `archetype`: `string` (Optional: default for entities)
  - `total_wins`: `int` (For IA grade scaling)
  - `entities`: `Array<Entity>`
    - `max_hp`: `int`
    - `auto_gen`: `boolean` (True -> Procedural generation)
    - `archetype`: `string` (Optional override)

### Response (Wrapped in [[api_standard_envelope]])
- `arena_id`: `string (UUID)`
- `initial_state`: `BoardState`

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /v1/arena/start`
- **Code Tag:** `@spec-link [[api_go_battle_start]]`
- **Go Handler:** `handler.HandleArenaStart`
- **Request Type:** `api.ArenaStartRequest`
- **Response Type:** `api.ArenaStartResponse`

## EXPECTATION (For Testing)
- Valid `ArenaStartRequest` -> Returns `200 OK` with `ArenaStartResponse`.
- Invalid JSON or missing required fields -> Returns `400 Bad Request` with `Success: false`.
- A well-formed request that the engine rejects on rule grounds (e.g. a skill payload carrying an unrecognized or malformed property shape) -> Returns a non-2xx status carrying a well-formed envelope with `Success: false`. This is a rule rejection, not a transport error: callers MUST treat `Success: false` as failure regardless of HTTP status, MUST NOT create or cache any arena state from the response, and MUST NOT infer success merely from the absence of a network/decode error.

**Prior wording (superseded 2026-08-24, ISS-131):** the EXPECTATION section previously covered only the two bullets above (valid request / malformed request) and said nothing about the engine-rule-rejection-with-valid-envelope case — a well-formed request the engine refuses for rule reasons (not a request-shape problem). That gap let a caller check only the transport-level error and never inspect `Success`, which is exactly how ISS-131's phantom-match bug happened.
