---
id: api_battle_proxy
human_name: Battle Proxy & Webhook API
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [battle, proxy, webhook, api]
parents:
  - [[api_laravel_gateway]]
  - [[api_standard_envelope]]
dependents:
  - [[api_websocket_arena_updates]]
---
# Battle Proxy & Webhook API

## INTENT
To facilitate communication between the player and the core game engine for active matches.

## THE RULE / LOGIC
- **Endpoint 1: Get Game State**
  - **URI:** `/api/v1/game/{match_id}`
  - **Verb:** `GET`
  - **Intent:** Tactical Synchronization
  - **Input:** 
    - `match_id`: (uuid) [Mandatory] The active arena identifier.
  - **Output:** Current board state, entity positions, and turn order.

- **Endpoint 2: Perform Action**
  - **URI:** `/api/v1/game/{match_id}/action`
  - **Verb:** `POST`
  - **Intent:** Command Transmission
  - **Input:** 
    - `entity_id`: (string) [Mandatory] The acting character identifier.
    - `type`: (string) [Mandatory] Action type must be **LOWERCASE**: 'move', 'attack', 'skill', or 'pass'.
    - `skill_id`: (string) [Mandatory for 'skill'] The skill identifier.
    - `target_coords`: (object) [Optional] Extra data like coordinates or target IDs.
  - **Output:** `{ "success": true, "result": "action_processed" }`

- **Endpoint 3: Forfeit Match**
  - **URI:** `/api/v1/game/{match_id}/forfeit`
  - **Verb:** `POST`
  - **Intent:** Sudden Concession
  - **Logic:** Defined in [[rule_forfeit_battle]].
  - **Output:** Standard success envelope.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `/api/v1/game/*`, `/api/webhook/*`
- **Code Tag:** `@spec-link [[api_battle_proxy]]`
- **Related Issue:** `ISS-007`
- **Test Names:** `TestActionProxying`, `TestWebhookUpdatesStateAndBroadcasts`

## EXPECTATION (For Testing)
- Action forwarded to Go carries the same `request_id`.
- Webhook receipt triggers `BoardUpdated` event in Laravel.
