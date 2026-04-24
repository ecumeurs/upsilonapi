---
id: domain_upsilon_engine_domain_upsilon_engine_resolution
human_name: UpsilonBattle Core Engine Resolution Domain
type: DOMAIN
layer: CUSTOMER
version: 1.0
status: STABLE
priority: 5
tags: []
parents:
  - [[domain_upsilon_engine]]
dependents: []
---
# UpsilonBattle Core Engine Resolution Domain

## INTENT
To formally define the conditions under which a battle arena is concluded and a winning team is declared.

## THE RULE / LOGIC
- **Completion Condition:** A battle is considered complete when only one `TeamID` has active, non-defeated entities remaining on the grid.
- **Victory Declaration:** The last remaining `TeamID` is declared the winner.
- **Reporting:** Upon conclusion, the `winner_team_id` MUST be broadcast to all connected controllers and stored as the final result of the match.
- **Draws:** If all remaining entities are eliminated simultaneously (e.g., area-of-effect self-damage), the match is resolved as a DRAW (WinnerTeamID: 0).

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[domain_upsilon_engine_domain_upsilon_engine_resolution]]`
