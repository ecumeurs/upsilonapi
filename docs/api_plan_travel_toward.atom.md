---
id: api_plan_travel_toward
status: DRAFT
type: API
version: 1.0
tags: pathfinding,scripting
parents:
  - [[upsilonbattle:mech_move_validation]]
dependents: []
human_name: Plan Travel Toward API
layer: IMPLEMENTATION
priority: 3
---

# New Atom

## INTENT
Streamline entity movement planning toward a target coordinate by automatically handling occupancy and movement credit constraints.

## THE RULE / LOGIC
1. Find the acting entity in the current board state.
2. Determine if the target tile is occupied (by another unit or an obstacle).
3. If the target is occupied, find the shortest path to the most accessible adjacent tile.
4. If the target is empty, find the shortest path directly to it.
5. Truncate the resulting path by the entity's available movement points.
6. Return the plan as an array of coordinates.

## TECHNICAL INTERFACE
- **API Endpoint:** `upsilon.planTravelToward(entityId, targetPos, board)`
- **Code Tag:** `@spec-link [[api_plan_travel_toward]]`
- **Return Value:** `Array<{x: number, y: number}>`

## EXPECTATION
1. If the target is already adjacent and occupied, an empty path is returned.
2. If the entity has 0 move points, an empty path is returned.
3. If the target is 5 cells away but the entity has 3 move points, a path of 3 steps is returned.
4. If all adjacent cells to an occupied target are blocked, an empty path is returned.
