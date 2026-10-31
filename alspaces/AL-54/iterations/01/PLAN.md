# AL-54 Iteration 01 Plan

## Approval and grounding

Source: user request on 2026-10-31, explicitly stating that plan and contract are pre-approved and requesting full delivery. No additional scope approval is required for this specification-aligned implementation. Frozen acceptance boundaries: ../../CONTRACT.md. Baseline: 85a2af94b5036b0918aaf9a7a89aed3a990f2650; initial checkout clean. Parent owns ticket-index status updates.

SkillFact already carries name and explicit success in backend/internal/domain/models.go and web/src/domain/types.ts; AgenticFactLedger has optional skills and ledgerSkills supplies the safe TypeScript accessor. CalculateFleetSummary already receives all fact ledgers and projects fleet metrics; BuildComparisonRows and deriveTicketComparisonRows separately project server and browser rows. FleetOverview fetches fleet summary and comparison rows with matching date/type/prefix filters. TicketComparisons loads full computed ledgers, applies pure filters, derives rows, calculates benchmarks, and feeds TicketScatterPlot. There is currently no TypeScript aggregations.ts file: create it as the pure owner requested by the specification. Existing API handlers under backend/internal/adapters/http serialize these domain projections, so a new endpoint is unnecessary.

## Implementation ownership and sequence

1. Coder owns application and initial focused test changes in backend/internal/domain/models.go, aggregations.go, relevant existing domain tests, web/src/domain/types.ts, new aggregations.ts and its tests, formatters.ts, FleetOverview.tsx, TicketComparisons.tsx, TicketScatterPlot.tsx, and directly affected existing fixtures/tests. Read TECHNICAL_GUIDELINES.md before changes. Implement ComputeSkillAdoption and its TypeScript equivalent with recorded-row multiplicity, distinct-ticket counting, explicit failure semantics, deterministic ranking, and absent-data safety. Extend existing server projections and mirror browser rows. Avoid unrelated edits.
2. Coder adds fleet card and comparison controls/column/tooltips. Keep skill selection client-side in TicketComparisons and include it before row/benchmark/chart derivation. Reuse MultiSelectDropdown and existing card styling; any required style edits are within scope. Keep recovery controls rendered at zero matches. Update colSpan and the existing table-column assertion coherently.
3. Tester owns only relevant test files and verification records while active; inspect existing peers before extending them. Cover missing analytics cases, actual fleet presentation, multi-select OR behavior combined with existing controls, absent-data inclusion without selection, and clearing selection after zero results. Synthetic ticket fixtures must remain independent of live alspaces content. Use the existing TicketScatterPlot tests or view integration test for tooltip behavior, and backend HTTP tests if needed to establish serialized projection fidelity. Run focused suites, frontend checks/build, and backend race tests; report precise outcomes.
4. Reviewer owns only iterations/01/REVIEW.md and performs a read-only application audit against the contract and engineering rules. Record diff/commit, actual verification evidence, gaps, findings, and APPROVED or CHANGES_REQUESTED. Parent routes fixes back to coder/tester and retains ownership of all Git and PR actions.

All agents share this checkout and must preserve changes owned by others. Parent passes these pointers, baseline, and review findings on each dispatch. No child may publish or ingest session telemetry.

## Selected regression prevention cases

Optional facts: aggregation reads supported persisted data without enrichment or source availability; absent skills produce safe empty arrays without claiming a measured success rate. Multiplicity: repeated skill calls increase invocation counts while per-ticket counts stay distinct. Failure classification: boolean success determines failure even when error text disagrees. UI recovery: skill dropdown options and controls survive zero matches, clearing selection restores missing-skill tickets, and type/prefix/search constraints compose correctly. Projection consistency: both server comparison rows and browser-derived rows retain distinct skill names; fleet aggregates follow active filters. Frontend embedding: canonical build updates ignored backend/internal/adapters/web/dist with matching index, JS, and CSS, and standalone binary serves its embedded bundle.

## Test owners and proportionality

backend/internal/domain/aggregations_test.go owns existing fleet and comparison projection behavior; extend those tables/tests for added fields and add a focused analytics test for the new algorithm. web/src/domain/formatters.test.ts owns browser row projection; new aggregations.test.ts owns the distinct new algorithm. FleetOverview.test.tsx owns card and filtered-fetch rendering. TicketComparisons.test.tsx already covers all columns, type/prefix filtering, chart integration, and empty recovery; extend coherent cases rather than duplicating those suites. TicketScatterPlot.test.tsx owns tooltip accessibility and content where needed. No unrelated tests are removed.

## Integration and completion

Parent commits reviewed implementation, merges current origin/main, and runs make on the integrated tree. Verify ignored embedded assets and standalone binary, record SUMMARY.md and PULL_REQUEST.md, move AL-54 specification to plan/done and remove only its index entry, and check AL-55 dependencies before promotion. Generate Codex telemetry with literal nonce and explicit feat type at the final workflow step, validate source/root/children and preservation, then commit final artifacts, push, create or update the correct open PR, and link it immediately. Record unavailable evidence honestly.

## Implementation notes

The shared multi-select previously collapsed selection of every known option to unrestricted All. The skill selector opts out through an optional component property so selecting every recorded skill still excludes tickets without skills. Skill option values use an injective skill: prefix, retaining literal names such as all and ALL without colliding with selection sentinels. Other consumers retain their existing default behavior. Comparative benchmarks now use the searched and selected skill population, matching table and chart. These are specification-aligned implementation choices within this iteration, without an acceptance change.

Focused implementation checks: five frontend files passed with 104 tests; TypeScript typecheck and lint passed; canonical make web-assets passed; focused Go skill, fleet, and comparison tests passed. Tester and reviewer own remaining contract evidence.

## Routine integration supervision (2026-10-31)

Source: MGT-1 owner instruction after predecessor PR #76 merged at 5da845a7c244a6b2582102c0dff0d221844d9bfc. This is routine integration under the existing approved acceptance boundaries, not a new iteration or scope. Merged current origin/main into AL-54 as 7deb1de2225ec63fb36115dc92f1e0f833f277c4 without conflicts. Preserve live refresh and prior dependency removals. Manager owns merge sequencing; AL-54 owner must not merge or enable auto-merge. Verify full make and exact embedded assets, refresh AL-54 telemetry at final step, push, and report exact-head readiness. AL-55 remains blocked by AL-52 and AL-53. No product implementation changes were required.
