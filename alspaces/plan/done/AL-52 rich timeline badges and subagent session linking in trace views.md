# AL-52: Rich Timeline Badges and Subagent Session Linking in Trace Views

- **Subsystem**: Presentation / Web / Components / Waterfall / Inspector / Tests
- **Phase**: Phase 8 (v2 Observability / Skills)
- **Status**: Completed
- **Blocked by**: None
- **Blocks**: AL-55

## Planning status (2026-10-31)

AL-51 merged in [PR #62](https://github.com/NiltonDuarte/AgenticLens/pull/62), providing optional skill/task facts and live attributes. Its dependency is satisfied; this presentation work remains open. Supported v4 ledgers may omit those optional fields; ordinary runtime rejects v1-v3 ledgers until explicit migration and source-backed backfill complete.

## Context
When inspecting agent execution traces in the Trace Waterfall or Live Trace Inspector, tool calls for `skill` and `task` appear as generic tool invocations. Operators must expand the JSON arguments manually to understand what skill was loaded or what subagent role was spawned.

## Goal
Surface rich, prominent visual badges for skill invocations and subagent delegations in the Trace Waterfall and Live Inspector message streams, with interactive jump links to navigate to child sessions.

## Scope Across Layers

### 1. Presentation Components (`web/src/components/TraceWaterfall.tsx`, `web/src/components/LiveTraceInspector.tsx`)
- Render distinct visual badges when a message part represents a skill or task:
  - Skill badge: `[skill: <name>]` styled with an accent badge (e.g. purple/indigo hue) indicating injected capability.
  - Subagent task badge: `[task -> <subagent_type>]` styled with a delegation badge (e.g. emerald/cyan hue).
- When a `child_session_id` is present on a task link or live part:
  - Render an interactive link/button that navigates the Live Observability tree or deep dive view directly to that child session.
  - If the child session is pending or absent from the current active tree, display an informative tooltip indicating the child session ID with a copy button.

### 2. Styling and Responsiveness (`web/src/index.css`)
- Ensure badges conform to dark and light theme tokens.
- Apply `white-space: nowrap` and `overflow: hidden` with `text-overflow: ellipsis` on skill names and subagent roles to prevent horizontal layout blowout on narrow viewports.

### 3. Automated Tests
- Vitest unit tests in `web/src/components/TraceWaterfall.test.tsx` verifying badge rendering for skill and task spans.
- Vitest unit tests in `web/src/components/LiveTraceInspector.test.tsx` verifying interactive jump links and tooltips.

## Acceptance Criteria
1. Trace Waterfall displays prominent `[skill: <name>]` badges on skill invocations.
2. Trace Waterfall and Live Inspector display `[task -> <subagent_type>]` badges on task delegations.
3. Clicking a task delegation with a valid child session identifier routes the view to inspect the child session.
4. Truncation and tooltips prevent badge text overflow on narrow viewports.

## Definition of Done
- Component changes authored and styled cleanly.
- Vitest test coverage for badge rendering and link navigation.
- All tests pass (`npm test`).
- Zero emojis across all code, tests, and documentation.
