---
id: rule_leaderboard_score_calculation
status: STABLE
human_name: Score Calculation Rule
type: RULE
layer: ARCHITECTURE
parents:
  - [[shared:us_leaderboard_view]]
dependents: []
priority: 5
version: 1.0
---

# Score Calculation Rule

## INTENT
Define the scoring algorithm and eligibility for the leaderboard.

## THE RULE / LOGIC
- **Scoring Formula:** `Score = Wins / MAX(1, Losses)`
  - 10 Wins, 0 Losses => 10.0
  - 0 Wins, 0 Losses => 0.0
  - 10 Wins, 10 Losses => 1.0
- **Eligibility:** Only users with at least one match recorded within the current time range are included in the leaderboard.
- **Search Filtering:** Retrieval queries must exclude any user with 0 matches in the selected mode/period.

## TECHNICAL INTERFACE
- **Code Tag:** `@spec-link [[rule_leaderboard_score_calculation]]`
- **Test Names:** `TestScoreCalculation`, `TestLeaderboardEligibility`

## EXPECTATION
- `Score = Wins / MAX(1, Losses)`: 10W/0L → 10.0, 10W/10L → 1.0, 0W/0L → 0.0.
- Users with zero matches in the selected mode/period are excluded from results (`TestLeaderboardEligibility`).
- `TestScoreCalculation` validates the formula across the win/loss boundary cases above.
