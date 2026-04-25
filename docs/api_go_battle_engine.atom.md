---
id: api_go_battle_engine
human_name: Go UpsilonBattle JSON API & Webhook Dispatcher
type: MODULE
layer: ARCHITECTURE
version: 1.1
status: STABLE
priority: 5
tags: [api, golang, rest, webhooks]
parents:
  - [[api_standard_envelope]]
dependents:
  - [[api_go_battle_action]]
  - [[api_go_battle_forfeit]]
  - [[api_go_battle_start]]
  - [[api_go_health_check]]
  - [[api_go_webhook_callback]]
  - [[battleui_upsilon_api_service]]
  - [[module_upsilonapi]]
---
# Go UpsilonBattle JSON API & Webhook Dispatcher

## INTENT
To define the external JSON boundary for UpsilonBattle, allowing the Laravel Gateway to instantiate arenas, proxy commands, and receive asynchronous state updates via webhooks.

## THE RULE / LOGIC
The Go Battle Engine API is composed of several specialized endpoints and a webhook dispatcher. It maintains a consolidated 'players' state where each player nests its live entities, including shared 'team' and 'is_self' status to prevent ID leakage. All communications follow the [[api_standard_envelope]].

## TECHNICAL INTERFACE (The Bridge)
- **Base Path:** `/internal`
- **Port:** `8081`
- **Code Tag:** `@spec-link [[api_go_battle_engine]]`

## EXPECTATION (For Testing)
- All endpoints must return a [[api_standard_envelope]] with `success: true` on successful operations.
- All endpoints must handle invalid input by returning a [[api_standard_envelope]] with `success: false` and a descriptive message.
