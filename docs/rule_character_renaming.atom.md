---
id: rule_character_renaming
status: STABLE
layer: BUSINESS
priority: 3
tags: profile,character,safety
parents:
  - [[shared:requirement_customer_player_profile]]
dependents: []
human_name: Character Renaming Rule
version: 1.0
type: RULE
---

# Character Renaming Rule

## INTENT
Allow survivors to personalize their digital entities while maintaining strictly validated identity strings.

## THE RULE / LOGIC
- **Length Constraint:** Minimum 3 characters, maximum 20 characters.
- **Character Set:** Only letters (A-Z, a-z), numbers (0-9), spaces, and underscores (`_`) are permitted.
- **Whitespacing:** Leading and trailing spaces MUST be trimmed before validation. Multiple contiguous spaces are permitted but discouraged.
- **Persistence:** Renaming is immediate and permanent across all system modules.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_character_renaming]]`
- **Validation Class:** `RenameCharacterRequest`
- **API Action:** `POST /api/v1/profile/character/{id}/rename`

## EXPECTATION
- Names with 2 or 21 characters are rejected (422 Unprocessable Entity).
- Names containing special characters (e.g., `!`, `@`, `#`) are rejected.
- Valid names (e.g., `Shadow_Runner`, `Agent 007`) are accepted and persist in the database.
