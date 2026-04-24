---
id: infra_mvp_docker
human_name: MVP Docker Infrastructure
type: BUILD
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 5
tags: [docker, infrastructure, mvp]
parents:
  - [[module_backend]]
dependents:
  - [[watch_services]]
---
# MVP Docker Infrastructure

## INTENT
Provide a lightweight, development-friendly Docker orchestration for the Upsilon system MVP.

## THE RULE / LOGIC
- **Base Images**:
  - BattleUI: `php:8.4-apache` (Custom Dockerfile in ./battleui)
  - WebSocket: same as BattleUI, different command.
  - Go Engine: `golang:1.25-alpine` (Custom Dockerfile in ./upsilonapi)
  - Database: `postgres:18-alpine`
- **Service Orchestration**:
  - `app`: Laravel/Vue via Apache. Port `8000:80`.
  - `ws`: Reverb WebSocket server. Port `8080:8080`.
  - `engine`: Go battle engine. Port `8081:8081`.
  - `db`: PostgreSQL. Port `5434:5432` (Host:Container).
  - `cli`: On-stack maintenance/tester container.
- **Data Persistence**:
  - Named volume `db_data` for PostgreSQL `/var/lib/postgresql/data`. Ensures data survives restarts and shutdowns.
- **Environment**:
  - Managed via root `.env` file generated from `env.example`.
  - Secrets (APP_KEY, REVERB_*) are generated automatically.
- **Initialization**:
  - `db-init` service automates `php artisan migrate --force` on every stack startup using `www-data` permissions.
  - **Rebuild strategy**: Any change to `battleui/Dockerfile` dependencies (like permissions) requires rebuilding `app`, `ws`, and `db-init` to maintain synchronization.

## BUILD AND EXECUTION PROCEDURE
- **Build strategy**:
  - **Context**: Must be executed from the **workspace root** to allow `upsilon*` cross-module resolution.
  - **Command**: `docker compose -f docker-compose.prod.yaml build`
- **Execution strategy**:
  - **Lifecycle**: Services must be started via `docker compose -f docker-compose.prod.yaml up -d`.
  - **Order**: `db` must be healthy before `app`, `ws`, and `engine` can function (handled via `depends_on`).
  - **Initialization**: Database migrations must be run manually after the initial startup: `docker compose -f docker-compose.prod.yaml exec app php artisan migrate`.

## TECHNICAL INTERFACE
- **Files**:
  - `docker-compose.prod.yaml` (root)
  - `env.example` (root)
  - `scripts/setup_prod.sh`
  - `setup.md`
- **Code Tag**: `@spec-link [[infra_mvp_docker]]`

## EXPECTATION
- `docker compose up` starts all 6 services (including db-init and cli).
- `app` is reachable at `http://localhost:8000`.
- Data persists across stack restarts.
- Secrets are unique and consistent across services.
- Dashboard can retrieve active match stats from the Go engine via the synchronized internal network.
