---
id: api_profile_export
human_name: Profile Data Export API
type: MODULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [api, gdpr, profile]
parents:
  - [[api_laravel_gateway]]
  - [[api_standard_envelope]]
  - [[rule_gdpr_compliance]]
dependents: []
---
# Profile Data Export API

## INTENT
Provides an authenticated endpoint for users to retrieve a complete dump of their personal data as required by GDPR.

## THE RULE / LOGIC
- **Endpoint:** `GET /api/profile/export`
- **Authentication:** Bearer token required.
- **Payload:** Returns a JSON object containing:
  - Account Information (Name, Address, Birth Date).
  - Game Statistics (Total Wins, Wins/Losses).
  - Active Character Roster.
- **Constraints:** Only the authenticated owner can access their own data.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[api_profile_export]]`
- **Test Names:** `TestApiProfileExportSuccess`, `TestApiProfileExportUnauthorized`
