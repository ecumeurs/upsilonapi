---
id: rule_admin_access_restriction
human_name: Administrator Access Restriction Rule
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [security, admin, privacy]
parents:
  - [[shared:req_security]]
  - [[upsilonbattle:entity_player]]
dependents:
  - [[uc_admin_login]]
---
# Administrator Access Restriction Rule

## INTENT
Ensures that while administrators have system management capabilities, they are strictly barred from viewing sensitive personal data of users.

## THE RULE / LOGIC
- **Restricted Fields:** Administrators MUST NOT have access to the following fields in `entity_player`:
  - `full_address`
  - `birth_date`
- **Enforcement:** Dashboard and API responses for administrators must censor or omit these fields, even when managing user accounts.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_admin_access_restriction]]`
- **Test Names:** `TestAdminCantSeeAddress`, `TestAdminCantSeeBirthDate`

## EXPECTATION (For Testing)
- Admin requests user profile -> `full_address` and `birth_date` are null or omitted.
- Database query from admin-privileged service layer -> fields are masked.
