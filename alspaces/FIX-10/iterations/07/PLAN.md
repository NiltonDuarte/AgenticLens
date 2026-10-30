# Iteration 07: Refresh the Merged AL-59 Ledger

## Goal

Regenerate the AL-59 metrics ledger from its recorded Codex rollout so it conforms to canonical schema v4 and is no longer skipped by workspace loading.

## Scope

- Recompute the AL-59 session facts from the exact Codex rollout ID.
- Persist the session user at the session level and omit invocation `id`, `parent_id`, and `user` fields.
- Refresh the AL-59 summary metrics from the regenerated ledger.
- Validate the ledger through the workspace scanner and schema assertions.

## Acceptance

- AL-59 is present in the workspace scanner output.
- Its ledger is schema v4, every session has a non-empty user, and invocation rows omit ownership fields.
- `git diff --check` passes.
