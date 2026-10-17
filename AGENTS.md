# Agent Instructions for AgenticLens

Mandatory operational policies for all agents and contributors working in this repository.

---

## 1. Ticket Tracking Mandate (Zero Untracked Work)

Every task, feature, bugfix, or refactoring in this repository must be tracked by an explicit ticket identifier.

### Rules
1. **Mandatory Ticket Identifier**:
   - A ticket ID is strictly required before starting any work. Agents must never begin work without it.
   - If missing from the initial request, the agent must prompt the user or immediately generate a `GEN-#` ticket (using configured prefix from `.agenticlens/config.json`, fallback: `GEN`). Formats: `<PREFIX>-<ID>` (e.g. `GEN-42` or `AL-42`).
2. **Single-Ticket Exclusivity**:
   - Agents are strictly prohibited from working on more than one ticket at a time.
   - Once work begins on a ticket, the agent is locked to that ticket. Agents must refuse any request to switch to another ticket or work on multiple tickets simultaneously within the same session.
3. **Prompt Steering & Iterations**:
   - Once a ticket is active, every subsequent user prompt must be interpreted as steering feedback for that specific ticket.
   - Agents must create and advance new iterations (`iterations/<NN>/`) as needed to incorporate steering directions.
4. **Dedicated Workspace Allocation**:
   - Every ticket must have its isolated workspace under `workspaces/<TICKET_ID>/`.
   - Never mix artifacts across tickets or leave untracked changes in the working directory.
5. **Workspace Deliverables Checklist**:
   - `workspaces/<TICKET_ID>/SUMMARY.md`: Summary log of work completed, changes made, and verification status.
   - `workspaces/<TICKET_ID>/agentic_metrics.json`: Raw session telemetry ledger.
   - `workspaces/<TICKET_ID>/PULL_REQUEST.md`: Pull request description and context.
6. **Telemetry Ingestion**:
   - Every iteration must extract telemetry facts and append metrics using the AgenticLens CLI:
     ```bash
     NONCE=$(od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
     ./bin/agentic-lens generate --nonce $NONCE --ticket-id <TICKET_ID> --json-output workspaces/<TICKET_ID>/agentic_metrics.json --append workspaces/<TICKET_ID>/SUMMARY.md
     ```

---

## 2. Multi-Agent SDLC & Role Boundaries

Work in this repository is orchestrated through specialized agents and skills:

- **Product / Scope Inception (`agentic-lens-po`)**:
  - Elicits requirements and decomposes features into tracer-bullet tickets.
  - Used when creating new tickets, structuring specifications, or defining acceptance criteria.
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
