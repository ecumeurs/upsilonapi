---
id: api_auth_login
human_name: Player Login API
type: API
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [auth, login, api]
parents:
  - [[uc_player_login]]
dependents:
  - [[mechanic_mech_cli_sensitive_data_masking]]
  - [[uc_player_login]]
---
# Player Login API

## INTENT
To authenticate a survivor by verifying credentials and issuing a secure access token.

## THE RULE / LOGIC
- **URI:** `/api/v1/auth/login`
- **Verb:** `POST`
- **Intent:** Identity Authentication
- **Fully Detailed Input:**
  - `account_name`: (string) [Mandatory] The unique tactical identifier.
  - `password`: (string) [Mandatory] The survivor's secret credential.
- **Fully Detailed Output:**
  - `user`: (object) Profile data (id, account_name, email).
  - `token`: (string) JWT Bearer token for session authorization.

**Requirement Boundary:** Authentication MUST strictly use `account_name`. Email-based authentication is explicitly forbidden to maintain thematic consistency and primary key privacy.

## TECHNICAL INTERFACE (The Bridge)
- **API Endpoint:** `POST /api/v1/auth/login`
- **Code Tag:** `@spec-link [[api_auth_login]]`
- **Related Issue:** `ISS-029`
- **Test Names:** `TestSuccessfulLogin`, `TestInvalidCredentials`

## EXPECTATION (For Testing)
- Correct credentials -> Return 200 OK with Token.
- Wrong password -> Return 401 Unauthorized.
