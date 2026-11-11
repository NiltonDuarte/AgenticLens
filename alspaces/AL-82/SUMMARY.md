# AL-82: Ticket tags with compact integrated filters

Delivered durable ticket-level tags and one compact Ticket tags control in each existing Fleet, Progression, Comparison and Deep Dive filter bar. Its popup selects a key and multiple values while retaining clauses for other keys. OR-within-key/AND-across-keys, Missing versus literals, clear/reset, whole-ticket totals and comparison identities remain intact. Closed controls no longer expand into one dropdown per tag key.

The user approved ticket-level ownership and later requested the compact integrated control. The [contract](CONTRACT.md), [ticket-ownership delta](iterations/02/PLAN.md) and [UI refinement](iterations/03/PLAN.md) record those accepted boundaries. Shared project defaults record dev, PO and review skill versions as `0.1.1`; generation persists them once on the ticket, with local/CLI precedence and independent session attribution. Version tags describe project skill versions, not proof every skill executed.

The original feature was verified at `57b1a45c`; the UI refinement began from `b10a4daa8ceb1d9ad963544d1d204d8b24809319`. Current reviewed/integrated application revision is `650d370cee83347f9d678f4e62e68292c04f4664`, matching corrected candidate SHA256 `d98e8897278a961d27d0625e4e0dba88eace0dc4e7d4747329a8895563cd19d2`. Fetch/merge of origin/main required no changes. Final documentation/telemetry commits do not change this application source.

Consolidated approval is APPROVED. Iteration 03 behavior, architecture and test effectiveness each ran gpt-6.1-sol/high and gpt-6-luna/medium. Initial Sol behavior/architecture requested a focus correction; initial Luna behavior approved its scoped investigation. All six final cells approved the corrected source. Data-integrity review carries from iteration 02 because backend, configuration, codecs, domain matching, API transport and numerical aggregation are unchanged. No required cell is cancelled or incomplete. Model/effort identity is dispatch-acknowledged, not independently source-observed.

[Iteration 03 retained findings](iterations/03/REVIEW_SUMMARY.json) preserve UI03-01 (4/10): tester/gpt-6.1-sol/high found empty-catalog focus loss; independent behavior and architecture reviewers on that profile found additional stale-value/nonempty-refresh triggers; the orchestrator supplied the heading-click trigger. Owned-descendant focus repair and open-only Escape handling fix the lineage. UI03-02 (4/10) records the orchestrator's positioning-height discovery, corrected before formal review. Coders implemented fixes; tester red/green cases and paired re-review verified them without extra discovery credit. [Iteration 02 findings](iterations/02/REVIEW_SUMMARY.json) retain the earlier Unicode, publication and refreshed-catalog fixes discovered respectively by behavior, data and test reviewers on gpt-6.1-sol/high. No unresolved findings or accepted debt remain.

Verification passed on `650d370c`: canonical `make` in an isolated network namespace, including Go race coverage, static/workflow/distribution checks, 230 frontend units, 781 frontend integrations and multi-platform builds; `make verify-build PLATFORMS=linux/amd64`; and standalone HTTP/asset verification from a temporary cwd with no Node on PATH. All three built frontend/embed files matched, including current Tag key UI and engine identity. The first host make failed only because the existing default-port test could not bind occupied 8085. The successful rerun used `unshare --user --map-current-user --net --keep-caps -- sh -c 'ip link set lo up && exec setpriv --ambient-caps=-all --inh-caps=-all make'`, preserving UID 1000. [UI test evidence](iterations/03/TEST_RESULTS.md) and [original boundary evidence](iterations/02/TEST_RESULTS.md) retain focused commands, failures and corrected results. Raw logs/transcripts stay outside Git.

Real rendered viewport/screen-reader and physical power-loss execution were not performed; layout evidence uses DOM/numeric bounds. The prior installed-manager configuration-glob limitation requires direct built-engine telemetry with the observed Codex store supplied explicitly. Existing dependency advisories remain outside this ticket. Final telemetry verified the exact Codex root, all 33 expected direct children, unique owner IDs and preservation of all 22 prior owner identities/source labels. Shared skill-version tags remain ticket-level. The root turn remains explicitly provisional rather than claiming completion.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 7
- **Active Agent Duration:** 3h 28m (recorded duration including 3158s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 361.2k tokens (80% of 450k healthy ceiling, 0 compactions) | Subagents peak 318.1k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.4 avg turns/run (peak: 4 turns) | Planner 2.0, Coder 2.7, Tester 3.5, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.4:1 (+994 test lines / +687 prod lines)
- **Code Delivered:** +1,681 / -55 lines across 45 code files (excluding docs)
- **Reasoning Effort:** 160,952 reasoning tokens (30.5% of generated output)
- **Tool Health & Self-Correction:** 935 tool calls (53 self-corrections, 94.3% success rate)

### Anomalies & Health Flags
- **Status:** 19 execution anomalies detected (10 CRITICAL, 9 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+249.3k tokens) — Context tokens surged by 249.3k to 249.3k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+253.4k tokens) — Context tokens surged by 253.4k to 253.4k tokens.
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+256.1k tokens) — Context tokens surged by 256.1k to 256.1k tokens.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (17 cycles) — Subagent roles 'tester' and 'coder' alternated for 17 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (7 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (11 cycles) — Subagent roles 'coder' and 'tester' alternated for 11 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (9 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (11 cycles) — Subagent roles 'tester' and 'coder' alternated for 11 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (8 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 8 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (9 cycles) — Subagent roles 'coder' and 'tester' alternated for 9 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (254.6k tokens) — Context tokens reached 254.6k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (253.4k tokens) — Context tokens reached 253.4k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (256.1k tokens) — Context tokens reached 256.1k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.1k tokens) — Context tokens reached 251.1k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: planner <-> coder (3 cycles) — Subagent roles 'planner' and 'coder' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-data' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-architecture (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-architecture' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 336,998 | 9,690 | 2,508,800 | 88.2% | 261.7k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **coder** | 3 | 283,125 | 68,569 | 31,479,040 | 99.1% | 311.6k | ELEVATED | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **tester** | 2 | 187,704 | 44,581 | 27,204,992 | 99.3% | 318.1k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Unknown** | 1 | 284,613 | 154,970 | 79,337,984 | 99.6% | 361.2k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 7 | 338,357 | 21,914 | 3,119,872 | 90.2% | 97.1k | HEALTHY | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **reviewer-behavior** | 9 | 444,900 | 31,069 | 3,834,368 | 89.6% | 78.5k | HEALTHY | Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations) |
| **reviewer-data** | 4 | 209,051 | 13,430 | 2,893,824 | 93.3% | 111.3k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-tests** | 6 | 330,675 | 21,894 | 3,614,592 | 91.6% | 105.9k | HEALTHY | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **Total** | **34** | **2,415,423** | **366,117** | **153,993,472** | **98.5%** | **361.2k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 34 unavailable / 34 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 13 | 540,081 | 31,568 | 4,443,136 | 89.2% | Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations) |
| **gpt-6.1-sol** | 21 | 1,875,342 | 334,549 | 149,550,336 | 98.8% | Unavailable (0 observed, 0 partial, 21 unavailable / 21 invocations) |
| **Total** | **34** | **2,415,423** | **366,117** | **153,993,472** | **98.5%** | **Unavailable (0 observed, 0 partial, 34 unavailable / 34 invocations)** |

Capture observation finalized: 2026-11-11T21:41:19.170187903Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 22; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 22; reserved bytes: 24211301. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1cbea-bf98-70d2-9122-d0894960d903: current; last successful source read 2026-11-11T21:41:18.18278776Z; accepted source checkpoint 2026-11-11T18:42:57.355Z.
Owner codex/01a1cbe3-88ff-7a12-80c3-58d840d18a5b: current; last successful source read 2026-11-11T21:41:18.211418174Z; accepted source checkpoint 2026-11-11T18:35:53.977Z.
Owner codex/01a1cbe8-9d65-7911-98e1-495ad47eda57: current; last successful source read 2026-11-11T21:41:18.247851443Z; accepted source checkpoint 2026-11-11T18:41:47.792Z.
Owner codex/01a1cbeb-2ab1-7440-b591-5603276c64ed: current; last successful source read 2026-11-11T21:41:18.280084002Z; accepted source checkpoint 2026-11-11T18:43:24.849Z.
Owner codex/01a1cbe1-5565-7822-a66a-811016c2b264: current; last successful source read 2026-11-11T21:41:18.313213979Z; accepted source checkpoint 2026-11-11T18:32:47.457Z.
Owner codex/01a1cbc8-55a6-7ff0-889a-4b602637ea08: current; last successful source read 2026-11-11T21:41:18.344817003Z; accepted source checkpoint 2026-11-11T18:15:55.21Z.
Owner codex/01a1cbe8-b89a-7210-9bde-54d58a00e90b: current; last successful source read 2026-11-11T21:41:18.372307314Z; accepted source checkpoint 2026-11-11T18:41:35.964Z.
Owner codex/01a1cbdb-ec11-7920-b24b-67fd46080cb7: current; last successful source read 2026-11-11T21:41:18.393118383Z; accepted source checkpoint 2026-11-11T18:26:24.943Z.
Owner codex/01a1cbda-471c-7291-9eac-e534877fc162: current; last successful source read 2026-11-11T21:41:18.424543928Z; accepted source checkpoint 2026-11-11T18:26:55.185Z.
Owner codex/01a1cbda-7a9b-7df3-87a6-c3cb76ed53ab: current; last successful source read 2026-11-11T21:41:18.475224497Z; accepted source checkpoint 2026-11-11T18:27:07.884Z.
Owner codex/01a1cbe1-1bf0-7eb2-8303-3f1993e8a528: current; last successful source read 2026-11-11T21:41:18.518404113Z; accepted source checkpoint 2026-11-11T18:33:32.688Z.
Owner codex/01a1cbea-6360-7372-985f-b19ac9f42a7a: current; last successful source read 2026-11-11T21:41:18.54471272Z; accepted source checkpoint 2026-11-11T18:42:45.57Z.
Owner codex/01a1cbe1-3658-7f13-8333-beb92692c69c: current; last successful source read 2026-11-11T21:41:18.588171281Z; accepted source checkpoint 2026-11-11T18:35:26.454Z.
Owner codex/01a1cbdb-1bf3-7b10-84ef-9516fd48187e: current; last successful source read 2026-11-11T21:41:18.611335934Z; accepted source checkpoint 2026-11-11T18:25:39.081Z.
Owner codex/01a1cbd2-1e4e-7e11-a10d-e15d05422815: current; last successful source read 2026-11-11T21:41:18.659098116Z; accepted source checkpoint 2026-11-11T18:38:53.08Z.
Owner codex/01a1cbe2-8b14-74f1-9498-f3d3dd40f895: current; last successful source read 2026-11-11T21:41:18.685168685Z; accepted source checkpoint 2026-11-11T18:34:04.735Z.
Additional owner freshness details remain in local capture diagnostics.
<!-- agentic-lens:metrics:end -->
