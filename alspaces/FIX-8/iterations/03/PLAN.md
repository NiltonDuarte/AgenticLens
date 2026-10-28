# FIX-8 Iteration 03 Plan

## Goal

Reconcile FIX-8 artifact locations with the configured workspace root, `alspaces`, while preserving the complete ticket history and keeping PR #41's artifact references accurate. This is a documentation and workspace-artifact relocation iteration. Do not change application code or tests.

## Planned changes

1. Move the entire `workspaces/FIX-8/` ticket workspace into `alspaces/FIX-8/`, preserving all files and relative structure, including `CONTRACT.md`, `SUMMARY.md`, `PULL_REQUEST.md`, `agentic_metrics.json`, the existing iteration plans and reviews, and this iteration 03 plan.
2. Move the completed FIX-8 specification from `workspaces/plan/done/` into `alspaces/plan/done/` without changing its contents.
3. Remove the obsolete `workspaces/plan/TICKETS.md` created by the prior iteration. Leave the canonical `alspaces/plan/TICKETS.md` unchanged: FIX-8 has no entry there, and no downstream ticket is shown as blocked by FIX-8.
4. Update operational Markdown pointers within the relocated FIX-8 artifacts so ticket-root references use `alspaces/FIX-8` (or an equivalent valid relative link). In particular, update the Context paths in `PULL_REQUEST.md` for the contract, summary, and iteration references. Keep PR #41's title and implementation summary consistent; this iteration changes artifact locations only and adds no application behavior. This historical plan records the original `workspaces/` source paths and may retain them.
5. Regenerate the FIX-8 telemetry ledger as the final workflow step before the final commit, using the configured `alspaces` paths and a unique literal nonce. This environment cannot resolve the current Codex session from its local OpenCode database, so include `--allow-empty` and the explicit ticket title:
   ```bash
   ./bin/agentic-lens generate --nonce <GENERATED_NONCE> --ticket-id FIX-8 --ticket-title "Preserve Codex rollout session trees during backfill" --ticket-type fix --allow-empty --json-output alspaces/FIX-8/agentic_metrics.json --append alspaces/FIX-8/SUMMARY.md
   ```

## Boundaries

- Only ticket workspace and plan artifacts are in scope, together with the obsolete `workspaces/plan/TICKETS.md` removal.
- Do not modify application source, tests, `.agenticlens-config.yaml`, `AGENTS.md`, or canonical `alspaces/plan/TICKETS.md`.
- Do not alter FIX-8 acceptance criteria, implementation claims, or verification results in the completed artifacts.
- Keep all artifacts free of emojis.

## Verification

- Confirm every file from the source workspace `workspaces/FIX-8/` exists under the destination `alspaces/FIX-8/` with its relative structure preserved, including `agentic_metrics.json`, iterations 01 and 02, and this iteration 03 `PLAN.md`.
- Confirm the FIX-8 specification exists in `alspaces/plan/done/` with its original contents and no longer exists in `workspaces/plan/done/`.
- Confirm `alspaces/FIX-8/` and `workspaces/plan/TICKETS.md` no longer exist.
- Check `PULL_REQUEST.md` and other operational pointers under `alspaces/FIX-8/` for stale `workspaces/` references; confirm PR #41's context points to the relocated artifacts. Historical source-path references in this plan are not operational pointers and are expected.
- Compare `alspaces/plan/TICKETS.md` before and after the iteration and confirm it is unchanged.
- Run `git diff --check` and review the final changed-path list to confirm it contains only workspace and plan artifact relocation, pointer updates, and obsolete register removal; confirm no application code or tests changed.

## Completion boundary

Iteration 03 completes the FIX-8 workspace-root reconciliation for the existing ticket artifacts and PR references. It does not revise the completed feature implementation or change the canonical ticket register.
