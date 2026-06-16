---
id: arch_api_id_masking_gateway
human_name: "Architectural API ID Masking Gateway"
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [security, api, masking, uuid]
parents:
  - [[shared:requirement_customer_user_id_privacy]]
dependents: []
---

# Architectural API ID Masking Gateway

## INTENT
To provide a secure translation layer between internal database identifiers (UUIDs) and public-facing semantic or masked identifiers, preventing reconnaissance and primary key enumeration.

## THE RULE / LOGIC
- **Internal vs Public Boundary:** All raw database UUIDs (User, Character) MUST be intercepted at the API Gateway (Laravel) before reaching the network.
- **Masking Mechanisms:**
  - **Boolean Flags:** Replace User IDs with `is_self: boolean` (e.g., in entities) or `current_player_is_self` (for turn state).
  - **Pseudonyms:** Use persistent, non-traceable keys for long-term identification (e.g., `ws_channel_key`).
  - **Team Identifiers:** Expose `winner_team_id` for reporting match outcomes without exposing the winning player's personal ID.
- **Inbound Validation (Ownership):** 
  - For every state-changing request (Actions, Upgrades), the Gateway MUST verify that the authenticated User owns the targeted Entity (Character/Match Participant) before proxying to the Battle Engine.
- **Match Scoping:** Match IDs are permissible in URLs but MUST be guarded by participant-level authorization.

## TECHNICAL INTERFACE (The Bridge)
- **Laravel Resources:** Use `toArray()` to filter out `id` and inject `is_self`.
- **Middleware/Policies:** `CharacterPolicy` and `MatchParticipantPolicy` for ownership enforcement.
- **Code Tag:** `@spec-link [[arch_api_id_masking_gateway]]`

## EXPECTATION (For Testing)
- `GET /api/v1/leaderboard` -> No `id` field present; `is_self` correctly identifies the caller.
- `GET /api/v1/game/{id}` -> `current_player_is_self` and `game_finished` provide state without UUID exposure.
- `winner_team_id` is exposed for team-level match resolution in logs (Unified from `winning_team_id`).
- `POST /api/v1/game/{id}/action` with an `entity_id` not owned by the user -> `403 Forbidden`.

## TRUST BOUNDARY: upsilonapi (port 8081) is INTERNAL-ONLY

**Decision recorded 2026-06-16** (ISS-098 re-scoping).

`upsilonapi` (Go battle engine bridge, port 8081) is a **trusted internal service** — it is not a public-facing API. Raw internal User UUIDs appearing as `player_id` in its `BoardState` DTOs are **accepted by design** on the internal hop.

### Masking gateway: battleui is the sole external surface

All external clients (web browser, WebSocket subscribers) receive board state exclusively through battleui (Laravel), which applies masking unconditionally:

- **`app/Http/Resources/BoardStateResource.php`** — `unset`s `player_id` and `current_player_id` from the DTO before serialisation.
- **`app/Events/BoardUpdated.php`** — broadcasts per-recipient state via the Resource above; raw IDs never reach the WebSocket channel.
- **`app/Http/Controllers/GameController.php`** — serves HTTP polling responses through the same Resource.

### Network isolation evidence
- `docker-compose.ci.yaml`: `engine` service has **no `ports:` mapping** — 8081 is reachable only within the compose network.
- `docker-compose.prod.yaml`: `engine` publishes `8081:8081` to the EC2 host, but the AWS EC2 security group (created by `upsilonaws/scripts/setup/01_networking.sh`) authorises **only ports 22, 80, and 443** from the public internet. Port 8081 is not in the SG ingress rules.
- The nginx reverse proxy on EC2 routes only `APP_PORT` (8000) and `WS_PORT` (8080); 8081 is not proxied.
- **Latent risk**: the `ports: "8081:8081"` line in `docker-compose.prod.yaml` means 8081 is bound on the EC2 host loopback/interface. If an EC2 SG rule for 8081 were ever added, it would become publicly reachable. This should be remediated by removing the host-port binding (use internal Docker networking only). Tracked as a Low-severity hardening item in ISS-098.

### Consequence for ISS-098
Because the web path is fully masked at battleui and 8081 is not publicly reachable via the current SG, ISS-098 is downgraded from High to Low severity. The recommended long-term fix remains: remove `"8081:8081"` from `docker-compose.prod.yaml` and ensure CI/prod never publish the engine port.
