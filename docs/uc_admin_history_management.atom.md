---
id: uc_admin_history_management
human_name: Administrative Match History Management Use Case
type: USER_STORY
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [admin, match-history]
parents:
  - [[shared:req_admin_experience]]
  - [[upsilonbattle:entity_game_match]]
dependents: []
---
# Administrative Match History Management Use Case

## INTENT
Provides administrators with the tools to review match results and purge old history to maintain system performance.

## THE RULE / LOGIC
- **Review Results:** Admin can list all completed matches and their outcomes.
- **History Purge:** 
  - Admin can perform a "clean" operation to remove match records older than a specific threshold (e.g., 90 days).
  - This operation is destructive and must be confirmed.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `GET /api/v1/admin/history` (paginated), `POST /api/v1/admin/history/purge`
- **Code Tag:** `@spec-link [[uc_admin_history_management]]`
- **Related Issue:** `#ISS-051`
- **Test Names:** `e2e_admin_history_management.js`

## EXPECTATION (For Testing)
- Admin requests match list -> results shown.
- Admin triggers history clean -> matches older than threshold are deleted from DB.
