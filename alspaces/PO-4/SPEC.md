# Feature Specification: CI/CD Automation, Release Pipelines, Multi-Version Migration, Review Skill, Live Observability, YAML Config Bridge, and Rich Artifact Rendering

## Problem Statement

AgenticLens has achieved core single-binary capability and telemetry extraction, but operates with key production gaps across seven critical areas:
1. **Unautomated Verification and Unenforced CI**: Pull requests and merges to `main` lack automated GitHub Actions CI checks (`vet`, `lint`, race testing, diff checks, build tests), risking regressions before release.
2. **Missing Semantic Versioning and Automated Release Distribution**: The application lacks compiled-in semantic versioning (`Version`, `GitCommit`, `BuildDate`), CLI version inspection, and automated GitHub workflows to publish Release Candidates (RC) on commit to `main` and official Stable releases on `vX.Y.Z` tags with a maintained `latest` release pointer.
3. **Absence of Universal Schema Migration Pipeline**: As telemetry schemas evolve (e.g. from unversioned v1 to v2 events and future revisions), there is no sequential, forward-migrating engine that guarantees existing historical `agentic_metrics.json` files from ANY previous version can be upgraded to the latest version deterministically.
4. **Codebase and Skill Drift**: Rapid feature development requires a comprehensive refactoring pass guided by a provided `agentic-lens-review` skill to systematically audit diffs and codebase structures for FCIS adherence, typing rules (zero `any`), emoji prohibition, and performance invariants.
5. **Underutilized JSON Configuration**: Project configuration is currently constrained to `.agenticlens/config.json` with minimal keys. The backend duplicates hardcoded values (ticket regexes, database pragmas, timeouts, role mappings). Migrating to `.agenticlens/config.yaml` provides human-readable, commented configuration and establishes concrete schema definitions that bridge agent workflows and backend engines.
6. **Lack of Non-Locking Live Observability**: Users and system operators have zero real-time visibility into active multi-agent sessions. There is no transient, non-locking dashboard to inspect concurrent root sessions and sub-sessions, live reasoning thoughts, inter-agent messages, config-driven case-insensitive ticket guessing, live anomaly detection, and prioritized failed tool calls without interfering with or locking the underlying SQLite database.
7. **Raw Monospace Workspace Artifact Rendering**: The web dashboard currently renders `CONTRACT.md` and `SUMMARY.md` inside raw `<pre><code>` code blocks instead of parsed Markdown, making telemetry summary tables, section headers, and task lists difficult to read.

## Target Personas & Use Cases
- **Platform Engineers & Maintainers**: Need automated CI quality gates, semantic versioning, and continuous release pipelines for RC and Stable builds.
- **Telemetry Authors & Migration Operators**: Need idempotent, chainable schema migration from all historical ledger versions to current formats.
- **Code Reviewers & Orchestrators**: Need systematic audits and refactoring execution guided by the provided `agentic-lens-review` skill.
- **Agent Developers & AI System Operators**: Need live, non-intrusive runtime visibility into executing agent sessions, under-the-hood reasoning traces, active tool calls, case-insensitive ticket detection, prominent failure alerts, and anomaly indicators.
- **Product Owners & Reviewers**: Need rich, formatted Markdown rendering for workspace `CONTRACT.md` and `SUMMARY.md` artifacts.

## User Stories
1. As a platform maintainer, I want CI to execute on every pull request and push to `main` and assert all checks pass before any release, so that regressions are blocked automatically.
2. As a release engineer, I want GitHub workflows to publish release candidate binaries on every commit to `main` and stable releases on tags `vX.Y.Z` with an updated `latest` pointer, so that users always have access to bleeding-edge and stable binaries.
3. As a telemetry operator, I want the software to migrate `agentic_metrics.json` from all previous schema versions to the latest version idempotently, so that historical ledgers remain fully compatible.
4. As a code maintainer, I want to execute a comprehensive codebase and skills refactoring pass guided by the provided `agentic-lens-review` skill, so that FCIS and typing invariants are enforced.
5. As an agent developer, I want the project configuration migrated to YAML with rich agent-backend bridge definitions, so that ticket matching, database pragmas, and live thresholds are centrally managed.
6. As an agent developer, I want a live observability dashboard that hooks into the database without locking it to view active multi-root sessions, sub-sessions, inter-agent messages, reasonings, case-insensitive ticket guessing, live anomaly detection, and prioritized failed tool calls in real time.
7. As a dashboard user, I want `CONTRACT.md` and `SUMMARY.md` rendered as formatted Markdown with tables, lists, and headers instead of raw code blocks, so that workspace artifacts are readable.

## Functional Requirements

### 1. Semantic Versioning and Continuous Integration (AL-13)
- Declare semantic version fields in Go: `Version`, `GitCommit`, `BuildDate`.
- Wire build-time `-ldflags` in `Makefile` and `go build`.
- Add `agentic-lens version` CLI command and `/api/version` HTTP endpoint.
- Render semantic version badge in the TypeScript SPA navigation bar.
- Add `.github/workflows/ci.yml` running `make check`, `make test-backend` (`-race`), `make test-frontend`, and `make build` on PRs and `main`.

### 2. Automated Release Workflows (AL-14)
- Release Candidate workflow (`.github/workflows/release-candidate.yml`): triggers on push to `main` after CI; generates `vX.Y.Z-rc.<run_number>+<short_sha>` pre-release with compiled binaries.
- Stable Release workflow (`.github/workflows/release-stable.yml`): triggers on `v*.*.*` tags; validates semver format, compiles multi-platform binaries, creates official GitHub Release, and updates the floating `latest` tag and release pointer.

### 3. Universal Multi-Version Schema Migration Engine (AL-15)
- Pure domain stepwise migration pipeline: `Migrate(input []byte, targetVersion int) (AgenticFactLedger, error)`.
- Sequential upgrade chain: `v1 -> v2 -> ... -> vLatest`, ensuring non-destructive transformations.
- Expose `agentic-lens migrate` (and `--migrate` option in `backfill`).
- Discovers ledgers across workspaces, identifies schema version, executes migration steps, and atomically updates files with backup options.
- Fixture test suites validating forward migration from all historical versions.

### 4. Codebase & Skills Refactoring with Provided Review Skill (AL-16)
- The `agentic-lens-review` skill is provided externally to guide deep multi-agent auditing.
- Execute systematic refactoring across Go backend and TypeScript frontend:
  - Consolidate duplicated formatters, date utilities, and chart helpers.
  - Eliminate dead code and verify zero `any` or empty `interface{}` across all layers.
  - Verify strict FCIS purity in `internal/domain/`.
  - Ensure zero emojis across all code, comments, logs, and artifacts.

### 5. YAML Configuration Migration and Agent-Backend Bridge Definitions (AL-21)
- Migrate `.agenticlens/config.json` to human-readable `.agenticlens/config.yaml`.
- Go adapter configuration loader with YAML parser (`gopkg.in/yaml.v3`) and backward-compatible JSON fallback.
- Enriched configuration schema bridging agent and backend:
  - `ticket_matching`: candidate prefixes (`["AL", "GEN"]`), case-insensitive flag, custom regex patterns.
  - `database`: path, read-only flag, busy timeout (5000ms), WAL mode.
  - `live_observability`: active session threshold (300s), poll interval, max trace buffer, anomaly thresholds.
  - `roles`: mapping agent names/IDs to canonical roles (`orchestrator`, `coder`, `tester`, `reviewer`, `planner`).
  - `code_metrics`: include, exclude, test, and production file patterns.

### 6. Live Observability Platform (AL-17, AL-18, AL-19, AL-20)
- **Non-Locking Database Hook & Config-Driven Ticket Inference (AL-17)**:
  - Read-only SQLite hook with non-blocking pragmas: `PRAGMA query_only = ON;`, `PRAGMA busy_timeout = 5000;`.
  - Reconstructs multi-root session hierarchy with nested sub-sessions.
  - Config-driven case-insensitive ticket inference: detects upper, lower, and mixed-case ticket references (`al-4`, `GEN-2`, `feat/al-12-...`), matching against configured prefixes and normalizes to uppercase (`AL-4`).
  - HTTP endpoints: `/api/live/sessions` and SSE `/api/live/stream`.
- **Live Trace Inspector, Reasoning Stream, and Prominent Tool Failures (AL-18)**:
  - In-memory transient trace feed for active messages, LLM reasoning thoughts, and tool executions.
  - Tool failure classifier prioritizing failed calls with high-severity indicators and error snippets.
  - Transient memory buffer with zero disk writes to `agentic_metrics.json` or database tables.
  - HTTP endpoint: `/api/live/sessions/:id/trace`.
- **Domain Live Anomaly Detection Engine (AL-19)**:
  - Pure functional rule engine detecting consecutive tool failures (streak >= 3), token velocity spikes, context window ceiling saturation (200k), and tool looping; streamed via SSE.
- **TypeScript Live Observability Dashboard (AL-20)**:
  - Reactive `/live` view in web dashboard showing active root/sub-session tree, guessed ticket indicators, under-the-hood message and reasoning drawer, prominent failed tool call error cards, and live anomaly notification banners.

### 7. Rich Formatted Markdown for Workspace Artifacts (AL-22)
- Refactor `ArtifactViewer.tsx` in `web/src/components/` to replace raw `<pre><code>` code blocks with rich formatted Markdown rendering.
- Render styled headings, bordered tables (with formatted headers and cell alignment for metrics tables), styled task list checkboxes, blockquotes, and code snippets with syntax highlighting.
- Provide interactive toggle between "Rich Document" and "Raw Source" views, plus a "Copy Content" quick action.

## Non-Functional & Architectural Invariants
- **Non-Locking Database Access**: All queries to the active OpenCode SQLite DB must execute in read-only mode with zero locks, zero writes, and zero interference with running agents.
- **Zero Persistence for Live Observability**: Live traces, reasonings, and active messages remain transient in-memory and in frontend state. They are never written to disk or `agentic_metrics.json` until an iteration is officially concluded and generated.
- **FCIS Separation**: Pure domain transformations in Go `internal/domain` (tree derivation, ticket guessing heuristics, anomaly rule engine); imperative shell in `internal/adapters` (SQLite queries, SSE streaming, HTTP routes).
- **Strict Typing**: Zero `any` or empty `interface{}` across Go and TypeScript.
- **Zero Emojis**: Zero emojis in all code, comments, logs, UI elements, API responses, or documentation.

## Decomposition & Delivery Roadmap

### Tier 1 (Ready / Unblocked)
- [ ] Ticket AL-13: `AL-13: Implement Semantic Versioning and Automated CI for Pull Requests and Main Merges` (Unblocked)
- [ ] Ticket AL-15: `AL-15: Implement Universal Multi-Version Schema Migration Engine for Telemetry Ledgers` (Unblocked)
- [ ] Ticket AL-16: `AL-16: Execute Codebase and Skills Refactoring with Provided Review Skill` (Unblocked)
- [ ] Ticket AL-21: `AL-21: Migrate Configuration to YAML and Establish Agent-Backend Bridge Definitions` (Unblocked)
- [ ] Ticket AL-22: `AL-22: Render Rich Markdown for Workspace Summary and Contract Artifacts` (Unblocked)

### Tier 2 (Pending)
- [ ] Ticket AL-14: `AL-14: Implement Automated GitHub Release Workflows for Release Candidates and Stable Versions` (Blocked by AL-13)
- [ ] Ticket AL-17: `AL-17: Implement Non-Locking Live Database Hook, Session Tree Engine, and Config-Driven Ticket Inference` (Blocked by AL-21)

### Tier 3 (Pending)
- [ ] Ticket AL-18: `AL-18: Implement Live Trace Inspector, Reasoning Stream, and Prominent Tool Failure Telemetry` (Blocked by AL-17)

### Tier 4 (Pending)
- [ ] Ticket AL-19: `AL-19: Implement Live Anomaly Detection Engine and Real-Time Session Health Rules` (Blocked by AL-18)

### Tier 5 (Pending)
- [ ] Ticket AL-20: `AL-20: Build TypeScript Live Observability Dashboard and Multi-Root Session Inspector` (Blocked by AL-18, AL-19)
