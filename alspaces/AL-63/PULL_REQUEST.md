# docs(audit): verify cross-harness source capabilities (AL-63)

Execute AL-63's existing discovery ticket and restore its specification to the pre-session text. Report an evidence-qualified matrix for OpenCode v1, OpenCode v2 and Codex across ingestion, historical hydration, live observation, skill/delegation facts, child navigation, execution errors and recovery.

Codex historical hydration succeeds on the AL-55 copy (232 events, all five owners). Native/recorded-source HTTP reads, existing fixtures and disposable preservation/refresh probes distinguish working paths from v2 live failures, configured Codex-home selection, metadata child-link loss and error-summary limitations. Findings propose bounded follow-ups; no application, schema, runtime or new test changes.

[Capability matrix and evidence](alspaces/AL-63/CAPABILITIES.md) and [verification/limits](alspaces/AL-63/SUMMARY.md) retain source versions, identities, commands, attribution and unexecuted cases. Validation reuses the successful full make for identical application/test/build inputs and adds focused tests, disposable probes, document consistency and original-ledger hash checks. Original AL-55/AL-62 ledgers are unchanged. This PR delivers an audit, not implementation approval or cross-harness parity.

Complete the audit lifecycle by archiving the unchanged [original acceptance](alspaces/plan/done/AL-63%20verify%20source%20capabilities%20across%20harnesses.md) and promoting AL-65 to Ready. This does not approve a backfill-default change.
