---
id: module_upsilonapi
human_name: UpsilonAPI Module
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [architecture, module, api, bridge]
parents:
  - [[api_go_battle_engine]]
dependents: []
---
# UpsilonAPI Module

## INTENT
To provide a scalable HTTP bridge for the UpsilonBattle engine, enabling external orchestration by the Laravel Gateway.

## THE RULE / LOGIC
- **Bridge Pattern:** It must maintain a registry of active `Ruler` instances mapped to `arena_id`.
- **Stateless HTTP:** The HTTP endpoints must be thin wrappers around actor message submissions.
- **Event Lifecycle:**
  - `START`: Creates a `Ruler` instance, adds `HTTPController`s, and returns the initial state.
  - `ACTION`: Proxies an action to the corresponding `Ruler`.
  - `EVENT`: Listens for `Ruler` broadcasts and pushes them to the registered `callback_url`.

## TECHNICAL INTERFACE (The Bridge)
Test Names: TestBattleFullRoundtrip

## EXPECTATION (For Testing)
- High availability for the HTTP server.
- Guaranteed delivery of webhooks (with retry logic).
- Correct mapping of multiple concurrent arenas.
