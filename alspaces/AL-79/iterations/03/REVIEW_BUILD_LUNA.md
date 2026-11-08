# AL-79 iteration 03 build and distribution review (gpt-6-luna / medium requested)

**Outcome: CHANGES_REQUESTED. Severity: 5/10.** Scoped to build artifact naming, release/bootstrap alignment, and active shipped developer guidance. Candidate `c84c6b274d12ef53544596f8214c1e5501cf6022883276fc0184bf528320e5cd`, based on `9c39cf9d4f44bdb968cb92b0212dccf577d10255`. The requested reviewer identity is codex / gpt-6-luna / medium; runtime metadata was not independently available.

## Finding

**Active contributor and architecture instructions still direct users to the removed build output and manager command.** [CONTRIBUTING.md](/home/niltonduarte/.t3/worktrees/AgenticLens/t3-350b0421/CONTRIBUTING.md:64) tells developers to execute `./bin/agentic-lens-<GOOS>-<GOARCH> server`; the revised Makefile emits the engine as `bin/agentic-lens-engine-<GOOS>-<GOARCH>` and the manager as `bin/agentic-lens-<GOOS>-<GOARCH>`. Thus this documented server command invokes the manager, which does not own `server`, instead of the built engine. The same contributor section at lines 56–64 still describes one binary, and lines 189–190 identifies the manager filename as the contributor artifact and says adopters use a common wrapper, though the accepted design removes that wrapper.

`ARCHITECTURE.md` retains direct engine examples under `./bin/agentic-lens` at lines 112–118 and 143–149. That path is no longer produced by `make build`; the platform-stamped `agentic-lens-*` name is the manager, while telemetry commands belong to the engine. Following these examples after a source build therefore fails to find the path or invokes the manager without a selected project engine. Update active instructions to use the generated engine path (or the installed manager with a configured project) and state which command is intended.

## Distribution checks and limits

The Makefile builds both artifact families with the approved names. `release-package` includes each manager and engine in platform archives and writes checksums for archives and loose binaries. Stable, candidate, and Release Please workflows call the shared release targets and upload `dist/bin/agentic-lens-*` plus `SHA256SUMS`, matching bootstrap's loose-artifact URLs. POSIX and PowerShell bootstrap scripts select/check the manager first and pass the distinct engine artifact name to native installation; engine installation checks release digest and host identity. No new release workflow glob defect found.

Evidence was source inspection, not an executed full build or published release. The parent owns `make` and integrated install verification. I did not run tests. Candidate freeze verification matched the supplied hash before review and after completing this report.
