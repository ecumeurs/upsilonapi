---
id: domain_skill_system
human_name: Entity Skill System Domain
type: DOMAIN
layer: CUSTOMER
version: 1.0
status: DRAFT
priority: 5
tags: [combat, skills, abilities]
parents:
  - [[domain_upsilon_engine]]
dependents: []
---
# Entity Skill System Domain

## INTENT
To provide a modular, structured capability matrix allowing entities to perform specialized, targeted combat actions beyond basic attacks.

## THE RULE / LOGIC
- **Skill Structure:** A skill is composed of four abstractions: Name, Targeting Mechanism, Effect Payload, and Action Cost/Cooldown.
- **Targeting Types:** Skills may target a single entity, multiple entities, grid positions (AoE), or act as delayed reactions/environmental traps.
- **Line Of Sight (LOS):** The system supports rules evaluating whether the targeted node is visible from the casting entity's perspective.
- **Effect Computing:** Skills define what property computations the Ruler should trigger (e.g., Damaging rules, Healing rules, Buff/Debuff application to entity properties).
- **Cooldown Execution:** Executing a skill prevents its usage for a pre-defined interval of turns/time.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[domain_skill_system]]`
- **Test Names:** `TestSkillTargetValidation`, `TestSkillCooldownLocking`

## EXPECTATION (For Testing)
- Entity uses advanced attack skill -> Effect modifies target HP -> Skill enters cooldown phase.
- Entity attempts to cast same skill next turn -> Action rejected due to cooldown lock.
