# feat(capture): capture OpenCode sessions across Git worktrees

## Pull request

OpenCode users can install a single capture plugin with `capture install-user --source opencode`. Each callback supplies its native session context to the installed manager. The selected engine applies the checkout's capture configuration and uses the shared worker to publish bound sessions. Operators can preview installation, explicitly migrate an unchanged legacy project plugin and remove the owned global plugin.

Verification includes native OpenCode 1.19.5 with a server outside the session worktree, a root binding through the plugin tool, a child created through OpenCode's task tool, and worker-produced ticket output. The local test provider supplied deterministic model responses. OpenCode supplied both session identities and their relationship. Integrated `make` passed. OpenCode 2.2.0 supported manual source extraction but did not discover the automatic plugin.

## Verification and review

The integrated candidate is `2f794e38` on `b6ab0471`. The native run verified three simultaneous roots and three children, including concurrent contributors to one ticket and a different ticket in the base checkout. Resuming from a disabled sibling delivered all 11 observed callbacks to the original worktree. A copied mixed-harness ledger retained its OpenCode owner while an exact Codex refresh replaced its token count. [Native verification](evidence/iterations/01/native-verification.md)

The behavior, data, architecture and security reviewers on `gpt-6.1-sol/high` independently found that legacy installation could add a duplicate while the default user plugin remained loaded beside a custom config directory. The architecture reviewer on the same model also demonstrated that a migrated plugin in the default directory could bind a session without scheduling publication. Both regressions failed on the initial candidate and passed after the installer and worker began checking both loaded homes. Native resume evidence resolved the `gpt-6-luna/medium` architecture reviewer's question about routing callbacks that carry only a session ID. The behavior, data, architecture, test and security re-reviews accepted the corrected candidate under both profiles.

Automatic capture was demonstrated on OpenCode 1.19.5. OpenCode 2.2.0 read its v2 source successfully through manual generation and did not discover this plugin. The installer can inspect the configured repository, its base and the current runtime's loaded homes. An unrelated checkout with a legacy plugin reports a conflict at callback admission and requires explicit migration. With no authoritative runtime database path, the integration cannot independently compare that path to the configured store.

The preliminary full build hit deadline failures during concurrent backend suites on this machine. The integrated `GOMAXPROCS=4 GOFLAGS="-p=1 -timeout=35m" make` passed. It ran Go race tests, installation and capture distribution checks, frontend type checking and lint, 231 frontend unit tests, 804 frontend integration tests, and built the manager and engine for Linux and macOS on amd64 and arm64. Consolidated review approved the candidate. [Retained review findings](evidence/iterations/01/REVIEW_SUMMARY.json)

Telemetry includes the Codex root, its nine native children and ten separately rooted tool-delegated reviews. The source reported the current turn as open and costs as unavailable. A temporary worktree-local configuration supplied the required publication state directory while retaining the base checkout’s disabled capture setting.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 35
- **Active Agent Duration:** 2h 13m (recorded duration including 2970s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 152.6k tokens (34% of 450k healthy ceiling, 0 compactions) | Subagents peak 145.1k (Reviewer-architecture)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.8 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 1.5, Tester 2.0, Reviewer 2.0, Reviewer-architecture 2.0, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 0.3:1 (+251 test lines / +832 prod lines)
- **Code Delivered:** +1,083 / -123 lines across 13 code files (excluding docs)
- **Reasoning Effort:** 41,262 reasoning tokens (18.6% of generated output)
- **Tool Health & Self-Correction:** 729 tool calls (54 self-corrections, 92.6% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (2 CRITICAL, 5 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: planner <-> tester (5 cycles) — Subagent roles 'planner' and 'tester' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (36 cycles) — Subagent roles 'coder' and 'tester' alternated for 36 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+80.2k tokens) — Context tokens surged by 80.2k to 126.2k tokens.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> tester (3 cycles) — Subagent roles 'reviewer-behavior' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> tester (3 cycles) — Subagent roles 'reviewer-data' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (3 cycles) — Subagent roles 'tester' and 'coder' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 57,010 | 3,086 | 340,864 | 85.7% | 55.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 200,647 | 32,052 | 10,413,440 | 98.1% | 135.0k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 130,235 | 25,409 | 12,072,576 | 98.9% | 142.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 1 | 84,807 | 8,421 | 2,340,864 | 96.5% | 99.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 11 | 948,874 | 65,275 | 28,363,392 | 96.8% | 152.6k | HEALTHY | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **reviewer-architecture** | 1 | 124,934 | 14,071 | 4,253,184 | 97.1% | 145.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 106,710 | 9,669 | 2,770,688 | 96.3% | 109.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 113,011 | 14,924 | 4,472,576 | 97.5% | 133.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 87,858 | 7,720 | 1,680,768 | 95.0% | 104.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **20** | **1,854,086** | **180,627** | **66,708,352** | **97.3%** | **152.6k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 20 unavailable / 20 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 10 | 629,972 | 37,627 | 4,849,792 | 88.5% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **gpt-6.1-sol** | 5 | 517,320 | 54,805 | 15,518,080 | 96.8% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **gpt-6.2-astra** | 5 | 706,794 | 88,195 | 46,340,480 | 98.5% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **20** | **1,854,086** | **180,627** | **66,708,352** | **97.3%** | **Unavailable (0 observed, 0 partial, 20 unavailable / 20 invocations)** |

Retained recorded history: 10 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
