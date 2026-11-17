# fix(live): report unsupported telemetry sources and failed reads (FIX-18)

## Pull request

The dashboard now checks whether its selected SQLite database can supply live OpenCode v1 queries before it starts polling. Startup, session listing, trace lookup and stream admission explain why a selected source is unsupported or unavailable. A supported, idle source still returns a measured empty tree.

Clean v2 databases continue to hydrate historical events. Competing mixed layouts receive an explicit live diagnosis, including when the old reader would have returned only their v1 sessions. The historical reader retains its existing whole-database schema selection, and optional source failures leave portable ledger facts readable.

The live view displays tree and session read failures, keeps the last usable observation, and clears the corresponding diagnostic after a successful refresh. Verification exercises actual server startup and routes against synthetic SQLite stores, stream failure and recovery, and dashboard REST/SSE ordering.

Integrated `make` passed at `f8953abb` with 231 frontend unit tests, 819 frontend integration tests, Go race tests and builds for four platforms. Standalone HTTP responses matched the embedded assets and the expected v2 diagnosis.


## Verification

The reviewed product implementation is commit `ed0a8460`, based on `b6ab0471`. The first integrated run matched a temporary connection alert before REST supplied the authoritative reason. Commit `f8953abb` made that one test wait for its specific expected text. The full frontend lane then passed 231 unit and 819 integration tests. Both test reviewers approved this assertion. Both integrations already contained the fetched main revision. Integrated `make` passed on `f8953abb` with Go race and integration tests, vet, lint, release, installation, capture and orchestration checks, 231 frontend unit tests, 819 frontend integration tests and manager/engine builds for Linux and macOS on amd64 and arm64. The resulting Linux amd64 engine passed the standalone HTTP and embedded-asset checks. The build version carried a dirty-worktree suffix for local artifact and policy edits. At the baseline, both v2 live REST routes returned HTTP 500 and SSE admitted a silent connection. Against the integrated engine the same fixture produced HTTP 503 with the same v2 diagnosis on all three routes. The standalone server's index, JavaScript and CSS responses matched the canonical build bytes.

The compiled-server regression exercises configured and explicit database selection, active and idle v1, legacy optional timestamps, clean v2 historical hydration, mixed layouts, absent, unknown, malformed, missing and corrupt sources. The Linux denied-file test passed under UID 1000. Its root-runner downgrade branch compiled and was inspected; runtime verification used an unprivileged account. The event-stream recovery case passed ten consecutive runs under the race detector. Source-backed and portable-ledger fixtures verified owner identity, parent links, independent historical access and unchanged ledger bytes after GET. The two focused frontend files passed 93 tests on the final candidate, including the original dashboard failure probes.

## Independent review

The first discovery pass paired `gpt-6.1-sol/high` and `gpt-6-luna/medium` across behavior, data, architecture and tests. The data reviewer on Sol found a failed trace read that erased observed anomalies. The tests reviewer on Luna identified the missing denied-file case; the tests reviewer on Sol demonstrated an unprotected pending trace response. These cases now have direct verification.

During correction review the updated Sol profile (`gpt-6.2-sol/high`) found stale session errors after reconnect and ordering defects between manual, fallback and trace requests. The architecture reviewer discovered the reconnect issue. Behavior and architecture independently reproduced both ordering defects, and the tests reviewer also reproduced the fallback/manual defect. A successful tree now reconciles source diagnostics using the observation order while preserving independent failures and valid pending trace data. The final behavior, architecture and tests rechecks approved on both configured profiles. Both data reviewers approved the corrected server; subsequent changes touched the browser and a Linux test only. The retained diagnostic tests use equal successful payloads in the cross-path cases. A disposable distinct-payload probe passed the implementation and failed a deliberate stale-write mutation. [Attributed findings](evidence/iterations/01/REVIEW_SUMMARY.json)

Consolidated review approved `f8953abb` after the integrated gates passed. All 30 focus/profile runs are accounted for across discovery and correction. The final one-line synchronization assertion changes only test timing. Native Codex turn contexts corroborated the recorded reviewer models and efforts.

The fixtures use controlled sources and browser responses. They establish the specified behavior without a fresh native OpenCode session. v2 and Codex remain historical-only sources for this dashboard.

## Telemetry scope

The integrated engine selected the configured local Codex store. A temporary worktree-local policy supplied the private publication state directory. Capture stayed disabled. The root extraction includes native child sessions. Fifteen host-delegated reviewer runs have separate source roots; their user assignments match the parent's recorded dispatches exactly. They were extracted by exact ID before the parent refresh to preserve their native identities and ancestry. The snapshot was taken before this conversation closed. The ledger contains the parent, 19 native children and 15 independent reviewer roots. The source did not report costs.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 54
- **Active Agent Duration:** 2h 41m (recorded duration including 983s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 208.3k tokens (46% of 450k healthy ceiling, 0 compactions) | Subagents peak 178.0k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.4 avg turns/run (peak: 5 turns) | Planner 1.0, Coder 2.5, Tester 5.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.8:1 (+716 test lines / +407 prod lines)
- **Code Delivered:** +1,123 / -221 lines across 15 code files (excluding docs)
- **Reasoning Effort:** 43,804 reasoning tokens (17.4% of generated output)
- **Tool Health & Self-Correction:** 898 tool calls (66 self-corrections, 92.7% success rate)

### Anomalies & Health Flags
- **Status:** 6 execution anomalies detected (4 CRITICAL, 2 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (26 cycles) — Subagent roles 'coder' and 'tester' alternated for 26 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (7 cycles) — Subagent roles 'coder' and 'tester' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (5 cycles) — Subagent roles 'tester' and 'coder' alternated for 5 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (8 cycles) — Subagent roles 'tester' and 'coder' alternated for 8 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> tester (3 cycles) — Subagent roles 'reviewer-tests' and 'tester' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 62,872 | 3,272 | 521,088 | 89.2% | 61.4k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 2 | 310,502 | 24,537 | 10,921,216 | 97.2% | 123.9k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 206,648 | 25,424 | 14,149,376 | 98.6% | 178.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 16 | 1,150,389 | 87,233 | 43,268,352 | 97.4% | 208.3k | HEALTHY | Unavailable (0 observed, 0 partial, 16 unavailable / 16 invocations) |
| **reviewer-architecture** | 4 | 266,150 | 15,511 | 3,005,184 | 91.9% | 85.0k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 4 | 262,225 | 16,783 | 3,157,376 | 92.3% | 84.0k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-data** | 2 | 114,816 | 9,763 | 1,700,736 | 93.7% | 83.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 5 | 255,246 | 26,040 | 5,192,576 | 95.3% | 89.0k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **35** | **2,628,848** | **208,563** | **81,915,904** | **96.9%** | **208.3k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 35 unavailable / 35 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 15 | 604,373 | 46,559 | 5,144,704 | 89.5% | Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations) |
| **gpt-6.1-sol** | 7 | 446,831 | 33,745 | 6,706,816 | 93.8% | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **gpt-6.2-astra** | 5 | 1,126,038 | 93,907 | 63,715,328 | 98.3% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **gpt-6.2-sol** | 8 | 451,606 | 34,352 | 6,349,056 | 93.4% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **35** | **2,628,848** | **208,563** | **81,915,904** | **96.9%** | **Unavailable (0 observed, 0 partial, 35 unavailable / 35 invocations)** |

Retained recorded history: 15 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
