# AL-46 Implementation Summary

## Overview
Made failed tool call details collapsible by default in the Live Trace Inspector while keeping the failure count prominent and the execution message stream immediately visible.

## Key Changes

### Domain
- No domain model changes; the inspector uses the existing failed tool call count and recent error facts.

### Adapters
- No network or adapter changes; expanded/collapsed state is local to the inspector component.

### Presentation
- Added a high-visibility `Failed Tool Calls: N` badge and accessible toggle with `aria-expanded` and `aria-controls`.
- Kept failed alert cards hidden until expanded, and omitted the section when there are no failed tool calls.
- Added responsive summary styling and visible keyboard focus treatment.
- Added coverage for default collapsed state, exact count, message stream visibility, click and keyboard toggling, and expanded failure details.

## Verification
- `npm run typecheck`: passed.
- `npm run lint`: passed.
- `npm test -- --run`: 33 test files and 822 tests passed.
- `make check`: passed from the repository root.
- `git diff --check`: passed.
