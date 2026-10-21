# Feature Specification: Natural Ticket Ordering, Dynamic Roles & Models, Contextual Info Icons, Date & Type Filtering, Custom Progression Windows

## Problem Statement

AgenticLens provides comprehensive multi-agent telemetry and dashboard analytics, but current visualization and aggregation mechanisms contain functional gaps that impede developer and managerial observability:

1. **Alphanumeric Ticket Sorting Inversion**: Sorting tickets alphabetically orders identifiers as `AL-1`, `AL-10`, `AL-11`, `AL-2`, `AL-22` instead of natural numerical sequence `AL-1`, `AL-2`, `AL-3` ... `AL-10`, `AL-11`, `AL-22`. This fragmentation affects the Landing Page Quick-Launch table, Fleet Overview directory, Ticket Deep Dive selector, and Comparison Matrix.
2. **Unsortable Landing Page Workspaces**: The Quick-Launch Workspaces table on the landing page is currently static, preventing users from re-ordering by ticket ID, title, status, iterations, or active duration.
3. **Hardcoded Agent Roles and Rigid Model Assumptions**: Aggregation routines and UI components enforce hardcoded roles (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`). Custom subagent roles (such as `advisor`, `po`, `researcher`, `architect`) are omitted or ignored in turn velocity calculations, peak context metrics, and effort distributions. Similarly, model breakdowns assume fixed groupings rather than dynamically accepting any model returned in telemetry.
4. **Missing Contextual Explanations on Metrics and Tables**: Visual components (KPI cards, distribution cards, data tables) lack contextual help, forcing users to guess metric definitions, calculation methodologies, or unit derivations.
5. **Absence of Fleet Overview Date Filtering**: Fleet Overview aggregates all-time workspace data without the ability to filter by completion date, making it impossible to evaluate fleet velocity across specific sprints or historical windows.
6. **Restricted Progression Windows and Granularities**: Metric Progression only supports coarse presets (7d, 30d, 90d, all) and lacks arbitrary custom date range selection, 1-day (24-hour) windows, and hourly aggregation granularity.
7. **Lack of Multi-Dimensional Filtering (Ticket Type and Prefix)**: Fleet Overview, Ticket Comparisons, and Metric Progression lack filtering by `ticket_type` and `workspace prefix`, hindering segmentation between feature tickets, bug fixes, documentation/PO tasks, and maintenance chores.
8. **Monolithic Feature Ticket Type Generation in Telemetry Skills**: Skills and CLI telemetry generation default non-feature tickets (including PO planning sessions and generic maintenance tasks) to `ticket_type: "feature"` instead of using standard Conventional Commit classifications (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).

---

## Target Personas & Use Cases

- **AI Platform Engineers & System Architects**: Need flexible date range and hourly granularity analysis to detect real-time cost runaway, evaluate token saturation across custom agent roles, and filter telemetry by workspace prefix.
- **Engineering Leads & Product Managers**: Need accurate `<PREFIX>-NN` ticket ordering, completion-date sprint filtering on the fleet overview, and explanatory metric tooltips to understand token unit economics and turn velocity without consulting documentation.
- **Agent Developers & Orchestrators**: Need dynamic support for specialized subagent roles (e.g., `researcher`, `advisor`, `po`), sortable quick-launch access, and clean Conventional Commit ticket categorization across telemetry ledgers.

---

## User Stories

1. As an engineering manager, I want tickets ordered naturally by numerical sequence (`AL-1`, `AL-2`, `AL-10`) across all tables and dropdowns, so that I can browse workspace progression chronologically without alphanumeric jumping.
2. As a platform user, I want the home Quick-Launch Workspaces table to be sortable by every column (Ticket ID, Title, Status, Iterations, Duration), so that I can immediately identify outlier tickets from the landing view.
3. As an AI system architect, I want any agent role (e.g., `advisor`, `po`, `architect`, `researcher`) and model identifier to be tracked, aggregated, and displayed dynamically without truncation or hardcoded role omissions.
4. As an analyst, I want hoverable information icons on every metric card and table header, so that I can understand how each metric is calculated directly within the dashboard.
5. As an engineering lead, I want Fleet Overview filterable by completion date (defaulting to project lifetime), so that I can inspect fleet velocity and cost across specific delivery cycles.
6. As an agent developer, I want Workspace Autopsy in Ticket Deep Dive sorted naturally by `<PREFIX>-NN` by default with an option to sort by completion date, so that I can toggle between ticket numbering and chronological delivery order.
7. As an AI platform engineer, I want Metric Progression to support custom date windows, a 1-day window option, and hourly granularity, so that I can inspect high-resolution intraday token spikes and cost progression.
8. As a tech lead, I want Fleet Overview, Comparisons, and Metric Progression filterable by `ticket_type` and `workspace prefix` (defaulting to `ticket_type: "feature"`), so that I can segment product feature metrics from chores, docs, and bug fixes.
9. As a workflow contributor, I want telemetry generation across all skills and CLI commands to classify tickets according to Conventional Commit types (`feat`, `fix`, `docs`, `chore`, `refactor`), so that PO and generic workspaces are accurately attributed.

---

## Functional Requirements

### 1. Natural Ticket Ordering & Sortable Tables
- **Pure Domain & Formatting**:
  - Implement natural ticket comparator: parse ticket identifiers into `<PREFIX>` (string) and `<NUMBER>` (integer), handling hyphens, underscores, or direct numbers (`AL-1`, `AL2`, `GEN-42`, `PO-5`).
  - Sort rule: compare prefix alphabetically; if prefixes match, compare numeric IDs numerically in ascending order; fallback to lexicographical if no numeric suffix is parsed.
- **Landing Page (Quick-Launch Workspaces)**:
  - Add stateful header sorting (`ticket_id`, `ticket_title`, `status`, `sdlc_iterations`, `active_duration_seconds`).
  - Ensure sorting by `ticket_id` uses natural ticket ordering.
- **Fleet Overview (Ticket Workspaces Directory)**:
  - Update `TicketTable.tsx` to sort by `ticket_id` using natural ordering.
- **Ticket Deep Dive (Workspace Autopsy)**:
  - Sort selector options by `<PREFIX>-NN` by default.
  - Provide sort toggle: by natural ID (default) or by completion date (`completed_at` descending, fallback to `updated_at`).

### 2. Dynamic Roles and Dynamic Models
- **Go Domain Engine (`internal/domain`)**:
  - Remove hardcoded iteration restrictions that ignore roles outside `SubagentRoleOrder`.
  - Treat all non-orchestrator roles as subagents when calculating `SubagentsPeakTokens`, `SubagentsPeakRole`, `SubagentTurnVelocity`, and `RoleEffortDistributionPct`.
  - Maintain canonical roles as priority display order, but append any dynamically discovered roles in alphabetical order.
  - Compute model metrics dynamically for all unique model strings present in invocation facts.
- **TypeScript Web Dashboard (`web/src`)**:
  - Expand `AgentRole` from fixed union to open string type (`type AgentRole = string`).
  - Update `SaturationBar.tsx`, `DistributionCard.tsx`, `KPICard.tsx`, and `TicketDeepDive.tsx` to render dynamic roles without truncation.
  - Ensure dynamic color hashing or palette assignment for custom roles and models.

### 3. Contextual Hover Information Icons
- **Information Tooltip Component**:
  - Author a reusable `InfoTooltip` / `InfoIcon` component adhering to zero-emoji policy.
  - Render a clean text-based indicator (e.g. `(i)` badge or SVG circle with "i") with accessible ARIA attributes (`aria-label`, `role="tooltip"`).
  - Reveal detailed explanation on mouse hover and focus.
- **Coverage**:
  - Add info icons to all KPI cards (Fleet, Deep Dive, Progression).
  - Add info icons to distribution cards (Role, Model).
  - Add info icons to table headers (Quick-Launch, Ticket Directory, Comparison Matrix, Live Thread View).

### 4. Fleet Overview Completion Date Filtering
- **Backend Filter Parameters**:
  - Support query parameters on `/api/fleet` and `/api/workspaces`: `window` (`all`, `1d`, `7d`, `30d`, `90d`, `custom`), `since` (ISO 8601), and `until` (ISO 8601).
  - Filter ledgers primarily by `completed_at`; if `completed_at` is null, evaluate `updated_at`.
  - Default filter: project lifetime (`all`).
- **Fleet Overview UI**:
  - Add date range selector in Fleet Overview header.
  - Recalculate fleet summary, distributions, and directory table reactively upon filter change.

### 5. Metric Progression Custom Windows & Hourly Granularity
- **Granularity & Window Expansion**:
  - Add `TimeGranularityHourly` (`"hourly"`) to Go domain and TypeScript definitions (`TimeGranularity = 'hourly' | 'daily' | 'weekly' | 'cumulative'`).
  - Add `1d` (24-hour) preset window.
  - Add `custom` window preset with `since` and `until` date-time pickers.
  - Keep default view: lifetime window (`all`) and daily granularity (`daily`).
- **Progression Bucketing**:
  - Implement hourly timestamp formatting and aggregation (`YYYY-MM-DDTHH:00:00Z`).
  - Update `TrendChart.tsx` to handle hourly bucket labels and formatting.

### 6. Multi-Dimensional Filtering (Ticket Type and Prefix)
- **Filter Parameters**:
  - Add `ticket_type` and `prefix` query filters to backend API endpoints (`/api/fleet`, `/api/workspaces`, `/api/comparison`, `/api/progression`).
  - Filter by `ticket_type`: match ledger `ticket_type` case-insensitively, handling aliases (`feature`/`feat`, `bug`/`fix`). Default selection: `"feature"`.
  - Filter by `prefix`: match workspace prefix extracted from ticket identifier (e.g. `AL`, `GEN`, `PO`, `ALL`). Default: `ALL`.
- **UI Integration**:
  - Provide filter dropdowns on Fleet Overview, Ticket Comparisons, and Metric Progression.

### 7. Conventional Commit Ticket Type Classification
- **Engine & Git Inference**:
  - Update `git.InferTicketType` to use a deterministic 3-tier resolution hierarchy:
    1. Explicit caller `--ticket-type` flag.
    2. Git branch prefix (`feat/*`, `fix/*`, `docs/*`, `chore/*`, `refactor/*`, `test/*`, `ci/*`, `perf/*`, `build/*`).
    3. Workspace ticket identifier prefix (`PO-*` -> `docs`, `GEN-*` -> `chore`).
  - Eliminate git commit message inspection (`git log --grep`) completely as commit contents are unreliable.
  - Stop inference fallbacks strictly at the workspace ticket prefix. Zero silent defaulting to `feature`.
  - Enforce explicit error or warning reporting when ticket type cannot be resolved, persisting `null`/untyped rather than falsifying telemetry.
- **Skill Manifests**:
  - Update `agentic-lens-po`, `agentic-lens-dev`, `agentic-lens-review`, orchestrator, and `AGENTS.md` to pass `--ticket-type <TYPE>` explicitly.

### 8. Fail Telemetry Generation on Unresolved Root Session
- **Strict Resolution Policy**:
  - `agentic-lens generate` must fail with non-zero exit code if no root session is identified from the provided information (nonce, session ID, ticket title, or workspace path) and no cached facts exist.
  - Zero silent output of empty ledgers (0 invocations, 0 tokens, 0 duration). Having no agentic information totally defeats the purpose of the tool.
  - Automatic fallback in session resolver checks workspace path (`workspaces/<ticket>/`) before failing.

---

## Non-Functional & Architectural Invariants

- **FCIS Discipline**: All sorting logic, date comparisons, bucket partitioning, and role/model derivations must reside in pure Go functions under `internal/domain/`. Zero I/O, zero network, zero clocks.
- **Strict Typing**: Zero `any` or empty `interface{}` across Go backend and TypeScript frontend.
- **Zero Emojis**: Strictly prohibited across code, comments, docstrings, UI icons, tooltips, and markdown artifacts.
- **Single Binary**: UI builds into `web/dist` and embeds into Go via `embed.FS` with zero Node.js runtime dependency.
- **Performance**: In-memory sorting and filtering must execute in under 1 millisecond for up to 10,000 workspaces.

---

## Decomposition & Delivery Roadmap

- [ ] **AL-23**: Implement Natural Ticket Identifier Ordering and Sortable Tables (Ready)
- [ ] **AL-24**: Implement Dynamic Role and Model Aggregations and Presentation (Ready)
- [ ] **AL-25**: Add Contextual Hover Information Icons for Cards and Tables (Ready)
- [ ] **AL-26**: Implement Fleet Overview Completion Date Filtering and Autopsy Date Sorting (Blocked by AL-23)
- [ ] **AL-27**: Implement Metric Progression Custom Window, 1-Day Window, and Hourly Granularity (Blocked by AL-26)
- [ ] **AL-28**: Implement Multi-Dimensional Filtering by Ticket Type and Workspace Prefix (Blocked by AL-26)
- [ ] **AL-29**: Implement Conventional Commit Classification in Telemetry Engine and Skill Manifests (Ready)
- [ ] **AL-30**: Fail Telemetry Generation on Unresolved Root Session (Ready)
