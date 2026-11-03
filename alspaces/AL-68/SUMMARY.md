# AL-68 automatic session capture and publication

Additive Codex/OpenCode lifecycle integrations ask the agent for explicit ticket/title/type, then schedule bounded native publication. Contributing agents need no generate call; manual generation remains recovery. Ledger schema v5 is preserved. The user approved iteration 03 implementation on 2026-11-03. Code review and integrated make passed; the new Codex human-trust smoke remains the delivery gate.

## Delivered behavior

- Domain: validated capture policy, deterministic metadata conflicts and cumulative exact-owner replacement preserve unknown source/cost facts. Completion records observability snapshot finalization independently of ticket or active-session completion.
- Adapters: durable request generations, per-ticket/source cursors, source-proven children, exact physical project/store identity and publication locking coordinate concurrent sessions. Different configured directories remain isolated even with the same ticket/store. Source errors preserve accepted facts and remain failed/pending.
- CLI/process: a periodic/coalesced native worker has idle/lifetime bounds, supported-event restart and validated version takeover, including acknowledged active sessions. Read-only status/diagnostics/verify expose accepted snapshots, terminal failures and bounded history. Pending retries cannot acknowledge source or collection failures. Authorized journal recovery preserves accepted artifacts across interruption.
- Setup/configuration: one user-level wrapper selects the authoritative shared YAML version and executing platform. Explicit installation preserves side-by-side native identities; local YAML overrides paths. Handler capability2 gates new glue while registry/manifest protocol1 remains readable. Source-build Make/CI commands run the newly built native directly. Shared/local/CLI policy declares project-local state by default, with an explicit external project-scoped option.
- Publication/diagnostics: new output is ledger plus summary. The local database retains diagnostic facts and indexed current heads; legacy metadata stays visibly unknown until bounded explicit reconciliation. Recovery validates physical project authority for an old report journal. Retired detailed_report is deprecated/no-effect; no capture_report.json is generated.
- Optional Git advisory: explicit install-check/uninstall-check preserve arbitrary existing hooks. The check resolves the committing worktree, warns on proven capture issues, honors local disablement and remains bounded/nonblocking.

## Integration with main

Main 99fbcd7 was integrated through 698aa45 and corrected at 7e09f41. Worktree skill bootstrap and explicit binary repair coexist with automatic capture and protocol-aware rollback. Installation/distribution reviewer gpt-6.1-sol/high discovered MERGE-01: setup rejected workflow RC tags containing +. Both gpt-6.1-sol/high and gpt-6-luna/medium verified the correction, installed-bundle/repair oracle and unsafe-tag rejection. The initial Luna approval and Sol changes-requested outcome remain historical, separate from later verification.

Main 8f0fd4f (GEN-16 compact review retention) was merged through 6a6b3cc. Its REVIEW_SUMMARY.json format replaces this branch's interim Markdown-only retention policy. Guidance preserves private working evidence, explicit model attribution and readable candidate revisions without mandatory report/prompt/policy hashes. That merge and retention adaptation changed documentation only; its existing integrated application verification carried forward.

Main b1742e9 (GEN-17) was merged without conflicts through 31d092e from a clean a14cbdf checkout. It preserves source personas and Unknown when unavailable, independently of root/child ancestry. Incoming Codex/domain/live source remains byte-identical to main; automatic merges retain AL-68 ownership/source-store checks and capture extraction. No hand-written product fix or iteration 03 implementation was needed. The standalone Linux binary's index and both referenced assets match the canonical web/dist and embedded asset tree.

## Review and verification

Iteration 01: all 12 confirmed CAP findings fixed on candidate-04. Paired behavior, data, architecture, tests, security and release reviewers (gpt-6.1-sol/high and gpt-6-luna/medium) verified their assigned corrections. Their final overall INCOMPLETE verdicts reflected missing real Codex evidence; the parent subsequently closed the required observed-path gate with the user's normal trust review and real exec/resume smoke. Final scoped outcome: APPROVED. Integrated make passed at 9db4672 and ba0b50a, including Go vet/race, frontend typecheck/lint/build, 781 tests / 35 files, 14 distribution checks and native builds.

Iteration 02: 19 confirmed AUTO defects/oracle gaps fixed and AUTO-05/AUTO-10 refuted. Final scoped outcome: APPROVED on candidate-02-07 (de31cb1), implementation 55b4dd4. Behavior/data/architecture/security paired approvals carry from candidate 06 over unchanged boundaries; release/tests paired approvals cover candidate 07. Verification actors are the dedicated reviewer-behavior/data/architecture/tests personas and generic reviewer with security or release responsibility, each on gpt-6.1-sol/high and gpt-6-luna/medium. These verification assignments do not grant discovery credit. Final integrated make at 7e09f41 passed backend race/vet, frontend typecheck/lint/build and 781 tests / 35 files, 14 installation checks, 15 capture-distribution checks, 6 release-workflow checks and Linux/Darwin builds.

Real runtime evidence is bounded to Linux Codex 0.166.0 and OpenCode 1.19.5. Iteration 01 verified Codex callback/context/request/agent-bind exchanges, independent roots and a completed native child after human trust review. Corrected iteration 02 evidence includes three Codex roots/two completed children and two OpenCode same-ticket/store directories with disjoint 12/3 owners. Source project/session/parent/model/usage and available costs were compared; contributing agents did not run generate. No trust bypass was used.

Codex SubagentStart delivery, active/interrupted children before Stop, other versions/interfaces/platforms (including OpenCode v2), actual forty-minute elapsed operation and physical power-loss behavior remain unverified. Synthetic growth/crash fixtures and cross-builds do not establish those runtime claims. Supported scope does not promise universal harness capture or protection from hostile concurrent filesystem mutation.

GEN-17 integration review: reviewer-data gpt-6.1-sol/high and gpt-6-luna/medium both APPROVED the scoped merge 31d092e with no findings (dispatch-acknowledged execution settings). This narrower review reuses GEN-17's completed paired behavior/data/architecture/test evidence for unchanged incoming code; fresh data/failure inspection covers its interaction with AL-68 capture, and full make verifies the union. Security/release paths are unchanged. Full make passed: backend race/vet, frontend typecheck/lint/build and 783 tests / 35 files, 6 release-workflow, 14 installation and 15 capture-distribution checks, four Linux/Darwin builds. No new live harness/platform smoke is claimed. Conservative accepted snapshots may retain historical roles at unchanged/equal checkpoints; the merge does not authorize bulk rewriting historical facts.

Iteration 03 plan-only review: gpt-6.1-sol/high architecture review approved with no findings; gpt-6-luna/medium requested baseline correction (PLAN3-01). Parent gpt-6.1-sol/high verified it and found stale telemetry text (PLAN3-02). Both documentation findings are fixed; see [compact findings](iterations/03/REVIEW_SUMMARY.json). At that historical checkpoint, the four recent plan/integration sessions were outside the 93-owner verification claim, implementation was not approved, and bootstrap/Windows demonstrations were pending. Later explicit approval and implementation verification are recorded below.

Iteration 03 implementation review covers candidate05 on the a051ada baseline. Behavior, data and architecture gpt-6.1-sol/high and gpt-6-luna/medium all APPROVED its final two-condition producer correction. Paired test-effectiveness approval carries from candidate04 with the additive actual parent-conflict before/after oracle verified by the parent; no test deletion or fixture-framework change. Paired security/release approval carries from candidate03: source authority, physical paths, process/argv handling, hooks, installers, packaging and sanitized fixed feedback are unchanged. Full working evidence stays private while runtime acceptance is incomplete. [Compact iteration 03 findings](iterations/03/REVIEW_SUMMARY.json) retain every candidate, refutation and reporting persona/model/effort tuple.

The initial full paired discovery found unsafe legacy recovery (data/security Sol), stale repair and missing-integration diagnostics (behavior Sol), unbounded diagnostic retrieval (behavior/data Luna), cache identity changes (architecture Sol), omitted Windows watchdog CI (tests Sol), and source-build/scaffold/Windows setup failures (release Sol, independently security Sol for Windows output). Subsequent data/behavior/architecture Sol re-reviews exposed same-ticket pre-budget scans, pending failure suppression, fatal collection/owner ambiguity and post-read parent-conflict suppression. Shared discoveries retain all actor tuples. Coder/tester/parent findings identify their implementation or verification origin separately; supplied-finding verification does not receive discovery credit. Owning source-session metadata confirms model/effort for all 42 implementation review runs.

| Iteration 03 candidate | Changes requested | Scoped approved | Incomplete |
| --- | ---: | ---: | ---: |
| 01, full paired discovery | 8 | 3 | 1 |
| 02, paired behavior/data correction | 3 | 1 | 0 |
| 03, all paired focuses | 1 | 11 | 0 |
| 04, paired core correction | 3 | 5 | 0 |
| 05, paired producer closure | 0 | 6 | 0 |

Final source gates passed domain/capture races (1.149s/35.667s) and full Capture CLI races (69.016s); original three parent-conflict probes pass while the prior producer fails the canonical oracle. POSIX distribution15, installation10 and release6 checks pass. Native Windows PowerShell/PE coverage passed eight valid full-run cases plus the corrected scaffold case; Git/Node owned-tree timeout evidence remains valid on unchanged paths. Those are platform-component results, not live Windows harness certification.

Implementation commit 8d86704 was clean before fetching/merging main, which was already up to date. Integrated make exited0: all backend vet/race packages, frontend typecheck/lint/build,783 tests in35files, release6/installation10/distribution15 checks and all four Linux/Darwin native builds. The embedded-asset adapter checks passed; generated frontend/embedded trees and binaries remain ignored. No product source changed during integration.

Final development-native OpenCode1.19.5 smoke passed on Linux: actual hook challenge, agenticlens_bind, background ledger/summary and source project/parent/model/raw-role/tokens/tools/reported-cost comparison. No generate call or report sidecar occurred; the optional check was quiet. An explicit version update preserved earlier native identities and exact hook/plugin contents. New Codex definitions are installed at /tmp/al68-iter03-codex-y_cxnj4y, awaiting the user's normal project/hook trust review and actual agent exchange. Earlier iteration Codex trust does not approve these definitions; no bypass was attempted.

The upgraded older OpenCode fixture also preserved accepted totals through explicit bounded diagnostic maintenance: three retained rows/2527bytes indexed, current metadata became complete, and old binding evidence correctly remained insufficient to certify the new installation. Fresh final-native state required no maintenance.

Fixture correction disclosure: Linux and Windows probes each created an identified synthetic native outside the intended test cache before isolation was corrected. Only owned artifacts/empty version parents were removed; user wrappers, profiles, trust, state and unrelated versions were preserved. Windows cleanup used matching build metadata/timestamp; exact hash evidence was not retained. Explicit fixture cache roots and canonical isolation guidance now protect these paths.

## Compact review retention

Permanent findings: [iteration 01 REVIEW_SUMMARY.json](iterations/01/REVIEW_SUMMARY.json) and [iteration 02 REVIEW_SUMMARY.json](iterations/02/REVIEW_SUMMARY.json). Together these two completed implementation reviews retain 34 candidates, all 49 distinct reporting persona/model/effort tuples, shared discovery and the two explicit refutations. All discovery identities are source-observed; specialist annotations resolve to the actual generic reviewer persona. CAP-12, AUTO-16–21 and MERGE-01 retain their re-review/integration discovery origin. Verification actors remain above, separate from discovery.

Severity was individually reassessed from documented supported impact and reachable triggers, not translated mechanically from P1/P2. Scores describe the historical defect before correction, not remaining risk: data corruption/ownership or unsafe writes score 8; capture loss, store-boundary violations or identity failures score 7; operational/configuration disruption scores 4–6; contradictory guidance scores 3. AUTO-05 and AUTO-10 score 0 because the claimed defect was refuted, not because a confirmed defect was fixed. No unsupported remote-exploit or catastrophic-loss claim was used to inflate scores.

The following compact history accounts for all 122 historical implementation runs, including no-new-findings approvals and cancelled/incomplete work. Cancellation is no approval; requested settings for unobserved cancelled runs remain unverified. Detailed per-run identities/evidence and original outcomes remain in Git history at b56e5a0 (consolidated reviews), 0360839 (finding actors) and 0e9cb88 (original full reports). These are historical references, not required runtime files.

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

Completed verbose REVIEW.md files and bulk archives are removed after validating compact findings and attribution. Historical unrelated ticket artifacts remain untouched. Parent verification checks exact finding IDs, all reporter tuples, refutations, numeric severity range, JSON parsing, references and skill validity. The earlier compaction left runtime source unchanged; later integration verification is recorded above.

Post-GEN-16 integration verification: full `make` passed on 6a6b3cc plus this documentation maintenance diff: backend race/vet, frontend typecheck/lint/build, 781 tests in 35 files, 6 release-workflow, 14 installation and 15 capture-distribution checks, and all four Linux/Darwin builds. Both skill validators, exact finding/reporter preservation checks and whitespace/current-link checks pass. No application source changed.

## Telemetry and remaining scope

The latest verified snapshot contains 93 same-project Codex owners (67 roots/26 native children), preserving all 93 prior owners. Direct source-prefix model/usage/project/parent comparisons pass. Seven prior foreign-directory smoke owners remain historical; they are not relabelled into this project's ledger. Coverage states remain conservative (55 pending / 38 current), with no failed/unassigned registrations. Pending alone does not prove missing published metrics. Work after the final snapshot is outside its boundary; raw transcripts remain outside Git.

Original CONTRACT remains frozen. Iteration 03 implementation, current scoped code reviews, main integration and make are complete. Overall acceptance is INCOMPLETE pending the new normally trusted Codex exchange. Root telemetry is deliberately still the prior 93-owner snapshot; the private source-backed inventory identifies 141 expected same-project owners for final ingestion. Newly observed smoke sessions belong to different physical directories and must not be relabelled into this ledger. Final root ingestion remains deferred until acceptance work finishes; the PR remains draft.

## Follow-up backlog

AL-76 defines compact ledger JSON serialization and a source-independent format command with all-workspace, single-ticket and ticket-set selection, dry-run/check modes, value preservation and safe concurrent publication. Its specification is Ready under plan/to-do. Implementation requires its own dedicated session; no format command is implemented in AL-68.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 249
- **Active Agent Duration:** 17h 31m (excludes user wait time)
- **Context Window Health:** CRITICAL — Root agent peak 454.1k tokens (227% of 200k healthy ceiling, 3 compactions) | Subagents peak 348.1k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.5 avg turns/run (peak: 6 turns) | Planner 1.3, Coder 3.0, Tester 6.0, Reviewer 1.2, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.1:1 (+5,291 test lines / +4,684 prod lines)
- **Code Delivered:** +10,026 / -49 lines across 56 code files (excluding docs)
- **Reasoning Effort:** 624,696 reasoning tokens (27.6% of generated output)
- **Tool Health & Self-Correction:** 3,085 tool calls (150 self-corrections, 95.1% success rate)

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
| **Unknown** | 1 | 2,688,102 | 394,577 | 119,759,744 | 97.8% | 454.1k | CRITICAL (227%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 4 | 524,799 | 31,831 | 4,718,080 | 90.0% | 106.1k | ELEVATED | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 4 | 476,895 | 27,453 | 4,110,976 | 89.6% | 98.6k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-data** | 4 | 383,552 | 32,720 | 4,905,856 | 92.7% | 99.3k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-tests** | 2 | 236,275 | 13,714 | 2,179,328 | 90.2% | 103.7k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **rootagent** | 66 | 7,336,451 | 841,164 | 141,539,968 | 95.1% | 453.9k | CRITICAL (227%) | Unavailable (0 observed, 0 partial, 66 unavailable / 66 invocations) |
| **Total** | **93** | **14,922,664** | **1,641,470** | **376,518,912** | **96.2%** | **454.1k** | **CRITICAL** | **Unavailable (0 observed, 0 partial, 93 unavailable / 93 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 32 | 2,557,710 | 162,470 | 18,918,272 | 88.1% | Unavailable (0 observed, 0 partial, 32 unavailable / 32 invocations) |
| **gpt-6.1-sol** | 61 | 12,364,954 | 1,479,000 | 357,600,640 | 96.7% | Unavailable (0 observed, 0 partial, 61 unavailable / 61 invocations) |
| **Total** | **93** | **14,922,664** | **1,641,470** | **376,518,912** | **96.2%** | **Unavailable (0 observed, 0 partial, 93 unavailable / 93 invocations)** |

Capture reconciliation completed: 2026-11-02T23:16:34.858242478Z (per-owner refreshed_at checkpoints; no common as-of snapshot). Current: 36; pending: 57; failed: 0; unassigned: 0. Detailed report disabled; any existing report is stale. Source failures preserve prior facts; coverage is limited to registered owners. Hook installation alone does not prove complete capture.
