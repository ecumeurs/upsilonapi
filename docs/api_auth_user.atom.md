---
id: api_auth_user
status: STABLE
type: API
layer: ARCHITECTURE
tags: [api, auth, account, gdpr]
parents:
  - [[shared:requirement_customer_user_account]]
dependents:
  - [[mechanic_mech_cli_sensitive_data_masking]]
  - [[upsilonbattle:mechanic_mech_cli_sensitive_data_masking]]
human_name: User Authentication & Account API
priority: 5
version: 1.0
---

# New Atom

## INTENT
To provide a centralized technical specification for all user authentication, identity synchronization, and account lifecycle endpoints.

## THE RULE / LOGIC
- **Registration/Login:** Standard JWT exchange via Laravel Sanctum.
- **Account Update:** Partial updates to user profile (nickname, email, birth date, address).
- **Security:** Password rotation requires current password verification.
- **GDPR compliance:** Soft delete with anonymization on `DELETE /auth/delete`. Data portability JSON on `GET /auth/export`.

## TECHNICAL INTERFACE
- **Controller:** `AuthController`
- **Code Tag:** `@spec-link [[api_auth_user]]`
- **Endpoints:**
    - `POST /api/v1/auth/login`
    - `POST /api/v1/auth/register`
    - `POST /api/v1/auth/logout`
    - `POST /api/v1/auth/update`
    - `POST /api/v1/auth/password`
    - `GET /api/v1/auth/export`
    - `DELETE /api/v1/auth/delete`

## EXPECTATION
