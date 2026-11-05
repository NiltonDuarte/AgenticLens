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
- Final telemetry is limited to the root and 27 authoritatively linked native children (28 owners). Twenty delegated gpt-6-luna reviews have no native parent metadata and are excluded; review execution/attribution is retained separately. Pricing is unavailable. Extraction verified the exact root, all 27 native children and preservation of every prior owner from the 19-owner Codex-only baseline. All 28 recorded invocations have unavailable pricing; no zero or complete total cost is claimed. Canonical ledger checks confirm positive-only provisional values and no retired duration metadata. Unsupported configured legacy capture state is preserved; explicit isolated capture state provides manual recovery. Raw source inventories remain outside Git.

Consolidated local approval: APPROVED after all eight correction cells and integrated gates. Initial policy uncertainty is resolved by the approved provisional-estimate semantics. Earlier unchanged scopes carry forward with the dependency rationale above. PR #121 is the delivery vehicle; hosting checks on the final pushed head are reported separately. No release or remote merge is authorized.

## Iteration 04 planning delta: telemetry checkpoint feedback

Subsequent user steering requests a [checkpoint-feedback plan](iterations/04/PLAN.md), with a default threshold of three minutes (180 seconds), every exact feedback variant and its triggering case, including multiple affected owners. This is authorized plan authorship; implementation approval is not inferred. The completed iteration 03 implementation and verification remain unchanged.

The proposed extension identifies session/invocation owners and suggests a real turn boundary only when the executing caller and native open interval are verified. Other owners, unknown caller identity and unavailable source evidence receive distinct qualified feedback. It must not equate extraction root with caller, fabricate completion, promise automatic continuation without verified support, or claim tokens/cost are final. Same-turn regeneration does not close an interval. Human tool feedback adds no ledger fields, duration reasoning or frontend behavior, and does not alter numeric accounting or portable refresh. This extends tool guidance while preserving iteration 03's rejection of mandatory liveness/accounting gates. Verification must cover threshold boundaries, ownership/identity, interval versus subtotal distinctions, fallback and multi-owner messages before any implementation is delivered. User steering is paraphrased.

Further steering on 2026-11-06 removes aggregation-only feedback from iteration 04. Individually known intervals shorter than 180 seconds produce no message even when their combined subtotal crosses the threshold, and must not fall through to F04. Qualified F04 feedback remains for unavailable individual interval evidence. Five exact message variants remain; implementation is still pending approval.

On 2026-11-06 the user authorized implementation of iteration 04 and renamed the unsupported-continuation notice to F05. The removed aggregation-only case remains excluded. Approval covers the five feedback templates, default 180-second threshold and optional threshold configuration, transient ownership/evidence and qualified fallbacks. Existing iteration 03 verification remains historical evidence for unchanged accounting; new feedback verification/review and final delivery are pending.

The iteration 04 [native capability investigation](iterations/04/RUNTIME_CAPABILITIES.md) distinguishes real source-open interval evidence from an executing-caller handoff. No installed generic-generate transport was verified for stronger F01/F02 advice or next-turn continuation. Production feedback therefore uses qualified F03/F04 where appropriate; F01/F02/F05 remain gated on independently verified caller/current-interval context. This is an approved fallback boundary, not proof of process liveness or a new accounting gate. No session was poked or automatically continued.

Iteration 04 CP-02 is frozen under manifest512721e4192e55dcad98f2f1d7a080360178b73ce016c0225b50840dc69ff629. All six behavior/data/test-effectiveness correction cells approve on gpt-6.1-sol/high and gpt-6-luna/medium; both initial architecture approvals carry forward with unchanged layer/contracts and a three-path correction. Shared Sol behavior/data discovery of native-idle qualification was corrected, Sol test supersession coverage strengthened, and the Luna stderr hypothesis refuted by descriptor access mode and failed-write preflight. Native capability limits remain explicit. Parent integration/full make/standalone and final delivery are pending.

The telemetry below remains the published iteration 03 snapshot; it will be regenerated at the final delivery step after implementation and review.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 39
- **Active Agent Duration:** 6h 18m (recorded duration including 768s provisional)
- **Context Window Health:** BLOATED — Root agent peak 450.2k tokens (100% of 450k healthy ceiling, 1 compactions) | Subagents peak 363.1k (Coder)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.6 avg turns/run (peak: 7 turns) | Planner 4.0, Coder 7.0, Tester 2.8, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 1.4:1 (+1,283 test lines / +891 prod lines)
- **Code Delivered:** +2,174 / -151 lines across 61 code files (excluding docs)
- **Reasoning Effort:** 214,496 reasoning tokens (25.7% of generated output)
- **Tool Health & Self-Correction:** 1,574 tool calls (106 self-corrections, 93.3% success rate)

### Anomalies & Health Flags
- **Status:** 12 execution anomalies detected (7 CRITICAL, 5 WARNING).
- [CRITICAL] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (450.2k tokens) — Context tokens reached 450.2k, breaching the 450k critical ceiling.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (21 cycles) — Subagent roles 'coder' and 'tester' alternated for 21 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-tests <-> reviewer-data (6 cycles) — Subagent roles 'reviewer-tests' and 'reviewer-data' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (6 cycles) — Subagent roles 'coder' and 'tester' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-tests (6 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-tests' alternated for 6 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (30 cycles) — Subagent roles 'coder' and 'tester' alternated for 30 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (7 cycles) — Subagent roles 'coder' and 'tester' alternated for 7 consecutive cycles.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (265.0k tokens) — Context tokens reached 265.0k, breaching the 250k warning ceiling.
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (250.4k tokens) — Context tokens reached 250.4k, breaching the 250k warning ceiling.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> tester (3 cycles) — Subagent roles 'reviewer-architecture' and 'tester' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-data <-> reviewer-behavior (3 cycles) — Subagent roles 'reviewer-data' and 'reviewer-behavior' alternated for 3 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-tests (4 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-tests' alternated for 4 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 352,850 | 17,799 | 2,950,528 | 89.3% | 147.6k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 1,061,045 | 91,109 | 43,572,864 | 97.6% | 363.1k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 4 | 1,279,879 | 123,440 | 43,072,640 | 97.1% | 222.0k | HEALTHY | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **Unknown** | 1 | 1,183,086 | 239,514 | 91,645,824 | 98.7% | 450.2k | BLOATED (100%) | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 6 | 473,532 | 39,898 | 6,420,992 | 93.1% | 159.4k | HEALTHY | Unavailable (0 observed, 0 partial, 6 unavailable / 6 invocations) |
| **reviewer-behavior** | 5 | 393,324 | 34,601 | 5,138,688 | 92.9% | 108.5k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-data** | 5 | 418,224 | 36,828 | 5,931,776 | 93.4% | 111.1k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-tests** | 5 | 376,742 | 37,447 | 5,731,584 | 93.8% | 109.5k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **Total** | **28** | **5,538,682** | **620,636** | **204,464,896** | **97.4%** | **450.2k** | **BLOATED** | **Unavailable (0 observed, 0 partial, 28 unavailable / 28 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 28 | 5,538,682 | 620,636 | 204,464,896 | 97.4% | Unavailable (0 observed, 0 partial, 28 unavailable / 28 invocations) |
| **Total** | **28** | **5,538,682** | **620,636** | **204,464,896** | **97.4%** | **Unavailable (0 observed, 0 partial, 28 unavailable / 28 invocations)** |

Capture observation finalized: 2026-11-05T20:03:53.677127982Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
