# feat(dashboard): show full local historical timestamps (AL-61)

Ticket Deep Dive previously showed time alone for historical activity. Display local calendar dates and times to the second in workspace lifecycle labels, trace events, skill load times, and the event drawer. Preserve timezone tooltips and source instants, reject impossible calendar dates without normalizing them, and allow longer labels to wrap.

Live Inspector precision, chart axes, ordering, durations, storage and API values remain unchanged.

See [contract](CONTRACT.md) and [summary](SUMMARY.md) for scope and delivery context.
