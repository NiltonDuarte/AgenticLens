# AL-64: active-turn duration

PR: https://github.com/NiltonDuarte/AgenticLens/pull/121

Implemented the approved [iteration 03 numeric contract](iterations/03/PLAN.md): active_duration_seconds includes the optional positive provisional_seconds portion. Zero is omitted. No reasoning field, setting, console explanation or frontend reasoning behavior is added. Source-open intervals accrue through an explicit calculation cutoff; a stopped process whose source interval remains open can overestimate time. Zero or absence does not establish completed coverage. The original [CONTRACT.md](CONTRACT.md) remains the frozen baseline; iteration 03 supersedes its richer duration representation and the unimplemented iteration 02 hybrid proposal. User approval and steering are paraphrased.

## Delivered behavior

The pure duration core retains deterministic interval/rounding rules and numeric rollups. Codex now counts unfinished turns; independently investigated OpenCode v1/v2 paths account for their native unfinished assistant/tool/text activity. Terminal events, future boundaries, invalid starts/completions and fallback cutoffs retain explicit handling. Completion replaces provisional snapshots; repeated refresh avoids duplicate owners or duration, and unavailable sources retain recorded facts.

Backend readers accept original v6, earlier draft evidence and lean numeric inputs. Canonical writes use only the total plus positive provisional scalar. Valid old/new dual forms must agree; aliases and contradictory values reject before historical typed decoding. Completed-migration cleanup preserves unrelated raw facts, optional-category presence and rollback bytes. Portable history needs no private receipt. CLI summaries, API projections and dashboard duration consumers show numeric recorded/provisional values. Upstream optional metrics and shared-session warnings remain preserved.

Investigation: [source-family analysis](iterations/01/INVESTIGATION.md), [native OpenCode evidence](iterations/01/RUNTIME_EVIDENCE.md), [lean implementation](iterations/03/IMPLEMENTATION.md) and [native Codex refresh](iterations/03/NATIVE_REFRESH_EVIDENCE.md).

## Review and verification

The implementation author is coder gpt-6.1-sol/high. Independent deterministic and native runtime testers use gpt-6.1-sol/high. Behavior, data/failure, architecture and test effectiveness each ran on gpt-6.1-sol/high and gpt-6-luna/medium. Discovery initially requested changes; all eight bounded correction rechecks approved LEAN-02, the 399-path manifest e40b0bb87a233eeeec0da9716cfc51311c1763ecc8752c95b280d6f92d7fc3b8 on baseline c68fbc90. Verification credit is separate from discovery credit.

The original six corrections and their discovering reviewers are retained in [iteration 01 compact findings](iterations/01/REVIEW_SUMMARY.json). The numeric rewrite findings and exact shared reporter tuples are retained in [iteration 03 compact findings](iterations/03/REVIEW_SUMMARY.json). Native/adversarial testers gpt-6.1-sol/high discovered the current-schema normalization gap; coder gpt-6.1-sol/high discovered the current-input alias gap. Cleanup publication was independently reported by behavior/data/architecture gpt-6.1-sol/high and data gpt-6-luna/medium. Historical alias decoding was independently reported by data gpt-6.1-sol/high and data gpt-6-luna/medium. Coder gpt-6.1-sol/high fixed those gaps, and tester gpt-6.1-sol/high verified original-code red overlays followed by passing affected race checks. The Sol test-effectiveness recheck also independently passed 97 cause/publication/compatibility probes.

LEAN-02 was committed as 64da0bfa; main 29699cc6 was merged as f8aeac2b. No application production code changed during this integration. Tester gpt-6.1-sol/high resolved schemas_test.go by retaining all 41 functions: AL-64 numeric assertions remain byte-identical and upstream policy/anomaly assertions retain their stronger causes/identity checks. Targeted adapter race checks pass. Parent integration assessment carries forward reviewed feature behavior because the other imported application changes only configure test/build orchestration; the full integrated gate exercises both test lanes and asset build.

Required make passed exit 0 on f8aeac2beadce9f5d46ea0c3cc25dbc53bfc2401: Go units and every race package, frontend 190 unit plus 678 integration tests (868 total), typecheck/lint/vet, 45 Python release/installation/capture/orchestration checks, canonical assets and five platform builds. Standalone binary commit f8aeac2b served exact HTML/JS/CSS matching web/dist and the embedded tree; assets index-zSnCKFwx.js and index-BGH-Pi_l.css. The first two disposable probe attempts omitted required config/schema fields; the corrected probe used the project config with isolated state. No product change was needed, and its server/workspace were cleaned. Raw make/asset logs remain outside Git under /tmp/al64-final-delivery/.

Affected migration/adapters/backfill/capture race and selected CLI race checks pass. The correction tests fail on 40 historical invalid-input cases and three cleanup publishers with original production overlays, then pass with 90 historical cases and three command variants after correction. Frontend verification before integration passed 868 tests, typecheck and lint. An earlier broad Go race run failed one unchanged capture-install deadline; its unchanged isolated retry passed. No timeout or assertion was weakened, and that failed run is not reported as a clean full-suite result. [Detailed test evidence](iterations/03/TEST_REPORT.md) retains command/revision limits.

Native OpenCode 1.19.5 and 2.2.0 probes used isolated sources. OpenCode v1 cancellation demonstrated the accepted stale-open interval risk. The native Codex probe normalized an authentic copied 19-owner ledger without losing numeric facts, then source-backed recomputation discovered one legitimate new child and verified completion/repeat replacement. A prior fresh copied-source ownership rejection and an intermediate normalization failure remain documented; neither is presented as success. The actual ticket ledger stayed unchanged during experiments. No new OpenCode provider experiment was run for the numeric rewrite; unchanged native source semantics plus current deterministic three-family generation/refresh fixtures are reused.

## Remaining limits and debt

- Pre-expansion strict readers can reject expanded v6 scalar output. Upgrade consumers before writer rollout; this ticket authorizes no release.
- INT-DATA-01: disabled-delegation registered Codex refresh may overwrite retained registry delegations. Data reviewer gpt-6-luna/medium discovered the concern; parent source comparison refuted its introduced-merge classification because the block was byte-identical to upstream main. Source-traced, not independently runtime-reproduced; not fixed. Owner: capture adapter maintainer; follow-up unassigned.
- LEAN-READ-02: legacy nested numeric null is interpreted as zero. Behavior and test-effectiveness gpt-6.1-sol/high and test-effectiveness gpt-6-luna/medium reported it. Exact baseline probes refute introduced acceptance; new flat null correctly rejects. Historical interpretation is retained. Owner: backend input compatibility maintainer; stricter legacy policy follow-up unassigned.
- Iteration 03 publication covered28 native sessions, excluding20 unlinked delegated reviews; pricing was unavailable. Current iteration04 coverage and remaining native-capability limits are recorded below.

Consolidated local approval: APPROVED after all eight correction cells and integrated gates. Initial policy uncertainty is resolved by the approved provisional-estimate semantics. Earlier unchanged scopes carry forward with the dependency rationale above. PR #121 is the delivery vehicle; hosting checks on the final pushed head are reported separately. No release or remote merge is authorized.

## Iteration 04: owner-scoped checkpoint feedback

The user approved implementation on 2026-11-06 after removing aggregation-only feedback and renaming the continuation notice to F05. The [approved plan](iterations/04/PLAN.md) contains all five exact templates and cases. Explicit generate adds nonfatal stderr feedback after successful publication, defaulting to 180 seconds via generation.feedback.provisional_threshold_seconds; zero disables it. Individually known short intervals remain silent even when their subtotal crosses the threshold. Feedback never pools owners/models/invocation rows, fabricates completion, schedules a new turn or enters ledger/summary/API/frontend payloads. Numeric duration/token/cost accounting remains unchanged.

Generic native caller/current-interval handoff remains unverified. Production supports exact source-open F03 and qualified F04; F01/F02/F05 are typed selector/renderer cases gated on independently verified context, not claimed installed native functionality. [Native capability evidence](iterations/04/RUNTIME_CAPABILITIES.md) separates installed help/protocol declarations, authentic Codex source-open observations and unexecuted caller/continuation cases. No session was poked or automatically continued. A real turn boundary can improve recorded usage freshness, but this feature does not establish token/cost finality.

Coder and independent tester identities are source-observed gpt-6.1-sol/high. [Implementation](iterations/04/IMPLEMENTATION.md) and [test evidence](iterations/04/TEST_REPORT.md) retain source ownership, commands, earlier failed/expected-red probes and limits. CP-02 on baseline d8a7c14c has 411-path manifest 512721e4192e55dcad98f2f1d7a080360178b73ce016c0225b50840dc69ff629; the code was committed as 73572098. Fetch/merge confirmed existing main 29699cc6 was already integrated with no new conflicts or production changes.

Initial review comprised four focuses on both profiles: six scoped approvals and two Sol behavior/data changes requested. All six behavior/data/test-effectiveness correction cells subsequently approved on gpt-6.1-sol/high and gpt-6-luna/medium. Initial architecture approvals on both profiles carry forward: the correction changes only private v2 adapter metadata qualification and two test files, with no new layer/domain/public/persistence contract; data rechecks specifically exercise optional-read failure semantics. Other source grouping/config/selector/renderer/capture/API/frontend inputs are unchanged. Final compact [findings and shared discovery credit](iterations/04/REVIEW_SUMMARY.json) retain:

- CP-IN-01, severity4: coder and tester gpt-6.1-sol/high independently exposed incomplete/historical source evidence overclaim. Coder gpt-6.1-sol/high corrected transient completeness/SourceOpen; tester gpt-6.1-sol/high verified partial/future-terminal numeric preservation.
- CP-IN-02, severity3: orchestrator gpt-6.1-sol/high identified synthetic Codex accounting IDs presented as native intervals. Coder gpt-6.1-sol/high tracked native-ID presence; tester gpt-6.1-sol/high verified qualified fallback without changing numeric accounting.
- CP-REV-01, severity4: behavior and data reviewers gpt-6.1-sol/high independently reproduced native v2 idle overridden by stale running-tool evidence. Coder gpt-6.1-sol/high qualified feedback only; tester gpt-6.1-sol/high proved eight original-source failures then13-case corrected controls.
- CP-TEST-01, severity3: test-effectiveness reviewer gpt-6.1-sol/high demonstrated missing future-message supersession guards. Tester gpt-6.1-sol/high added both-family assertions that fail guard-removal mutations; no product fix was needed.
- CP-OBS-01, severity0/refuted: test-effectiveness reviewer gpt-6-luna/medium raised a hypothetical root-permission weakness. Parent/source inspection and tester gpt-6.1-sol/high verified an O_RDONLY descriptor rather than chmod-only protection; an explicit failed-write preflight now proves the setup. No product defect is claimed.

Verification credit is separate from discovery: Sol behavior independently exercised 13 optional-query-failure controls plus seven real generate controls; Sol data injected query/scan/missing-row failures preserving numeric facts; Sol test effectiveness independently exercised 13 optional-query-failure controls. Luna data executed the focused idle race test; other Luna rechecks reused reported evidence and source inspection. Public unit190 frontend, affected race and focused CLI gates pass. Required make passed exit 0 on 73572098: every backend unit/race package (full CLI 116.521s), frontend 190 unit plus 678 integration tests (868 total), typecheck/lint/vet, 45 Python release/installation/capture/orchestration checks and five platform builds. main 29699cc6 was already integrated; no integration source changes. The standalone host binary reports73572098 and serves exact canonical/embedded HTML, index-zSnCKFwx.js and index-BGH-Pi_l.css from an isolated configured empty workspace; server/workspace cleaned. Fresh generate help documents default180/zero-disable and extraction-root versus caller identity. Raw make/help/asset logs remain outside Git under /tmp/al64-04-final-delivery/.

Final source-backed extraction verified root 01a1a7d8-b9d2-7383-a6f7-3c7002ba88e6 and all 38 native children (39 owners), preserved every prior owner from the 28-owner Codex-only baseline, and added no retired duration or checkpoint fields. All 39 invocation costs remain unavailable; no zero/complete dollar total is claimed. Twenty-seven delegated gpt-6-luna reviews lack native parent links and remain excluded. Configured unsupported legacy capture state is preserved; explicit isolated capture state enabled manual generation. Real native generation emitted exactly one F03 on stderr for the root's source-open turn, with 544s provisional at the 180s threshold; stdout, ledger and summary contain no runtime feedback. Native source ID/start independently match the rollout; this does not verify executing-caller handoff or next-turn continuation. No same-turn retry or fabricated terminal was performed. Raw inventories/logs remain outside Git.

Consolidated local approval: APPROVED after six correction cells, two justified carried architecture approvals and required integrated/standalone gates. No unresolved product or test finding remains in scope; native caller/current-interval handoff and natural completion-to-next-turn refresh remain unverified. PR121 is the delivery vehicle; hosting checks on the final pushed head are reported separately. No release or remote merge is authorized.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 57
- **Active Agent Duration:** 8h 12m (recorded duration including 544s provisional)
- **Context Window Health:** BLOATED — Root agent peak 450.2k tokens (100% of 450k healthy ceiling, 1 compactions) | Subagents peak 363.1k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.5 avg turns/run (peak: 7 turns) | Planner 2.5, Coder 4.5, Tester 2.6, Explorer 1.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.5:1 (+1,967 test lines / +1,321 prod lines)
- **Code Delivered:** +3,288 / -190 lines across 74 code files (excluding docs)
- **Reasoning Effort:** 307,322 reasoning tokens (26.4% of generated output)
- **Tool Health & Self-Correction:** 2,161 tool calls (143 self-corrections, 93.4% success rate)

### Anomalies & Health Flags
- **Status:** 16 execution anomalies detected (9 CRITICAL, 7 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (21 cycles) — Subagent roles 'coder' and 'tester' alternated for 21 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-data (6 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-data' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (6 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (30 cycles) — Subagent roles 'coder' and 'tester' alternated for 30 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (7 cycles) — Subagent roles 'coder' and 'tester' alternated for 7 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> coder (23 cycles) — Subagent roles 'tester' and 'coder' alternated for 23 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (9 cycles) — Subagent roles 'coder' and 'tester' alternated for 9 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (265.0k tokens) — Context tokens reached 265.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.4k tokens) — Context tokens reached 250.4k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> tester (3 cycles) — Subagent roles 'reviewer-architecture' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 4 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (3 cycles) — Subagent roles 'coder' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-behavior' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 2 | 414,853 | 26,097 | 3,597,312 | 89.7% | 147.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **coder** | 2 | 1,305,002 | 121,573 | 50,659,328 | 97.5% | 363.1k | ELEVATED | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **tester** | 5 | 1,507,677 | 156,257 | 51,962,112 | 97.2% | 222.0k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Unknown** | 1 | 1,489,844 | 343,287 | 128,477,824 | 98.9% | 450.2k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **explorer** | 1 | 54,209 | 8,882 | 898,816 | 94.3% | 72.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 7 | 560,048 | 45,290 | 7,528,832 | 93.1% | 159.4k | HEALTHY | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **reviewer-behavior** | 7 | 542,868 | 50,874 | 7,629,184 | 93.4% | 108.5k | HEALTHY | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **reviewer-data** | 7 | 557,472 | 50,971 | 8,134,400 | 93.6% | 111.1k | HEALTHY | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **reviewer-tests** | 7 | 517,920 | 53,492 | 7,794,560 | 93.8% | 109.5k | HEALTHY | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **Total** | **39** | **6,949,893** | **856,723** | **266,682,368** | **97.5%** | **450.2k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 39 unavailable / 39 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 39 | 6,949,893 | 856,723 | 266,682,368 | 97.5% | Unavailable (0 observed, 0 partial, 39 unavailable / 39 invocations) |
| **Total** | **39** | **6,949,893** | **856,723** | **266,682,368** | **97.5%** | **Unavailable (0 observed, 0 partial, 39 unavailable / 39 invocations)** |

Capture observation finalized: 2026-11-05T23:49:38.530524927Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
