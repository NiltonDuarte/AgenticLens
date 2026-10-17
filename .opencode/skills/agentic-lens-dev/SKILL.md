---
name: agentic-lens-dev
description: Use when building features, fixing bugs, or refactoring the standalone AgenticLens application (/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens) in Go (backend/CLI/engine) and TypeScript (frontend dashboard) with multi-agent orchestration (planner, coder, tester, reviewer). Requires a Jira ticket or task identifier.
---

# AgenticLens Multi-Agent Development Workflow

Coordinates autonomous development in the standalone **AgenticLens** project (`/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens`) across specialized subagents (`planner`, `coder`, `tester`, `reviewer`). Architecture is built in **Go** (core aggregations, boundary adapters, CLI, server) and **TypeScript** (modern web dashboard embedded into the Go binary via `embed.FS`). Persona instructions and layer boundaries live in `.opencode/agent/`.

The skill itself uses the AgenticLens CLI application to parse, validate, and render developer telemetry and lifecycle metrics for every development iteration.

## Configuration & Parametrization (`.agenticlens/config.json`)
At startup, inspect `.agenticlens/config.json` in the project root to load parameters (falling back to defaults if not present):
- `ticket_prefix`: Prefix for ticket identifiers (fallback default: `"GEN"`). If a numeric ID is passed (e.g. `42`), format as `<ticket_prefix>-<ID>` (e.g. `GEN-42` or `AL-42`).
- `workspaces_dir`: Directory for workspace artifacts (default: `"workspaces"`).
- `project_name`: Project name (default: `"AgenticLens"`).

Throughout this workflow:
- `<WORKSPACES_DIR>` refers to the configured `workspaces_dir` (e.g. `workspaces`).
- `<TICKET_ID>` uses the configured `ticket_prefix` (e.g. `AL-00042` or `GEN-00042`).

## Core Invariants
- **Orchestrator Coordinates Only**: Never write application code, edit manifests/schemas/views, or execute manual fixes directly.
- **Mandatory User Gate**: STOP after planning. Present `<WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md` and `iterations/<NN>/PLAN.md` to the user via the `question` tool. NEVER dispatch `coder` without explicit user confirmation.
- **Context Pointers Over Bloat**: Pass file pointers to subagents (`CONTRACT.md`, `iterations/<NN>/PLAN.md`, `git diff`), not long duplicated prompt instructions.
- **Self-Reflexive Telemetry via AgenticLens**: Use the configured telemetry command (`python3 scripts/extract_agentic_metrics.py ...`) to ingest the extracted fact ledger, validate schemas, calculate domain aggregations, update `SUMMARY.md`, and render development metrics.

## Setup & Workspace Structure
1. Require `<TICKET_ID>` (prefixed with configured `ticket_prefix`, e.g. `AL-XXXXX` or `GEN-XXXXX`).
2. Checkout ticket branch: `git checkout -b <type>/<TICKET_ID>-<slug>`.
3. Workspace layout under `<WORKSPACES_DIR>/<TICKET_ID>/`:
   - `CONTRACT.md`: Immutable baseline (problem statement, FCIS architectural invariants, Go/TS scope boundaries).
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
      prompt="Ticket: <TICKET_ID>. Request: <USER_REQUEST>. Config: .agenticlens/config.json. Ground against existing AgenticLens layers (Go backend in internal/domain, internal/adapters, cmd/ and TypeScript frontend in web/) and TECHNICAL_GUIDELINES.md. Write <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md and <WORKSPACES_DIR>/<TICKET_ID>/iterations/01/PLAN.md. Zero code blocks.",
  )
  ```
- **Iteration > 01**:
  ```python
  task(
      subagent_type="planner",
      description="Plan iteration delta",
      prompt="Ticket: <TICKET_ID>. Follow-up: <USER_REQUEST>. Config: .agenticlens/config.json. Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md and prior iterations. Write delta plan to <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md.",
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
       prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md, TECHNICAL_GUIDELINES.md, git diff. Execute pure Go domain tests (zero mocks) with go test -race ./..., boundary schema tests, and TypeScript test suites. Assert edge cases (zero tokens, division by zero, empty ledgers). Zero silent skips allowed.",
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
   # TypeScript verification and build (if frontend modified)
   test -d web && (cd web && npm run typecheck && npm run lint && npm run build) || true
   # Go verification and single binary build
   go vet ./... && go test -race ./... && go build -o bin/agentic-lens ./cmd/agentic-lens
   git diff --check
   ```
2. Update `<WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md`: Unified summary of everything developed for the ticket grouped by layer (`domain`, `adapters`, `presentation`), plus verification status.
3. Extract raw session facts ledger:
   ```bash
   NONCE=$(python3 -c "import uuid; print(uuid.uuid4().hex[:8])" || od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
   python3 scripts/extract_agentic_metrics.py --nonce $NONCE --ticket-id <TICKET_ID> --json-output <WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json
   ```
4. **Provide Metrics via AgenticLens Application**:
   Run AgenticLens Go CLI to parse the ledger, validate boundary schemas, derive domain metrics, update `SUMMARY.md`, and output KPIs:
   ```bash
   NONCE=$(python3 -c "import uuid; print(uuid.uuid4().hex[:8])" 2>/dev/null || od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
   python3 scripts/extract_agentic_metrics.py --nonce $NONCE --ticket-id <TICKET_ID> --json-output workspaces/<TICKET_ID>/agentic_metrics.json
   ```
5. Update `<WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`: Conventional PR draft (max 3000 chars, pointers to `CONTRACT.md` and `SUMMARY.md`).
6. Print concise handoff summary to user (active duration, cost, cache hits, test status, context window health).
