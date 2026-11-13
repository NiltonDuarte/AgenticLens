# feat(AL-84): support flexible workspace layouts and adopt OpenSpec

## Pull request

AgenticLens now discovers ledgers across explicitly configured roots and follows recorded ticket identity through archive and arbitrary directory moves. First generation supplies a canonical ID and ledger destination; planning files and folder names establish no runtime identity. Writers emit explicit root-only depth 0/budget 300; this repository selects its three containers at depth 1/budget 300. Distinct same-ID ledgers require location selection.

Publication journals and locks use local `agenticlens.state_dir`, separate from capture storage. A transient adjacent pending indicator keeps interrupted sets incomplete when private recovery data is unavailable. Completed supported ledgers remain readable without sources or private state. Reads remain bound to the resolved physical home or return a retry. This is a breaking configuration/API update: adopt `workspaces.roots` and explicit local state; retired scalar keys, runtime root overrides and old journal adapters are rejected.

Pinned OpenSpec 2.0.0 now holds selected execution scope, tasks and evidence. Roadmap intent and priority stay in `alspaces/plan`. Only AL-84 execution artifacts moved; other active owners and historical ledgers remain intact.

Validation: full make passed on `273d90aa16966fe34984b78543fc42cdb6e0c7c0`, including backend race suites, 231 frontend units, 804 frontend integrations, distribution checks and all four native build targets. Three OpenSpec workflow tests and current standalone embedded HTML/API reads passed. Native Codex/OpenCode probes preserved exact owners through archive, external movement and late refresh. OpenCode provider counters were synthetic; native records were real. Native macOS execution and physical power-loss testing were not performed.

Hosted checks and the actual PR are attached at [PR144](https://github.com/NiltonDuarte/AgenticLens/pull/144). The Pages follow-up passed both exact caller commands plus11 Ruby/9 preview tests.

## Verified delivery

The user approved [iteration01](design.md#approved-contract) before implementation. Candidate `623facc7` and test corrections `b4b18cdd`/`273d90aa` integrated main without conflicts; main remained at the original baseline. Core verification content fingerprint: `6af70a82b7335cebe294cd8aafe7e025c18d4f304c6535c0a166661325aa525f`. A subsequent Pages-only caller correction is verified separately; Go/TypeScript and runtime-schema inputs remain unchanged from the passing core gate. The existing roadmap specification moved to done and its active index entry was removed based on verified outcomes; archive and ticket completion remain separate decisions. Its PO-24 navigation link was repaired without changing historical findings.

## Verification and attributed review

Final consolidated review: APPROVED after integrated make. Four core focuses each ran `gpt-6.1-sol`/high and `gpt-6-luna`/medium; initial changes-requested findings were corrected, and matching read-bound rechecks approved. The final six-file test-only delta received both test profiles; all runtime paths were hash-identical, so other approvals carried forward. Source metadata confirmed reporter model/effort. The bounded Pages caller correction received both behavior profiles, with unchanged core scopes carried forward.

[Retained findings](evidence/iterations/01/REVIEW_SUMMARY.json) preserves all eight canonical issues and every reporting tuple. Sol/high behavior, data, architecture and test reviewers found filename admission, physical read identity, private durability, maintenance selection and regression-oracle gaps. Luna/medium behavior found incomplete-scan duplicate publication; Luna/medium data identified the residual same-ID read window. Its original uncorrelated title count was qualified, then a deterministic replacement test and guard mutation established the fix. The parent identified the trusted-base-workflow Pages caller mismatch, retained as CI-01. Verification of supplied issues earns no additional discovery credit.

[Test evidence](evidence/iterations/01/TEST_RESULTS.md) retains actual commands/results, tester discovery credit, earlier failed gates and restart limits. Linux syscall tracing and injected fsync failures verified durable ordering before public changes, not physical power-loss behavior. Native harness proof used manual source/publication; installed automatic-hook execution with real harnesses was not claimed.

## Telemetry

Final extraction verified all 37 source-backed owners: one root, 22 native children and 14 independent delegated review roots. All facts use the actual Codex source; native parent edges are preserved and no T3 parent links are invented. Costs are unavailable. The active root turn remains provisional and its invoking caller unverified, as reported by the source; no completion/rollover was fabricated. Explicit extraction bounds were 128 sessions/128 MiB, based on the observed source set. Raw transcripts, private state and source stores stay outside Git.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 80
- **Active Agent Duration:** 9h 24m (recorded duration including 252s provisional)
- **Context Window Health:** BLOATED — Root agent peak 448.8k tokens (100% of 450k healthy ceiling, 4 compactions) | Subagents peak 450.5k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.0 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 1.7, Tester 2.4, Reviewer-architecture 1.5, Reviewer-behavior 2.5, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 1.5:1 (+3,818 test lines / +2,540 prod lines)
- **Code Delivered:** +6,358 / -2,323 lines across 144 code files (excluding docs)
- **Reasoning Effort:** 418,379 reasoning tokens (30.1% of generated output)
- **Tool Health & Self-Correction:** 2,146 tool calls (183 self-corrections, 91.5% success rate)

### Anomalies & Health Flags
- **Status:** 21 execution anomalies detected (10 CRITICAL, 11 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+293.0k tokens) — Context tokens surged by 293.0k to 293.0k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+321.2k tokens) — Context tokens surged by 321.2k to 321.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.3k tokens) — Context tokens reached 450.3k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (88 cycles) — Subagent roles 'coder' and 'tester' alternated for 88 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (24 cycles) — Subagent roles 'coder' and 'tester' alternated for 24 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (22 cycles) — Subagent roles 'tester' and 'coder' alternated for 22 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (16 cycles) — Subagent roles 'coder' and 'tester' alternated for 16 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (5 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (11 cycles) — Subagent roles 'coder' and 'tester' alternated for 11 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (255.5k tokens) — Context tokens reached 255.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (256.3k tokens) — Context tokens reached 256.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (255.5k tokens) — Context tokens reached 255.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (293.0k tokens) — Context tokens reached 293.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (321.2k tokens) — Context tokens reached 321.2k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.6k tokens) — Context tokens reached 253.6k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-tests (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-tests' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 3 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (4 consecutive errors) — Tool 'command' encountered 4 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 76,885 | 7,226 | 662,400 | 89.6% | 82.3k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 6 | 1,519,432 | 246,642 | 107,852,672 | 98.6% | 450.2k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **tester** | 7 | 1,616,672 | 232,788 | 99,606,528 | 98.4% | 450.5k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **Unknown** | 15 | 2,355,835 | 400,216 | 143,770,368 | 98.4% | 448.8k | ELEVATED | Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations) |
| **reviewer-architecture** | 2 | 286,054 | 15,286 | 3,268,608 | 92.0% | 133.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 540,005 | 20,277 | 4,560,512 | 89.4% | 135.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 316,668 | 21,588 | 5,259,648 | 94.3% | 129.4k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 433,215 | 26,585 | 6,653,312 | 93.9% | 169.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **37** | **7,144,766** | **970,608** | **371,634,048** | **98.1%** | **450.5k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 37 unavailable / 37 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 14 | 999,858 | 58,166 | 8,000,896 | 88.9% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **gpt-6.1-sol** | 23 | 6,144,908 | 912,442 | 363,633,152 | 98.3% | Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations) |
| **Total** | **37** | **7,144,766** | **970,608** | **371,634,048** | **98.1%** | **Unavailable (0 observed, 0 partial, 37 unavailable / 37 invocations)** |

Retained recorded history: 14 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
