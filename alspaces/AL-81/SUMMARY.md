# AL-81: Generate metrics without capture registry access

Manual `generate` now reads authoritative Codex rollout files or OpenCode source SQLite data without opening, creating or changing the AgenticLens capture registry. Capture installation, enablement, state and worker policy are unnecessary. `--capture-dir` remains accepted without a manual state effect; explicit budget values must remain positive. Source/project validation, output containment, portable owner refresh, shared publication locking and interruption recovery remain intact. Automatic generation retains strict enabled-policy and registered-owner checks; explicit session/nonce selectors cannot bypass them.

The user approved the [contract](CONTRACT.md) and [plan](iterations/01/PLAN.md) on 2026-11-11. Work began from clean baseline `4e5c458322a2cddfdb5911b26a4569ea75f06a09`. The latest integration, requested on 2026-11-13, merges `origin/main` at `f1339266` into the existing branch as `2c652c1dd75fcdf644f40b4fe8521a9931b5d390`. The only conflict was the ticket index; its resolution preserves both completed-ticket removals. Integrated verification used that merge revision. The final evidence/telemetry commit changes no application or test code.

Independent behavior, data, architecture and test reviews ran under both `gpt-6.1-sol`/high and `gpt-6-luna`/medium. Initial review had six scoped approvals and two data change requests. All eight corrected-candidate rechecks approved fingerprint `5de23fa986778b4e8048c9a42f13321e950e2e4d5b3c0ab36bdc7ee06e65e62b` for the original delivery. Those approvals remain applicable to the AL-81 selection, preservation and registry guards after merge impact inspection: upstream tag handling leaves the manual loader and capture-policy split intact, and publication rollback remains registry independent under the same lock. The integrated gate and repeated native comparison pass; no product conflict resolution or new acceptance boundary was needed. [Retained findings](iterations/01/REVIEW_SUMMARY.json) preserve every candidate, severity, disposition and discovery actor; no unresolved finding or accepted debt remains.

Reviewer-data Sol/high discovered grouping warnings could erase an accepted child link; reviewer-data Luna/medium identified silent discovery skips, concretely verified for nonce scanning after valid native child metadata. The coder now rejects incomplete selected grouping/discovery before publication while tolerating unrelated/unidentifiable corrupt files. Reviewer-architecture Sol/high identified unused manual admission/recovery branches, now removed. The parent independently reproduced the grouping defect; the tester supplied durable preservation/control tests and before-fix discovery proof. Test reviewer Sol independently demonstrated that removing manual locking loses newer updates and removing the grouping guard fails its regression. Verification actors are distinct from discovery credit.

`make` passed completely: Go units/vet/race integration (including the full CLI), 230 frontend units, 781 frontend integration tests, typecheck/lint, installation/distribution/release/orchestration checks and multi-platform builds. The standalone Linux engine and manager reported the tested commit; embedded assets match the canonical build. [Test evidence](iterations/01/TESTING.md) records commands and results.

[Native evidence](iterations/01/NATIVE_EVIDENCE.json) records T3 Code/Codex 0.167.0, root `01a1cbc8-e25f-7b63-8c5e-da429f951a17` and eleven source-proven children. Identical unmodified native rollout copies yielded matching source facts with registry absent/present; absent state stayed absent and present schema-compatible SQLite state was unchanged. Separate cutoffs can change provisional/active durations, excluded from that comparison. This proves manual extraction, not native automatic hook/trust coverage; OpenCode runtime evidence remains isolated SQLite fixtures. Native cost evidence is unavailable and remains explicit. Raw rollouts and private probes stay outside Git. Final telemetry from the configured Codex home was verified against the complete native root/child ID set with no duplicates or source errors. The initial ledger had no earlier AL-81 history; the latest final refresh preserves previously committed owners. Cross-harness preservation remains covered by isolated regression evidence.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 10
- **Active Agent Duration:** 1h 38m (recorded duration including 784s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 269.5k tokens (60% of 450k healthy ceiling, 0 compactions) | Subagents peak 190.9k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.3 avg turns/run (peak: 4 turns) | Planner 2.0, Coder 4.0, Tester 3.0, Reviewer-architecture 2.0, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 3.2:1 (+602 test lines / +187 prod lines)
- **Code Delivered:** +789 / -308 lines across 14 code files (excluding docs)
- **Reasoning Effort:** 67,314 reasoning tokens (27.8% of generated output)
- **Tool Health & Self-Correction:** 561 tool calls (32 self-corrections, 94.3% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (3 CRITICAL, 4 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (17 cycles) — Subagent roles 'coder' and 'tester' alternated for 17 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-architecture (5 cycles) — Subagent roles 'reviewer-data' and 'reviewer-architecture' alternated for 5 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (265.6k tokens) — Context tokens reached 265.6k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (4 consecutive errors) — Tool 'command' encountered 4 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 74,530 | 3,159 | 684,672 | 90.2% | 75.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 126,845 | 12,976 | 7,354,112 | 98.3% | 151.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 106,725 | 27,883 | 12,615,424 | 99.2% | 190.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 718,467 | 85,901 | 36,172,800 | 98.1% | 269.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 2 | 157,045 | 8,288 | 2,303,104 | 93.6% | 109.4k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 320,344 | 10,944 | 3,097,216 | 90.6% | 132.1k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 168,698 | 14,649 | 3,854,976 | 95.8% | 117.7k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 160,340 | 11,185 | 2,648,832 | 94.3% | 121.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **12** | **1,832,994** | **174,985** | **68,731,136** | **97.4%** | **269.5k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 266,043 | 16,211 | 3,322,496 | 92.6% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 8 | 1,566,951 | 158,774 | 65,408,640 | 97.7% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **12** | **1,832,994** | **174,985** | **68,731,136** | **97.4%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |
<!-- agentic-lens:metrics:end -->
