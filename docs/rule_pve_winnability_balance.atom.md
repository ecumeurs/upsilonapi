---
id: rule_pve_winnability_balance
human_name: "PVE Winnability Balance"
type: RULE
layer: BUSINESS
version: 1.0
status: STABLE
priority: 4
tags: [pve, balance, gameplay]
parents:
  - [[battleui:req_ui_look_and_feel]]
dependents: []
---

# PVE Winnability Balance

## INTENT
To ensure that any PVE engagement is technically winnable by preventing AI defense stats from making them invulnerable to the player's strongest attack.

## THE RULE / LOGIC
- When generating AI entities for a PVE match:
  - **The "Glass Ceiling" Rule:** No AI entity may have a `defense` stat that is greater than or equal to the `highest attack` value among all player-controlled entities in the same match.
  - `AI_Defense < Max(Player_Attacks)`
- If a generated AI entity violates this rule, its `defense` must be capped at `Max(Player_Attacks) - 1`.

## TECHNICAL INTERFACE (The Bridge)
- **Logic Location:** `MatchMakingController` during AI entity roll.
- **Code Tag:** `@spec-link [[rule_pve_winnability_balance]]`
- **Input:** Player team entity roster.

## EXPECTATION (For Testing)
- In a 1v1 PVE match, if the player's strongest unit has 10 Attack, no AI unit can have 10 or more Defense.
- Compliance is checked at the moment of Match Creation.
