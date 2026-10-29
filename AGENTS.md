# Agent Instructions for AgenticLens

Mandatory operational policies for all agents and contributors working in this repository.

---

## 1. Ticketed and Ad Hoc Work

Work may be ticketed or ad hoc. Do not create a ticket, ticket specification, or ticket workspace solely to satisfy this policy. When a task is associated with an existing ticket, follow that ticket's lifecycle and workspace requirements below.

### Configurable Workspace Root
- Define `<WORKSPACES_DIR>` as the `workspaces.dir` value in `.agenticlens-config.yaml`.
- If `.agenticlens-config.yaml` or the `workspaces.dir` key is absent, use `workspaces` as the fallback.
- The current configuration sets `workspaces.dir` to `alspaces`, so `<WORKSPACES_DIR>` currently resolves to `alspaces`.
- Use `<WORKSPACES_DIR>` as the root for all ticket workspaces and plan paths below.

### Rules
1. **Existing Ticket or Ad Hoc Work**:
   - If the request identifies an existing ticket, use that ticket ID. Do not create a duplicate ticket or specification.
   - If no ticket is identified and the task is ad hoc, proceed without creating a ticket ID, ticket specification, or ticket workspace. Ask for clarification only when ticket ownership materially affects scope or dependencies.
   - Create or register a ticket only when the user explicitly requests ticket creation or the work must enter the project backlog.
2. **Single-Ticket Exclusivity**:
   - Agents are strictly prohibited from working on more than one ticket at a time.
   - Once work begins on a ticket, the agent is locked to that ticket. Agents must refuse any request to switch to another ticket or work on multiple tickets simultaneously within the same session. Ad hoc work has no ticket lock, but must remain a single scoped task.
3. **Prompt Steering & Iterations**:
   - Once a ticket is active, every subsequent user prompt must be interpreted as steering feedback for that specific ticket.
   - Agents must create and advance new iterations (`iterations/<NN>/`) as needed to incorporate steering directions.
4. **Dedicated Workspace Allocation**:
   - Ticketed work must have its isolated workspace under `<WORKSPACES_DIR>/<TICKET_ID>/`. Ad hoc work does not require a ticket workspace.
   - Never mix artifacts across tickets or leave untracked changes in the working directory.
5. **Workspace Deliverables Checklist**:
   - The following workspace deliverables apply to ticketed work; ad hoc tasks do not need a ticket workspace or ticket-specific artifacts.
   - `<WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md`: Summary log of work completed, changes made, and verification status.
   - `<WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`: Pull request description and context.
   - `<WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md`: Frozen technical contract and acceptance boundaries, when required by the configured artifact list.
   - **Plan Lifecycle Synchronization**: When a matching specification exists in `<WORKSPACES_DIR>/plan/to-do/<TICKET_ID>*.md`, move it to `<WORKSPACES_DIR>/plan/done/` when the ticket is completed. Remove the completed ticket entry from `<WORKSPACES_DIR>/plan/TICKETS.md` only if one exists; do not create a specification, ticket entry, or Done section just to satisfy this lifecycle rule. For ticketed work, unblock downstream tickets whose dependencies are now fully satisfied by promoting them to `## Ready` with `Status: Ready`.
   - `<WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json`: Raw session telemetry ledger (generated at the very final step of the iteration).
   - Store iteration plans and reviews under `<WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md` and `<WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/REVIEW.md`.
6. **Telemetry Ingestion as the Final Step & Conventional Commit Type Classification**:
   - **Delayed Ingestion Mandate**: Generating telemetry must always be delayed to the very last step of the workflow. Running telemetry early causes subsequent turns, tool calls, compactions, and subagent sessions to be omitted from the ledger.
   - **Literal Nonce Generation**: The agent must generate a unique random nonce directly (by any means, e.g. an 8 to 16 character alphanumeric or hex string) and supply the literal token directly in the command. Shell variables (`$NONCE`) or command substitutions (`$(...)`) must not be used because OpenCode records tool inputs into SQLite before subshell expansion occurs, which causes database searches for the expanded token to fail.
   - For ticketed work, execute AgenticLens CLI with an explicit conventional commit `type` immediately before final commit:
     ```bash
     ./bin/agentic-lens generate --nonce <GENERATED_NONCE> --ticket-id <TICKET_ID> --ticket-type <CONVENTIONAL_TYPE> --json-output <WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json --append <WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md
     ```
   - `<CONVENTIONAL_TYPE>` must match standard conventional commit classifications (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).
   - PO tickets (`PO-#`) must strictly use `--ticket-type docs`.
   - Generic tickets (`GEN-#`) must use `--ticket-type chore` unless explicitly a feature (`feat`).
   - Developer tickets must use the corresponding conventional commit type matching their branch prefix (`feat`, `fix`, `refactor`, etc.). Never default all tickets to `feature`.
   - Do not create a ticket or ticket workspace solely to generate telemetry for ad hoc work. Generate ticket-specific telemetry only when the task has an associated ticket or the user explicitly requests a ticketing strategy.
7. **Pull Request Delivery & Handoff**:
   - Commit all implementation deliverables on the feature branch first (`git add -A && git commit -m "$COMMIT_MSG"`). Git requires a clean working tree before merging.
   - Merge `main` into the feature branch (`git fetch origin main && git merge origin/main --no-edit`) to resolve any upstream conflicts, and verify test suites pass (`make`).
   - Extract the PR title directly from the first line of `PULL_REQUEST.md` (e.g. `PR_TITLE=$(head -n 1 <WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md | sed 's/^# *//')`).
   - Commit final deliverables and telemetry facts on the feature branch, push upstream, and open a GitHub Pull Request via `gh pr create --title "$PR_TITLE" --body-file <WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`.
   - Avoid directly merging or checking out `main` within agent worktrees.
   - The Pull Request URL must be prominently displayed at the top of the final handoff report.

---

## 2. Multi-Agent SDLC & Role Boundaries

Work in this repository is orchestrated through specialized agents and skills:

- **Product / Scope Inception (`agentic-lens-po`)**:
  - Elicits requirements and decomposes features into tracer-bullet tickets.
  - Used when creating new tickets, structuring specifications, or defining acceptance criteria.
  - **PO Non-Execution Invariant**: Project Owners are strictly prohibited from writing or modifying application code, test files, runtime scripts, or adapters. Project Owners only plan tasks, author product specifications (`SPEC.md`), and decompose roadmaps into Jira tickets. All implementation must be performed through `agentic-lens-dev`.
  - **PO Session Separation**: The PO session strictly ends once specifications and planning tickets are authored. Development must never take place in the same session and must always run in a separate, dedicated session.
- **Development Orchestrator (`orchestrator` / `agentic-lens-dev`)**:
  - Coordinates iterations across subagents (`planner` -> `coder` -> `tester` -> `reviewer`).
  - **Orchestrator Non-Execution Invariant**: Orchestrators never write application code directly.
  - **Mandatory User Confirmation Gate**: Stop and confirm proposed plan with the user via `question` before writing code.
- **Subagents**:
  - `planner`: Contracts, architecture, and zero-code plans.
  - `coder`: Go backend and TypeScript frontend implementation.
  - `tester`: Adversarial unit and boundary tests. Zero mocks in domain, zero silent skips.
  - `reviewer`: Read-only diff audit along Spec Alignment and Engineering Standards.

---

## 3. Architecture & Code Invariants

All code must strictly satisfy `TECHNICAL_GUIDELINES.md`:

1. **Stack**:
   - Go 1.23+ for backend, CLI, and HTTP services (`cmd/agentic-lens`, `internal/domain`, `internal/adapters`).
   - TypeScript 5.x+ strict for frontend (`web/`).
2. **Single Binary Distribution**:
   - The compiled TypeScript frontend (`web/dist`) must be embedded into the Go executable via `embed.FS`.
   - Zero Node.js runtime dependency in production.
3. **FCIS (Functional Core, Imperative Shell)**:
   - `internal/domain`: Pure functions, value semantics, immutable structs. Zero I/O, zero network, zero clocks.
   - `internal/adapters` & `cmd/`: Imperative shell. Traversal, validation, HTTP routing, asset embedding.
4. **Typing & Zero Any**:
   - Zero `any` or `interface{}` across Go and TypeScript.
   - Discriminated unions for multi-state representations.
5. **Zero Emojis**:
   - Zero emojis permitted in code, comments, docstrings, commit messages, PR drafts, UI strings, logs, or workspace artifacts.
