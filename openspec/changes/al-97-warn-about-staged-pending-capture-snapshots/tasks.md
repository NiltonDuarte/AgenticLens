# AL-97: Execution tasks

Use the [design approval record](design.md#implementation-approval) before coding and the [staged snapshot requirements](specs/staged-pending-capture-advisory/spec.md) for acceptance.

## Implementation and focused verification

- [x] **Coder · Setup guidance.** Update the setup skill and manual to offer an explicit preview and install choice. Document the existing local preference, custom hook location, and manual composition for foreign or edited hooks.
- [x] **Coder · Native CLI and adapters.** Inspect bounded staged Git objects under the resolved pending directory using the committing worktree's effective index. Classify valid snapshots, apply exact read-only discard status, and render both pending states within the existing hook's output limit.
- [x] **Tester · Git and capture fixtures.** Cover custom directories, exact staged bytes when the working copy differs or disappears, a real path-limited commit whose commit index differs from the ordinary index, initial commits, and quiet cases for unchanged or deleted files. Exercise both states together, confirmed discard, unavailable local state, disabled warnings, and incomplete inspection. Verify that a direct hook run leaves the index intact and that a real commit succeeds with a warning.

## Integration and delivery

- [x] **Orchestrator · Native verification.** Run the installed hook path through the stable manager and selected engine. Record the tested revision, check setup ownership and linked-worktree behavior, and retain the relevant capture-health regression results.
- [x] **Focused reviewers · Review.** Verify effective index selection, hook ownership, bounded object and registry reads, portable snapshots, and fail-open output. Recheck product corrections and verify both final budget tests against targeted mutations.
- [x] **Orchestrator · Delivery.** Run integrated `make`, record actual evidence and limits, refresh final source-backed telemetry, and prepare the PR handoff.

## Workflow follow-up

Evaluate archiving and roadmap completion after verified delivery.
