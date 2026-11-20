# fix(AL-96): verify automatic capture and support reversible discard

## Outcome

The installed worker records native activity before ticket binding. After verified binding and destination selection, it publishes the ticket ledger and removes the pending snapshot once its facts are durable. Existing ledgers keep their canonical location and history.

A completed Codex session can bind after its rollout moves into the native archive. The worker verifies the exact session, project, physical store and accepted parentage before saving the new path and binding together. [Archive binding design](design.md#codex-archives-a-session-before-binding)

Users and agents can use `capture discard`, `capture restore`, and `capture status --capture-id` for a verified capture. Discard removes its pending snapshot and suppresses publication for that capture while preserving accepted canonical history and Git staging. Restore rereads the available native history, including activity during discard, and confirms completion after the worker writes the entire verified group. Activity for independent captures continues. The advisory Git hook is tracked in AL-97. [Design](design.md#reversible-discard-of-ad-hoc-capture) · [AL-97](../../../alspaces/plan/to-do/AL-97%20warn%20about%20staged%20pending%20capture%20snapshots.md)

## Verification

The late-binding regression failed at `189a1e75`. The corrected revision `d0ed47e5` passed focused Go race tests for late operator binding and the signed inbox, including foreign project, duplicate owner, changed parent and physically external archive rejections. The two independent reviewers checked the final ownership guards. The integrated `GOMAXPROCS=6 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` passed on the same revision over main at `a5c8a25e`. The prior installed-client evidence below covers the existing discard and restore flows.

The checksummed Linux amd64 package `v3.0.0-rc.0+al96.discard.598d4ce8` passed the installed gate on Codex 0.170.0 and OpenCode 1.19.5. Both clients exercised unbound and bound discard and restore, with exact native owner and duration comparisons, a separate unrelated session, and additional activity during discard. The gate compared canonical ledger and summary bytes from before bound discard and checked exact staged Git entries throughout. [Results](evidence/native-discard-result.json) · [Gate notes](evidence/native-gate.md#integrated-discard-and-restore-package)

Focused Go races covered interrupted cleanup, stale publication and discovery cursors, native credentials, portable history, and restoration after a temporarily unavailable Codex child returned. The Codex fixture restored four exact owners and 59 root input tokens and reproduced the same facts on the next pass. The OpenCode fixture restored a nested group with one source session per pass, and eight terminal sibling roots from one wake. Earlier focused runs observed the same restore continue across two worker processes. Thirteen Python oracle tests and strict OpenSpec validation passed.

The full integrated `GOMAXPROCS=6 VITEST_MAX_FORKS=2 VITEST_MIN_FORKS=1 make -j2` passed at `576c5184` over main at `a5c8a25e`. The preceding run exposed five fixtures that held the legacy exclusive lease to pause current workers. One advisory test also assumed its seeding publication left no lock file. Only test fixtures changed after the native package run. The updated fixtures use current worker election, and the focused legacy exclusion and read-only advisory tests still pass. Independent behavior, architecture, data, and test reviews used gpt-6.2-sol/high and gpt-6-luna/medium. Earlier approvals cover unchanged authorization, recovery, and native oracles. Follow-up reviews covered bounded scheduling, exact cursor fencing and cross-pass materialization. [Review findings](evidence/REVIEW_SUMMARY.json)

## Earlier delivery evidence

The operator-prepared destination gate passed on product `02ff2608` with runner `32fd1347`. Delayed binding and first publication passed on `8588f84d`, and integrated Make passed on `6c178e54`. Those runs remain available with their original support boundaries. [Native gate history](evidence/native-gate.md)

## Limits

The packaged checks cover Linux amd64 and the tested client versions. OpenCode 2.2.0 has earlier manual extraction evidence. The stable release and PR merge remain separate. The existing worktree installation `v3.0.0-rc.0+al96.8588f84d` automatically published this chat's verified Codex root and preserved its 32 previously recorded owners. The root and its 35 recorded descendants matched their native rollout identities and parent links. Exact coverage of separately launched reviews depends on the native records available to the worktree worker. [Development ledger](agentic_metrics.json)

## Pull request

Codex sessions now bind after archival through both operator binding and the authenticated inbox. Recovery checks the exact session and project, store containment and accepted ancestry before recording the archived path with the binding. Focused Go races and integrated `make -j2` passed at `d0ed47e5`.

The installed worker records ticket metrics before binding, keeps pending facts through an unavailable destination, and publishes automatically after verified destination selection. Users and agents can discard a capture and restore its available native history later. The worker preserves accepted canonical history during discard and waits for durable facts from the complete restore group before confirming completion.

The Linux amd64 candidate `v3.0.0-rc.0+al96.discard.598d4ce8` passed with Codex 0.170.0 and OpenCode 1.19.5. Both runs verified exact native owners, durations, Git staging, and canonical ledger and summary preservation through discard. The integrated `make -j2` passed at `576c5184`, which adds test-only fixture fixes to the packaged revision. Focused Go races, thirteen native-oracle tests, and strict OpenSpec validation passed.

## Pull request record

[AgenticLens PR 156](https://github.com/NiltonDuarte/AgenticLens/pull/156)

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 136
- **Active Agent Duration:** 14h 48m (recorded duration including 807s provisional)
- **Context Window Health:** BLOATED — Root agent peak 450.1k tokens (100% of 450k healthy ceiling, 11 compactions) | Subagents peak 451.4k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 2.6 avg turns/run (peak: 8 turns) | Planner 1.0, Coder 3.1, Tester 4.6, Reviewer-architecture 1.8, Reviewer-behavior 2.0, Reviewer-data 2.4, Reviewer-tests 2.0
- **Test-to-Production Code Ratio:** 0.8:1 (+4,717 test lines / +5,609 prod lines)
- **Code Delivered:** +10,326 / -491 lines across 76 code files (excluding docs)
- **Reasoning Effort:** 667,416 reasoning tokens (29.8% of generated output)
- **Tool Health & Self-Correction:** 4,498 tool calls (250 self-corrections, 94.4% success rate)

### Anomalies & Health Flags
- **Status:** 34 execution anomalies detected (15 CRITICAL, 19 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.1k tokens) — Context tokens reached 450.1k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.3k tokens) — Context tokens reached 450.3k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.3k tokens) — Context tokens reached 450.3k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (451.2k tokens) — Context tokens reached 451.2k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (451.3k tokens) — Context tokens reached 451.3k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+132.7k tokens) — Context tokens surged by 132.7k to 132.7k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+402.4k tokens) — Context tokens surged by 402.4k to 402.4k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+416.3k tokens) — Context tokens surged by 416.3k to 416.3k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+416.3k tokens) — Context tokens surged by 416.3k to 416.3k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+425.4k tokens) — Context tokens surged by 425.4k to 425.4k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+437.0k tokens) — Context tokens surged by 437.0k to 437.0k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+301.5k tokens) — Context tokens surged by 301.5k to 301.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+331.6k tokens) — Context tokens surged by 331.6k to 331.6k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.0k tokens) — Context tokens reached 450.0k, breaching the 450k critical ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.8k tokens) — Context tokens reached 253.8k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (265.1k tokens) — Context tokens reached 265.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (402.4k tokens) — Context tokens reached 402.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (416.3k tokens) — Context tokens reached 416.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (412.9k tokens) — Context tokens reached 412.9k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (416.3k tokens) — Context tokens reached 416.3k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (425.4k tokens) — Context tokens reached 425.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (437.0k tokens) — Context tokens reached 437.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (301.5k tokens) — Context tokens reached 301.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (331.4k tokens) — Context tokens reached 331.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (331.6k tokens) — Context tokens reached 331.6k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (331.9k tokens) — Context tokens reached 331.9k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (359.5k tokens) — Context tokens reached 359.5k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (372.6k tokens) — Context tokens reached 372.6k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (322.2k tokens) — Context tokens reached 322.2k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (322.4k tokens) — Context tokens reached 322.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.7k tokens) — Context tokens reached 251.7k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (252.1k tokens) — Context tokens reached 252.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.0k tokens) — Context tokens reached 250.0k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 3 | 805,902 | 12,669 | 9,139,968 | 91.9% | 421.9k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **coder** | 8 | 1,967,207 | 429,984 | 229,503,488 | 99.2% | 451.3k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **tester** | 7 | 1,849,349 | 420,992 | 253,473,792 | 99.3% | 451.4k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **Unknown** | 23 | 3,955,857 | 500,395 | 369,690,624 | 98.9% | 450.1k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 23 unavailable / 23 invocations) |
| **reviewer-architecture** | 5 | 681,655 | 47,997 | 16,603,904 | 96.1% | 240.9k | ELEVATED | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-behavior** | 6 | 956,369 | 66,315 | 23,083,136 | 96.0% | 289.6k | ELEVATED | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **reviewer-data** | 5 | 723,514 | 53,198 | 18,096,512 | 96.2% | 228.8k | ELEVATED | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-tests** | 5 | 509,110 | 38,142 | 9,818,240 | 95.1% | 136.1k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **62** | **11,448,963** | **1,569,692** | **929,409,664** | **98.8%** | **451.4k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 62 unavailable / 62 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 21 | 1,302,179 | 75,577 | 9,090,048 | 87.5% | Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations) |
| **gpt-6.2-sol** | 41 | 10,146,784 | 1,494,115 | 920,319,616 | 98.9% | Unavailable (0 observed, 0 partial, 41 unavailable / 41 invocations) |
| **Total** | **62** | **11,448,963** | **1,569,692** | **929,409,664** | **98.8%** | **Unavailable (0 observed, 0 partial, 62 unavailable / 62 invocations)** |

Capture observation finalized: 2026-11-20T15:07:25.571961159Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 4; retained older facts: 58; pending: 58; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 62; reserved bytes: 536870912. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1f87d-23cf-74d2-832c-2048d28e1726: pending; last successful source read 2026-11-20T15:05:39.685395732Z; accepted source checkpoint 2026-11-20T10:27:11.163Z.
Owner codex/01a1f57e-c71d-7853-8cfd-cd5b570d1010: pending; last successful source read 2026-11-20T15:05:40.392060836Z; accepted source checkpoint 2026-11-19T20:53:31.915Z.
Owner codex/01a1f87c-5810-7ea3-86cf-dc664a25e23c: pending; last successful source read 2026-11-20T15:05:59.156715814Z; accepted source checkpoint 2026-11-20T11:19:59.755Z.
Owner codex/01a1f57d-fa76-7f62-913b-05599a379a47: pending; last successful source read 2026-11-20T15:05:59.901014269Z; accepted source checkpoint 2026-11-19T21:13:02.834Z.
Owner codex/01a1f8bb-3d2b-7733-9837-70218878288d: pending; last successful source read 2026-11-20T15:06:00.536326757Z; accepted source checkpoint 2026-11-20T11:34:09.17Z.
Owner codex/01a1f582-ee58-7a42-b354-86b18f20573e: pending; last successful source read 2026-11-20T15:06:01.27773075Z; accepted source checkpoint 2026-11-19T21:14:44.581Z.
Owner codex/01a1f54a-9c34-7761-a171-c3eb88406a9f: pending; last successful source read 2026-11-20T15:06:20.009957876Z; accepted source checkpoint 2026-11-19T19:36:29.648Z.
Owner codex/01a1f54b-28ce-7361-9943-1445c827c4c1: pending; last successful source read 2026-11-20T15:06:20.929218203Z; accepted source checkpoint 2026-11-19T19:34:11.429Z.
Owner codex/01a1f520-2f74-7ed3-8820-bea7b7d55f08: pending; last successful source read 2026-11-20T15:06:21.660518272Z; accepted source checkpoint 2026-11-19T19:43:39.465Z.
Owner codex/01a1f520-8020-7d53-ac66-56852de9c4ef: pending; last successful source read 2026-11-20T15:06:22.431682694Z; accepted source checkpoint 2026-11-19T19:53:33.551Z.
Owner codex/01a1f871-e22b-74a1-a39a-621213714559: pending; last successful source read 2026-11-20T15:06:41.123123294Z; accepted source checkpoint 2026-11-20T11:01:25.415Z.
Owner codex/01a1f86c-1bb4-7390-b214-a62f617c579a: pending; last successful source read 2026-11-20T15:06:42.020529545Z; accepted source checkpoint 2026-11-20T11:32:21.801Z.
Owner codex/01a1f87c-a752-73d1-b3fa-a4270ea9fe08: pending; last successful source read 2026-11-20T15:06:42.74281665Z; accepted source checkpoint 2026-11-20T11:02:17.686Z.
Owner codex/01a1f591-8636-7f93-858b-65ee553873d7: pending; last successful source read 2026-11-20T15:06:43.485557911Z; accepted source checkpoint 2026-11-19T21:15:07.276Z.
Owner codex/01a1f89e-a268-7aa3-b942-03b10df25297: pending; last successful source read 2026-11-20T15:07:02.022934611Z; accepted source checkpoint 2026-11-20T11:04:07.671Z.
Owner codex/01a1f538-9bff-7581-878d-3ac7d0230e21: pending; last successful source read 2026-11-20T15:07:02.771964831Z; accepted source checkpoint 2026-11-19T19:16:06.629Z.
Additional owner freshness details remain in local capture diagnostics.

Retained recorded history: 58 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
