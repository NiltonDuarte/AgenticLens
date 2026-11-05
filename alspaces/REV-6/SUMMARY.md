# REV-6: test performance and unit/integration boundaries

[PR #113](https://github.com/NiltonDuarte/AgenticLens/pull/113). Implementation and consolidated review APPROVED at source `8280c2bd3cc5bc03f2f4f3f401cefcc0524bdc4b`; delivered head `2b91574d` adds artifacts only. Publication does not authorize merging.

## Scope and approval

Iteration 01 was a review only, approved 2026-11-04 against `a84adbe5`. Its [contract](CONTRACT.md), [completed plan](iterations/01/PLAN.md) and [compact findings](iterations/01/REVIEW_SUMMARY.json) preserve the original boundaries. Ten reviews covered behavior, data/failure, architecture, tests and performance, each with gpt-6.1-sol/high and gpt-6-luna/medium; all scoped APPROVED after integrated make passed at `48b1b00e`. Local pure Go domain and frontend domain Node command medians were 0.39s and 1.85s; these are historical command cohorts, not a speedup claim.

User-requested main integration brought `40b16def` into the branch. Its make run at `3beb779e` failed because two incoming schema-test fixtures lacked required ticket_id. The user then explicitly approved iteration 02's [contract delta](iterations/02/CONTRACT_DELTA.md) and [plan](iterations/02/PLAN.md) on 2026-11-05. Implementation baseline was `23b38163`; the review-only prohibition is superseded only within that approved delta.

## Delivered behavior

- Repair schema fixtures and assert intended policy/anomaly failures, decoded identity, nil/empty distinctions and valid canonical rewriting; production validators and schemas are unchanged.
- Go units cover internal/domain and its subtree; all complementary packages remain integration. Domain race verification plus complementary package race runs preserve full original Go race coverage.
- Frontend domain and pure utility units run in Node. Complementary tests retain jsdom/setup and per-file annotations; real filesystem asset-sync stays integration.
- Make/CI integration and its setup wait for both Go and frontend unit success, including parallel mixed goals. Standalone npm is frontend-scoped. Shared assets, build-only WEB_DIR overrides and historical selected-release check/test interfaces remain supported.
- No runtime feature changes, dependency/lockfile changes, test deletion, timeout inflation or silent skips.

## Verification and review

Parent orchestrator, coder, tester and planner ran gpt-6.1-sol/high, confirmed from source metadata. Eight independent iteration 02 discovery runs covered behavior, data/failure, architecture and test effectiveness with both gpt-6.1-sol/high and gpt-6-luna/medium. Four paired behavior/architecture correction runs approved source `8280c2bd` without new findings. Both test-effectiveness reviews approved that source.

Behavior gpt-6.1-sol/high initially requested changes at `df3e02c9`; behavior gpt-6-luna/medium initially approved it without findings. Data gpt-6.1-sol/high requested changes solely for the shared release-test blocker and found no blocking data-contract issue; data gpt-6-luna/medium approved `ee82c243`. Their unchanged schema/selector/state-test investigation carries forward explicitly, rather than being relabeled as later child approval. Later Make-path, release-consumer and compiler-inclusion changes were verified by paired behavior/architecture correction review and actual checks. Architecture's two initial approvals were at `ee82c243`. Final consolidated APPROVED followed the complete parent verification.

[Iteration 02 compact findings](iterations/02/REVIEW_SUMMARY.json) preserve all six fixed root causes and every discovery tuple. Orchestrator gpt-6.1-sol/high discovered COORD-01/02/04/05; tester gpt-6.1-sol/high fixed self-contained discovery, and coder gpt-6.1-sol/high fixed compatibility asset ordering, spaced Make paths and compiler inclusion. COORD-03 shares discovery between orchestrator, behavior reviewer and data reviewer, all gpt-6.1-sol/high; tester gpt-6.1-sol/high repaired the release-step consumer. Behavior reviewer gpt-6.1-sol/high discovered REL-01; coder gpt-6.1-sol/high preserved historical release interfaces, verified by tester gpt-6.1-sol/high. Correction reviewers verified supplied issues and receive no additional discovery credit. No iteration 02 blocker remains.

Required full make passed at `8280c2bd` from absent generated frontend/embed assets: 229.67s wall, 425.66s user, 127.14s system, 654448KiB maximum RSS. It covered Go units/race/integration, frontend units/integration, Python release/installer/distribution/orchestration checks (14/10/15/6), strict compiler/lint checks and configured platform builds. Tester proved complete disjoint accounting: 14 Go packages (2 unit,12 complementary) and 853 frontend cases across 37 files (176 unit,677 integration), preserving duplicate multiplicity with no pending cases. Real per-language unit failures blocked integration under Make/npm, including parallel/mixed goals; positive controls reached it. Selected schema-validator overlays and a removed-Make-gate mutation were detected.

Repeated canonical frontend builds and embedded assets matched byte-for-byte. The native binary served those exact assets with Node absent, bounded HTTP/cleanup and unique disposable API identity. Earlier make candidates failed on a stale release-step expectation (`df3e02c9`) and spaced WEB_DIR parsing (`ee82c243`), both fixed before final verification. The first private native probe lacked required configuration thresholds; correcting that fixture required no application change.

Limits: Linux/WSL, Go 1.26, Node 24/npm 11, GNU Make 4.3. Hosted CI/Node20, native Windows, real providers, historical tagged builds and release publication were not locally executed. Outer race instrumentation does not instrument separately built children. No billing saving, general mutation equivalence or whole-repository correctness claim. Historical accepted follow-ups remain in iteration01 compact findings, including AL-74 readiness ownership, optional ACK consolidation and diagnostic-backfill profiling; cleanup does not resolve or hide them.

## Artifact cleanup and telemetry

At the user's request, remove redundant inventories, detailed measurement/verification records and repeated planning narratives. Essential facts are consolidated here; prior committed detail remains in Git history at `2b91574d`. Approved contracts, concise completed plans, both compact finding records, PR description and source ledger remain. No application/test/CI changes; prior source verification remains valid. Links and JSON are checked; no test rerun is needed for artifact-only cleanup. The completed specification remains in plan/done, with no open REV-6 index entry or downstream dependency to promote.

Telemetry uses the source-native CLI, explicit Codex source and isolated manual capture state. Configured state lacks its worktree mapping table and remains untouched. Root source `01a1a7db-7264-7d71-869d-64e1c5906a2f` has real planner/coder/tester delegation edges; 22 independent T3 review roots have unavailable parent/role attribution. The retained ledger has 26 sessions (23 Unknown roles plus planner/coder/tester), preserving all 12 pre-implementation sessions. Costs are unavailable, not zero. Raw rollouts and private state remain outside Git; the final refresh and cutoff follow below.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 56
- **Active Agent Duration:** 4h 8m (excludes user wait time)
- **Context Window Health:** BLOATED — Root agent peak 450.1k tokens (100% of 450k healthy ceiling, 1 compactions) | Subagents peak 407.8k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 4.0 avg turns/run (peak: 5 turns) | Planner 2.0, Coder 5.0, Tester 5.0
- **Test-to-Production Code Ratio:** 37:0 (+37 test lines / +0 prod lines)
- **Code Delivered:** +72 / -24 lines across 7 code files (excluding docs)
- **Reasoning Effort:** 184,996 reasoning tokens (31.7% of generated output)
- **Tool Health & Self-Correction:** 666 tool calls (45 self-corrections, 93.2% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (4 CRITICAL, 3 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.1k tokens) — Context tokens reached 450.1k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+316.9k tokens) — Context tokens surged by 316.9k to 316.9k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+317.5k tokens) — Context tokens surged by 317.5k to 317.5k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (21 cycles) — Subagent roles 'coder' and 'tester' alternated for 21 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (316.9k tokens) — Context tokens reached 316.9k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (317.5k tokens) — Context tokens reached 317.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.4k tokens) — Context tokens reached 251.4k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 167,034 | 13,862 | 1,428,096 | 89.5% | 108.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 407,184 | 33,434 | 22,571,648 | 98.2% | 388.6k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 65,352 | 52,009 | 30,916,096 | 99.8% | 407.8k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 23 | 2,366,434 | 299,173 | 86,653,696 | 97.3% | 450.1k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations) |
| **Total** | **26** | **3,006,004** | **398,478** | **141,569,536** | **97.9%** | **450.1k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 26 unavailable / 26 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 11 | 705,240 | 38,734 | 3,645,440 | 83.8% | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **gpt-6.1-sol** | 15 | 2,300,764 | 359,744 | 137,924,096 | 98.4% | Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations) |
| **Total** | **26** | **3,006,004** | **398,478** | **141,569,536** | **97.9%** | **Unavailable (0 observed, 0 partial, 26 unavailable / 26 invocations)** |

Capture observation finalized: 2026-11-05T17:26:20.705275724Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 26; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 26; reserved bytes: 32609102. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1ab7d-db70-7dc2-a23b-510473ebe1b4: current; last successful source read 2026-11-05T17:26:19.90840856Z; accepted source checkpoint 2026-11-05T11:37:19.427Z.
Owner codex/01a1a7f0-ae81-7e60-90f7-f89c6dde0447: current; last successful source read 2026-11-05T17:26:19.93399747Z; accepted source checkpoint 2026-11-04T19:03:50.831Z.
Owner codex/01a1ab58-b06d-7590-88cb-266601d044f4: current; last successful source read 2026-11-05T17:26:19.961661164Z; accepted source checkpoint 2026-11-05T10:56:55.02Z.
Owner codex/01a1ab75-5d33-7f60-957a-9b63fadf3514: current; last successful source read 2026-11-05T17:26:19.993400806Z; accepted source checkpoint 2026-11-05T11:28:43.449Z.
Owner codex/01a1a7fd-a6d1-7772-abf8-4c9a4adc4863: current; last successful source read 2026-11-05T17:26:20.024793495Z; accepted source checkpoint 2026-11-04T19:18:10.103Z.
Owner codex/01a1a7db-7264-7d71-869d-64e1c5906a2f: current; last successful source read 2026-11-05T17:26:20.27816828Z; accepted source checkpoint 2026-11-05T17:26:19.151Z.
Owner codex/01a1a7fc-317e-7fb1-8d67-ef6b03325642: current; last successful source read 2026-11-05T17:26:20.310392873Z; accepted source checkpoint 2026-11-04T19:16:18.157Z.
Owner codex/01a1a7db-ce2f-7391-b1f2-a646b932ce13: current; last successful source read 2026-11-05T17:26:20.364380435Z; accepted source checkpoint 2026-11-05T00:56:55.201Z.
Owner codex/01a1a7ed-c79d-7e42-9c23-015dc99ca004: current; last successful source read 2026-11-05T17:26:20.40374169Z; accepted source checkpoint 2026-11-04T19:02:07.549Z.
Owner codex/01a1ab80-cede-7d33-8287-451861d732f9: current; last successful source read 2026-11-05T17:26:20.438995346Z; accepted source checkpoint 2026-11-05T11:39:48.275Z.
Owner codex/01a1a7f7-7427-70e2-8900-c647a0f73bc8: current; last successful source read 2026-11-05T17:26:20.467727119Z; accepted source checkpoint 2026-11-04T19:11:22.593Z.
Owner codex/01a1a7f3-c7d5-77f2-adf0-9a6a3c12ffcd: current; last successful source read 2026-11-05T17:26:19.328458637Z; accepted source checkpoint 2026-11-04T19:09:29.222Z.
Owner codex/01a1ab40-744d-7dd1-8597-d46ea3e1d211: current; last successful source read 2026-11-05T17:26:19.408312015Z; accepted source checkpoint 2026-11-05T11:23:12.517Z.
Owner codex/01a1a800-3dcc-78e3-94d6-41511e1016aa: current; last successful source read 2026-11-05T17:26:19.44408841Z; accepted source checkpoint 2026-11-04T19:20:26.566Z.
Owner codex/01a1ab77-e9a0-7e00-b9e7-4a14dcf0f5b9: current; last successful source read 2026-11-05T17:26:19.475768909Z; accepted source checkpoint 2026-11-05T11:29:55.07Z.
Owner codex/01a1ab57-98d2-72b0-8756-bd732e611231: current; last successful source read 2026-11-05T17:26:19.506823983Z; accepted source checkpoint 2026-11-05T10:54:36.788Z.
Additional owner freshness details remain in local capture diagnostics.

Final cleanup telemetry: source-native CLI built from2b91574d, explicit Codex source/root and literal nonce37d074498dca5c12; isolated manual state preserved the incompatible configured state. Generation exited0; independently verified all26 prior Codex session IDs and all three native delegation edges, with both compact finding files unchanged. Cutoff: 2026-11-05T17:26:33.228151+00:00. T3 hierarchy/roles and costs remain unavailable as described above; validation/commit/push follow this snapshot.
