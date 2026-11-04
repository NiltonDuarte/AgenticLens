# feat(dashboard): warn about sessions shared across tickets (AL-77)

Fleet Overview now warns when an exact stored session ID appears in multiple tickets. Expand the warning to inspect every affected ticket, source context and existing ticket navigation. Inspection uses the unfiltered project ledger listing, describes partial or unavailable scope, and refreshes when records are corrected.

Require a recorded, nonblank `ticket_id` at the canonical ledger boundary. Missing IDs are rejected with workspace diagnostics rather than inferred from directories or requested paths; valid siblings remain usable. Reads preserve ledger files and valid totals. No automatic reassignment, deletion or cross-ticket deduplication is introduced.

- [Contract](https://github.com/NiltonDuarte/AgenticLens/blob/feat/AL-77-warn-duplicate-sessions/alspaces/AL-77/CONTRACT.md)
- [Approved correction](https://github.com/NiltonDuarte/AgenticLens/blob/feat/AL-77-warn-duplicate-sessions/alspaces/AL-77/iterations/02/PLAN.md)
- [Work summary and verification](https://github.com/NiltonDuarte/AgenticLens/blob/feat/AL-77-warn-duplicate-sessions/alspaces/AL-77/SUMMARY.md)
