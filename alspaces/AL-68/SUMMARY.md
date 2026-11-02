# AL-68 automatic session capture and publication

Implemented iteration 02: additive Codex/OpenCode lifecycle integrations ask the agent for explicit ticket/title/type, then schedule bounded native publication. Contributing agents need no generate call; manual generation remains recovery. Ledger schema v5 is preserved.

## Delivered behavior

- Domain: validated capture policy, deterministic metadata conflicts and cumulative exact-owner replacement preserve unknown source/cost facts. Completion records observability snapshot finalization independently of ticket or active-session completion.
- Adapters: durable request generations, per-ticket/source cursors, source-proven children, exact physical project/store identity and publication locking coordinate concurrent sessions. Different configured directories remain isolated even with the same ticket/store. Source errors preserve accepted facts and remain failed/pending.
- CLI/process: a periodic/coalesced native worker has idle/lifetime bounds, supported-event restart, disable/uninstall checks, state-drift diagnostics and manual recovery. Authorized journal recovery preserves accepted artifacts across interruption; it does not promise simultaneous visibility or physical power-loss durability.
- Setup/configuration: additive owned installation preserves user hooks and checks handler compatibility with rollback. Shared/local YAML and CLI resolve one policy. Minimal project-local state is default; an explicitly selected external base is project-scoped. Runtime detailed-report support remains optional; this repository disables it and its capture_report.json is absent. No mandatory Git/CI hooks or external daemon/runtime.

## Integration with main

Main 99fbcd7 was integrated through 698aa45 and corrected at 7e09f41. Worktree skill bootstrap and explicit binary repair coexist with automatic capture and protocol-aware rollback. Installation/distribution reviewer gpt-6.1-sol/high discovered MERGE-01: setup rejected workflow RC tags containing +. Both gpt-6.1-sol/high and gpt-6-luna/medium verified the correction, installed-bundle/repair oracle and unsafe-tag rejection. The initial Luna approval and Sol changes-requested outcome remain historical, separate from later verification.

Main 8f0fd4f (GEN-16 compact review retention) was merged through 6a6b3cc. Its REVIEW_SUMMARY.json format replaces this branch's interim Markdown-only retention policy. Guidance preserves private working evidence, explicit model attribution and readable candidate revisions without mandatory report/prompt/policy hashes. This latest merge and adaptation change documentation only; existing integrated application verification carries forward.

## Review and verification

Iteration 01: all 12 confirmed CAP findings fixed on candidate-04. Paired behavior, data, architecture, tests, security and release reviewers (gpt-6.1-sol/high and gpt-6-luna/medium) verified their assigned corrections. Their final overall INCOMPLETE verdicts reflected missing real Codex evidence; the parent subsequently closed the required observed-path gate with the user's normal trust review and real exec/resume smoke. Final scoped outcome: APPROVED. Integrated make passed at 9db4672 and ba0b50a, including Go vet/race, frontend typecheck/lint/build, 781 tests / 35 files, 14 distribution checks and native builds.

Iteration 02: 19 confirmed AUTO defects/oracle gaps fixed and AUTO-05/AUTO-10 refuted. Final scoped outcome: APPROVED on candidate-02-07 (de31cb1), implementation 55b4dd4. Behavior/data/architecture/security paired approvals carry from candidate 06 over unchanged boundaries; release/tests paired approvals cover candidate 07. Verification actors are the dedicated reviewer-behavior/data/architecture/tests personas and generic reviewer with security or release responsibility, each on gpt-6.1-sol/high and gpt-6-luna/medium. These verification assignments do not grant discovery credit. Final integrated make at 7e09f41 passed backend race/vet, frontend typecheck/lint/build and 781 tests / 35 files, 14 installation checks, 15 capture-distribution checks, 6 release-workflow checks and Linux/Darwin builds.

Real runtime evidence is bounded to Linux Codex 0.166.0 and OpenCode 1.19.5. Iteration 01 verified Codex callback/context/request/agent-bind exchanges, independent roots and a completed native child after human trust review. Corrected iteration 02 evidence includes three Codex roots/two completed children and two OpenCode same-ticket/store directories with disjoint 12/3 owners. Source project/session/parent/model/usage and available costs were compared; contributing agents did not run generate. No trust bypass was used.

Codex SubagentStart delivery, active/interrupted children before Stop, other versions/interfaces/platforms (including OpenCode v2), actual forty-minute elapsed operation and physical power-loss behavior remain unverified. Synthetic growth/crash fixtures and cross-builds do not establish those runtime claims. Supported scope does not promise universal harness capture or protection from hostile concurrent filesystem mutation.

## Compact review retention

Permanent findings: [iteration 01 REVIEW_SUMMARY.json](iterations/01/REVIEW_SUMMARY.json) and [iteration 02 REVIEW_SUMMARY.json](iterations/02/REVIEW_SUMMARY.json). Together they retain 34 candidates, all 49 distinct reporting persona/model/effort tuples, shared discovery and the two explicit refutations. All discovery identities are source-observed; specialist annotations resolve to the actual generic reviewer persona. CAP-12, AUTO-16–21 and MERGE-01 retain their re-review/integration discovery origin. Verification actors remain above, separate from discovery.

Severity was individually reassessed from documented supported impact and reachable triggers, not translated mechanically from P1/P2. Scores describe the historical defect before correction, not remaining risk: data corruption/ownership or unsafe writes score 8; capture loss, store-boundary violations or identity failures score 7; operational/configuration disruption scores 4–6; contradictory guidance scores 3. AUTO-05 and AUTO-10 score 0 because the claimed defect was refuted, not because a confirmed defect was fixed. No unsupported remote-exploit or catastrophic-loss claim was used to inflate scores.

The following compact history accounts for all 122 runs, including no-new-findings approvals and cancelled/incomplete work. Cancellation is no approval; requested settings for unobserved cancelled runs remain unverified. Detailed per-run identities/evidence and original outcomes remain in Git history at b56e5a0 (consolidated reviews), 0360839 (finding actors) and 0e9cb88 (original full reports). These are historical references, not required runtime files.

| Reviewed candidate | Changes requested | Approved | Completed incomplete | Cancelled |
| --- | ---: | ---: | ---: | ---: |
| iteration 01 candidate-01 | 9 | 0 | 3 | 0 |
| iteration 01 candidate-02 | 3 | 0 | 0 | 9 |
| iteration 01 candidate-03 | 2 | 0 | 1 | 9 |
| iteration 01 candidate-04 | 0 | 0 | 12 | 0 |
| candidate-02-01 | 10 | 2 | 0 | 0 |
| candidate-02-02 | 1 | 3 | 0 | 8 |
| candidate-02-03 | 2 | 2 | 0 | 6 |
| candidate-02-04 | 2 | 1 | 0 | 7 |
| candidate-02-05 | 1 | 3 | 0 | 6 |
| candidate-02-06 | 1 | 10 | 0 | 1 |
| candidate-02-07 | 0 | 4 | 0 | 0 |
| 698aa45 | 1 | 1 | 0 | 0 |
| 7e09f41 | 0 | 2 | 0 | 0 |

Completed verbose REVIEW.md files and bulk archives are removed after validating compact findings and attribution. Historical unrelated ticket artifacts remain untouched. Parent verification checks exact finding IDs, all reporter tuples, refutations, numeric severity range, JSON parsing, references and skill validity. Runtime source is unchanged, so the integrated make evidence above is reused.

Post-GEN-16 integration verification: full `make` passed on 6a6b3cc plus this documentation maintenance diff: backend race/vet, frontend typecheck/lint/build, 781 tests in 35 files, 6 release-workflow, 14 installation and 15 capture-distribution checks, and all four Linux/Darwin builds. Both skill validators, exact finding/reporter preservation checks and whitespace/current-link checks pass. No application source changed.

## Telemetry and remaining scope

The latest verified snapshot contains 93 same-project Codex owners (67 roots/26 native children), preserving all 93 prior owners. Direct source-prefix model/usage/project/parent comparisons pass. Seven prior foreign-directory smoke owners remain historical; they are not relabelled into this project's ledger. Coverage states remain conservative (55 pending / 38 current), with no failed/unassigned registrations. Pending alone does not prove missing published metrics. Work after the final snapshot is outside its boundary; raw transcripts remain outside Git.

Original CONTRACT remains frozen. Iteration 03 is an unapproved draft: shared user-PATH wrapper, YAML version pin, user-level native cache, summary-only runtime output, YAML state defaults/local path overrides, optional warning-only Git hook and early registration/setup verification are not implemented here. This branch's local detailed-report opt-out/removal does not claim the broader iteration 03 runtime removal is delivered.

## Follow-up backlog

AL-76 defines compact ledger JSON serialization and a source-independent format command with all-workspace, single-ticket and ticket-set selection, dry-run/check modes, value preservation and safe concurrent publication. Its specification is Ready under plan/to-do. Implementation requires its own dedicated session; no format command is implemented in AL-68.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 246
- **Active Agent Duration:** 17h 17m (excludes user wait time)
- **Context Window Health:** CRITICAL — Root agent peak 454.1k tokens (227% of 200k healthy ceiling, 3 compactions) | Subagents peak 348.1k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.5 avg turns/run (peak: 6 turns) | Planner 1.3, Coder 3.0, Tester 6.0, Reviewer 1.2, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.1:1 (+5,291 test lines / +4,684 prod lines)
- **Code Delivered:** +10,026 / -49 lines across 56 code files (excluding docs)
- **Reasoning Effort:** 612,235 reasoning tokens (27.4% of generated output)
- **Tool Health & Self-Correction:** 3,043 tool calls (150 self-corrections, 95.1% success rate)

### Anomalies & Health Flags
- **Status:** 31 execution anomalies detected (16 CRITICAL, 15 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (203.5k tokens) — Context tokens reached 203.5k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (207.3k tokens) — Context tokens reached 207.3k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+300.5k tokens) — Context tokens surged by 300.5k to 300.5k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (300.5k tokens) — Context tokens reached 300.5k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (207.2k tokens) — Context tokens reached 207.2k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (203.6k tokens) — Context tokens reached 203.6k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.4k tokens) — Context tokens reached 200.4k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.5k tokens) — Context tokens reached 200.5k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (336.6k tokens) — Context tokens reached 336.6k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (303.6k tokens) — Context tokens reached 303.6k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (453.9k tokens) — Context tokens reached 453.9k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (210.0k tokens) — Context tokens reached 210.0k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (236.5k tokens) — Context tokens reached 236.5k, breaching the 200k critical ceiling.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (348.1k tokens) — Context tokens reached 348.1k, breaching the 200k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> reviewer-tests (8 cycles) — Subagent roles 'reviewer' and 'reviewer-tests' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer <-> reviewer-tests (7 cycles) — Subagent roles 'reviewer' and 'reviewer-tests' alternated for 7 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.3k tokens) — Context tokens reached 160.3k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (174.0k tokens) — Context tokens reached 174.0k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (190.7k tokens) — Context tokens reached 190.7k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (199.1k tokens) — Context tokens reached 199.1k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.1k tokens) — Context tokens reached 161.1k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (163.4k tokens) — Context tokens reached 163.4k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (161.0k tokens) — Context tokens reached 161.0k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (174.3k tokens) — Context tokens reached 174.3k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-architecture (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-architecture' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-architecture (4 cycles) — Subagent roles 'reviewer-data' and 'reviewer-architecture' alternated for 4 consecutive cycles.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.
- [WARNING] **TOOL_RETRY_STREAK**: Tool retry streak: command (3 consecutive errors) — Tool 'command' encountered 3 consecutive execution failures.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 3 | 814,398 | 26,345 | 4,034,944 | 83.2% | 236.5k | BLOATED (118%) | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **coder** | 3 | 1,306,732 | 155,437 | 54,191,872 | 97.6% | 348.1k | BLOATED (174%) | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **tester** | 1 | 692,408 | 81,990 | 35,824,384 | 98.1% | 303.6k | BLOATED (152%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 5 | 463,052 | 36,239 | 5,253,760 | 91.9% | 105.5k | ELEVATED | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-architecture** | 4 | 524,799 | 31,831 | 4,718,080 | 90.0% | 106.1k | ELEVATED | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 4 | 476,895 | 27,453 | 4,110,976 | 89.6% | 98.6k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-data** | 4 | 383,552 | 32,720 | 4,905,856 | 92.7% | 99.3k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-tests** | 2 | 236,275 | 13,714 | 2,179,328 | 90.2% | 103.7k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **rootagent** | 67 | 9,948,517 | 1,214,038 | 255,486,336 | 96.3% | 454.1k | CRITICAL (227%) | Unavailable (0 observed, 0 partial, 67 unavailable / 67 invocations) |
| **Total** | **93** | **14,846,628** | **1,619,767** | **370,705,536** | **96.1%** | **454.1k** | **CRITICAL** | **Unavailable (0 observed, 0 partial, 93 unavailable / 93 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 32 | 2,557,710 | 162,470 | 18,918,272 | 88.1% | Unavailable (0 observed, 0 partial, 32 unavailable / 32 invocations) |
| **gpt-6.1-sol** | 61 | 12,288,918 | 1,457,297 | 351,787,264 | 96.6% | Unavailable (0 observed, 0 partial, 61 unavailable / 61 invocations) |
| **Total** | **93** | **14,846,628** | **1,619,767** | **370,705,536** | **96.1%** | **Unavailable (0 observed, 0 partial, 93 unavailable / 93 invocations)** |

Capture reconciliation completed: 2026-11-02T22:22:36.676329638Z (per-owner refreshed_at checkpoints; no common as-of snapshot). Current: 36; pending: 57; failed: 0; unassigned: 0. Detailed report disabled; any existing report is stale. Source failures preserve prior facts; coverage is limited to registered owners. Hook installation alone does not prove complete capture.
