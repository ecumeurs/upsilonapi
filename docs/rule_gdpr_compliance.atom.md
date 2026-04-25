---
id: rule_gdpr_compliance
human_name: GDPR Compliance Rules
type: RULE
layer: ARCHITECTURE
version: 1.0
status: STABLE
priority: 5
tags: [gdpr, privacy]
parents:
  - [[shared:req_security]]
  - [[shared:requirement_customer_user_account]]
dependents:
  - [[api_profile_export]]
---
# GDPR Compliance Rules

## INTENT
Ensures personal data protection through secure deletion (soft delete) and anonymization of sensitive information.

## THE RULE / LOGIC
- **Account Deletion (Right to be Forgotten):** 
  - Deletion of an account MUST be a **soft delete** using Laravel's `SoftDeletes`.
  - The record is marked as deleted but remains in the database for audit/integrity until a purge cycle.
- **Anonymization:**
  - Upon soft deletion or upon request for anonymization, sensitive data fields MUST be overwritten with non-identifiable placeholders.
  - Sensitive Fields: `full_address`, `birth_date`.
  - Placeholder: `ANONYMIZED`.
- **Right to Portability:**
  - Authenticated users MUST have the ability to download a machine-readable dump of all their personal data (JSON format).
  - Scope: User account info, win/loss records, and character rosters.

## TECHNICAL INTERFACE (The Bridge)
- **Code Tag:** `@spec-link [[rule_gdpr_compliance]]`
- **Test Names:** `GdprTest::test_user_account_soft_delete_and_anonymization`, `GdprTest::test_user_hard_delete_cascades_to_characters`

## EXPECTATION (For Testing)
- Request account delete -> `deleted_at` timestamp set -> User cannot login.
- Audit `users` table -> `full_address` and `birth_date` show "ANONYMIZED" for deleted user.
