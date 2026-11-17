# feat(AL-84): support flexible workspace layouts and adopt OpenSpec

## Pull request

AgenticLens now discovers ledgers across explicitly configured roots and follows recorded ticket identity through archive and arbitrary directory moves. First generation supplies a canonical ID and ledger destination; planning files and folder names establish no runtime identity. Writers emit explicit root-only depth 0/budget 300; this repository selects its three containers at depth 1/budget 300. Distinct same-ID ledgers require location selection. Publication requires a nonblank owning project; no destination-based project inference or optional old call signature remains.

Publication journals and locks use local `agenticlens.state_dir`, separate from capture storage. A transient adjacent pending indicator keeps interrupted sets incomplete when private recovery data is unavailable. Completed supported ledgers remain readable without sources or private state. Reads and artifact publication remain bound to the selected physical home or return a retry. Pending-ledger uncertainty blocks implicit uniqueness and first-generation absence; discovery and explicit selection share physical-depth limits. This is a breaking configuration/API update: adopt `workspaces.roots` and explicit local state; retired scalar keys, runtime root overrides and old journal adapters are rejected.

Pinned OpenSpec 2.0.0 now holds selected execution scope, tasks and evidence. Roadmap intent and priority stay in `alspaces/plan`. Only AL-84 execution artifacts moved; other active owners and historical ledgers remain intact.

Validation: the integrated gate completed on test-only correction `adf9ece5` across recorded runs: backend race suites with serial Go package scheduling, 231 frontend units, 804 frontend integrations with at most two workers, distribution checks and all four native build targets. Earlier default-concurrency runs hit capture/frontend deadlines; their causes remain unresolved. Full make previously passed on the unchanged production code at `c550193e`. Three OpenSpec workflow tests and current standalone embedded HTML/API reads passed. Native Codex/OpenCode probes preserved exact owners through archive, external movement and late refresh. OpenCode provider counters were synthetic; native records were real. Native macOS execution and physical power-loss testing were not performed. One earlier callback deadline/SQLite failure did not recur in six targeted runs, the broad rerun or integrated make; its cause remains unexplained.

Hosted checks and the actual PR are attached at [PR144](https://github.com/NiltonDuarte/AgenticLens/pull/144). The Pages follow-up passed both exact caller commands plus11 Ruby/9 preview tests. Hosted checks for the new correction commit are separate from the passing local gate.

## Verified delivery

The user approved [iteration01](design.md#approved-contract) before implementation. Candidate `623facc7` and test corrections `b4b18cdd`/`273d90aa` integrated main without conflicts; main remained at the original baseline. Core verification content fingerprint: `6af70a82b7335cebe294cd8aafe7e025c18d4f304c6535c0a166661325aa525f`. The Pages caller correction was verified separately. The later requested Astra audit found three additional defects; correction `9382e92f` integrated unchanged main and passed full make. Its 538-file review fingerprint is `a10e916ca8c02c8552ff8287690c3e9fc6de2ec9531b2197462e77dc16366f20`. The existing roadmap specification moved to done and its active index entry was removed based on verified outcomes; archive and ticket completion remain separate decisions. Its PO-24 navigation link was repaired without changing historical findings.

## Verification and attributed review

Original consolidated review: APPROVED after its integrated make. Four core focuses each ran `gpt-6.1-sol`/high and `gpt-6-luna`/medium; initial changes-requested findings were corrected, and matching read-bound rechecks approved. The final six-file test-only delta received both test profiles; all runtime paths were hash-identical, so other approvals carried forward. Source metadata confirmed reporter model/effort. The bounded Pages caller correction received both behavior profiles, with unchanged core scopes carried forward.

[Retained findings](evidence/iterations/01/REVIEW_SUMMARY.json) preserves all thirteen canonical issues and every reporting tuple. Sol/high behavior, data, architecture and test reviewers found filename admission, physical read identity, private durability, maintenance selection and regression-oracle gaps. Luna/medium behavior found incomplete-scan duplicate publication; Luna/medium data identified the residual same-ID read window. Its original uncorrelated title count was qualified, then a deterministic replacement test and guard mutation established the fix. The parent identified the trusted-base-workflow Pages caller mismatch, retained as CI-01. Verification of supplied issues earns no additional discovery credit.

The user requested a fresh Astra review with fixes. All four focuses ran `gpt-6-astra`/high against `2b41187`, replacing the default paired profiles for this request only; source metadata verifies execution identity. Initial verdicts were CHANGES_REQUESTED. Behavior, architecture and test reviewers independently found pending-ledger identity exclusion (7/10); behavior also proved false first-generation absence. The data reviewer found same-ticket replacement writes during artifact publication (6/10). Architecture found inconsistent symlink depth between discovery and explicit selection (4/10). Severity is normalized to the repository's 0–10 scale. The parent reproduced all three; the coder fixed them and the tester added consequential regressions, including five disposable mutation proofs. All four fresh Astra/high correction rechecks approved the matching candidate without new findings. Final consolidated review is APPROVED after make at `9382e92f`; this is scoped to AL-84 and does not authorize merging.

The unexplained callback deadline/SQLite rollback observation remains visible in test evidence. No timeout or assertion was weakened. Six targeted runs, the broad rerun and integrated make passed; those passes do not establish its cause or prove it unrelated. Parent retains follow-up ownership if it recurs.

[Test evidence](evidence/iterations/01/TEST_RESULTS.md) retains actual commands/results, tester discovery credit, earlier failed gates and restart limits. Linux syscall tracing and injected fsync failures verified durable ordering before public changes, not physical power-loss behavior. Native harness proof used manual source/publication; installed automatic-hook execution with real harnesses was not claimed.

The user then requested removal of the remaining internal call compatibility. Correction `c550193e` makes the owning project mandatory for publication and computed snapshot saving, carries selected project context through relocation/formatting, and updates every caller. The parent directly inspected this bounded signature/caller correction; prior unrelated review scopes carry forward, without claiming new independent reviewer approval. The external-home regression rejects blank selection and verifies that unrelated enclosing configuration cannot supply publication state. Restoring destination inference in a disposable overlay fails the intended assertion. Focused ordinary/race checks and full integrated make passed; main was unchanged. Pages configuration behavior was outside this request.

The hosted Backend Check at `b69df5bb` exposed a test-fixture race: three parallel anomaly subtests rewrote shared YAML while siblings read it. Test-only correction `adf9ece5` moves preparation before `t.Parallel()` and preserves parallel CLI execution and every assertion. The pre-fix repeated race run failed; all 30 corrected repetitions passed. Parent inspection covers this small test-only change; prior runtime reviews carry forward. Two default-concurrency full-gate attempts encountered separate capture deadline failures, retained in the test evidence. The integrated gate completed in stages with serial Go packages and at most two frontend workers, reusing same-candidate passing stages. A frontend deadline failure also remains recorded. No timeout or assertion was changed; this does not establish that default-concurrency execution is reliable.

## Telemetry

Final extraction verified 47 source-backed owners: one root, 22 native children and 24 independent delegated roots, including ten Astra review/correction sessions. All 37 prior owners are retained. All facts use the actual Codex source; native parent edges are preserved and no T3 parent links are invented. Costs are unavailable. The active root turn remains provisional and its invoking caller unverified, as reported by the source; no completion/rollover was fabricated. Explicit extraction bounds were 128 sessions/128 MiB, based on the observed source set. Raw transcripts, private state and source stores stay outside Git.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 119
- **Active Agent Duration:** 11h 18m (recorded duration including 27s provisional)
- **Context Window Health:** BLOATED — Root agent peak 448.8k tokens (100% of 450k healthy ceiling, 6 compactions) | Subagents peak 450.5k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.0 avg turns/run (peak: 4 turns) | Planner 1.0, Coder 1.7, Tester 2.4, Reviewer-architecture 1.5, Reviewer-behavior 2.5, Reviewer-data 2.0, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 1.6:1 (+4,138 test lines / +2,641 prod lines)
- **Code Delivered:** +6,779 / -2,338 lines across 147 code files (excluding docs)
- **Reasoning Effort:** 438,419 reasoning tokens (28.8% of generated output)
- **Tool Health & Self-Correction:** 2,526 tool calls (206 self-corrections, 91.8% success rate)

### Anomalies & Health Flags
- **Status:** 22 execution anomalies detected (11 CRITICAL, 11 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+293.0k tokens) — Context tokens surged by 293.0k to 293.0k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+321.2k tokens) — Context tokens surged by 321.2k to 321.2k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.3k tokens) — Context tokens reached 450.3k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+189.5k tokens) — Context tokens surged by 189.5k to 243.6k tokens.
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
| **Unknown** | 26 | 4,043,207 | 512,348 | 188,223,360 | 97.9% | 448.8k | ELEVATED | Unavailable (0 observed, 0 partial, 26 unavailable / 26 invocations) |
| **reviewer-architecture** | 2 | 286,054 | 15,286 | 3,268,608 | 92.0% | 133.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 540,005 | 20,277 | 4,560,512 | 89.4% | 135.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 316,668 | 21,588 | 5,259,648 | 94.3% | 129.4k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 433,215 | 26,585 | 6,653,312 | 93.9% | 169.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **48** | **8,832,138** | **1,082,740** | **416,087,040** | **97.9%** | **450.5k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 48 unavailable / 48 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-astra** | 11 | 1,288,115 | 101,315 | 40,510,208 | 96.9% | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **gpt-6-luna** | 14 | 999,858 | 58,166 | 8,000,896 | 88.9% | Unavailable (0 observed, 0 partial, 14 unavailable / 14 invocations) |
| **gpt-6.1-sol** | 23 | 6,544,165 | 923,259 | 367,575,936 | 98.3% | Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations) |
| **Total** | **48** | **8,832,138** | **1,082,740** | **416,087,040** | **97.9%** | **Unavailable (0 observed, 0 partial, 48 unavailable / 48 invocations)** |

Retained recorded history: 24 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
