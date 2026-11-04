# GEN-19: Release candidates reuse Release Please's version

## Current delivered behavior

RCs run after Release Please succeeds on main, or by manual dispatch. The workflow captures the pending same-repository release PR's immutable head, reads its root manifest version, validates it, and appends the existing run/commit suffix. Missing PRs skip; API, ambiguous-selection and invalid-version failures stop the job.

Before publication, one exact stable-tag lookup remains: HTTP 200 skips, 404 permits publication, other statuses or transport failures stop it. The RC tag is explicitly pinned to EXPECTED_SHA, matching the checked-out source and binary metadata. Existing build, artifact verification, prerelease and retention steps remain.

The user's explicit "make it simplier" instruction approved the immediately preceding proposal. It supersedes iteration 02's generated-parent assumption, repeated latest-main/PR checks and historical release scans. Later main or PR updates no longer invalidate a captured candidate. These are intentional accepted guarantees, not unresolved review findings. No standalone calculator or extra package dependency remains. CONTRACT.md records each accepted supersession; [iteration 03 plan](iterations/03/PLAN.md) records this delta.

## Baseline and integration

GEN-19 was allocated after checking the configured alspaces root, ticket index, remote workspace tree and branch names. Initial working tree was clean at 036d808. Branch chore/GEN-19-rc-target-version and type chore persist across all steering. Iteration03 began clean at d29d834; its implementation commit is a9d220e.

Reviewed integration/tested revision: c787a9669abc64014b87696b39af7bee82ad9630. Main advanced during delivery; origin/main was merged without conflicts. RC workflow, release tests, stable workflow, release configuration and manifest were unchanged by integration. CONTRIBUTING merged unrelated installed-entrypoint guidance. Shared source build/asset/capture tooling changed upstream, but release metadata variables and release-build/package/verify consumers stayed compatible. Required integrated make exercised the updated code; no ticket-owned feature changes were added to those upstream modules.

## Verification and review

- Full make passed after integration: Go vet and race suite, frontend type/lint checks, 36 frontend files / 784 tests, 14 release tests, 10 installation tests, 15 capture-distribution tests, frontend embedding and five-platform native builds. Valid cached results for unchanged packages are retained; no conditional stage was silently skipped.
- The 14 release tests preserve six stable revision tests. Actual workflow shell checks immutable manifest version/output, no-PR/API/ambiguity/malformed failures and publication wiring. Actual gh/gojq evaluates raw local JSON; actual curl hits local HTTP fixtures and validates exact tag URL, 200/404/403/500 and transport behavior. Wrong-version, allow-existing-tag and wrong-URL mutations fail. Removed assertions correspond only to explicitly retired guarantees; exact-tag coverage replaces historical-release checks.
- Coder gpt-6.1-sol/high performed live read-only complete metadata and exact-tag probes, producing v2.1.0-rc.114+d29d834 and eligibility=true for missing stable v2.1.0. No release command was invoked. Parent standalone host native version --json identifies the integrated commit, and web/dist matches backend/internal/adapters/web/dist exactly. Development binaries use git describe; RC builds receive validated VERSION explicitly.
- Final consolidated verdict APPROVED after integrated make. All four responsibilities (behavior/integration, data/failure, architecture/maintenance, test effectiveness) ran under gpt-6.1-sol/high and gpt-6-luna/medium. Exact profiles were verified in Codex source records. Both behavior profiles rechecked the publication binding fix; unchanged data/failure evidence carried forward on its unchanged query/guard boundary. Other discovery cells reviewed the corrected source. All cells approved, with no unresolved finding.

[Iteration03 compact finding](iterations/03/REVIEW_SUMMARY.json): REL-02 severity 5 was discovered by behavior/integration reviewer gpt-6.1-sol/high. Existing omission of target_commitish could make GitHub create a tag at newer default-branch HEAD while the binaries retain captured SHA. Parent gpt-6.1-sol/high added the action's built-in EXPECTED_SHA target and one assertion. Behavior reviewers gpt-6.1-sol/high and gpt-6-luna/medium verified the fix; test reviewer gpt-6.1-sol/high confirmed a target_commitish=main mutation fails. Discovery is distinct from verification. The [action input](https://github.com/softprops/action-gh-release/blob/v2/action.yml) and [GitHub release API](https://docs.github.com/en/rest/releases/releases#create-a-release) support the binding. The [reference API](https://docs.github.com/en/rest/git/refs#get-a-reference) provides exact tag lookup rather than prefix matching.

Detailed working reviews and records were compacted after retaining every finding, reporting tuple, disposition, verification actor and reviewed revision. Upstream guidance now prefers private scratch evidence; no detailed reports or persistent content-hash bookkeeping remain in the delivered tree.

## Earlier iterations

Iteration 01 used an independent engine. Iteration 02 removed it after the user's explicit approval and consumed the existing release PR. Their code is superseded by accepted iteration 03 behavior; their evidence does not automatically certify its relaxed guards.

[Iteration 01 findings](iterations/01/REVIEW_SUMMARY.json): REL-01 severity 2 was shared discovery by behavior and architecture reviewers gpt-6.1-sol/high; coder gpt-6.1-sol/high resolved the dependency support discrepancy by removing the engine. TEST-01 severity 3 was discovery by test reviewer gpt-6.1-sol/high; tester gpt-6.1-sol/high resolved successful-output coverage with executable metadata/mutation assertions.

[Iteration 02 findings](iterations/02/REVIEW_SUMMARY.json): TEST-02 severity 4 was discovery by test reviewer gpt-6-luna/medium (real-filter oracle bypass); TEST-03 severity 3 by test reviewer gpt-6.1-sol/high (omitted pagination undetected); TEST-04 severity 3 by test reviewer gpt-6.1-sol/high during re-review (earlier query overwritten). Tester gpt-6.1-sol/high fixed them, with both test profiles verifying. Current tests retain meaningful output/filter coverage; historical stable-pagination requirements and their tests are explicitly retired by iteration 03. No accepted debt remains from those findings.

## Limits and delivery

Hosted workflow scheduling, tag creation, upload/download and retention deletion were not exercised. A final tag read and release creation are not atomic. Release packaging was not rerun; an earlier Windows ZIP probe lacked zip and is not claimed successful. Current Linux make does not establish native Windows runner behavior. No releases were published/renamed/deleted and the stable release PR was not merged.

No matching GEN-19 plan specification/index entry exists, so no lifecycle move or downstream promotion is needed. Parent owns commits, integration, push and PR #98 update. Source ownership for workflow/guidance and main tests remained coder/tester gpt-6.1-sol/high; parent made the single small integration correction above.

Telemetry is refreshed only at the final step using explicit Codex source and chore type. Previous 24 session IDs must be preserved; current source tree contains 34 sessions. The installed shared agentic-lens command is unavailable on PATH, so the freshly built current host native is used directly, as supported for source-development execution; its generate help confirms the manual flags. No shared release cache or pin is changed. Raw transcripts/probe logs stay outside Git; final ledger validation follows extraction.

Telemetry recovery: the first current-native generation correctly refused legacy ledger owners without capture provenance. Exact source metadata verified all 34 sessions belong to this physical project. Supported capture request/bind commands established those source-verified GEN-19 associations using native challenge tokens, without printing tokens, installing hooks, deleting facts or bypassing authorization. A final fresh-nonce extraction follows this recovery.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 8
- **Active Agent Duration:** 2h 45m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 322.7k tokens (72% of 450k healthy ceiling, 0 compactions) | Subagents peak 153.5k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.2 avg turns/run (peak: 5 turns) | Planner 3.0, Coder 3.0, Tester 5.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 121,517 reasoning tokens (30.1% of generated output)
- **Tool Health & Self-Correction:** 599 tool calls (27 self-corrections, 95.5% success rate)

### Anomalies & Health Flags
- **Status:** 20 execution anomalies detected (13 CRITICAL, 7 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (200.9k tokens) — Context tokens reached 200.9k, breaching the 200k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (15 cycles) — Subagent roles 'coder' and 'tester' alternated for 15 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (10 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (7 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (6 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (7 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (15 cycles) — Subagent roles 'coder' and 'tester' alternated for 15 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (10 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 10 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (7 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (6 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (7 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 7 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context saturation spike (+60.3k tokens) — Context tokens surged by 60.3k to 60.3k tokens.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (160.2k tokens) — Context tokens reached 160.2k, breaching the 160k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (322.7k tokens) — Context tokens reached 322.7k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 150,115 | 6,023 | 1,037,056 | 87.4% | 82.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 96,309 | 18,254 | 5,313,280 | 98.2% | 126.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 84,531 | 30,094 | 7,246,208 | 98.8% | 153.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 673,275 | 142,773 | 41,737,600 | 98.4% | 322.7k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 6 | 317,122 | 15,482 | 1,709,824 | 84.4% | 69.2k | HEALTHY | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **reviewer-behavior** | 8 | 384,868 | 23,588 | 2,905,088 | 88.3% | 71.7k | HEALTHY | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **reviewer-data** | 6 | 225,865 | 17,304 | 1,794,816 | 88.8% | 63.1k | HEALTHY | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **reviewer-tests** | 10 | 406,341 | 28,905 | 3,198,848 | 88.7% | 60.3k | HEALTHY | Unavailable (0 observed, 0 partial, 10 unavailable / 10 invocations) |
| **Total** | **34** | **2,338,426** | **282,423** | **64,942,720** | **96.5%** | **322.7k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 34 unavailable / 34 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 15 | 659,710 | 32,195 | 3,131,264 | 82.6% | Unavailable (0 observed, 0 partial, 15 unavailable / 15 invocations) |
| **gpt-6.1-sol** | 19 | 1,678,716 | 250,228 | 61,811,456 | 97.4% | Unavailable (0 observed, 0 partial, 19 unavailable / 19 invocations) |
| **Total** | **34** | **2,338,426** | **282,423** | **64,942,720** | **96.5%** | **Unavailable (0 observed, 0 partial, 34 unavailable / 34 invocations)** |
