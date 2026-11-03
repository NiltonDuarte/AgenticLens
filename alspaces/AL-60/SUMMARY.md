# AL-60 Work Summary

User pre-approved the specification on 2026-11-03. Initial working tree clean at 122e983f8c365a2aa248b1b6d492252555e922ae. Work is isolated on feat/AL-60-search-tickets. MGT-2 owns merging; this thread publishes the reviewed head only.

## Delivered behavior

Presentation: visibly and accessibly labeled Search tickets input narrows exclusion choices by case-insensitive ID/title matching. Query state is independent of exclusions and chart derivation. Hidden selections survive searching and clearing; explicit Clear search and no-match copy retain recovery. Native keyboard controls, point activation, chart defaults, axes, scales and tooltips remain available.

Domain and adapters: existing identity/title contracts reused unchanged; no API, extraction or metric additions. AL-55 telemetry preserved. Generated frontend assets remain ignored under the existing tracking policy.

## Review

Reviewed source: 7a50a5e99927928995a72e13128d597c1317e9d3, source diff SHA256 8703b8f22f375420ebd76f8e7a62f22e2a8b6902563133ea837fac91c4294914. Origin/main merge was already up to date; no integration code delta. Later changes concern delivery artifacts only.

Behavior/integration, data/failure handling, architecture/maintainability and test-effectiveness reviewers each ran independently under gpt-6.1-sol/high and gpt-6-luna/medium. Settings are dispatcher-acknowledged. No defect candidates or confirmed issues were reported. Seven scoped initial verdicts were APPROVED; architecture gpt-6-luna/medium was initially INCOMPLETE solely because build/asset evidence had not been supplied. Parent resolved that exact gap with successful canonical build, byte-equal generated/embed trees and isolated standalone HTTP asset verification. No discrepancy or product correction required another discovery pass.

The planner authored the contract/plan; coder implemented the control/styles; tester extended coherent interaction cases. These verification actors ran under inherited gpt-6.1-sol/high. Parent owns integration, independent evidence reconciliation, final approval and publication.

Final consolidated verdict: INCOMPLETE. Search source review has no findings; required integrated make remains failed in unchanged shared capture code. No publication/merge approval is asserted. Detailed review evidence remains open pending the gate; do not compact it as approved.

## Verification

- Tester: TicketScatterPlot focused suite 6/6 passed; typecheck/lint and diff check passed.
- Tester initial frontend full suite: 784 passed and two existing 5-second timeouts in TicketComparisons and TicketDeepDive. Both affected files passed serially, 83/83. No timeout or test configuration changes.
- Parent initial make: static/frontend/script checks passed; unchanged capture suites failed publication/version timeouts and manifest identity checks under parallel load. Bounded full make also failed (exit 2), with two-CPU affinity, GOMAXPROCS=2 and GOFLAGS=-p=1: TestCaptureLaterAgentChannelReportsBoundPublicationFailure (native lifetime exhausted; signal killed) and TestCaptureAutomaticSharedStoreAndTicketStayInExactProjects (worker cleanup deadline). Capture adapter package passed on this rerun; all remaining packages passed/cached. Make stopped before frontend tests/build, which were exercised separately.
- Parent canonical make build passed, including frontend build/sync and all configured platform binaries. Compared every web/dist asset byte-for-byte with backend/internal/adapters/web/dist. Started the standalone host binary against a disposable workspace, fetched index and its JS over HTTP, and confirmed both search and empty-state strings; server terminated afterward. No Node runtime used to serve the app.

Limits: real-browser viewport geometry and hosted CI have not been exercised. Existing Vite chunk-size and React act warnings are not AL-60 changes. Hosted CI not run; no PR published while delivery gate is incomplete.

## Telemetry

Checkpoint ingestion verified all 13 exact Codex sessions: primary root 01a1a27e-f9db-75b3-8e95-f6b1bbbe7f05, four structurally linked native children (planner, coder, tester, read-only explorer), and eight explicitly selected T3 review roots. Every refresh preserved prior session IDs; all source labels are codex. No prior other-harness ledger existed for AL-60. Built CLI help confirms generate --source/--session-id/--nonce; legacy --al-offline and --al-project flags are unsupported in the current native CLI. Commands run from the repository root using the built binary. T3 review sessions have source=vscode and no parent linkage; Their exact source IDs were ingested separately, preserving the recorded hierarchy. Review roots have no parent/role evidence in source, so aggregate root/role counts reflect that harness limitation. Do not invent parent links or reviewer roles in source facts.

## Blocked checkpoint evidence

Full serial frontend suite independently passed 786/786 across 36 files (113.94s), using npm test -- --maxWorkers=1 --minWorkers=1. The two failing bounded CLI tests passed in a targeted rerun without CPU affinity: GOMAXPROCS=2 go test -race -p=1 ./cmd/agentic-lens -run 'TestCaptureLaterAgentChannelReportsBoundPublicationFailure|TestCaptureAutomaticSharedStoreAndTicketStayInExactProjects' -count=1 (8.291s). This supports resource sensitivity, not a passing full gate.

User reports AL-61 failures on unchanged main 122e983 in LaterAgentChannelReportsBoundPublicationFailure, CLIOpenCodePluginOwnershipAndMissingHandler and NativeWorkerPublishesMultipleRootsAndRefreshesWithoutGenerate; manager baseline corroborates native version/capability and worker timeout failures. These are reported external observations, not locally rerun baseline results. Capture repair expansion is explicitly unapproved. Read-only diagnosis and minimal boundary: [CAPTURE_DIAGNOSIS.md](iterations/01/CAPTURE_DIAGNOSIS.md).

### Workflow Efficiency
- **Software Development Life Cycle Iterations:** 22
- **Active Agent Duration:** 15m 30s (excludes user wait time)
- **Context Window Health:** ELEVATED — Root agent peak 126.8k tokens (63% of 200k healthy ceiling, 0 compactions) | Subagents peak 139.9k (Explorer)
- **Role Effort Distribution (Subagents):** Unavailable: Complete nonzero subagent cost is required
- **Subagent Turn Velocity & Thrashing:** 1.0 avg turns/run (peak: 1 turns) | Planner 1.0, Coder 1.0, Tester 1.0, Explorer 1.0
- **Test-to-Production Code Ratio:** 4.8:1 (+77 test lines / +16 prod lines)
- **Code Delivered:** +93 / -1 lines across 3 code files (excluding docs)
- **Reasoning Effort:** 9,670 reasoning tokens (15.5% of generated output)
- **Tool Health & Self-Correction:** 203 tool calls (16 self-corrections, 92.1% success rate)

### Anomalies & Health Flags
- **Status:** No execution anomalies detected (clean session).

### Agentic Execution Metrics

**Contributors:** Nilton Duarte

#### By Subagent Role
| Role | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Peak Context | Health | Cost (USD) |
|---|---|---|---|---|---|---|---|---|
| **planner** | 1 | 50,552 | 2,423 | 270,080 | 84.2% | 52.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **coder** | 1 | 105,174 | 1,265 | 341,760 | 76.5% | 65.8k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **tester** | 1 | 55,548 | 3,343 | 1,513,472 | 96.5% | 80.0k | HEALTHY | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Unknown** | 9 | 603,103 | 42,980 | 11,389,952 | 95.0% | 126.8k | ELEVATED | Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations) |
| **explorer** | 1 | 115,406 | 2,546 | 1,165,568 | 91.0% | 139.9k | ELEVATED | Unavailable (0 observed, 0 partial, 1 unavailable / 1 invocations) |
| **Total** | **13** | **929,783** | **52,557** | **14,680,832** | **94.0%** | **139.9k** | **ELEVATED** | **Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations)** |

#### By Model
| Model | Invocations | Input Tokens | Output Tokens | Cache Read | Cache Hit % | Cost (USD) |
|---|---|---|---|---|---|---|
| **gpt-6-luna** | 4 | 158,984 | 10,472 | 998,272 | 86.3% | Unavailable (0 observed, 0 partial, 4 unavailable / 4 invocations) |
| **gpt-6.1-sol** | 9 | 770,799 | 42,085 | 13,682,560 | 94.7% | Unavailable (0 observed, 0 partial, 9 unavailable / 9 invocations) |
| **Total** | **13** | **929,783** | **52,557** | **14,680,832** | **94.0%** | **Unavailable (0 observed, 0 partial, 13 unavailable / 13 invocations)** |

Capture observation finalized: 2026-11-03T18:03:58.021271402Z (successful owner reads have independent checkpoints; no common as-of snapshot). Refreshed: 13; retained older facts: 0; pending: 0; failed: 0; unassigned within this collection: 0. Unresolved child evidence: 0; parent links: 0; discovery continuation: false. Source attempts: 13; reserved bytes: 2885147. Source failures preserve accepted facts; pending processing does not establish freshness. Only observed, bound owners are included; unobserved sessions and harness trust remain unverified. Inspect local capture diagnostics and retry capture reconcile for unavailable sources or unfinished child recovery.
Owner codex/01a1a285-a7e2-7083-aad5-781dfee72767: current; last successful source read 2026-11-03T18:03:57.883975073Z; accepted source checkpoint 2026-11-03T17:47:46.391Z.
Owner codex/01a1a280-9d93-7471-9927-ea58a45f6e99: current; last successful source read 2026-11-03T18:03:57.890639719Z; accepted source checkpoint 2026-11-03T17:42:54.585Z.
Owner codex/01a1a285-f25d-79f1-bef1-db5606c7ef63: current; last successful source read 2026-11-03T18:03:57.897207161Z; accepted source checkpoint 2026-11-03T17:48:17.873Z.
Owner codex/01a1a27f-2156-7d82-951e-0831e5c48114: current; last successful source read 2026-11-03T18:03:57.903568027Z; accepted source checkpoint 2026-11-03T17:41:34.612Z.
Owner codex/01a1a287-55e9-73b0-91a6-30c361793268: current; last successful source read 2026-11-03T18:03:57.909564704Z; accepted source checkpoint 2026-11-03T17:49:46.235Z.
Owner codex/01a1a285-48ae-7673-8991-f4f6890125ff: current; last successful source read 2026-11-03T18:03:57.916401928Z; accepted source checkpoint 2026-11-03T17:47:56.345Z.
Owner codex/01a1a284-6527-77b0-8445-85af9362f511: current; last successful source read 2026-11-03T18:03:57.92488152Z; accepted source checkpoint 2026-11-03T17:47:07.76Z.
Owner codex/01a1a292-1334-7753-b1c2-a03b6db06572: current; last successful source read 2026-11-03T18:03:57.933154325Z; accepted source checkpoint 2026-11-03T18:02:25.425Z.
Owner codex/01a1a286-a6ba-7ea0-9f57-9998f53aea35: current; last successful source read 2026-11-03T18:03:57.940820664Z; accepted source checkpoint 2026-11-03T17:49:52.526Z.
Owner codex/01a1a281-c2de-7ec3-9541-aa8d191c94b0: current; last successful source read 2026-11-03T18:03:57.947384817Z; accepted source checkpoint 2026-11-03T17:46:02.873Z.
Owner codex/01a1a287-ec4f-7e73-997b-f887745b5d9e: current; last successful source read 2026-11-03T18:03:57.962027929Z; accepted source checkpoint 2026-11-03T17:50:22.308Z.
Owner codex/01a1a283-3ddc-7f31-8a32-01e8bbacbf87: current; last successful source read 2026-11-03T18:03:57.968975279Z; accepted source checkpoint 2026-11-03T17:45:45.018Z.
Owner codex/01a1a27e-f9db-75b3-8e95-f6b1bbbe7f05: current; last successful source read 2026-11-03T18:03:58.012981226Z; accepted source checkpoint 2026-11-03T18:03:57.584Z.
