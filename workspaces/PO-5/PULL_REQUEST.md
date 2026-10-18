# PR: PO-5: Natural Ticket Ordering, Dynamic Roles & Models, Contextual Info Icons, Completion Date Filtering, Custom Progression Windows, and Multi-Dimensional Type Filtering

## Summary
Defines the product specification, architectural contracts, and 7 vertical tracer-bullet Jira tickets across natural ticket sorting (`<PREFIX>-NN`), sortable Landing Page workspaces, dynamic roles and models, contextual hover info icons, Fleet Overview completion date filtering, Metric Progression custom windows and hourly granularity, multi-dimensional filtering by ticket type and workspace prefix, and Conventional Commit telemetry classification across skill manifests.

## Initiatives Covered

1. **Natural Ticket Identifier Ordering & Sortable Tables (AL-23)**:
   - Pure domain natural comparator `CompareTicketIDs(a, b string) int` in Go and `compareTicketIds(a, b string): number` in TypeScript.
   - Guarantees natural sequence: `AL-1`, `AL-2`, ..., `AL-9`, `AL-10`, `AL-11`, ..., `AL-22` instead of alphanumeric sorting.
   - Adds stateful ascending/descending column sorting to the Landing Page Quick-Launch Workspaces directory table.
   - Wires natural sort to Fleet Overview Ticket Directory and Ticket Deep Dive selector.

2. **Dynamic Roles & Models Support (AL-24)**:
   - Eliminates hardcoded role enumerations (`orchestrator`, `planner`, `coder`, `tester`, `reviewer`) across turn velocity, peak context, effort distribution, and visualizations.
   - Treats all non-orchestrator roles as subagents and dynamically discovers, aggregates, and renders arbitrary roles (e.g. `advisor`, `po`, `researcher`, `architect`) and arbitrary models.

3. **Contextual Hover Information Icons for Cards and Tables (AL-25)**:
   - Reusable `InfoIcon` / `InfoTooltip` component adhering strictly to zero-emoji policy.
   - Integrated across KPI cards, distribution cards, pulse cards, and table headers to reveal calculation explanations on hover and focus.

4. **Fleet Overview Completion Date Filtering & Autopsy Sorting (AL-26)**:
   - Filters fleet ledgers by `completed_at` (falling back to `updated_at`), defaulting to Project Lifetime (`all`).
   - Adds completion date sorting toggle to Ticket Deep Dive Workspace Autopsy selector.

5. **Metric Progression Custom Windows & Hourly Granularity (AL-27)**:
   - Adds `1d` (24-hour) window preset, arbitrary `custom` date range pickers, and `hourly` temporal aggregation granularity.
   - Preserves defaults to Project Lifetime window and Daily granularity.

6. **Multi-Dimensional Filtering by Ticket Type and Workspace Prefix (AL-28)**:
   - Adds `ticket_type` (defaulting to `"feature"`) and `workspace prefix` filters to Fleet Overview, Ticket Comparisons, and Metric Progression.

7. **Conventional Commit Classification in Telemetry Engine & Skills (AL-29)**:
   - Upgrades CLI and git inference with deterministic 3-tier resolution (flag -> branch -> ticket prefix).
   - Banishes git commit message inspection, stops inference fallbacks at the workspace ticket prefix, and emits explicit warnings/errors on missing ticket types instead of silently defaulting to `feature`.
   - Updates skill manifests (`agentic-lens-po`, `agentic-lens-dev`, orchestrator, `AGENTS.md`) to explicitly provide `--ticket-type`.

8. **Fail Telemetry Generation on Unresolved Root Session (AL-30)**:
   - Mandates that `agentic-lens generate` must fail with non-zero exit code if no root session is identified from provided information, preventing empty/misleading telemetry ledgers.

## Ticket Breakdown
- Ready: `AL-23`, `AL-24`, `AL-25`, `AL-29`, `AL-30`
- Pending: `AL-26` (blocked by AL-23), `AL-27` (blocked by AL-26), `AL-28` (blocked by AL-26)

## Pointers
- Specification: `workspaces/PO-5/SPEC.md`
- Shared Roadmap: `workspaces/plan/TICKETS.md`
- Session Summary: `workspaces/PO-5/SUMMARY.md`
