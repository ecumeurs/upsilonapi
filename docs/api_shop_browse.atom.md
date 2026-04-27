---
id: api_shop_browse
status: STABLE
layer: ARCHITECTURE
version: 2.0
priority: 5
human_name: Shop Catalog Browse API
type: API
tags: [api, shop, catalog, iss-074]
parents:
  - [[entity_shop_item]]
  - [[upsilonbattle:mec_credit_spending_shop]]
dependents: []
---

# New Atom

## INTENT
To expose the shop catalog (all `available=true` rows of `shop_items`) so authenticated users can browse what's purchasable. Read-only.

## THE RULE / LOGIC
- **Endpoint:** `GET /api/v1/shop/items`
- **Auth:** Sanctum (any authenticated user).
- **Response:** Standard envelope with `data: ShopItemResource[]`.
- **`ShopItemResource` shape:** `{ id, name, type, slot, properties, cost, available }`.
- **Filtering:** None in V2.0 (3-item catalog). Server returns all `available=true` rows.
- **Pagination:** None in V2.0.
- **Caching:** Catalog is effectively static between deploys; controller may cache the query result for 60s.

## TECHNICAL INTERFACE
- **API Endpoint:** `GET /api/v1/shop/items`
- **Controller:** `ShopController@index`
- **Code Tag:** `@spec-link [[api_shop_browse]]`
- **Resource:** `App\Http\Resources\ShopItemResource`

## EXPECTATION
- Authenticated GET returns 200 with array of 3 items in V2.0.
- Unauthenticated GET returns 401.
- Items with `available=false` are excluded.
- Response envelope conforms to `[[shared:api_standard_envelope]]`.
