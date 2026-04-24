---
id: api_go_health_check
human_name: Go Engine Health Check Endpoint
type: API
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 3
tags: [api, health, ci, docker]
parents:
  - [[api_go_battle_engine]]
dependents: []
---
# Go Engine Health Check Endpoint

## INTENT
To provide a lightweight readiness probe for Docker healthchecks and CI orchestration, confirming the Go engine is booted and serving HTTP requests.

## THE RULE / LOGIC
- `GET /health` returns HTTP 200 with a JSON body containing `{"status": "ok", "revision": "<git_hash>"}`.
- The endpoint is unauthenticated and publicly accessible.
- The revision is extracted from Go's `debug.ReadBuildInfo()` VCS metadata at startup.
- If the endpoint responds, the engine is considered healthy and ready to accept arena requests.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `GET /health`
- **Port:** `8081`
- **Code Tag:** `@spec-link [[api_go_health_check]]`
- **Docker Usage:** `curl -f http://localhost:8081/health`

## EXPECTATION (For Testing)
- `GET /health` must return HTTP 200 with `{"status": "ok"}`.
- The response must include a non-empty `revision` field.
- The endpoint must respond within 1 second under normal conditions.
