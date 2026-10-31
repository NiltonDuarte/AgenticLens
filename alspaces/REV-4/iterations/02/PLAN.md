# REV-4 iteration 02 plan: remove 50 approved low-value tests

## Scope and existing approval

Read the [iteration delta contract](CONTRACT.md), [root review](../../REVIEW.md), [baseline inventory](../../TEST_INVENTORY.md), and [iteration 01 verification](../01/REVIEW.md). The committed R01–R50 ledger is authoritative; temporary scripts are aids only. Baseline is `833ceebea739a74b53c1366c6d350743d3c1959b` and the initial checkout was clean.

The user's explicit “Review and delete approved candidates” selection authorizes these exact removals and satisfies the gate under AGENTS.md. No additional question is needed. Conditional recommendations remain outside this iteration.

## Coder writable scope

The coder owns only these eight existing test paths. The two all-candidate files may be deleted; all other paths receive named test removals and necessary test-local cleanup only. Preserve edits made by others in the shared checkout.

### `backend/cmd/agentic-lens/release_verification_test.go`

| ID | Approved declaration | Historical source line |
|---|---|---|
| R01 | `TestReleaseStable_SemVerRegexValidation` | 15 |
| R02 | `TestReleaseCandidate_VersionFormatRegex` | 70 |
| R03 | `TestReleaseCandidate_BaseTagDerivation` | 114 |
| R04 | `TestChecksumVerification_DetectsCorruptedArchives` | 299 |

### `backend/internal/domain/config_test.go`

| ID | Approved declaration | Historical source line |
|---|---|---|
| R05 | `TestWorkspaceArtifactConfig` | 244 |
| R06 | `TestProjectConfig_HelperFixture` | 264 |

### `web/src/components/Components.test.tsx`

| ID | Approved declaration | Historical source line |
|---|---|---|
| R07 | `dispatches onSelectTab when tab is clicked` | 368 |
| R08 | `renders empty placeholder when data array is empty` | 379 |
| R09 | `renders bar series rect elements` | 414 |
| R10 | `renders markdown artifact content and toggles between contract and summary` | 967 |
| R11 | `renders empty notice when content is null` | 986 |

### `web/src/domain/events.test.ts`

| ID | Approved declaration | Historical source line |
|---|---|---|
| R12 | `correctly discriminates event_type enums` | 96 |
| R13 | `preserves tool call error excerpt and execution duration` | 100 |
| R14 | `preserves compaction events with turn and context volume` | 109 |
| R15 | `verifies chronological ascending ordering and turn tie-breaking logic` | 118 |

### `web/src/domain/formatters.test.ts`

| ID | Approved declaration | Historical source line |
|---|---|---|
| R16 | `accepts arbitrary role strings under AgentRole` | 1011 |
| R17 | `accepts all 4 valid severity values` | 1026 |
| R18 | `accepts all 4 health status values` | 1035 |
| R19 | `supports canonical conventional commit types and custom ticket types` | 1046 |

### `web/src/domain/live.test.ts`

Delete this complete file: every declaration is approved.

| ID | Approved declaration | Historical source line |
|---|---|---|
| R20 | `deserializes alert across all severity tiers and canonical rule names` | 16 |
| R21 | `deserializes alert with canonical live anomaly rule` | 47 |
| R22 | `handles hostile trigger_event strings including path traversal and SQL injection patterns` | 62 |
| R23 | `deserializes complete tool execution part with input and output snippets` | 89 |
| R24 | `deserializes skill and task delegation attributes while preserving legacy omission` | 114 |
| R25 | `deserializes failing tool part with error message and high severity` | 129 |
| R26 | `handles optional fields omitted or null gracefully` | 149 |
| R27 | `deserializes failing tool part where error_message is null` | 167 |
| R28 | `deserializes varied non-tool part types such as text, step_finish, and compaction` | 187 |
| R29 | `deserializes message fact with full token breakdown and reasoning content` | 207 |
| R30 | `deserializes message fact when reasoning_content is omitted or null` | 249 |
| R31 | `deserializes message fact with zero tokens across all token fields without failure` | 287 |
| R32 | `deserializes message fact with saturated context window exceeding healthy ceiling` | 311 |
| R33 | `deserializes message fact containing multiple mixed tool and non-tool parts` | 333 |
| R34 | `deserializes root session node with null or omitted parent_id` | 386 |
| R35 | `deserializes child sub-session node with populated parent_id and attached anomalies` | 430 |
| R36 | `deserializes deeply nested 3-tier session hierarchy (root -> child -> grandchild)` | 463 |
| R37 | `deserializes nested multi-level tree hierarchy and aggregated active anomalies` | 532 |
| R38 | `deserializes empty session tree safely` | 615 |
| R39 | `deserializes session tree with multiple concurrent root sessions and distinct ticket formats` | 629 |
| R40 | `deserializes trace view with message history, tool statistics, and anomalies` | 677 |
| R41 | `deserializes empty trace view safely` | 734 |
| R42 | `deserializes trace with 100% tool failure rate and multiple recent errors` | 753 |

### `web/src/domain/migration.test.ts`

Delete this complete file: every declaration is approved.

| ID | Approved declaration | Historical source line |
|---|---|---|
| R43 | `deserializes raw historical v1 ledger without schema_version or events` | 5 |
| R44 | `accurately parses version 2 ledger with migration metadata and empty events array` | 69 |
| R45 | `accurately handles partial migration status representation` | 110 |
| R46 | `accurately parses version 4 ledger with atomic user_iterations on invocations and migration metadata` | 126 |
| R47 | `handles legacy ledgers where user_iterations is undefined without runtime error` | 206 |
| R48 | `handles empty code facts defensively` | 238 |
| R49 | `preserves populated chronological events during schema upgrades` | 252 |

### `web/src/utils/theme.test.ts`

| ID | Approved declaration | Historical source line |
|---|---|---|
| R50 | `is strictly deterministic across 50 repetitions` | 101 |

## Implementation sequence

1. Identify declaration boundaries by parsing Go/TypeScript structure, not by deleting between adjacent title matches. Preserve neighboring helper definitions, comments belonging to survivors, and shared setup. Use Go parsing that recognizes multiline signatures.
2. Remove six approved Go functions from the two Go paths and 44 approved TypeScript declarations from the six web paths, including complete deletion of the two all-candidate files. Remove test-local copied comparators, fixture arrays, and empty describe groups made unused by these removals.
3. Remove only imports that become unused. Expected candidates include `crypto/sha256` and `regexp` in the release verification file, EventType/mockEvents-related setup in events tests, literal-only type imports in formatter tests, and ArtifactViewer in the broad component suite. Confirm actual references before removing each; `fmt`, other JSON/contracts, testProjectConfig, and helpers may still be needed.
4. Do not change `backend/internal/domain/tool_execution_test.go`, `aggregations_test.go`, `web/src/domain/anomalies.test.ts`, or `web/src/views/TicketDeepDive.test.tsx` conditional cases. Keep the broad TrendChart reference-line case, Navbar workspace/loading assertions, and compiler-negative anomaly contracts unchanged.
5. Return exact removed declarations and per-file line deltas to the parent. Make no commits or PR updates.

## Tester verification

- Compare parsed declaration sets before and after using Go AST or multiline-aware matching and TypeScript compiler AST. Require the difference to equal the committed 50-entry ledger exactly; detect additions and unrelated removals. Retain `TestFilterLedgersByCompletionDate`, which was missing only from the old inventory.
- Verify expected corrected declaration totals: 511 to 505 Go, 742 to 698 TypeScript, 1,253 to 1,203 overall. Record actual expanded Go subtest and frontend runtime case counts separately; Vitest should report 749 cases after deleting 44 plain declarations.
- Run affected backend packages (`./cmd/agentic-lens` and `./internal/domain`) with the race detector, and affected surviving frontend files plus dedicated Navbar, ArtifactViewer, TrendChart, live API, progression/events and negative compiler contracts as relevant. Typecheck/lint detect import cleanup mistakes. Existing frontend dependencies/assets should be reused unless absent.
- Avoid redundant verification when the parent-owned full `make` already covers the same unchanged diff. Coordinate command ownership and preserve valid results. No new test additions are needed for this pure subtraction.
- Report the tested commit or exact diff, commands, pass/fail/skip counts, and environment limitations. Any failure caused by unrelated code requires reporting, not expanding scope or silently skipping.

## Reviewer verification

Audit contract alignment and engineering standards against the exact removed set. Confirm no useful surviving assertion was deleted incidentally, no shared fixture remains orphaned unnecessarily, no production/manifest/runtime changes occurred, and all conditional candidates remain. Re-read the five duplicate component cases against dedicated surviving coverage. Verify source references remain explicitly historical after line shifts, and review the parent-updated removal rationale/status rather than interpreting recommendations as completed work without evidence. Write verdict and evidence in `iterations/02/REVIEW.md` only.

## Parent integration and delivery

Update root REVIEW.md to mark R01–R50 as removed and append actual verification, without erasing original rationale. Correct TEST_INVENTORY.md's missed multiline Go declaration and mark exactly 50 historical entries REMOVED; describe retained baseline references and final totals. Update SUMMARY.md and rewrite PULL_REQUEST.md around actual reductions (conventional `test` classification); preserve the frozen root contract and iteration 01 history.

Commit reviewed implementation and artifacts, merge `origin/main`, run the required full `make`, and verify the canonical embedded asset tree without adding ignored outputs. Check PR state before pushing: update PR72 if still open; if merged, compare against current main and create a new PR for remaining work. Generate final telemetry only after planner/coder/tester/reviewer and integration work finish, using source Codex, a literal unique nonce, conventional type `test`, and the existing ledger so earlier sessions remain preserved. Verify expected root/children and existing cross-harness sessions before committing. Parent alone commits, pushes, links the PR, and hands off.

## Progress

Planning and explicit approval recorded. Coder completed exactly 50 removals and necessary test-local cleanup. Tester independently verified exact declaration/body preservation and all 749 frontend cases. Reviewer APPROVED with no blocking findings. Implementation committed as 2a78bad; main already up to date. Parent post-integration make passed in full, including749 frontend cases and four platform builds; embedded asset bytes verified in the standalone binary. Final telemetry and PR update follow.
