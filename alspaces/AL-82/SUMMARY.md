# AL-82: Filter and compare metrics by ticket tags

Delivered durable ticket-level tags from shared YAML, private local YAML and repeatable generation flags through ledger persistence, APIs, dashboard filters and comparison colors. Per-key precedence is CLI over local over shared; absent inputs preserve saved keys. Tags belong to the whole ticket, independently of session users and harnesses. Filters select complete tickets and totals. Comparison coloring preserves point identity, coordinates, completion details and exclusions, with Missing distinct from literal tag text.

The user approved with “Approved. Be sure the tags are in the ticket scope/level”. The [iteration 02 delta](iterations/02/PLAN.md) supersedes the original session-level ownership, partial-owner aggregation and Mixed contributor proposal. The [contract](CONTRACT.md) records current acceptance. Untagged schema-6 ledgers remain valid; reads, backfill, supported explicit migration and Git sharing preserve recorded tags without relabeling from reader configuration. Effective private tags deliberately become shared ledger data. The shared project configuration now defaults `agentic-lens-dev`, `agentic-lens-po` and `agentic-lens-review` to `0.1.1`, matching their VERSION files, as requested in the delivery follow-up. Generation persists those defaults once on each ticket ledger, including AL-82, with existing local/CLI precedence and session attribution preserved. These values describe project skill versions, not proof that every skill ran in a given session.

The reviewed source began at `4e5c458322a2cddfdb5911b26a4569ea75f06a09`. Final reviewed candidate SHA256 was `a6bf5c30a5390bad2e22cacacc25659f68c1128e7b33dcbadb54231e0e6bc4f5`; implementation/integrated revision is `57b1a45c3f706cd79d010db2c33920df8a41f6b5`. Fetch/merge of origin/main required no changes. Final artifacts and telemetry do not alter this verified application source.

Consolidated review is APPROVED. Architecture, behavior, data integrity and test effectiveness each received gpt-6.1-sol/high and gpt-6-luna/medium review. Initial Sol behavior/data reviews requested changes; Sol test review then found the refreshed-catalog defect. Initial Luna scoped reviews and Sol architecture found no blockers. All affected cells subsequently approved the corrected source; data Sol's candidate-2 approval carries forward because candidate 3 changes only eight frontend catalog/view/scatter paths, leaving publication, Unicode validation, persistence and backend tests byte-identical. No required cell is cancelled or incomplete. Model/effort attribution is dispatch-acknowledged, not independently source-observed.

Three findings were fixed, with discovery retained in [REVIEW_SUMMARY.json](iterations/02/REVIEW_SUMMARY.json): reviewer-behavior gpt-6.1-sol/high found BEH-01 (5/10), inconsistent Go/JavaScript Unicode whitespace rejection; reviewer-data gpt-6.1-sol/high found DATA-01 (6/10), unpublished tags surviving a handled late-summary failure; reviewer-tests gpt-6.1-sol/high found BEH-02 (5/10), refreshed values absent from catalog/legend/colors. Coders implemented corrections; the tester independently demonstrated publication red/green and source preservation; focused re-reviewers verified fixes without receiving discovery credit. No unresolved findings or accepted debt remain.

Verification on the integrated revision passed: full `make` (Go race suites, static and workflow/distribution checks, 230 frontend unit and 768 frontend integration tests, multi-platform binaries); `make verify-build PLATFORMS=linux/amd64`; and `python3 /tmp/al82-standalone-check.py`. The standalone Linux engine served matching index/JS and build identity from a temporary cwd with no Node on PATH; all three current frontend/embed files matched. [Targeted test evidence](iterations/02/TEST_RESULTS.md) preserves actual worker, cross-developer/idempotence, whole-ticket totals, migration/backfill and publication red/green coverage. Frontend catalog regressions also failed before correction and passed afterward (230 units/44 affected integrations). Raw logs and transcripts stay outside Git.

Physical power-loss/real sync-error injection and real assistive-technology execution were not performed; incomplete rollback or post-commit ambiguity is reported honestly. npm reported 12 existing dependency advisories; dependency upgrades are outside this ticket. The installed manager rejected an existing configuration glob, so telemetry uses the verified built engine directly; its help confirms source/session/nonce/type/output options. This worktree has no configured Codex source path, so the observed source store is supplied explicitly. Final extraction verified schema 6, the exact Codex root `01a1cbc8-082d-74d3-8617-2f0262209ca9`, all 21 expected direct child sessions, unique owner IDs and Codex source labels. No prior ledger or other-harness sessions existed in this ticket workspace. The root turn remains open, so its recorded duration is explicitly provisional; the checkpoint does not claim a verified invoking caller or final turn completion.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 6
- **Active Agent Duration:** 1h 48m (recorded duration including 207s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 239.2k tokens (53% of 450k healthy ceiling, 0 compactions) | Subagents peak 185.2k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.4 avg turns/run (peak: 4 turns) | Planner 3.0, Coder 3.0, Tester 4.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.3:1 (+704 test lines / +563 prod lines)
- **Code Delivered:** +1,267 / -54 lines across 43 code files (excluding docs)
- **Reasoning Effort:** 64,871 reasoning tokens (24.5% of generated output)
- **Tool Health & Self-Correction:** 628 tool calls (36 self-corrections, 94.3% success rate)

### Anomalies & Health Flags
- **Status:** 7 execution anomalies detected (4 CRITICAL, 3 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (17 cycles) — Subagent roles 'tester' and 'coder' alternated for 17 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (7 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (11 cycles) — Subagent roles 'coder' and 'tester' alternated for 11 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (9 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 9 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: planner <-> coder (3 cycles) — Subagent roles 'planner' and 'coder' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-data' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 76,588 | 7,391 | 1,220,992 | 94.1% | 81.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 228,237 | 40,537 | 15,000,960 | 98.5% | 163.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 124,601 | 18,076 | 8,806,656 | 98.6% | 185.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 206,401 | 75,515 | 35,953,536 | 99.4% | 239.2k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 4 | 219,163 | 11,908 | 1,886,464 | 89.6% | 97.1k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 5 | 279,464 | 17,288 | 2,359,808 | 89.4% | 78.5k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-data** | 4 | 209,051 | 13,430 | 2,893,824 | 93.3% | 111.3k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-tests** | 4 | 228,468 | 15,272 | 2,645,760 | 92.1% | 105.9k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Total** | **22** | **1,571,973** | **199,417** | **70,768,000** | **97.8%** | **239.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 22 unavailable / 22 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 9 | 391,927 | 21,039 | 3,117,824 | 88.8% | Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations) |
| **gpt-6.1-sol** | 13 | 1,180,046 | 178,378 | 67,650,176 | 98.3% | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **Total** | **22** | **1,571,973** | **199,417** | **70,768,000** | **97.8%** | **Unavailable (0 observed, 0 partial, 22 unavailable / 22 invocations)** |

Capture observation finalized: 2026-11-11T19:53:25.347354543Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 22; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 22; reserved bytes: 21964151. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1cbea-bf98-70d2-9122-d0894960d903: current; last successful source read 2026-11-11T19:53:23.412117403Z; accepted source checkpoint 2026-11-11T18:42:57.355Z.
Owner codex/01a1cbe3-88ff-7a12-80c3-58d840d18a5b: current; last successful source read 2026-11-11T19:53:23.523471921Z; accepted source checkpoint 2026-11-11T18:35:53.977Z.
Owner codex/01a1cbe8-9d65-7911-98e1-495ad47eda57: current; last successful source read 2026-11-11T19:53:23.594086422Z; accepted source checkpoint 2026-11-11T18:41:47.792Z.
Owner codex/01a1cbeb-2ab1-7440-b591-5603276c64ed: current; last successful source read 2026-11-11T19:53:23.629315595Z; accepted source checkpoint 2026-11-11T18:43:24.849Z.
Owner codex/01a1cbe1-5565-7822-a66a-811016c2b264: current; last successful source read 2026-11-11T19:53:23.705295377Z; accepted source checkpoint 2026-11-11T18:32:47.457Z.
Owner codex/01a1cbc8-55a6-7ff0-889a-4b602637ea08: current; last successful source read 2026-11-11T19:53:23.808128407Z; accepted source checkpoint 2026-11-11T18:15:55.21Z.
Owner codex/01a1cbe8-b89a-7210-9bde-54d58a00e90b: current; last successful source read 2026-11-11T19:53:23.890935379Z; accepted source checkpoint 2026-11-11T18:41:35.964Z.
Owner codex/01a1cbdb-ec11-7920-b24b-67fd46080cb7: current; last successful source read 2026-11-11T19:53:24.010896628Z; accepted source checkpoint 2026-11-11T18:26:24.943Z.
Owner codex/01a1cbda-471c-7291-9eac-e534877fc162: current; last successful source read 2026-11-11T19:53:24.133549213Z; accepted source checkpoint 2026-11-11T18:26:55.185Z.
Owner codex/01a1cbda-7a9b-7df3-87a6-c3cb76ed53ab: current; last successful source read 2026-11-11T19:53:24.326045897Z; accepted source checkpoint 2026-11-11T18:27:07.884Z.
Owner codex/01a1cbe1-1bf0-7eb2-8303-3f1993e8a528: current; last successful source read 2026-11-11T19:53:24.392241667Z; accepted source checkpoint 2026-11-11T18:33:32.688Z.
Owner codex/01a1cbea-6360-7372-985f-b19ac9f42a7a: current; last successful source read 2026-11-11T19:53:24.420036282Z; accepted source checkpoint 2026-11-11T18:42:45.57Z.
Owner codex/01a1cbe1-3658-7f13-8333-beb92692c69c: current; last successful source read 2026-11-11T19:53:24.499033299Z; accepted source checkpoint 2026-11-11T18:35:26.454Z.
Owner codex/01a1cbdb-1bf3-7b10-84ef-9516fd48187e: current; last successful source read 2026-11-11T19:53:24.54565181Z; accepted source checkpoint 2026-11-11T18:25:39.081Z.
Owner codex/01a1cbd2-1e4e-7e11-a10d-e15d05422815: current; last successful source read 2026-11-11T19:53:24.624738515Z; accepted source checkpoint 2026-11-11T18:38:53.08Z.
Owner codex/01a1cbe2-8b14-74f1-9498-f3d3dd40f895: current; last successful source read 2026-11-11T19:53:24.665001741Z; accepted source checkpoint 2026-11-11T18:34:04.735Z.
Additional owner freshness details remain in local capture diagnostics.
<!-- agentic-lens:metrics:end -->
