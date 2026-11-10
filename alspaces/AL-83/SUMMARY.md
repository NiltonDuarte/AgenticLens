# AL-83: Public AgenticLens presentation

The implementation replaces the minimal Jekyll welcome page with an English product presentation for developers and engineering leads. It explains the workflow, three investigation use cases, harness modes, cost/evidence limits and local architecture. Four captioned SVG illustrations contain no measured results or private session captures. Setup links preserve exact release selection and distinguish current development capabilities from installed releases.

The existing single Pages pipeline, roadmap routes and embedded dashboard boundary are preserved. The builder stages five named presentation assets and verifies the generated names and bytes before replacing output. README links to the presentation. Scope was approved by the user's “$agentic-lens-dev continue” instruction on 2026-11-10; see the [contract](CONTRACT.md) and [plan](iterations/01/PLAN.md).

Baseline was `59f7ddeeed4ad4b95283acb2d0fd445317536057` on `AL-83/present-agenticlens-on-github-pages`, with an initially clean checkout. Reviewed product/test/docs fingerprint: `8d6e9221141ad2bd252179ed98d10be52a574dae1afd603b583f123edd54d495`. [Verification](iterations/01/TEST_RESULTS.md) records native Jekyll: 8 tests/335 assertions, no failures/errors/skips; project/root builds: 103 HTML pages/244 local links each, no errors; Chromium 320/375/768/1440px: no overflow or failed assets, working keyboard skip/deep anchors, zero axe WCAG A/AA violations and JavaScript-disabled content. Tester gpt-6.1-sol/high discovered a brand accessible-name mismatch; coder fixed it and tester verified the correction.

Four review focuses ran independently under gpt-6.1-sol/high and gpt-6-luna/medium. Product hashes matched throughout. Parent consolidated approval for implementation/PR delivery followed the integrated `make` pass at `a648bba10fea51297cf6f5cb1b3996c8aef74a67`; fetch/merge of `origin/main` was already up to date. All source hashes are unchanged from the reviewed candidate. Full gate passed Go units/race/integration, frontend units/typecheck/lint/integration (686 integration tests), distribution/release/install/orchestration checks and multi-platform native builds. A production Jekyll build at that commit also passed. Both data reviewers found unchanged output-replacement debt; the parent verified identical baseline ordering and classifies it as nonblocking existing debt. Pages builder maintainers own follow-up transactional replacement. Luna requested changes on this point; its dissent remains recorded. Tests Luna requested copy assertions; the parent refutes this as a blocker because current required static statements were verified, and literal copy snapshots are not an acceptance requirement. No current content defect was reported. Behavior/architecture and tests Sol approved; data Luna and tests Luna initially requested changes, reconciled above. Exact model/effort was observed in Codex source metadata. The permanent [review findings](iterations/01/REVIEW_SUMMARY.json) retain all discovery credit and dissent; the tester is separately credited for verification. No introduced blocker or material local evidence gap remains.

Publication remains incomplete until an authorized merge and successful main deployment, followed by public page/asset/anchor/roadmap checks. Pages settings already use GitHub Actions with HTTPS, but the live welcome page is not evidence of this change's deployment. No merge is authorized. Manual assistive-technology testing and remote deployment were not performed. Final CLI extraction verified all 12 Codex source sessions: this parent, seven native children and four explicitly selected T3 review roots. The four roots were preserved when refreshing the parent tree; no prior cross-harness ledger existed. No parent links or session facts were synthesized. The source reports an open turn and unverified invoking caller, so duration is provisional at extraction. Raw transcripts remain outside Git.

<!-- agentic-lens:metrics:start -->
### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 16
- **Active Agent Duration:** 1h 12m (recorded duration including 512s provisional)
- **Context Window Health:** HEALTHY — Root agent peak 139.8k tokens (31% of 450k healthy ceiling, 0 compactions) | Subagents peak 114.7k (Tester)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.3 avg turns/run (peak: 2 turns) | Planner 1.0, Coder 2.0, Tester 2.0, Reviewer-architecture 1.0, Reviewer-behavior 1.0, Reviewer-data 1.0, Reviewer-tests 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 23,207 reasoning tokens (20.6% of generated output)
- **Tool Health & Self-Correction:** 293 tool calls (25 self-corrections, 91.5% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 CRITICAL).
- [CRITICAL] **SUBAGENT_REWORK_LOOP**: Subagent rework loop: coder <-> tester (8 cycles) — Subagent roles 'coder' and 'tester' alternated for 8 consecutive cycles.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 52,547 | 2,580 | 424,960 | 89.0% | 53.5k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 53,395 | 14,101 | 1,348,096 | 96.2% | 86.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 54,504 | 16,110 | 6,349,824 | 99.1% | 114.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 5 | 337,471 | 38,111 | 12,441,472 | 97.4% | 139.8k | HEALTHY | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **reviewer-architecture** | 1 | 49,012 | 3,482 | 373,504 | 88.4% | 63.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-behavior** | 1 | 74,454 | 5,168 | 742,400 | 90.9% | 77.2k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-data** | 1 | 51,066 | 6,082 | 870,400 | 94.5% | 67.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **reviewer-tests** | 1 | 112,286 | 4,005 | 377,984 | 77.1% | 57.3k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **12** | **784,735** | **89,639** | **22,928,640** | **96.7%** | **139.8k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 205,315 | 11,682 | 1,221,632 | 85.6% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 8 | 579,420 | 77,957 | 21,707,008 | 97.4% | Unavailable (0 observed, 0 partial, 8 unavailable / 8 invocations) |
| **Total** | **12** | **784,735** | **89,639** | **22,928,640** | **96.7%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

Capture observation finalized: 2026-11-10T22:27:50.439897456Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 4; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 4; reserved bytes: 2232452. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1c788-28f2-7381-87b7-1b99c3734371: current; last successful source read 2026-11-10T22:27:50.398093036Z; accepted source checkpoint 2026-11-10T22:17:06.913Z.
Owner codex/01a1c786-a2de-74d2-b7ed-e96defd2ef79: current; last successful source read 2026-11-10T22:27:50.412812733Z; accepted source checkpoint 2026-11-10T22:15:09.251Z.
Owner codex/01a1c78a-2c46-7711-97bb-def193cce8bd: current; last successful source read 2026-11-10T22:27:50.363083058Z; accepted source checkpoint 2026-11-10T22:19:16.967Z.
Owner codex/01a1c789-1428-7451-87f4-12f661bc3910: current; last successful source read 2026-11-10T22:27:50.379977209Z; accepted source checkpoint 2026-11-10T22:17:41.744Z.
<!-- agentic-lens:metrics:end -->
