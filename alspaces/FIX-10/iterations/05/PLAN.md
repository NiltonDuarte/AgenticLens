# Iteration 05: Ticket and Plan Lifecycle Requirements

## Goal

Clarify repository instructions so existing tickets and specifications are handled consistently, while ad hoc work can proceed without creating ticket artifacts.

## Scope

- Use an existing ticket when work is associated with one; do not create a duplicate.
- Allow ad hoc work without a ticket ID, ticket specification, ticket workspace, or ticket-specific telemetry.
- Move a matching specification from `plan/to-do/` to `plan/done/` when the ticket is completed, but only when that specification exists.
- Remove a completed ticket entry from `plan/TICKETS.md` only when present. Do not create artifacts just to satisfy lifecycle synchronization.
- Align `AGENTS.md`, `CONTRIBUTING.md`, and the standard policy block in `docs/SETUP_MANUAL.md`.

## Acceptance

- The three instruction sources consistently distinguish ticketed and ad hoc work.
- No instruction requires creating a `GEN` ticket for an ad hoc request.
- Plan completion is conditional on an existing to-do specification; index removal is conditional on an existing entry.
- Documentation diff passes `git diff --check`.
