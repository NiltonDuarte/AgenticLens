# feat(live-inspector): collapse failed tool call details by default (AL-46)

## Description
Add a collapsed-by-default Failed Tool Calls section to the Live Trace Inspector. The header keeps the exact failure count prominent while preserving immediate visibility of the execution message stream.

- Add an accessible expand/collapse control and high-visibility failure badge.
- Show failed alert cards only when expanded; omit the section when there are no failures.
- Add component styles and tests for count, initial state, stream visibility, toggling, keyboard activation, and failure details.

## Context
- Contract: [CONTRACT.md](alspaces/AL-46/CONTRACT.md)
- Plan: [iteration 01](alspaces/AL-46/iterations/01/PLAN.md), [iteration 02](alspaces/AL-46/iterations/02/PLAN.md)
- Summary: [SUMMARY.md](alspaces/AL-46/SUMMARY.md)
