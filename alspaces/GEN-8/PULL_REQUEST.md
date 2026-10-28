# chore: rename .opencode directory to .agents

Rename the repository's agent configuration and skills directory to `.agents/`, updating active references outside `alspaces/` and `testdata/`.

## Changes

- Move tracked agent personas and skills under `.agents/`.
- Update setup instructions, release download paths, ignore rules, and scripts.
- Preserve archived ticket and fixture references under `alspaces/` and `testdata/`.

## Verification

`make` passed, including backend checks, 727 frontend tests, and multi-platform builds. A scoped search found no `.opencode` references outside `alspaces/` and `testdata/`.
