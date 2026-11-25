# fix(AL-94): make explicit backfill repeatable and report coverage

## Outcome

Explicit backfill verifies recorded sessions from the chosen source and preserves equivalent ledger and summary bytes and modification times. A requested summary can be repaired on its own while preserving authored text. Actual ledger and summary writes are counted separately from the existing `Updated` count.

Coverage identifies changed, unchanged, retained and newly discovered sessions per selected harness, including Claude Code after integration with AL-98. It distinguishes unavailable collections from verified empty collections, and preserves verified observations when interruption or a rejected candidate prevents publication. Migration state remains independent. Store preflight reports blocked work and keeps the established failure exit. OpenCode reads recorded owners before spending the remaining budget on newly discovered descendants.

## Verification and review

Candidate `da90e8e0` included `origin/main` at `db4477fa`. Full integrated `make`, the three OpenSpec workflow tests and strict change validation passed. [Test evidence](evidence/TESTS.md)

The branch then integrated main through `be630e36`, which brings AL-98 Claude Code backfill and the FIX-21 capture fix. Claude Code recovery now reports coverage like the other harnesses ([design steering](design.md#steering)). A quick independent Opus review of the merge found one defect: retained Claude Code owners showed a generic reason. That is fixed in `7bf89f51`, with regression assertions. On merge `390dccb7`, the backfill package passed. The integrated `make` run on this host failed only in timing-sensitive capture tests that AL-94 does not change, and the CLI and capture packages exceeded Go's 10-minute test timeout under load average 16 on 12 cores. The failed tests passed when rerun in isolation: twice on the branch and twice on `origin/main`. CI on the pushed head is the final gate for those packages.

Independent behavior, data integrity, architecture and test-effectiveness reviews ran under gpt-6.2-sol high and gpt-6-luna medium. The candidate has no unresolved blocking findings. The final behavior and data integrity reviews cover rejected preflight, cancellation and preservation. Both architecture profiles verified the final observation-only fallback. Earlier Sol behavior and data reviews carry forward for their unchanged Git/source preflight and publication paths. Earlier test effectiveness verdicts carry forward for their unchanged cases. Both architecture profiles verified the added regression, and Sol demonstrated that discarding source observations makes it fail. [Attributed review findings](evidence/REVIEW_SUMMARY.json)

## Limits

The CLI renderer's uncommitted-count output was verified with an isolated probe and mutation test. The permanent CLI suite covers other reporting fields, and the adapter regression covers uncommitted owner counts. AL-95 retains ownership of preview and rollout decisions.

Codex telemetry records 28 owners: the bound root, its three source-linked native children and all 19 independently identified reviewer sessions. An uncommitted automatic regeneration by the pre-FIX-21 capture worker had dropped the root's three planner, coder and tester child links. It was discarded. Automatic capture by the merged engine (`generate --capture-auto`) restored the ledger: all 28 sessions, sources and parents match the committed `b2e7bb4d` ledger, with three of three child links present. All 27 AL-94 registrations reconciled to current. One reviewer lacked a verified automatic binding and was recovered by exact source session ID. Runs that did not return an independent native ID could not be matched individually. The final generation, by the merged engine, adds this Claude Code integration session and its Opus reviewer as a source-proven child, for 30 owners, without changing any Codex owner or link. The Claude Code root turn was still open at extraction, so its last interval is provisional. Incomplete cost information remains unavailable.

## Pull request

- Preserve equivalent backfill artifacts, and repair the requested summary independently.
- Report selected-source coverage, retained owners and effective writes while keeping `Updated` and existing failure exits. Claude Code sources from AL-98 report the same coverage.
- Verify native Codex and OpenCode fixtures, mixed-source preflight, summary rollback, cancellation and bounded retry. Strict OpenSpec validation passed. Integrated `make` passed on the first candidate. After the main merges, only load-sensitive capture tests failed locally; they pass in isolation, and CI gates them.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 82
- **Active Agent Duration:** 10h 44m (recorded duration including 436s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 272.5k tokens (61% of 450k healthy ceiling, 0 compactions) | Subagents peak 298.0k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.8 avg turns/run (peak: 5 turns) | Planner 1.0, Coder 4.0, Tester 5.0, Reviewer 1.0
- **Test-to-Production Code Ratio:** 0.9:1 (+923 test lines / +1,040 prod lines)
- **Code Delivered:** +1,963 / -318 lines across 19 code files (excluding docs)
- **Reasoning Effort:** 105,317 reasoning tokens (21.4% of generated output)
- **Tool Health & Self-Correction:** 1,193 tool calls (100 self-corrections, 91.6% success rate)

### Anomalies & Health Flags
- **Status:** 4 execution anomalies detected (1 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (53 cycles) — Subagent roles 'coder' and 'tester' alternated for 53 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.6k tokens) — Context tokens reached 250.6k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.4k tokens) — Context tokens reached 251.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (252.6k tokens) — Context tokens reached 252.6k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 135,794 | 1,361 | 765,952 | 84.9% | 134.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 217,749 | 64,952 | 32,035,072 | 99.3% | 298.0k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 211,927 | 58,924 | 34,244,864 | 99.4% | 289.4k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 1 | 48 | 16,578 | 2,108,321 | 100.0% | 123.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 26 | 2,652,170 | 245,438 | 94,071,138 | 97.3% | 272.5k | ELEVATED | Unavailable (0 observed, 0 partial, 26 unavailable / 26 invocations) |
| **Total** | **30** | **3,217,688** | **387,253** | **163,225,347** | **98.1%** | **298.0k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 30 unavailable / 30 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **claude-opus-5-5** | 2 | 252 | 67,140 | 16,175,235 | 100.0% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **gpt-6-luna** | 10 | 591,113 | 42,973 | 5,821,824 | 90.8% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **gpt-6.2-sol** | 18 | 2,626,323 | 277,140 | 141,228,288 | 98.2% | Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations) |
| **Total** | **30** | **3,217,688** | **387,253** | **163,225,347** | **98.1%** | **Unavailable (0 observed, 0 partial, 30 unavailable / 30 invocations)** |

Retained recorded history: 28 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
