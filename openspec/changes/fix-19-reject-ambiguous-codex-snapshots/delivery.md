# fix(FIX-19): reject ambiguous Codex owner snapshots

Codex generation, backfill and historical hydration now reject multiple native snapshots for one owner instead of adding cumulative facts or selecting a readable copy. Registered automatic extraction also checks uniqueness; a registration does not grant authority to select one duplicate. Incomplete owners retain accepted history, and hydration continues with unaffected owners. Genuine repeated invocations, skills and delegations remain intact.

## Verification and review

Candidate: `1f32e6df7c4045948b4440d39cdfa1daf371080b` (production unchanged from `2618c04a`), based on the original main revision `b6ab0471224c8bcea7bc00adce867d3ba0de4a37`. Main integration found no additional changes. [Tester evidence](evidence/iterations/01/TESTS.md) records executable baseline counterexamples, regressions that fail before the fix, and corrected focused race results across all four consumer packages. The initial full changed-package race passed before the registered-capture correction.

All eight focus/model discovery cells are complete (one superseded Luna architecture run was cancelled and replaced; no approval was credited to it), with affected architecture/test re-reviews on the final candidate. Behavior/data evidence carries forward only for unchanged dependencies; the final budget delta is covered by both architecture and test profiles. Architecture reviewer Sol 6.1/high discovered ARCH-01: registered automatic generation bypassed the first guard and replaced accepted facts from one duplicate. The correction validates registered owner uniqueness and reports failed owners explicitly. Its durable automatic-generation regression passes for identical, divergent gzip and warning-bearing copies, preserving accepted session facts across repeats. Architecture Sol also found uncharged repeated discovery reads (ARCH-02); architecture Luna found candidate-count reset across scheduler passes (ARCH-03). Both were fixed with shared, charged operation-owned discovery accounting and durable red/green regressions. Test Luna confirmed the supplied candidate-count coverage gap. Data Luna’s cross-store checkpoint claim was refuted by source-store identity semantics and an independently executed existing collision/recovery test; it remains retained with attribution. The final test-only fixture correction was reviewed without findings under the updated policy by Sol 6.2/high and Luna 6/medium. Sol used a disposable broken-scheduler mutation to confirm the fairness assertion still detects starvation. Full capture race passed after the correction. Consolidated review: APPROVED after integrated verification. The full `make` gate passed on `1f32e6df`: Go unit/race and adapter/CLI integration, vet, frontend typecheck/lint, 231 frontend unit tests, 804 frontend integration tests, release/installation/capture/test-orchestration checks and four-platform native builds. Direct host engine `version --json` and manager `--manager-version --json` passed and identify commit `1f32e6df`; the dirty suffix reflects documentation/local-policy edits, with no application/test delta. Strict OpenSpec validation and diff whitespace checks pass. Initial full-gate failures exposed legacy fixture allowances; exact discovery-plus-extraction sizing preserves their assertions and full capture race then passed. Detailed review evidence is compacted in [attributed findings](evidence/iterations/01/REVIEW_SUMMARY.json).

## Integration with AL-80 and AL-73

Merge `54230a767b3404c4970783b21a1323ef7c29dc4d` combines FIX-19 head `5c1b5737` with main `4e86badf`. Git merged the two existing roadmap closures cleanly. The shared roadmap index was the only file both branches edited. FIX-19's Codex discovery, combination, hydration, backfill and registered capture implementation and tests match the previously reviewed branch. AL-80's installer, callback routing and worker match main. The worker still passes its session and byte allowances to registered generation. This preserves the basis for the earlier architecture, data and test reviews. The integration review reopened behavior at that call boundary and the shared reconciliation cursor.

Focused race checks passed for all five consuming packages at the merge. They exercised manual and registered automatic generation with ambiguous sources, capture budget accounting and pending history, backfill and hydration, OpenCode user installation, migration, worktree admission and worker publication with explicit budgets. AL-80's native OpenCode 1.19.5 callback and worker publication evidence carries forward for its unchanged integration. The native OpenCode session was not repeated in this run. [AL-80 native verification](../archive/2026-11-17-al-80-capture-opencode-sessions-across-worktrees/evidence/iterations/01/native-verification.md)

The `gpt-6.2-sol/high` behavior reviewer reproduced BEH-INT-01. With a 16 KiB allowance, eight passes retained an OpenCode owner’s 11 input tokens while its source held 37. The AL-80 parent refreshed both owners under that allowance. The `gpt-6-luna/medium` first pass found no blocker. Commit `136504cb613435d7b352d4f3a930abff97b97f66` keeps the cursor before an owner deferred by earlier charged work. An owner that exhausts a full allowance still advances. Both reviewers approved the correction. The independent probe and the durable zero-remainder, partial-remainder and oversized-owner regressions pass. Existing capture budget and scheduler race tests also pass. The reviewed change affects cursor advancement and its operation-local state. Codex ancestry, snapshot uniqueness, capture identity, fact replacement, byte accounting and AL-80 installer and callback code are unchanged from their respective reviewed revisions.

Integrated verification passed on `136504cb613435d7b352d4f3a930abff97b97f66`. Full `make` completed Go vet and race suites, 231 frontend unit tests, 804 integration tests, release and installation checks, all 24 capture-distribution cases, orchestration checks and four-platform native builds. The host manager routed to the integrated engine and both version checks identified `136504cb`. Strict OpenSpec validation passed. The prior merge-only make run was superseded when BEH-INT-01 was confirmed. The compact review record retains the discovery credit and disposition.

## Limits

Fixtures are isolated synthetic native records; no real source store was edited. Before-fix CLI gzip generation encountered a pre-existing ownership decoding failure; uncompressed CLI generation and gzip hydration/backfill independently established the defect. Failed family backfill conservatively retains its accepted family; exact hydration exposes unambiguous owners separately. Source stores can change after an inspection, so this is snapshot ambiguity rejection rather than an atomic lock over external Codex writers.

## Telemetry coverage

The integrated CLI refreshed the FIX-19 schema 6 ledger from the configured local Codex source. Native root `01a1eb81-91c3-71a3-91c3-edd6caca28b4` has its original eight children and the new Sol integration reviewer. The previous eight independent T3/Luna sessions and the two integration Luna sessions have their own native root IDs. Their exact task prompts and rollout metadata establish review provenance. No independent session was attached as a native child. Invocation costs are unavailable and the active root duration is provisional.

## Pull request

Reject identical and divergent native Codex snapshots before ancestry resolution, combination or complete-owner certification. Preserve accepted history on failed reads, expose explicit ambiguity diagnostics, and keep genuine occurrence multiplicity and unaffected owners usable.

Add actual CLI generation, registered automatic generation, backfill, hydration, discovery and nested-combination regressions, including conflicting children, warning-bearing copies, portable history and unchanged repeat runs. The baseline demonstrates doubled facts and arbitrary complete hydration; the corrected focused race checks pass. Validation: full integrated `make`, focused consumer race regressions, full capture race, native host execution and strict OpenSpec validation pass. Integration with AL-80 exposed a fourth issue. A Codex scan could repeatedly consume the allowance ahead of a fitting OpenCode owner. The cursor now gives that deferred owner another turn while preserving its accepted facts. Both integration behavior reviewers approved the correction. Full `make` passed on `136504cb` atop main `4e86badf`.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 37
- **Active Agent Duration:** 2h 52m (recorded duration including 980s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 218.5k tokens (49% of 450k healthy ceiling, 0 compactions) | Subagents peak 180.3k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.8 avg turns/run (peak: 5 turns) | Planner 1.0, Coder 5.0, Tester 5.0, Reviewer-architecture 4.0, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 4.2:1 (+857 test lines / +202 prod lines)
- **Code Delivered:** +1,059 / -64 lines across 17 code files (excluding docs)
- **Reasoning Effort:** 55,936 reasoning tokens (21.0% of generated output)
- **Tool Health & Self-Correction:** 923 tool calls (76 self-corrections, 91.8% success rate)

### Anomalies & Health Flags
- **Status:** 10 execution anomalies detected (4 CRITICAL, 6 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (12 cycles) — Subagent roles 'coder' and 'tester' alternated for 12 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (18 cycles) — Subagent roles 'coder' and 'tester' alternated for 18 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (6 cycles) — Subagent roles 'tester' and 'coder' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (5 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 5 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-architecture (4 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-architecture' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> reviewer-architecture (4 cycles) — Subagent roles 'tester' and 'reviewer-architecture' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (4 cycles) — Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> tester (3 cycles) — Subagent roles 'reviewer-tests' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (7 consecutive errors) — Tool 'command' encountered 7 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 61,707 | 3,128 | 383,104 | 86.1% | 63.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 101,570 | 23,251 | 9,694,848 | 99.0% | 134.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 264,727 | 39,818 | 23,241,728 | 98.9% | 180.3k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 12 | 1,107,672 | 88,872 | 49,229,696 | 97.8% | 218.5k | HEALTHY | Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations) |
| **reviewer-architecture** | 1 | 115,708 | 16,493 | 5,397,376 | 97.9% | 135.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 2 | 231,011 | 20,625 | 4,641,792 | 95.3% | 139.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 1 | 122,569 | 7,511 | 1,816,832 | 93.7% | 117.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 2 | 145,955 | 10,704 | 2,416,000 | 94.3% | 99.5k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **21** | **2,150,919** | **210,402** | **96,821,376** | **97.8%** | **218.5k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 10 | 524,804 | 35,789 | 3,383,680 | 86.6% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **gpt-6.1-sol** | 8 | 1,021,885 | 145,691 | 72,407,680 | 98.6% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **gpt-6.2-astra** | 1 | 412,752 | 13,151 | 17,623,680 | 97.7% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **gpt-6.2-sol** | 2 | 191,478 | 15,771 | 3,406,336 | 94.7% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **21** | **2,150,919** | **210,402** | **96,821,376** | **97.8%** | **Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations)** |

Retained recorded history: 10 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
