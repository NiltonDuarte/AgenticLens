---
name: agentic-lens-dev
description: Use when building features, fixing bugs, or refactoring the standalone AgenticLens application (/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens) with multi-agent orchestration (planner, coder, tester, reviewer). Requires a Jira ticket or task identifier.
---

# AgenticLens Multi-Agent Development Workflow

Coordinates autonomous development in the standalone **AgenticLens** project (`/Users/guedesduarte/workspace/projects/SinkLayer/AgenticLens`) across specialized subagents (`planner`, `coder`, `tester`, `reviewer`). Persona instructions and layer boundaries live in `.opencode/agent/`.

The skill itself uses the AgenticLens application to parse, validate, and render developer telemetry and lifecycle metrics for every development iteration.

## Core Invariants
- **Orchestrator Coordinates Only**: Never write application code, edit schemas/views, or execute manual fixes directly.
- **Mandatory User Gate**: STOP after planning. Present `CONTRACT.md` and `iterations/<NN>/PLAN.md` to the user via the `question` tool. NEVER dispatch `coder` without explicit user confirmation.
- **Context Pointers Over Bloat**: Pass file pointers to subagents (`CONTRACT.md`, `iterations/<NN>/PLAN.md`, `git diff`), not long duplicated prompt instructions.
- **Self-Reflexive Telemetry via AgenticLens**: Use the `agentic_lens` application itself (`uv run python -m agentic_lens ...`) to ingest the extracted fact ledger, validate schemas, calculate domain aggregations, update `SUMMARY.md`, and render development metrics.

## Setup & Workspace Structure
1. Require `<TICKET_ID>`.
2. Checkout ticket branch: `git checkout -b <type>/<TICKET_ID>-<slug>`.
3. Workspace layout under `workspaces/<TICKET_ID>/`:
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
      prompt="Ticket: <TICKET_ID>. Request: <USER_REQUEST>. Ground against existing AgenticLens layers (domain, adapters, presentation). Write workspaces/<TICKET_ID>/CONTRACT.md and iterations/01/PLAN.md. Zero code blocks.",
  )
  ```
- **Iteration > 01**:
  ```python
  task(
      subagent_type="planner",
      description="Plan iteration delta",
      prompt="Ticket: <TICKET_ID>. Follow-up: <USER_REQUEST>. Context: workspaces/<TICKET_ID>/CONTRACT.md and prior iterations. Write delta plan to workspaces/<TICKET_ID>/iterations/<NN>/PLAN.md.",
  )
  ```

### User Gate (HARD STOP)
Present planning artifacts. Wait for explicit user confirmation via `question` tool before touching code. If changes requested, re-dispatch `planner`.

### Phase 2: Implementation, Testing & Audit Loop
Run `coder` -> `tester` -> `reviewer` on `workspaces/<TICKET_ID>/iterations/<NN>/`:

1. **Code (`coder`)**:
   ```python
   task(
       subagent_type="coder",
       description="Implement iteration <NN>",
       prompt="Context: workspaces/<TICKET_ID>/CONTRACT.md, iterations/<NN>/PLAN.md (and REVIEW.md if rework). Implement domain models, aggregations, adapters, or presentation views adhering to FCIS, static typing, zero Any, and zero emojis. Return STATUS: BLOCKED if conflicting constraints arise.",
   )
   ```
2. **Test (`tester`)**:
   ```python
   task(
       subagent_type="tester",
       description="Test iteration <NN>",
       prompt="Context: workspaces/<TICKET_ID>/CONTRACT.md, iterations/<NN>/PLAN.md, git diff. Execute pure domain tests (zero mocks), boundary schema tests, and repository tests with uv run pytest. Assert edge cases (zero tokens, division by zero, empty ledgers). Zero silent skips allowed.",
   )
   ```
3. **Audit (`reviewer`)**:
   ```python
   task(
       subagent_type="reviewer",
       description="Audit iteration <NN>",
       prompt="Context: workspaces/<TICKET_ID>/CONTRACT.md, iterations/<NN>/PLAN.md, git diff. Audit along Axis 1 (Spec & Contract Alignment) and Axis 2 (Engineering Standards: FCIS isolation, zero Any, zero emojis, Streamlit width='stretch'). Save verdict to workspaces/<TICKET_ID>/iterations/<NN>/REVIEW.md.",
   )
   ```

**Loop Control**:
- `APPROVED`: Proceed to Phase 3.
- `CHANGES_REQUESTED`: Feed findings back to `coder`/`tester`.
- `STATUS: BLOCKED`: Re-engage `planner` with details, update `PLAN.md`, and re-confirm with user.

### Phase 3: Delivery Sync (Every Iteration)
1. Verify code quality:
   ```bash
   python3 -m py_compile agentic_lens/**/*.py && uv run ruff check agentic_lens/ && git diff --check
   ```
2. Update `workspaces/<TICKET_ID>/SUMMARY.md`: Unified summary of everything developed for the ticket grouped by layer (`domain`, `adapters`, `presentation`), plus verification status.
3. Extract raw session facts ledger:
   ```bash
   NONCE=$(python3 -c "import uuid; print(uuid.uuid4().hex[:8])")
   python3 scripts/extract_agentic_metrics.py --nonce $NONCE --ticket-id <TICKET_ID> --json-output workspaces/<TICKET_ID>/agentic_metrics.json
   ```
4. **Provide Metrics via AgenticLens Application**:
   Run AgenticLens to parse the ledger, validate boundary schemas, derive domain metrics, update `SUMMARY.md`, and output KPIs:
   ```bash
   uv run python -m agentic_lens --ticket-id <TICKET_ID> --append workspaces/<TICKET_ID>/SUMMARY.md
   ```
5. Update `workspaces/<TICKET_ID>/PULL_REQUEST.md`: Conventional PR draft (max 3000 chars, pointers to `CONTRACT.md` and `SUMMARY.md`).
6. Print concise handoff summary to user (active duration, cost, cache hits, test status, context window health) and offer to launch the dashboard:
   ```bash
   uv run streamlit run agentic_lens/presentation/app.py
   ```
