---
id: infra_mvp_docker
human_name: MVP Docker Infrastructure
type: MECHANIC
layer: IMPLEMENTATION
version: 2.0
status: STABLE
priority: 5
tags: [docker, infrastructure, mvp]
parents:
  - [[upsilonbattle:module_backend]]
dependents:
  - [[upsiloncli:watch_services]]
---
# MVP Docker Infrastructure

## INTENT
Provide a lightweight, development-friendly Docker orchestration for the Upsilon system MVP.

## THE RULE / LOGIC
- **Base Images**:
  - Hub: `golang:1.25-alpine` builder + `node:22-alpine` SPA stage → distroless static runtime (Custom Dockerfile in ./upsilonhub, umbrella-root context)
  - Go Engine: `golang:1.25-alpine` (Custom Dockerfile in ./upsilonapi)
  - Front door: `caddy:2-alpine`
  - Database: `postgres:18-alpine`
- **Service Orchestration**:
  - `hub`: Go platform gateway serving API + SSE + SPA on `:8090` (internal).
  - `proxy`: Caddy front door. Port `8000:8085` (prod) / `8085:8085` (dev/CI).
  - `engine`: Go battle engine. Port `8081:8081`.
  - `db`: PostgreSQL.
  - `cli`: On-stack maintenance/tester container.
- **Data Persistence**:
  - Named volume `db_data` for PostgreSQL. Ensures data survives restarts and shutdowns.
- **Environment**:
  - Managed via root `.env` file generated from `env.example` (DATABASE_URL, ADMIN_INITIAL_PASSWORD, UPSILON_* endpoints).
- **Initialization**:
  - `db-init` service runs the hub image with `-migrate-mode baseline` (prod: adopts the Laravel-migrated schema, then applies newer migrations) or `-migrate-mode full` + `-seed` (CI: fresh database).
  - **Rebuild strategy**: the hub image embeds both the Go binary and the SPA build — any change to `upsilonhub/` or `upsilonbattleui/` requires rebuilding `hub` and `db-init` (same image) to maintain synchronization.

## BUILD AND EXECUTION PROCEDURE
- **Build strategy**:
  - **Context**: Must be executed from the **workspace root** to allow `upsilon*` cross-module resolution.
  - **Command**: `docker compose -f docker-compose.prod.yaml build`
- **Execution strategy**:
  - **Lifecycle**: Services must be started via `docker compose -f docker-compose.prod.yaml up -d --wait`.
  - **Order**: `db` healthy → `db-init` completes → `hub` starts → `proxy` healthcheck (`:8085/up`) gates dependents (handled via `depends_on`).
  - **Initialization**: migrations run automatically in `db-init`; no manual step.

## TECHNICAL INTERFACE
- **Files**:
  - `docker-compose.prod.yaml` (root)
  - `env.example` (root)
  - `scripts/setup_prod.sh`
  - `setup.md`
- **Code Tag**: `@spec-link [[infra_mvp_docker]]`

## EXPECTATION
- `docker compose up` starts all services (including db-init and cli).
- The stack is reachable at `http://localhost:8000` (prod front door).
- Data persists across stack restarts.
- Dashboard can retrieve active match stats from the Go engine via the synchronized internal network.
