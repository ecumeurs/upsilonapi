---
id: api_shop_purchase
status: STABLE
layer: ARCHITECTURE
version: 2.0
human_name: Shop Purchase API
type: API
priority: 5
tags: [api, shop, purchase, iss-074]
parents:
  - [[upsilonbattle:entity_player_credits]]
  - [[upsilonbattle:entity_player_inventory]]
  - [[upsilonbattle:mec_credit_spending_shop]]
dependents:
  - [[battleui:ui_shop]]
---

# New Atom

## INTENT
To deduct credits from a user's balance and add the purchased item to their inventory in a single transactional step. The only state-changing shop endpoint.

## THE RULE / LOGIC
- **Endpoint:** `POST /api/v1/shop/purchase`
- **Auth:** Sanctum.
- **Body:** `{ "shop_item_id": "<uuid>", "quantity": <int, optional, default 1> }`.
- **Service:** `ShopService::purchase` runs in a DB transaction:
  1. Lock user row (`SELECT ... FOR UPDATE`).
  2. Resolve `ShopItem`; if `available=false` or not found → 404.
  3. Compute `total_cost = cost × quantity`.
  4. If `users.credits < total_cost` → 422 with `meta.reason='insufficient_credits'`.
  5. Compute `new_quantity = existing_quantity + quantity`. If `new_quantity > 99` → 422 with `meta.reason='quantity_cap'` per `[[rule_quantity_cap]]`.
  6. Debit `users.credits`.
  7. Upsert `player_inventory` row.
  8. Insert `inventory_transactions` (transaction_type=`purchase`).
  9. Insert `credit_transactions` (source=`shop_purchase`).
- **Response (200):** `{ credits: <new_balance>, inventory_item: InventoryItemResource }`.
- **Failure modes:** 400 (validation), 401 (unauth), 404 (unknown item), 422 (insufficient_credits | quantity_cap).
- **Crash early:** No silent fallback. Any exception inside the service rolls the transaction back.

## TECHNICAL INTERFACE
- **API Endpoint:** `POST /api/v1/shop/purchase`
- **Controller:** `ShopController@purchase`
- **Service:** `App\Services\ShopService::purchase`
- **Code Tag:** `@spec-link [[api_shop_purchase]]`
- **Request:** `App\Http\Requests\PurchaseShopItemRequest`

## EXPECTATION
- Successful purchase returns 200, decrements credits, increments inventory.
- Insufficient credits returns 422 with `meta.reason='insufficient_credits'`; balance and inventory unchanged.
- Quantity cap returns 422; balance and inventory unchanged.
- Unknown shop_item_id returns 404.
- Concurrent purchases on the same user serialize via `FOR UPDATE`.
