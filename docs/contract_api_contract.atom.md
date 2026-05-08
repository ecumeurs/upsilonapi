---
id: contract_api_contract
status: STABLE
human_name: UpsilonAPI Contract
layer: BUSINESS
version: 1.0
priority: 1
tags: [governance, contract, api]
dependents:
  - [[rule_api_bridge_orchestration]]
type: CONTRACT
parents:
  - [[shared:contract_upsilon_contract]]
---

# New Atom

## INTENT
Establish the implementation standards and constraints for the UpsilonAPI project.

## THE RULE / LOGIC
- **Communication:** Must strictly adhere to `[[api_standard_envelope]]` for all request and response payloads.
- **Auth:** All V1 endpoints must validate JWT tokens provided by the Laravel gateway.
- **Performance:** Handle message proxying to the engine with < 10ms overhead.
- **Traceability:** Maintain `@spec-link` coverage for all handlers and DTOs.
- **Error Handling:** Propagate `error_key` codes from the engine to the client without transformation.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[api_contract]]`
- **Related Atoms:** `[[api_standard_envelope]]`, `[[shared:upsilon_contract]]`

## EXPECTATION
