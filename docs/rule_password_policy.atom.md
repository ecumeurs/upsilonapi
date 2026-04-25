---
id: rule_password_policy
human_name: Password Strength Policy Rule
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [security, password]
parents:
  - [[shared:req_security]]
  - [[shared:requirement_customer_user_account]]
dependents: []
---
# Password Strength Policy Rule

## INTENT
Ensures that all player account passwords meet strict complexity requirements to prevent brute-force attacks.

## THE RULE / LOGIC
All passwords MUST meet the following criteria:
- **Minimum Length:** 15 characters.
- **Complexity Requirements:**
  - At least 1 Uppercase letter.
  - At least 1 Numeric digit.
  - At least 1 Special symbol (e.g., !, @, #, $, %).

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_password_policy]]`
- **Test Names:** `TestPasswordLengthMin15`, `TestPasswordUppercaseRequirement`, `TestPasswordNumericRequirement`, `TestPasswordSymbolRequirement`

## EXPECTATION (For Testing)
- Submit password "short1" -> REJECTED.
- Submit password "thisIsALongPasswordWithNoNumberOrSymbol" -> REJECTED.
- Submit password "VeryLongPassword123!" -> ACCEPTED.
