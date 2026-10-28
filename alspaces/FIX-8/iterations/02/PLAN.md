# FIX-8 Iteration 02 Plan

## Goal

Update the repository's `AGENTS.md` workspace policy so every workspace and plan path is rooted at the configured `workspaces.dir` value from `.agenticlens-config.yaml`, with an explicit fallback for missing configuration. This is a documentation-only follow-up; do not modify application code or `AGENTS.md` in this planning iteration.

## Policy to specify

1. Define `<WORKSPACES_DIR>` once as the `workspaces.dir` value in `.agenticlens-config.yaml`. If the file or key is absent, use `workspaces` as the documented default. Give a brief example showing that the current configuration resolves to `alspaces`.
2. Use `<WORKSPACES_DIR>` consistently for all ticket artifacts and instructions, including:
   - ticket workspaces and their `SUMMARY.md`, `PULL_REQUEST.md`, `CONTRACT.md`, and `agentic_metrics.json` files;
   - plan intake and completion paths (`plan/to-do/` and `plan/done/`);
   - `plan/TICKETS.md` lifecycle references;
   - iteration `PLAN.md` and `REVIEW.md` paths;
   - telemetry CLI output and append arguments;
   - PR title extraction and `gh pr create --body-file` examples.
3. Ensure examples use the placeholder in command arguments rather than literal `workspaces/` paths, so the instructions work when configured to another directory.
4. Retain the existing lifecycle requirements and ticket directory layout beneath `<WORKSPACES_DIR>`; this change only makes their root configurable.

## Implementation boundaries

- In the subsequent implementation iteration, edit only `AGENTS.md` for the documentation change.
- Do not alter application code, tests, configuration, or any existing ticket artifact locations.
- Do not relocate existing artifacts. The current FIX-8 artifacts remain where they are; relocation is not needed to establish this policy.
- Preserve all unrelated ticket, telemetry, commit, and PR requirements in `AGENTS.md`.
- Keep all documentation free of emojis.

## Verification

- Review every workspace and plan path in `AGENTS.md` and confirm it is expressed relative to `<WORKSPACES_DIR>` rather than a hard-coded `workspaces/` root.
- Confirm the fallback and current `alspaces` example agree with `.agenticlens-config.yaml`.
- Check that the telemetry command, PR title extraction, and PR creation examples consistently use `<WORKSPACES_DIR>`.
- Confirm the implementation diff contains only the intended `AGENTS.md` documentation update and passes `git diff --check`.

## Completion boundary

Iteration 02 defines and implements the configurable workspace-root documentation policy only. It does not change runtime workspace resolution, configuration parsing, application behavior, or existing ticket artifact paths.
