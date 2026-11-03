# AL-61 Work Summary

## Delivered behavior

Ticket Deep Dive shows viewer-local calendar dates and time to the second for workspace Started/Completed/Updated, trace events, skill Loaded At and selected historical event Timestamp. Full tooltips retain timezone/offset context and valid source instants remain machine-readable. Missing states and invalid raw evidence remain explicit. Scoped styles permit longer labels to wrap. The historical validator rejects impossible calendar dates before conversion.

Domain/adapters: no AL-61 Go, storage, API, duration or ordering changes. Live millisecond formatting, chart axes and raw thread diagnostics remain unchanged. AL-55 telemetry is untouched. Scope and accepted label inventory are frozen in [CONTRACT.md](CONTRACT.md); delivery steering is recorded in [iteration plan](iterations/01/PLAN.md).

## Revision and integration

User pre-approved the ticket on 2026-11-03. Initial Git status was clean at 122e983f8c365a2aa248b1b6d492252555e922ae. Corrected historical implementation/test revision: 2e4762d. Draft PR #104 initially published the reviewed checkpoint; documentation-only changes followed.

After MGT-2's actual AL-60 merge notification, integrated current main exactly once: parent c40e0022aca6c5ae40633b260ce93c379e2ee1c7 and upstream 6244585dc5cc41ef012cd3e1fb3b964ccd910a95 produced tested/reviewed merge 3c0e4cd59979309a14ad785882d609b62d40ad03 without conflicts. AL-60 comparison search, capture repairs, CI and completed ticket/index artifacts remain byte-identical to upstream. The historical formatter, four consumer owners and tests remain byte-identical to the corrected revision; shared CSS contains both independent sets of rules. Final cleanup changes ticket documentation/telemetry only. MGT-2 owns merging; no merge or auto-merge is requested by this thread.

## Review and disposition

Consolidated verdict: APPROVED after integrated make and standalone asset verification passed at 3c0e4cd59979309a14ad785882d609b62d40ad03. Narrow independent generic reviews covered historical behavior, data/failure semantics, architecture and test effectiveness under both configured profiles; rationale is in the plan. [Compact findings](iterations/01/REVIEW_SUMMARY.json) retain all candidates and discovery credit.

Initial reviewer gpt-6.1-sol/high requested changes after discovering SOL-1 (severity 5): JavaScript normalized impossible calendar dates into fabricated dates. Initial reviewer gpt-6-luna/medium reported no candidate. Coder gpt-6.1-sol/high added shared historical calendar validation; tester gpt-6.1-sol/high extended existing regression cases. Both reviewer profiles verified the correction and approved it. Both profiles independently reviewed the predecessor integration and approved it without new findings. Luna's integration verdict initially remained INCOMPLETE pending parent make/asset evidence; after inspecting recorded passing evidence it changed to APPROVED. Verification credit is distinct from discovery: Luna verified SOL-1 but did not discover it.

Model/effort identity was confirmed from child rollout session metadata and turn_context: gpt-6.1-sol/high for planner/coder/tester/Sol reviews, gpt-6-luna/medium for Luna reviews. All nine native children identify the source root 01a1a27f-326c-79b1-bb26-5c71de3375b1. This source check supersedes earlier dispatch-only identity limits. Raw transcripts remain outside Git. Detailed review working reports were compacted only after final approval.

## Verification

Tester completed 244 focused cases before integration: 73 formatter, 20 drawer, 84 trace and 67 TicketDeepDive, with TypeScript and ESLint. Controlled America/New_York tests cover calendar rollover, spring DST offsets, impossible dates, leap days, offset equivalents, missing/invalid values, source datetime/tooltips and preserved live precision. The timezone environment is restored. Independent Sol probes additionally covered autumn DST and adapter timestamp precision/offset formats.

Coordinator ran required make at merged revision 3c0e4cd: PASS. This includes Go race tests for every package (including CLI/capture), vet, TypeScript, ESLint, release/installation/capture-distribution Python checks (14/10/15 tests), all 790 frontend tests across 36 files and builds for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64 and windows/amd64. Prior full make at 2e4762d failed three unchanged capture tests; [original failed gate evidence](iterations/01/MAKE_FAILED.log) is preserved. AL-60 alone supplied the authorized repairs, and the integrated gate passed without AL-61 test weakening or duplicate capture fixes.

Canonical build synchronized web/dist and backend/internal/adapters/web/dist byte-for-byte. A disposable empty-workspace HTTP probe of the standalone Linux binary returned matching assets: index-B2kf5x4y.js SHA256 08fd43e36bf73da220961685eda43d03f6d7403d390988f92532829c34408bc8; index-BrFBNE4n.css SHA256 74c8e8925ce877e0f0fd4bc8b3842b55683332641d82dc2efe788ed70a865614. Build/index paths and both feature sources are preserved. Generated assets remain ignored under the repository tracking policy.

Limits: browser viewport geometry, assistive technology and remote CI were not certified by local verification. DOM semantics and scoped CSS inspection support layout/accessibility behavior but do not prove rendered geometry. Upstream AL-60's separately recorded nonblocking test-oracle debt remains owned by AL-60; this ticket neither modifies nor hides it.

## AL-60 follow-up integration and current readiness

User notified verified merge of AL-60 PR #105 at 3fbecd2576270772462fe37235d90c5f33e2164c. Merged current main into 64d1e3d once without conflicts, producing tested/reviewed merge 8c5e3568a53ce357d4d6870ddaec9289ed089985. Four upstream Go test files add real worker-completion teardown barriers and delayed-start/early-return regression assertions; other incoming changes are AL-60 artifacts. Production, CI, all frontend source/tests and historical timestamp inputs are byte-identical to the previously approved head. Backend/CI exactly match upstream. AL-60's completed ticket-index state, search behavior and historical timestamp behavior remain preserved.

The coordinator inspected the four-file merge and its cleanup ordering, completion probes, retained isolation/publication assertions and upstream review/verification. Existing Sol gpt-6.1-sol/high and Luna gpt-6-luna/medium feature/integration approvals carry forward because none of their production or frontend-test dependencies changed. The upstream test-only repair remains AL-60-owned; its review evidence is retained in AL-60 iteration 03. No feature-workflow restart, duplicate repair or new finding. Consolidated approval remains APPROVED after the refreshed required checks.

Required full make at 8c5e3568: PASS, including CLI Go race tests (96.194s), every other Go package, vet, TypeScript/ESLint, Python workflow/install/distribution checks (14/10/15), all 790 frontend tests in 36 files and five platform builds. Rebuilt asset trees match, and the updated standalone Linux binary served byte-identical JS/CSS with the same hashes recorded above. [Current verification](iterations/01/VERIFICATION.md) records the tested revision and carry-forward limits. Final publication changes only AL-61 documentation/telemetry; hosted checks must refresh on the pushed final head. MGT-2 owns merging.

## Telemetry

Final Codex extraction follows all completed review/integration work. Expected source root: 01a1a27f-326c-79b1-bb26-5c71de3375b1; nine source-linked child sessions. The initial ledger contained 10 Codex sessions; subsequent refreshes preserve those session identities. No other-harness sessions are present. Extraction/session completeness is checked against actual source metadata; no costs or session facts are inferred.

Extraction verified: all 10 ledger sessions exactly match actual rollout IDs (one root and all nine children), every source is Codex, and each child points to the expected root. No existing AL-61 sessions from another harness were present. Cost evidence is unavailable for all invocations; no numeric cost is asserted. The source root has no recorded role, so generated metrics correctly retain Unknown rather than invent orchestrator attribution. Source extraction succeeded with the current built native CLI from the project directory; wrapper-only --al-offline/--al-project options are not native flags, and the native command required an explicit ticket title. The successful final command used --source codex, explicit root, literal nonce, title and feat type. Earlier argument failures produced no fabricated ledger. The latest generated metrics below report source-derived activity, cache and context facts; idle supervision time is excluded from active durations.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 15
- **Active Agent Duration:** 50m 0s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 171.5k tokens (86% of 200k healthy ceiling, 0 compactions) | Subagents peak 85.6k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.3 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 2.0, Tester 2.0, Reviewer 1.2
- **Test-to-Production Code Ratio:** 1.6:1 (+99 test lines / +61 prod lines)
- **Code Delivered:** +160 / -17 lines across 10 code files (excluding docs)
- **Reasoning Effort:** 14,402 reasoning tokens (16.2% of generated output)
- **Tool Health & Self-Correction:** 280 tool calls (10 self-corrections, 96.4% success rate)

### Anomalies & Health Flags
- **Status:** 2 execution anomalies detected (2 WARNING).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (167.9k tokens) — Context tokens reached 167.9k, breaching the 160k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (4 cycles) — Subagent roles 'coder' and 'tester' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 59,593 | 2,675 | 360,192 | 85.8% | 61.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 107,633 | 4,773 | 1,608,704 | 93.7% | 82.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 133,651 | 5,973 | 2,389,760 | 94.7% | 85.6k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer** | 6 | 322,214 | 25,585 | 3,238,272 | 91.0% | 73.0k | HEALTHY | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **Unknown** | 1 | 324,199 | 35,625 | 17,511,040 | 98.2% | 171.5k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **10** | **947,290** | **74,631** | **25,107,968** | **96.4%** | **171.5k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 3 | 149,766 | 11,049 | 1,546,752 | 91.2% | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **gpt-6.1-sol** | 7 | 797,524 | 63,582 | 23,561,216 | 96.7% | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **Total** | **10** | **947,290** | **74,631** | **25,107,968** | **96.4%** | **Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations)** |

Capture observation finalized: 2026-11-03T20:54:51.702505655Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 10; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 10; reserved bytes: 8286616. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1a281-613f-7fd1-8918-2d7847fc3f1a: current; last successful source read 2026-11-03T20:54:51.51457495Z; accepted source checkpoint 2026-11-03T17:48:52.427Z.
Owner codex/01a1a283-9c15-74c1-be26-747e97958030: current; last successful source read 2026-11-03T20:54:51.53198463Z; accepted source checkpoint 2026-11-03T17:47:53.186Z.
Owner codex/01a1a286-ea21-7a93-bd0d-68d3b710ecd1: current; last successful source read 2026-11-03T20:54:51.548655962Z; accepted source checkpoint 2026-11-03T17:50:33.779Z.
Owner codex/01a1a283-af6c-7162-a103-9548bdbdfac0: current; last successful source read 2026-11-03T20:54:51.563469329Z; accepted source checkpoint 2026-11-03T17:45:54.888Z.
Owner codex/01a1a27f-5e46-7410-b31a-be93ba0401ac: current; last successful source read 2026-11-03T20:54:51.578532868Z; accepted source checkpoint 2026-11-03T17:41:36.094Z.
Owner codex/01a1a2f6-5912-7890-ad62-caf9bf7887d6: current; last successful source read 2026-11-03T20:54:51.595495433Z; accepted source checkpoint 2026-11-03T19:54:58.997Z.
Owner codex/01a1a280-bd1e-7340-929a-59edc79a5430: current; last successful source read 2026-11-03T20:54:51.61072646Z; accepted source checkpoint 2026-11-03T17:48:22.431Z.
Owner codex/01a1a27f-326c-79b1-bb26-5c71de3375b1: current; last successful source read 2026-11-03T20:54:51.661503451Z; accepted source checkpoint 2026-11-03T20:54:51.485Z.
Owner codex/01a1a287-0eb8-7692-925d-df396eaa7fa3: current; last successful source read 2026-11-03T20:54:51.676352276Z; accepted source checkpoint 2026-11-03T17:51:22.099Z.
Owner codex/01a1a2f6-46fd-7db0-8868-88e1e6163abd: current; last successful source read 2026-11-03T20:54:51.695098973Z; accepted source checkpoint 2026-11-03T19:51:49.885Z.
