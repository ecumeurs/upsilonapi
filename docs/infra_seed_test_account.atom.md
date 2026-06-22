---
id: infra_seed_test_account
human_name: "Test Account Seeding Requirement"
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 3
tags: [infra, seed, test]
parents:
  - [[upsilonbattle:entity_player]]
dependents: []
---

# Test Account Seeding Requirement

## INTENT
Ensures that a standard player account is available for manual testing and CI scenarios without manual registration.

## THE RULE / LOGIC
- **Account Details:**
  - `account_name`: `testuser`
  - `email`: `test@example.com`
  - `password`: `TestUserPassword123!`
  - `role`: `Player`
- **Initial State:**
  - Must have a standard roster of characters (3 characters).
  - Must have starting credits (1000).
- **Persistence:** Seeding must be idempotent.

## TECHNICAL INTERFACE (The Bridge)
- **Seeder Class:** `TestAccountSeeder.php`
- **Code Tag:** `@spec-link [[infra_seed_test_account]]`

## EXPECTATION (For Testing)
- Run seeder -> `testuser` exists with 3 characters.
