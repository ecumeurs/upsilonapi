---
id: api_help_endpoint
status: STABLE
tags: api,discovery,meta
human_name: API Help & Discovery Endpoint
layer: ARCHITECTURE
version: 1.0
parents:
  - [[shared:requirement_customer_api_first]]
dependents: []
type: API
priority: 3
---

# New Atom

## INTENT
To provide a live, code-driven index of the entire system surface (REST & WebSockets) by introspecting the Laravel router and ATD dependency graph.

## THE RULE / LOGIC
The endpoint uses **Code-First Introspection** to build the response:
1. **Route Reflection**: Uses `Route::getRoutes()` to identify all `v1` endpoints.
2. **Metadata Extraction**: Reads docblocks to find `@spec-link [[atom_id]]`.
3. **Semantic Enrichment**: Fetches the `## INTENT` section from the linked Atom file.
4. **Input Discovery**: Introspects `FormRequest` classes to extract validation rules and types.
5. **WebSocket Registry**: Hardcoded registry documenting `pusher` protocol handshakes, channel patterns (`private-user`, `private-arena`), and event payloads (`match.found`, `board.updated`).

## TECHNICAL INTERFACE
- **API Endpoint:** `GET /api/v1/help`
- **Code Tag:** `@spec-link [[api_help_endpoint]]`
- **Primary Service:** `App\Services\CodeDiscoveryService`
- **Test Names:** `TestHelpEndpointStructure`, `TestHelpEndpointDiscovery`

## EXPECTATION
- Request returns 200 OK JSON via `ApiResponder`.
- Response contains accurate DTO schemas reflecting semantic flags (`is_self`, etc.) rather than masked UUIDs.
- Response contains:
    - `envelope`: Documentation of the standard response structure.
    - `endpoints`: A non-empty array of introspected API routes with Intent (from Atoms), Input (from FormRequest), and Auth status.
    - `websocket`: Comprehensive registry of Reverb channels and events.
    - `workflow`: Leading text for common onboarding/battle journeys.
- Every endpoint entry MUST have an `atd_link` mapping back to a SPEC atomic ID.
