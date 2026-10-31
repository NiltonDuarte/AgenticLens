# AL-57 iteration 02: Independent implementation audit

Status: APPROVED. All three bounded review findings were corrected and rechecked. This is an implementation audit verdict, separate from user scope approval and the parent integration/delivery gate.

## Reviewed boundary and evidence

Baseline: 45cb679f21475cd14b7dd4cbc25996a39d734403, branch fix/AL-57-cost-semantics. Reviewed the working-tree diff against HEAD, including untracked application files, approved CONTRACT.md and iterations/02/PLAN.md, TECHNICAL_GUIDELINES.md and applicable regression checklist. The current user amendment governs: valid legacy numeric values including historical Codex zero are observed; there is no special historical type, count or display. Iteration 01 is historical, not the accepted cost policy.

Reviewed backend/web source SHA256: 82d6d7261258a2b33dd8350b64a5215185c2e24929a5f5301e0f25d9f6ac5c81. Method: sorted unique paths from git ls-files -co --exclude-standard -- backend web, then hash each path, NUL, file bytes, NUL. Ignored dependencies and generated asset trees are excluded. This fingerprint includes the precision ordering, task rendering, registration error-handling fixes and their regression tests. Further source changes require affected follow-up verification.

## Findings

1. P2, FIXED: web/src/components/SkillsSubagentsPanel.tsx task cost rendering previously called formatCurrencyUSD(task.cost_usd) and ignored admitted partial task evidence. A valid v5 task with cost_usd null and cost_evidence state partial / known_subtotal_usd 2 is accepted by the shared boundary validator and serialized correctly, but displays only Unavailable, losing its known contribution. The new shared formatFactCostUSD renders the known partial subtotal without invocation coverage claims. The tester verified the actual deep-dive task row and separately proved DeriveMetricsLedger excludes auxiliary task cost from invocation totals. Frontend focused 139 tests/typecheck/lint passed; tester independently passed all 64 deep-dive tests and the actual domain exclusion case under race.
2. P2, FIXED: backend/internal/domain/migration/engine.go, NewEngine initially discarded the newly added Register(NewStepV4ToV5CostEvidence()) error with `_ =`. The canonical rule prohibits discarding domain-execution errors. NewEngine now explicitly checks all four static built-in registrations and raises a contextual invariant panic for an impossible programmer-error failure, preserving its public constructor signature and caller-supplied registry behavior. Coder migration/backfill tests and targeted vet passed; the corrected constructor was re-read during final audit.
3. P2, FIXED: backend/internal/domain/aggregations.go previously checked complete evidence but sorted model/thread/ledger costs by rounded internal CostUSD. Distinct source amounts such as 0.00001 and 0.00002 could therefore sort by identity/date despite differing exported amounts. Equal ledger costs also used UpdatedAt rather than the accepted stable identity tie. The corrected comparisons use CompleteAmount and stable identity in both directions. Coder focused tests/vet passed; tester added precision/tie regression cases and reports domain/HTTP race checks passing.

## Axis 1: Contract alignment

- Traced current Codex and both OpenCode source families, coalesced present/missing evidence, presence-aware v4 runtime projection, v1-v3 explicit migration, v4-v5 conversion, persistence, exact-owner backfill entrypoints, pure rollups, HTTP/CLI/Markdown and frontend consumers.
- Reviewed strict v5 contradiction checks and v4 numeric/null/omitted handling. Numeric historical costs remain ordinary observed values, including zero. Newly missing current source costs remain unavailable; partial coalesced contributions retain a known subtotal. There is no harness-based historical rewrite.
- Reviewed cost summary composition/counts across role/model/fleet/time projections, unknown complete totals, empty states, incomplete cost denominators, cumulative availability, and task-reference exclusion from invocation totals.
- Reviewed real HTTP mixed-v4/v5 fixture and unchanged-file assertions, flattened v1-v3 migration presence/multiplicity cases, both SQLite source-family zero/partial/null/negative cases, actual API client rejection tests, DOM subtotal/share/sort tests, and chart gap/non-cost scatter cases.
- Confirmed tester corrections for unavailable fleet-map coverage and accepted cost_usd sort alias. No meaningful tests were removed; changed expected loading/zero-denominator behavior follows the approved contract. Existing code precision behavior is now checked by the focused follow-up regressions.

## Axis 2: Engineering standards

New financial algebra remains pure with typed Go evidence/summary enums and strict TypeScript evidence unions. Boundary adapters own JSON validation/projection; source/database/filesystem effects do not enter the functional cost core. Changed application additions were scanned for forbidden any/interface{} and emoji symbols; no new occurrences found. git diff --check passed. The previously ignored registration error is now handled explicitly; no open engineering-standard finding remains.

Compatibility and API changes are documented in docs/COST_SEMANTICS.md, linked from README and setup guidance. Canonical v4 stays readable without GET writes; old binaries rejecting v5 is disclosed. Estimates, rate cards, AL2 dispatch and live-history bulk mutation are outside this change.

## Verification reuse and limits

Reused the tester's recorded make test evidence: all 12 Go packages under the race detector and 754 frontend tests passed without skips before the precision follow-up. The tester records later affected precision checks separately. Initial coder Go tests/vet and frontend typecheck/lint/build also passed. These valid results were not duplicated merely for review. Final task-row rendering and auxiliary-cost-exclusion follow-up cases also passed; the parent integrated make gate will cover the combined final source.

No browser rendering or geometry validation is claimed. Current frontend source changes require rebuilding ignored web/dist and backend/internal/adapters/web/dist after their final fix, which the parent integrated gate owns. Parent still owns current-main integration, required full make, standalone binary/embedded asset smoke checks, final source-grounded telemetry, clean-tree commits and PR delivery. No findings remain open; the earlier complete suite and later affected checks are recorded distinctly.

Final verdict: APPROVED.
