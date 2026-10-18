# GEN-4: Update TICKETS.md and Move Merged Tickets to Done

## Summary
Updates `workspaces/plan/TICKETS.md` and ticket specification files for completed tickets AL-15 and AL-30 whose workspaces and implementations have already been merged into `main`. Moves their specification files from `workspaces/plan/to-do/` to `workspaces/plan/done/`, removes them from the `Ready` list, and marks their status as `Done`.

## Changes
- Moved `workspaces/plan/to-do/AL-15 implement universal multiversion schema migration engine for telemetry ledgers.md` to `workspaces/plan/done/`.
- Moved `workspaces/plan/to-do/AL-30 fail telemetry generation on missing root session.md` to `workspaces/plan/done/`.
- Updated `workspaces/plan/TICKETS.md`:
  - Updated individual ticket definition links to reflect `done/` locations for AL-15 and AL-30.
  - Removed AL-15 and AL-30 from `## Ready`.
  - Added AL-15 and AL-30 to `## Done` with `Status: Done` and updated file paths.
- Allocated ticket workspace under `workspaces/GEN-4/`.
