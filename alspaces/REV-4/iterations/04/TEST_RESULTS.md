# REV-4 iteration 04 independent test verification

## Tested scope and identity

Baseline: `e4fb51060afc40726153954872f400e6b36bdca3`. Tested the completed coder diff while the parent concurrently updated ticket documentation. Ownership was limited to this report; no tests were added or edited by the tester.

Implementation fingerprint: `51b92d4642c4ee8f016b637b18f918bbaca37461ae67f5c2b53bf44f2d9132ed`, SHA-256 of `git diff e4fb51060afc40726153954872f400e6b36bdca3 --` followed by the three authorized paths in lexicographic order:

- `backend/cmd/agentic-lens/release_verification_test.go`
- `backend/internal/domain/aggregations_test.go`
- `backend/internal/domain/enums_test.go`

Implementation delta: 11 lines added, 297 lines deleted, **286 net test lines removed**. Production, workflows, Makefile, frontend, manifests, schemas, runtime scripts and fixture data remain unchanged.

## Independent structural audit

Compared every baseline/current Go function with the Go standard library AST parser, identifying declarations by file/name and hashing exact source text. This handles multiline signatures and includes helpers/TestMain in preservation checks.

Exactly six top-level tests disappear, with no additions:

- `TestWorkflowFiles_Integrity`
- `TestMakefile_ReleaseTargetsListedInHelp`
- `TestMakefile_ReleaseVerifyFailsWhenChecksumMissing`
- `TestMakefile_ReleasePackageFailsWhenBinaryMissing`
- `TestDateFiltersAndAggregations`
- `TestOrderRoles_And_OrderSubagentRoles_AdversarialEdgeCases`

The release file is deleted in full. Its `findRepoRoot` helper is the only removed non-test Go function; reference search confirms it had no callers outside the deleted file and no remaining reference under backend. The only changed surviving function is `TestOrderRoles_And_OrderSubagentRoles_AdversarialComprehensive` in enums_test.go. Every other surviving Go test/helper/TestMain body is byte-for-byte unchanged, including `TestResolveDatePreset`, the other aggregation role tests, and the generic AgentRole ordering test.

Counts independently verified: **499 Go declarations** remain, down from 505. A fresh TypeScript compiler AST traversal confirms **698 TypeScript declarations**, with no web diff. Total current declarations: **1,197**. Parameterized runtime cases are counted separately from declarations.

## Role-table preservation and transfer

Parsed the existing comprehensive test's testCases composite literal and compared baseline/current row source text. Both tables contain **19 rows**: 16 are byte-for-byte unchanged; exactly three rows are refined. Replacing the table literal with a marker yields identical runner source hashes before/after, confirming that both public entrypoints, length checks and element comparisons remain unchanged.

| Refined contract | Current fixture and expected result |
|---|---|
| Partial canonical priority plus sorted remainder | Input reviewer/zeta/tester/coder/planner/orchestrator/alpha; priority planner/coder/tester/reviewer; expected planner/coder/tester/reviewer/alpha/orchestrator/zeta. The three-element out-of-order remainder preserves real alphabetical sorting coverage. |
| Complete canonical priority | Reversed canonical input reviewer/tester/coder/planner/orchestrator; full priority orchestrator/planner/coder/tester/reviewer; expected full priority order. |
| Canonical default alphabetical order | Reverse-alphabetical input tester/reviewer/planner/orchestrator/coder with no custom order; expected coder/orchestrator/planner/reviewer/tester. |

Nil, empty and whitespace-only input rows remain unchanged. Existing custom-role, duplicate/whitespace, nil/empty custom priority, unknown priority, duplicate priority, case distinctions and single-role rows remain. Production inspection confirms both public ordering functions delegate to the same helper; the actual tests still call both independently.

No top-level tests, table rows, new helpers, replacement release tests, or snapshots were introduced to conceal the reduction.

## Executed verification

Tester-owned command:

`cd backend && go test -race -json ./cmd/agentic-lens ./internal/domain > /tmp/rev4-iteration04-package-race.jsonl`

Exit status 0. Inspected JSON completion events: **1,139 test/subtest completions**, zero failed events and zero skipped events. Both package results passed: CLI 5.654s and domain 1.235s in this run. These elapsed values are local diagnostics, not a speedup claim.

The same JSON run directly verifies the targeted contracts without a redundant extra suite:

- Comprehensive role test: **38 passing child cases**, exactly 19 rows through each of OrderRoles and OrderSubagentRoles.
- All three transferred canonical configurations passed through both public functions.
- Nil, empty and whitespace-only inputs each passed through both entrypoints.
- `TestResolveDatePreset` and all three children passed: 24-hour endpoints, exact seven-day endpoints, and unknown-preset nil behavior. Its body remains unchanged.

`git diff --check` passed independently. Raw JSON and scratch AST-analysis scripts remain outside Git. The unchanged frontend's prior 749-case evidence is reused; no new frontend run is claimed by this report.

## Remaining parent gate and limitations

The parent owns full post-integration `make`, canonical embedded assets, telemetry, commits and PR update. This report does not claim that later gate has already passed. No mutation campaign, live release/publish exercise, browser layout check or controlled performance comparison was performed.

The deleted missing-input release checks exercised genuine narrow Makefile failure branches; removing them is user-approved coverage reduction, not evidence that CI validation improved. The role-ordering wrapper is consolidated into equivalent enum contracts rather than removing its useful behavior.

Status: PASS for exact scope, body preservation, role-table consolidation, remaining counts, affected-package race checks and targeted JSON cases.
