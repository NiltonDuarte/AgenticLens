# AL-82: Ticket tags with removable active filters

Delivered durable ticket-level tags and one compact Ticket tags control in each existing Fleet, Progression, Comparison and Deep Dive filter bar. Its popup selects a key and multiple values while retaining clauses for other keys. OR-within-key/AND-across-keys, Missing versus literals, clear/reset, whole-ticket totals and comparison identities remain intact. The popup lists active key clauses with their values and an accessible X for each. One quiet Clear all action replaces the old clearing buttons. Closed controls remain compact.

The user approved ticket-level ownership and later requested the compact integrated control. The [contract](CONTRACT.md), [ticket-ownership delta](iterations/02/PLAN.md), [compact UI refinement](iterations/03/PLAN.md) and [active-list refinement](iterations/04/PLAN.md) record those accepted boundaries. Shared project defaults record dev, PO and review skill versions as `0.1.1`; generation persists them once on the ticket, with local/CLI precedence and independent session attribution. Version tags describe project skill versions, not proof every skill executed.

The original feature was verified at `57b1a45c`; the UI refinement began from `b10a4daa8ceb1d9ad963544d1d204d8b24809319`. The compact UI was verified at `650d370c`. Current reviewed/integrated application revision is `1655ba4541a451ec266b4286be357f775b65d203`, matching candidate SHA256 `8d28749f56a8007014e4f329e470fdaea8db0567c540824a5914d22ade5e5676`. Fetch/merge of origin/main required no changes. Final documentation/telemetry commits do not change this application source.

Consolidated approval is APPROVED. Iteration 04 behavior and test-effectiveness reviews each ran gpt-6.1-sol/high and gpt-6-luna/medium; all four approved with [no new findings](iterations/04/REVIEW_SUMMARY.json). Architecture carries from iteration 03 because this refinement adds no hooks, state, types, dependencies or ownership changes. Data review carries from iteration 02 because backend, configuration, codecs, matching, transport and aggregation are unchanged. No required cell is incomplete or cancelled. Model/effort identity is dispatch-acknowledged, not independently source-observed. Iteration 03's initial Sol behavior/architecture requested correction, initial Luna behavior approved its scope, and all six final cells approved corrected source.

[Iteration 03 retained findings](iterations/03/REVIEW_SUMMARY.json) preserve UI03-01 (4/10): tester/gpt-6.1-sol/high found empty-catalog focus loss; independent behavior and architecture reviewers on that profile found additional stale-value/nonempty-refresh triggers; the orchestrator supplied the heading-click trigger. Owned-descendant focus repair and open-only Escape handling fix the lineage. UI03-02 (4/10) records the orchestrator's positioning-height discovery, corrected before formal review. Coders implemented fixes; tester red/green cases and paired re-review verified them without extra discovery credit. [Iteration 02 findings](iterations/02/REVIEW_SUMMARY.json) retain the earlier Unicode, publication and refreshed-catalog fixes discovered respectively by behavior, data and test reviewers on gpt-6.1-sol/high. No unresolved findings or accepted debt remain.

Verification passed on `1655ba45`: canonical `make` in an isolated network namespace, including Go race coverage, static/workflow/distribution checks, 230 frontend units, 781 frontend integrations and multi-platform builds; `make verify-build PLATFORMS=linux/amd64`; and standalone HTTP/asset verification from a temporary cwd with no Node on PATH. All three built frontend/embed files matched, including the current Active filters/Clear all UI and engine identity. Iteration 03's host make failed only because an existing default-port test could not bind occupied 8085; subsequent full gates use isolation. The successful rerun used `unshare --user --map-current-user --net --keep-caps -- sh -c 'ip link set lo up && exec setpriv --ambient-caps=-all --inh-caps=-all make'`, preserving UID 1000. [Active-list test evidence](iterations/04/TEST_RESULTS.md), [compact-editor evidence](iterations/03/TEST_RESULTS.md) and [original boundary evidence](iterations/02/TEST_RESULTS.md) retain focused commands, failures and corrected results. Raw logs/transcripts stay outside Git.

Real rendered viewport/screen-reader and physical power-loss execution were not performed; layout evidence uses DOM/numeric bounds. The prior installed-manager configuration-glob limitation requires direct built-engine telemetry with the observed Codex store supplied explicitly. Existing dependency advisories remain outside this ticket. Final telemetry verified the exact Codex root, all 40 expected direct children, unique owner IDs and preservation of all 34 prior owner identities/source labels. Shared skill-version tags remain ticket-level. The root turn remains explicitly provisional rather than claiming completion.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 9
- **Active Agent Duration:** 4h 5m (recorded duration including 1356s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 411.2k tokens (91% of 450k healthy ceiling, 0 compactions) | Subagents peak 392.5k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.3 avg turns/run (peak: 4 turns) | Planner 1.7, Coder 2.3, Tester 3.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.4:1 (+1,020 test lines / +713 prod lines)
- **Code Delivered:** +1,733 / -55 lines across 45 code files (excluding docs)
- **Reasoning Effort:** 197,420 reasoning tokens (31.6% of generated output)
- **Tool Health & Self-Correction:** 1,046 tool calls (59 self-corrections, 94.4% success rate)

### Anomalies & Health Flags
- **Status:** 27 execution anomalies detected (14 CRITICAL, 13 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+249.3k tokens) — Context tokens surged by 249.3k to 249.3k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+253.4k tokens) — Context tokens surged by 253.4k to 253.4k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+256.1k tokens) — Context tokens surged by 256.1k to 256.1k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+380.1k tokens) — Context tokens surged by 380.1k to 380.1k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+372.4k tokens) — Context tokens surged by 372.4k to 372.4k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+372.7k tokens) — Context tokens surged by 372.7k to 372.7k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (17 cycles) — Subagent roles 'tester' and 'coder' alternated for 17 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (7 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (11 cycles) — Subagent roles 'coder' and 'tester' alternated for 11 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (9 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (11 cycles) — Subagent roles 'tester' and 'coder' alternated for 11 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (8 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (9 cycles) — Subagent roles 'coder' and 'tester' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (5 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 5 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.6k tokens) — Context tokens reached 254.6k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.4k tokens) — Context tokens reached 253.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (256.1k tokens) — Context tokens reached 256.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (380.1k tokens) — Context tokens reached 380.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (372.4k tokens) — Context tokens reached 372.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (372.7k tokens) — Context tokens reached 372.7k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.1k tokens) — Context tokens reached 251.1k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: planner <-> coder (3 cycles) — Subagent roles 'planner' and 'coder' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-data' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 3 | 717,902 | 12,972 | 4,402,688 | 86.0% | 381.7k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **coder** | 4 | 302,006 | 76,841 | 37,201,792 | 99.2% | 389.6k | ELEVATED | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **tester** | 3 | 211,693 | 48,916 | 32,563,456 | 99.4% | 392.5k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **Unknown** | 1 | 679,548 | 190,756 | 98,648,576 | 99.3% | 411.2k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 7 | 338,357 | 21,914 | 3,119,872 | 90.2% | 97.1k | HEALTHY | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **reviewer-behavior** | 11 | 543,684 | 36,461 | 4,361,984 | 88.9% | 78.5k | HEALTHY | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **reviewer-data** | 4 | 209,051 | 13,430 | 2,893,824 | 93.3% | 111.3k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-tests** | 8 | 421,182 | 26,185 | 4,182,016 | 90.9% | 105.9k | HEALTHY | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **41** | **3,423,423** | **427,475** | **187,374,208** | **98.2%** | **411.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 41 unavailable / 41 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 15 | 615,909 | 36,345 | 5,009,792 | 89.1% | Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations) |
| **gpt-6.1-sol** | 26 | 2,807,514 | 391,130 | 182,364,416 | 98.5% | Unavailable (0 observed, 0 partial, 26 unavailable / 26 invocations) |
| **Total** | **41** | **3,423,423** | **427,475** | **187,374,208** | **98.2%** | **Unavailable (0 observed, 0 partial, 41 unavailable / 41 invocations)** |

Capture observation finalized: 2026-11-12T20:14:53.949616017Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 34; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 34; reserved bytes: 39909554. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1cbea-bf98-70d2-9122-d0894960d903: current; last successful source read 2026-11-12T20:14:52.826113493Z; accepted source checkpoint 2026-11-11T18:42:57.355Z.
Owner codex/01a1cc79-58bb-7c62-b4ce-749c48d8e418: current; last successful source read 2026-11-12T20:14:52.84901879Z; accepted source checkpoint 2026-11-11T21:18:51.733Z.
Owner codex/01a1cc6d-61f6-7e11-baf7-dbc468fcbef7: current; last successful source read 2026-11-12T20:14:52.871168884Z; accepted source checkpoint 2026-11-11T21:05:38.653Z.
Owner codex/01a1cc60-0aa0-7103-9637-39898131ee79: current; last successful source read 2026-11-12T20:14:52.927699278Z; accepted source checkpoint 2026-11-11T21:12:32.126Z.
Owner codex/01a1cbe3-88ff-7a12-80c3-58d840d18a5b: current; last successful source read 2026-11-12T20:14:52.947176062Z; accepted source checkpoint 2026-11-11T18:35:53.977Z.
Owner codex/01a1cc77-119d-7df2-9454-f8d917fec3c0: current; last successful source read 2026-11-12T20:14:52.968125114Z; accepted source checkpoint 2026-11-11T21:16:53.786Z.
Owner codex/01a1cbe8-9d65-7911-98e1-495ad47eda57: current; last successful source read 2026-11-12T20:14:52.989027176Z; accepted source checkpoint 2026-11-11T18:41:47.792Z.
Owner codex/01a1cc6d-7f83-7553-9148-5dd01b6c9537: current; last successful source read 2026-11-12T20:14:53.010827019Z; accepted source checkpoint 2026-11-11T21:07:11.791Z.
Owner codex/01a1cc78-d534-7bd3-9448-925b88a1da03: current; last successful source read 2026-11-12T20:14:53.027073009Z; accepted source checkpoint 2026-11-11T21:18:34.584Z.
Owner codex/01a1cc5f-6c96-7131-b4b8-419ca14956c3: current; last successful source read 2026-11-12T20:14:53.055161941Z; accepted source checkpoint 2026-11-11T20:50:52.206Z.
Owner codex/01a1cbeb-2ab1-7440-b591-5603276c64ed: current; last successful source read 2026-11-12T20:14:53.073396953Z; accepted source checkpoint 2026-11-11T18:43:24.849Z.
Owner codex/01a1cbe1-5565-7822-a66a-811016c2b264: current; last successful source read 2026-11-12T20:14:53.091521761Z; accepted source checkpoint 2026-11-11T18:32:47.457Z.
Owner codex/01a1cbc8-55a6-7ff0-889a-4b602637ea08: current; last successful source read 2026-11-12T20:14:53.110863614Z; accepted source checkpoint 2026-11-11T18:15:55.21Z.
Owner codex/01a1cbe8-b89a-7210-9bde-54d58a00e90b: current; last successful source read 2026-11-12T20:14:53.129103686Z; accepted source checkpoint 2026-11-11T18:41:35.964Z.
Owner codex/01a1cbdb-ec11-7920-b24b-67fd46080cb7: current; last successful source read 2026-11-12T20:14:53.143775128Z; accepted source checkpoint 2026-11-11T18:26:24.943Z.
Owner codex/01a1cbda-471c-7291-9eac-e534877fc162: current; last successful source read 2026-11-12T20:14:53.166585401Z; accepted source checkpoint 2026-11-11T18:26:55.185Z.
Additional owner freshness details remain in local capture diagnostics.
<!-- agentic-lens:metrics:end -->
