# AL-57 iteration 02: Approved simple legacy cost semantics

Status: APPROVED for implementation. CONTRACT.md is the first approved baseline. User approval of the proposed scope includes the amendment: “remove legacy_unverified type. treat legacy costs as observed. keep it simple”. No repeated approval gate is required for this accepted scope.

## Steering delta

- Source: direct user response to the contract/plan approval question, relayed verbatim by the orchestrator.
- Changed acceptance: valid legacy numeric amounts, including zero and historical Codex zero, are observed in read projections, migration, aggregates, CLI/API and UI. Remove historical special states, counters, labels, separate amounts and provenance reconstruction.
- Superseded proposal: iteration 01's separate historical-cost classification, exclusion from observed subtotal, additional count, and special display. Iteration 01 remains an unapproved historical planning record; its contradictory text has no current authority.
- Preserved guarantees: partial coalesced current-source rows retain known contributions; absent/null costs remain unavailable; nonfinite/invalid costs are rejected; fresh Codex cost is unavailable; measured zero survives; coverage counts invocation facts; v5 persistence and read-only canonical-v4 support remain; v1-v3 use explicit migration; exact-owner recovery and AL-55 safeguards persist; no estimates/rate cards/invoice claims; AL2 stays queued.
- Verification impact: replace all formerly proposed special-history expectations with ordinary observed numeric behavior. Add presence-sensitive legacy decoding coverage so missing optional values do not become observed zero. Reuse the original source-to-consumer inventory and unaffected verification cases.

## Current implementation plan

1. Define only observed, partial and unavailable invocation cost states, plus empty/complete/incomplete aggregate representation. Coverage counts are fully_observed, partial, unavailable and total; the first three sum to total. Known subtotal includes observed values and reported contributions of partial rows exactly once. Invalid states/value combinations fail at the JSON/source boundary.
2. In backend/internal/adapters/schemas.go, repository.go and backend/internal/domain/migration/engine.go, decode legacy amount presence before scalar defaults. Valid numbers become observed, absent/null becomes unavailable, and invalid amounts fail explicitly. Preserve valid numeric values exactly, including historical Codex zero. Retain required current identity/token fields and reject ambiguous legacy root collections on ordinary ingestion. The compatibility projection may accept missing/null cost in otherwise valid canonical-v4 input because this contract defines those costs as unavailable; it must not relax unrelated validation.
3. Add v4-v5 migration and v5 generation/persistence through existing registry, validator and migration runner. Keep valid canonical-v4 reads read-only and source-independent. Migration retains valid original numbers and all ownership/non-cost facts, with existing backup/dry-run/ticket filtering/atomicity. Decoding older v1-v3 input must likewise retain cost presence until its explicit chain reaches v5; do not first erase missing/null into float64 zero.
4. Update current OpenCode source accumulators to preserve all-present/mixed/all-missing evidence inside existing grouped rows, and current Codex extraction to unavailable cost. Apply accepted complete-tree source refresh semantics; an unavailable tree preserves the stored observed historical value. A successful explicit refresh may replace historical zero with current unavailable evidence, but a harness name alone never rewrites it.
5. Update all pure aggregates, fleet maps, comparisons, role-effort distributions, time/cumulative buckets, HTTP serialization, TypeScript runtime validation, and local calculations using the original iteration 01 owner inventory. Complete totals require nonempty all-observed facts; missing or zero denominators make cost-dependent ratios unavailable; cumulative complete totals stay unavailable after a missing bucket.
6. Update the listed CLI/Markdown/UI surfaces together. Historical numbers use normal observed formatting and ordering with no special labels. Partial amounts retain labels/counts; unavailable cost is never coerced to zero. Decode optional TaskLinkFact cost by presence as well: present valid number including zero is observed, missing/null is unavailable. Task-link costs remain outside invocation totals; no new producer or inferred child cost is introduced.
7. Tester and reviewer apply the accepted contract and this delta, using iteration 01's unchanged grounded file/test-owner inventory only as reference. Ship one coherent fact-to-display change. Parent owns integration, full make, embedded asset/binary checks, task lifecycle, final telemetry and PR delivery. No implementation is claimed by this plan itself.

## Acceptance and focused regression changes

| Input/path | Required result |
| --- | --- |
| Canonical-v4 numeric 7 + observed 2 | Known subtotal 9, complete total 9, two fully observed facts |
| Canonical-v4 historical Codex numeric 0 | Observed zero in ordinary read, migration, sorting and display; no source lookup needed |
| Legacy missing/null amount + observed 2 | Known subtotal 2 labeled partial, complete total unavailable, one unavailable and one observed fact |
| Explicit historical task cost 0 versus omitted task cost | Present zero observed; omitted unavailable; neither contributes to invocation totals |
| Current coalesced row: reported 2 + missing source amount | One partial invocation with known subtotal 2 and unavailable complete amount |
| Above partial row + observed historical 1 | Known subtotal 3, incomplete complete total, one partial and one fully observed fact |
| Unavailable exact-owner refresh | Prior historical numeric facts retained with recovery diagnostics |
| Successful explicit refresh from present source logs | New complete tree replaces its owner's snapshot using current cost evidence; other owners and repeated rows preserved |
| Nonfinite/invalid amount or contradictory v5 state/value | Explicit rejection/diagnostic; no measured-zero fallback |

Extend the existing schema/repository/migration tests named in iteration 01 for numeric zero/nonzero, omitted/null and invalid boundaries. Migration invariants preserve numbers and ownership, not a removed historical state. Exercise actual API/client serialization and representative CLI/UI displays so legacy zero cannot be hidden by a special display or counted twice. Preserve the earlier mixed-source, partial-coalesced, denominator, coverage/filter, cumulative-chart, hydration-failure and embedded-asset cases. Synthetic self-contained fixtures only; no live ticket ledgers or source databases.

## Planning verification

Only the contract and this delta plan are changed by the planner. Application implementation and tests are not yet run by the planner. Parent records the approved-baseline checkpoint and proceeds to coder/tester/reviewer; the iteration 01 grounding remains valid where this delta does not supersede it.


## Review correction: preserve JSON key order

Source: user asked, "can we keep the order of the json keys? the json generated in this session has a different key order". This is a serialization regression correction within iteration 02, not a new product scope or approval gate.

Changed acceptance: generated persisted JSON and exported JSON retain the established relative order of existing DTO fields at every object level. New cost evidence stays adjacent to the related amount. Cost normalization preserves its input object order recursively, including nested maps; any newly inserted compatibility fields have deterministic placement. Invocation custom serialization must retain cost_usd before compactions/user_iterations rather than moving shadowed fields to the end.

Preserved guarantees: all approved cost states, null/zero/partial distinctions, legacy numeric observed policy, coverage, schema versions, validation, identities and array multiplicity remain unchanged. Superseded implementation detail: accidental alphabetical reordering through map-based JSON projection. The frozen semantic contract is unchanged.

Verification impact: extend actual repository/serializer boundary tests to check byte-level relative order and preserved meaning, normalization idempotency and nested objects. Run focused checks, independent correction review and required integrated make. Finally regenerate this ticket's real Codex ledger with the corrected built serializer, preserving all prior root/child sessions, and update the existing PR if still open.
