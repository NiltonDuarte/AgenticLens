# REV-4 iteration 05 plan

Read the [delta contract](CONTRACT.md), [47-entry candidate map](CANDIDATES.md), frozen root CONTRACT.md, prior iteration 04 review, current score catalog and TECHNICAL_GUIDELINES.md. Current explicit user authorization satisfies the gate for this bounded evidence-backed reduction. No production implementation is included.

## Ownership

Coder A owns only:

- backend/internal/domain/aggregations_test.go
- backend/internal/domain/enums_test.go
- backend/internal/domain/version_test.go
- web/src/api/client.test.ts
- web/src/domain/anomalies.test.ts
- web/src/utils/theme.test.ts

Among the role-ordering tests, only enums_test.go needs bounded recipient edits if the mixed-case default-order assertion from L09 is not already independently present; retained canonical/fallback tests stay unchanged. No new top-level declarations or rows solely to mirror deleted fixtures.

Coder B owns only:

- web/src/components/AnomalyBadge.test.tsx
- web/src/components/Components.test.tsx
- web/src/components/MultiSelectDropdown.test.tsx
- web/src/components/Navbar.test.tsx
- web/src/components/TicketTable.test.tsx
- web/src/components/TraceWaterfall.test.tsx
- web/src/views/FleetOverview.test.tsx
- web/src/views/MetricProgression.test.tsx
- web/src/views/TicketDeepDive.test.tsx

The parent owns all root ticket documents, numeric catalog and summary updates. Coders do not edit documentation or each other's test files. Preserve unrelated/shared edits. Tester writes iterations/05/TEST_RESULTS.md only and adds no tests; reviewer writes iterations/05/REVIEW.md only and otherwise audits read-only. Parent owns commits, main integration, full make, telemetry and PR delivery.

## Implementation

Apply CANDIDATES.md L01-L47 exactly. Retain L04, L15, L22, L34 and L46 as independent tests. The other 42 donors are deleted after any mandatory assertion transfer. Recipients may gain assertions and short rerender sequences, but retain their existing titles for unambiguous AST identity tracking. No new top-level tests, test.each expansion, helper test wrappers or snapshot suites. Remove only newly unused imports, test-local setup and empty describe groups.

Coder A: merge exact parser string-dispatch assertions into TestEnumsParsing, remove redundant role wrappers only after checking enum coverage including the mixed-case default order, drop the Go value-semantics snapshot overconstraint, remove duplicate mocked 404 coverage, retain positive compiler narrowing inside the four-member production guard test, and fold canonical model repeatability into the existing stronger model-sample loop. Keep both enum data guards and all-token color guard unchanged.

Coder B: use named existing fixtures/recipients for badge props, tooltip optional states, keyboard negatives, table-header accessibility, metadata rerenders and view help wiring. Avoid copying mounts or creating a universal tooltip helper merely to reduce visible lines. Reuse existing models/roles in fixtures; preserve behavioral distinctions rather than spelling every old test literal again. Keep ThreadViewer empty collection coverage and the substantial view-owned autopsy tooltip test. Merge the two CSS-source guard bodies unchanged into the existing responsive-wrapper contract, reading CSS once; do not claim real geometry validation. No change to actual styles or components.

For rerender flows, query current elements after updates as appropriate; preserve all original callback, prop precedence, absence, counts and metadata assertions. In Navbar project context, explicit name, configured name and explicit empty override remain distinct; in version metadata, version/commit/date all change on the update. In AnomalyBadge, preserve visible custom label independently of accessible name, title absence/default/override, and sizes. In TrendChart, preserve supplied/omitted/empty tooltip for populated/empty data plus existing keyboard interaction.

## Verification

Coders run affected focused checks and coordinate to avoid duplicate full suites. Coder A owns focused Go domain checks and TypeScript typecheck/lint after both patches settle; Coder B owns focused changed frontend files unless the tester owns the full JSON frontend run. No new domain mocks, skips, weak assertions, fixture replacement or production fixes.

Tester independently parses baseline/current Go and TypeScript ASTs: exactly five Go and 37 TS donor removals if all 42 are implemented, no additions, and no unrelated changes. Track permitted receiver changes separately from untouched test bodies. Confirm five retained donors remain without score inflation. Audit each MERGE's mandatory oracle in its named recipient, especially narrowing under tsc, all CSS regex guards, metadata and explicit-empty precedence, tooltip presence/absence matrices, and both role entrypoints. Verify zero silent skips and actual frontend count, not just process exit. Record tested diff fingerprint, commands and limits. Existing negative @ts-expect-error directives remain effective.

Reviewer audits every donor/recipient transfer along scope alignment and engineering standards, confirming meaningful protection survives and reduction is not merely hiding independent failures in unrelated large tests. Reuse valid test results on unchanged diff. Give explicit APPROVED or CHANGES_REQUESTED and report any candidate that needs retention instead.

Parent updates the current score catalog to actual remaining AST identities, sorted ascending; preserve scores/confidence unless a clearly recorded content change warrants reconsideration, never raise five retained scores to clear the threshold. Record each executed donor outcome and any exception in the candidate audit. Commit reviewed paths, merge origin/main, run required full make and inspect all stages, then generate Codex telemetry at the final preparation step with literal nonce and type test. Preserve original root/child sessions and validate newly spawned sessions from real metadata. Push/update existing PR if still open, link immediately and hand off correct URL with behavior/verification/limitations.

## Completion

All 42 selected donors removed with the seven DELETE and 35 MERGE obligations verified; five independent guards retained at unchanged scores. Tester and reviewer completed; reviewer APPROVED. Implementation committed as 3ab3ece; main already up to date. Full post-integration make passed, including all 712 frontend cases and four target builds. Catalog: 1,155 sorted entries validated. Final telemetry and PR delivery follow.

## Routine main integration follow-up

Source: user requested "merge main into branch". Initial tree was clean at `13d6b57bef1964a245251a9b8c03d489cd801ce4`. Fetched origin/main `575bb78`; merged with ort and no conflicts as `ca1fec8b8a681c813a23c0cc162850a96b15a5c4`. This is routine integration within iteration 05, not a new acceptance scope.

Upstream changes are contributor/technical guidance, regression-prevention documentation and REV-3 historical artifacts. No backend, frontend or Makefile changes came from the merge, so the accepted test reductions, five retained guards and source/catalog identities remain unchanged. No prior approval or behavioral guarantee is superseded.

Reviewed the new docs/REGRESSION_PREVENTION.md checklist and owning guidance. Applicable integration cases: real full make stages with frontend dependencies/assets present, current embed tree and standalone binary, preserving task-owned source, and final Codex root/child provenance. New boundary implementation and absent/repeated build-path experiments are not applicable: this merge changes no build/CI/runtime paths. Full post-merge verification results are recorded in SUMMARY.md.
