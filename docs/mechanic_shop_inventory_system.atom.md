---
id: mechanic_shop_inventory_system
human_name: Shop & Inventory System
type: MECHANIC
layer: IMPLEMENTATION
version: 2.1
status: DRAFT
priority: 5
tags: [economy, shop, inventory, equipment]
parents:
  - [[domain_credit_economy]]
dependents:
  - [[api_inventory_list]]
  - [[api_shop_browse]]
  - [[api_shop_purchase]]
  - [[entity_player_inventory]]
  - [[mechanic_daily_random_shop_roll]]
  - [[upsilontypes:entity_shop_item]]
---

# Shop & Inventory System

## INTENT
To implement the shop & inventory system where players spend credits to browse and purchase skills and equipment — covering the catalog (browse + filtering), affordability, DB-transactional purchase with full audit trail, and ownership/inventory management. Prices in V2.0 are fixed (seeded), with procedural pricing (Skill Weight / equipment tier) deferred.

## THE RULE / LOGIC
**V2.0 Fixed Pricing (ISS-074):**
- Per `[[upsilontypes:rule_item_pricing_simple]]`, V2.0 catalog entries carry an explicit `cost` column populated at seed time. No procedural pricing in V2.0.
- Reference catalog (V2.0):
  - Basic Armor — 200 credits — slot=armor — properties={ArmorRating:5}
  - Basic Sword — 300 credits — slot=weapon — properties={WeaponBaseDamage:5, WeaponType:"One-Handed Melee", WeaponRange:1}
  - Swift Boots — 150 credits — slot=utility — properties={Movement:1}

**V2.1+ Procedural Pricing (deferred):**
- Skill cost = Total Positive Skill Weight × 2 credits.
- Equipment cost = Equipment Power Rating × Tier Multiplier, where Power Rating is the sum of weighted property values (Damage weighted higher than HP) and Tier Multiplier ranges Common 1.0× → Legendary 5.0×.
- Reintroduced when the procedural item generator lands.

**Catalog & Inventory Browse:**
- **Categories:** Skills (organized by I–V grade) and Equipment (Armor / Utility / Weapon).
- **Equipment classification:** Armor = defensive slots (Armor Rating / Defense); Utility = stat enhancements, MP/SP pools, unique effects; Weapons = one/two-handed melee or ranged (attack power + range).
- **Tier system:** Common → Uncommon → Rare → Epic → Legendary, with increasing stat bonuses and specialized properties.
- **Visibility / prerequisites:** items shown only when the character meets the minimum level and any required precursor skills.
- **Filtering & sorting:** by category, grade, stat focus, price range, tier, or affordability; affordable items highlighted against the current credit balance. Text search by name/property.
- **Comparison:** side-by-side equipped-vs-candidate view projecting stat changes before purchase.

**Purchase Mechanics (V2.0):**
- Endpoint: `[[api_shop_purchase]]` — `POST /v1/shop/purchase` body `{ shop_item_id, quantity? }`.
- Service (`ShopService::purchase`) is DB-transactional:
  1. Lock user row, check `users.credits >= shop_items.cost × quantity`.
  2. Insufficient → 422 with `meta.reason = "insufficient_credits"`.
  3. Debit credits.
  4. Upsert `player_inventory` row (increment quantity if exists, capped at 99 per `[[rule_quantity_cap]]`).
  5. Insert `inventory_transactions` audit row (transaction_type=`purchase`).
  6. Insert `credit_transactions` audit row (source=`shop_purchase`).
- **Crash early:** any service-level failure rolls back the whole transaction.
- Purchased items land in the character's unequipped inventory; binding to one of the three active slots (Armor / Utility / Weapon) and the resulting stat recalculation are governed by `[[api_equipment_management]]` and `[[upsilonbattle:mechanic_three_slot_equipment_system]]`.

**Shop Browse (V2.0):**
- Endpoint: `[[api_shop_browse]]` — `GET /v1/shop/items` returns all `available=true` rows. No filtering / pagination in V2.0 (3-item catalog). The daily-rotation variant is `[[mechanic_daily_random_shop_roll]]`.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[mechanic_shop_inventory_system]]`
- **API Endpoints:** `[[api_shop_browse]]`, `[[api_shop_purchase]]`
- **Service:** `App\Services\ShopService::purchase` (Laravel)
- **Migration:** `*_create_item_system_tables.php` (creates `shop_items`, `player_inventory`, `inventory_transactions`)
- **Related:** `[[entity_player_inventory]]`, `[[upsilontypes:entity_shop_item]]`, `[[api_equipment_management]]`, `[[mechanic_daily_random_shop_roll]]`
