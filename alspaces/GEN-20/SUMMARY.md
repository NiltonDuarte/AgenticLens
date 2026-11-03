# GEN-20: Unique worktree directory identifiers

## Delivered behavior

Capture coordination directories use the linked Git worktree administrative name qualified by its observed identity, or a persisted 16-hex random identifier when no linked-worktree identifier is available. Concurrent first opens share the committed mapping for the same scope and isolate distinct scopes. Bounded collisions skip existing filesystem entries without adopting or overwriting them.

Adapters and CLI consumers resolve the mapping consistently for read-only opens, aliases, provisional previews, installation, policy checks, diagnostics, workers and supported relocation/resume. Internal numeric SQL IDs remain unchanged. Domain calculations, portable ledger authority and frontend behavior are unchanged.

## Scope and baseline

- Initial revision: dc44619ae15586fe1194f872743033f523635c92; working tree was clean. Ticket branch: chore/GEN-20-unique-worktree-identifiers; configured workspace root: alspaces.
- The user approved the baseline contract and iteration 01 plan, then clarified that database compatibility is unnecessary because the database-introducing version is unreleased. [Iteration 02](iterations/02/PLAN.md) supersedes numeric-path preservation, protocol upgrades and old-binary fencing in the historical [contract](CONTRACT.md). Registry protocol remains 2 with the mapping in its current schema. Existing unrelated protocol-1 recovery remains supported.
- Unsupported local layouts fail explicitly without rewriting the database. Existing numeric layouts require fresh capture state; this ticket performs no automatic reset or migration.
- Initial implementation: 67d2a8bdd7039e7638e5dfb68fc435993252433b. Corrected revision: 13a6ff38c97e9c9d3019d76a0944ce46afe1814e. Both were synchronized with origin/main; upstream remained at the baseline.

## Verification

Coder and tester: gpt-6.1-sol, high reasoning effort. The coder executed focused capture/CLI race checks; the tester inspected fixture reachability, verified test discovery, and reused those executed results. [Test evidence](iterations/02/TEST_REPORT.md) distinguishes inspection, discovery and execution.

The coordinator (gpt-6.1-sol, high reasoning effort) verified the corrections: the new direct-root and physical-alias regression failed before the path fix; affected capture/CLI race tests passed afterward. The Git naming mutation that passed the original test now fails the stronger administrative-name assertion. Unknown/foreign aliases, redirected coordinates, concurrent allocation, fallback stability, installation and relocation coverage still pass.

Integrated make passed on the original reviewed revision: backend races, frontend typecheck/lint and 790 tests, Python release/installation/distribution checks, frontend embedding and all configured platform builds. Final make also passed on corrected revision 13a6ff38c97e9c9d3019d76a0944ce46afe1814e; its native binary reports that commit.

## Review findings and attribution

Eight independent first-pass reviews covered behavior, data/failure handling, architecture and tests on 67d2a8b, each under acknowledged gpt-6.1-sol/high and gpt-6-luna/medium settings. The Sol behavior, data and architecture reviews requested changes for one shared defect; their Luna counterparts found no blocking issues. The Sol test review approved with one useful suggestion; the Luna test review approved without findings. A mistyped base SHA in two Luna reports did not change their reviewed diff: the coordinator verified that the reviewed candidate parent is the requested baseline.

- Severity 5: ordinary configured bases beneath a directory named worktrees were misclassified as coordination aliases. Independently discovered by behavior, data and architecture reviewers, each gpt-6.1-sol/high. Fixed by the coordinator, gpt-6.1-sol/high, by requiring the projects/<name>/worktrees/<token> shape; direct and physical-alias before/after tests verify the fix while alias rejection remains intact.
- Severity 3 suggestion: the Git naming oracle accepted randomized substitutes. Discovered by the test reviewer, gpt-6.1-sol/high, using an isolated mutation. The coordinator, gpt-6.1-sol/high, strengthened the existing concurrency test using independently observed Git administrative names and hexadecimal qualification, then confirmed the mutation is rejected.

The host requires small review corrections to be fixed locally without further delegates. The coordinator therefore verified these narrow corrections directly and carried forward the unaffected independent review evidence. No reviewer is claimed to have reviewed 13a6ff3; final consolidated approval is APPROVED after the corrected revision passed make. Compact findings and every discovery attribution are retained in [REVIEW_SUMMARY.json](iterations/02/REVIEW_SUMMARY.json). Earlier implementation self-review also corrected an arbitrary probe-child check in root validation; focused tests passed after that correction.

## Telemetry and limits

Final ingestion succeeded with an explicit source codex, ticket type chore and ticket title. The configured local registry uses the unsupported earlier layout; isolated scratch capture state was used, the temporary local configuration was removed, and a byte hash confirmed the original database was unchanged. The built native CLI takes generate flags directly; dispatcher-only --al-offline/--al-project flags are not supported by the native executable.

Source inspection identifies root 01a1a395-ebd9-7820-8931-b7c6a91a5906 and native planner, coder and tester child sessions. Eight T3 reviewer sessions are recorded by Codex as separate roots without native parent lineage. Ledger membership was verified against real Codex metadata: exactly the root and three expected native children are present, all source codex with the correct parent links. The eight separately rooted T3 reviewer sessions are omitted because their source metadata lacks native parent lineage; their usage is not included. Pricing is unavailable for the four included invocations, so cost remains unavailable. The ledger is explicitly incomplete for total ticket usage. No prior GEN-20 ledger or other-harness sessions existed at baseline. Raw rollouts and detailed working run/finding records remain outside Git.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 15
- **Active Agent Duration:** 40m 18s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 160.0k tokens (80% of 200k healthy ceiling, 0 compactions) | Subagents peak 131.9k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.3 avg turns/run (peak: 3 turns) | Planner 2.0, Coder 2.0, Tester 3.0
- **Test-to-Production Code Ratio:** 1.3:1 (+410 test lines / +308 prod lines)
- **Code Delivered:** +718 / -57 lines across 17 code files (excluding docs)
- **Reasoning Effort:** 22,518 reasoning tokens (23.2% of generated output)
- **Tool Health & Self-Correction:** 206 tool calls (15 self-corrections, 92.7% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (4 cycles) — Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 134,214 | 6,216 | 795,904 | 85.6% | 81.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 94,051 | 21,073 | 4,431,872 | 97.9% | 131.9k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 73,139 | 7,921 | 2,708,224 | 97.4% | 126.2k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 142,644 | 39,271 | 13,368,448 | 98.9% | 160.0k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **4** | **444,048** | **74,481** | **21,304,448** | **98.0%** | **160.0k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 4 | 444,048 | 74,481 | 21,304,448 | 98.0% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **4** | **444,048** | **74,481** | **21,304,448** | **98.0%** | **Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations)** |

Capture observation finalized: 2026-11-03T23:18:55.625649578Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
