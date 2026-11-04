# AL-67 work summary

Documentation-only architecture assessment and recommendations. No application, test, schema, manifest or runtime-script changes were made. [Assessment](ARCHITECTURE_ASSESSMENT.md), [contract](CONTRACT.md), [approved iteration plan](iterations/01/PLAN.md), [compact review](iterations/01/REVIEW_SUMMARY.json).

## Baseline and approval

Initial status was clean at `a84adbe5be46c2ef52ede0987a417562715d38e1`, on `AL-67/review-telemetry-repository-interfaces-and-dependency-injection`. Work uses `docs/AL-67-telemetry-repository-review` and the configured `alspaces/AL-67` workspace. Planner prepared the contract/plan; user explicitly approved the documentation-only audit on 2026-11-04. There was no implementation scope expansion, coder dispatch or new iteration. Unrelated history, including AL-55, remains untouched.

## Delivered scope by layer

- Domain: inventory of normalized facts, coverage and pure aggregation guarantees; proposed interfaces remain in imperative consumers, without domain I/O.
- Adapters/composition: actual ledger/source/service contracts, constructors and dependency maps, OpenCode v1/v2/Codex operation matrix, generate/backfill/capture/server/hydration/live selection and lifecycle traces, source-grounded findings and staged target design.
- Presentation: historical/live API and UI trace showing persisted baseline visibility, coverage validation and obsolete-selection guards. No UI change.

There is no common native telemetry-source repository today. The target proposes small exact-owner/event capabilities, separate discovery and live observation, independent ledger persistence, and explicit composition. Historical v2 hydration is distinct from v1-only live SQL. All recommended corrections remain follow-up debt, with proposed responsibility and verification in the assessment; no downstream ticket was created or newly blocked.

## Consolidated review

Application revision reviewed by every focus: `a84adbe5be46c2ef52ede0987a417562715d38e1`. Candidate documentation commit: `436fb095ab06db423f2c43cad0821a98a9952160`, plus final owned documentation. Fetch/merge of origin/main was already up to date; the application tree is unchanged. Final consolidated verdict: APPROVED after the required integrated verification. Approval concerns assessment accuracy and completed evidence, not remediation of existing defects or merge authorization.

Four responsibilities each ran independently under both configured profiles: behavior/integration, data/failure handling, architecture/maintainability and test effectiveness; `gpt-6.1-sol` / high and `gpt-6-luna` / medium. Eight focus/profile cells completed in nine runs. Luna data initially returned INCOMPLETE; a new same-profile completion run closed its named v1/v2/Codex/HTTP/UI gaps and returned APPROVED. All other scoped passes returned APPROVED, including passes with no substantiated blocking application finding. Dispatch acknowledgments establish selected model/effort; native reviewer-session model identity and costs are not independently source-verified.

Discovering reviewers and dispositions below retain shared credit. The coordinator (`gpt-6.1-sol` / high) separately verified consequential source paths; coordinator validation does not gain discovery credit. Exact reporting tuples and numeric severity survive in REVIEW_SUMMARY.json.

| Finding | Discovering responsibilities and model/effort | Disposition |
| --- | --- | --- |
| F01 configured Codex home, 6 | Behavior, architecture: `gpt-6.1-sol` / high and `gpt-6-luna` / medium; tests: `gpt-6.1-sol` / high | Confirmed; CLI composition follow-up, no fix. |
| F02 relative paths, 5 | Behavior, architecture, tests: `gpt-6.1-sol` / high | Confirmed; composition follow-up preserving CLI-relative semantics. |
| F03 v2 live capability, 5 | Behavior, architecture: `gpt-6.1-sol` / high and `gpt-6-luna` / medium | Confirmed; live capability reporting follow-up; historical hydration retained. |
| F04 duplicate Codex snapshots, 6 | Data: `gpt-6.1-sol` / high; completion-phase co-discovery by data: `gpt-6-luna` / medium, disputing impact | Coordinator confirms grouping appends rows before unique-owner validation; Codex ambiguity follow-up. |
| F05 child lookup diagnostic, 4 | Data, architecture, tests: `gpt-6.1-sol` / high | Confirmed failure-state gap. Architecture stressed optional link enrichment; reconcile by preserving content plus explicit availability per checklist. |
| F06 cancellation, 4 | Architecture: `gpt-6.1-sol` / high and `gpt-6-luna` / medium | Context propagation debt; latency not measured. |
| F07 native coupling/dispatch, 3 | Architecture: `gpt-6.1-sol` / high and `gpt-6-luna` / medium | Proposed narrow consumer ports; mixed-layout compatibility changes need approval. |
| F08 snapshot boundary, 3; F09 tie ordering, 2 | Data: `gpt-6.1-sol` / high | Contract/determinism recommendations, no observed corruption. |
| F10 column-omission oracle, 3 | Tests: `gpt-6.1-sol` / high and `gpt-6-luna` / medium | Confirmed coverage limit; production column validation still exists. |
| F11 alternate-source fixture, 2 | Tests: `gpt-6.1-sol` / high | Confirmed coverage limit, no demonstrated fallback defect. |
| F12 oversized-interface caution, 1 | Architecture: `gpt-6.1-sol` / high and `gpt-6-luna` / medium | Design recommendation only. |

Test reviewers also verified meaningful positive fixture oracles: exact ownership, real repeated rows, mixed-source partitioning, unavailable children, persisted HTTP baseline and no GET writes. These strengths are preserved alongside composition/test gaps. Twelve canonical findings/recommendations remain explicitly unimplemented. Source counterexamples establish bounded consequences, not production incidents. Detailed process reports are compacted only after final approval; the architecture assessment remains the durable audit product.

## Verification

Verification actor: coordinator (`gpt-6.1-sol` / high); test reviewers independently inspected oracles and reused valid execution. Source inspected at a84adbe5; integrated gate at 436fb095 plus documentation-only clarification. Application and test trees unchanged, so focused results remain applicable.

- Frontend dependencies installed and canonical `make web-assets` passed. A host source binary built and generate/backfill/server help was checked.
- Race-enabled focused Go checks passed for config, OpenCode, generation, backfill source/multi-machine cases, hydration/live and HTTP persisted visibility. Exact command is retained in the assessment. Codex matched no tests in that filter and was not counted there.
- Separate `go test -race ./internal/adapters/codex -count=1` passed.
- Integrated `make`: passed: Go vet/race suites; TypeScript typecheck/lint; release, installation and capture-distribution checks; 36 frontend test files / 824 tests; Linux/Darwin amd64/arm64 and Windows amd64 builds. Frontend stages were not silently skipped. Build output may include a non-fatal bundle-size advisory.
- Whitespace, local artifact/source links and documentation-only scope checked. Embedded assets and standalone binary verification: three asset hashes match web/dist and the embedded tree; standalone binary served all three assets with identical bytes in a disposable fixture, then server/fixture were cleaned.

No new tests, dynamic cancellation/duplicate-file mutation probe, browser session or release execution. Native harness compatibility tests creating new source sessions were not run. Fixtures do not establish installed OpenCode/Codex versions. Hosted PR CI/review is independent of local checks.

## Telemetry source boundaries

Native metadata inventory proves Codex root `01a1a7da-055d-7eb2-ade9-d3db0bf494ca` and planner child `01a1a7da-d618-7583-8183-863cae857761`, with the native parent link. The nine T3 service review runs have acknowledged model/effort settings but their native parent linkage is not proven by available session_meta. Service workflow coverage is therefore broader than the provable native source graph. Do not invent child IDs, parent facts or usage to close that gap. Raw native transcripts stay outside Git.

The shared pinned release is not installed at its configured cache path. The final extraction uses the freshly built native CLI from this project directory, with explicit Codex source and docs type; wrapper-only --al-offline/--al-project flags cannot be passed to a direct native. Capture state is isolated via the supported explicit manual override, preserving existing shared operational state. Final extraction succeeded with literal nonce `6c90b8e2d571af34`, explicit exact root, `--source codex`, `--ticket-type docs`, and isolated `--capture-dir`. Independent metadata inventory matches both saved session IDs, Codex labels and parent relationships exactly. Schema is 6. No prior AL-67 ledger exists, so preservation of earlier harness owners is not applicable. The ledger and rendered metrics cover two native sessions; the nine service review runs remain excluded from native usage totals because linkage is unproven. Costs remain unavailable.

The CLI diagnosed unsupported configured capture coordinates (missing worktree mapping table); the supported explicit manual override used healthy isolated state and preserved the configured state. Capture-registry counters of zero describe that isolated registry, not an absence of manual source extraction. Manual discovery/read found two native rollouts. The rendered 13 SDLC iterations are the CLI-derived telemetry metric, not thirteen approved ticket iterations; this ticket has one approved iteration. Root persona is not source-proven and remains Unknown in the rendered role table. No application fix or synthetic fact was used to alter those source results.

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 13
- **Active Agent Duration:** 10m 54s (excludes user wait time)
- **Context Window Health:** HEALTHY — Root agent peak 172.9k tokens (38% of 450k healthy ceiling, 0 compactions) | Subagents peak 72.6k (Planner)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0
- **Test-to-Production Code Ratio:** 0:0 (+0 test lines / +0 prod lines)
- **Code Delivered:** +0 / -0 lines across 0 code files (excluding docs)
- **Reasoning Effort:** 6,614 reasoning tokens (13.9% of generated output)
- **Tool Health & Self-Correction:** 76 tool calls (3 self-corrections, 96.1% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 69,169 | 4,670 | 425,728 | 86.0% | 72.6k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 1 | 152,854 | 36,196 | 7,437,696 | 98.0% | 172.9k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **2** | **222,023** | **40,866** | **7,863,424** | **97.3%** | **172.9k** | **HEALTHY** | **Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6.1-sol** | 2 | 222,023 | 40,866 | 7,863,424 | 97.3% | Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations) |
| **Total** | **2** | **222,023** | **40,866** | **7,863,424** | **97.3%** | **Unavailable (0 observed, 0 partial, 2 unavailable / 2 invocations)** |

Capture observation finalized: 2026-11-04T19:13:44.993891359Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 0; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 0; reserved bytes: 0. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
