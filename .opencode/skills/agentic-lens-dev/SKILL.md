---
name: agentic-lens-dev
description: Use when building features, fixing bugs, or refactoring the standalone AgenticLens application (/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens) in Go (backend/CLI/engine) and TypeScript (frontend dashboard) with multi-agent orchestration (planner, coder, tester, reviewer). Requires a Jira ticket or task identifier.
---

# AgenticLens Multi-Agent Development Workflow

Coordinates autonomous development in the standalone **AgenticLens** project (`/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens`) across specialized subagents (`planner`, `coder`, `tester`, `reviewer`). Architecture is built in **Go** (core aggregations, boundary adapters, CLI, server) and **TypeScript** (modern web dashboard embedded into the Go binary via `embed.FS`). Persona instructions and layer boundaries live in `.opencode/agent/`.

The skill itself uses the AgenticLens CLI application to parse, validate, and render developer telemetry and lifecycle metrics for every development iteration.

## Configuration & Parametrization (`.agenticlens/config.yaml`)
At startup, inspect `.agenticlens/config.yaml` in the project root to load parameters (falling back to defaults if not present):
- `ticket.prefixes`: Supported prefix identifiers and descriptions (defaults: `AL`, `GEN`, `PO`). If a numeric ID is passed (e.g. `42`), format as `<ticket_prefix>-<ID>` (e.g. `AL-42`, `GEN-42`, or `PO-2`).
- `workspaces.dir`: Directory for workspace artifacts (default: `"workspaces"`).
- `project_name`: Project name (default: `"AgenticLens"`).

Throughout this workflow:
- `<WORKSPACES_DIR>` refers to the configured `workspaces.dir` (e.g. `workspaces`).
- `<TICKET_ID>` uses a configured ticket prefix (e.g. `AL-42`, `GEN-42`, or `PO-2`).

## Core Invariants
- **Single-Ticket Exclusivity**: Never work on more than one ticket at a time. Refuse any switch to another ticket. Treat all subsequent prompts as steering for the active ticket, creating new iterations as needed.
- **Orchestrator Coordinates Only**: Never write application code, edit manifests/schemas/views, or execute manual fixes directly.
- **Mandatory User Gate**: STOP after planning. Present `<WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md` and `iterations/<NN>/PLAN.md` to the user via the `question` tool. NEVER dispatch `coder` without explicit user confirmation.
- **Context Pointers Over Bloat**: Pass file pointers to subagents (`CONTRACT.md`, `iterations/<NN>/PLAN.md`, `git diff`), not long duplicated prompt instructions.
- **Self-Reflexive Telemetry via AgenticLens**: Use the native AgenticLens CLI command (`./bin/agentic-lens generate ...`) to ingest the extracted fact ledger, validate schemas, calculate domain aggregations, update `SUMMARY.md`, and render development metrics.

## Setup & Workspace Structure
1. Require `<TICKET_ID>` (prefixed with configured ticket prefix, e.g. `AL-#`, `GEN-#`, or `PO-#`). Pick up specification from `<WORKSPACES_DIR>/plan/to-do/<TICKET_ID>*.md` and update ticket status to `In Progress` in `<WORKSPACES_DIR>/plan/TICKETS.md`.
2. Checkout ticket branch: `git checkout -b <type>/<TICKET_ID>-<slug>`.
3. Workspace layout under `<WORKSPACES_DIR>/<TICKET_ID>/`:
   - `CONTRACT.md`: Immutable baseline authored by the developer (problem statement, FCIS architectural invariants, Go/TS scope boundaries).
   - `SUMMARY.md` & `agentic_metrics.json`: Unified summary of everything developed for the ticket grouped by layer (`domain`, `adapters`, `presentation`), plus AgenticLens-rendered metrics table.
   - `PULL_REQUEST.md`: Conventional PR draft (max 3000 chars, pointers to `CONTRACT.md` & `SUMMARY.md`, summary of changes; no test evidence).
   - `iterations/<NN>/`: `PLAN.md` and `REVIEW.md` per iteration.

---

## Iteration Lifecycle

### Phase 1: Grounding, Planning & Steering
- **Iteration 01**:
  ```python
  task(
      subagent_type="planner",
      description="Ground, contract & plan iteration 01",
      prompt="Ticket: <TICKET_ID>. Request: <USER_REQUEST>. Config: .agenticlens/config.yaml. Ground against existing AgenticLens layers (Go backend in internal/domain, internal/adapters, cmd/ and TypeScript frontend in web/) and TECHNICAL_GUIDELINES.md. Write <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md and <WORKSPACES_DIR>/<TICKET_ID>/iterations/01/PLAN.md. Zero code blocks.",
  )
  ```
- **Iteration > 01**:
  ```python
  task(
      subagent_type="planner",
      description="Plan iteration delta",
      prompt="Ticket: <TICKET_ID>. Follow-up: <USER_REQUEST>. Config: .agenticlens/config.yaml. Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md and prior iterations. Write delta plan to <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md.",
  )
  ```

### User Gate (HARD STOP)
Present planning artifacts. Wait for explicit user confirmation via `question` tool before touching code. If changes requested, re-dispatch `planner`.

### Phase 2: Implementation, Testing & Audit Loop
Run `coder` -> `tester` -> `reviewer` on `<WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/`:

1. **Code (`coder`)**:
   ```python
   task(
       subagent_type="coder",
       description="Implement iteration <NN>",
       prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md (and REVIEW.md if rework), TECHNICAL_GUIDELINES.md. Implement Go domain models/aggregations, Go adapters/CLI, or TypeScript frontend components adhering to FCIS, static typing, zero any in Go/TS, and zero emojis. Return STATUS: BLOCKED if conflicting constraints arise.",
   )
   ```
2. **Test (`tester`)**:
   ```python
   task(
       subagent_type="tester",
       description="Test iteration <NN>",
        prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md, TECHNICAL_GUIDELINES.md, git diff. Execute pure Go domain tests (zero mocks) with make test (or go test -race ./...), boundary schema tests, and TypeScript test suites. Assert edge cases (zero tokens, division by zero, empty ledgers). Zero silent skips allowed.",
   )
   ```
3. **Audit (`reviewer`)**:
   ```python
   task(
       subagent_type="reviewer",
       description="Audit iteration <NN>",
       prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md, TECHNICAL_GUIDELINES.md, git diff. Audit along Axis 1 (Spec & Contract Alignment) and Axis 2 (Engineering Standards in TECHNICAL_GUIDELINES.md: Go FCIS isolation, TypeScript strictness, zero any, zero emojis). Save verdict to <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/REVIEW.md.",
   )
   ```

**Loop Control**:
- `APPROVED`: Proceed to Phase 3.
- `CHANGES_REQUESTED`: Feed findings back to `coder`/`tester`.
- `STATUS: BLOCKED`: Re-engage `planner` with details, update `PLAN.md`, and re-confirm with user.

### Phase 3: Delivery Sync (Every Iteration)
1. Verify code quality:
   ```bash
   # Run automated checks, tests, and build via Makefile
   make
   # or run individually: make check && make test && make build
   ```
2. Update `<WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md`: Unified summary of everything developed for the ticket grouped by layer (`domain`, `adapters`, `presentation`), plus verification status.
3. **Extract Telemetry & Render Metrics via AgenticLens**:
   Run AgenticLens Go CLI to query OpenCode SQLite and Git, construct the fact ledger, validate boundary schemas, derive domain metrics, update `SUMMARY.md`, and output KPIs. Pass explicit `--ticket-type` matching the conventional commit classification (`feat`, `fix`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`):
   ```bash
   NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
   ./bin/agentic-lens generate --nonce $NONCE --ticket-id <TICKET_ID> --ticket-type <CONVENTIONAL_TYPE> --json-output <WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json --append <WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md
   ```
4. Update `<WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`: Conventional PR draft (max 3000 chars, pointers to `CONTRACT.md` and `SUMMARY.md`).
5. Print concise handoff summary to user (active duration, cost, cache hits, test status, context window health).
