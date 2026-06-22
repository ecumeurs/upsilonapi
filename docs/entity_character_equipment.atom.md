---
id: entity_character_equipment
status: DRAFT
type: ENTITY
layer: ARCHITECTURE
version: 2.0
priority: 5
tags: [equipment, slots, characters, iss-074]
human_name: Character Equipment Entity
dependents:
  - [[upsilonbattle:mechanic_item_buff_application]]
parents:
  - [[entity_equipment_system]]
  - [[entity_player_inventory]]
  - [[upsilonbattle:mechanic_three_slot_equipment_system]]
---

# Character Equipment Entity

## INTENT
To define the slot-binding entity — exactly three nullable slots (armor / utility / weapon) per character pointing into the owner's `player_inventory`. This is the single source of truth for "what a character has equipped"; per decision D1 of ISS-074, this concern is NOT replicated on `player_inventory`.

## THE RULE / LOGIC
**Character Equipment Entity:**

**Core Fields (`character_equipment` table):**
- **character_id:** UUID primary key + FK → `characters.id` ON DELETE CASCADE. One row per character (created lazily on first equip, or eagerly at character creation).
- **armor_item_id:** Nullable UUID FK → `player_inventory.id` ON DELETE SET NULL.
- **utility_item_id:** Nullable UUID FK → `player_inventory.id` ON DELETE SET NULL.
- **weapon_item_id:** Nullable UUID FK → `player_inventory.id` ON DELETE SET NULL.

**Slot semantics:** A slot is `NULL` when empty. The slot column matches the catalog `slot` of its referenced inventory row — enforced by service-layer validation, not DB constraint.

**Mutual exclusivity:**
- A given `player_inventory` row can be referenced from at most ONE slot across ALL of the user's characters at any time.
- Equipping an item already bound to character A onto character B is allowed but atomic: A's slot is cleared in the same DB transaction.
- If the destination slot already holds another item, the previous binding is cleared.

**No skill slots here:** Skill equipment lives in a future `character_skills` join table (ISS-073). Equipment slots are strictly for items.

**Engine projection:**
- At battle init, the resolver in `bridge.go` reads each character's three slots, joins to `player_inventory.shop_item_id → shop_items.properties`, and registers each item as a `Forever=true` buff on the entity. See `[[upsilonbattle:mechanic_item_buff_application]]`.
- Unequip removes the buff via `RemoveBuffsByOrigin(item_id)`.

**Privacy:**
- Equipment is visible to the owner unconditionally; visible to enemies only via the same masking rules that hide their stat block during a match.

**Lifecycle:**
- Created on first equip (or seeded on character creation in V2.1).
- Mutated by `[[api_equipment_management]]`.
- Cascade deleted with the character; FK cascade SET NULL preserves rows when an inventory entry is deleted.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[entity_character_equipment]]`
- **Laravel Model:** `App\Models\CharacterEquipment`
- **Migration:** `*_create_item_system_tables.php`
- **Resource:** `App\Http\Resources\CharacterEquipmentResource`

## EXPECTATION
- Each character has at most one row in `character_equipment`.
- Cross-character re-equip atomically frees the prior character's slot.
- Cascade behavior: deleting a character drops the row; deleting an inventory entry NULLs the referencing slot.
- Service layer rejects equip when item slot ≠ destination slot.
