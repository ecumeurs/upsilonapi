---
id: domain_credit_economy
human_name: Credit Economy Domain
type: DOMAIN
layer: BUSINESS
version: 2.0
status: DRAFT
priority: 5
tags: [economy, credits, progression]
parents:
  - [[domain_upsilon_engine]]
dependents:
  - [[shared:rule_credit_earning_damage]]
  - [[shared:rule_credit_earning_status_effects]]
  - [[shared:rule_credit_earning_support]]
  - [[shared:rule_starting_credits_1000]]
  - [[upsilonbattle:entity_player_credits]]
  - [[upsilonbattle:mec_credit_spending_shop]]
  - [[upsilontypes:entity_shop_item]]
---

# Credit Economy Domain

## INTENT
To establish the credit economy as the primary progression currency, enabling players to earn credits through combat performance and spend them on skills and equipment.

## THE RULE / LOGIC
**Credit Sources:**
- **Damage Dealing:** 1 HP damage = 1 credit
- **Healing:** 1 HP healed = 1 credit
- **Damage Mitigation:** 1 HP blocked/shielded = 1 credit (for caster)
- **Status Effects:** SkillWeight/10 credits per application

**Credit Sinks:**
- **Skill Purchases:** Skill Weight × 2 credits
- **Equipment Purchases:** Base cost × tier multiplier
- **Skill Reforging:** Modification costs based on grade change

**Economy Principles:**
- **Transparent Earning:** Clear 1:1 HP-to-credit ratio
- **Meaningful Spending:** Credits purchase permanent character upgrades
- **Balanced Progression:** Credit earning scales with character power
- **Support Viability:** Multiple earning paths for different playstyles

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[domain_credit_economy]]`
- **Test Names:** `TestCreditEconomyBalance`, `TestSupportCreditViability`
