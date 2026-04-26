---
id: api_inventory_list
status: DRAFT
tags: [api, inventory, iss-074]
parents:
  - [[entity_player_inventory]]
  - [[mec_shop_inventory_system]]
type: API
version: 2.0
human_name: Player Inventory List API
dependents:
  - [[battleui:ui_inventory]]
  - [[ui_inventory]]
layer: ARCHITECTURE
priority: 5
---

# New Atom

## INTENT
To expose the authenticated user's owned items, with each row annotated by its current equip status (which character, if any, has it bound). Read-only.

## THE RULE / LOGIC
- **Endpoint:** `GET /api/v1/profile/inventory`
- **Auth:** Sanctum. Scope: own inventory only — cross-user reads return 403.
- **Response:** Standard envelope with `data: InventoryItemResource[]`.
- **`InventoryItemResource` shape:** `{ id, shop_item: ShopItemResource, quantity, purchased_at, equipped_on: { character_id, character_name, slot } | null }`.
- **Equip annotation:** LEFT JOIN `character_equipment` across user's characters; if any of `(armor_item_id, utility_item_id, weapon_item_id)` references this inventory row, populate `equipped_on`. Otherwise `null`.
- **Filtering / sorting:** None in V2.0; client-side tabs handle the slot-category filter.

## TECHNICAL INTERFACE
- **API Endpoint:** `GET /api/v1/profile/inventory`
- **Controller:** `InventoryController@index`
- **Code Tag:** `@spec-link [[api_inventory_list]]`
- **Resource:** `App\Http\Resources\InventoryItemResource`

## EXPECTATION
- Authenticated GET returns 200 with array of owned items.
- Items currently equipped have populated `equipped_on`; unequipped items have `equipped_on=null`.
- Cross-user data never leaks (scoped by `auth()->id()`).
