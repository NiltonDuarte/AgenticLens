# Decomposed Jira Tickets for Python to Go + TypeScript Rewrite

Individual ticket definitions are located in `workspaces/plan/to-do/`:
- `workspaces/plan/to-do/AL-1 go domain aggregations scanner cli.md`
- `workspaces/plan/to-do/AL-2 go embedded http server and api.md`
- `workspaces/plan/to-do/AL-3 typescript spa shell and fleet overview.md`
- `workspaces/plan/to-do/AL-4 typescript ticket deep dive and progression.md`
- `workspaces/plan/to-do/AL-5 single binary integration parity and deprecation.md`

---

## Execution Tiers & Dependency Graph

### Tier 1 (Ready to Execute)
- **AL-1**: Implement Go Domain Aggregations, Workspace Scanner, and CLI Engine
  - **Subsystem**: Domain / Adapters / CLI
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: AL-2
  - **File**: `workspaces/plan/to-do/AL-1 go domain aggregations scanner cli.md`

### Tier 2 (Blocked by Tier 1)
- **AL-2**: Implement Go HTTP Server, Embedded Static FS, and Telemetry REST API
  - **Subsystem**: Adapters / HTTP / Server
  - **Status**: Pending
  - **Blocked by**: AL-1
  - **Blocks**: AL-3
  - **File**: `workspaces/plan/to-do/AL-2 go embedded http server and api.md`

### Tier 3 (Blocked by Tier 2)
- **AL-3**: Build TypeScript SPA Shell and Fleet Overview Dashboard
  - **Subsystem**: Presentation / Web
  - **Status**: Pending
  - **Blocked by**: AL-2
  - **Blocks**: AL-4
  - **File**: `workspaces/plan/to-do/AL-3 typescript spa shell and fleet overview.md`

### Tier 4 (Blocked by Tier 3)
- **AL-4**: Implement Ticket Deep Dive, Progression Charts, and Comparisons
  - **Subsystem**: Presentation / Web
  - **Status**: Pending
  - **Blocked by**: AL-3
  - **Blocks**: AL-5
  - **File**: `workspaces/plan/to-do/AL-4 typescript ticket deep dive and progression.md`

### Tier 5 (Blocked by Tier 4)
- **AL-5**: Integrate Single Binary Build, Verify Parity, and Deprecate Python App
  - **Subsystem**: Build / Packaging / Tests
  - **Status**: Pending
  - **Blocked by**: AL-4
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-5 single binary integration parity and deprecation.md`
