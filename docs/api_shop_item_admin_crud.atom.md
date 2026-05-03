---
id: api_shop_item_admin_crud
status: STABLE
parents:
  - [[upsilontypes:entity_shop_item]]
  - [[rule_admin_content_authority]]
dependents: []
type: API
layer: ARCHITECTURE
version: 2.0
priority: 5
tags: [api, shop, admin, crud, iss-086]
human_name: Admin Shop Item CRUD API
---

# New Atom

## INTENT
To give admins full CRUD control over the shop item catalog — enabling availability toggling, pricing changes, and addition of exotic items without code deploys.

## THE RULE / LOGIC
**Admin-only endpoints (Sanctum + `admin` middleware):**

- `GET    /api/v1/admin/shop-items` — list all items including unavailable ones.
- `GET    /api/v1/admin/shop-items/{id}` — get single item (404 if not found).
- `POST   /api/v1/admin/shop-items` — create item. Required: `name`, `slot` (in:armor,utility,weapon), `cost` (int). Optional: `properties` (JSON map), `type` (tag string), `available` (bool), `skill_template_id` (UUID FK — D11 exotic items).
- `PUT    /api/v1/admin/shop-items/{id}` — full update.
- `DELETE /api/v1/admin/shop-items/{id}` — delete item; returns 404 if not found.

**D11 exotic items:** When `skill_template_id` is set, the bridge appends the linked skill template's snapshot to the entity's `EquippedSkills` at arena initialization with `origin='item:{inventory_item_id}'`.

## TECHNICAL INTERFACE
- **Controller:** `App\Http\Controllers\API\Admin\AdminShopItemController`
- **Form Requests:** `StoreShopItemRequest`, `UpdateShopItemRequest`
- **Resource:** `App\Http\Resources\ShopItemResource`
- **Code Tag:** `@spec-link [[api_shop_item_admin_crud]]`

## EXPECTATION
- Non-admin requests return 403.
- Store returns 201 with created item.
- Update returns 200 with refreshed item.
- Delete returns 200 with success message.
- GET/PUT/DELETE on nonexistent ID returns 404.
- Player `GET /shop/items` excludes items where `available=false`.
