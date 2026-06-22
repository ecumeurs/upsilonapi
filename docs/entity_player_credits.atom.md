---
id: entity_player_credits
status: STABLE
priority: 5
layer: BUSINESS
version: 2.0
tags: ["credits", "economy", "player"]
parents:
  - [[domain_credit_economy]]
dependents:
  - [[api_shop_purchase]]
human_name: Player Credit Entity
type: ENTITY
---

# Player Credit Entity

## INTENT
To define the credit system entity for players and characters, tracking credit balances, transactions, and earning sources. Credits serve as the primary currency for skill purchases, equipment acquisition, and character progression.

## THE RULE / LOGIC
**Player Credits Entity:**

**Core Fields:**
- **player_id:** UUID primary key (references users table)
- **balance:** Current credit balance (can be negative for debt system)
- **total_earned:** Lifetime credits earned (statistics)
- **total_spent:** Lifetime credits spent (statistics)
- **created_at:** Account creation timestamp
- **updated_at:** Last balance update timestamp

**Credit Transactions (Optional Audit):**
- **id:** UUID primary key
- **player_id:** References player
- **character_id:** References character (if character-specific)
- **amount:** Positive (earned) or negative (spent)
- **source:** Credit source ('damage', 'healing', 'mitigation', 'status_effect', 'shop_purchase')
- **reference_id:** Related action or skill ID
- **created_at:** Transaction timestamp

**Character Credits (Optional):**
- **character_id:** UUID primary key
- **player_id:** References player (owner)
- **balance:** Character-specific credit balance (if separate from player)
- **last_earned:** Timestamp of last credit earning

**Balance Rules:**
- Player balance is sum of all transactions
- Character balance is subset of player transactions (if tracked separately)
- Shop purchases deduct from appropriate balance
- Cannot spend more credits than available balance

**Credit Values:**
- Base rule: 1 HP damage = 1 credit
- Healing: 1 HP healed = 1 credit
- Mitigation: 1 HP shielded = 1 credit (to caster)
- Status effect: SkillWeight / 10 credits per application

## TECHNICAL INTERFACE

## EXPECTATION
