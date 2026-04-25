---
id: api_profile_credits
status: STABLE
priority: 3
human_name: Profile Credit Balance API
type: API
layer: ARCHITECTURE
version: 1.0
tags: ["api","profile","credits","economy"]
parents:
  - [[api_laravel_gateway]]
dependents: []
---

# New Atom

## INTENT
To provide a lightweight, dedicated endpoint for retrieving a player's current credit balance for shop display and UI economy widgets.

## THE RULE / LOGIC
1. **Access:** Requires authenticated user session (Sanctum).
2. **Scope:** Returns the total credit balance for the authenticated user account (sum of career earnings minus expenditures).
3. **Caching:** Should be fast-path; Laravel may cache this value if balance updates are low-frequency, but currently reads from the `users` table `credits` column.
4. **Integration:** Primary source for the Neon Shop UI.

## TECHNICAL INTERFACE
- **API Endpoint:** `GET /api/v1/profile/credits`
- **Controller:** `ProfileController@getCredits`
- **Code Tag:** `@spec-link [[api_profile_credits]]`
- **Response Format:** `{"success": true, "data": {"credits": 1234}}`

## EXPECTATION
- Request returns HTTP 200.
- `credits` value matches the database sum of transactions for the user.
- Response time < 200ms.
