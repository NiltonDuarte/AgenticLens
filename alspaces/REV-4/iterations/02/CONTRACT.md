# REV-4 iteration 02 contract delta: delete approved candidates

## Baseline and approval

Implementation baseline: `833ceebea739a74b53c1366c6d350743d3c1959b` on `REV-4/reduce-low-value-tests`; working tree was clean when this iteration began. Existing PR: https://github.com/NiltonDuarte/AgenticLens/pull/72 (open at grounding).

The user explicitly selected “Review and delete approved candidates” in response to the review-versus-deletion question. That selection approves exactly R01–R50 in the committed [review](../../REVIEW.md). Per repository AGENTS.md, this explicit scope approval satisfies the development gate. Do not request the same approval again. The frozen [root contract](../../CONTRACT.md) describes iteration 01 and remains unchanged; this delta authorizes the subsequent implementation.

## Accepted changes

Remove the 50 named test declarations, including their now-unreferenced test-local setup, empty describe wrappers, and unused imports. Remove `web/src/domain/live.test.ts` and `web/src/domain/migration.test.ts` entirely because all their 30 declarations are approved for removal. Preserve shared setup referenced by remaining tests. No production behavior, production files, manifests, schemas, runtime scripts, fixture contents, source data, or API contracts may change.

Retain every conditional CONSOLIDATE, REWORK, KEEP TYPE CONTRACT, and KEEP entry. No extra consolidation or test deletion is authorized. No replacement tests, new test helpers, snapshots, or parameterization that obscures the actual count reduction. Preserve TypeScript negative compiler checks and all independently useful boundary/regression coverage identified in the review.

## Acceptance

- All and only R01–R50 are absent; all other test bodies and assertions remain unchanged apart from necessary unused test imports/setup and empty wrapper cleanup.
- Expected declaration reduction: six Go tests and 44 TypeScript tests. Corrected baseline: 511 Go plus 742 TypeScript declarations, 1,253 total. Expected final: 505 Go plus 698 TypeScript declarations, 1,203 total. Iteration 01 missed the multiline `TestFilterLedgersByCompletionDate` declaration; do not delete or change it while correcting inventory documentation.
- Expected Vitest runtime cases: 793 minus 44, or 749, unless evidence demonstrates parameter expansion differences. Count actual results and explain any mismatch; do not weaken tests to fit arithmetic.
- The root review explains the reason for each executed removal; the inventory preserves historical source references and distinguishes removed from retained entries. Parent owns those documentation corrections.
- Focused affected suites and required post-integration `make` pass, with no silent skips, production changes, or stale tracked generated assets. Keep raw test output outside Git and record exact results and remaining limitations.
- No claim of mutation-equivalent coverage or controlled runtime improvement without corresponding evidence. This change reduces demonstrated noise and duplicate coverage.

## Role and ownership boundaries

The coder owns only the eight test paths enumerated in PLAN.md. The tester may run checks and report failures, but adds no tests for this subtraction-only change. The reviewer audits read-only application/test diffs and owns `iterations/02/REVIEW.md`. The parent owns integration, ticket documentation, telemetry, commits, and PR delivery. All agents share this checkout and must preserve others' changes.
