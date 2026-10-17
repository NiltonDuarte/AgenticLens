# chore(plan): discard AL-10 and incorporate git metrics extraction into AL-11

## Summary of Changes
Rationalizes telemetry extraction tickets by discarding AL-10 in favor of direct native Go extraction in AL-11, avoiding throwaway maintenance on the legacy Python extractor script.

- **AL-10 Discard**: Moved `AL-10 fix git metrics calculation in python extractor and backfill ledgers.md` to `workspaces/plan/discarded/` with detailed discard reasoning.
- **AL-11 Scope Enhancement**: Amended `workspaces/plan/to-do/AL-11 implement native go telemetry extractor and session generation.md` to include all multi-language git metrics calculation fixes (universal file patterns for Go, TypeScript, and Python; configuration overrides from `.agenticlens/config.json`; robust merge-base and commit/branch resolution; division-by-zero protection).
- **Ticket Index & Dependency Update**: Updated `workspaces/plan/TICKETS.md` marking AL-10 as Discarded and unblocking AL-11.

## References
- Work & Telemetry Summary: `workspaces/AL-10/SUMMARY.md`
- Discarded Ticket Specification: `workspaces/plan/discarded/AL-10 fix git metrics calculation in python extractor and backfill ledgers.md`
- Updated Target Specification: `workspaces/plan/to-do/AL-11 implement native go telemetry extractor and session generation.md`
