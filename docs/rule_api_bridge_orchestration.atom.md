---
id: api_bridge_orchestration
status: DRAFT
priority: 2
parents:
  - [[contract_api_contract]]
type: RULE
version: 1.0
tags: [api,bridge,orchestration]
dependents: []
human_name: "API Bridge Orchestration"
layer: BUSINESS
---

# New Atom

## INTENT
Define the orchestration logic for bridging HTTP requests to the battle engine.

## THE RULE / LOGIC
- **Ruler Registry:** Maintain a map of active BattleArenas.
- **Request Proxying:** Transform incoming HTTP JSON payloads into internal engine command structures.
- **Event Forwarding:** Capture engine events and push them to the registered webhook callback.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[api_bridge_orchestration]]`

## EXPECTATION
The API must maintain a 1:1 mapping between HTTP requests and engine actor states for every active arena.
