---
id: api_skill_grading_computation
human_name: Skill Grading Computation API
type: API
layer: ARCHITECTURE
version: 2.0
status: DRAFT
priority: 5
tags: [api, skills, grading, computation]
parents:
  - [[mech_skill_weight_calculator]]
  - [[rule_skill_grading_system]]
dependents: []
---

# Skill Grading Computation API

## INTENT
To provide API endpoints for computing skill grades, calculating skill weight, and determining skill availability based on character level progression.

## THE RULE / LOGIC
**Grade Computation Endpoint:**
- **Input:** Skill properties (damage, crit, range, zone, etc.)
- **Process:** Calculate Total Positive SW using benefit table
- **Output:** Skill grade (I-V) and credit cost (SW × 2)

**Weight Computation Endpoint:**
- **Benefits Calculation:** Sum all positive SW from skill properties
- **Payments Calculation:** Sum all negative SW from skill costs
- **Net SW:** Benefits + Payments (should equal 0 for balanced skills)

**Skill Availability Endpoint:**
- **Input:** Character level, existing skills
- **Process:** Filter available skills by grade level access
- **Output:** List of 3 random skills from appropriate grade pool

**Grade Level Access:**
- **Level 1-9:** Return Grade I-II skills
- **Level 10-19:** Return Grade II-III skills
- **Level 20-29:** Return Grade III-IV skills
- **Level 30+:** Return Grade IV-V skills

**API Endpoints:**
- `GET /api/v1/skills/grade-compute` - Compute grade from skill properties
- `GET /api/v1/skills/weight-compute` - Compute skill weight
- `GET /api/v1/character/{id}/available-skills` - Get skill selection options

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[api_skill_grading_computation]]`
- **Controller:** `SkillGradingController`
- **Response Format:** JSON with grade, credit cost, weight breakdown
