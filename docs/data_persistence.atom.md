---
id: data_persistence
human_name: PostgreSQL Database Persistence
type: DATA
layer: IMPLEMENTATION
version: 1.0
status: STABLE
priority: 5
tags: [database, postgresql, state]
parents: []
dependents:
  - [[entity_game_match]]
  - [[entity_users]]
---
# PostgreSQL Database Persistence

## INTENT
Serve as the centralized, persistent source of truth for accounts, characters, and historical match statistics.

## THE RULE / LOGIC
- Technology Stack: Must be strictly deployed on PostgreSQL.
- Primary Game Logic Entities:
  - [[entity_users]] (authentication credentials, role-based access, win/loss metrics, WebSocket channel keys).
  - [[entity_character]] (HP, Movement, Attack, Defense stats linked to a User via player_id, includes initial_movement for progression tracking).
  - [[entity_game_match]] (matches historical state, game_state_cache, grid_cache, turn tracking, version for deduplication).
  - [[entity_match_participants]] (junction table linking users/AI to matches, supports nullable player_id for PvE modes).
  - Matchmaking Queues (active queues with JSON-based character selection, game_mode specification).
- Infrastructure Tables (Laravel-specific, not game logic):
  - Session management (sessions, cache, cache_locks)
  - Job/Queue processing (jobs, job_batches, failed_jobs)
  - Authentication utilities (password_reset_tokens, personal_access_tokens)
- Integration Note: Since Laravel orchestrates authentication and Go orchestrates active combat, both services share the PostgreSQL instance with clear responsibility boundaries.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[data_persistence]]`
- **Test Names:** `TestPostgresPlayerSchema`, `TestPostgresCharacterSchema`

## EXPECTATION (For Testing)
- Game Ends via Go API -> Service updates Player Win/Loss record in PostgreSQL -> Laravel queries updated stats for the Leaderboard.
