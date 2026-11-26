# GEN-34: Technical design

## Scope and boundaries
Investigate and report on the duration and timeout failures of the integrated `make`. The deliverable is [the report](evidence/REPORT.md) with its raw measurements in `evidence/`. The Makefile, tests, production code, CI and other tickets' changes or ledgers are out of scope and stay untouched.

## Decisions and ownership
This session is the only owner of `openspec/changes/gen-34-investigate-make-duration/`. Measurements ran in this ticket's worktree and use its own build outputs: `web/node_modules`, `backend/internal/adapters/web/dist` and `bin/`. Those outputs are ignored by Git, and nothing else is written.

Measurements use the repository as-is. Instrumentation stays outside the repository: a `make SHELL=` wrapper that timestamps each top-level recipe, a load sampler, `go test -json`, `/usr/bin/time -v` on precompiled test binaries, and the Vitest JSON reporter. None of these changes what the tests do.

## Implementation approval
The user's ticket request authorizes the investigation, measurements, report, telemetry, commit, push and an unmerged PR. It explicitly excludes changes to the Makefile, tests, production code and CI.

## Implementation approach
See [tasks.md](tasks.md). The report says, for each measurement, how loaded the machine was and how many times the measurement was repeated, and it separates measured facts from inference.
