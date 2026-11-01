# Decomposed Jira Tickets for AgenticLens

Individual ticket definitions are located in `alspaces/plan/done/`, `alspaces/plan/to-do/`, and `alspaces/plan/discarded/`. Completed tickets are removed from this index; their specifications remain in `done/`.

---

## In Progress




---

## Ready

### Post-v2.0 review: Ready

- **AL-60**: Add Search tickets to the comparison exclusion panel
  - **Subsystem**: Dashboard comparisons
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL-60 search tickets in comparison exclusion panel.md`

- **AL-61**: Show full local historical timestamps in Ticket Deep Dive
  - **Subsystem**: Dashboard historical timestamps
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL-61 show full local timestamps in ticket deep dive.md`

- **AL-62**: Fix Codex skill and delegation facts with granular source ownership
  - **Subsystem**: Domain / Codex and OpenCode Adapters / Schema Migration / Backfill / CLI / API / Dashboard / Tests
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: AL-63
  - **File**: `alspaces/plan/to-do/AL-62 fix codex skill and delegation facts with granular ownership.md`

- **AL-64**: Investigate and fix active-turn duration across Codex and OpenCode v1/v2
  - **Subsystem**: Domain / Codex / OpenCode v1 and v2 / Generation / Backfill / Ledger / Presentation
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL-64 investigate and fix active turn duration across codex and opencode.md`

- **AL-66**: Configure optional skills, task_links and anomalies output for generate
  - **Subsystem**: Configuration / Domain / Source Adapters / CLI / Persistence / Summary / Dashboard Compatibility / Tests
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL-66 configure optional metric output for generate.md`

- **AL-67**: Review shared telemetry repository interfaces and source injection
  - **Subsystem**: Architecture / Domain Ports / Source Adapters / Composition / Backfill / Hydration / HTTP and Live Services / Tests
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL-67 review telemetry repository interfaces and dependency injection.md`

- **AL-68**: Install harness lifecycle hooks for automatic session capture
  - **Subsystem**: Installation / Setup / Domain / Codex and OpenCode Adapters / CLI / Persistence / Tests
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL-68 install harness lifecycle hooks for automatic session capture.md`

### Existing ready work

- **AL2-1**: End-to-End Agent Anomaly Debug Prompt Generator, CLI, and Interactive Web Drawer
  - **Subsystem**: Domain / Adapters / HTTP / CLI / Presentation / Web / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL2-1 end-to-end agent anomaly debug prompt generator cli and interactive web drawer.md`

- **AL2-2**: End-to-End Workspace Archive Engine, CLI Compaction, and In-Memory Decompression Ingestion
  - **Subsystem**: Domain / Adapters / CLI / HTTP / Presentation / Web / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL2-2 end-to-end workspace archive engine cli compaction and in-memory decompression ingestion.md`

- **AL2-3**: Multi-Project Configuration Schema, Domain Aggregations, and Multi-Folder Scanner
  - **Subsystem**: Domain / Adapters / CLI / Config / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: AL2-4, AL2-5
  - **File**: `alspaces/plan/to-do/AL2-3 multi project configuration schema domain aggregations and multi folder scanner.md`

- **AL2-6**: Model API Pricing Configuration, Domain Rate-Card Cost Approximation Engine, and Hybrid Ledger Ingestion
  - **Subsystem**: Config / Domain / Adapters / CLI / HTTP / Presentation / Tests
  - **Phase**: Phase 8 (v2 Cost Modeling & Parity)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: AL2-7
  - **File**: `alspaces/plan/to-do/AL2-6 model api pricing configuration and rate card cost approximation engine.md`

- **AL2-8**: Ingest Model Reasoning Effort Variants and Surface Deliberation Settings across Platform Views
  - **Subsystem**: Domain / Adapters / OpenCode / HTTP / Presentation / Web / Tests
  - **Phase**: Phase 8 (v2 Observability & Models)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL2-8 reasoning effort telemetry integration across fleet kpis comparison tables and progression.md`

- **AL2-9**: Evaluate Cryptographic Signing and Verification for Telemetry Ledgers
  - **Subsystem**: Product / Security / CLI / Documentation
  - **Phase**: Phase 8 (v2 Security and Provenance Discovery)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL2-9 evaluate cryptographic signing and verification for telemetry ledgers.md`

---

## Blocked

### Post-v2.0 review: Dependencies

- **AL-63**: Audit source capability parity across OpenCode v1, OpenCode v2, and Codex
  - **Subsystem**: Multi-source ingestion / hydration / recovery
  - **Status**: Blocked
  - **Blocked by**: AL-62
  - **Blocks**: AL-65
  - **File**: `alspaces/plan/to-do/AL-63 audit capabilities across opencode v1 v2 and codex.md`

- **AL-65**: Research default source-backed recomputation and the future of Force
  - **Subsystem**: Backfill CLI semantics / source recovery
  - **Status**: Blocked
  - **Blocked by**: AL-63
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL-65 research safe default source backed backfill.md`

### Existing blocked work

- **AL2-4**: Remote Git Repository Ingestion, Shallow Cache Manager, and Project Management API
  - **Subsystem**: Adapters / Git / HTTP / Server / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Pending
  - **Blocked by**: AL2-3
  - **Blocks**: AL2-5
  - **File**: `alspaces/plan/to-do/AL2-4 remote git repository ingestion shallow cache manager and project management api.md`

- **AL2-5**: Multi-Project Web Dashboard Selector, Cross-Project Fleet Overview, and Comparative Analytics
  - **Subsystem**: Presentation / Web / Components / State / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Pending
  - **Blocked by**: AL2-3, AL2-4
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL2-5 multi project web dashboard selector cross project fleet overview and comparative analytics.md`

- **AL2-7**: Interactive Cost Estimation Dashboards, Hybrid Fleet Analytics, and Rate-Card Simulation
  - **Subsystem**: Presentation / Web / Components / State / Analytics / Tests
  - **Phase**: Phase 8 (v2 Cost Modeling & Parity)
  - **Status**: Pending
  - **Blocked by**: AL2-6
  - **Blocks**: None
  - **File**: `alspaces/plan/to-do/AL2-7 interactive cost estimation dashboards hybrid fleet analytics and rate card simulation.md`

---

## Discarded

- **AL-18**: Implement Live Trace Inspector, Reasoning Stream, and Prominent Tool Failure Telemetry
  - **Subsystem**: Domain / Adapters / HTTP / Tests
  - **Status**: Discarded (Merged into AL-17)
  - **File**: `alspaces/plan/discarded/AL-18 implement live trace inspector reasoning stream and prominent tool failure telemetry.md`

- **AL-10**: Fix Git Metrics Calculation in Python Extractor and Backfill Ledgers
  - **Subsystem**: Adapters / Python / Tests
  - **Status**: Discarded (Superseded by AL-11)
  - **File**: `alspaces/plan/discarded/AL-10 fix git metrics calculation in python extractor and backfill ledgers.md`
