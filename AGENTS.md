# Agent Instructions for AgenticLens

Mandatory operational policies for all agents and contributors working in this repository.

---

## 1. Ticketed and Ad Hoc Work

Every task, feature, bugfix, or refactoring in this repository must have a ticket identifier and an isolated ticket workspace. A planning specification or index entry under `plan/` is not required for an ad hoc ticket.

### Configurable Workspace Root
- Define `<WORKSPACES_DIR>` as the `workspaces.dir` value in `.agenticlens-config.yaml`.
- If `.agenticlens-config.yaml` or the `workspaces.dir` key is absent, use `workspaces` as the fallback.
- The current configuration sets `workspaces.dir` to `alspaces`, so `<WORKSPACES_DIR>` currently resolves to `alspaces`.
- Use `<WORKSPACES_DIR>` as the root for all ticket workspaces and plan paths below.

### Rules
1. **Mandatory Ticket Identifier**:
   - A ticket ID is required before starting any work. If the request does not provide one, ask the user or assign an ad hoc `GEN-#` identifier using the configured prefixes in `.agenticlens-config.yaml` (fallback: `GEN`).
   - Use an existing ticket ID when one is provided; do not create a duplicate.
   - An ad hoc ticket ID does not require a corresponding specification or entry under `<WORKSPACES_DIR>/plan/`.
2. **Single-Ticket Exclusivity**:
   - Agents are strictly prohibited from working on more than one ticket at a time.
   - Once work begins on a ticket, the agent is locked to that ticket. Agents must refuse any request to switch to another ticket or work on multiple tickets simultaneously within the same session.
3. **Prompt Steering & Iterations**:
   - Once a ticket is active, every subsequent user prompt must be interpreted as steering feedback for that specific ticket.
   - Advance iterations (`iterations/<NN>/`) when scope or acceptance boundaries change. Keep clarification, status questions, and review corrections within the current iteration; do not create a new plan for each message. Preserve approval and completed verification across compaction. Honor explicit requests to consolidate plans and remove empty iteration folders.
4. **Dedicated Workspace Allocation**:
   - Every task must have its isolated workspace under `<WORKSPACES_DIR>/<TICKET_ID>/`, including ad hoc tickets.
   - Never mix artifacts across tickets or leave task-owned changes untracked at handoff. Record the initial Git status and preserve unrelated user or agent changes; stage only files owned by this task.
   - Check the existing workspace and ticket index before allocating an ad hoc identifier. A generated branch name is not proof that its ticket number is unused.
   - A rename or relocation of ticket artifacts must not require application or test changes. Tests must use self-contained fixtures, never the live ticket workspace.
5. **Workspace Deliverables Checklist**:
   - `<WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md`: Summary log of work completed, changes made, and verification status.
   - `<WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`: Pull request description and context.
   - `<WORKSPACES_DIR>/<TICKET_ID>/CONTRACT.md`: Frozen technical contract and acceptance boundaries, when required by the configured artifact list.
   - **Plan Lifecycle Synchronization**: When a matching specification exists in `<WORKSPACES_DIR>/plan/to-do/<TICKET_ID>*.md`, move it to `<WORKSPACES_DIR>/plan/done/` when the ticket is completed. Remove the completed ticket entry from `<WORKSPACES_DIR>/plan/TICKETS.md` only if one exists; do not create a specification, ticket entry, or Done section just to satisfy this lifecycle rule. For ticketed work, unblock downstream tickets whose dependencies are now fully satisfied by promoting them to `## Ready` with `Status: Ready`.
   - `<WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json`: Raw session telemetry ledger (generated at the very final step of the iteration).
   - Store iteration plans and reviews under `<WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/PLAN.md` and `<WORKSPACES_DIR>/<TICKET_ID>/iterations/<NN>/REVIEW.md`.
6. **Telemetry Ingestion as the Final Step & Conventional Commit Type Classification**:
   - **Delayed Ingestion Mandate**: Generating telemetry must always be delayed to the very last step of the workflow. Running telemetry early causes subsequent turns, tool calls, compactions, and subagent sessions to be omitted from the ledger.
   - **Literal Nonce Generation**: The agent must generate a unique random nonce directly (by any means, e.g. an 8 to 16 character alphanumeric or hex string) and supply the literal token directly in the command. Shell variables (`$NONCE`) or command substitutions (`$(...)`) must not be used because OpenCode records tool inputs into SQLite before subshell expansion occurs, which causes database searches for the expanded token to fail.
   - Execute AgenticLens CLI with an explicit conventional commit `type` immediately before final commit:
     ```bash
     ./bin/agentic-lens generate --nonce <GENERATED_NONCE> --ticket-id <TICKET_ID> --ticket-type <CONVENTIONAL_TYPE> --json-output <WORKSPACES_DIR>/<TICKET_ID>/agentic_metrics.json --append <WORKSPACES_DIR>/<TICKET_ID>/SUMMARY.md
     ```
   - Select the actual harness explicitly when defaults are ambiguous. For a Codex session, add `--source codex`; use `--codex-home` only when its session store is outside the normal configured location. Confirm supported options with the built CLI.
   - Before committing telemetry, verify the root session, source, expected child sessions, and preservation of existing sessions from other harnesses. A successful exit alone does not establish complete extraction. Record unavailable facts or source errors; never manufacture a successful or empty ledger.
   - Commit the generated ledger and rendered summary. Keep raw transcripts outside Git. Backfill requires the machine holding the original source sessions; unavailable source data must remain explicitly incomplete.
   - `<CONVENTIONAL_TYPE>` must match standard conventional commit classifications (`feat`, `fix`, `docs`, `chore`, `refactor`, `test`, `ci`, `perf`, `build`).
   - PO tickets (`PO-#`) must strictly use `--ticket-type docs`.
   - Generic tickets (`GEN-#`) must use `--ticket-type chore` unless explicitly a feature (`feat`).
   - Developer tickets must use the corresponding conventional commit type matching their branch prefix (`feat`, `fix`, `refactor`, etc.). Never default all tickets to `feature`.
   - Ad hoc tickets use their assigned ticket ID and isolated workspace for telemetry. They do not require a specification or index entry under `plan/`.
7. **Pull Request Delivery & Handoff**:
   - Commit all implementation deliverables on the feature branch first (stage the reviewed task-owned paths, then commit). Git requires a clean working tree before merging.
   - Merge `main` into the feature branch (`git fetch origin main && git merge origin/main --no-edit`) to resolve any upstream conflicts, and verify test suites pass (`make`).
   - Extract the PR title directly from the first line of `PULL_REQUEST.md` (e.g. `PR_TITLE=$(head -n 1 <WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md | sed 's/^# *//')`).
   - Check the existing PR state before follow-up delivery. If it has merged, compare the remaining branch changes with current `origin/main` and open a new PR for any undelivered changes; do not report a merged PR as delivering later work.
   - Commit final deliverables and telemetry facts on the feature branch, push upstream, and open a GitHub Pull Request via `gh pr create --title "$PR_TITLE" --body-file <WORKSPACES_DIR>/<TICKET_ID>/PULL_REQUEST.md`.
   - Avoid directly merging or checking out `main` within agent worktrees.
   - The Pull Request URL must be prominently displayed at the top of the final handoff report.

---

## 2. Multi-Agent SDLC & Role Boundaries

Application development in this repository is orchestrated through specialized agents and skills. Documentation-only maintenance and session audits may directly update instructions and ticket artifacts without launching a full development cycle. If the work reaches application code, tests, schemas, or runtime scripts, use the development workflow below.

For delegated work, name one owner for each writable path, state the role and acceptance boundaries, and pass the contract, current plan, review findings, and baseline commit. Tell agents that they share the checkout and must preserve others' changes. The parent owns integration, telemetry, commits, and PR delivery unless explicitly delegated. A tester or reviewer must not publish or continue product implementation on its own.

Specialized roles:

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
   - Go 1.23+ for backend, CLI, and HTTP services under `backend/` (`backend/cmd/agentic-lens`, `backend/internal/domain`, `backend/internal/adapters`).
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


## 4. Session-Grounded Execution Practices

- **Ground before changing**: Inspect the configuration, owning module, existing tests, and current command help. Use configured workspace roots, code classification patterns, and source locations; never infer them from this repository's ticket names or one developer's checkout.
- **Keep evidence distinct from inference**: For telemetry work, record source and session identifiers and validate against real source records. Synthetic fixtures are for deterministic tests, not evidence of a real session. Unknown records, missing children, unavailable pricing, and partial migrations must be surfaced according to the data contract rather than silently treated as complete or zero.
- **Preserve baseline functionality**: Optional hydration enriches stored facts. Test the persisted-data path with enrichment unavailable as well as available; enrichment failure must not hide supported ledger data.
- **Name by responsibility and version**: Version-specific adapters and tests should identify both the schema family and responsibility (for example, `schema_v2_session_extraction.go`). Shared logic should not claim to be version-specific. Keep public comments consistent with actual dispatch behavior.
- **Bound investigation and output**: Use targeted searches and bounded reads. After a repeated failure, inspect the cause and change the hypothesis or command before retrying. Summarize resolved findings and remaining work in the active iteration so a compacted session does not restart discovery.
- **Coordinate verification**: Record the tested commit or diff, command, result, and environment limitation in the review. Reuse valid results for unchanged code. Re-run affected checks after fixes and the required `make` gate after integration; reviewers need not duplicate a completed suite without a concrete concern. Report skipped stages explicitly.
- **Keep the shipped UI current**: Frontend changes require the canonical frontend build and verification of the embedded asset tree under `backend/internal/adapters/web/dist/`. Do not restore stale tracked assets merely to reduce the diff. Verify the standalone binary's assets as well as development-source behavior.
- **Close with evidence**: State delivered behavior, verification, remaining limitations, and the correct PR URL. Do not call work complete because a plan exists, a child reports success, or an earlier PR passed checks.

Rationale and session references: [GEN-12 session audit](alspaces/GEN-12/iterations/01/REVIEW.md).
