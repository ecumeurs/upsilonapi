---
id: api_equipment_management
human_name: Equipment Management API
type: API
layer: ARCHITECTURE
version: 2.1
status: DRAFT
priority: 5
tags: [api, equipment, inventory]
parents:
  - [[upsilonbattle:entity_equipment_system]]
  - [[upsilonbattle:mec_three_slot_equipment_system]]
dependents:
  - [[battleui:ui_character_equipment_panel]]
  - [[battleui:ui_inventory]]
  - [[ui_character_equipment_panel]]
  - [[ui_inventory]]
---

# Equipment Management API

## INTENT
To provide API endpoints for character equipment management: viewing the 3-slot equipment configuration, equipping an inventory item (slot inferred from item type), and unequipping by slot. Equip / unequip are the only state-changing endpoints; purchase lives in `api_shop_purchase`.

## THE RULE / LOGIC
**Equipment Management Endpoints:**

**View Character Equipment:**
- **Endpoint:** `GET /api/v1/profile/character/{id}/equipment`
- **Auth:** Sanctum + ownership policy (user must own the character).
- **Returns:** The 3-slot configuration (armor, utility, weapon). Each slot is either `null` or an `InventoryItemResource` with the underlying `ShopItemResource` and effective property contributions.

**Equip Item (single endpoint, slot inferred):**
- **Endpoint:** `POST /api/v1/profile/character/{id}/equip`
- **Body:** `{ "item_id": "<player_inventory_uuid>" }`
- **Slot Resolution:** Looked up from `shop_items.slot` of the underlying catalog row. The client MUST NOT pass a slot — the server is the authority.
- **Validation:**
  - User owns the character (Policy `equip`).
  - User owns the inventory row.
  - If another character of the same user has the item equipped, the previous binding is cleared atomically (cross-character mutual exclusivity in a single DB transaction).
  - If the slot already holds another item, that item is unbound and returned to inventory.
- **Response:** Updated `CharacterEquipmentResource` plus the recomputed effective stats.

**Unequip Item:**
- **Endpoint:** `DELETE /api/v1/profile/character/{id}/unequip/{slot}` where `slot ∈ {armor, utility, weapon}`.
- **Auth:** Sanctum + ownership policy.
- **Behavior:** If the slot is empty, returns 404. Otherwise clears the slot and returns the updated `CharacterEquipmentResource`.

**Stat Recalculation:**
- Equipment changes do not mutate base character columns. The engine resolves equipment contributions at battle init via Forever buffs (see `mech_item_buff_application`). The dashboard composes effective stats client-side via the equipment summary in `CharacterResource`.

**Inventory & Purchase (separate atoms):**
- Inventory list: see `api_inventory_list` (`GET /v1/profile/inventory`).
- Shop browse / purchase: see `api_shop_browse`, `api_shop_purchase`.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[api_equipment_management]]`
- **Controller:** `EquipmentController` (Laravel)
- **Models:** `App\Models\Character`, `App\Models\CharacterEquipment`, `App\Models\PlayerInventory`, `App\Models\ShopItem`
- **Policy:** `CharacterPolicy::equip`, `CharacterPolicy::unequip`
