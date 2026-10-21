# PR: PO-6: Telemetry Event Pruning, Database Message Stripping, Dynamic Local Hydration, and Structural Hierarchy Integrity

## Summary
Defines the product specification, architectural contracts, and 3 consolidated vertical tracer-bullet Jira tickets addressing `RFC-telemetry-sanitization-and-local-hydration.md` and `ISSUE-root-session-role-conflation.md`, alongside consolidating Phase 4 live backend tickets (AL-17 & AL-18).

## Initiatives Covered

0. **Consolidation of Live Backend Slice (AL-17 & AL-18)**:
   - Merged AL-18 into AL-17 (`Implement Non-Locking Live Database Hook, Session Tree Engine, Ticket Inference, and Live Trace Telemetry`), unifying active session tree discovery, config-driven ticket inference, and live trace message/reasoning extraction into a single vertical backend slice.
   - Moved AL-18 to `workspaces/plan/discarded/` as superseded.
   - Updated downstream dependencies: AL-19, AL-20, and AL-33 now link directly to AL-17.

1. **Decouple Root Session Structural Hierarchy from Persona Roles (AL-31)**:
   - Eliminates conflation between structural hierarchy (`ParentID == nil` vs `ParentID != nil`) and persona names (`general`, `orchestrator`, `build`, etc.).
   - Evaluates root coordination metrics (`OrchestratorPeakTokens`, saturation percentages) from root invocations regardless of persona name.
   - Restricts subagent metrics (`SubagentsPeakTokens`, `SubagentsPeakRole`, velocity, effort distribution) strictly to child invocations (`ParentID != nil`).
   - Fixes 8 failing test cases in `backend/internal/adapters/parity_test.go` and updates baseline fixtures.

2. **Telemetry Event Pruning, Database Message Stripping, Agent Investigation, and Backfill (AL-32)**:
   - Enforces zero database message persistence in Git: bypasses fragile regex sanitization and completely strips database error messages, prompts, and assistant text from committed ledgers.
   - Persists bare tool failure metadata (`tool_name`, `status: "error"`, turn, timestamp, duration) with `error_message: nil`.
   - Adds an investigation step to explore obtaining structured error metadata directly from the agent runtime versus relying solely on local hydration.
   - Prunes routine completed events to keep committed ledgers compact (<5 KB).
   - Registers schema migration step (`StepV2ToV3Sanitized`) and integrates backfill runner to retroactively prune routine events and strip database text from historical workspace ledgers.

3. **End-to-End Dynamic Local Database Event Hydration and Dual-Plane Waterfall (AL-33)**:
   - Delivers a full vertical slice across Go backend and TypeScript frontend for dynamic local database hydration.
   - Backend: exposes `GET /api/workspaces/:id/events` with non-locking read-only SQLite queries (`PRAGMA query_only = ON;`, WAL mode) to dynamically reconstruct full prompt, tool, and reasoning traces in memory from local OpenCode database when available (`is_hydrated: true`), seamlessly degrading to committed bare events when offline (`is_hydrated: false`).
   - Frontend: updates `TraceWaterfall.tsx` and autopsy views to render rich expandable message and reasoning drawers when hydrated locally, and a compact bare error timeline with an informational offline banner when unhydrated.

## Ticket Breakdown
- Ready: `AL-31`, `AL-32`
- Pending: `AL-33` (blocked by AL-32, AL-17)

## Pointers
- Specification: `workspaces/PO-6/SPEC.md`
- Shared Roadmap: `workspaces/plan/TICKETS.md`
- Session Summary: `workspaces/PO-6/SUMMARY.md`
