# REV-4 scored test review: iteration 03

Iteration 03 originally scored 1,203 declarations. After the earlier targeted deletions and iteration 05 low-score audit, the [current catalog](TEST_SCORES.md) has **1,155 declarations** (494 Go and 661 TypeScript), with 712 expanded frontend cases. Numeric ratings remain unchanged for surviving declarations.

Iteration 05 examined all 47 candidates scoring at most four: **7 deletions, 35 assertion-preserving consolidations, and 5 retained independent guards**. A low score alone was not a proof of redundancy. The five residual scores remain unchanged; they were not promoted merely to clear a threshold. [CANDIDATES.md](iterations/05/CANDIDATES.md) records every candidate and its named protection. Earlier findings below are historical unless labeled with a later iteration; the current catalog is authoritative for membership.

Baseline: `2e87ab15ffee382b1685bf38f233eca3d76fe72e`; initial working tree clean. The user requested numeric scores rather than binary dispositions. Iteration 01/02 documents remain intact as historical removal evidence; this is the iteration 03 scoring rubric and findings, with a current-catalog update below.

## Tests sorted by score

The current entries are ordered from lowest to highest score, with source path and line breaking ties. Numeric ratings for survivors are unchanged.

- 0/10 — no tests
- 1/10 — no tests
- [2/10 — 1 tests](TEST_SCORES.md#score-210)
- [3/10 — 1 tests](TEST_SCORES.md#score-310)
- [4/10 — 3 tests](TEST_SCORES.md#score-410)
- [5/10 — 65 tests](TEST_SCORES.md#score-510)
- [6/10 — 307 tests](TEST_SCORES.md#score-610)
- [7/10 — 521 tests](TEST_SCORES.md#score-710)
- [8/10 — 230 tests](TEST_SCORES.md#score-810)
- [9/10 — 16 tests](TEST_SCORES.md#score-910)
- [10/10 — 11 tests](TEST_SCORES.md#score-1010)

## What the number means

The score is the current declaration's **marginal regression-protection value in this repository**, considering the rest of the suite. It is an ordinal engineering judgment, not a measured probability, code-coverage percentage, or mutation score. A small cheap test can score highly; a large slow integration test can also be essential. Passing tests do not establish usefulness.

| Score | Interpretation |
|---:|---|
| 0 | Definitely expendable: no additional regression oracle beyond a specifically named surviving test. Its counterpart must remain. |
| 1 | Almost no application signal; local assignments, runtime/library facts, or assertions that miss the claimed contract. |
| 2 | Very weak marginal protection: constants, aliases, descriptive output, or a much weaker overlapping example. |
| 3 | Low priority: brittle source-shape checks, shallow presentation, or substantial overlap with a condition before consolidation. |
| 4 | Small real contract: optional presentation or a narrow dispatch branch; a compact representative matrix may suffice. |
| 5 | Useful bounded utility/prop behavior; worthwhile but not independently critical. |
| 6 | Meaningful feature behavior, rendering, or route contract; no demonstrated safe redundancy. |
| 7 | Important calculation, query, interaction, configuration, or boundary behavior. |
| 8 | High-value error recovery, provenance, compiler rejection, isolation, or persistence/source boundary. |
| 9 | Very important adversarial integrity or integration coverage with a concrete failure mode; avoid losing the contract in consolidation. |
| 10 | Critical protection against corruption, destructive migration, invented ownership, inherited usage, lost committed facts, or cross-ticket evidence. This contract must survive. |

No current test is assigned 1 merely to populate every integer. A 10 protects a contract, not necessarily its exact existing wrapper forever: a refactor needs demonstrated equivalent protection before replacing it. A 3 is not authorization to drop the whole test without its unique case or compiler assertion. A score must not be inferred from whether a test was previously labeled a conditional consolidation candidate.

## Basis and limits

All declarations were extracted from current source using Go AST parsing (including multiline function signatures) and the installed TypeScript compiler AST (including multiline it.each). Each catalog row records current line, exact title, score, reason, and observed calls/assertions. Duplicate titles are distinguished by file/line; parameterized tests are scored as the combined declaration, not independently certified table rows.

**Focused** rows received direct body/peer or critical-contract inspection, including the three zero scores, the selected ten scores, prior conditional clusters, and calibration of misleading cosmetic keywords. **Screened** rows are provisional source-assisted ratings using declaration bodies, target calls, assertions and contract families. They have not each received exhaustive pairwise overlap analysis. Source-signal rules accelerated broad scoring; the shortlist and critical examples were manually calibrated against actual tests and implementation. This distinction prevents a blanket score from pretending to prove uniqueness.

In particular, tooltip/badge wording must not hide zero-denominator, event-propagation, malformed-time, or persisted-versus-hydrated provenance contracts. Negative compiler tests are not fixture-only runtime checks. Helpers containing assertions count: testCodexRefresh and assertFencedVerbatim were inspected, rather than rating their callers as assertion-free.

No mutation testing or temporary deletion experiment was performed. The zero-score evidence is a static assertion/behavior containment argument; it is not a claim that every adversarial implementation mutation was tried. Scores 2-4 identify the next review work, not an automatic mass-deletion threshold. Do not delete both a zero-scored case and the representative named in its explanation.

Further calibration distinguished the small optional-collection accessor tests (5-6) from actual event reconciliation (8-9), and inspected parity fixtures: explicit externally recorded reference totals are independent oracles (8), even when their fixture text is large.

## Definite lowest-value declarations

### 0/10: `renders empty notice when items dictionary is empty`

- Source: `web/src/components/Components.test.tsx:174`.
- Reason: Same component, empty items, zero total and exact same message assertion as the earlier empty-message test. Only title differs; production empty branch does not use title for its result. No additional regression oracle.
- Representative: Components.test.tsx: renders empty message when items dictionary is empty (score 5).

### 0/10: `toggles granularity between hourly, daily, and weekly`

- Source: `web/src/views/MetricProgression.test.tsx:144`.
- Reason: Repeats the same ledger fixture, hourly/weekly/daily clicks, active-class assertions and absent-cumulative assertion already present in the earlier stronger test. Its Total Invocations wait is independently asserted by the initial KPI test, so it adds no independent contract.
- Representative: MetricProgression.test.tsx: toggles granularity and verifies cumulative option is removed (score 7); renders loading state initially and then displays progression KPIs and trend charts (score 6).

### 0/10: `renders zero telemetry state when no ledgers exist`

- Source: `web/src/views/MetricProgression.test.tsx:190`.
- Reason: The same mocked empty response and rendered view are asserted by the later zero-ledger test, which also checks the explanatory text. This regex heading assertion adds no distinct behavior.
- Representative: MetricProgression.test.tsx: renders gracefully with zero ledgers displaying empty state without runtime errors (score 6).

These are newly identified residual duplicates, distinct from the previous 50 removals. Their stronger counterparts retain scores 5, 7 and 6 respectively. They remain in source because this iteration is a review.

## Other low-score priorities

These entries have low marginal value, but some protect a small unique case. Their numeric scores deliberately distinguish them from the exact duplicates above.

| Score | Declaration | Why the score is low |
|---:|---|---|
| 2 | `backend/cmd/agentic-lens/release_verification_test.go:88` / `TestMakefile_ReleaseTargetsListedInHelp` | Invokes actual make help but checks target names in descriptive text, not target executability. Mostly documentation spelling coverage. |
| 2 | `backend/internal/domain/aggregations_test.go:626` / `TestDateFiltersAndAggregations` | Only verifies a seven-day duration. TestResolveDatePreset checks exact midnight start/end for seven days plus 24h and unknown inputs. Little extra value from a nearby reference date. |
| 2 | `backend/internal/domain/enums_test.go:471` / `TestAgentRole_FallbackConstants` | Asserts two string constants. These labels are user-visible fallback conventions, but this pins spelling without proving fallback selection. Actual fallback/serialization tests provide stronger evidence. |
| 2 | `web/src/components/AnomalyBadge.test.tsx:45` / `renders small, medium, and large sizes according to size prop` | Only verifies three CSS class tokens; no dimensions, accessibility or behavior are exercised. |
| 2 | `web/src/components/Components.test.tsx:293` / `applies custom className to container element` | Only locks a pass-through CSS class on a wrapper. It cannot establish tooltip placement, visibility or usability. |
| 2 | `web/src/components/Components.test.tsx:327` / `exports InfoIcon alias identically` | Pins reference identity of a one-line alias. Existing consumers would also fail if the export disappeared; little user behavior is demonstrated. Compatibility intent should be confirmed before pruning. |
| 3 | `backend/cmd/agentic-lens/release_verification_test.go:13` / `TestWorkflowFiles_Integrity` | Reads real workflow files but only searches for substrings. Comments or disconnected steps can satisfy it; missing, reordered or ineffective release behavior is not reliably detected. |
| 3 | `backend/internal/domain/aggregations_test.go:1409` / `TestOrderRoles` | Real ordering behavior, substantially repeated by enums_test.go comprehensive ordering tables. Consolidation needs a case-by-case transfer of custom/default order and arbitrary-role assertions. |
| 3 | `backend/internal/domain/aggregations_test.go:1439` / `TestOrderSubagentRoles` | Real ordering behavior, substantially repeated by enums_test.go comprehensive ordering tables. Consolidation needs a case-by-case transfer of custom/default order and arbitrary-role assertions. |
| 3 | `backend/internal/domain/aggregations_test.go:1795` / `TestOrderRoles_And_OrderSubagentRoles_AdversarialEdgeCases` | Real ordering behavior, substantially repeated by enums_test.go comprehensive ordering tables. Consolidation needs a case-by-case transfer of custom/default order and arbitrary-role assertions. |
| 3 | `backend/internal/domain/version_test.go:255` / `TestGetBuildInfo_ImmutabilityValueSemantics` | Changes source globals after reading a plain value struct and checks snapshot values. Primarily guards Go value semantics; limited application logic beyond GetBuildInfo construction. |
| 3 | `web/src/api/client.test.ts:193` / `throws structured error when ticketId does not exist on 404` | Repeats the same fetchWorkspaceDetail 404 message branch already exercised nearby with a different ticket literal. No distinct server lookup occurs in either mocked response. |
| 3 | `web/src/components/Components.test.tsx:83` / `handles zero total cost without divide-by-zero distortion` | Real rendering occurs, but the later zero-cost case asserts exact percentage count and absence of NaN. This earlier case adds mainly canonical labels and title. |
| 3 | `web/src/components/Components.test.tsx:301` / `applies custom placement classes correctly` | Checks placement-class mapping but not actual geometry. Some prop-branch coverage exists; low marginal value beside interaction tests. |
| 3 | `web/src/domain/anomalies.test.ts:97` / `narrows candidate to ToolFailureAnomalyRule via type guard across all valid members` | Runtime membership duplicates adjacent positive/negative guard tests. The compile-time narrowing assignment has value only if kept under tsc; transfer that before removing the wrapper. |
| 3 | `web/src/utils/theme.test.ts:15` / `defines valid hex colors across all theme tokens` | Validates real exported constants, but enforces a six-digit-hex implementation choice rather than semantic colors/contrast. Future valid CSS colors would fail. |
| 3 | `web/src/utils/theme.test.ts:71` / `returns deterministic hex colors for models via getModelColor` | Repeated model lookup and hex-shape checks overlap the complex-tag test exercising the same wrapper; no model-specific branch is shown. |
| 3 | `web/src/views/FleetOverview.test.tsx:78` / `displays empty state when total_tickets is zero` | Overlaps the later empty-list/zero-ticket case, which additionally checks refresh. The earlier fixture also zeros total cost, so treat this as near-overlap, not a mechanically identical duplicate. |
| 3 | `web/src/views/TicketDeepDive.test.tsx:999` / `ensures .table-card and .thread-accordion-item in web/src/index.css have overflow: visible to prevent tooltip clipping` | Reads actual CSS but regex matches source spelling and exact values, not layout/clipping. Low-strength oracle for a real historical regression; a geometry-based replacement is needed before dropping all protection. |
| 3 | `web/src/views/TicketDeepDive.test.tsx:1011` / `ensures .table-responsive, .data-table, and .thread-models-table rules in web/src/index.css satisfy responsive constraints` | Reads actual CSS but regex matches source spelling and exact values, not layout/clipping. Low-strength oracle for a real historical regression; a geometry-based replacement is needed before dropping all protection. |

## Critical integrity contracts

Only the following directly examined declarations receive 10. Scores of 8 or 9 elsewhere can still protect important independent boundaries; absence from this list is not evidence of expendability.

| Score | Declaration | Concrete protection |
|---:|---|---|
| 10 | `backend/cmd/agentic-lens/main_codex_test.go:112` / `TestGenerateCodexPreservesChildIdentityAndExcludesInheritedUsage` | Runs the CLI twice against isolated source rollouts and verifies root/child identity, token totals, replay exclusion and privacy. Domain-only tests cannot prove this end-to-end attribution. |
| 10 | `backend/internal/adapters/backfill/migration_runner_test.go:844` / `TestMigrationRunner_RejectsMixedCanonicalAndLegacyCollectionsAtomically` | Real filesystem migration rejects mixed-shape input and compares ledger bytes afterward. Essential protection against destructive writes on invalid input. |
| 10 | `backend/internal/adapters/codex/invocation_ownership_regression_test.go:11` / `TestCombinedCodexSessionsPreserveRepeatedRowsAndNestedOwners` | Extraction/combine boundary preserves repeated contributions and nested ownership rather than deduplicating legitimate rows. Protects actual exported metrics. |
| 10 | `backend/internal/adapters/http/server_test.go:2182` / `TestRecordedV4FactsSurviveDetailEventsAndOptionalHydration` | Exercises detail/events HTTP endpoints with optional hydration, owners, payload enrichment and immutable workspace snapshots. Protects persisted facts all the way to the API. |
| 10 | `backend/internal/adapters/live/sqlite_test.go:1469` / `TestHydrationFailuresRetainRecordedEventsAndLedgerBytes` | Runs unavailable/empty/corrupt enrichment modes across both source families while preserving committed events and file bytes. Optional enrichment must never erase supported facts. |
| 10 | `backend/internal/adapters/schemas_test.go:151` / `TestInvocationSerializationPreservesRowsAndSessionOwnershipWithoutID` | Actual DTO/serialization round trip preserves repeated session-owned rows without reintroducing obsolete invocation IDs. A wire-shape error would destroy attribution or collapse metrics. |
| 10 | `backend/internal/domain/migration/engine_test.go:564` / `TestInvariantValidator` | Adversarially mutates identity, schema, timestamps, code counters, invocations/tools and events; validates invariant-violation errors. Prevents migration from silently changing stored facts. |
| 10 | `backend/internal/domain/migration/invocation_ownership_regression_test.go:13` / `TestMigrationPreservesRepeatedRowsWithinLegacySessionOwners` | Real migration across legacy shapes preserves owners and every repeated row, serialized shape, and independently expected totals. Losing this can silently change historical usage. |
| 10 | `backend/internal/domain/session_facts_test.go:27` / `TestSessionsFromFactsPreservesOwnersAndEveryMetricRow` | Preserves repeated metric rows, explicit owners/source/user and separates tools/events across sessions, then checks derived totals and role/model counts. Prevents silent telemetry corruption. |
| 10 | `backend/internal/domain/session_facts_test.go:89` / `TestSessionsFromFactsRequiresExplicitInvocationOwner` | Production rejects unowned invocation rows rather than fabricating attribution. Essential integrity barrier for all downstream totals. |
| 10 | `web/src/views/TicketDeepDive.test.tsx:398` / `ignores ticket A %s enrichment after selecting ticket B and resets the drawer and disclosure` | Delays a response from the old ticket and checks the current ticket state/inspector. Prevents cross-ticket stale async data from appearing as current evidence. |

## Iteration 03 coverage and verification (historical)

- Numeric catalog: 1203 unique file/line entries, all integer scores in [0,10]; 102 Focused and 1101 Screened. All current declarations occur exactly once; no deleted test, helper or TestMain is counted.
- Declaration source fingerprint: `cf143e3784fdb64d306ba95b08c69e36eeb86af7c1258179539400dff9e3cf41` (SHA-256 over sorted file/line/body-hash identities).
- The same test diff was already validated in iteration 02 by independent AST/body comparison and full `make`, including 749 frontend cases, race tests, typecheck/lint and standalone embedded assets. No application/test changes have occurred since that gate.
- Fresh profile: `cd backend && go test -count=1 -json ./...` passed. Inspected 505 top-level tests; 0 failed and 0 skipped events. JSON event span 3.38s; compilation before the first event is excluded, so this is not a total command wall-clock measurement.
- Post-integration `make` results for documentation delivery are recorded in SUMMARY.md. Current frontend source is unchanged; no claim is made that a profiling run alone substitutes for the required gate.

Slowest ordinary Go test completions in the fresh profile:

| Seconds | Test | Score |
|---:|---|---:|
| 0.26 | `agenticlens/cmd/agentic-lens/TestCLIGenerate_SessionRefreshMatchesFreshExtraction` | 9 |
| 0.23 | `agenticlens/cmd/agentic-lens/TestCLIGenerate_ThreeDeviceHandoffAccumulation` | 7 |
| 0.23 | `agenticlens/cmd/agentic-lens/TestCLIGenerateCommand` | 7 |
| 0.22 | `agenticlens/internal/adapters/backfill/TestBackfillEngine_MigrationThenSourceBackfillMakesLedgerReady` | 8 |
| 0.19 | `agenticlens/internal/adapters/git/TestResolveBaseBranch_ProbeCandidates` | 7 |
| 0.18 | `agenticlens/internal/adapters/git/TestResolveBaseBranch_Configured` | 7 |
| 0.18 | `agenticlens/cmd/agentic-lens/TestCLIGenerate_SparseAnomalyEvidenceAndZeroErrorMessage` | 7 |
| 0.16 | `agenticlens/internal/adapters/live/TestOfflineAndLiveToolExecutionParity` | 8 |

Timing is not a deletion criterion on its own. This ordinary uncached run and the earlier parallel/race runs have different conditions; no speedup is inferred. Broad table declarations can contain both strong boundary checks and weak repeated examples, so a high declaration score does not endorse every subcase. No numeric target runtime is justified by this review.

## Original iteration 03 next-step boundaries (historical)

Start with the three zero-score duplicates if the user requests another deletion pass. For the 2-3 band, inspect exact assertion transfer first: role ordering includes arbitrary/default priorities, tool classification retains a distinct segmentation-fault case (score 5), negative type narrowing needs tsc coverage, CSS regex tests protect a historical issue despite a weak layout oracle, and the basic reference-line smoke retains an otherwise unique label check (score 6).

Do not convert these ratings into a bulk deletion script. Preserve critical independent layer coverage: a domain merge test does not establish HTTP persistence safety, a mocked API test does not establish SQLite access mode, and a component test does not establish standalone assets. No new deletion approval is inferred from requesting scores.

Final delivery gate: full post-integration `make` passed, including all 749 frontend runtime cases, Go race checks, frontend typecheck/lint, canonical assets and four-platform binary builds. Source/test bodies remain unchanged from the scoring baseline.

## Iteration 04 catalog update

The user explicitly selected the complete release test file and date smoke for deletion, and requested merging the adversarial role wrapper into the enum test. The release tests covered source/help text and missing-input errors, not a successful CI release/publish run. The missing-input checks were genuine but limited; the user chose to remove that coverage. The stronger TestResolveDatePreset still checks exact boundaries.

Canonical role priorities/default order now live in three refined rows of the existing 19-row comprehensive enum table. Both entrypoints exercise every row, including a multi-element alphabetical remainder with alpha/orchestrator/zeta. No new top-level tests or table rows were added.

TEST_SCORES.md has 1,197 current rows: 499 Go and 698 TypeScript. Six obsolete rows are omitted; all other ratings persist, with refreshed source locations and the enum consolidation rationale. Current basis counts: 96 Focused and 1,101 Screened. The original three zero-score duplicates and eleven critical ten-score contracts remain unchanged. Earlier low-priority rows above document historical judgments, including the now-deleted cases; the current catalog is authoritative for membership.

See iterations/04/TEST_RESULTS.md, iterations/04/REVIEW.md and SUMMARY.md for final verification. No additional deletion is inferred from any remaining score.

## Iteration 05: can every score 0-4 test disappear?

No. Direct inspection supported eliminating 42 of the 47 donor declarations, while preserving their meaningful protection in existing tests. Seven were true duplicates or explicit alias/language overconstraints; 35 used existing recipients with their unique assertions retained. No new standalone tests, parameterized test declarations or helper suites were added.

Five independent guards remain, with scores unchanged:

- **2**: TestAgentRole_FallbackConstants pins both wire-facing fallback labels; consumer tests comparing to those constants do not pin their spelling.
- **3**: all-theme-token validation checks every exported entry, not only sampled hash outputs.
- **4**: TestCanonicalTicketTypes checks the actual exported list/order independently of parser acceptance.
- **4**: ThreadViewer empty collection checks a different render branch from a populated thread expansion.
- **4**: the substantial TicketDeepDive tooltip-wiring test owns view-specific bindings/cardinality and now also receives related interactions; placing it in a general KPI smoke would obscure the contract.

CSS source guards remain unchanged inside the existing responsive-wrapper contract, without claiming browser geometry coverage. Type narrowing still executes through the actual type guard under tsc, and separate negative compiler checks remain. Header aria associations, tooltip presence/absence, metadata precedence and existing callback assertions survive their transfers. Assertions were not merely packed into unrelated large tests to manufacture a count reduction.

See iterations/05/TEST_RESULTS.md and iterations/05/REVIEW.md for independent execution and semantic audits. No mutation-equivalence or blanket deletion guarantee is claimed.
