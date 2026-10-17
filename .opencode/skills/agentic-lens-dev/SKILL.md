---
name: agentic-lens-dev
description: Use when building features, fixing bugs, or refactoring the standalone AgenticLens application (/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens) with multi-agent orchestration (planner, coder, tester, reviewer). Requires a Jira ticket or task identifier.
---

# AgenticLens Multi-Agent Development Workflow

Coordinates autonomous development in the standalone **AgenticLens** project (`/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens`) across specialized subagents (`planner`, `coder`, `tester`, `reviewer`). Persona instructions and layer boundaries live in `.opencode/agent/`.

The skill itself uses the AgenticLens application to parse, validate, and render developer telemetry and lifecycle metrics for every development iteration.

## Configuration & Parametrization (`.agenticlens/config.json`)
At startup, inspect `.agenticlens/config.json` in the project root to load parameters (falling back to defaults if not present):
- `ticket_prefix`: Prefix for ticket identifiers (default: `"AL"`). If a numeric ID is passed (e.g. `42`), format as `<ticket_prefix>-<ID>` (e.g. `AL-42`).
- `workspaces_dir`: Directory for workspace artifacts (default: `"workspaces"`).
- `project_name`: Project name (default: `"AgenticLens"`).
- `guidelines_file`: Technical guidelines document (default: `"TECHNICAL_GUIDELINES.md"`).
- `telemetry_command`: Command to invoke AgenticLens telemetry CLI (default: `"uv run python -m agentic_lens"`).

Throughout this workflow:
- `<WORKSPACES_DIR>` refers to the configured `workspaces_dir` (e.g. `workspaces`).
- `<TICKET_ID>` uses the configured `ticket_prefix` (e.g. `AL-00042`).
- `<GUIDELINES>` refers to the configured `guidelines_file`.

## Core Invariants
- **Orchestrator Coordinates Only**: Never write application code, edit schemas/views, or execute manual fixes directly.
- **Mandatory User Gate**: STOP after planning. Present `<WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md` and `iterations/<NN>/PLAN.md` to the user via the `question` tool. NEVER dispatch `coder` without explicit user confirmation.
- **Context Pointers Over Bloat**: Pass file pointers to subagents (`CONTRACT.md`, `iterations/<NN>/PLAN.md`, `git diff`), not long duplicated prompt instructions.
- **Self-Reflexive Telemetry via AgenticLens**: Use the configured telemetry command (`uv run python -m agentic_lens ...`) to ingest the extracted fact ledger, validate schemas, calculate domain aggregations, update `SUMMARY.md`, and render development metrics.

## Setup & Workspace Structure
1. Require `<TICKET_ID>` (prefixed with configured `ticket_prefix`, e.g. `AL-XXXXX`).
2. Checkout ticket branch: `git checkout -b <type>/<TICKET_ID>-<slug>`.
3. Workspace layout under `<WORKSPACES_DIR>/<TICKET_ID>/`:
   - `CONTRACT.md`: Immutable baseline (problem statement, FCIS architectural invariants, scope boundaries).
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
      prompt="Ticket: <TICKET_ID>. Request: <USER_REQUEST>. Config: .agenticlens/config.json. Ground against existing AgenticLens layers (domain, adapters, presentation) and <GUIDELINES>. Write <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md and <WORKSPACES_DIR>/<TICKET_ID>/iterations/01/PLAN.md. Zero code blocks.",
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
       prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md (and REVIEW.md if rework), <GUIDELINES>. Implement domain models, aggregations, adapters, or presentation views adhering to FCIS, static typing, zero Any, and zero emojis. Return STATUS: BLOCKED if conflicting constraints arise.",
   )
   ```
2. **Test (`tester`)**:
   ```python
   task(
       subagent_type="tester",
       description="Test iteration <NN>",
       prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md, <GUIDELINES>, git diff. Execute pure domain tests (zero mocks), boundary schema tests, and repository tests with uv run pytest. Assert edge cases (zero tokens, division by zero, empty ledgers). Zero silent skips allowed.",
   )
   ```
3. **Audit (`reviewer`)**:
   ```python
   task(
       subagent_type="reviewer",
       description="Audit iteration <NN>",
       prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md, <GUIDELINES>, git diff. Audit along Axis 1 (Spec & Contract Alignment) and Axis 2 (Engineering Standards in <GUIDELINES>: FCIS isolation, zero Any, zero emojis, Streamlit width='stretch'). Save verdict to <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/REVIEW.md.",
   )
   ```

**Loop Control**:
- `APPROVED`: Proceed to Phase 3.
- `CHANGES_REQUESTED`: Feed findings back to `coder`/`tester`.
- `STATUS: BLOCKED`: Re-engage `planner` with details, update `PLAN.md`, and re-confirm with user.

### Phase 3: Delivery Sync (Every Iteration)
1. Verify code quality:
   ```bash
   python3 -m py_compile agentic_lens/**/*.py && uv run ruff check agentic_lens/ tests/ && git diff --check
   ```
2. Update `<WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md`: Unified summary of everything developed for the ticket grouped by layer (`domain`, `adapters`, `presentation`), plus verification status.
3. Extract raw session facts ledger:
   ```bash
   NONCE=$(python3 -c "import uuid; print(uuid.uuid4().hex[:8])")
   python3 scripts/extract_agentic_metrics.py --nonce $NONCE --ticket-id <TICKET_ID> --json-output <WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json
   ```
4. **Provide Metrics via AgenticLens Application**:
   Run AgenticLens to parse the ledger, validate boundary schemas, derive domain metrics, update `SUMMARY.md`, and output KPIs:
   ```bash
   uv run python -m agentic_lens --ticket-id <TICKET_ID> --append <WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md
   ```
5. Update `workspaces/<TICKET_ID>/PULL_REQUEST.md`: Conventional PR draft (max 3000 chars, pointers to `CONTRACT.md` and `SUMMARY.md`).
6. Print concise handoff summary to user (active duration, cost, cache hits, test status, context window health) and offer to launch the dashboard:
   ```bash
   uv run streamlit run agentic_lens/presentation/app.py
   ```
