---
id: uc_admin_login
human_name: "Admin Login Use Case"
type: USER_STORY
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [admin, login]
parents:
  - [[infra_seed_admin]]
  - [[rule_admin_access_restriction]]
  - [[shared:req_admin_experience]]
dependents:
  - [[battleui:ui_admin_dashboard]]
---

# Admin Login Use Case

## INTENT
To authenticate administrative users via a secure route and grant management privileges.

## THE RULE / LOGIC
1. Administrator provides administrative credentials via a dedicated secure endpoint.
2. System validates credentials and enforces [[rule_admin_access_restriction]].
3. Upon success, System issues a high-privilege JWT.
4. Administrator is redirected to the Admin Dashboard (UC-5/UC-6).

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `GET /admin/login` (Portal), `POST /login` (Handled via Role Redirect)
- **Code Tag:** `@spec-link [[uc_admin_login]]`
- **Related Issue:** `#admin-auth`

## EXPECTATION (For Testing)
- Unauthorized users are blocked from admin routes.
- Successful authentication grants admin-level permissions.
- Failed authentication returns a 401 Unauthorized error.
