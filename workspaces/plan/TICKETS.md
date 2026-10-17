# Decomposed Jira Tickets for AgenticLens

Individual ticket definitions are located in `workspaces/plan/done/`, `workspaces/plan/to-do/`, and `workspaces/plan/discarded/`:
- `workspaces/plan/done/AL-1 go domain aggregations scanner cli.md`
- `workspaces/plan/done/AL-2 go embedded http server and api.md`
- `workspaces/plan/done/AL-3 typescript spa shell and fleet overview.md`
- `workspaces/plan/done/AL-11 implement native go telemetry extractor and session generation.md`
- `workspaces/plan/to-do/AL-4 typescript ticket deep dive and progression.md`
- `workspaces/plan/to-do/AL-5 single binary integration parity and deprecation.md`
- `workspaces/plan/to-do/AL-6 versioned schema migration and chronological event extraction.md`
- `workspaces/plan/to-do/AL-7 domain anomaly detection and rule engine.md`
- `workspaces/plan/to-do/AL-8 anomaly api endpoints and fleet health aggregation.md`
- `workspaces/plan/to-do/AL-9 visual trace waterfall and anomaly inspector.md`
- `workspaces/plan/discarded/AL-10 fix git metrics calculation in python extractor and backfill ledgers.md`
- `workspaces/plan/to-do/AL-12 implement go telemetry backfill engine and schema migration.md`

---

## Ready

- **AL-4**: Implement Ticket Deep Dive, Progression Charts, and Comparisons
  - **Subsystem**: Presentation / Web
  - **Phase**: Phase 1
  - **Status**: Ready
  - **Blocked by**: None (AL-2 and AL-3 Done)
  - **Blocks**: AL-5
  - **File**: `workspaces/plan/to-do/AL-4 typescript ticket deep dive and progression.md`

- **AL-12**: Implement Go Telemetry Backfill Engine and Schema Migration Command
  - **Subsystem**: Adapters / Domain / CLI
  - **Phase**: Phase 1
  - **Status**: Ready
  - **Blocked by**: None (AL-11 Done)
  - **Blocks**: AL-5, AL-6
  - **File**: `workspaces/plan/to-do/AL-12 implement go telemetry backfill engine and schema migration.md`

---

## Pending

- **AL-5**: Integrate Single Binary Build, Verify Parity, and Deprecate Python App
  - **Subsystem**: Build / Packaging / Tests
  - **Phase**: Phase 1
  - **Status**: Pending
  - **Blocked by**: AL-4, AL-11 (Done), AL-12
  - **Blocks**: AL-6
  - **File**: `workspaces/plan/to-do/AL-5 single binary integration parity and deprecation.md`

- **AL-6**: Versioned Schema Migration and Chronological Event Extraction
  - **Subsystem**: Adapters / Domain / Scripts / Tests
  - **Phase**: Phase 2
  - **Status**: Pending
  - **Blocked by**: AL-5, AL-12
  - **Blocks**: AL-7
  - **File**: `workspaces/plan/to-do/AL-6 versioned schema migration and chronological event extraction.md`

- **AL-7**: Domain Anomaly Detection and Rule Engine
  - **Subsystem**: Domain / CLI / Tests
  - **Phase**: Phase 2
  - **Status**: Pending
  - **Blocked by**: AL-6
  - **Blocks**: AL-8
  - **File**: `workspaces/plan/to-do/AL-7 domain anomaly detection and rule engine.md`

- **AL-8**: Anomaly API Endpoints and Fleet Health Aggregation
  - **Subsystem**: Adapters / HTTP / Server / Tests
  - **Phase**: Phase 2
  - **Status**: Pending
  - **Blocked by**: AL-7
  - **Blocks**: AL-9
  - **File**: `workspaces/plan/to-do/AL-8 anomaly api endpoints and fleet health aggregation.md`

- **AL-9**: Visual Trace Waterfall and Anomaly Inspector
  - **Subsystem**: Presentation / Web / Tests
  - **Phase**: Phase 2
  - **Status**: Pending
  - **Blocked by**: AL-8
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-9 visual trace waterfall and anomaly inspector.md`

---

## Done

- **AL-1**: Implement Go Domain Aggregations, Workspace Scanner, and CLI Engine
  - **Subsystem**: Domain / Adapters / CLI
  - **Phase**: Phase 1
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: AL-2
  - **File**: `workspaces/plan/done/AL-1 go domain aggregations scanner cli.md`

- **AL-2**: Implement Go HTTP Server, Embedded Static FS, and Telemetry REST API
  - **Subsystem**: Adapters / HTTP / Server
  - **Phase**: Phase 1
  - **Status**: Done
  - **Blocked by**: AL-1 (Done)
  - **Blocks**: AL-4, AL-5, AL-11
  - **File**: `workspaces/plan/done/AL-2 go embedded http server and api.md`

- **AL-3**: Build TypeScript SPA Shell and Fleet Overview Dashboard
  - **Subsystem**: Presentation / Web
  - **Phase**: Phase 1
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: AL-4
  - **File**: `workspaces/plan/done/AL-3 typescript spa shell and fleet overview.md`

- **AL-11**: Implement Native Go Telemetry Extractor and Active Session Metric Generation
  - **Subsystem**: Adapters / Domain / CLI
  - **Phase**: Phase 1
  - **Status**: Done
  - **Blocked by**: AL-2 (Done)
  - **Blocks**: AL-12, AL-5
  - **File**: `workspaces/plan/done/AL-11 implement native go telemetry extractor and session generation.md`

- **AL-10**: Fix Git Metrics Calculation in Python Extractor and Backfill Existing Workspace Ledgers
  - **Subsystem**: Scripts / Python / Workspaces / Telemetry
  - **Phase**: Phase 1
  - **Status**: Discarded (Superseded by AL-11)
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/discarded/AL-10 fix git metrics calculation in python extractor and backfill ledgers.md`
