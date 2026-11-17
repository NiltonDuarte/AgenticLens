# AL-95: Report exact-source refresh coverage

## Goal

An operator can identify which recorded owners need a retry, even when a ledger has already completed migration.

## Context

A 12-ledger batch in [AL-65 discovery, PR #145](https://github.com/NiltonDuarte/AgenticLens/pull/145) reported 12 completed migrations and four incomplete refreshes after exhausting its read budget. The completion count alone gives an unattended caller insufficient evidence of source coverage.

## Acceptance criteria

- For each selected ticket and harness, distinguish refreshed owners, owners verified unchanged, and owners whose stored facts were retained. Identify retained session IDs and reasons such as a missing child, unsupported layout or exhausted budget. Distinguish a retained skill/task snapshot from a fully refreshed tree.
- Show migration status separately from refresh coverage. A completed migration may have incomplete coverage. A verified empty collection is distinguishable from an unavailable read. An unselected harness stays outside the coverage denominator.
- For a missing selected store, preserve the failure signal and disclose which requested work could not run, including when selecting both stores prevents any publication. Keep partial owner evidence and prior unrelated history intact.
- Verify complete, partial and all-unavailable selections and a budget-limited batch. Counts must reconcile with the per-ticket results and identify an exact-ID retry. Keep existing exit semantics unless a separate compatibility decision approves a change.
