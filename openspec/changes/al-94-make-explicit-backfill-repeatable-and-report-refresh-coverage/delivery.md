# fix(AL-94): make explicit backfill repeatable and report coverage

## Outcome

Explicit backfill verifies recorded sessions from the chosen source and preserves equivalent ledger and summary bytes and modification times. A requested summary can be repaired on its own while preserving authored text. Actual ledger and summary writes are counted separately from the existing `Updated` count.

Coverage identifies changed, unchanged, retained and newly discovered sessions per selected harness. It distinguishes unavailable collections from verified empty collections, and preserves verified observations when interruption or a rejected candidate prevents publication. Migration state remains independent. Store preflight reports blocked work and keeps the established failure exit. OpenCode reads recorded owners before spending the remaining budget on newly discovered descendants.

## Verification and review

Candidate `da90e8e0` includes `origin/main` at `db4477fa`. Full integrated `make`, the three OpenSpec workflow tests and strict change validation passed. [Test evidence](evidence/TESTS.md)

Independent behavior, data integrity, architecture and test-effectiveness reviews ran under gpt-6.2-sol high and gpt-6-luna medium. The candidate has no unresolved blocking findings. The final behavior and data integrity reviews cover rejected preflight, cancellation and preservation. Both architecture profiles verified the final observation-only fallback. Earlier Sol behavior and data reviews carry forward for their unchanged Git/source preflight and publication paths. Earlier test effectiveness verdicts carry forward for their unchanged cases. Both architecture profiles verified the added regression, and Sol demonstrated that discarding source observations makes it fail. [Attributed review findings](evidence/REVIEW_SUMMARY.json)

## Limits

The CLI renderer's uncommitted-count output was verified with an isolated probe and mutation test. The permanent CLI suite covers other reporting fields, and the adapter regression covers uncommitted owner counts. AL-95 retains ownership of preview and rollout decisions.

Final Codex generation preserved all 27 previously accepted owners and recorded 28 owners in total, including the bound root, its three source-linked native children and all 19 independently identified reviewer sessions. One reviewer lacked a verified automatic binding and was recovered by exact source session ID. Runs that did not return an independent native ID could not be matched individually. No prior ledger owners from another harness were present. The root turn was still open at extraction, and incomplete cost information remains unavailable.

## Pull request

- Preserve equivalent backfill artifacts, and repair the requested summary independently.
- Report selected-source coverage, retained owners and effective writes while keeping `Updated` and existing failure exits.
- Verify native Codex and OpenCode fixtures, mixed-source preflight, summary rollback, cancellation and bounded retry. Integrated `make` and strict OpenSpec validation passed.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 79
- **Active Agent Duration:** 10h 6m (recorded duration including 782s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 269.2k tokens (60% of 450k healthy ceiling, 0 compactions) | Subagents peak 298.0k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 3.3 avg turns/run (peak: 5 turns) | Planner 1.0, Coder 4.0, Tester 5.0
- **Test-to-Production Code Ratio:** 0.9:1 (+896 test lines / +981 prod lines)
- **Code Delivered:** +1,877 / -299 lines across 17 code files (excluding docs)
- **Reasoning Effort:** 85,792 reasoning tokens (21.2% of generated output)
- **Tool Health & Self-Correction:** 1,064 tool calls (99 self-corrections, 90.7% success rate)

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
| **Unknown** | 25 | 2,646,727 | 192,861 | 77,570,688 | 96.7% | 269.2k | ELEVATED | Unavailable (0 observed, 0 partial, 25 unavailable / 25 invocations) |
| **Total** | **28** | **3,212,197** | **318,098** | **144,616,576** | **97.8%** | **298.0k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 28 unavailable / 28 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 10 | 591,113 | 42,973 | 5,821,824 | 90.8% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **gpt-6.2-sol** | 18 | 2,621,084 | 275,125 | 138,794,752 | 98.1% | Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations) |
| **Total** | **28** | **3,212,197** | **318,098** | **144,616,576** | **97.8%** | **Unavailable (0 observed, 0 partial, 28 unavailable / 28 invocations)** |

Retained recorded history: 24 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
