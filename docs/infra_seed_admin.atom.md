---
id: infra_seed_admin
human_name: Administrator Account Seeding Requirement
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [infra, seed, admin]
parents:
  - [[entity_player]]
dependents: []
---
# Administrator Account Seeding Requirement

## INTENT
Ensures that a default administrator account is available upon system deployment.

## THE RULE / LOGIC
- **Account Details:**
  - `account_name`: `admin`
  - `email`: `admin@admin.com`
  - `role`: `Admin`
- **Security:**
  - The password MUST NOT be hardcoded. 
  - It must be provided via an environment variable at deployment/seed time (e.g., `ADMIN_INITIAL_PASSWORD`).
- **Persistence:** Seeding must be idempotent; it should not overwrite an existing admin or fail if already seeded.

## TECHNICAL INTERFACE (The Bridge)
- **Seeder Class:** `DatabaseSeeder.php`
- **ENV Context:** `ADMIN_INITIAL_PASSWORD`
- **Code Tag:** `@spec-link [[infra_seed_admin]]`
- **Test Names:** `TestAdminAccountSeededCorrectly`

## EXPECTATION (For Testing)
- Run seeder with ENV variable -> admin user exists in DB with role Admin.
- Env variable missing -> Seeding fails or logs warning (must not use default password).
