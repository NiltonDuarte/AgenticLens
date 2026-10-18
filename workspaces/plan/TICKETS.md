# Decomposed Jira Tickets for AgenticLens

Individual ticket definitions are located in `workspaces/plan/done/`, `workspaces/plan/to-do/`, and `workspaces/plan/discarded/`:
- `workspaces/plan/done/AL-1 go domain aggregations scanner cli.md`
- `workspaces/plan/done/AL-2 go embedded http server and api.md`
- `workspaces/plan/done/AL-3 typescript spa shell and fleet overview.md`
- `workspaces/plan/done/AL-4 typescript ticket deep dive and progression.md`
- `workspaces/plan/done/AL-5 single binary integration parity and deprecation.md`
- `workspaces/plan/done/AL-6 versioned schema migration and chronological event extraction.md`
- `workspaces/plan/done/AL-7 domain anomaly detection and rule engine.md`
- `workspaces/plan/done/AL-11 implement native go telemetry extractor and session generation.md`
- `workspaces/plan/done/AL-12 implement go telemetry backfill engine and schema migration.md`
- `workspaces/plan/to-do/AL-8 anomaly api endpoints and fleet health aggregation.md`
- `workspaces/plan/to-do/AL-9 visual trace waterfall and anomaly inspector.md`
- `workspaces/plan/to-do/AL-13 implement semantic versioning and automated ci for pull requests and main merges.md`
- `workspaces/plan/to-do/AL-14 implement automated github release workflows for release candidates and stable versions.md`
- `workspaces/plan/to-do/AL-15 implement universal multiversion schema migration engine for telemetry ledgers.md`
- `workspaces/plan/to-do/AL-16 execute codebase and skills refactoring with provided review skill.md`
- `workspaces/plan/to-do/AL-17 implement nonlocking live database hook session tree engine and ticket inference.md`
- `workspaces/plan/to-do/AL-18 implement live trace inspector reasoning stream and prominent tool failure telemetry.md`
- `workspaces/plan/to-do/AL-19 implement live anomaly detection engine and realtime session health rules.md`
- `workspaces/plan/to-do/AL-20 build typescript live observability dashboard and multiroot session inspector.md`
- `workspaces/plan/to-do/AL-21 migrate configuration to yaml and establish agent backend bridge definitions.md`
- `workspaces/plan/to-do/AL-22 render rich markdown for workspace summary and contract artifacts.md`
- `workspaces/plan/discarded/AL-10 fix git metrics calculation in python extractor and backfill ledgers.md`

---

## Ready

- **AL-8**: Anomaly API Endpoints and Fleet Health Aggregation
  - **Subsystem**: Adapters / HTTP / Server / Tests
  - **Phase**: Phase 2
  - **Status**: Ready
  - **Blocked by**: None (AL-7 Done)
  - **Blocks**: AL-9
  - **File**: `workspaces/plan/to-do/AL-8 anomaly api endpoints and fleet health aggregation.md`

- **AL-13**: Implement Semantic Versioning and Automated CI for Pull Requests and Main Merges
  - **Subsystem**: Domain / Adapters / Presentation / CLI / Tests
  - **Phase**: Phase 3
  - **Status**: Ready
  - **Blocked by**: None (AL-5 Done)
  - **Blocks**: AL-14
  - **File**: `workspaces/plan/to-do/AL-13 implement semantic versioning and automated ci for pull requests and main merges.md`

- **AL-15**: Implement Universal Multi-Version Schema Migration Engine for Telemetry Ledgers
  - **Subsystem**: Domain / Adapters / CLI / Tests
  - **Phase**: Phase 3
  - **Status**: Ready
  - **Blocked by**: None (AL-12 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-15 implement universal multiversion schema migration engine for telemetry ledgers.md`

- **AL-16**: Execute Codebase and Skills Refactoring with Provided Review Skill
  - **Subsystem**: Domain / Adapters / Presentation / Tests
  - **Phase**: Phase 3
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-16 execute codebase and skills refactoring with provided review skill.md`

- **AL-21**: Migrate Configuration to YAML and Establish Agent-Backend Bridge Definitions
  - **Subsystem**: Adapters / Domain / Config / Tests
  - **Phase**: Phase 4
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: AL-17
  - **File**: `workspaces/plan/to-do/AL-21 migrate configuration to yaml and establish agent backend bridge definitions.md`

- **AL-22**: Render Rich Markdown for Workspace Summary and Contract Artifacts
  - **Subsystem**: Presentation / Web / Tests
  - **Phase**: Phase 4
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-22 render rich markdown for workspace summary and contract artifacts.md`

---

## Pending

- **AL-9**: Visual Trace Waterfall and Anomaly Inspector
  - **Subsystem**: Presentation / Web / Tests
  - **Phase**: Phase 2
  - **Status**: Pending
  - **Blocked by**: AL-8
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-9 visual trace waterfall and anomaly inspector.md`

- **AL-14**: Implement Automated GitHub Release Workflows for Release Candidates and Stable Versions
  - **Subsystem**: CLI / Adapters / Build / Tests
  - **Phase**: Phase 3
  - **Status**: Pending
  - **Blocked by**: AL-13
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-14 implement automated github release workflows for release candidates and stable versions.md`

- **AL-17**: Implement Non-Locking Live Database Hook, Session Tree Engine, and Config-Driven Ticket Inference
  - **Subsystem**: Adapters / Domain / HTTP / Tests
  - **Phase**: Phase 4
  - **Status**: Pending
  - **Blocked by**: AL-21
  - **Blocks**: AL-18
  - **File**: `workspaces/plan/to-do/AL-17 implement nonlocking live database hook session tree engine and ticket inference.md`

- **AL-18**: Implement Live Trace Inspector, Reasoning Stream, and Prominent Tool Failure Telemetry
  - **Subsystem**: Domain / Adapters / HTTP / Tests
  - **Phase**: Phase 4
  - **Status**: Pending
  - **Blocked by**: AL-17
  - **Blocks**: AL-19, AL-20
  - **File**: `workspaces/plan/to-do/AL-18 implement live trace inspector reasoning stream and prominent tool failure telemetry.md`

- **AL-19**: Implement Live Anomaly Detection Engine and Real-Time Session Health Rules
  - **Subsystem**: Domain / Adapters / HTTP / Tests
  - **Phase**: Phase 4
  - **Status**: Pending
  - **Blocked by**: AL-18
  - **Blocks**: AL-20
  - **File**: `workspaces/plan/to-do/AL-19 implement live anomaly detection engine and realtime session health rules.md`

- **AL-20**: Build TypeScript Live Observability Dashboard and Multi-Root Session Inspector
  - **Subsystem**: Presentation / Web / Tests
  - **Phase**: Phase 4
  - **Status**: Pending
  - **Blocked by**: AL-18, AL-19
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-20 build typescript live observability dashboard and multiroot session inspector.md`

---

## Done

- **AL-7**: Domain Anomaly Detection and Rule Engine
  - **Subsystem**: Domain / CLI / Tests
  - **Phase**: Phase 2
  - **Status**: Done
  - **Blocked by**: None (AL-6 Done)
  - **Blocks**: AL-8
  - **File**: `workspaces/plan/done/AL-7 domain anomaly detection and rule engine.md`

- **AL-6**: Chronological Event Extraction and Schema Evolution in Go
  - **Subsystem**: Adapters / Domain / CLI / Tests
  - **Phase**: Phase 2
  - **Status**: Done
  - **Blocked by**: None (AL-5, AL-12 Done)
  - **Blocks**: AL-7
  - **File**: `workspaces/plan/done/AL-6 versioned schema migration and chronological event extraction.md`

- **AL-5**: Integrate Single Binary Build, Verify Parity, and Deprecate Python App
  - **Subsystem**: Build / Packaging / Tests
  - **Phase**: Phase 1
  - **Status**: Done
  - **Blocked by**: None (AL-4, AL-11, AL-12 Done)
  - **Blocks**: AL-6
  - **File**: `workspaces/plan/done/AL-5 single binary integration parity and deprecation.md`

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

- **AL-4**: Implement Ticket Deep Dive, Progression Charts, and Comparisons
  - **Subsystem**: Presentation / Web
  - **Phase**: Phase 1
  - **Status**: Done
  - **Blocked by**: None (AL-2 and AL-3 Done)
  - **Blocks**: AL-5
  - **File**: `workspaces/plan/done/AL-4 typescript ticket deep dive and progression.md`

- **AL-11**: Implement Native Go Telemetry Extractor and Active Session Metric Generation
  - **Subsystem**: Adapters / Domain / CLI
  - **Phase**: Phase 1
  - **Status**: Done
  - **Blocked by**: AL-2 (Done)
  - **Blocks**: AL-12, AL-5
  - **File**: `workspaces/plan/done/AL-11 implement native go telemetry extractor and session generation.md`

- **AL-12**: Implement Go Telemetry Backfill Engine and Schema Migration Command
  - **Subsystem**: Adapters / Domain / CLI
  - **Phase**: Phase 1
  - **Status**: Done
  - **Blocked by**: AL-11 (Done)
  - **Blocks**: AL-5, AL-6
  - **File**: `workspaces/plan/done/AL-12 implement go telemetry backfill engine and schema migration.md`

- **AL-10**: Fix Git Metrics Calculation in Python Extractor and Backfill Existing Workspace Ledgers
  - **Subsystem**: Scripts / Python / Workspaces / Telemetry
  - **Phase**: Phase 1
  - **Status**: Discarded (Superseded by AL-11)
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/discarded/AL-10 fix git metrics calculation in python extractor and backfill ledgers.md`
