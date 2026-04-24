---
id: api_laravel_health_check
human_name: "Laravel Health Check Endpoint"
type: API
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 2
tags: [devops, health, connectivity]
parents:
  - [[api_laravel_gateway]]
dependents: []
---

# Laravel Health Check Endpoint

## INTENT
To provide a lightweight readiness probe for Docker orchestration and CI pipelines, confirming the Laravel application is booted and serving HTTP requests.

## THE RULE / LOGIC
- `GET /up` must return HTTP 200 with an empty body or simple "OK" status.
- The endpoint is registered via `health: '/up'` in `bootstrap/app.php`.
- The endpoint must NOT be intercepted by the web catch-all router.
- Success indicates that the service providers are loaded and the app is ready for requests.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `GET /up`
- **Code Tag:** `@spec-link [[api_laravel_health_check]]`
- **Docker Usage:** `curl -f http://localhost/up`

## EXPECTATION (For Testing)
- `GET /up` returns status 200.
- Response time is < 500ms.
- Does not require authentication or session state.
