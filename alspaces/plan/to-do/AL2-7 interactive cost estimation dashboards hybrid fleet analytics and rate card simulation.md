h3. Title: AL2-7: Interactive Cost Estimation Dashboards, Hybrid Fleet Analytics, and Rate-Card Simulation

*GOAL*

Analysts can switch cost views and simulate alternate rate cards without changing measured evidence or saved telemetry.

*CONTEXT*

AL2-6 supplies cost provenance and rate-card estimates. Fleet, progression and comparison views need consistent projections so subscription-equivalent estimates are not mistaken for measured spend.

*ACCEPTANCE CRITERIA*

* Provide Measured Spend, Estimated Equivalent and Combined Hybrid modes across fleet, progression and comparisons, retaining the user preference on refresh. The default preserves current measured-spend behavior.
* Measured shows recorded billing with coverage. Estimated shows eligible rate-card estimates with coverage. Hybrid combines measured facts and estimates for uncovered eligible facts once each; it never adds a second estimate over already measured spend. Unknowns remain explicit.
* HTTP fleet/workspace/comparison projections accept cost_mode=measured|estimated|hybrid consistently. Mode changes update chart scales, coordinates, cards and tooltips with the selected basis clearly labeled; preserve ticket identity and existing filters.
* A rate-card simulation drawer validates alternate finite nonnegative rates and projects historical token categories without writing configuration, measured facts or ledgers. Reset restores the saved baseline; unmatched/unknown inputs remain unavailable.
* Cost-mode and simulation changes handle loading, failures and stale responses without relabeling old values as the new projection. Recorded estimates and temporary simulation values are distinguishable, including for partial mixed-harness history.

*Status*: Blocked
*Blocked by*: AL2-6
*Work type*: feat

*IMPLEMENTATION NOTES*

No financial return claim is implied by an API-equivalent estimate. Preserve AL2-6 token-category, provenance and no-double-counting semantics. Pure calculations belong to domain projections; avoid independent pricing arithmetic in each view.

Follow [engineering rules](../../../TECHNICAL_GUIDELINES.md) and [regression boundaries](../../../docs/REGRESSION_PREVENTION.md). [PO-19](../../PO-19/SPEC.md) preserves the earlier development reconciliation; historical observations require revalidation, not an assumption of current runtime support.
