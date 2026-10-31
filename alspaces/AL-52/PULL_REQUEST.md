# feat(traces): show invocation badges and child session links (AL-52)

Skill and task invocations now show named badges in the Trace Waterfall and Live Trace Inspector. Recorded child links open available sessions in the current view; pending or unavailable children expose the full session identifier and a copy action.

Retrospective metadata is matched conservatively so ambiguous invocation facts cannot navigate to the wrong child. Badges use theme-aware colors and truncated labels with full-value access.

Contract: [alspaces/AL-52/CONTRACT.md](alspaces/AL-52/CONTRACT.md)
Summary and verification: [alspaces/AL-52/SUMMARY.md](alspaces/AL-52/SUMMARY.md)
