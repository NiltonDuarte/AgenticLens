# feat(dashboard): show full local historical timestamps (AL-61)

Ticket Deep Dive previously showed time alone for historical activity. Display local calendar dates and times to the second in workspace lifecycle labels, trace events, skill load times, and the event drawer. Preserve timezone tooltips and source instants, reject impossible calendar dates without normalizing them, and allow longer labels to wrap.

Live Inspector precision, chart axes, ordering, durations, storage and API values remain unchanged.

Draft checkpoint only; not merge-ready. User authorized draft publication while retaining these pending gates:

- Await MGT-2 notification of the actual AL-60 merge, then integrate the predecessor once, preserving both features and rebuilt embedded assets.
- Rerun required integrated verification after that integration. The earlier full make failed in unchanged capture tests; AL-60 owns the approved shared capture repair. No duplicate repair or premature full-suite retry.
- Complete integration review, final telemetry and ticket lifecycle closure before readiness. MGT-2 owns merging.

See [contract](alspaces/AL-61/CONTRACT.md) and [summary](alspaces/AL-61/SUMMARY.md) for scope and delivery context.
