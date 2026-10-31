# REV-4 iteration 02 test results

## Tested scope and identity

Baseline: `833ceebea739a74b53c1366c6d350743d3c1959b`. Tested the eight-path working test diff after the coder completed the approved removals; documentation was edited concurrently by the parent and excluded from the implementation fingerprint.

Test-diff SHA-256: `8fd88fd904afbb84e83421f7f9dec1d9363770c3c103c37460370795f3e06d5f` (SHA-256 of `git diff 833ceebea739a74b53c1366c6d350743d3c1959b --` followed by the eight authorized paths in lexicographic order).

## Independent structural verification

Compared baseline and working tree test declarations using the Go standard library AST parser and installed TypeScript compiler AST. The Go parser handles multiline signatures, including the retained `TestFilterLedgersByCompletionDate`. TypeScript parsing handles plain tests and parameterized declarations. Comparison uses multisets of file, declaration name, and SHA-256 of the exact source text for each function/test call, preserving duplicate names rather than collapsing them.

| Language | Baseline declarations | Remaining declarations | Removed | Added or changed surviving bodies |
|---|---:|---:|---:|---:|
| Go | 511 | 505 | 6 | 0 |
| TypeScript | 742 | 698 | 44 | 0 |
| Total | 1,253 | 1,203 | 50 | 0 |

- The removed declaration set equals the authoritative R01-R50 ledger from the committed baseline REVIEW.md exactly. No unrelated removals or new declarations.
- Every surviving Go test function and TypeScript test call remains byte-for-byte identical. Every Go test helper and TestMain function also remains byte-for-byte identical.
- All conditional consolidation/rework candidates remain. The negative compiler assertions in anomalies.test.ts and the retained multiline completion-date test remain unchanged.
- The only non-ticket-documentation changed paths are the eight authorized test paths. No production code, manifests, runtime scripts, fixtures, or API contracts changed.
- Two approved fixture-only files are deleted completely: domain/live.test.ts and domain/migration.test.ts. Remaining shared setup is retained; removed setup/imports are local to removed tests. Inspection of the TypeScript diff confirmed the copied comparator/mock event array and empty describe groups disappear with their approved tests.
- Implementation line delta: 1 added import-replacement line, 1,611 deleted test lines, **1,610 net test lines removed**. No replacement tests or helpers were added.

## Executed verification

Tester-owned command:

`cd web && npm test -- --reporter=json --outputFile=/tmp/rev4-iteration02-vitest.json`

Exit status 0. Independently inspected the JSON rather than relying on a successful process exit: **749 passed / 749 total cases**, 0 failures, 0 pending, 0 todo, all assertion statuses `passed`, across **34 files**. The 793-case original suite lost exactly 44 plain runtime cases; parameterized surviving cases remain. This is an execution-health result, not a mutation-equivalence claim or controlled timing comparison.

Coder-owned focused checks, coordinated to avoid duplication and reported directly by the coder:

- `cd backend && go test -race ./cmd/agentic-lens ./internal/domain`: passed; CLI 7.599s, domain 1.112s. Package output did not report skips; no separate Go JSON skip audit is claimed here.
- `cd web && npm run typecheck && npm run lint`: passed, including existing negative compiler directives and unused-import checks.
- `git diff --check`: passed; independently rerun by the tester after the structural audit.

## Scope of evidence and remaining gate

No new tests were written because this is a precisely bounded subtraction. No tests or assertions were weakened to match the expected totals. The completed frontend run covers the retained live API parser/stream behavior, actual event projection, component interactions, and view-level hydration scenarios. Focused Go race results are reused for unchanged code.

The parent owns the required full post-integration `make` gate, canonical embedded asset verification, final telemetry, and PR delivery. Those results belong in SUMMARY.md and the final reviewer report; this report does not claim that a future integration run already passed. No mutation campaign or browser layout validation was run, and no coverage-equivalence or speedup claim is made.

Status: PASS for exact approved subtraction, surviving-body preservation, frontend runtime verification, and coordinated focused checks. Required parent integration gate remains separate.
