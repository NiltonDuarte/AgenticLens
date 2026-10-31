# AL-54 Technical Contract

## Authority and baseline

The user explicitly pre-approved the plan and contract and authorized full implementation on 2026-10-31. This contract records the supplied AL-54 specification without expanding its acceptance boundaries. Baseline commit: 85a2af94b5036b0918aaf9a7a89aed3a990f2650. Initial working tree was clean on AL-54/fleet-skill-adoption-analytics-filter. Workspace root is configured as alspaces.

## Delivered behavior

Aggregate recorded skill invocations across accepted ticket ledgers in pure Go and TypeScript functions. SkillAdoptionFact contains name, total_invocations, tickets_count, failure_count, and failure_rate. Count every invocation row, including repeated calls; count each ticket once per skill. Use exact skill names as identities. Deduplicate ticket counts by nonempty ticket ID; unidentified inputs count as separate ledger entries. Failure means SkillFact.Success is false (TypeScript success is false), never error-text inference. failure_rate is failures divided by invocations, a fraction from zero to one. Sort by invocation count descending and exact name ascending for ties, consistently across languages. Inputs are never mutated.

Absent, nil, or empty optional skills represent no recorded skills and produce empty results safely. This does not introduce runtime support for schema versions rejected by existing validation. Existing persisted facts remain the read baseline; skill analytics require no source hydration, migration, or telemetry extraction changes.

Fleet Overview shows unique recorded skills, top three ranked names with invocation counts, and invocation-weighted overall success percentage. A fleet with no recorded invocations displays an explicit no-recorded-skills state and an unavailable success rate rather than claiming measured success. Existing date, ticket type, and prefix filters constrain these metrics alongside the other fleet cards.

Ticket Comparisons shows a Skills Used column listing distinct skill names alphabetically, and includes those names in scatter plot tooltips. Missing recorded skills have explicit presentation copy. Filter by Skill uses the existing accessible multi-select component: multiple selected skills match any selected name (OR); the skill constraint composes with type, prefix, and search constraints (AND). All or empty selection means unrestricted and includes tickets lacking skill facts. Filter options derive from the loaded source set before skill filtering and remain usable when the resulting rows are empty. Table, chart, and comparative benchmark summaries use the same selected skill population.

## Architectural boundaries

Go models and ComputeSkillAdoption belong in backend/internal/domain/models.go and aggregations.go. Add skill adoption to the existing FleetSummary read projection and skill names to TicketComparisonRow where needed; do not add persisted metrics or change ledger schemas. CalculateFleetSummary and BuildComparisonRows remain the canonical server projections. Mirror read types in web/src/domain/types.ts, add pure analytics in web/src/domain/aggregations.ts, and keep derived comparison rows in web/src/domain/formatters.ts consistent. UI belongs in FleetOverview, TicketComparisons, and TicketScatterPlot. Reuse existing HTTP serialization without introducing a new endpoint. New read fields may be optional in TypeScript for older response/fixture compatibility, with explicit safe UI defaults.

FCIS, strict typing, zero any/interface{}, zero emojis, self-contained test fixtures, and existing error/diagnostic handling remain mandatory. No dependencies or Node runtime requirement are introduced. Canonical frontend builds must update the ignored embedded asset tree and the standalone executable.

## Acceptance and verification

Prove Go/TypeScript parity for repeated calls, distinct tickets, mixed success, ordering ties, and empty/missing skill facts. Prove fleet card values and no-data copy, comparison OR selection and empty-to-recovered behavior, skills column and chart tooltip content. Existing test owners should receive coherent extensions. Run focused checks during implementation, then full make after integration; inspect embedded assets and verify the built binary. Record actual commands and limitations in REVIEW.md and SUMMARY.md. Parent owns integration, telemetry, commits, lifecycle synchronization, push, and PR delivery.
