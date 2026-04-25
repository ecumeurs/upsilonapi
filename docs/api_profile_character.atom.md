---
id: api_profile_character
human_name: Character Management API
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [profile, character, api]
parents:
  - [[shared:us_character_reroll]]
dependents: []
---
# Character Management API

## INTENT
To view and modify character entities, including statistical progression and cosmetic identity.

## THE RULE / LOGIC
- **Endpoint 1: List Roster**
  - **URI:** `/api/v1/profile/characters`
  - **Verb:** `GET`
  - **Intent:** Inventory Review
  - **Input:** []
  - **Output:** Array of character objects.

- **Endpoint 2: Character Detail**
  - **URI:** `/api/v1/profile/character/{characterId}`
  - **Verb:** `GET`
  - **Intent:** Deep Inspection
  - **Input:** 
    - `characterId`: (uuid) [Mandatory] Target character identifier.
  - **Output:** Detailed character data including stats, traits, and bio.

- **Endpoint 3: Upgrade stats**
  - **URI:** `/api/v1/profile/character/{characterId}/upgrade`
  - **Verb:** `POST`
  - **Intent:** Progression Allocation
  - **Input:** 
    - `stats`: (object) Map of stat names to increase values.
  - **Output:** Updated character state.

- **Endpoint 4: Rename**
  - **URI:** `/api/v1/profile/character/{characterId}/rename`
  - **Verb:** `POST`
  - **Intent:** Identity Modification
  - **Input:** 
    - `name`: (string) [Mandatory] New tactical name.
  - **Output:** `{ "success": true, "name": "new_name" }`

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `/api/v1/profile/*`
- **Code Tag:** `@spec-link [[api_profile_character]]`
- **Related Issue:** `ISS-016`
- **Test Names:** `TestGetCharacters`, `TestRerollRestricted`, `TestLevelUpStatAllocation`

## EXPECTATION (For Testing)
- Requesting character list -> Return array of characters.
- Upgrading beyond available points (wins) -> Return 400 Bad Request.
- Upgrading that violates [[rule_progression]] (e.g., movement limit) -> Return 400 Bad Request.
- Rerolling after account is "Stable" -> Return 403 Forbidden.
