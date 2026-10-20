# Feature Specification: AgenticLens v2 - Agent Debug Prompt Generator, Multi-Project Observability, and Workspace Archive Engine

## Problem Statement

As multi-agent development scales across engineering teams, three major operational and architectural limitations emerge in the v1 platform:

1. **Manual Anomaly Diagnosis Friction**:
   When an autonomous coding agent encounters execution failures, tool retry loops, context exhaustion spikes, or subagent thrashing, engineering operators must manually sift through event waterfalls, locate stderr traces, interpret anomaly rule thresholds, and formulate natural-language guidance prompts for the agent. This manual triage creates cognitive friction, slows down incident resolution, and often results in vague instructions that cause agents to repeat previous mistakes. The platform lacks an automated capability to synthesize failure diagnostics, tool invocation parameters, and error traces into a ready-to-run, context-aware debugging prompt formatted directly for LLM coding agents.

2. **Single-Project Isolation in Multi-Repository Environments**:
   AgenticLens currently binds to a single local workspace directory (`workspaces/`) and a single project configuration manifest. Engineering leads and project managers overseeing multiple microservices, client applications, or cross-functional repositories are forced to run separate server instances on multiple ports, with no unified fleet dashboard, cross-project metric aggregation, or central health observability. The platform cannot natively monitor multiple local folders, git repositories, or remote git URLs within a single server instance, nor does the web dashboard allow selecting and comparing one or multiple projects simultaneously.

3. **Workspace File Bloat & Repository Accumulation**:
   Every completed ticket workspace stores a directory tree containing `agentic_metrics.json`, `SUMMARY.md`, `CONTRACT.md`, `PULL_REQUEST.md`, and iteration review files (`iterations/<NN>/`). In active repositories and continuous integration pipelines, accumulating hundreds of workspace folders bloats Git repositories, wastes filesystem inodes, and degrades directory traversal performance. Teams need a standardized command to compress older completed ticket workspaces into single compacted archives based on completion dates or interactive selection, without losing the ability of the AgenticLens server to transparently ingest, query, and visualize those archived ledgers.

---

## Target Personas & Use Cases

- **AI System Operators & Software Engineers**:
  Encounter a stuck agent or failing tool during an autonomous coding run; click "Generate Debug Prompt" in the dashboard (or run the CLI prompt command) to immediately receive a structured markdown prompt equipped with error traces, causal context, and prescriptive directives to copy to the agent.

- **Project Managers & Engineering Directors**:
  Launch a single AgenticLens instance configured to track multiple repositories across the engineering organization (local directories and remote git URLs); filter the web dashboard across individual projects or aggregate fleet velocity, token expenditures, and anomaly distributions across multiple teams.

- **DevOps Engineers & CI/CD Pipelines**:
  Automate workspace lifecycle maintenance by scheduling `agentic-lens archive --before <date>` in CI or pre-commit hooks, archiving older ticket workspaces into verified `.tar.gz` archives to keep the repository lean while ensuring historical telemetry remains queryable by the server.

---

## User Stories

1. As an agent operator, I want to request a structured debug prompt from the platform when an anomaly or tool failure occurs, so that I can immediately supply my agent with precise failure forensics and actionable instructions to fix the issue.
2. As a CLI user, I want to run `agentic-lens prompt --ticket <TICKET_ID>` to generate a markdown debug prompt directly in my terminal or pipe it into an agent session.
3. As a project manager overseeing several repositories, I want to configure a single AgenticLens server with multiple project paths and git remotes, so that all project metrics are ingested into a central platform without managing multiple server processes.
4. As a dashboard viewer, I want a multi-project selector in the navigation bar to toggle between individual projects or view aggregate fleet metrics across selected projects.
5. As an engineering lead, I want the server to dynamically read each project's own `.agenticlens-config.yaml` and ingest its workspaces, respecting project-specific ticket prefixes, role mappings, and anomaly thresholds.
6. As a CI engineer, I want an archive CLI command that compacts older workspace directories prior to a specified date into a `.tar.gz` archive and cleans up individual folders, so that repository bloat is eliminated.
7. As an interactive terminal user, I want the archive command to prompt me with a selection list when run without a date flag, so that I can pick specific workspaces to archive safely.
8. As a dashboard user, I want the server and web UI to transparently load and display metrics from archived workspace files alongside live workspaces, identifying archived tickets with clear visual indicators.

---

## Functional Scope & Layer Architecture

### 1. Agent Anomaly & Tool Error Debug Prompt Generator (AL2-1)

#### Domain Core (`internal/domain`)
- **Domain Models**:
  - `DebugPrompt`: Structured record containing `TicketID`, `SessionID`, `Turn`, `RuleID`, `Severity`, `Title`, `FormattedPrompt` (markdown string), `ToolContext`, `RemediationDirectives`, and `GeneratedAt`.
  - `DebugPromptRequest`: Parameters including `TicketID`, optional `Turn`, optional `RuleID`, optional `ToolName`, and formatting style flags.
- **Pure Prompt Synthesis Engine**:
  - `BuildDebugPrompt(anomaly *Anomaly, event *EventFact, recentEvents []EventFact, ticketID string, projectCfg ProjectConfig) DebugPrompt`: Pure functional builder that takes the detected anomaly, failing event fact, causal preceding event sequence (3 to 5 preceding turns/tool invocations), and project configuration.
  - Generates deterministic, highly structured markdown with dedicated sections:
    1. `# Incident Overview`: Ticket ID, Session ID, Turn Number, Role, Model, Timestamp, Severity.
    2. `## Anomaly Diagnostics`: Rule ID, plain-English symptom description, triggered thresholds (e.g. 3 consecutive tool failures, 85% context ceiling breach).
    3. `## Tool Invocation Forensics`: Exact tool name, input arguments formatted as code/JSON, exit status, duration, error message, and stderr payload.
    4. `## Preceding Execution Trace`: Brief chronological summary of prior turns to explain how the agent arrived at the current error state.
    5. `## Prescriptive Remediation Directives`: Rule-tailored instructions instructing the agent what to avoid and what concrete diagnostic actions to take next:
       - `TOOL_RETRY_STREAK`: Instruct agent to stop repeating identical arguments, inspect underlying filesystem/environment state first, and verify preconditions.
       - `TOOL_HIGH_FAILURE_RATE`: Instruct agent to verify command syntax, check path existence, or switch to alternative tools.
       - `CONTEXT_SATURATION_SPIKE`: Instruct agent to summarize findings, avoid dumping large files into context, truncate logs, or trigger compaction.
       - `SUBAGENT_REWORK_LOOP`: Instruct agent to re-read `CONTRACT.md`, verify requirements, and reconcile divergent test expectations.
       - `LIVE_ANOMALY` / Stalls: Instruct agent to check for hanging processes, deadlocks, or long-running commands.

#### Adapters & HTTP Shell (`internal/adapters/http`, `cmd/agentic-lens`)
- **HTTP REST Endpoint**:
  - `GET /api/workspaces/{ticketId}/debug-prompt`: Accepts optional query parameters `turn`, `rule_id`, `tool_name`. Resolves workspace ledger, locates target anomaly/event, synthesizes prompt via domain engine, and returns `DebugPromptResponse` (JSON containing raw text and metadata).
- **CLI Subcommand**:
  - `agentic-lens prompt --ticket <ticketId> [--turn <n>] [--rule <rule_id>] [--output <file>]`: Evaluates workspace ledger, builds prompt, prints formatted markdown to stdout, or writes to file.

#### Presentation (`web/`)
- **Anomaly Drawer Integration**:
  - Add "Generate Debug Prompt" action button inside `AnomalyDrawer.tsx` beside the Rule Diagnosis section.
  - Dedicated Debug Prompt View with syntax-highlighted Markdown, raw toggle, and one-click "Copy Prompt to Clipboard" button with transient feedback state.
- **Trace Waterfall & Live Banner Actions**:
  - Provide a quick "Debug Prompt" button on error nodes in the Waterfall trace inspector and on active alerts in `LiveAnomalyBanner.tsx`.

---

### 2. Multi-Project Workspace Observability & Remote Ingestion (AL2-3, AL2-4, AL2-5)

#### Domain Core (`internal/domain`)
- **Multi-Project Models**:
  - `ProjectDefinition`: `ID` (slug), `Name` (display label), `SourceType` (`local_folder`, `git_repo`, `git_remote`), `SourcePath` (path or URL), `Branch` (git branch, defaults to `main`), `WorkspacesDir` (relative path, defaults to `workspaces`), `ConfigPath` (optional path to project config).
  - `MultiProjectConfig`: Slice of `ProjectDefinition` and global multi-project options.
  - Ledger Tagging: Extend `AgenticMetricsLedger` and `ScanReport` with `ProjectID string` and `ProjectName string`.
- **Pure Multi-Project Aggregations**:
  - `FilterLedgersByProjects(ledgers []AgenticMetricsLedger, projectIDs []string) []AgenticMetricsLedger`: Pure filtering supporting single or multi-project inclusion.
  - `CalculateMultiProjectSummary(ledgers []AgenticMetricsLedger) MultiProjectSummary`: Computes global totals alongside per-project rollups (cost, tokens, turn velocity, anomaly counts).

#### Adapters & Repository Shell (`internal/adapters`, `internal/adapters/git`, `cmd/agentic-lens`)
- **Configuration Schema**:
  - Extend `.agenticlens-config.yaml` to accept an optional `projects:` block:
    ```yaml
    projects:
      - id: core-platform
        name: Core Platform
        path: .
      - id: billing-service
        name: Billing Service
        path: /home/user/repos/billing
      - id: auth-worker
        name: Auth Worker
        path: https://github.com/myorg/auth-worker.git
        branch: main
    ```
- **CLI Integration**:
  - Server and scan commands accept repeatable `--project <id>=<path_or_url>` flags and `--projects-config <path>` to augment or override configuration.
- **Remote Git Ingestion Manager (`internal/adapters/git`)**:
  - Local shallow cache directory: `~/.cache/agenticlens/remotes/<project_id>/`.
  - Automatic shallow clone/fetch (`git clone --depth 1` / `git fetch --depth 1`) for remote URLs.
  - Safe URL validation (HTTP/HTTPS/SSH) and credential isolation.
- **Multi-Directory Workspace Scanner**:
  - Extends `ScanWorkspaces` to scan all configured projects.
  - For each project, loads that project's specific configuration manifest (for ticket prefixes, role mappings, and custom metrics), falling back to defaults if absent.
  - Ingests all raw metrics ledgers (`agentic_metrics.json` and archives) from each project's workspace folder, tagging each ledger with `ProjectID` and `ProjectName`.
- **HTTP API Expansion**:
  - `GET /api/projects`: Returns list of monitored projects with status, source type, path, ledger count, and last sync timestamp.
  - `POST /api/projects/sync`: Triggers asynchronous refresh of git remotes and rescans local folders.
  - Update `GET /api/workspaces`, `GET /api/fleet`, `GET /api/comparison`, and `GET /api/progression` to support `?project=<id>` (repeatable, or comma-separated `?project=p1,p2`).

#### Presentation (`web/`)
- **Global Project Selector**:
  - Integrated into the primary `Navbar.tsx` and filter bars across views.
  - Allows selecting "All Projects" or checking individual project checkboxes.
  - Synchronizes selection with URL query parameter (`?projects=core-platform,billing-service`) and localStorage for persistence.
- **Fleet & Comparison Views**:
  - Fleet Overview displays multi-project summary cards and comparative bar/line charts comparing velocity, cost, and health across projects.
  - Workspaces table displays a "Project" column with distinct visual color badges.
  - Ticket Deep Dive displays project breadcrumbs and origin metadata.

---

### 3. Workspace Archive Command & Compressed Ingestion (AL2-2)

#### Domain Core (`internal/domain`)
- **Archive Invariants & Models**:
  - `ArchiveManifest`: Metadata record stored inside archive containing `ArchiveVersion`, `CreatedAt`, `ArchivedBy`, `TicketCount`, and `Tickets []ArchivedTicketEntry`.
  - `ArchivedTicketEntry`: `TicketID`, `CompletedAt`, `FileCount`, `ByteSize`, `ChecksumSHA256`.
  - Ledger Annotation: `IsArchived bool` and `ArchiveSource string` attached to `AgenticMetricsLedger`.
- **Pure Date Filtering**:
  - `FilterWorkspacesForArchive(entries []WorkspaceDirectoryInfo, cutoffDate time.Time) []WorkspaceDirectoryInfo`: Pure evaluation of eligible directories.

#### Adapters & Storage Shell (`internal/adapters`, `cmd/agentic-lens`)
- **Archive Creation Engine**:
  - Subcommand `agentic-lens archive`:
    - `--before <YYYY-MM-DD|RFC3339>`: Non-interactive batch mode for CI and automated scripts. Archives all workspace directories completed prior to `<cutoff>`.
    - `--workspaces <dir>`: Target workspaces directory.
    - `--output <path>`: Destination archive file path (defaults to `<workspaces>/archive-<timestamp>.tar.gz` or `<workspaces>/archives/`).
    - `--keep`: Preserves source workspace directories on disk (default behavior: removes archived directories after verifying archive integrity).
    - `--dry-run`: Reports which directories would be archived without modifying disk.
  - Interactive Terminal Mode:
    - When executed in a TTY without `--before`, scans existing workspace directories, extracts ticket IDs, completion dates, and disk usage, and presents an interactive selection prompt to the user.
  - Archive Safety & Integrity Verification:
    - Writes archive atomically to a temporary file (`.tmp`) before moving to final destination.
    - Standard `.tar.gz` format incorporating `manifest.json`, each ticket's `agentic_metrics.json`, `SUMMARY.md`, `CONTRACT.md`, `PULL_REQUEST.md`, and iteration folders.
    - Reads back and validates tarball entries and SHA256 checksums before deleting any source directory.
- **Transparent In-Memory Archive Ingestion**:
  - `ScanWorkspaces` enhanced to discover both workspace directories and `.tar.gz` archives in the workspaces root (and `archives/` subfolder).
  - Uses `archive/tar` and `compress/gzip` to parse `agentic_metrics.json` directly from the tar stream without extracting files to disk.
  - Adapters `LoadTicketWorkspace` and `ReadTicketArtifacts` query in-memory index or stream artifact contents directly from the tar archive.
  - Deduplication: If a ticket ID exists both on disk and in an archive, the live on-disk directory takes precedence.

#### Presentation (`web/`)
- Workspaces table displays an "Archived" badge for archived tickets.
- Filter toggle in the filter bar: "All Workspaces", "Active Only", "Archived Only".
- Ticket Deep Dive displays archive origin banner and source archive filename.

---

## Non-Functional & Architectural Invariants

1. **Functional Core, Imperative Shell (FCIS)**:
   - All prompt generation formatting, multi-project metric rollups, date filtering, and archive manifest calculations must reside in `internal/domain` with pure functions, value semantics, and zero I/O imports.
   - All filesystem operations, tarball compression/decompression, git remote cloning, and HTTP handlers reside in `internal/adapters` and `cmd/`.
2. **Zero Any & Strict Typing**:
   - Zero `any` or `interface{}` across Go packages and TypeScript code.
   - Discriminated unions and strict interfaces for all API payloads and UI components.
3. **Zero Mocks in Domain Tests**:
   - All Go domain logic tested using table-driven test cases with zero mocks.
   - Adapter tests use real temporary directories (`t.TempDir()`), real tar archives, and local git fixtures.
4. **Zero Emojis**:
   - Strictly zero emojis in code, logs, comments, commit messages, PR drafts, or UI components.
5. **Memory & Streaming Efficiency**:
   - Archive scanning must stream entries without reading entire tarballs into memory all at once.
   - Remote git caching must use shallow clones (`--depth 1`) to minimize network bandwidth and disk consumption.

---

## Decomposition & Delivery Roadmap (Phase 8: AL2 Prefix)

The initiative is decomposed into 5 vertical tracer-bullet development tickets structured across 3 execution tiers:

### Tier 1 (Ready - Unblocked)
- **AL2-1**: End-to-End Agent Anomaly Debug Prompt Generator, CLI, and Interactive Web Drawer
  - *Subsystem*: Domain / Adapters / HTTP / CLI / Presentation / Web / Tests
  - *Blocked by*: None
  - *Blocks*: None
- **AL2-2**: End-to-End Workspace Archive Engine, CLI Compaction, and In-Memory Decompression Ingestion
  - *Subsystem*: Domain / Adapters / CLI / HTTP / Presentation / Web / Tests
  - *Blocked by*: None
  - *Blocks*: None
- **AL2-3**: Multi-Project Configuration Schema, Domain Aggregations, and Multi-Folder Scanner
  - *Subsystem*: Domain / Adapters / CLI / Config / Tests
  - *Blocked by*: None
  - *Blocks*: AL2-4, AL2-5

### Tier 2 (Pending - Blocked by Tier 1)
- **AL2-4**: Remote Git Repository Ingestion, Shallow Cache Manager, and Project Management API
  - *Subsystem*: Adapters / Git / HTTP / Server / Tests
  - *Blocked by*: AL2-3
  - *Blocks*: AL2-5

### Tier 3 (Pending - Blocked by Tier 2)
- **AL2-5**: Multi-Project Web Dashboard Selector, Cross-Project Fleet Overview, and Comparative Analytics
  - *Subsystem*: Presentation / Web / Components / State / Tests
  - *Blocked by*: AL2-3, AL2-4
  - *Blocks*: None
