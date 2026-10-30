# Consolidated GEN-10 Plan

This plan consolidates the accepted scope and work from the earlier iterations.

## Scope and work

1. Review README installation guidance, the setup skill, the setup manual, and release asset conventions.
2. Focus the README installation section on a copyable prompt that directs an OpenCode agent to the AgenticLens setup skill; leave manual setup detail in the linked setup manual.
3. Correct the README Codex generation example to select a root session and provide the required `--source` flag in the README and Architecture backfill examples.
4. Remove fixed schema-version claims from Architecture migration guidance.
5. Leave Contributor release guidance unchanged per user direction.
6. Add the MIT license file referenced by README.
7. Consolidate the iteration plans and reviews in this folder and remove their superseded folders.

## Acceptance and verification

- Installation guidance in README centers on agent-assisted setup.
- CLI examples include required source/session selection.
- Architecture migration docs do not claim a fixed current schema version.
- The README license link resolves to the added MIT license.
- `git diff --check` and `make build` pass; tests are not required for these documentation and workspace changes.
