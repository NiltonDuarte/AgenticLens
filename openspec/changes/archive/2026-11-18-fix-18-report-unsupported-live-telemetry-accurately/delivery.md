# fix(live): report unsupported telemetry sources and failed reads (FIX-18)

## Pull request

The dashboard now checks whether its selected SQLite database can supply live OpenCode v1 queries before it starts polling. Startup, session listing, trace lookup and stream admission explain why a selected source is unsupported or unavailable. A supported, idle source still returns a measured empty tree.

Clean v2 databases continue to hydrate historical events. Competing mixed layouts receive an explicit live diagnosis, including when the old reader would have returned only their v1 sessions. The historical reader retains its existing whole-database schema selection, and optional source failures leave portable ledger facts readable.

The live view displays tree and session read failures, keeps the last usable observation, and clears the corresponding diagnostic after a successful refresh. Verification exercises actual server startup and routes against synthetic SQLite stores, stream failure and recovery, and dashboard REST/SSE ordering.

The branch includes queue base `fc489ff9`. Integrated `make` passed at `29a7de6a` with 231 frontend unit tests, 819 frontend integration tests, Go race tests and builds for four platforms. The standalone Linux engine served the canonical embedded assets and returned the expected v2 diagnosis on all three live routes.


## Queue integration verification

Merge `29a7de6a309fff63791b80f59c506821fd0a866f` integrated `fc489ff921254d6421110dcda24b585ae3f7a940` into `f22bca4a`. The roadmap index matches the queue base with the completed FIX-18 entry removed. FIX-18, FIX-19, AL-73, AL-80 and AL-65 are in Done. AL-94 is Ready and AL-95 depends on it. AL-65 changed roadmap and OpenSpec documents only.

The two shared Go files merged without a manual source edit. The server keeps the selected source available for historical requests after a failed live start. Historical Codex hydration rejects duplicate locations for a recorded owner, reports the ambiguity, and continues with other owners. Registered generation reports failed or pending owners while retaining accepted facts. The dashboard keeps its last usable observations through read failures and reconciles recovery in request order.

The tester ran `go test -race -count=1` with focused selectors in the engine CLI and the Codex, live, backfill and HTTP adapters. All 51 leaf cases passed with zero skips. The focused Vitest live API and dashboard suites passed all 93 tests. Full `make` passed on the merge, including release and capture checks and four-platform builds. A separate check launched the resulting Linux amd64 engine, compared its served index, JavaScript and CSS byte-for-byte against the canonical build, and verified identical HTTP 503 v2 diagnoses for session listing, trace and SSE admission.

A separate Go overlay combined the SQLite v2 history fixture with identical, divergent compressed, and warning-bearing Codex copies. All three variants passed under the race detector. Each verified the v2 live startup diagnosis, usable v2 and healthy Codex history, explicit exclusion of the ambiguous child, repeatable events, and unchanged ledger facts.

Behavior and data reviewers on `gpt-6.2-sol/high` and `gpt-6-luna/medium` approved the combined paths without new findings. They traced both merge parents through the actual service and reused the unchanged dashboard, architecture and test evidence. Runtime checks used isolated source fixtures. The previous release evidence follows.

## Verification

The reviewed product implementation is commit `ed0a8460`, based on `b6ab0471`. The first integrated run matched a temporary connection alert before REST supplied the authoritative reason. Commit `f8953abb` made that one test wait for its specific expected text. The full frontend lane then passed 231 unit and 819 integration tests. Both test reviewers approved this assertion. Both integrations already contained the fetched main revision. Integrated `make` passed on `f8953abb` with Go race and integration tests, vet, lint, release, installation, capture and orchestration checks, 231 frontend unit tests, 819 frontend integration tests and manager/engine builds for Linux and macOS on amd64 and arm64. The resulting Linux amd64 engine passed the standalone HTTP and embedded-asset checks. The build version carried a dirty-worktree suffix for local artifact and policy edits. At the baseline, both v2 live REST routes returned HTTP 500 and SSE admitted a silent connection. Against the integrated engine the same fixture produced HTTP 503 with the same v2 diagnosis on all three routes. The standalone server's index, JavaScript and CSS responses matched the canonical build bytes.

The compiled-server regression exercises configured and explicit database selection, active and idle v1, legacy optional timestamps, clean v2 historical hydration, mixed layouts, absent, unknown, malformed, missing and corrupt sources. The Linux denied-file test passed under UID 1000. Its root-runner downgrade branch compiled and was inspected; runtime verification used an unprivileged account. The event-stream recovery case passed ten consecutive runs under the race detector. Source-backed and portable-ledger fixtures verified owner identity, parent links, independent historical access and unchanged ledger bytes after GET. The two focused frontend files passed 93 tests on the final candidate, including the original dashboard failure probes.

## Independent review

The first discovery pass paired `gpt-6.1-sol/high` and `gpt-6-luna/medium` across behavior, data, architecture and tests. The data reviewer on Sol found a failed trace read that erased observed anomalies. The tests reviewer on Luna identified the missing denied-file case; the tests reviewer on Sol demonstrated an unprotected pending trace response. These cases now have direct verification.

During correction review the updated Sol profile (`gpt-6.2-sol/high`) found stale session errors after reconnect and ordering defects between manual, fallback and trace requests. The architecture reviewer discovered the reconnect issue. Behavior and architecture independently reproduced both ordering defects, and the tests reviewer also reproduced the fallback/manual defect. A successful tree now reconciles source diagnostics using the observation order while preserving independent failures and valid pending trace data. The final behavior, architecture and tests rechecks approved on both configured profiles. Both data reviewers approved the corrected server; subsequent changes touched the browser and a Linux test only. The retained diagnostic tests use equal successful payloads in the cross-path cases. A disposable distinct-payload probe passed the implementation and failed a deliberate stale-write mutation. [Attributed findings](evidence/iterations/01/REVIEW_SUMMARY.json)

Consolidated review approved `f8953abb` after the integrated gates passed. All 30 focus/profile runs are accounted for across discovery and correction. The final one-line synchronization assertion changes only test timing. Native Codex turn contexts corroborated the recorded reviewer models and efforts.

The fixtures use controlled sources and browser responses. They establish the specified behavior without a fresh native OpenCode session. v2 and Codex remain historical-only sources for this dashboard.

## Telemetry scope

The integrated engine selected the configured local Codex store. A temporary worktree-local policy supplied the private publication state directory. Capture stayed disabled. The root extraction includes native child sessions. Seventeen host-delegated reviewer runs have separate source roots; their user assignments match the parent's recorded dispatches exactly. They were extracted by exact ID before the parent refresh to preserve their native identities and ancestry. The snapshot was taken before this conversation closed. The ledger contains the parent, 21 native children and 17 independent reviewer roots. The source did not report costs.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 63
- **Active Agent Duration:** 3h 21m (recorded duration including 948s provisional)
- **Context Window Health:** ELEVATED — Root agent peak 244.7k tokens (54% of 450k healthy ceiling, 1 compactions) | Subagents peak 187.0k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.4 avg turns/run (peak: 6 turns) | Planner 1.0, Coder 3.0, Tester 6.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.8:1 (+716 test lines / +407 prod lines)
- **Code Delivered:** +1,123 / -221 lines across 15 code files (excluding docs)
- **Reasoning Effort:** 46,553 reasoning tokens (16.4% of generated output)
- **Tool Health & Self-Correction:** 1,032 tool calls (69 self-corrections, 93.3% success rate)

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
| **coder** | 2 | 438,757 | 24,957 | 11,300,096 | 96.3% | 127.9k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 1 | 376,609 | 26,964 | 15,632,768 | 97.6% | 187.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 18 | 1,568,966 | 106,296 | 57,731,584 | 97.4% | 244.7k | ELEVATED | Unavailable (0 observed, 0 partial, 18 unavailable / 18 invocations) |
| **reviewer-architecture** | 4 | 266,150 | 15,511 | 3,005,184 | 91.9% | 85.0k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **reviewer-behavior** | 5 | 343,448 | 21,141 | 3,871,232 | 91.9% | 84.0k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-data** | 3 | 181,254 | 12,693 | 2,067,584 | 91.9% | 83.8k | HEALTHY | Unavailable (0 observed, 0 partial, 3 unavailable / 3 invocations) |
| **reviewer-tests** | 5 | 255,246 | 26,040 | 5,192,576 | 95.3% | 89.0k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **39** | **3,493,302** | **236,874** | **99,322,112** | **96.6%** | **244.7k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 39 unavailable / 39 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 17 | 736,804 | 52,295 | 7,113,088 | 90.6% | Unavailable (0 observed, 0 partial, 17 unavailable / 17 invocations) |
| **gpt-6.1-sol** | 7 | 446,831 | 33,745 | 6,706,816 | 93.8% | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **gpt-6.2-astra** | 5 | 1,710,400 | 109,194 | 78,072,448 | 97.9% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **gpt-6.2-sol** | 10 | 599,267 | 41,640 | 7,429,760 | 92.5% | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **Total** | **39** | **3,493,302** | **236,874** | **99,322,112** | **96.6%** | **Unavailable (0 observed, 0 partial, 39 unavailable / 39 invocations)** |

Retained recorded history: 17 session(s), unchanged because their sources were not refreshed. They do not establish capture coverage in this worktree.
<!-- agentic-lens:metrics:end -->
