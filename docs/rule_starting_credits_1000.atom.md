---
id: rule_starting_credits_1000
status: STABLE
priority: 5
parents:
  - [[domain_credit_economy]]
type: RULE
version: 2.0
tags: [credits, onboarding, economy, iss-074]
human_name: Starting Credit Balance Rule
dependents: []
layer: BUSINESS
---

# Starting Credit Balance Rule

## INTENT
To define the starting credit balance granted to every new user upon registration. V2 design decision: 1000 credits — enough to acquire one V2.0 item (300 max) and retain meaningful runway for additional purchases as wins generate further credits.

## THE RULE / LOGIC
- **Starting balance:** 1000 credits, granted at registration time and persisted on `users.credits`.
- **Database mechanism:** `users.credits` column has `default(1000)`. Existing zero-balance users are backfilled idempotently via migration (`update users set credits = 1000 where credits = 0`).
- **Rationale:**
  - V2.0 catalog totals: Basic Armor (200) + Basic Sword (300) + Swift Boots (150) = 650 credits. 1000 leaves budget for one full equipment loadout plus reserve.
  - Avoids the dead-end onboarding state where a fresh account cannot afford any item.
  - Permanent V2 design decision (not a testing flag) — captured here and referenced by the registration flow.
- **Crash early:** registration must persist exactly 1000; any deviation is a defect.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_starting_credits_1000]]`
- **Migration:** `*_default_credits_1000.php`
- **Models:** `App\Models\User` (registration path)
- **Test Names:** `TestRegistration_StartsWith1000Credits`

## EXPECTATION
- A freshly registered user has `users.credits == 1000` immediately after `POST /v1/auth/register`.
- An existing user with `credits=0` at migration time is updated to 1000.
- An existing user with `credits>0` is not affected by the backfill.
