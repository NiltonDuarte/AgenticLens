# feat(dashboard): show full local historical timestamps (AL-61)

Ticket Deep Dive previously showed time alone for historical activity. Show local calendar dates and time to the second in workspace lifecycle labels, trace events, skill load times and the event drawer. Preserve timezone tooltips and source instants, reject impossible calendar dates without normalization, and let longer labels wrap.

Live Inspector precision, chart axes, ordering, durations, storage and API values remain unchanged.

Integrated AL-60's merged comparison search and capture repairs while preserving both features, completed ticket records and shared embedded assets. The predecessor integration and local delivery gates are complete; MGT-2 owns merging. See the [contract](alspaces/AL-61/CONTRACT.md), [summary and verification](alspaces/AL-61/SUMMARY.md), and [compact findings](alspaces/AL-61/iterations/01/REVIEW_SUMMARY.json).
