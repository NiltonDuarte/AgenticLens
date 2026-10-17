# Agent Instructions for AgenticLens

Mandatory operational policies for all agents and contributors working in this repository.

---

## 1. Ticket Tracking Mandate (Zero Untracked Work)

Every task, feature, bugfix, or refactoring in this repository must be tracked by an explicit ticket identifier.

### Rules
1. **Existing Ticket or Immediate Creation**:
   - Before executing changes or dispatching subagents, verify whether the task has an existing ticket ID (e.g. `AL-00101` or `GEN-00101`).
   - If no ticket ID is provided by the user, you must create one immediately before touching code or planning an iteration.
   - Use the configured ticket prefix from `.agenticlens/config.json` (fallback: `AL` or `GEN`). Formats: `<PREFIX>-<ID>` (e.g. `AL-00042`).
2. **Dedicated Workspace Allocation**:
   - Every ticket must have its isolated workspace under `workspaces/<TICKET_ID>/`.
   - Never mix artifacts across tickets or leave untracked changes in the working directory.
3. **Workspace Deliverables Checklist**:
   - `workspaces/<TICKET_ID>/SUMMARY.md`: Summary log of work completed, changes made, and verification status.
   - `workspaces/<TICKET_ID>/agentic_metrics.json`: Raw session telemetry ledger.
   - `workspaces/<TICKET_ID>/PULL_REQUEST.md`: Pull request description and context.
4. **Telemetry Ingestion**:
   - Every iteration must extract telemetry facts and append metrics using the AgenticLens CLI:
     ```bash
     NONCE=$(python3 -c "import uuid; print(uuid.uuid4().hex[:8])" 2>/dev/null || od -N 4 -tx1 /dev/urandom | head -1 | tr -d ' ')
     python3 scripts/extract_agentic_metrics.py --nonce $NONCE --ticket-id <TICKET_ID> --json-output workspaces/<TICKET_ID>/agentic_metrics.json
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
