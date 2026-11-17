# fix(FIX-19): reject ambiguous Codex owner snapshots

Codex generation, backfill and historical hydration now reject multiple native snapshots for one owner instead of adding cumulative facts or selecting a readable copy. Registered automatic extraction also checks uniqueness; a registration does not grant authority to select one duplicate. Incomplete owners retain accepted history, and hydration continues with unaffected owners. Genuine repeated invocations, skills and delegations remain intact.

## Verification and review

Candidate: `1f32e6df7c4045948b4440d39cdfa1daf371080b` (production unchanged from `2618c04a`), based on current main `b6ab0471224c8bcea7bc00adce867d3ba0de4a37`. Main integration found no additional changes. [Tester evidence](evidence/iterations/01/TESTS.md) records executable baseline counterexamples, regressions that fail before the fix, and corrected focused race results across all four consumer packages. The initial full changed-package race passed before the registered-capture correction.

All eight focus/model discovery cells are complete (one superseded Luna architecture run was cancelled and replaced; no approval was credited to it), with affected architecture/test re-reviews on the final candidate. Behavior/data evidence carries forward only for unchanged dependencies; the final budget delta is covered by both architecture and test profiles. Architecture reviewer Sol 6.1/high discovered ARCH-01: registered automatic generation bypassed the first guard and replaced accepted facts from one duplicate. The correction validates registered owner uniqueness and reports failed owners explicitly. Its durable automatic-generation regression passes for identical, divergent gzip and warning-bearing copies, preserving accepted session facts across repeats. Architecture Sol also found uncharged repeated discovery reads (ARCH-02); architecture Luna found candidate-count reset across scheduler passes (ARCH-03). Both were fixed with shared, charged operation-owned discovery accounting and durable red/green regressions. Test Luna confirmed the supplied candidate-count coverage gap. Data Luna’s cross-store checkpoint claim was refuted by source-store identity semantics and an independently executed existing collision/recovery test; it remains retained with attribution. The final test-only fixture correction was reviewed without findings under the updated policy by Sol 6.2/high and Luna 6/medium. Sol used a disposable broken-scheduler mutation to confirm the fairness assertion still detects starvation. Full capture race passed after the correction. Consolidated review: APPROVED after integrated verification. The full `make` gate passed on `1f32e6df`: Go unit/race and adapter/CLI integration, vet, frontend typecheck/lint, 231 frontend unit tests, 804 frontend integration tests, release/installation/capture/test-orchestration checks and four-platform native builds. Direct host engine `version --json` and manager `--manager-version --json` passed and identify commit `1f32e6df`; the dirty suffix reflects documentation/local-policy edits, with no application/test delta. Strict OpenSpec validation and diff whitespace checks pass. Initial full-gate failures exposed legacy fixture allowances; exact discovery-plus-extraction sizing preserves their assertions and full capture race then passed. Detailed review evidence is compacted in [attributed findings](evidence/iterations/01/REVIEW_SUMMARY.json).

## Limits

Fixtures are isolated synthetic native records; no real source store was edited. Before-fix CLI gzip generation encountered a pre-existing ownership decoding failure; uncompressed CLI generation and gzip hydration/backfill independently established the defect. Failed family backfill conservatively retains its accepted family; exact hydration exposes unambiguous owners separately. Source stores can change after an inspection, so this is snapshot ambiguity rejection rather than an atomic lock over external Codex writers.

## Telemetry coverage

The built native engine generated schema 6 telemetry with explicit source `codex`, ticket `FIX-19`, type `fix`, root `01a1eb81-91c3-71a3-91c3-edd6caca28b4` and literal nonce `FIX-19-final-20261117-95c2d6478a83`. Native engine invocation runs directly in this project; manager-only routing flags are unnecessary. Verified nine unique Codex owners: the exact root and eight native children with matching parent IDs, containing eight Sol 6.1 and one Sol 6.2 invocation. No prior ledger or other-harness facts existed to preserve. Separately delegated T3/Luna sessions are not native descendants in the selected source family; their telemetry/cost is missing from this ledger, not zero. All nine captured invocation costs report unavailable evidence. Root duration remains provisional while this session is open; no complete team-cost or final-duration claim is made.

## Pull request

Reject identical and divergent native Codex snapshots before ancestry resolution, combination or complete-owner certification. Preserve accepted history on failed reads, expose explicit ambiguity diagnostics, and keep genuine occurrence multiplicity and unaffected owners usable.

Add actual CLI generation, registered automatic generation, backfill, hydration, discovery and nested-combination regressions, including conflicting children, warning-bearing copies, portable history and unchanged repeat runs. The baseline demonstrates doubled facts and arbitrary complete hydration; the corrected focused race checks pass. Validation: full integrated `make`, focused consumer race regressions, full capture race, native host execution and strict OpenSpec validation pass. Independent reviews found and resolved three issues; the compact findings retain discovery credit and the refuted claim.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 29
- **Active Agent Duration:** 2h 21m (recorded duration including 126s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 180.0k tokens (40% of 450k healthy ceiling, 0 compactions) | Subagents peak 164.2k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.6 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 4.0, Tester 4.0, Reviewer-architecture 4.0, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 3.8:1 (+705 test lines / +184 prod lines)
- **Code Delivered:** +889 / -57 lines across 16 code files (excluding docs)
- **Reasoning Effort:** 46,500 reasoning tokens (21.3% of generated output)
- **Tool Health & Self-Correction:** 741 tool calls (68 self-corrections, 90.8% success rate)

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
| **coder** | 1 | 89,913 | 19,688 | 8,146,560 | 98.9% | 121.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 252,450 | 32,718 | 19,401,216 | 98.7% | 164.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 10 | 777,602 | 73,466 | 32,825,856 | 97.7% | 180.0k | HEALTHY | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **reviewer-architecture** | 1 | 115,708 | 16,493 | 5,397,376 | 97.9% | 135.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 102,053 | 7,724 | 1,719,552 | 94.4% | 105.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 122,569 | 7,511 | 1,816,832 | 93.7% | 117.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 2 | 145,955 | 10,704 | 2,416,000 | 94.3% | 99.5k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **18** | **1,667,957** | **171,432** | **72,106,496** | **97.7%** | **180.0k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 8 | 426,222 | 30,610 | 2,845,440 | 87.0% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **gpt-6.1-sol** | 8 | 997,951 | 135,028 | 67,018,880 | 98.5% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **gpt-6.2-astra** | 1 | 181,264 | 2,924 | 1,758,080 | 90.7% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **gpt-6.2-sol** | 1 | 62,520 | 2,870 | 484,096 | 88.6% | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **18** | **1,667,957** | **171,432** | **72,106,496** | **97.7%** | **Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations)** |

Retained recorded history: 8 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
