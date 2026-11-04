# FIX-17: Explicit manual capture-state recovery

Explicit manual generation can use a healthy/fresh capture directory despite unsupported configured coordinate mappings. Selected-state validation remains mandatory. Healthy configured manifests, literal configured paths and discoverable installed integrations retain ownership, pin/template/native/wrapper/source and owned-integration checks. Default, native and automatic capture remain strict. The old registry is not migrated or reset; portable ledger authority, missing/other-harness facts, schemas and metric formulas are unchanged.

## Scope and delivery

Initial tree clean at 86bea597afe7208cb1836f7cc5747083f7b95f47. Existing GEN-13 documentation artifacts preserved; FIX-17 allocation checked against workspace and ticket index. Branch fix/FIX-17-telemetry-capture; reviewed implementation d4a9e6edb7b36247a5c35db15c2e0e64a050d775. Main integration was already up to date. No matching plan specification/index entry or downstream dependency required lifecycle changes. Later commits are ticket documentation/telemetry only.

The user explicitly preapproved this recovery boundary after providing GEN-21 evidence, then required real-source final capture for this ticket and authorized early source plus final completion notification to the controlling thread. Source commit was sent without a publication-readiness claim. The controller reports clean source-equivalent integration as GEN-21 commit 12048ab5. Its first combined gate hit SQLITE_BUSY in unchanged TestCoordinationDirectoryNamesConcurrentFirstOpenAndStableReopen (shared_state_test.go:353); five isolated race-enabled repetitions passed, and its full gate was being rerun. This is external integration evidence, not a FIX-17 code change or a claim that retry resolves the contention defect. GEN-21 owns integration and its exact root plus eleven-child generation; this root has not changed its artifacts.

By layer: CLI shell selects the recovery policy and retains installation admission; capture adapter classifies only known missing coordinate-table/mapping outcomes with a typed sentinel. No domain or presentation changes. Contract and steering: [CONTRACT.md](CONTRACT.md), [iteration plan](iterations/01/PLAN.md).

## Review and verification

Reviewed final four-file diff 6cc147ec3e399a59e5589d7a19df5882cd98571b0e9db848c1cdc73772aeb439, committed as d4a9e6edb7b36247a5c35db15c2e0e64a050d775. Behavior, data/failure, architecture and test-effectiveness reviewers each ran under gpt-6.1-sol/high and gpt-6-luna/medium. All eight final cells approved. Native session_meta/turn_context inspection confirmed those model/effort identities and this root plus eleven children.

Two severity-6 findings were corrected within iteration 01. Behavior reviewer gpt-6.1-sol/high found loss of the healthy configured manifest when anchor and hooks both vanish. Architecture reviewer gpt-6.1-sol/high and data/failure reviewer gpt-6-luna/medium independently found loss of a surviving manifest at a literal configured coordination directory when mappings are unsupported. Coder gpt-6.1-sol/high fixed both by retaining healthy configured checks and checking the literal known configured path during typed recovery, without guessing unavailable coordinates. Both discovering focuses reapproved; the other final reviews found no additional defects.

Tester gpt-6.1-sol/high independently reproduced baseline failure, demonstrated that removing installation validation makes the regression fail, and verified the first correction. Test-effectiveness reviewer gpt-6.1-sol/high independently removed the final literal-path check with a disposable overlay: the new regression detected incorrect publication; unmodified control passed. Discovery credit and verification are distinct.

Nine isolated recovery/installation cases and installed-state drift coverage pass under race detection. Coverage includes fresh/healthy success; default, incompatible selected, automatic and native rejection; byte-identical old registry/config; real missing-hook and missing-anchor/known-manifest failures. Existing exact-source/project and portable-history tests preserve unavailable sessions, children and other-harness facts without new receipts. Prior corrected CLI/capture race suites passed (104.554s/62.046s); final targeted race passed (4.172s).

Final consolidated source approval: APPROVED after all eight final focused cells and required integrated make passed on d4a9e6edb7b36247a5c35db15c2e0e64a050d775. Retained [compact findings](iterations/01/REVIEW_SUMMARY.json) preserve both findings and every discovering persona/model/effort tuple. No unresolved findings or accepted debt were hidden by cleanup. Detailed working review/test reports were removed after consolidation; raw source transcripts remain outside Git.

Integrated verification: make completed with exit 0, including dependency preparation, static checks, release/installation/distribution tests, all Go race packages, all 824 frontend tests, frontend asset synchronization and five platform binary builds. Host native version confirms implementation commit d4a9e6e. No test stages were silently skipped; packages with no test files are not test skips. Full-check raw logs remain private in /tmp/fix17-review. The first full run visibly passed the same checks/tests but lost final exit retrieval; the final logged run supplied durable exit 0 without product changes.

## Runtime limits and final telemetry

Configured capture remains incompatible and its producing version is unknown. Manual recovery does not certify automatic-capture health, reconstruct unavailable old coordinates or repair installed integrations. Real installed provenance failures intentionally remain blocking. Use a source-built native containing the implementation; the existing shared release pin does not ship this fix. Cross-platform compilation is separate from execution on those operating systems.

Final delayed real-source recovery PASSED using the reviewed native, explicit Codex root 01a1a79f-d472-7483-a987-98e38cd2f817, literal nonce c2d8f6b9154ae073 and a fresh capture override. The schema-v6 fix ledger contains exactly the root and all eleven source-proven children, all Codex; identities match native metadata. Configured database and WAL SHA-256 hashes are unchanged. [Runtime evidence](iterations/01/TELEMETRY_STATUS.json) records identities, before/after hashes, native revision and limits. First invocation recovered past the coordinate error but required explicit ticket-title; it wrote no ledger and left old state unchanged. Successful invocation used another fresh directory plus the title.

No prior FIX-17 ledger or other-harness facts existed; existing mixed-owner regression coverage remains the preservation evidence. Twelve manual source extractions are distinct from the fresh registry's zero bound capture-owner observations. Costs remain unavailable for all twelve source invocations and the root role remains Unknown as extracted; no values or provenance were manufactured. Metrics cover the final generation snapshot, excluding subsequent artifact verification and delivery calls. GEN-21's actual root/eleven-child extraction is exclusively its owner's acceptance step.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 9
- **Active Agent Duration:** 38m 13s (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 130.1k tokens (29% of 450k healthy ceiling, 0 compactions) | Subagents peak 107.6k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.7 avg turns/run (peak: 2 turns) | Planner 2.0, Coder 2.0, Tester 2.0, Reviewer-architecture 1.5, Reviewer-behavior 2.0, Reviewer-data 2.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 2.9:1 (+97 test lines / +34 prod lines)
- **Code Delivered:** +131 / -11 lines across 4 code files (excluding docs)
- **Reasoning Effort:** 21,404 reasoning tokens (18.7% of generated output)
- **Tool Health & Self-Correction:** 341 tool calls (25 self-corrections, 92.7% success rate)

### Anomalies & Health Flags
- **Status:** 3 execution anomalies detected (2 CRITICAL, 1 WARNING).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: tester <-> reviewer-behavior (9 cycles) — Subagent roles 'tester' and 'reviewer-behavior' alternated for 9 consecutive cycles.
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-behavior <-> reviewer-data (5 cycles) — Subagent roles 'reviewer-behavior' and 'reviewer-data' alternated for 5 consecutive cycles.
- [WARNING] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: reviewer-architecture <-> reviewer-data (3 cycles) — Subagent roles 'reviewer-architecture' and 'reviewer-data' alternated for 3 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 58,384 | 2,589 | 313,856 | 84.3% | 60.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 77,293 | 12,388 | 4,375,168 | 98.3% | 105.3k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 47,130 | 6,652 | 2,956,032 | 98.4% | 107.6k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 123,672 | 32,266 | 13,932,288 | 99.1% | 130.1k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-architecture** | 2 | 99,829 | 8,279 | 1,400,576 | 93.3% | 86.8k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-behavior** | 2 | 215,494 | 11,928 | 2,376,192 | 91.7% | 98.0k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-data** | 2 | 131,661 | 10,897 | 1,795,712 | 93.2% | 89.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **reviewer-tests** | 2 | 88,564 | 7,782 | 1,311,104 | 93.7% | 76.6k | HEALTHY | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **12** | **842,027** | **92,781** | **28,460,928** | **97.1%** | **130.1k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 136,943 | 11,747 | 1,472,384 | 91.5% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 8 | 705,084 | 81,034 | 26,988,544 | 97.5% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **12** | **842,027** | **92,781** | **28,460,928** | **97.1%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

Capture observation finalized: 2026-11-04T18:06:14.878396689Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
