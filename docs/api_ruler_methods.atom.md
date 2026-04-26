---
id: api_ruler_methods
human_name: Ruler Message Methods API
type: API
layer: ARCHITECTURE
version: 1.2
status: STABLE
priority: 5
tags: [api, messaging, queue]
parents:
  - [[domain_ruler_state]]
dependents:
  - [[mech_controller_communication_sequence]]
  - [[rule_battle_readiness]]
  - [[rule_ruler_test_robustness]]
  - [[shared:rule_battle_readiness]]
  - [[shared:rule_ruler_test_robustness]]
  - [[api_controller_methods]]
  - [[upsilonbattle:mech_controller_communication_sequence]]
---
# Ruler Message Methods API

## INTENT
To define the explicit actor-message structures required to ingest data into the Ruler and extract state changes.

## THE RULE / LOGIC
Interaction with the backend engine is strictly channeled through `messagequeue` structs.

**State Commands (Read & Init):**
- `AddController`: Ingests a new player. Replies with `AddControllerReply` containing Grid, Entities, and TurnState.
- `GetGridState`: Requests board data (`GetGridStateReply`). Optional filtering via `AsController`.
- `GetEntitiesState`: Requests live roster data (`GetEntitiesStateReply`).

**Action Commands (Write):**
- `ControllerMove`: Issues a navigation path. Replies `ControllerMoveReply` containing the updated Entity state.
- `ControllerAttack`: Issues a basic attack against a target node. Replies `ControllerAttackReply`.
- `ControllerUseSkill`: Issues a complex skill against a target node. Replies `ControllerUseSkillReply`.
- `EndOfTurn`: Manually completes an entity's turn segment.
- `ControllerQuit`: Disconnects the controller from the session loop.

**Internal Engine Notifications:**
- `Timeout`: Internal notification triggered by the ShotClock to force end-of-turn processing. Must include `TurnIndex` for version validation.

**Broadcast Events (Engine to Clients):**
- `BattleStart`: Indicates the initial transition from setup to combat.
- `ControllerNextTurn`: Informs clients who just became the active entity.
- `EntitiesStateChanged`: Emits updated states after movement, damage, or healing.
- `ControllerSkillUsed` / `ControllerAttacked`: Specialized action notification for UX logging. **Must include `CreditAwards`** to notify participants of earned economy points.
- `BattleEnd`: Fires upon victory condition met; defines the winning `WinnerTeamID` and the legacy `WinnerControllerID`.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** Implicit RPC/Message Queue over `actor` communication channels (e.g., `github.com/ecumeurs/upsilontools/tools/messagequeue/message`).
- **Code Tag:** `@spec-link [[api_ruler_methods]]`
- **Related Issue:** `#None`
- **Test Names:** `N/A` (Interface def)

## EXPECTATION (For Testing)
- Submit `ControllerMove` msg -> Validated by Ruler -> Broadcasts `EntitiesStateChanged` -> Returns `ControllerMoveReply` to caller.
