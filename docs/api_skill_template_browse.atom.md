---
id: api_skill_template_browse
status: STABLE
human_name: Skill Template Browse API
dependents:
  - [[api_character_skill_inventory]]
layer: ARCHITECTURE
priority: 5
tags: [api, skills, templates, catalog, iss-086]
parents:
  - [[entity_skill_template]]
type: API
version: 2.0
---

# New Atom

## INTENT
To expose the player-facing skill template catalog so authenticated users can browse which skills exist and inspect their stats before rolling.

## THE RULE / LOGIC
**Endpoints (Sanctum auth required):**

- `GET /api/v1/skills/templates` — returns all templates where `available=true`, ordered by grade then name. No pagination in V2.0.
- `GET /api/v1/skills/templates/{id}` — returns a single template by UUID. Returns 404 if not found or `available=false`.

**Response shape (`SkillTemplateResource`):** `{ id, name, behavior, targeting, costs, effect, grade, weight_positive, weight_negative, available, version, created_at, updated_at }`

**Filtering:** Only `available=true` rows are visible to players. Admin-facing CRUD (including unavailable items) is handled by `[[api_skill_template_admin_crud]]`.

## TECHNICAL INTERFACE
- **Controller:** `App\Http\Controllers\API\SkillTemplateController`
- **Resource:** `App\Http\Resources\SkillTemplateResource`
- **Code Tag:** `@spec-link [[api_skill_template_browse]]`

## EXPECTATION
- Authenticated GET /skills/templates returns 200 with array of available templates (at least 3 in seeded environment).
- Unavailable templates are excluded from player listing.
- GET /skills/templates/{nonexistent-uuid} returns 404.
- Unauthenticated requests return 401.
