# AL-94: Avoid unchanged writes on exact-source refresh

## Goal

An operator can repeat an explicit source refresh and repair a requested summary without disturbing a current ledger.

## Context

An unchanged second DB pass in [AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145) advanced updated_at and both file modification times. A summary can also be missing its metrics while the ledger already contains the correct facts.

## Acceptance criteria

- With the harness explicitly selected and --recompute-db requested, verify the recorded identities even on a completed ledger. When persisted facts and migration state are equivalent, ledger bytes, updated_at and modification time remain identical. An equivalent repeat of a still-partial refresh preserves its pending state without timestamp churn.
- Compare the requested summary independently. Leave identical rendered content untouched. Create a missing requested summary or repair stale metrics while preserving authored text and the unchanged ledger. Without a summary request, leave the summary alone.
- When original evidence changes a selected fact or resolves a pending step, publish the genuine change with the existing publication safeguards. Verify a changed first pass followed by a stable repeat, including a missing child and another harness whose prior facts survive. A source failure retains accepted files and existing failure signals.
- Exercise explicit DB and combined DB/Git calls. A genuine requested Git change must still publish. Force keeps its stored-fact maintenance purpose. Ordinary backfill keeps its existing selection.
