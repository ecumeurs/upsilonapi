---
id: mechanic_daily_random_shop_roll
status: DRAFT
human_name: Daily Deterministic Shop Roll
priority: 3
dependents: []
type: MECHANIC
layer: IMPLEMENTATION
parents:
  - [[mechanic_shop_inventory_system]]
version: 1.0
---

# Daily Deterministic Shop Roll

## INTENT
To provide a unique, non-exploitable daily rotating shop inventory for each player.

## THE RULE / LOGIC
- **Seed Generation:**
  Seed = Hash(user_id + account_creation_date + current_date_string)
- **Deterministic Rolling:**
  1. Initialize a PRNG with the generated Seed.
  2. Perform 3 standard rolls:
     - Roll Category (Weighted: Weapon 30%, Armor 30%, Utility 20%, Skill 20%)
     - Select random item from Category Registry using PRNG.
  3. Perform 1 bonus roll:
     - IF PRNG.Next(0, 100) < HighValueThreshold (e.g., 5%):
       - Roll 4th item from "High Value" pool.
- **Data Stability:**
  The `current_date_string` must be anchored to a specific timezone (UTC recommended) to ensure consistent rollover worldwide.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[mechanic_daily_random_shop_roll]]`
- **Related Issue:** `ISS-089`
- **Related Atoms:** `[[mechanic_shop_inventory_system]]`

## EXPECTATION
- Generating a shop inventory multiple times for the same user on the same day results in the exact same list of items.
- Generating a shop inventory for the same user on a different day results in a different list of items.
- Two different users on the same day get different inventories.
- The 4th item appears only when the PRNG roll is below the defined threshold.
- The item pool includes Weapons, Armor, Utilities, and Skills.
