# REV-6: measured test performance and boundaries

Consolidated review APPROVED. Delivered [measured findings, sequencing and follow-ups](iterations/01/TEST_REVIEW.md), [commands/resources](iterations/01/MEASUREMENTS.json), [syntax inventory](iterations/01/INVENTORY.md), and [compact findings with all reporting profiles](iterations/01/REVIEW_SUMMARY.json). Pure Go domain command median: 0.39 s; frontend domain Node median: 1.85 s. Full later suites remain necessary. No application, test, fixture, schema, runtime or CI changes.

## Scope, approval and revision

The user invoked agentic-lens-dev for existing REV-6 and explicitly approved CONTRACT.md/iteration 01 on 2026-11-04. Initial Git status was clean on the existing REV-6/review-test-performance-unit-integration-boundaries branch. Source baseline a84adbe5be46c2ef52ede0987a417562715d38e1 matched fetched origin/main and included AL-62 merged PR #109 (86bea597), not closed/unmerged #103. AL-74 remains separately owned. Candidate 48b1b00e committed the review artifacts; upstream merge was already current. Tracked application/test/runtime/CI source bytes are identical to baseline. Subsequent edits are documentation/telemetry and ticket lifecycle only; valid source review/check evidence carries forward without re-running unchanged suites.

Domain, adapters and presentation implementation: unchanged. Go-domain and frontend-domain early gates are recommendations; no Make target or hosted needs edge was implemented.

## Review and attribution

All ten source/measurement investigations completed with scoped APPROVED: behavior/integration, data/failure, architecture/CI, test effectiveness and performance specialist, each under gpt-6.1-sol/high and gpt-6-luna/medium. Parent issued final APPROVED after integrated make passed, with no unresolved material evidence gap for this review-only contract. Dispatch receipts and local Codex turn_context records confirmed every requested model/effort. Requests used capacity-limited waves; no provider execution overlap is claimed. One parent owned all expensive timing; reviewers reused it and did not repeat suites.

The table preserves reporting credit, including shared observations; it is not a count of new product bugs. Existing debt and conditional risks were not fixed. The AL-74 finding is a revalidation, not a claim of original discovery. Performance specialists gpt-6.1-sol/high and gpt-6-luna/medium checked the supplied timing/race/phase claims against raw evidence; those confirmations receive verification credit, not additional discovery credit. Data gpt-6-luna/medium, test effectiveness gpt-6-luna/medium and performance gpt-6-luna/medium returned no separate canonical candidates. No initial/final child verdict was INCOMPLETE or CHANGES_REQUESTED.

| Finding | Severity 0–10 | Issue | Reporting actor/model/effort | Disposition and owner |
| --- | ---: | --- | --- | --- |
| REV6-01 | 3 | Filtered CLI commands still execute the native TestMain build. | reviewer-behavior gpt-6.1-sol/high; reviewer-behavior gpt-6-luna/medium; reviewer-data gpt-6.1-sol/high | Documented setup separately; future fixture-sharing owner must preserve build identity. No implementation. |
| REV6-02 | 4 | A faster split must preserve independent native and persisted-state oracles. | reviewer-behavior gpt-6.1-sol/high | Preservation constraint adopted; all existing coverage retained. No present product defect asserted. |
| REV6-03 | 4 | Current local and hosted graphs lack an explicit pure-unit prerequisite. | reviewer-behavior gpt-6.1-sol/high; reviewer-data gpt-6.1-sol/high; reviewer-architecture gpt-6.1-sol/high; reviewer-architecture gpt-6-luna/medium | Accepted follow-up for a separate CI sequencing owner; commands and needs edges specified, no CI edit. |
| REV6-04 | 2 | ACK tests repeat one transition while retaining different namespace assertions. | reviewer-data gpt-6.1-sol/high | Optional P3 only; publication-control is the recipient for the repeated block, namespace/native proof stays. No measured saving. |
| REV6-05 | 5 | Ordinary PR CI omits the three Unix Python check targets. | reviewer-architecture gpt-6.1-sol/high | Accepted P1 coverage-gap follow-up for a separate CI owner; add gated Unix checks, no installer bug claimed. |
| REV6-06 | 4 | Parallel Make can race dependency and embedded-asset preparation. | reviewer-architecture gpt-6.1-sol/high | Accepted debt: route separate FIX before enabling parallel Make; owner unassigned. Serial make remains the gate. |
| REV6-07 | 3 | Outer race tests do not instrument separately built application children. | reviewer-behavior gpt-6.1-sol/high; reviewer-behavior gpt-6-luna/medium; reviewer-data gpt-6.1-sol/high; reviewer-architecture gpt-6.1-sol/high | Limitation disclosed; explicit native-race cohort requires its own measured follow-up, no race defect asserted. |
| REV6-08 | 4 | HTTP readiness can block past the outer advertised deadline. | reviewer-tests gpt-6.1-sol/high | Existing AL-74 accepted debt revalidated; AL-74 owns bounded requests and nonresponding-listener proof. No duplicate ticket or fix. |
| REV6-09 | 2 | Diagnostic-history timing combines quadratic fixture width and real index backfill. | orchestrator gpt-6.1-sol/high | Separate performance follow-up should time setup/backfill/read and preserve index-seek/bounded-head proof; no savings or production complexity claimed. |

Orchestrator gpt-6.1-sol/high adjudicated source counterexamples and measured values. Test-effectiveness reviewer gpt-6.1-sol/high additionally compared frontend file/name/status multisets and duplicate multiplicity: the same 158 cases passed in both environments and were contained in the full 824 cases. Architecture reviewer gpt-6.1-sol/high identified the workflow setup-input/module-requirement distinction; orchestrator gpt-6.1-sol/high corrected the report's wording from source (module Go 1.26.0, setup input 1.23, actual hosted compiler unobserved). This minor factual documentation correction did not change acceptance, selectors or source and required no extra delegated review.

## Verification and limits

- Parent timing owner, orchestrator gpt-6.1-sol/high: 27 retained profiling/preparation commands passed, plus npm ci. Full capture/CLI cohorts with/without race reported no runtime skips; selected hotspots and pure gates were repeated. A separate CPU-profile cohort supports backfill/SQLite and instrumentation attribution, not an exact phase wall-time decomposition.
- Required integrated make at 48b1b00e: PASS, 228.21 s wall, 435.71 s user, 118.65 s system, 632128 KiB max RSS. Vet, TypeScript typecheck/lint, Python release/installation/distribution (14/10/15), full backend race tests, all 824 frontend cases and multi-platform build ran. Serial Make; source packages were not reported cached in output. Existing non-failing React act, Vite chunk and npm policy notices were not suppressed.
- Local WSL2/Go 1.26.0/Node 24 evidence on a shared host with observed contention. No cold-cache, hosted workflow, Windows-native or actual external OpenCode compatibility execution. Cross-compilation is not Windows runtime proof. Cgroup limits and aggregate concurrent process-tree memory were not established; no CI billing saving is claimed. Proposed budgets are local investigation alerts, not new test timeouts.
- Parent verified source/test/runtime/CI diff empty, JSON parsed, artifact links/whitespace and compact finding/reporting-tuple preservation. No tests were removed or modified. Pure gates are explicit initial subsets; full later suites retain other tests. Static Make graph and readiness counterexamples were not dynamically reproduced.
- Detailed reports/run/finding records were private working evidence; after compact retention and final approval they are removed. Retained compact records include accepted debt/owners and verification distinctions; no historical ticket audit was cleaned.

## Delivery and telemetry

The existing specification is completed and moved to plan/done; the matching index entry is removed. No downstream dependency on REV-6 required promotion. Final source-backed telemetry follows all review/verification work.

The installed agentic-lens wrapper was absent on PATH. The current source-native CLI was built and its help/version inspected; wrapper-only --al-offline/--al-project flags are not native flags. Final extraction uses the built native from the repository working directory and explicit --source codex; no global install, pin or configuration was changed. Exact session ingestion supplements root discovery for T3 review sessions whose source metadata records independent vscode roots, preserving that missing-parent limitation rather than inventing edges. The native planner has its recorded Codex parent edge. Expected source owners: root, planner and ten reviews. Existing other-harness sessions: none at baseline. Telemetry verification/facts are recorded at extraction below.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 28
- **Active Agent Duration:** 1h 4m (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 278.3k tokens (62% of 450k healthy ceiling, 0 compactions) | Subagents peak 69.7k (Planner)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 38,120 reasoning tokens (23.1% of generated output)
- **Tool Health & Self-Correction:** 259 tool calls (20 self-corrections, 92.3% success rate)

### Anomalies & Health Flags
- **Status:** 1 execution anomaly detected (1 WARNING).
- [WARNING] **CONTEXT_SATURATION_SPIKE**: Context ceiling breached (251.4k tokens) — Context tokens reached 251.4k, breaching the 250k warning ceiling.

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 66,593 | 4,579 | 473,088 | 87.7% | 69.7k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 11 | 1,256,151 | 121,984 | 30,703,104 | 96.1% | 278.3k | ELEVATED | Unavailable (0 observed, 0 partial, 11 unavailable / 11 invocations) |
| **Total** | **12** | **1,322,744** | **126,563** | **31,176,192** | **95.9%** | **278.3k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 5 | 325,605 | 19,657 | 1,901,184 | 85.4% | Unavailable (0 observed, 0 partial, 5 unavailable / 5 invocations) |
| **gpt-6.1-sol** | 7 | 997,139 | 106,906 | 29,275,008 | 96.7% | Unavailable (0 observed, 0 partial, 7 unavailable / 7 invocations) |
| **Total** | **12** | **1,322,744** | **126,563** | **31,176,192** | **95.9%** | **Unavailable (0 observed, 0 partial, 12 unavailable / 12 invocations)** |

Capture observation finalized: 2026-11-05T00:50:22.023467189Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 12; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 12; reserved bytes: 11960950. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1a7f0-ae81-7e60-90f7-f89c6dde0447: current; last successful source read 2026-11-05T00:50:21.73629618Z; accepted source checkpoint 2026-11-04T19:03:50.831Z.
Owner codex/01a1a7fd-a6d1-7772-abf8-4c9a4adc4863: current; last successful source read 2026-11-05T00:50:21.75369396Z; accepted source checkpoint 2026-11-04T19:18:10.103Z.
Owner codex/01a1a7db-7264-7d71-869d-64e1c5906a2f: current; last successful source read 2026-11-05T00:50:21.854697658Z; accepted source checkpoint 2026-11-05T00:50:21.635Z.
Owner codex/01a1a7fc-317e-7fb1-8d67-ef6b03325642: current; last successful source read 2026-11-05T00:50:21.878715347Z; accepted source checkpoint 2026-11-04T19:16:18.157Z.
Owner codex/01a1a7db-ce2f-7391-b1f2-a646b932ce13: current; last successful source read 2026-11-05T00:50:21.902585564Z; accepted source checkpoint 2026-11-04T18:41:31.704Z.
Owner codex/01a1a7ed-c79d-7e42-9c23-015dc99ca004: current; last successful source read 2026-11-05T00:50:21.92806974Z; accepted source checkpoint 2026-11-04T19:02:07.549Z.
Owner codex/01a1a7f7-7427-70e2-8900-c647a0f73bc8: current; last successful source read 2026-11-05T00:50:21.94451763Z; accepted source checkpoint 2026-11-04T19:11:22.593Z.
Owner codex/01a1a7f3-c7d5-77f2-adf0-9a6a3c12ffcd: current; last successful source read 2026-11-05T00:50:21.644804417Z; accepted source checkpoint 2026-11-04T19:09:29.222Z.
Owner codex/01a1a800-3dcc-78e3-94d6-41511e1016aa: current; last successful source read 2026-11-05T00:50:21.660688487Z; accepted source checkpoint 2026-11-04T19:20:26.566Z.
Owner codex/01a1a7f9-fd32-7d00-86b6-6cee5278a7ae: current; last successful source read 2026-11-05T00:50:21.68009216Z; accepted source checkpoint 2026-11-04T19:14:35.214Z.
Owner codex/01a1a7eb-39fb-7460-94e1-fe524cb8a72e: current; last successful source read 2026-11-05T00:50:21.701327182Z; accepted source checkpoint 2026-11-04T18:58:54.434Z.
Owner codex/01a1a7f2-ac77-7a90-bf87-2e534216eba8: current; last successful source read 2026-11-05T00:50:21.718927053Z; accepted source checkpoint 2026-11-04T19:05:31.355Z.

## Final telemetry verification

- Source-native CLI: bin/agentic-lens-linux-amd64, built commit 48b1b00e. Explicit Codex source at normal ~/.codex; root nonce 8805979c8fb821aa was supplied literally together with exact root session ID 01a1a7db-7264-7d71-869d-64e1c5906a2f. Native root extraction discovered root plus planner 01a1a7db-ce2f-7391-b1f2-a646b932ce13. Ten supplementary exact-session imports used their source-observed IDs; final ledger set equality against all expected source sessions passed: 12 sessions, all Codex, no duplicate IDs. Existing supplementary sessions survived final root refresh; the planner's recorded parent edge is preserved. No other-harness REV-6 sessions existed to preserve.
- Initial extraction failed because configured capture coordinates lacked the required worktree mapping table. No successful/empty ledger was fabricated. Supported explicit manual --capture-dir recovery used private disposable state under /tmp; configured state, installation and project configuration were preserved. The source warning remains disclosed here. All ten retries and final root extraction exited 0, with actual source/session verification beyond exit status.
- T3 reviews are independent vscode roots in Codex metadata: no native parent links were invented. Eleven invocation roles remain Unknown (root and ten T3 sessions); planner is recorded as planner. Thus source-session inclusion is verified, while T3 hierarchy/role attribution remains incomplete. Capture footer counts describe its selected owner collection, not a proof of complete T3 parentage.
- Snapshot aggregate active time: 1571 s (26 m 11 s); summed actor activity, not elapsed wall time. Cache-hit metric: 95.5%. All 12 cost-evidence states are unavailable; zero stored subtotals are not an observed zero-dollar cost. Peak context 251871 tokens triggered the configured warning threshold; rendered health is ELEVATED, with two context warnings. This reports source metrics, not an independently known harness context capacity.
- Extraction cutoff is the ledger updated_at value 2026-11-04T19:29:50.484012+00:00; final verification/commit/push calls necessarily follow that snapshot. Raw source rollouts, measurement logs, private capture state and transcripts are outside Git. No raw transcript was committed.


## Main integration, 2026-11-05

User requested merging main into the existing REV-6 branch. Initial working tree was clean. Integrated origin/main 40b16def, including delivered AL-66, AL-67 and AL-77. The only conflict was plan/TICKETS.md: retain upstream lifecycle changes and keep completed REV-6 and AL-77 out of the open index. No application conflict or implementation change was authored. The original measured review remains historical evidence for its explicitly recorded baseline; these newer upstream application changes have not been reprofiled or re-audited under REV-6. Fresh integrated make at 3beb779e FAILED in the incoming adapter tests: TestCurrentLedgerGenerationPolicyReadBoundary (historical_absent/all_enabled/all_disabled) and TestCurrentLedgerUnavailableAnomalyBoundary (omitted_unavailable/unmarked_measured_empty), all with "ticket_id is required and cannot be blank". schemas_test.go:1610 and :1656 construct self-contained current-schema payloads without ticket_id; they do not read REV-6 artifacts. The application/test/configuration diff from origin/main is empty. This is an incoming-main failure, not a resolved green gate; no AL-66 test implementation was changed under REV-6. Vet, frontend typecheck/lint, Python release/installation/distribution checks and other backend packages passed; the later frontend-test and multi-platform-build stages were not reached. Earlier review approval and timings remain scoped to their historical revision. The user-requested merge is delivered with this integration limit explicitly disclosed. Raw make output is private at /tmp/rev6-main-merge-make.log.

Merge-handoff telemetry refreshed from the source-native CLI built from 3beb779e, with literal nonce 03872e6210eb5fa7 and explicit Codex root. Verified all 12 prior source-session IDs survived with no additions/deletions; T3 parent/role and unavailable-cost limitations remain. The existing isolated manual capture state was used, preserving the incompatible configured state. This refreshed telemetry does not turn the failed integrated make into approval.


## Iteration 02 steering: test repair and staged execution

Source: user request on 2026-11-05, "fix the test, enhance the testing suite, split them into unit and integration tests. integration test should only run after units tests run successfully". This materially expands acceptance from the completed review to implementation within the same active REV-6 ticket. Initial status for this iteration was clean at 23b38163. The original CONTRACT.md and iteration01 review/measurements remain frozen historical evidence; iterations/02/CONTRACT_DELTA.md and PLAN.md define the proposed extension. The review-only implementation prohibition is superseded as product intent by this explicit request; the development skill still requires approval of the concrete new plan before coder dispatch. No implementation has started.

Preserved guarantees: valid required ledger identity, meaningful negative metadata/anomaly oracles, every existing test/assertion and native/storage/platform boundary, no shortened timeout or silent skip, no unrelated AL-74 remediation. Verification impact: repair the reproduced schema fixtures; prove unit/integration partition completeness and that failed applicable unit gates prevent integration execution/setup locally and in the configured CI graph, including parallel Make. Repository Make/CI gates include both Go/frontend units; standalone npm gates remain frontend-scoped and portable. Telemetry is deferred until the iteration's final delivery.

Iteration02 concrete contract/plan approval: user selected "Approve implementation plan" on 2026-11-05. Coder and tester started with disjoint source ownership at baseline 23b38163. Approval includes repository-global Make/CI gates and frontend-scoped standalone npm gates. Telemetry remains deferred until final delivery.


Iteration02 implementation checkpoint: coder owns the Make/npm/CI/compatibility graph and lane configs, tester owns fixture/oracle and orchestration tests. Existing application runtime logic and lockfiles remain untouched. Targeted repaired schema tables pass; isolated policy-arity, anomaly-marker and optional-array mutations were detected. A removed-Make-gate mutation permits integration despite failing units and is detected by the new harness. Initial actual-Make harness passed six meaningful tests with both-language failures, parallel mixed goals, positive controls, standalone npm scope, build-only override, fixture package partition and CI edge checks. Final source/evidence freeze and consolidated review follow; none of these targeted results alone claims integrated completion.
