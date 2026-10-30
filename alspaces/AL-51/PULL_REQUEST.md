# feat(observability): ingest OpenCode skills and task links (AL-51)

## Summary

- Add optional skill and task delegation facts to historical ledgers and live trace parts.
- Extract facts from both supported OpenCode SQLite schema families and persist them through ledger generation and backfill.
- Stream skill/task attributes through live trace updates, correlating child sessions only from explicit records.
- Preserve schema v4 and legacy ledger compatibility; add no migration or UI presentation.

## Context

- Contract: `alspaces/AL-51/CONTRACT.md`
- Work and verification summary: `alspaces/AL-51/SUMMARY.md`
- Iteration plans and reviews: `alspaces/AL-51/iterations/`
