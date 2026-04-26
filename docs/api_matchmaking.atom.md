---
id: api_matchmaking
human_name: Matchmaking API
type: API
layer: ARCHITECTURE
version: 1.1
status: STABLE
priority: 5
tags: [matchmaking, queue, api]
parents:
  - [[shared:uc_matchmaking]]
dependents: []
---
# Matchmaking API

## INTENT
To manage survivor entry and exit from competitive and cooperative matchmaking queues.

## THE RULE / LOGIC
- **Endpoint 1: Join Queue**
  - **URI:** `/api/v1/matchmaking/join`
  - **Verb:** `POST`
  - **Intent:** Enter Search Pool
  - **Input:** 
    - `game_mode`: (string) [Mandatory] '1v1_PVP', '1v1_PVE', '2v2_PVP', '2v2_PVE'.
  - **Output:** `{ "status": "queued", "estimated_wait": 30 }`

- **Endpoint 2: Leave Queue**
  - **URI:** `/api/v1/matchmaking/leave`
  - **Verb:** `DELETE`
  - **Intent:** Exit Search Pool
  - **Input:** []
  - **Output:** `{ "status": "idle" }`

- **Endpoint 3: Match Status**
  - **URI:** `/api/v1/matchmaking/status`
  - **Verb:** `GET`
  - **Intent:** Poll Search State
  - **Input:** []
  - **Output:** `{ "status": "queued|matched|idle", "match_id": "optional-uuid" }`

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `/api/v1/matchmaking/*`
- **Code Tag:** `@spec-link [[api_matchmaking]]`
- **Related Issue:** `ISS-007`
- **Test Names:** `TestJoinQueue`, `TestLeaveQueue`, `TestMatchFinding`

## EXPECTATION (For Testing)
- Join -> Player ID and characters stored in database `matchmaking_pool` (or equivalent persistent store).
- Leave -> Entry removed from database.
- Two compatible entries in pool -> Call Go `arena/start` -> Broadcast `game.started` to both.
