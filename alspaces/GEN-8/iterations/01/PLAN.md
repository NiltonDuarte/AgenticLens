# GEN-8 Iteration 01 Plan

Rename the repository's agent configuration directory from `.opencode/` to `.agents/` and update active references outside `alspaces/` and `testdata/`.

## Scope

- Move all tracked `.opencode/` contents to `.agents/`.
- Update scripts, skill instructions, documentation, and ignore rules to use `.agents/`.
- Leave `alspaces/` and `testdata/` unchanged.
- Verify no active `.opencode` references remain outside those exclusions and run `make`.
