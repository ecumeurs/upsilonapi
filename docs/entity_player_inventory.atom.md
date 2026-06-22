---
id: entity_player_inventory
status: DRAFT
type: ENTITY
layer: ARCHITECTURE
tags: [inventory, ownership, items, iss-074]
human_name: Player Inventory Entity
version: 2.0
priority: 5
parents:
  - [[mechanic_shop_inventory_system]]
  - [[upsilontypes:entity_shop_item]]
dependents:
  - [[api_inventory_list]]
  - [[api_shop_purchase]]
  - [[entity_character_equipment]]
  - [[rule_quantity_cap]]
---

# Player Inventory Entity

## INTENT
To define the per-user inventory entity — the set of items a user has purchased. Inventory tracks ownership only; equip status is held in `[[entity_character_equipment]]`. This is a deliberate split (decision D1 of ISS-074): inventory rows do not encode "which character has this equipped".

## THE RULE / LOGIC
**Player Inventory Entity:**

**Core Fields (`player_inventory` table):**
- **id:** UUID primary key.
- **player_id:** UUID FK → `users.id` ON DELETE CASCADE. Inventory dies with the user.
- **shop_item_id:** UUID FK → `shop_items.id` ON DELETE CASCADE. Catalog reference (D13).
- **quantity:** Integer default 1, hard-capped at 99 per `[[rule_quantity_cap]]`.
- **purchased_at:** Timestamp default `now()`.

**Constraints:**
- Unique `(player_id, shop_item_id)` — duplicate purchases increment `quantity` rather than creating new rows.
- Quantity > 0 always; rows with quantity=0 are deleted (V2.1+ once stacking UI exists).
- No `character_id` column. Per D1 of ISS-074, equip state lives in `character_equipment`.

**Audit Trail:**
- Every insert / quantity-increment is mirrored in `inventory_transactions` (see `[[mechanic_shop_inventory_system]]`).
- Every credit-debiting purchase is also mirrored in `credit_transactions` with `source='shop_purchase'`.

**Privacy:**
- Inventory is visible only to its owner (Sanctum + ownership check in controller).

**Lifecycle:**
- Created by `[[api_shop_purchase]]`.
- Read by `[[api_inventory_list]]` and joined into `[[api_equipment_management]]` for the equip flow.
- Deleted via cascade (user deletion OR shop item deletion).

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[entity_player_inventory]]`
- **Laravel Model:** `App\Models\PlayerInventory`
- **Migration:** `*_create_item_system_tables.php`
- **Resource:** `App\Http\Resources\InventoryItemResource`

## EXPECTATION
- A purchase increments `quantity` rather than inserting a duplicate row.
- Cascade delete on `users` removes all inventory rows.
- `quantity` cannot exceed 99 (rejected at service layer with 422).
- Cross-user reads return 403.
