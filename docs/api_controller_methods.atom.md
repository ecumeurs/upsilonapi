---
id: api_controller_methods
human_name: Controller Message Methods API
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [api, controller, messaging]
parents:
  - [[api_ruler_methods]]
dependents:
  - [[mech_controller_communication_sequence]]
  - [[mech_controller_handshake]]
---
# Controller Message Methods API

## INTENT
To define the messages that the Ruler (or other actors) can send to a Controller.

## THE RULE / LOGIC
Controllers must handle the following actor-message structures:

- **SetQueue**: Sent by the Ruler during registration. Contains the Ruler's actor reference.
- **Send**: A generic "fire and forget" message for triggering controller-side logic.
- **ReceiveAPIMessage**: Used to pass raw API data to the controller if bridged.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[api_controller_methods]]`
- **Related Issue:** `#None`
- **Test Names:** `TestRulerBattleBegin`

## EXPECTATION (For Testing)
- Controller receives `SetQueue` -> Stores `Ruler` reference -> Can now send messages to Ruler.
