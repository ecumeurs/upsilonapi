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
