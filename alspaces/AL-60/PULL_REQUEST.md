# fix(capture): wait for native fixture writers before teardown (AL-60)

The shared-source isolation test could remove its temporary directories while a detached native worker was still starting or closing its registry. Generation/ACK equality and a momentarily free worker lease did not establish process completion.

Wait for inherited log writers to close after shutdown: FIFO EOF on Unix and exclusive log-handle access on Windows. Preserve automatic publication, exact project/source/token isolation, pending/ACK and exclusive ownership assertions. Add delayed-start and early-return regression coverage without changing production worker policy or increasing timeouts.

This is a new AL-60 follow-up to merged PR #101 and precedes AL-61 PR #104. Scope and delivery context: [contract](alspaces/AL-60/CONTRACT.md), [iteration plan](alspaces/AL-60/iterations/03/PLAN.md) and [summary](alspaces/AL-60/SUMMARY.md).
