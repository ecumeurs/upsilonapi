---
id: api_skill_generation
human_name: "Procedural Skill Generation API"
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 4
tags: [api, golang, skills, procedural]
parents:
  - [[api_go_battle_engine]]
  - [[api_standard_envelope]]
dependents: []
---

# Procedural Skill Generation API

## INTENT
To generate a random, balanced tactical skill based on optional constraints (Grade, Tags), supporting the procedural archetype system and external content tools.

## THE RULE / LOGIC
**Endpoint:** `POST /v1/skills/generate`

### Request (Wrapped in [[api_standard_envelope]])
- **Grade** (string): Optional. I to V (default: I).
- **IncludeTags** (Array<string>): Optional. Filter for specific themes (e.g. "fire", "heal").
- **ExcludeTags** (Array<string>): Optional. Filter out specific themes.

### Response (Wrapped in [[api_standard_envelope]])
Returns a `SkillGenerateResponse` containing:
- Full skill definition (Targeting, Costs, Effect).
- Balanced weights (Positive/Negative).
- Grade and associated tags.

## TECHNICAL INTERFACE
- **API Endpoint:** `POST /v1/skills/generate`
- **Code Tag:** `@spec-link [[api_skill_generation]]`
- **Go Handler:** `handler.HandleSkillGenerate`
- **Request Type:** `skillgenerator.GenerateRequest`
- **Response Type:** `api.SkillGenerateResponse`

## EXPECTATION
- Valid request -> Returns `200 OK` with a balanced skill.
- Impossible constraints -> Returns `400 Bad Request`.
