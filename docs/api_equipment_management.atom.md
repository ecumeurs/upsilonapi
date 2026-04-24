---
id: api_equipment_management
human_name: Equipment Management API
type: API
layer: ARCHITECTURE
version: 2.0
status: DRAFT
priority: 5
tags: [api, equipment, inventory]
parents:
  - [[entity_equipment_system]]
  - [[mec_three_slot_equipment_system]]
dependents: []
---

# Equipment Management API

## INTENT
To provide API endpoints for equipment management including equipping/unequipping items, viewing equipment inventory, and managing character equipment slots.

## THE RULE / LOGIC
**Equipment Management Endpoints:**

**Character Equipment View:**
- **Endpoint:** `GET /api/v1/character/{id}/equipment`
- **Returns:** Currently equipped items in 3 slots (armor, utility, weapon)
- **Response:** Equipment details, stat bonuses, total stat changes

**Equip Item:**
- **Endpoint:** `POST /api/v1/character/{id}/equip`
- **Request:** Item ID and target slot
- **Process:** Validate slot compatibility, unequip current item, equip new item, recalculate stats
- **Response:** Updated character stats and equipment configuration

**Unequip Item:**
- **Endpoint:** `POST /api/v1/character/{id}/unequip`
- **Request:** Slot to empty
- **Process:** Remove item from slot, return to inventory, recalculate stats
- **Response:** Updated character stats and empty slot

**Equipment Inventory:**
- **Endpoint:** `GET /api/v1/character/{id}/inventory`
- **Returns:** All owned but unequipped items
- **Process:** Filter by item type, sort by various criteria
- **Response:** List of available items with properties and costs

**Equipment Shop Integration:**
- **Endpoint:** `GET /api/v1/shop/equipment`
- **Returns:** Available equipment for purchase filtered by character level
- **Process:** Show equipment that character can equip based on level and class
- **Response:** Equipment details, costs, stat bonuses

**Stat Recalculation:**
- **Process:** Automatically recalculate character stats when equipment changes
- **Logic:** Base stats + equipment bonuses = final stats
- **Validation:** Ensure no stat exceeds maximum caps

**API Response Format:**
```json
{
  "character_id": "uuid",
  "equipment": {
    "armor": {
      "item_id": "uuid",
      "name": "Iron Armor",
      "slot": "armor",
      "stat_bonuses": {
        "defense": 5,
        "armor_rating": 3
      }
    },
    "utility": {...},
    "weapon": {...}
  },
  "total_stats": {
    "attack": 15,
    "defense": 10,
    "movement": 3
  }
}
```

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[api_equipment_management]]`
- **Controller:** `EquipmentController`
- **Model:** `App\Models\Character` (Laravel), `App\Models\Equipment`
