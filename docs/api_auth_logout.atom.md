---
id: api_auth_logout
human_name: "Player Logout API"
type: API
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 3
tags: [auth, logout]
parents:
  - [[api_laravel_gateway]]
  - [[api_standard_envelope]]
  - [[uc_auth_logout]]
dependents: []
---

# Player Logout API

## INTENT
To securely terminate an active session by revoking the client's current access token.

## THE RULE / LOGIC
- **URI:** `/api/v1/auth/logout`
- **Verb:** `POST`
- **Intent:** Session Termination
- **Fully Detailed Input:** [] (Requires Authorization Header)
- **Fully Detailed Output:**
  - `success`: (boolean) Confirmation of token revocation.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /api/v1/auth/logout`
- **Code Tag:** `@spec-link [[api_auth_logout]]`
- **Security:** Middleware `auth:sanctum` mandatory.

## EXPECTATION (For Testing)
1. Return `success: true` in the standard envelope.
2. Subsequent requests with the same token must return `401 Unauthorized`.
