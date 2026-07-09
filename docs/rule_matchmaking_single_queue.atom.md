---
id: rule_matchmaking_single_queue
human_name: "Single Queue Restriction"
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [matchmaking, rule]
parents:
  - [[shared:uc_matchmaking]]
dependents:
  - [[upsilonbattleui:ui_dashboard_matchmaking]]
---

# Single Queue Restriction

## INTENT
To prevent a single player from occupying multiple queue slots or participating in multiple disparate matches simultaneously, ensuring fair resource allocation and consistent state.

## THE RULE / LOGIC
1. A Player (User) MUST NOT be permitted to join a matchmaking queue if they already have an active entry in ANY matchmaking queue.
2. A Player (User) MUST NOT be permitted to join a matchmaking queue if they are currently an active participant in an un-concluded match.
3. Attempts to join a queue while restricted MUST return a `409 Conflict` or equivalent error status with a descriptive message.
4. "Leave Queue" actions MUST clear all queue entries for the User across all game modes.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_matchmaking_single_queue]]`
- **Controller:** `MatchMakingController@joinMatch`
- **Validation:** Check `MatchmakingQueue` and `MatchParticipant` (active matches).

## EXPECTATION (For Testing)
- Join PvP -> Success.
- Join PvE while PvP is active -> Rejection (409).
- Leave Queue -> Success.
- Join PvE -> Success.
