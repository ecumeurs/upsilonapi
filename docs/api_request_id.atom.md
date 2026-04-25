---
id: api_request_id
human_name: API Request Identification
type: MECHANIC
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 5
tags: [api, tracing, uuid, header]
parents: []
dependents:
  - [[api_standard_envelope]]
---
# API Request Identification

## INTENT
To uniquely identify every transaction flow within the Upsilon ecosystem, enabling deterministic tracing across distributed components (Vue, Laravel, Go).

## THE RULE / LOGIC
Every logical interaction sequence MUST be assigned a unique identifier:

1.  **Format:** UUID v7 (for chronological sortability).
2.  **Originator Responsibility:** The entity initiating the primary interaction is responsible for generating the ID.
    - **Frontend (Vue):** Generates it for user actions (clicks, form submits).
    - **Backend Gateway (Laravel):** Forwards the frontend's ID or generates a new one if missing.
    - **Engine (Go):** Maintains the ID in all synchronous responses.
    - **Asynchronous Events:** If Go initiates a new flow (e.g., webhook, broadcast), it generates a *new* UUID v7.
3.  **Transport:**
    - **JSON Envelope:** Present in the `request_id` field of the [[api_standard_envelope]].
    - **HTTP Header:** SHOULD be present in the `X-Request-ID` header. This is optional but highly recommended for major error tracking and load balancer/proxy log correlation.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** Universal Middleware
- **Header Key:** `X-Request-ID`
- **Code Tag:** `@spec-link [[api_request_id]]`
- **Test Names:** `TestRequestIdGeneration`, `TestHeaderPropagation`

## EXPECTATION (For Testing)
- The `request_id` remains constant across a single request-response cycle spanning multiple services.
- `X-Request-ID` value matches the JSON `request_id` if both are present.
- UUIDv7 timestamp component matches the approximate real-world creation time.
