# feat(AL-84): support flexible workspace layouts and adopt OpenSpec

## Pull request

AgenticLens now discovers ledgers across explicitly configured roots and follows recorded ticket identity through archive and arbitrary directory moves. First generation supplies a canonical ID and ledger destination; planning files and folder names establish no runtime identity. Writers emit explicit root-only depth 0/budget 300; this repository selects its three containers at depth 1/budget 300. Distinct same-ID ledgers require location selection.

Publication journals and locks use local `agenticlens.state_dir`, separate from capture storage. A transient adjacent pending indicator keeps interrupted sets incomplete when private recovery data is unavailable. Completed supported ledgers remain readable without sources or private state. Reads remain bound to the resolved physical home or return a retry. This is a breaking configuration/API update: adopt `workspaces.roots` and explicit local state; retired scalar keys, runtime root overrides and old journal adapters are rejected.

Pinned OpenSpec 2.0.0 now holds selected execution scope, tasks and evidence. Roadmap intent and priority stay in `alspaces/plan`. Only AL-84 execution artifacts moved; other active owners and historical ledgers remain intact.

Validation: full make passed on `273d90aa16966fe34984b78543fc42cdb6e0c7c0`, including backend race suites, 231 frontend units, 804 frontend integrations, distribution checks and all four native build targets. Three OpenSpec workflow tests and current standalone embedded HTML/API reads passed. Native Codex/OpenCode probes preserved exact owners through archive, external movement and late refresh. OpenCode provider counters were synthetic; native records were real. Native macOS execution and physical power-loss testing were not performed.

## Verified delivery

The user approved [iteration01](design.md#approved-contract) before implementation. Candidate `623facc7` and test corrections `b4b18cdd`/`273d90aa` integrated main without conflicts; main remained at the original baseline. Final runtime/test content fingerprint: `6af70a82b7335cebe294cd8aafe7e025c18d4f304c6535c0a166661325aa525f`. Final evidence/telemetry changes are artifact-only. The existing roadmap specification moved to done and its active index entry was removed based on verified outcomes; archive and ticket completion remain separate decisions. Its PO-24 navigation link was repaired without changing historical findings.

## Verification and attributed review

Final consolidated review: APPROVED after integrated make. Four core focuses each ran `gpt-6.1-sol`/high and `gpt-6-luna`/medium; initial changes-requested findings were corrected, and matching read-bound rechecks approved. The final six-file test-only delta received both test profiles; all runtime paths were hash-identical, so other approvals carried forward. Source metadata confirmed reporter model/effort.

[Retained findings](evidence/iterations/01/REVIEW_SUMMARY.json) preserves all seven canonical issues and every reporting tuple. Sol/high behavior, data, architecture and test reviewers found filename admission, physical read identity, private durability, maintenance selection and regression-oracle gaps. Luna/medium behavior found incomplete-scan duplicate publication; Luna/medium data identified the residual same-ID read window. Its original uncorrelated title count was qualified, then a deterministic replacement test and guard mutation established the fix. Verification of supplied issues earns no additional discovery credit.

[Test evidence](evidence/iterations/01/TEST_RESULTS.md) retains actual commands/results, tester discovery credit, earlier failed gates and restart limits. Linux syscall tracing and injected fsync failures verified durable ordering before public changes, not physical power-loss behavior. Native harness proof used manual source/publication; installed automatic-hook execution with real harnesses was not claimed.

## Telemetry

Final extraction verified all36 source-backed owners: root1, native children22 and independent delegated review roots13. All facts use the actual Codex source; native parent edges are preserved and no T3 parent links are invented. Costs are unavailable. The active root turn remains provisional and its invoking caller unverified, as reported by the source; no completion/rollover was fabricated. Explicit extraction bounds were128 sessions/128MiB, based on the observed source set. Raw transcripts, private state and source stores stay outside Git.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 77
- **Active Agent Duration:** 8h 53m (recorded duration including 1766s provisional)
- **Context Window Health:** BLOATED — Root agent peak 448.8k tokens (100% of 450k healthy ceiling, 4 compactions) | Subagents peak 450.5k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.9 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 1.5, Tester 2.3, Reviewer-architecture 1.5, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 1.5:1 (+3,818 test lines / +2,540 prod lines)
- **Code Delivered:** +6,358 / -2,323 lines across 144 code files (excluding docs)
- **Reasoning Effort:** 392,085 reasoning tokens (29.8% of generated output)
- **Tool Health & Self-Correction:** 2,065 tool calls (176 self-corrections, 91.5% success rate)

### Anomalies & Health Flags
- **Status:** 20 execution anomalies detected (9 CRITICAL, 11 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+293.0k tokens) — Context tokens surged by 293.0k to 293.0k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+321.2k tokens) — Context tokens surged by 321.2k to 321.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.3k tokens) — Context tokens reached 450.3k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (88 cycles) — Subagent roles 'coder' and 'tester' alternated for 88 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (24 cycles) — Subagent roles 'coder' and 'tester' alternated for 24 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (22 cycles) — Subagent roles 'tester' and 'coder' alternated for 22 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (16 cycles) — Subagent roles 'coder' and 'tester' alternated for 16 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (5 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 5 consecutive cycles.
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
| **coder** | 6 | 1,409,640 | 240,656 | 105,040,384 | 98.7% | 450.2k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **tester** | 7 | 1,592,577 | 224,307 | 97,384,960 | 98.4% | 450.5k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **Unknown** | 14 | 2,283,275 | 370,556 | 132,967,168 | 98.3% | 448.8k | ELEVATED | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **reviewer-architecture** | 2 | 286,054 | 15,286 | 3,268,608 | 92.0% | 133.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 406,117 | 17,173 | 3,787,392 | 90.3% | 126.2k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 316,668 | 21,588 | 5,259,648 | 94.3% | 129.4k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 433,215 | 26,585 | 6,653,312 | 93.9% | 169.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **36** | **6,804,431** | **923,377** | **355,023,872** | **98.1%** | **450.5k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 36 unavailable / 36 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 13 | 951,506 | 56,408 | 7,841,664 | 89.2% | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **gpt-6.1-sol** | 23 | 5,852,925 | 866,969 | 347,182,208 | 98.3% | Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations) |
| **Total** | **36** | **6,804,431** | **923,377** | **355,023,872** | **98.1%** | **Unavailable (0 observed, 0 partial, 36 unavailable / 36 invocations)** |

Retained recorded history: 13 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
