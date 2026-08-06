---
id: domain_upsilon_engine
human_name: UpsilonBattle Core Engine Domain
type: DOMAIN
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: []
parents:
  - [[shared:req_tech_debt_backlog]]
dependents:
  - [[domain_credit_economy]]
  - [[domain_skill_system]]
  - [[rule_api_bridge_orchestration]]
  - [[upsilonbattle:mechanic_backstab_detection_algorithm]]
  - [[upsilonbattle:module_actor_concurrency]]
---
# UpsilonBattle Core Engine Domain

## INTENT
To define the UpsilonBattle Core Engine as the authoritative combat domain that the Arena API exposes to clients: it orchestrates client controllers, owns the battle lifecycle, evaluates combat from entity stats, and resolves match outcomes — keeping all combat state isolated from UI concerns.

## THE RULE / LOGIC
- **API Orchestration:** The engine accepts independent client Controllers connecting to an active Arena API; each controller drives one participant through the exposed battle-rule surface.
- **Combat Isolation:** The engine is strictly responsible for managing the battle lifecycle. UI-specific state handling is kept out of the engine boundary.
- **Entity & Stat Integration:** Combat is evaluated using entity properties (HP, Defense, Attack, and related stats).
- **Resolution:** A battle is complete when only one `TeamID` retains active, non-defeated entities; that team is declared the winner and `winner_team_id` is broadcast to all connected controllers and persisted as the final result. If all remaining entities are eliminated simultaneously (e.g. AoE self-damage), the match resolves as a DRAW (`WinnerTeamID: 0`). The full end-of-match flow (history persistence, CP award, progression transition) is specified in [[shared:uc_match_resolution]].

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[domain_upsilon_engine]]`
- **Boundary:** Arena API ⇄ engine; controllers interact only through the exposed battle-rule API.
