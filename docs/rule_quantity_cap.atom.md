---
id: rule_quantity_cap
status: STABLE
version: 2.0
priority: 5
human_name: Inventory Quantity Cap Rule
dependents: []
tags: [inventory, validation, iss-074]
parents:
  - [[entity_player_inventory]]
type: RULE
layer: ARCHITECTURE
---

# Inventory Quantity Cap Rule

## INTENT
To bound `player_inventory.quantity` to a maximum of 99 units per item per user, preventing trivial credit-burn exploits and bounding inventory cardinality for UI rendering.

## THE RULE / LOGIC
- **Cap:** `player_inventory.quantity <= 99`.
- **Enforcement:** Service-layer validation in `ShopService::purchase`. A purchase that would push quantity over 99 is rejected with HTTP 422 and `meta.reason = "quantity_cap"`.
- **No DB constraint:** the cap is enforced in code, not at the column level, so the rule can be relaxed in V2.1 (stacking polish) without a destructive schema change.
- **No partial fulfillment:** if requesting `quantity=10` but only 5 would fit, the entire purchase is rejected (no silent truncation). Crash early.
- **V2.0 UX implication:** since the catalog is 3 items and most uses are 1× per item, the cap is rarely hit in practice; it's a guardrail.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[upsilonapi:rule_quantity_cap]]` — `upsilonhub/internal/platform/economy/pg_inventory.go` (`Purchase`, the `newQuantity > QuantityCap` guard).
- **Service:** `Purchase` (Go port of `ShopService::purchase`).
- **Test Names:** `TestPurchaseRejectsQuantityCap` (`upsilonhub/internal/gateway/shop_test.go`); CLI edge coverage in `edge_quantity_cap_99.js`.

## EXPECTATION
- Purchasing the 100th unit of any item returns 422.
- Inventory and credit balance unchanged on rejection.
- Mixed-quantity purchases that would partially exceed the cap are wholly rejected.
