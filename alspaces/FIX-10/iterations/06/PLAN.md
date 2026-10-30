# Iteration 06: Require Ticket IDs and Workspaces

## Goal

Correct the prior clarification: all work still requires a ticket ID and isolated workspace. Only the planning record under `plan/` is optional for ad hoc tickets.

## Scope

- Require a ticket ID and workspace for every task, including ad hoc work.
- Use an existing ID when provided; otherwise assign a `GEN-#` ID for ad hoc work.
- Keep the plan lifecycle conditional: move an existing matching specification to `plan/done/` and remove an existing index entry, but do not create either solely for lifecycle compliance.
- Align `AGENTS.md`, `CONTRIBUTING.md`, and the policy block in `docs/SETUP_MANUAL.md`.

## Acceptance

- All three instruction sources require ticket IDs and workspaces.
- Ad hoc tickets may omit a `plan/` specification and index entry.
- Existing to-do specifications still move to `plan/done/` when completed.
- `git diff --check` passes and no conflicting wording remains in the authoritative instruction sources.
