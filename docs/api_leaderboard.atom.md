---
id: api_leaderboard
status: DRAFT
dependents: []
type: API
layer: ARCHITECTURE
version: 1.0
human_name: Leaderboard API Contract
priority: 5
parents:
  - [[ui_leaderboard]]
---

# New Atom

## INTENT
Define the data contract for fetching leaderboard rankings.

## THE RULE / LOGIC
- Returns a JSON object containing global rankings for a specific mode.
- Filters matches based on `rule_leaderboard_cycle` (current week).
- Applies `rule_leaderboard_score_calculation` for ranking.

- **URI:** `/api/v1/leaderboard`
- **Verb:** `GET`
- **Intent:** Fetch competitive rankings with temporal filtering and pagination.
- **Fully Detailed Input:**
  - `mode`: `1v1_PVP|2v2_PVP|1v1_PVE|2v2_PVE` (Required)
  - `page`: Result page index (Integer)
  - `search`: Filter by account name (String)

- **Fully Detailed Output:**
  `{ success: true, data: { results: [], self: {}, meta: {} } }`

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[api_leaderboard]]`
- **Related Issue:** `#456`
- **Test Names:** `LeaderboardTest`

## EXPECTATION
