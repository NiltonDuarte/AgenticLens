# Decomposed Jira Tickets for AgenticLens

Individual ticket definitions are located in `workspaces/plan/done/` and `workspaces/plan/to-do/`:
- `workspaces/plan/done/AL-1 go domain aggregations scanner cli.md`
- `workspaces/plan/to-do/AL-2 go embedded http server and api.md`
- `workspaces/plan/to-do/AL-3 typescript spa shell and fleet overview.md`
- `workspaces/plan/to-do/AL-4 typescript ticket deep dive and progression.md`
- `workspaces/plan/to-do/AL-5 single binary integration parity and deprecation.md`
- `workspaces/plan/to-do/AL-6 versioned schema migration and chronological event extraction.md`
- `workspaces/plan/to-do/AL-7 domain anomaly detection and rule engine.md`
- `workspaces/plan/to-do/AL-8 anomaly api endpoints and fleet health aggregation.md`
- `workspaces/plan/to-do/AL-9 visual trace waterfall and anomaly inspector.md`

---

## Execution Tiers & Dependency Graph

### Tier 1 (Ready to Execute)
- **AL-1**: Implement Go Domain Aggregations, Workspace Scanner, and CLI Engine
  - **Subsystem**: Domain / Adapters / CLI
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: AL-2
  - **File**: `workspaces/plan/done/AL-1 go domain aggregations scanner cli.md`

- **AL-2**: Implement Go HTTP Server, Embedded Static FS, and Telemetry REST API
  - **Subsystem**: Adapters / HTTP / Server
  - **Status**: Ready
  - **Blocked by**: AL-1 (Done)
  - **Blocks**: AL-4, AL-5
  - **File**: `workspaces/plan/to-do/AL-2 go embedded http server and api.md`

- **AL-3**: Build TypeScript SPA Shell and Fleet Overview Dashboard
  - **Subsystem**: Presentation / Web
  - **Status**: Ready (Decoupled with mock fixture harness)
  - **Blocked by**: None
  - **Blocks**: AL-4
  - **File**: `workspaces/plan/to-do/AL-3 typescript spa shell and fleet overview.md`

### Tier 2 (Blocked by AL-2 and AL-3)
- **AL-4**: Implement Ticket Deep Dive, Progression Charts, and Comparisons
  - **Subsystem**: Presentation / Web
  - **Status**: Pending
  - **Blocked by**: AL-2, AL-3
  - **Blocks**: AL-5
  - **File**: `workspaces/plan/to-do/AL-4 typescript ticket deep dive and progression.md`

### Tier 3 (Blocked by Tier 2)
- **AL-5**: Integrate Single Binary Build, Verify Parity, and Deprecate Python App
  - **Subsystem**: Build / Packaging / Tests
  - **Status**: Pending
  - **Blocked by**: AL-4
  - **Blocks**: AL-6
  - **File**: `workspaces/plan/to-do/AL-5 single binary integration parity and deprecation.md`

---

## Phase 2: Schema v2, Anomaly Engine, and Visual Waterfall

### Tier 6 (Blocked by Tier 5)
- **AL-6**: Versioned Schema Migration and Chronological Event Extraction
  - **Subsystem**: Adapters / Domain / Scripts / Tests
  - **Status**: Pending
  - **Blocked by**: AL-5
  - **Blocks**: AL-7
  - **File**: `workspaces/plan/to-do/AL-6 versioned schema migration and chronological event extraction.md`

### Tier 7 (Blocked by Tier 6)
- **AL-7**: Domain Anomaly Detection and Rule Engine
  - **Subsystem**: Domain / CLI / Tests
  - **Status**: Pending
  - **Blocked by**: AL-6
  - **Blocks**: AL-8
  - **File**: `workspaces/plan/to-do/AL-7 domain anomaly detection and rule engine.md`

### Tier 8 (Blocked by Tier 7)
- **AL-8**: Anomaly API Endpoints and Fleet Health Aggregation
  - **Subsystem**: Adapters / HTTP / Server / Tests
  - **Status**: Pending
  - **Blocked by**: AL-7
  - **Blocks**: AL-9
  - **File**: `workspaces/plan/to-do/AL-8 anomaly api endpoints and fleet health aggregation.md`

### Tier 9 (Blocked by Tier 8)
- **AL-9**: Visual Trace Waterfall and Anomaly Inspector
  - **Subsystem**: Presentation / Web / Tests
  - **Status**: Pending
  - **Blocked by**: AL-8
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-9 visual trace waterfall and anomaly inspector.md`
