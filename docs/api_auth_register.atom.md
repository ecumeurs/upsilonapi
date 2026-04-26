---
id: api_auth_register
human_name: Player Registration API
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [auth, register, api]
parents:
  - [[api_laravel_gateway]]
  - [[api_standard_envelope]]
dependents:
  - [[mechanic_mech_cli_sensitive_data_masking]]
  - [[upsilonbattle:mechanic_mech_cli_sensitive_data_masking]]
---
# Player Registration API

## INTENT
To initialize a new survivor entity by creating an account and generating an initial character roster.

## THE RULE / LOGIC
- **URI:** `/api/v1/auth/register`
- **Verb:** `POST`
- **Intent:** Entity Initialization
- **Fully Detailed Input:**
  - `account_name`: (string) [Mandatory] Unique tactical identifier.
  - `email`: (string) [Mandatory] Valid communication address.
  - `password`: (string) [Mandatory] Must meet entropy requirements.
  - `password_confirmation`: (string) [Mandatory] Matching credential verification.
- **Fully Detailed Output:**
  - `user`: (object) Newly created profile data.
  - `token`: (string) Immediate access token.
  - `roster`: (array) Initial character entities generated for the user.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /api/v1/auth/register`
- **Code Tag:** `@spec-link [[api_auth_register]]`
- **Related Issue:** `ISS-007`
- **Test Names:** `TestSuccessfulRegistration`, `TestRegistrationValidationFails`

## EXPECTATION (For Testing)
- Valid data -> User created in DB -> Return 201 Created with Token.
- Duplicate email -> Return 422 Unprocessable Entity.
