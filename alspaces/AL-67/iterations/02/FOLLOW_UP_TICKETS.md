# AL-67 follow-up tickets

All twelve assessment findings have one tracking owner. Four new tickets are Ready; three existing related tickets gain the AL-67 evidence instead of being duplicated. Nothing is implemented by this planning delivery.

| Ticket | Plain-language issue | Assessment findings |
| --- | --- | --- |
| [FIX-18](../../../plan/to-do/FIX-18%20report%20unsupported%20live%20telemetry%20accurately.md) | A v2 database can be announced as live-active even though live queries need v1 tables. Report the actual capability without breaking historical traces. | F03 |
| [FIX-19](../../../plan/to-do/FIX-19%20reject%20ambiguous%20codex%20session%20snapshots.md) | An active rollout and its archive copy can be counted as two cumulative snapshots of one session. Diagnose ambiguous owners instead of doubling facts. | F04 |
| [FIX-20](../../../plan/to-do/FIX-20%20stop%20source%20reads%20when%20callers%20cancel.md) | Cancelling a request does not reliably stop the SQL/file read already in progress. Propagate cancellation and release resources while preserving history. | F06 |
| [REV-7](../../../plan/to-do/REV-7%20review%20source%20read%20guarantees.md) | Investigate inconsistent source-read rules, snapshot boundaries and tied-row ordering before deciding on small shared interfaces. | F07, F08, F09, F12 |
| [AL-70, existing](../../../plan/to-do/AL-70%20honor%20configuration%20paths%20and%20ticket%20identity.md) | Generate, server and backfill can search different stores for the same configuration. Keep the configuration fix and its no-fallback/composition tests together. | F01, F02, F11 |
| [AL-69, existing](../../../plan/to-do/AL-69%20preserve%20accurate%20telemetry%20and%20evidence.md) | Child lookup failure currently looks like no children. Preserve useful trace content and expose the unavailable relationship lookup. | F05 |
| [AL-74, existing](../../../plan/to-do/AL-74%20make%20regression%20checks%20reliable%20and%20remove%20unused%20code.md) | Column-omission tests always fail the mixed-schema guard, so they can miss a broken column check. Keep the oracle repair in its existing ticket. | F10 |

Examples in the tickets are illustrative source counterexamples, not invented production incidents. In particular, one real invocation copied into two files must not become two invocations; distinct real repeated invocations must remain distinct.

REV-7 reviews guarantees and proposes bounded follow-ups; it does not authorize a universal repository refactor. AL-63 retains source capability/runtime verification, AL-64 duration, AL-66 generation toggles and AL-62 ownership work. There are no new blocking edges among the four tickets. Shared-file coordination is not a prerequisite.

The assessment and compact reviewer evidence remain under AL-67. Planned tickets have specifications only, without premature developer workspaces or contracts. The user explicitly authorized FIX/REV identifiers; numbering follows existing FIX-17 and REV-6 without renumbering history.
