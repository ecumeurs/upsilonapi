---
id: api_standard_envelope
human_name: Standard JSON Message Envelope
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [api, json, envelope, standard]
parents:
  - [[api_request_id]]
dependents:
  - [[rule_tracing_logging]]
  - [[shared:rule_tracing_logging]]
  - [[api_auth_logout]]
  - [[api_auth_register]]
  - [[api_battle_proxy]]
  - [[api_go_battle_action]]
  - [[api_go_battle_engine]]
  - [[api_go_battle_forfeit]]
  - [[api_go_battle_start]]
  - [[api_go_webhook_callback]]
  - [[api_profile_export]]
---
# Standard JSON Message Envelope

## INTENT
To establish a universal, predictable structure for all JSON exchanges between entities (Vue, Laravel, Go) to guarantee tracability, consistent error handling, and extensibility.

## THE RULE / LOGIC
Every JSON payload transmitted over HTTP or WebSocket between system units MUST conform to the following root structure:

```json
{
  "request_id": "018f5a...", // string (UUIDv7). Detailed in [[api_request_id]].
  "message": "...",         // string: A one-liner intent, status summary, or error message.
  "success": true,          // boolean: Indicates if the operation was successful.
  "data": {},               // object|array|null: The core JSON payload of the query or response.
  "meta": {}                // object: Arbitrary, undocumented side information (optional).
}
```

### Constraints:
*   **Request Identification:** Every envelope MUST carry a `request_id` following the rules defined in [[api_request_id]].

## TECHNICAL INTERFACE (The Bridge)
*   **API Endpoint:** Universal (Global Request/Response Middleware)
*   **Code Tag:** `@spec-link [[api_standard_envelope]]`
*   **Related Issues:** None
*   **Test Names:** `TestJsonEnvelopeValidation`, `TestProxyMaintainsRequestId`

## EXPECTATION (For Testing)
- A request lacking a `request_id` or with invalid UUID format MUST return `success: false` with HTTP 400.
- Malformed JSON or missing mandatory fields in wrapped payloads MUST return `success: false` with HTTP 400.
- The literal key names `request_id`, `message`, `success`, `data`, and `meta` must always exist in a response, even if null/empty.
