# AgenticLens Roadmap

Active specifications live in `alspaces/plan/to-do/`; completed specifications remain in `alspaces/plan/done/` and are removed from this index. The directory name is storage, not a To Do status. Historical discarded specifications remain outside the active roadmap.

Reconciled 2026-11-05 against merged `origin/main` at `03f5b05d` and open PRs. See [PO-19 evidence and ticket review](../PO-19/SPEC.md). Ready means no outstanding ticket prerequisite; it does not imply implementation approval or passing acceptance.

2026-11-08: PO-20 adds the user-level capture initiative as two vertical slices: AL-79 is ready and unblocks AL-80. Existing roadmap entries and dependencies are unchanged.

## Ready

- **AL-81**: Make manual generate independent of the capture registry
  - **Subsystem**: Domain / CLI / Source Adapters / Ledger Publication
  - **Priority**: P1
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-81 decouple manual generate from capture registry.md](to-do/AL-81%20decouple%20manual%20generate%20from%20capture%20registry.md)
  - **Specification**: [PO-20](../PO-20/SPEC.md)

- **AL-78**: Make open-turn feedback effective across Codex and OpenCode
  - **Subsystem**: CLI Feedback / Native Harness Lifecycle / Capture and Metrics Refresh
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-78 make open turn feedback effective across codex and opencode.md](to-do/AL-78%20make%20open%20turn%20feedback%20effective%20across%20codex%20and%20opencode.md)

- **AL-63**: Audit source capability parity across OpenCode v1, OpenCode v2, and Codex
  - **Subsystem**: Multi-source ingestion / hydration / recovery
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: AL-65
  - **File**: [AL-63 audit capabilities across opencode v1 v2 and codex.md](to-do/AL-63%20audit%20capabilities%20across%20opencode%20v1%20v2%20and%20codex.md)

- **AL-69**: Correct telemetry accuracy and preserve evidence
  - **Subsystem**: Domain / Source Adapters / CLI / Persistence / Anomaly Presentation
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-69 telemetry accuracy and preservation.md](to-do/AL-69%20telemetry%20accuracy%20and%20preservation.md)

- **AL-70**: Preserve configuration, paths and ticket identity
  - **Subsystem**: Configuration / CLI / Git and OpenCode Adapters / HTTP Listener / Frontend Identity
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-70 configuration paths and ticket identity.md](to-do/AL-70%20configuration%20paths%20and%20ticket%20identity.md)

- **AL-71**: Make frontend requests and live data handling reliable
  - **Subsystem**: Frontend Fleet Comparisons Progression / Live REST SSE Polling
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-71 frontend requests and live data handling.md](to-do/AL-71%20frontend%20requests%20and%20live%20data%20handling.md)

- **AL-72**: Correct UI interactions, calculations and accessibility
  - **Subsystem**: Frontend Trace Controls / Clipboard / Charts / Tables / Dialog and Tooltip
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-72 ui correctness and accessibility.md](to-do/AL-72%20ui%20correctness%20and%20accessibility.md)

- **AL-73**: Make installation and release verification reliable
  - **Subsystem**: Installer / Downloader / Packaged Setup / Release Packaging
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-73 installation and release reliability.md](to-do/AL-73%20installation%20and%20release%20reliability.md)

- **AL-74**: Strengthen test reliability and remove unused code
  - **Subsystem**: Domain Migration and Aggregation / Adapter and CLI Tests / Frontend Cleanup / Compatibility Fixtures
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-74 test reliability and unused code cleanup.md](to-do/AL-74%20test%20reliability%20and%20unused%20code%20cleanup.md)

- **AL-75**: Verify dependency integrity and reproducible builds
  - **Subsystem**: Dependency Maintenance / Lockfile / Docker / CI Toolchain
  - **Priority**: P3
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL-75 dependency and build reproducibility.md](to-do/AL-75%20dependency%20and%20build%20reproducibility.md)

- **AL2-1**: End-to-End Agent Anomaly Debug Prompt Generator, CLI, and Interactive Web Drawer
  - **Subsystem**: Domain / Adapters / HTTP / CLI / Presentation / Web / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL2-1 end-to-end agent anomaly debug prompt generator cli and interactive web drawer.md](to-do/AL2-1%20end-to-end%20agent%20anomaly%20debug%20prompt%20generator%20cli%20and%20interactive%20web%20drawer.md)

- **AL2-2**: End-to-End Workspace Archive Engine, CLI Compaction, and In-Memory Decompression Ingestion
  - **Subsystem**: Domain / Adapters / CLI / HTTP / Presentation / Web / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL2-2 end-to-end workspace archive engine cli compaction and in-memory decompression ingestion.md](to-do/AL2-2%20end-to-end%20workspace%20archive%20engine%20cli%20compaction%20and%20in-memory%20decompression%20ingestion.md)

- **AL2-3**: Multi-Project Configuration Schema, Domain Aggregations, and Multi-Folder Scanner
  - **Subsystem**: Domain / Adapters / CLI / Config / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: AL2-4, AL2-5
  - **File**: [AL2-3 multi project configuration schema domain aggregations and multi folder scanner.md](to-do/AL2-3%20multi%20project%20configuration%20schema%20domain%20aggregations%20and%20multi%20folder%20scanner.md)

- **AL2-6**: Model API Pricing Configuration, Domain Rate-Card Cost Approximation Engine, and Hybrid Ledger Ingestion
  - **Subsystem**: Config / Domain / Adapters / CLI / HTTP / Presentation / Tests
  - **Phase**: Phase 8 (v2 Cost Modeling & Parity)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: AL2-7
  - **File**: [AL2-6 model api pricing configuration and rate card cost approximation engine.md](to-do/AL2-6%20model%20api%20pricing%20configuration%20and%20rate%20card%20cost%20approximation%20engine.md)

- **AL2-8**: Ingest Model Reasoning Effort Variants and Surface Deliberation Settings across Platform Views
  - **Subsystem**: Domain / Adapters / OpenCode / HTTP / Presentation / Web / Tests
  - **Phase**: Phase 8 (v2 Observability & Models)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL2-8 reasoning effort telemetry integration across fleet kpis comparison tables and progression.md](to-do/AL2-8%20reasoning%20effort%20telemetry%20integration%20across%20fleet%20kpis%20comparison%20tables%20and%20progression.md)

- **AL2-9**: Evaluate Cryptographic Signing and Verification for Telemetry Ledgers
  - **Subsystem**: Product / Security / CLI / Documentation
  - **Phase**: Phase 8 (v2 Security and Provenance Discovery)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [AL2-9 evaluate cryptographic signing and verification for telemetry ledgers.md](to-do/AL2-9%20evaluate%20cryptographic%20signing%20and%20verification%20for%20telemetry%20ledgers.md)

- **FIX-18**: Report unsupported live telemetry schemas accurately
  - **Subsystem**: OpenCode / Live Service / HTTP SSE / Dashboard
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [FIX-18 report unsupported live telemetry schemas accurately.md](to-do/FIX-18%20report%20unsupported%20live%20telemetry%20schemas%20accurately.md)

- **FIX-19**: Reject ambiguous Codex owner snapshots
  - **Subsystem**: Codex / Generate / Backfill / Historical Hydration
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [FIX-19 reject ambiguous codex owner snapshots.md](to-do/FIX-19%20reject%20ambiguous%20codex%20owner%20snapshots.md)

- **FIX-20**: Cancel native source reads with their callers
  - **Subsystem**: Source Adapters / Live and Historical HTTP / Backfill
  - **Priority**: P2
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [FIX-20 cancel native source reads with their callers.md](to-do/FIX-20%20cancel%20native%20source%20reads%20with%20their%20callers.md)

- **REV-7**: Review shared source read guarantees
  - **Subsystem**: Source Contracts / Snapshot and Ordering / Composition
  - **Priority**: P3
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: [REV-7 review shared source read guarantees.md](to-do/REV-7%20review%20shared%20source%20read%20guarantees.md)

## In Progress

- **AL-79**: Deliver native user-level Codex capture across worktrees
  - **Subsystem**: Domain / Configuration / Native Capture / Codex / CLI / Setup
  - **Priority**: P1
  - **Status**: In Progress
  - **Blocked by**: None
  - **Blocks**: AL-80
  - **File**: [AL-79 native user level codex capture across worktrees.md](to-do/AL-79%20native%20user%20level%20codex%20capture%20across%20worktrees.md)
  - **Specification**: [PO-20](../PO-20/SPEC.md)

## Blocked

- **AL-80**: Deliver global OpenCode capture using the shared worktree runtime
  - **Subsystem**: OpenCode / TypeScript Plugin / Native Capture / CLI / Tests
  - **Priority**: P1
  - **Status**: Blocked
  - **Blocked by**: AL-79
  - **Blocks**: None
  - **File**: [AL-80 global opencode capture with shared worktree runtime.md](to-do/AL-80%20global%20opencode%20capture%20with%20shared%20worktree%20runtime.md)
  - **Specification**: [PO-20](../PO-20/SPEC.md)

- **AL-65**: Research default source-backed recomputation and the future of Force
  - **Subsystem**: Backfill CLI semantics / source recovery
  - **Status**: Blocked
  - **Blocked by**: AL-63
  - **Blocks**: None
  - **File**: [AL-65 research safe default source backed backfill.md](to-do/AL-65%20research%20safe%20default%20source%20backed%20backfill.md)

- **AL2-4**: Remote Git Repository Ingestion, Shallow Cache Manager, and Project Management API
  - **Subsystem**: Adapters / Git / HTTP / Server / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Blocked
  - **Blocked by**: AL2-3
  - **Blocks**: AL2-5
  - **File**: [AL2-4 remote git repository ingestion shallow cache manager and project management api.md](to-do/AL2-4%20remote%20git%20repository%20ingestion%20shallow%20cache%20manager%20and%20project%20management%20api.md)

- **AL2-5**: Multi-Project Web Dashboard Selector, Cross-Project Fleet Overview, and Comparative Analytics
  - **Subsystem**: Presentation / Web / Components / State / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Blocked
  - **Blocked by**: AL2-3, AL2-4
  - **Blocks**: None
  - **File**: [AL2-5 multi project web dashboard selector cross project fleet overview and comparative analytics.md](to-do/AL2-5%20multi%20project%20web%20dashboard%20selector%20cross%20project%20fleet%20overview%20and%20comparative%20analytics.md)

- **AL2-7**: Interactive Cost Estimation Dashboards, Hybrid Fleet Analytics, and Rate-Card Simulation
  - **Subsystem**: Presentation / Web / Components / State / Analytics / Tests
  - **Phase**: Phase 8 (v2 Cost Modeling & Parity)
  - **Status**: Blocked
  - **Blocked by**: AL2-6
  - **Blocks**: None
  - **File**: [AL2-7 interactive cost estimation dashboards hybrid fleet analytics and rate card simulation.md](to-do/AL2-7%20interactive%20cost%20estimation%20dashboards%20hybrid%20fleet%20analytics%20and%20rate%20card%20simulation.md)
