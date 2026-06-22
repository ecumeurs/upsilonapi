---
id: uc_admin_user_management
human_name: Administrative User Management Use Case
type: USER_STORY
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [admin, user-management]
parents:
  - [[shared:req_admin_experience]]
  - [[upsilonbattle:entity_player]]
dependents: []
---
# Administrative User Management Use Case

## INTENT
Allows administrators to perform system maintenance tasks on user accounts, specifically soft deletions.

## THE RULE / LOGIC
- **List Users:** Admin can retrieve a list of all accounts (excluding private data).
- **Soft Delete:** 
  - Admin can mark an account as deleted.
  - This action must follow [[rule_gdpr_compliance]] for soft deletion and anonymization.
  - Admin cannot reverse a "Right to be Forgotten" anonymization once finalized.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoints:**
  - `GET /admin/users` (List)
  - `POST /admin/users/{user:account_name}/anonymize` (GDPR)
  - `DELETE /admin/users/{user:account_name}` (Soft Delete)
- **Code Tag:** `@spec-link [[uc_admin_user_management]]`
- **Test Names:** `TestAdminUserSoftDelete`, `TestAdminUserAnonymization`

## EXPECTATION (For Testing)
- Admin selects "Delete" on user -> User `deleted_at` is set -> Data is anonymized -> Admin confirms user is no longer active.
