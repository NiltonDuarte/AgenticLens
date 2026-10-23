# AL-53: Dedicated Skills and Subagents Autopsy Panel in Ticket Deep Dive

- **Subsystem**: Presentation / Web / Views / Deep Dive / Components / Tests
- **Phase**: Phase 8 (v2 Observability / Skills)
- **Status**: Pending
- **Blocked by**: AL-51
- **Blocks**: AL-55

## Context
Ticket Deep Dive provides an autopsy of an agent's run across tokens, costs, tools, and files. However, there is no high-level overview of which modular skills were loaded during the session or how subagents were delegated. Operators must comb through hundreds of individual tool spans in the waterfall to reconstruct the delegation graph.

## Goal
Build a dedicated "Skills & Delegations" autopsy panel in Ticket Deep Dive that summarizes all modular skills injected and subagents spawned during the ticket lifecycle, with defensive handling for legacy ledgers.

## Scope Across Layers

### 1. Presentation View (`web/src/views/TicketDeepDive.tsx`)
- Add a new collapsible autopsy card/section: `Skills & Subagents`.
- Defensive fallback: if `ledger.skills` is undefined, empty, or un-migrated, render an elegant empty state: `"No modular skills or subagent delegations recorded for this ticket."`
- When skills exist:
  - Table displaying: Skill Name, Loaded At (Turn / Time), Invoking Role (Orchestrator, Coder, etc.), Execution Status (Success / Failed), and Error Details if any.
- When task delegations exist:
  - Subagent Delegation Tree/Table displaying: Subagent Type, Description/Task Prompt Summary, Parent Turn, Duration, Cost (USD), and Jump Link to Child Session.

### 2. Formatting and Accessibility (`web/src/domain/formatters.ts`)
- Add formatters for skill success badges and subagent role labels.
- Provide accessible ARIA attributes and keyboard navigation on table rows.

### 3. Automated Tests
- Vitest tests in `web/src/views/TicketDeepDive.test.tsx`:
  - Renders empty state when `skills` and `task_links` are undefined (legacy v3 ledgers).
  - Renders skills table when `skills` array is populated.
  - Renders delegation links when `task_links` array is populated.

## Acceptance Criteria
1. Dedicated "Skills & Subagents" section appears in Ticket Deep Dive.
2. Legacy v3 ledgers without `skills` or `task_links` render without console errors or UI crashes.
3. Skills table clearly identifies skill names, invocation turns, caller roles, and error states.
4. Delegation section identifies subagent roles and links to child sessions when available.

## Definition of Done
- Component and view tests written with Vitest.
- Visual hierarchy matches existing deep dive cards.
- All tests pass (`npm test`).
- Zero emojis across all code, tests, and documentation.
