# AL-94: Make explicit backfill repeatable and report refresh coverage

## Roadmap ticket

[AL-94](../../../alspaces/plan/to-do/AL-94%20make%20explicit%20backfill%20repeatable%20and%20report%20refresh%20coverage.md)

## Why

Operators need to know what an explicit source refresh verified and which files it changed. The engine already avoids replacing equivalent ledgers, but it still rewrites identical requested summaries. A completed migration can coexist with missing source evidence, and the CLI's `Updated` count reflects successful processing even when no ledger is replaced.

## What changes

- Verify exact recorded sessions for the explicitly selected harness with `--recompute-db`, including on completed ledgers. Preserve equivalent ledger bytes and timestamps on complete and still-partial repeats. Publish genuine changes to facts or migration state through the existing safeguards, including changes from requested Git recomputation.
- Compare the requested rendered summary independently. Leave identical content untouched and repair missing or stale managed metrics while preserving authored text. Update the summary only when requested.
- Report coverage per ticket and selected harness. Distinguish recorded sessions with changed facts, sessions verified unchanged, and sessions whose prior facts were retained. Identify retained session IDs and reasons. Describe verified-empty collections, unavailable reads, and retained skill/task snapshots explicitly. Keep newly discovered sessions outside the recorded-session counts.
- Keep migration status and the existing `Updated` counter. Add separate counts of ledgers and summaries actually written. When a selected store is unavailable, identify the requested work blocked before publication and retain the existing failure signal. Make budget-limited results actionable through a retry of the affected ticket and source.

## Capabilities

- `explicit-backfill-refresh` covers repeatable publication, independent summary repair, and coverage reporting for explicitly selected sources.

## Impact

The change affects backfill source recovery, ledger and summary publication, and CLI reporting. Existing accepted history, unrelated harnesses, cancellation safeguards, exit semantics, ordinary backfill selection, and the stored-fact purpose of Force remain in place. Opt-in preview and default rollout are scoped to AL-95.
