# REV-4 iteration 04 plan

## Authorized scope and rationale

The latest user request approves exactly this follow-up. Baseline `e4fb51060afc40726153954872f400e6b36bdca3` is clean; PR #72 is open. Read the [delta contract](CONTRACT.md), root CONTRACT.md, SCORE_REVIEW.md, TEST_SCORES.md, TECHNICAL_GUIDELINES.md, and prior iteration evidence. No additional approval question is needed.

## Coder ownership

Only these existing test paths are writable by the coder:

- `backend/cmd/agentic-lens/release_verification_test.go`
- `backend/internal/domain/aggregations_test.go`
- `backend/internal/domain/enums_test.go`

Delete the release file completely. It contains TestWorkflowFiles_Integrity, TestMakefile_ReleaseTargetsListedInHelp, TestMakefile_ReleaseVerifyFailsWhenChecksumMissing, TestMakefile_ReleasePackageFailsWhenBinaryMissing, and findRepoRoot. Grounded reference search found findRepoRoot only within this file; confirm before deletion. Two cases check workflow/help text; the other two execute genuine missing-input Makefile failures but do not validate successful release/publish behavior. User explicitly accepts removing this entire low-yield file. Do not change the actual workflows or Makefile or add substitute CI tests.

Remove TestDateFiltersAndAggregations from aggregations_test.go. It only checks a seven-day duration, while TestResolveDatePreset in the same file independently asserts exact seven-day boundaries, 24-hour boundaries and unknown inputs. Preserve TestResolveDatePreset unchanged.

Remove TestOrderRoles_And_OrderSubagentRoles_AdversarialEdgeCases from aggregations_test.go after the case transfer below. Leave the other role-order tests and all unrelated aggregation functions unchanged; this is not permission for a broader cleanup.

## Role case-transfer map

Both public ordering functions delegate to the same orderRolesHelper in enums.go. The existing TestOrderRoles_And_OrderSubagentRoles_AdversarialComprehensive table in enums_test.go already runs every row through both entrypoints, checks lengths, and compares every element. Use that existing runner; do not add a wrapper or new top-level test.

| Removed wrapper contract | Existing enum protection or bounded transfer |
|---|---|
| Nil, empty and whitespace-only inputs for both functions | Already covered by the first three comprehensive table rows, with length checks for both entrypoints. No new row. |
| Custom role names sorted alphabetically and deduplicated | Already covered by the comprehensive duplicate/whitespace/default-order rows and the dedicated subagent ordering table. Retain those rows; changing zeta/beta names adds no new branch. |
| All canonical roles follow a complete custom priority | Refine the existing comprehensive row for customOrder matching all input roles in reverse order to use reversed canonical input and canonical custom priority. Preserve its full-priority contract and both entrypoint assertions. |
| Canonical subagent priority excludes orchestrator from the priority list and leaves it in the alphabetical remainder | Refine the existing comprehensive priority-first/alphabetical-remainder row to use the canonical input plus out-of-order custom remainder names (for example zeta and alpha), planner/coder/tester/reviewer priority and the alphabetical alpha/orchestrator/zeta remainder. This preserves the old multi-element remainder-sorting assertion while transferring the orchestrator case. Both entrypoints still run this row. |
| Canonical roles default to alphabetical ordering | Refine the existing reverse-sorted-input row to use canonical strings in reverse alphabetical order, expecting coder/orchestrator/planner/reviewer/tester. This retains the existing reverse-sort condition and transfers the canonical-name example without adding a row. |

Adjust the three refined row names to accurately describe their fixtures. Preserve generic/custom-name coverage in the other existing rows, including nil/empty custom order, trimming, duplicate order entries, unknown priorities, case distinctions and single roles. The existing generic AgentRole test stays unchanged. The comprehensive table's row count must not increase.

## Verification responsibilities

Coder: run gofmt only on edited surviving Go tests, focused enum/date tests, and diff whitespace checks. Return exact changes; no commits or publication.

Tester: inspect the baseline/current Go AST diff, confirming exactly six removed top-level declarations and no new top-level declarations. Only the existing comprehensive enum test body may change beyond those removals. Confirm findRepoRoot has no outside references, table row count is unchanged, all three transfer configurations run through both entrypoints, and all unrelated bodies remain unchanged. Run affected packages with the race detector and targeted verbose/JSON role/date tests; report failures and skip counts. Reuse valid unchanged frontend evidence; parent owns full post-integration make. Do not write replacement tests.

Reviewer: verify scope, actual transferred contracts, generic-role coverage, accurate names and the absence of new rows/tests. Review removal rationale without claiming the real missing-input checks were entirely fake. Write verdict in iterations/04/REVIEW.md; do not modify code or publish.

Parent: record each removal and transfer, update score/catalog references as historical where deleted or shifted, preserve prior numeric findings, and report current 499 Go + 698 TypeScript = 1,197 declarations. Commit reviewed work, merge origin/main, run full make, record actual results, generate final Codex telemetry last with conventional type test while preserving earlier root/children, validate sessions, commit/push, update the existing PR if still open, and link immediately. Source/test fixture references must not depend on this ticket workspace.

## Completion evidence

Coder, tester and reviewer completed; reviewer APPROVED. Implementation committed as a10c61c; main already up to date. Full post-integration make passed including all 749 frontend cases and four target builds. Final telemetry and PR update follow.
