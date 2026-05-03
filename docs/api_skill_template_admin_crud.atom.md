---
id: api_skill_template_admin_crud
status: STABLE
layer: ARCHITECTURE
priority: 5
version: 2.0
parents:
  - [[upsilontypes:entity_skill_template]]
  - [[rule_admin_content_authority]]
dependents: []
tags: [api, skills, admin, crud, iss-086]
human_name: Admin Skill Template CRUD API
type: API
---

# New Atom

## INTENT
To give admins full CRUD control over the skill template registry — enabling content management without code deploys.

## THE RULE / LOGIC
**Admin-only endpoints (Sanctum + `admin` middleware):**

- `GET    /api/v1/admin/skill-templates` — list all templates including unavailable ones.
- `GET    /api/v1/admin/skill-templates/{id}` — get single template (404 if not found).
- `POST   /api/v1/admin/skill-templates` — create template. Required: `name`, `behavior` (in:Direct,Reaction,Passive,Counter,Trap), `grade` (in:I,II,III,IV,V), `weight_positive`, `weight_negative`. Optional: `targeting`, `costs`, `effect` (JSON maps), `available` (bool, default true).
- `PUT    /api/v1/admin/skill-templates/{id}` — full update (all fields `sometimes`).
- `DELETE /api/v1/admin/skill-templates/{id}` — soft or hard delete; returns 404 if not found.

**Validation:** `behavior` and `grade` enforced by DB CHECK constraint and Laravel form request.

## TECHNICAL INTERFACE
- **Controller:** `App\Http\Controllers\API\Admin\AdminSkillTemplateController`
- **Form Requests:** `StoreSkillTemplateRequest`, `UpdateSkillTemplateRequest`
- **Resource:** `App\Http\Resources\SkillTemplateResource`
- **Code Tag:** `@spec-link [[api_skill_template_admin_crud]]`

## EXPECTATION
- Non-admin requests return 403.
- Store returns 201 with created template.
- Update returns 200 with refreshed template.
- Delete returns 200 with success message.
- GET/PUT/DELETE on nonexistent ID returns 404.
