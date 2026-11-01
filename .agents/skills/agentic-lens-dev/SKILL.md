---
name: agentic-lens-dev
description: Use when building features, fixing bugs, or refactoring the standalone AgenticLens application in Go (backend/CLI/engine) and TypeScript (frontend dashboard) with multi-agent orchestration (planner, coder, tester, reviewer). Requires a Jira ticket or task identifier.
---

# AgenticLens Multi-Agent Development Workflow

Coordinates autonomous development in the standalone **AgenticLens** project across specialized subagents (`planner`, `coder`, `tester`, `reviewer`). Architecture is built in **Go** (core aggregations, boundary adapters, CLI, server) and **TypeScript** (modern web dashboard embedded into the Go binary via `embed.FS`). Persona instructions and layer boundaries live in `.agents/agent/`.

The skill itself uses the AgenticLens CLI application to parse, validate, and render developer telemetry and lifecycle metrics for every development iteration.

## Configuration & Parametrization (`.agenticlens-config.yaml`)
At startup, inspect `.agenticlens-config.yaml` in the project root to load parameters (falling back to defaults if not present):
- `ticket.prefixes`: Supported prefix identifiers and descriptions (defaults: `AL`, `GEN`, `PO`). If a numeric ID is passed (e.g. `42`), format as `<ticket_prefix>-<ID>` (e.g. `AL-42`, `GEN-42`, or `PO-2`).
- `workspaces.dir`: Directory for workspace artifacts (default: `"workspaces"`).
- `project_name`: Project name (default: `"AgenticLens"`).

Throughout this workflow:
- `<WORKSPACES_DIR>` refers to the configured `workspaces.dir` (e.g. `workspaces`).
- `<TICKET_ID>` uses a configured ticket prefix (e.g. `AL-42`, `GEN-42`, or `PO-2`).

Follow the repository [testing policy](../../../TECHNICAL_GUIDELINES.md#test-purpose-and-proportionality) and [test review checklist](../../../docs/REGRESSION_PREVENTION.md#adding-reviewing-or-reducing-tests) when planning additions, consolidation, or removal.

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
      prompt="Ticket: <TICKET_ID>. Request: <USER_REQUEST>. Config: .agenticlens-config.yaml. Ground against existing AgenticLens layers (Go backend in backend/internal/domain, backend/internal/adapters, backend/cmd/ and TypeScript frontend in web/) and TECHNICAL_GUIDELINES.md. Write <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md and <WORKSPACES_DIR>/<TICKET_ID>/iterations/01/PLAN.md. Zero code blocks.",
  )
  ```
- **Iteration > 01**:
  ```python
  task(
      subagent_type="planner",
      description="Plan iteration delta",
      prompt="Ticket: <TICKET_ID>. Follow-up: <USER_REQUEST>. Config: .agenticlens-config.yaml. Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md and prior iterations. Write delta plan to <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md.",
  )
  ```

### User Gate (HARD STOP)
Present planning artifacts. Wait for explicit user confirmation via `question` tool before touching code. If changes requested, re-dispatch `planner`.

### Phase 2: Implementation, Testing & Audit Loop
Give the reviewer a fresh context where supported: requirements, baseline/head, raw diff, canonical review procedure and evidence locations; avoid inheriting the author's conclusions. Scale parallel reviewers to independent risks rather than a fixed agent count. For a disputed finding, verify its concrete counterexample rather than vote on approval.

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
        prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md, TECHNICAL_GUIDELINES.md, git diff. Execute pure Go domain tests (zero mocks) with make test (or go test -race ./...), boundary schema tests, and TypeScript test suites. Inspect existing coverage and verify only relevant missing contract cases under the canonical testing policy; extend coherent existing tests before adding declarations. For reductions, verify deletion rationales and assertion-to-recipient transfers. Zero silent skips allowed.",
   )
   ```
3. **Audit (`reviewer`)**:
   ```python
   task(
       subagent_type="reviewer",
       description="Audit iteration <NN>",
       prompt="Context: <WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md, <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md, TECHNICAL_GUIDELINES.md, git diff. Follow .agents/skills/agentic-lens-review/SKILL.md for independent correctness investigation, including relevant unchanged callers and risk-directed evidence. Form failure hypotheses from requirements/source before reconciling implementation reports. Save verdict to <WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/REVIEW.md.",
   )
   ```

**Loop Control**:
- `APPROVED`: Proceed to Phase 3 for the reviewed revision and scope; this does not certify unrelated code or authorize merging.
- `INCOMPLETE`: Resolve the named material evidence/access gap before delivery; do not treat missing findings as approval.
- `CHANGES_REQUESTED`: Feed findings back to `coder`/`tester`.
- `STATUS: BLOCKED`: Re-engage `planner` with details, update `PLAN.md`, and re-confirm with user.

### Phase 3: Delivery Sync (Every Iteration)
1. **Commit Implementation**:
   Commit all implementation files and test changes on the feature branch first to ensure a clean working tree:
   ```bash
   git add -A
   git commit -m "<type>(<scope>): <summary> (<TICKET_ID>)"
   ```
2. **Sync Upstream & Verify Code Quality**:
   Merge `main` into the feature branch with a clean working tree to resolve any conflicts, then run automated checks:
   ```bash
   git fetch origin main && git merge origin/main --no-edit
   # Run automated checks, tests, and build via Makefile
   make
   # or run individually: make check && make test && make build
   ```
3. Update `<WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md`: Unified summary of everything developed for the ticket grouped by layer (`domain`, `adapters`, `presentation`), plus verification status.
4. Update `<WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`: Conventional PR draft with the first line formatted as `# <type>(<scope>): <summary> (<TICKET_ID>)` (max 3000 chars, pointers to `CONTRACT.md` and `SUMMARY.md`).
5. **Plan Lifecycle Synchronization**:
   Move specification file from `<WORKSPACES_DIR>/plan/to-do/<TICKET_ID>*.md` to `<WORKSPACES_DIR>/plan/done/`, remove the completed ticket entry from `<WORKSPACES_DIR>/plan/TICKETS.md` (do not create or maintain a Done section; completed specifications remain in `plan/done/`), and unblock downstream tickets: check tickets blocked by `<TICKET_ID>` and promote any whose dependencies are now fully satisfied to `## Ready` with `Status: Ready`.
6. **Extract Telemetry & Render Metrics (Final Step Before PR Commit)**:
   Delayed strictly to the final step before committing so that fewer turns, tool calls, compactions, or subagent sessions are omitted.
   Generate a unique random nonce directly (e.g. an 8 to 16 character alphanumeric or hex string) and supply the literal token directly to `--nonce <GENERATED_NONCE>`. Shell variables (`$NONCE`) must not be used because OpenCode records tool inputs into SQLite before subshell expansion occurs.
   Run AgenticLens Go CLI to query OpenCode SQLite and Git, construct the fact ledger, validate boundary schemas, derive domain metrics, update `SUMMARY.md`, and output KPIs. Pass explicit `--ticket-type` matching the conventional commit classification (`feat`, `fix`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`):
   ```bash
   ./bin/agentic-lens generate --nonce <GENERATED_NONCE> --ticket-id <TICKET_ID> --ticket-type <CONVENTIONAL_TYPE> --json-output <WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json --append <WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md
   ```
7. **Commit, Push & Open Pull Request**:
   Extract title from the first line of `PULL_REQUEST.md`, commit final workspace and telemetry files, push upstream, and open a GitHub Pull Request:
   ```bash
   PR_TITLE=$(head -n 1 <WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md | sed 's/^# *//')
   git add -A
   git commit -m "chore(telemetry): update workspace metrics and summary ($PR_TITLE)" # or git commit --amend
   git push -u origin <branch>
   PR_URL=$(gh pr create --title "$PR_TITLE" --body-file <WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md)
   ```
   If running in an environment with pull request linking tools (e.g. `link_pull_request`), register the pull request URL immediately.
8. **Print Handoff Summary & Pull Request Link**:
   Print concise handoff summary to user, with the Pull Request link prominently displayed at the top:
   - **Pull Request**: `<PR_URL>`
   - Active duration, cost, cache hits, test status, context window health.
