# AL-78: Execution tasks

Start after the [implementation approval](design.md#implementation-approval) is recorded. Verify the [observable acceptance](specs/completion-metrics-freshness/spec.md) against the approved scope.

## Implementation and focused verification

- [x] Orchestrator and planner assign the CLI and guidance to the coder, checkpoint tests to the CLI tester, and the packaged fixture to the native tester. The host has Codex 0.170.0 and OpenCode 1.19.5. The isolated runs select their own source stores and workspaces.
- [x] Tester added the opt-in packaged native scenario under `scripts/tests/fixtures/capture-release-gate/`. The baseline reproduced a real code commit, the qualified warning in model input at the default threshold, exact native completion, automatic publication, and external manual refresh on both clients. See [native evidence](evidence/native-feedback.md).
- [x] Coder updated the CLI checkpoint to point to an automatic refresh or exact-session external generation after native completion. The candidate preserves threshold, attribution, and ledger semantics. Native latency verification remains below.
- [x] Targeted race tests verified checkpoint qualification, same-owner idempotent refresh, unavailable and other-harness preservation, and unknown cost. The packaged scenario also verified external generation after native completion on both clients.
- [x] Coder updated the setup manual, operator CLI reference, and developer guidance to cover natural completion, automatic refresh, external generation, and a later telemetry commit.

## Integration and delivery

- [x] The packaged candidate passed the strengthened timing and final-usage checks on both native clients. The [native evidence](evidence/native-feedback.md) compares it with the baseline and records the measurement limits.
- [x] Independent behavior, data, architecture, and test reviews approved the scoped corrections on both model profiles. Integrated verification remains open.
- [ ] Orchestrator runs integrated `make` and applicable native checks, records tested revisions, results, and unresolved limitations in `delivery.md`, and completes final telemetry and PR delivery through the repository workflow.

## Workflow follow-up

Archive after verified delivery and reconcile the roadmap separately.
