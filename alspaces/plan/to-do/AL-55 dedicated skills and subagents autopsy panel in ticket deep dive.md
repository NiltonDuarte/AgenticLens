h3. Title: AL-55: Dedicated Skills and Subagents Autopsy Panel in Ticket Deep Dive

*TARGET*: AgenticLens (presentation / web / views / deep dive / components / tests)

*CONTEXT*
 
`TicketDeepDive.tsx` serves as the comprehensive retrospective autopsy view for completed tickets. While it provides detailed role breakdowns (input/output tokens, cache hits, cost) in the "Role Effort Distribution" section, it contains zero visibility into what modular skills were loaded during the run, when they were loaded, or how subagent tasks branched from the root orchestrator.

*GOAL*
 
Add a dedicated "Skills & Subagents" section in `TicketDeepDive.tsx` featuring an interactive skills usage table and a subagent task delegation tree.

*VERTICAL SLICE SCOPE*
 
* Presentation / Deep Dive:
  - Add a "Skills & Subagents" card in `TicketDeepDive.tsx` below Role Effort Distribution.
  - Render an interactive Skills Table listing:
    - Skill Name
    - Loaded By Role
    - Invocations Count
    - Error Count
    - First Turn Loaded
  - If `task_links` are present, render a Subagent Delegation Tree showing root orchestrator to child subagent handoffs, turn numbers, task descriptions, and execution status.
  - Provide an empty-state indicator when no skills were loaded during the session.
* Quality:
  - Unit and view tests in `TicketDeepDive.test.tsx` verifying table rendering with populated `skills` data, subagent delegation mapping, and empty states.

*ACCEPTANCE CRITERIA*
 
* "Skills & Subagents" section renders in Ticket Deep Dive whenever a ticket workspace is selected.
* Loaded skills are listed with name, caller role, call counts, errors, and first turn.
* Subagent delegation tree displays parent-to-child session handoffs with descriptions.
* Empty state renders cleanly when a ticket contains zero skill invocations.
* Zero layout shifts or hydration warnings.
* `npm test`, `npm run typecheck`, and `npm run lint` pass cleanly.

*DEFINITION OF DONE*
 
* Code and tests merged to target branch.
* Quality gates passing (`make check`, `make test`, or `make`).
* Go and TypeScript type checking clean with zero any.
* Telemetry tested with `agentic-lens` CLI.
