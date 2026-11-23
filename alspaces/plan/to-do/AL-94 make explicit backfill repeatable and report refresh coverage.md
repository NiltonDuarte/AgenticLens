# AL-94: Make explicit backfill repeatable and report refresh coverage

[OpenSpec proposal](../../../openspec/changes/al-94-make-explicit-backfill-repeatable-and-report-refresh-coverage/proposal.md)

## Goal

Operators can repeat an exact-source refresh, repair a requested summary, and tell which recorded owners still need a retry.

## Context

A second pass over unchanged source facts rewrote both the ledger and summary. In a separate batch, all 12 ledgers reported completed migrations while four refreshes exhausted the source-read budget. Operators need to know both what changed on disk and what the refresh verified. [AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145)

## Acceptance criteria

- With the harness explicitly selected and --recompute-db requested, verify the recorded identities even on a completed ledger. When persisted facts and migration state are equivalent, ledger bytes, updated_at and modification time remain identical. An equivalent repeat of a still-partial refresh preserves its pending state without timestamp churn.
- Compare the requested summary independently. Leave identical rendered content untouched. Create a missing requested summary or repair stale metrics while preserving authored text and the unchanged ledger. Without a summary request, leave the summary alone.
- When original evidence changes a selected fact or resolves a pending step, publish the genuine change with the existing publication safeguards. Verify a changed first pass followed by a stable repeat, including a missing child and another harness whose prior facts survive. A source failure retains accepted files and existing failure signals.
- Exercise explicit DB and combined DB/Git calls. A genuine requested Git change must still publish. Force keeps its stored-fact maintenance purpose. Ordinary backfill keeps its existing selection.

- For each selected ticket and harness, distinguish refreshed owners, owners verified unchanged, and owners whose stored facts were retained. Identify retained session IDs and reasons such as a missing child, unsupported layout or exhausted budget. Distinguish a retained skill/task snapshot from a fully refreshed tree.
- Show migration status separately from refresh coverage. A completed migration may have incomplete coverage. A verified empty collection is distinguishable from an unavailable read. An unselected harness stays outside the coverage denominator.
- For a missing selected store, preserve the failure signal and disclose which requested work could not run, including when selecting both stores prevents any publication. Keep partial owner evidence and prior unrelated history intact.
- Verify complete, partial and all-unavailable selections and a budget-limited batch. Counts must reconcile with the per-ticket results and identify an exact-ID retry. Keep existing exit semantics unless a separate compatibility decision approves a change.
