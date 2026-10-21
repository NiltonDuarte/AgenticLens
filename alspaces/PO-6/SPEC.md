# Feature Specification: Telemetry Event Pruning, Database Message Stripping, Dynamic Local Hydration, and Structural Hierarchy Integrity

## Problem Statement

AgenticLens currently faces two critical operational and architectural challenges across its telemetry collection, storage, and aggregation pipeline:

1. **Security Vulnerabilities and Repository Bloat from Raw Chronological Events**:
   - `AgenticLens` CLI dumps raw chronological event streams directly into `workspaces/<TICKET_ID>/agentic_metrics.json`.
   - In sessions with hundreds of tool invocations, this risks permanently leaking environment variables, API keys, credentials, local filesystem paths, and customer data into Git repository history.
   - Attempting regex-based secret redaction on database messages is inherently fragile and leak-prone. The committed Git plane must strictly enforce zero database message retention: no prompts, no assistant messages, and no raw database error strings are ever written to Git.
   - Routine successful tool calls (`read`, `edit`, `bash`) generate thousands of empty JSON shells (`status: "completed"`) that consume over 99% of ledger file size (~200 KB per session) with zero diagnostic value.
   - Failed tool calls in committed ledgers should be recorded as bare failure indicators without contextual database text; full context is reserved strictly for local hydration from the developer's SQLite database.

2. **Conflation of Root Session Hierarchy with Persona Role**:
   - The domain metrics engine and extraction adapters conflate structural hierarchy (`ParentID == nil` vs `ParentID != nil`) with agent persona names (`general`, `orchestrator`, `build`, etc.).
   - `domain.IsSubagentRole(role)` erroneously classifies any non-empty role other than `"orchestrator"` as a delegated subagent.
   - When a user interacts directly with OpenCode, the root session is assigned the persona `"general"`. Because `"general" != "orchestrator"`, the root session is misclassified as a subagent:
     - The root session's peak context tokens are misattributed to `SubagentsPeakTokens` with `SubagentsPeakRole = "general"`.
     - Root coordination metrics (`OrchestratorPeakTokens`, saturation percentages) drop to `0`.
     - Subagent effort distribution is diluted and distorted.
     - 8 test cases fail in `backend/internal/adapters/parity_test.go` against authentic workspace ledgers.

---

## Target Personas & Use Cases

- **AI Platform Engineers & Security Reviewers**: Need absolute security guarantees that no raw database text, prompts, or error messages ever leak into Git ledgers while maintaining ultra-compact (<5 KB) repository files.
- **Agent Developers & Operators**: Need full-fidelity interaction history (prompts, tool parameters, reasoning thoughts, raw error logs) when debugging locally on developer workstations via dynamic hydration, while retaining clean bare error timelines in remote/CI environments.
- **Engineering Leads & Managers**: Need trustworthy, accurate metrics where root session coordination and subagent effort distributions reflect actual structural delegation rather than brittle persona names.

---

## User Stories

1. As a security engineer, I want zero database messages, prompts, or error text committed to Git, so that credentials and sensitive customer data can never leak into repository history.
2. As a repository maintainer, I want routine successful tool executions and turn messages pruned from committed ledgers, so that `agentic_metrics.json` remains compact (<5 KB) with minimal diff churn.
3. As an agent developer, I want bare tool errors recorded in committed ledgers (tool name, turn, timestamp, status) without DB context, with full rich diagnostic details available on-demand via local database hydration.
4. As an operator inspecting a ticket locally, I want the web dashboard to dynamically hydrate full prompt, tool, and reasoning traces from my local OpenCode database, so that I have rich diagnostic observability without persisting massive logs to disk.
5. As a CI/headless runner, I want the web dashboard and API to gracefully fall back to the bare committed ledger when no local database exists, so that historical comparisons and health checks continue working portably.
6. As a telemetry analyst, I want root sessions (`ParentID == nil`) tracked as root coordination metrics regardless of persona name (`general`, `orchestrator`, `build`), so that subagent metrics measure only actual delegated child tasks.
7. As a quality engineer, I want `parity_test.go` and domain unit tests passing cleanly against authentic root session data, so that regression test suites are reliable.

---

## Functional Requirements

### 1. Structural Hierarchy Decoupling (Root Session vs Subagents)
- **Pure Domain Engine (`internal/domain`)**:
  - Determine whether an invocation is a root session strictly by checking `ParentID == nil` (or inclusion in `RootThreadIDs`), never by checking `Role == "orchestrator"`.
  - Invocations with `ParentID == nil` must never be counted or aggregated as subagents, regardless of role name (`general`, `orchestrator`, `build`, or custom).
  - Root coordination metrics (`OrchestratorPeakTokens`, `OrchestratorSaturationPct`) must evaluate root invocations (`ParentID == nil`), capturing peak context and saturation accurately.
  - Subagent metrics (`SubagentsPeakTokens`, `SubagentsPeakRole`, `subCosts`, `roleEffortDistributionPct`, and `SubagentTurnVelocity`) must strictly filter to child invocations (`ParentID != nil`).
  - Single-invocation sessions with no children must report `SubagentsPeakTokens: 0`, `SubagentsPeakRole: "none"`, and empty subagent effort distribution.
- **Adapters & Parity Baseline (`internal/adapters`)**:
  - Update `ExtractFacts` and `schemas.go` to preserve authentic root personas without forcing synthetic `"orchestrator"` remapping.
  - Update `backend/internal/adapters/parity_test.go` baselines and fixtures to reflect authentic `"general"` root sessions and decoupled subagent calculations.

### 2. Committed Git Plane Event Pruning, Message Stripping & Ledger Backfill
- **Zero Database Message Retention**:
  - Do not use regex sanitization; completely omit database messages, prompts, and raw error strings from committed `agentic_metrics.json`.
  - Failed tool calls persist only bare metadata: `tool_name`, `status: "error"`, turn number, timestamp, and duration (`error_message` is omitted or nil in committed ledgers).
  - Prune all routine completed turn messages and successful tool calls.
  - Retain strictly bare error events, compaction events, and events linked to detected domain anomalies.
- **In-Memory Aggregations**:
  - Compute all aggregate metrics (tokens, duration, cost, tool health, anomalies) in-memory over the full raw event stream before event pruning.
- **Agent Metadata Investigation**:
  - Investigate and document pathways to obtain structured diagnostic metadata directly from the agent runtime (e.g. self-reported exit codes, error classifications) versus relying strictly on local hydration.
- **Historical Ledger Backfill and Schema Migration**:
  - Register migration step `StepV2ToV3Sanitized` in `internal/domain/migration` to retroactively prune routine events and strip legacy database message strings across existing workspace ledgers.
  - Guarantee that top-level metrics, code facts, invocations, and anomalies remain 100% bit-identical after migration.
  - Support `agentic-lens migrate` with `--dry-run` and batch processing.

### 3. End-to-End Dynamic Local SQLite Hydration and Dual-Plane Waterfall
- **Backend Hydration Engine (`internal/adapters/http/handlers.go`)**:
  - Endpoint `GET /api/workspaces/:id/events`:
    - Checks whether ticket's session ID exists in the local OpenCode SQLite database (`opencode.db`).
    - Local DB available: uses read-only, non-locking queries (`PRAGMA query_only = ON;`, WAL mode) to extract full messages, reasoning thoughts, and tool inputs/outputs. Returns rich events with `is_hydrated: true` and `data_source: "local_sqlite"`.
    - Local DB unavailable (CI / remote): falls back to bare committed events from `workspaces/<ID>/agentic_metrics.json` with `is_hydrated: false` and `data_source: "committed_ledger"`.
  - Zero disk writes: all hydrated event data remains transient and strictly in-memory.
- **Frontend Dual-Plane Trace Waterfall (`web/src/components/TraceWaterfall.tsx`)**:
  - Intelligently renders based on `is_hydrated` flag:
    - `is_hydrated == true`: Full waterfall with expandable drawers for message prompts, LLM reasoning thought blocks, and formatted tool parameters/outputs.
    - `is_hydrated == false`: Compact bare error waterfall with a clear badge/banner ("Committed Git Ledger — Local Database Offline"), informing the user that full prompts, tool inputs, and error contexts are available only when running locally against the session database.
  - Detail drawer/modal for inspecting tool arguments and outputs adhering strictly to zero-emoji design.
  - Zero emojis, strict TypeScript types with zero `any`.

---

## Non-Functional & Architectural Invariants

1. **FCIS (Functional Core, Imperative Shell)**:
   - `internal/domain`: Pure aggregation and evaluation functions. Zero disk I/O, zero database queries, zero clocks.
   - `internal/adapters` and `cmd/`: Imperative shell for SQLite queries, filesystem operations, and HTTP routing.
2. **Zero Sensitive Context in Git**:
   - Prompts, assistant reasoning streams, tool inputs, and raw database error strings must never be written to `workspaces/` or committed to Git.
3. **Ledger Compactness**:
   - Persisted `agentic_metrics.json` must remain <5 KB for typical multi-turn sessions.
4. **Strict Typing & Zero Any**:
   - Zero `any` or `interface{}` across Go backend and TypeScript frontend.
5. **Zero Emojis**:
   - Zero emojis permitted in code, comments, docstrings, commit messages, PR drafts, UI strings, logs, or workspace artifacts.

---

## Decomposition & Delivery Roadmap

- [ ] **AL-31**: Decouple Root Session Structural Hierarchy from Persona Roles in Domain Engine and Tests *(Phase 6, Tier 1 - Ready, Unblocked)*
- [ ] **AL-32**: Telemetry Event Pruning, Database Message Stripping, Agent Data Investigation, and Ledger Backfill *(Phase 6, Tier 1 - Ready, Unblocked)*
- [ ] **AL-33**: End-to-End Dynamic Local Database Event Hydration and Dual-Plane Waterfall *(Phase 6, Tier 2 - Pending, Blocked by AL-32, AL-17)*
