# Decomposed Jira Tickets for AgenticLens

Individual ticket definitions are located in `workspaces/plan/done/`, `workspaces/plan/to-do/`, and `workspaces/plan/discarded/`.

---

## In Progress

(None)

---

## Ready

- **AL-37**: Cross-Project Platform Setup Manual and Telemetry Integration Guide
  - **Subsystem**: Docs / Packaging / CLI
  - **Phase**: Phase 7
  - **Status**: Ready
  - **Blocked by**: None (AL-36 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-37 cross-project setup manual and telemetry integration guide.md`

- **AL-39**: Decouple Git Analyzer from Hardcoded Branches and Folder Prefixes
  - **Subsystem**: Adapters / Git / Domain / Config / Tests
  - **Phase**: Phase 7
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-39 decouple git analyzer from hardcoded branches and folder prefixes.md`

- **AL-40**: Dynamic Configuration REST API Endpoint and Observability Dashboard Branding
  - **Subsystem**: Adapters / HTTP / Presentation / Web / Domain / Tests
  - **Phase**: Phase 7
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL-40 dynamic configuration rest api endpoint and observability dashboard branding.md`

- **AL2-1**: End-to-End Agent Anomaly Debug Prompt Generator, CLI, and Interactive Web Drawer
  - **Subsystem**: Domain / Adapters / HTTP / CLI / Presentation / Web / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL2-1 end-to-end agent anomaly debug prompt generator cli and interactive web drawer.md`

- **AL2-2**: End-to-End Workspace Archive Engine, CLI Compaction, and In-Memory Decompression Ingestion
  - **Subsystem**: Domain / Adapters / CLI / HTTP / Presentation / Web / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL2-2 end-to-end workspace archive engine cli compaction and in-memory decompression ingestion.md`

- **AL2-3**: Multi-Project Configuration Schema, Domain Aggregations, and Multi-Folder Scanner
  - **Subsystem**: Domain / Adapters / CLI / Config / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Ready
  - **Blocked by**: None
  - **Blocks**: AL2-4, AL2-5
  - **File**: `workspaces/plan/to-do/AL2-3 multi project configuration schema domain aggregations and multi folder scanner.md`

---

## Pending

(None)

- **AL2-4**: Remote Git Repository Ingestion, Shallow Cache Manager, and Project Management API
  - **Subsystem**: Adapters / Git / HTTP / Server / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Pending
  - **Blocked by**: AL2-3
  - **Blocks**: AL2-5
  - **File**: `workspaces/plan/to-do/AL2-4 remote git repository ingestion shallow cache manager and project management api.md`

- **AL2-5**: Multi-Project Web Dashboard Selector, Cross-Project Fleet Overview, and Comparative Analytics
  - **Subsystem**: Presentation / Web / Components / State / Tests
  - **Phase**: Phase 8 (v2 Phase 2)
  - **Status**: Pending
  - **Blocked by**: AL2-3, AL2-4
  - **Blocks**: None
  - **File**: `workspaces/plan/to-do/AL2-5 multi project web dashboard selector cross project fleet overview and comparative analytics.md`

---

## Done

- **AL-36**: Cross-Project Interactive Setup Skill, Automated Artifact Downloader, and Agent Workflow Integrator
  - **Subsystem**: Skills / CLI / Adapters / Scripts / Tests
  - **Phase**: Phase 7
  - **Status**: Done
  - **Blocked by**: None (AL-38 Done)
  - **Blocks**: AL-37
  - **File**: `workspaces/plan/done/AL-36 cross-project interactive setup skill automated artifact downloader and agent workflow integrator.md`

- **AL-41**: Eliminate Subprocess Churn, Artificial Sleeps, and Test Execution Bottlenecks
  - **Subsystem**: Tests / Adapters / Git / CLI / HTTP
  - **Phase**: Phase 7
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-41 eliminate subprocess churn artificial sleeps and test execution bottlenecks.md`

- **AL-35**: Pre-Release Source Code Audit, Configuration Leakage Inspection, and Review Artifact Generation
  - **Subsystem**: Review / Domain / Adapters / Presentation / Tests
  - **Phase**: Phase 7
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-35 pre-release source code audit configuration leakage inspection and review artifact generation.md`

- **AL-38**: Migrate Configuration Manifest to Root .agenticlens-config.yaml
  - **Subsystem**: Adapters / CLI / Domain / Tests
  - **Phase**: Phase 7
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: AL-36
  - **File**: `workspaces/plan/done/AL-38 migrate configuration manifest to root agenticlens-config-yaml.md`

- **AL-20**: Build TypeScript Live Observability Dashboard and Multi-Root Session Inspector
  - **Subsystem**: Presentation / Web / Tests
  - **Phase**: Phase 4
  - **Status**: Done
  - **Blocked by**: None (AL-17, AL-19 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-20 build typescript live observability dashboard and multiroot session inspector.md`

- **AL-33**: End-to-End Dynamic Local Database Event Hydration and Dual-Plane Waterfall
  - **Subsystem**: Domain / Adapters / HTTP / Presentation / Web / Tests
  - **Phase**: Phase 6
  - **Status**: Done
  - **Blocked by**: None (AL-17, AL-32 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-33 dynamic local database event hydration and fallback dual plane api.md`

- **AL-19**: Implement Live Anomaly Detection Engine and Realtime Session Health Rules
  - **Subsystem**: Domain / Adapters / HTTP / Tests
  - **Phase**: Phase 4
  - **Status**: Done
  - **Blocked by**: None (AL-17 Done)
  - **Blocks**: AL-20
  - **File**: `workspaces/plan/done/AL-19 implement live anomaly detection engine and realtime session health rules.md`

- **AL-34**: Eliminate Cumulative Granularity and Implement Bi-Directional Code Churn Visualization
  - **Subsystem**: Domain / Adapters / Presentation / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-34 eliminate cumulative granularity and implement bidirectional code churn visualization.md`

- **AL-17**: Implement Non-Locking Live Database Hook, Session Tree Engine, Ticket Inference, and Live Trace Telemetry
  - **Subsystem**: Adapters / Domain / HTTP / Tests
  - **Phase**: Phase 4
  - **Status**: Done
  - **Blocked by**: None (AL-21 Done)
  - **Blocks**: AL-19, AL-20, AL-33
  - **File**: `workspaces/plan/done/AL-17 implement nonlocking live database hook session tree engine and ticket inference.md`

- **AL-32**: Telemetry Event Pruning, Database Message Stripping, Agent Data Investigation, and Ledger Backfill
  - **Subsystem**: Domain / Adapters / CLI / Migration / Tests
  - **Phase**: Phase 6
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: AL-33
  - **File**: `workspaces/plan/done/AL-32 telemetry event pruning database message stripping and ledger backfill.md`

- **AL-31**: Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests
  - **Subsystem**: Domain / Adapters / Tests
  - **Phase**: Phase 6
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-31 decouple root session structural hierarchy from persona roles.md`

- **AL-28**: Multi-Dimensional Filtering by Ticket Type and Workspace Prefix
  - **Subsystem**: Domain / Adapters / Presentation / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None (AL-26 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-28 multidimensional filtering by ticket type and workspace prefix.md`

- **AL-29**: Conventional Commit Ticket Type Classification in Telemetry Engine
  - **Subsystem**: CLI / Adapters / Domain / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-29 conventional commit ticket type classification in telemetry engine.md`

- **AL-27**: Metric Progression Custom Window, 1-Day Window, and Hourly Granularity
  - **Subsystem**: Domain / Adapters / Presentation / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None (AL-26 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-27 metric progression custom window one-day range and hourly granularity.md`

- **AL-26**: Fleet Overview Completion Date Filtering and Autopsy Date Sorting
  - **Subsystem**: Domain / Adapters / Presentation / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None (AL-23 Done)
  - **Blocks**: AL-27, AL-28
  - **File**: `workspaces/plan/done/AL-26 fleet overview completion date filtering and autopsy date sorting.md`

- **AL-24**: Dynamic Role and Model Aggregations and Presentation
  - **Subsystem**: Domain / Presentation / Adapters / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-24 dynamic role and model aggregations and presentation.md`

- **AL-25**: Contextual Hover Information Icons for Cards and Tables
  - **Subsystem**: Presentation / Components / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-25 contextual hover information icons for cards and tables.md`

- **AL-9**: Visual Trace Waterfall and Anomaly Inspector
  - **Subsystem**: Presentation / Web / Tests
  - **Phase**: Phase 2
  - **Status**: Done
  - **Blocked by**: None (AL-8 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-9 visual trace waterfall and anomaly inspector.md`

- **AL-30**: Fail Telemetry Generation on Unresolved Root Session
  - **Subsystem**: CLI / Adapters / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-30 fail telemetry generation on missing root session.md`

- **AL-23**: Natural Ticket Identifier Ordering and Sortable Tables
  - **Subsystem**: Domain / Presentation / Adapters / Tests
  - **Phase**: Phase 5
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: AL-26
  - **File**: `workspaces/plan/done/AL-23 natural ticket identifier ordering and sortable tables.md`

- **AL-16**: Execute Codebase and Skills Refactoring with Provided Review skill
  - **Subsystem**: Domain / Adapters / Presentation / Tests
  - **Phase**: Phase 3
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-16 execute codebase and skills refactoring with provided review skill.md`

- **AL-15**: Implement Universal Multi-Version Schema Migration Engine for Telemetry Ledgers
  - **Subsystem**: Domain / Adapters / CLI / Tests
  - **Phase**: Phase 3
  - **Status**: Done
  - **Blocked by**: None (AL-12 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-15 implement universal multiversion schema migration engine for telemetry ledgers.md`

- **AL-13**: Implement Semantic Versioning and Automated CI for Pull Requests and Main Merges
  - **Subsystem**: Domain / Adapters / Presentation / CLI / Tests
  - **Phase**: Phase 3
  - **Status**: Done
  - **Blocked by**: None (AL-5 Done)
  - **Blocks**: AL-14
  - **File**: `workspaces/plan/done/AL-13 implement semantic versioning and automated ci for pull requests and main merges.md`

- **AL-14**: Implement Automated GitHub Release Workflows for Release Candidates and Stable Versions
  - **Subsystem**: CLI / Adapters / Build / Tests
  - **Phase**: Phase 3
  - **Status**: Done
  - **Blocked by**: None (AL-13 Done)
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-14 implement automated github release workflows for release candidates and stable versions.md`

- **AL-21**: Migrate Configuration to YAML and Establish Agent-Backend Bridge Definitions
  - **Subsystem**: Adapters / Domain / Config / Tests
  - **Phase**: Phase 4
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: AL-17
  - **File**: `workspaces/plan/done/AL-21 migrate configuration to yaml and establish agent backend bridge definitions.md`

- **AL-22**: Render Rich Markdown for Workspace Summary and Contract Artifacts
  - **Subsystem**: Presentation / Web / Tests
  - **Phase**: Phase 4
  - **Status**: Done
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/done/AL-22 render rich markdown for workspace summary and contract artifacts.md`

- **AL-8**: Anomaly API Endpoints and Fleet Health Aggregation
  - **Subsystem**: Adapters / HTTP / Server / Tests
  - **Phase**: Phase 2
  - **Status**: Done
  - **Blocked by**: None (AL-7 Done)
  - **Blocks**: AL-9
  - **File**: `workspaces/plan/done/AL-8 anomaly api endpoints and fleet health aggregation.md`

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

- **AL-18**: Implement Live Trace Inspector, Reasoning Stream, and Prominent Tool Failure Telemetry
  - **Subsystem**: Domain / Adapters / HTTP / Tests
  - **Phase**: Phase 4
  - **Status**: Discarded (Superseded and Merged into AL-17)
  - **Blocked by**: None
  - **Blocks**: None
  - **File**: `workspaces/plan/discarded/AL-18 implement live trace inspector reasoning stream and prominent tool failure telemetry.md`
