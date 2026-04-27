---
id: mech_webhook_delivery
human_name: "Asynchronous Webhook Delivery"
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 3
tags: [webhook, network, performance]
parents:
  - [[upsilonbattle:rule_credit_action_communication_layer]]
dependents: []
---

# Asynchronous Webhook Delivery

## INTENT
To deliver game events and state updates to the callback URL (Laravel) without blocking the engine's actor processing loop.

## THE RULE / LOGIC
- **Dispatch:** Every event intended for the external API must be dispatched in a non-blocking background goroutine.
- **Error Handling:** Network failures or non-200 responses must be logged as warnings but must NOT interrupt the engine's internal state machine.
- **Battle End Lifecycle:** For `game.ended` events, the background task is responsible for triggering `DestroyArena` after the notification is sent.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[mech_webhook_delivery]]`
- **Implementation:** `upsilonapi/bridge/http_controller.go`

## EXPECTATION (For Testing)
- Engine continues processing turns even if the webhook receiver is artificially delayed by 10 seconds.
- MatchId remains active in the engine until the `game.ended` webhook is dispatched.
