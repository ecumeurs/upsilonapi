---
id: api_character_skill_inventory
status: STABLE
type: API
priority: 5
version: 2.0
human_name: Character Skill Inventory API
layer: ARCHITECTURE
tags: [api, skills, inventory, iss-073]
parents:
  - [[entity_character_skill_inventory]]
  - [[rule_character_skill_slots]]
dependents: []
---

# New Atom

## INTENT
To expose per-character skill acquisition (roll), equip, and unequip operations so players can build battle-ready skill loadouts within their slot limit.

## THE RULE / LOGIC
**Endpoints (all require Sanctum auth + character ownership via CharacterPolicy):**

- `GET  /api/v1/profile/character/{characterId}/skills` — list all skills in character's inventory (equipped and unequipped).
- `POST /api/v1/profile/character/{characterId}/skills/roll` — acquire a new random skill; server picks a template via weighted lottery from `available=true` skill templates. Returns 201 with the new `CharacterSkillResource`.
- `POST /api/v1/profile/character/{characterId}/skills/{skillId}/equip` — move skill into an active slot. Fails 422 if equipped count already equals `character.skill_slots`.
- `DELETE /api/v1/profile/character/{characterId}/skills/{skillId}/unequip` — remove skill from active slot. Fails 422 if skill is not currently equipped.

**Skill snapshot model:** `character_skills.instance_data` is frozen JSON copied from the template at acquisition time. Later template changes do not retroactively affect existing inventory entries.

**Roll lottery:** Uses `weight_positive` / `weight_negative` to bias random selection (ISS-086). Falls back to 503 if no templates are available.

## TECHNICAL INTERFACE
- **Controller:** `App\Http\Controllers\API\CharacterSkillController`
- **Service:** `App\Services\SkillService`
- **Policy:** `App\Policies\CharacterPolicy` (acquireSkill, equipSkill, unequipSkill)
- **Resource:** `App\Http\Resources\CharacterSkillResource`
- **Code Tag:** `@spec-link [[api_character_skill_inventory]]`

## EXPECTATION
- Authenticated player can roll, equip, and unequip skills on their own character.
- Equipping beyond `skill_slots` returns 422.
- Roll on another player's character returns 403.
- Roll with no available templates returns 503.
- `character_skills.source` is `'roll'` for rolled skills.
